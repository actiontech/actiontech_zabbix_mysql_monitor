package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	SHOW_STATUS = iota
	SHOW_VARIABLES
	SHOW_SLAVE_STATUS
	SHOW_MASTER_LOGS
	SHOW_PROCESSLIST
	SHOW_INNODB_STATUS
	SELECT_FROM_QUERY_RESPONSE_TIME_PERCONA
	SELECT_FROM_QUERY_RESPONSE_TIME_MYSQL
)

// pre key
const (
	SHOW_PROCESSLIST_STATE_PRE = "show_processlist_state_"
	SHOW_PROCESSLIST_TIME_PRE  = "show_processlist_time_"
)

var (
	mysqlSsl = false // Whether to use SSL to connect to MySQL.
	// TODO:support ssl conn
	/*	mysqlSslKey  = "/etc/pki/tls/certs/mysql/client-key.pem"
		mysqlSslCert = "/etc/pki/tls/certs/mysql/client-cert.pem"
		mysqlSslCa   = "/etc/pki/tls/certs/mysql/ca-cert.pem"*/

	host              = flag.String("host", "127.0.0.1", "`MySQL host`")
	user              = flag.String("user", "", "`MySQL username` (default: no default)")
	pass              = flag.String("pass", "", "`MySQL password` (default: no default)")
	port              = flag.String("port", "3306", "`MySQL port`")
	pollTime          = flag.Int("poll_time", 30, "Adjust to match your `polling interval`.if change, make sure change the wrapper.sh file too.")
	nocache           = flag.Bool("nocache", false, "Do not cache results in a file (default: false)")
	items             = flag.String("items", "", "-items <`item`,...> Comma-separated list of the items whose data you want (default: no default)")
	debugLog          = flag.String("debug_log", "", "If `debuglog` is a filename, it'll be used. (default: no default)")
	cacheDir          = flag.String("cache_dir", "/tmp", "A `path` for saving cache. if change, make sure change the wrapper.sh file too.")
	heartbeat         = flag.Bool("heartbeat", false, "Whether to use pt-heartbeat table for repl. delay calculation. (default: false)")
	heartbeatUtc      = flag.Bool("heartbeat_utc", false, "Whether pt-heartbeat is run with --utc option. (default: false)")
	heartbeatServerId = flag.String("heartbeat_server_id", "0", "`Server id` to associate with a heartbeat. Leave 0 if no preference. (default: 0)")
	heartbeatTable    = flag.String("heartbeat_table", "percona.heartbeat", "`db.tbl`.")
	innodb            = flag.Bool("innodb", true, "Whether to check InnoDB statistics")
	master            = flag.Bool("master", true, "Whether to check binary logging")
	slave             = flag.Bool("slave", true, "Whether to check slave status")
	procs             = flag.Bool("procs", true, "Whether to check SHOW PROCESSLIST")
	getQrtPercona     = flag.Bool("get_qrt_percona", true, "Whether to get response times from Percona Server or MariaDB")
	getQrtMysql       = flag.Bool("get_qrt_mysql", false, "Whether to get response times from MySQL (default: false)")
	discoveryPort     = flag.Bool("discovery_port", false, "`discovery mysqld port`, print in json format (default: false)")
	useSudo           = flag.Bool("sudo", true, "Use `sudo netstat...`")
	version           = flag.Bool("version", false, "print version")

	// log
	debugLogFile *os.File

	//regexps
	regSpaces     = regexp.MustCompile("\\s+")
	regNumbers    = regexp.MustCompile("\\d+")
	regIndividual = regexp.MustCompile("(?s)INDIVIDUAL BUFFER POOL INFO.*ROW OPERATIONS")

	Version string
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("version:", Version)
		os.Exit(1)
	}

	//debug file
	{
		if *debugLog != "" {
			var err error
			debugLogFile, err = os.OpenFile(*debugLog, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if nil != err {
				fmt.Printf("Can not create file using (%s as path), err: %s", *debugLog, err)
			}
			defer debugLogFile.Close()
			debugLogFile.Truncate(0)
		}
		log.SetOutput(debugLogFile)
	}

	if *discoveryPort {
		discoveryMysqldPort()
		log.Fatalln("discovery Mysqld Port done, exit.")
	}

	if *items == "mysqld_port_listen" {
		// default port: 3306
		if verifyMysqldPort(*port) {
			fmt.Println(*port, ":", 1)
		} else {
			fmt.Println(*port, ":", 0)
		}
		log.Fatalln("verify Mysqld Port done, exit.")
	}

	//param preprocessing
	{
		if *user == "" || *pass == "" {
			fmt.Fprintln(os.Stderr, "Please set mysql user and password use --user= --pass= !")
			log.Fatalln("flag user or pass is empty, exit !")
		}
		*host = strings.Replace(*host, ":", "", -1)
		*host = strings.Replace(*host, "/", "_", -1)
	}

	// check cache
	var cacheFile *os.File
	{
		cacheFilePath := *cacheDir + "/actiontech_" + *host + "_" + *port + "-mysql_zabbix_stats.txt"
		log.Printf("cacheFilePath = %s\n", cacheFilePath)
		if !(*nocache) {
			var err error
			cacheFile, err = checkCache(cacheFilePath)
			if nil != err {
				log.Fatalf("checkCache err:%s\n", err)
			}
			defer cacheFile.Close()
		} else {
			log.Println("Caching is disabled.")
		}

		// print item
		if !(*nocache) && (*items != "") && (cacheFile == nil) {
			printItemFromCacheFile(*items, cacheFilePath)
			log.Fatalln("read from cache file done.")
		}
	}

	collectionExist, collectionInfo := collect()
	result := parse(collectionExist, collectionInfo)
	print(result, cacheFile)
}

func collect() ([]bool, []map[string]string) {
	// Connect to MySQL.
	var db *sql.DB
	if mysqlSsl {
		log.Println("// TODO: Use mysql ssl")
	} else {
		var err error
		log.Printf("Connecting mysql: user %s, pass %s, host %s, port %s", *user, *pass, *host, *port)
		db, err = sql.Open("mysql", *user+":"+*pass+"@tcp("+*host+":"+*port+")/")
		if nil != err {
			log.Fatalf("sql.Open err:(%s) exit !", err)
		}
		defer db.Close()
		if err := db.Ping(); nil != err {
			log.Fatalf("db.Ping err:(%s) exit !", err)
		}
	}

	// Collecting ...
	collectionInfo := make([]map[string]string, 8)
	collectionExist := []bool{true, true, false, false, false, false, false, false}

	collectionInfo[SHOW_STATUS] = collectAllRowsToMap("variable_name", "value", db, "SHOW /*!50002 GLOBAL */ STATUS")
	collectionInfo[SHOW_VARIABLES] = collectAllRowsToMap("variable_name", "value", db, "SHOW VARIABLES")
	if *slave {
		collectionExist[SHOW_SLAVE_STATUS] = true
		collectionInfo[SHOW_SLAVE_STATUS] = make(map[string]string)
		// Leverage lock-free SHOW SLAVE STATUS if available
		queryResult := tryQueryIfAvailable(db, "SHOW SLAVE STATUS NONBLOCKING", "SHOW SLAVE STATUS NOLOCK", "SHOW SLAVE STATUS")
		log.Println("get slave status: ", queryResult)
		if 0 == len(queryResult) {
			log.Println("show slave empty, assume it is a master")
		} else {
			queryResult[0] = changeKeyCase(queryResult[0])
			if 1 < len(queryResult) {
				// Multi source replication
				log.Println("show slave multi rows, assume it is a multi source replication")

				var maxLag int64 = -1
				var totalRelayLogSpace int64 = 0
				for i, resultMap := range queryResult {
					resultMap = changeKeyCase(resultMap)
					if resultMap["slave_io_running"] != "Yes" {
						log.Printf("%dth row slave_io_running != Yes", i)
						queryResult[0]["slave_io_running"] = "No"
					}
					if resultMap["slave_sql_running"] != "Yes" {
						log.Printf("%dth row slave_sql_running != Yes", i)
						queryResult[0]["slave_sql_running"] = "No"
					}

					// get max slave_lag
					if resultMap["seconds_behind_master"] != "NULL" && convStrToInt64(resultMap["seconds_behind_master"]) > maxLag {
						log.Printf("%dth row seconds_behind_master may be max", i)
						maxLag = convStrToInt64(resultMap["seconds_behind_master"])
						queryResult[0]["seconds_behind_master"] = resultMap["seconds_behind_master"]
					}

					//get total relay_log_space
					log.Printf("%dth row relay_log_space is %s", i, resultMap["relay_log_space"])
					totalRelayLogSpace += convStrToInt64(resultMap["relay_log_space"])
				}
				queryResult[0]["relay_log_space"] = strconv.FormatInt(totalRelayLogSpace, 10)
			}
			stringMapAdd(collectionInfo[SHOW_SLAVE_STATUS], queryResult[0])
			/*			stringMapAdd(collectionInfo[SHOW_SLAVE_STATUS], collectFirstRowAsMapValue("relay_log_space", "relay_log_space", db, "SHOW SLAVE STATUS NONBLOCKING", "SHOW SLAVE STATUS NOLOCK", "SHOW SLAVE STATUS"))
						stringMapAdd(collectionInfo[SHOW_SLAVE_STATUS], collectFirstRowAsMapValue("seconds_behind_master", "seconds_behind_master", db, "SHOW SLAVE STATUS NONBLOCKING", "SHOW SLAVE STATUS NOLOCK", "SHOW SLAVE STATUS"))
			*/
			//Check replication heartbeat, if present.
			stringMapAdd(collectionInfo[SHOW_SLAVE_STATUS], getDelayFromHeartbeat(db))
		}
		log.Println("collectionInfo slave: ", collectionInfo[SHOW_SLAVE_STATUS])
	}

	// Get SHOW MASTER STATUS
	if *master && collectionInfo[SHOW_VARIABLES]["log_bin"] == "ON" {
		collectionExist[SHOW_MASTER_LOGS] = true
		// See issue percona #8
		collectionInfo[SHOW_MASTER_LOGS] = collectAllRowsAsMapValue("show_master_logs_value_", "file_size", db, "SHOW MASTER LOGS")
		log.Println("collectionInfo master : ", collectionInfo[SHOW_MASTER_LOGS])
	}

	// Get SHOW PROCESSLIST and aggregate it by state, sort by time
	if *procs {
		collectionExist[SHOW_PROCESSLIST] = true
		collectionInfo[SHOW_PROCESSLIST] = collectMultiColumnAllRowsAsMapValue([]string{SHOW_PROCESSLIST_STATE_PRE, SHOW_PROCESSLIST_TIME_PRE},
			[]string{"state", "time"}, db, "SHOW PROCESSLIST")
		log.Println("collectionInfo show processlist:", collectionInfo[SHOW_PROCESSLIST])
	}

	// Get SHOW INNODB STATUS and extract the desired metrics from it
	if *innodb {
		collectionExist[SHOW_INNODB_STATUS] = true
		engineMap := collectAllRowsToMap("engine", "support", db, "SHOW ENGINES")
		if value, ok := engineMap["InnoDB"]; ok && (value == "YES" || value == "DEFAULT") {
			collectionInfo[SHOW_INNODB_STATUS] = collectFirstRowAsMapValue("innodb_status_text", "status", db, "SHOW /*!50000 ENGINE*/ INNODB STATUS")
			// Get response time histogram from Percona Server or MariaDB if enabled.
			if !*getQrtPercona {
				log.Println("Not getting time histogram percona because it is not enabled")
			} else if collectionInfo[SHOW_VARIABLES]["have_response_time_distribution"] == "YES" || collectionInfo[SHOW_VARIABLES]["query_response_time_stats"] == "ON" {
				collectionExist[SELECT_FROM_QUERY_RESPONSE_TIME_PERCONA] = true
				collectionInfo[SELECT_FROM_QUERY_RESPONSE_TIME_PERCONA] = make(map[string]string)
				log.Println("Getting query time histogram percona")
				stringMapAdd(collectionInfo[SELECT_FROM_QUERY_RESPONSE_TIME_PERCONA], collectRowsAsMapValue("Query_time_count_", "count", 14, db, "SELECT `count`, ROUND(total * 1000000) AS total FROM INFORMATION_SCHEMA.QUERY_RESPONSE_TIME WHERE `time` <> 'TOO LONG'"))
				stringMapAdd(collectionInfo[SELECT_FROM_QUERY_RESPONSE_TIME_PERCONA], collectRowsAsMapValue("Query_time_total_", "total", 14, db, "SELECT `count`, ROUND(total * 1000000) AS total FROM INFORMATION_SCHEMA.QUERY_RESPONSE_TIME WHERE `time` <> 'TOO LONG'"))
			}
		}
	}

	// Get Query Response Time histogram from MySQL if enable.
	if (!*getQrtMysql) || (collectionInfo[SHOW_VARIABLES]["performance_schema"] != "ON") {
		log.Println("Not getting time histogram mysql because it is not enabled")
	} else {
		collectionExist[SELECT_FROM_QUERY_RESPONSE_TIME_MYSQL] = true
		collectionInfo[SELECT_FROM_QUERY_RESPONSE_TIME_MYSQL] = make(map[string]string)
		query := `SELECT 'query_rt100s' as rt, ifnull(sum(COUNT_STAR),0) as cnt FROM performance_schema.events_statements_summary_by_digest WHERE AVG_TIMER_WAIT >= 100000000000000 UNION
SELECT 'query_rt10s', ifnull(sum(COUNT_STAR),0) as cnt FROM performance_schema.events_statements_summary_by_digest WHERE AVG_TIMER_WAIT BETWEEN 10000000000000 AND 10000000000000 UNION 
SELECT 'query_rt1s', ifnull(sum(COUNT_STAR),0) as cnt FROM performance_schema.events_statements_summary_by_digest WHERE AVG_TIMER_WAIT BETWEEN 1000000000000 AND 10000000000000 UNION 
SELECT 'query_rt100ms', ifnull(sum(COUNT_STAR),0) as cnt FROM performance_schema.events_statements_summary_by_digest WHERE AVG_TIMER_WAIT BETWEEN 100000000000 AND 1000000000000 UNION 
SELECT 'query_rt10ms', ifnull(sum(COUNT_STAR),0) as cnt FROM performance_schema.events_statements_summary_by_digest WHERE AVG_TIMER_WAIT BETWEEN 10000000000 AND 100000000000 UNION 
SELECT 'query_rt1ms', ifnull(sum(COUNT_STAR),0) as cnt FROM performance_schema.events_statements_summary_by_digest WHERE AVG_TIMER_WAIT BETWEEN 1000000000 AND 10000000000 UNION 
SELECT 'query_rt100us', ifnull(sum(COUNT_STAR),0) as cnt FROM performance_schema.events_statements_summary_by_digest WHERE AVG_TIMER_WAIT <= 1000000000`

		stringMapAdd(collectionInfo[SELECT_FROM_QUERY_RESPONSE_TIME_MYSQL], collectAllRowsToMap("rt", "cnt", db, query))
		stringMapAdd(collectionInfo[SELECT_FROM_QUERY_RESPONSE_TIME_MYSQL], collectFirstRowAsMapValue("query_avgrt", "avgrt", db, "select round(avg(AVG_TIMER_WAIT)/1000/1000/1000,2) as avgrt from performance_schema.events_statements_summary_by_digest"))
	}

	return collectionExist, collectionInfo
}

func parse(collectionExist []bool, collectionInfo []map[string]string) map[string]string {
	// Holds the result of SHOW STATUS, SHOW INNODB STATUS, etc
	stat := make(map[string]string)
	stringMapAdd(stat, collectionInfo[SHOW_STATUS])
	stringMapAdd(stat, collectionInfo[SHOW_VARIABLES])

	// Make table_open_cache backwards-compatible (issue 63).
	if val, ok := stat["table_open_cache"]; ok {
		stat["table_cache"] = val
	}

	// Compute how much of the key buffer is used and unflushed (issue 127).
	if val1, err := strconv.ParseInt(stat["key_buffer_size"], 10, 64); err == nil {
		if val2, err := strconv.ParseInt(stat["Key_blocks_unused"], 10, 64); err == nil {
			if val3, err := strconv.ParseInt(stat["key_cache_block_size"], 10, 64); err == nil {
				if val4, err := strconv.ParseInt(stat["Key_blocks_not_flushed"], 10, 64); err == nil {
					log.Println("calc Key_buf_bytes_used, Key_buf_bytes_unflushed...")
					stat["Key_buf_bytes_used"] = strconv.FormatInt(val1-(val2*val3), 10)
					stat["Key_buf_bytes_unflushed"] = strconv.FormatInt(val4*val3, 10)
				}
			}
		}
	}

	if collectionExist[SHOW_SLAVE_STATUS] {
		if 0 == len(collectionInfo[SHOW_SLAVE_STATUS]) {
			// it is a master, set running_slave = 1 to avoid "slave stop" trigger
			log.Println("it is a master!set running_slave = 1")
			stat["running_slave"] = "1"
			stat["slave_lag"] = "0"
		} else if collectionInfo[SHOW_SLAVE_STATUS]["slave_io_running"] == "Yes" && collectionInfo[SHOW_SLAVE_STATUS]["slave_sql_running"] == "Yes" {
			stat["running_slave"] = "1"
		} else {
			stat["running_slave"] = "0"
		}

		if v, ok := collectionInfo[SHOW_SLAVE_STATUS]["relay_log_space"]; ok {
			stat["relay_log_space"] = v
		}

		if v, ok := collectionInfo[SHOW_SLAVE_STATUS]["seconds_behind_master"]; ok {
			stat["slave_lag"] = v
		}

		if v, ok := collectionInfo[SHOW_SLAVE_STATUS]["delay_from_heartbeat"]; ok {
			stat["slave_lag"] = v
		}

		//Scale slave_running and slave_stopped relative to the slave lag.
		if collectionInfo[SHOW_SLAVE_STATUS]["slave_sql_running"] == "Yes" {
			stat["slave_running"] = stat["slave_lag"]
			stat["slave_stopped"] = "0"
		} else {
			stat["slave_running"] = "0"
			stat["slave_stopped"] = stat["slave_lag"]
		}
	}

	if collectionExist[SHOW_MASTER_LOGS] {
		var binaryLogSpace int64 = 0
		// Older versions of MySQL may not have the File_size column in the
		// results of the command.  Zero-size files indicate the user is
		// deleting binlogs manually from disk (bad user! bad!).
		for _, value := range collectionInfo[SHOW_MASTER_LOGS] {
			if size, err := strconv.ParseInt(value, 10, 64); nil != err {
				log.Println("convert file_size to int fail:", err)
			} else if size > 0 {
				binaryLogSpace += size
			}
		}
		if binaryLogSpace > 0 {
			stat["binary_log_space"] = strconv.FormatInt(binaryLogSpace, 10)
		} else {
			stat["binary_log_space"] = "NULL"
		}
	}

	{
		procsStateMap := map[string]int64{
			// Values for the 'State' column from SHOW PROCESSLIST (converted to
			// lowercase, with spaces replaced by underscores)
			"State_closing_tables":       0,
			"State_copying_to_tmp_table": 0,
			"State_end":                  0,
			"State_freeing_items":        0,
			"State_init":                 0,
			"State_locked":               0,
			"State_login":                0,
			"State_preparing":            0,
			"State_reading_from_net":     0,
			"State_sending_data":         0,
			"State_sorting_result":       0,
			"State_statistics":           0,
			"State_updating":             0,
			"State_writing_to_net":       0,
			"State_none":                 0,
			"State_other":                0, // Everything not listed above
		}
		procsTimeMap := map[string]int64{
			"Time_top_1":  0,
			"Time_top_2":  0,
			"Time_top_3":  0,
			"Time_top_4":  0,
			"Time_top_5":  0,
			"Time_top_6":  0,
			"Time_top_7":  0,
			"Time_top_8":  0,
			"Time_top_9":  0,
			"Time_top_10": 0,
		}
		procsTimeSort := []int{}
		if collectionExist[SHOW_PROCESSLIST] {
			var state string
			reg := regexp.MustCompile("^(Table lock|Waiting for .*lock)$")
			for key, value := range collectionInfo[SHOW_PROCESSLIST] {
				if strings.HasPrefix(key, SHOW_PROCESSLIST_STATE_PRE) {
					if value == "" {
						value = "none"
					}
					// MySQL 5.5 replaces the 'Locked' state with a variety of "Waiting for
					// X lock" types of statuses.  Wrap these all back into "Locked" because
					// we don't really care about the type of locking it is.
					state = reg.ReplaceAllString(value, "locked")
					state = strings.Replace(strings.ToLower(value), " ", "_", -1)
					if _, ok := procsStateMap["State_"+state]; ok {
						procsStateMap["State_"+state]++
					} else {
						procsStateMap["State_other"]++
					}

				} else if strings.HasPrefix(key, SHOW_PROCESSLIST_TIME_PRE) {
					time, err := strconv.Atoi(value)
					if nil != err {
						log.Printf("show processlist: convert time %v error %v\n", value, err)
					}
					procsTimeSort = append(procsTimeSort, time)
				} else {
					log.Printf("show processlist: unexpect show processlist pre key %v\n", key)
				}

			}
		}

		sort.Sort(sort.Reverse(sort.IntSlice(procsTimeSort)))
		for i, t := range procsTimeSort {
			if _, ok := procsTimeMap["Time_top_"+strconv.Itoa(i+1)]; ok {
				procsTimeMap["Time_top_"+strconv.Itoa(i+1)] = int64(t)
			}
		}

		intMapAdd(stat, procsStateMap)
		intMapAdd(stat, procsTimeMap)
	}

	if value, ok := collectionInfo[SHOW_INNODB_STATUS]["innodb_status_text"]; ok {
		result := parseInnodbStatusWithRule(value)
		// percona comments:
		// TODO: I'm not sure what the deal is here; need to debug this.  But the
		// unflushed log bytes spikes a lot sometimes and it's impossible for it to
		// be more than the log buffer.
		log.Println("Unflushed log: ", result["unflushed_log"])
		val := convStrToInt64(stat["innodb_log_buffer_size"])
		if result["unflushed_log"] > 0 && result["unflushed_log"] < val {
			result["unflushed_log"] = val
		}

		// Now copy the values into stat.
		intMapAdd(stat, result)

		// Override values from InnoDB parsing with values from SHOW STATUS,
		// because InnoDB status might not have everything and the SHOW STATUS is
		// to be preferred where possible.
		overrides := map[string]string{
			"Innodb_buffer_pool_pages_data":    "database_pages",
			"Innodb_buffer_pool_pages_dirty":   "modified_pages",
			"Innodb_buffer_pool_pages_free":    "free_pages",
			"Innodb_buffer_pool_pages_total":   "pool_size",
			"Innodb_data_fsyncs":               "file_fsyncs",
			"Innodb_data_pending_reads":        "pending_normal_aio_reads",
			"Innodb_data_pending_writes":       "pending_normal_aio_writes",
			"Innodb_os_log_pending_fsyncs":     "pending_log_flushes",
			"Innodb_pages_created":             "pages_created",
			"Innodb_pages_read":                "pages_read",
			"Innodb_pages_written":             "pages_written",
			"Innodb_rows_deleted":              "rows_deleted",
			"Innodb_rows_inserted":             "rows_inserted",
			"Innodb_rows_read":                 "rows_read",
			"Innodb_rows_updated":              "rows_updated",
			"Innodb_buffer_pool_reads":         "pool_reads",
			"Innodb_buffer_pool_read_requests": "pool_read_requests",
		}

		// If the SHOW STATUS value exists, override...
		for k, v := range overrides {
			if val, ok := stat[k]; ok {
				log.Println("Override " + k)
				stat[v] = val
			}
		}
	}

	if collectionExist[SELECT_FROM_QUERY_RESPONSE_TIME_PERCONA] {
		stringMapAdd(stat, collectionInfo[SELECT_FROM_QUERY_RESPONSE_TIME_PERCONA])
	}

	if collectionExist[SELECT_FROM_QUERY_RESPONSE_TIME_MYSQL] {
		stringMapAdd(stat, collectionInfo[SELECT_FROM_QUERY_RESPONSE_TIME_MYSQL])
	}

	return stat

}

func print(result map[string]string, fp *os.File) {
	// Define the variables to output.
	key := []string{
		"Key_read_requests",
		"Key_reads",
		"Key_write_requests",
		"Key_writes",
		"history_list",
		"innodb_transactions",
		"read_views",
		"current_transactions",
		"locked_transactions",
		"active_transactions",
		"pool_size",
		"free_pages",
		"database_pages",
		"modified_pages",
		"pages_read",
		"pages_created",
		"pages_written",
		"file_fsyncs",
		"file_reads",
		"file_writes",
		"log_writes",
		"pending_aio_log_ios",
		"pending_aio_sync_ios",
		"pending_buf_pool_flushes",
		"pending_chkp_writes",
		"pending_ibuf_aio_reads",
		"pending_log_flushes",
		"pending_log_writes",
		"pending_normal_aio_reads",
		"pending_normal_aio_writes",
		"ibuf_inserts",
		"ibuf_merged",
		"ibuf_merges",
		"spin_waits",
		"spin_rounds",
		"os_waits",
		"rows_inserted",
		"rows_updated",
		"rows_deleted",
		"rows_read",
		"Table_locks_waited",
		"Table_locks_immediate",
		"Slow_queries",
		"Open_files",
		"Open_tables",
		"Opened_tables",
		"innodb_open_files",
		"open_files_limit",
		"table_cache",
		"Aborted_clients",
		"Aborted_connects",
		"Max_used_connections",
		"Slow_launch_threads",
		"Threads_cached",
		"Threads_connected",
		"Threads_created",
		"Threads_running",
		"max_connections",
		"thread_cache_size",
		"Connections",
		"slave_running",
		"slave_stopped",
		"Slave_retried_transactions",
		"slave_lag",
		"Slave_open_temp_tables",
		"Qcache_free_blocks",
		"Qcache_free_memory",
		"Qcache_hits",
		"Qcache_inserts",
		"Qcache_lowmem_prunes",
		"Qcache_not_cached",
		"Qcache_queries_in_cache",
		"Qcache_total_blocks",
		"query_cache_size",
		"Questions",
		"Com_update",
		"Com_insert",
		"Com_select",
		"Com_delete",
		"Com_replace",
		"Com_load",
		"Com_update_multi",
		"Com_insert_select",
		"Com_delete_multi",
		"Com_replace_select",
		"Select_full_join",
		"Select_full_range_join",
		"Select_range",
		"Select_range_check",
		"Select_scan",
		"Sort_merge_passes",
		"Sort_range",
		"Sort_rows",
		"Sort_scan",
		"Created_tmp_tables",
		"Created_tmp_disk_tables",
		"Created_tmp_files",
		"Bytes_sent",
		"Bytes_received",
		"innodb_log_buffer_size",
		"unflushed_log",
		"log_bytes_flushed",
		"log_bytes_written",
		"relay_log_space",
		"binlog_cache_size",
		"Binlog_cache_disk_use",
		"Binlog_cache_use",
		"binary_log_space",
		"innodb_locked_tables",
		"innodb_lock_structs",
		"State_closing_tables",
		"State_copying_to_tmp_table",
		"State_end",
		"State_freeing_items",
		"State_init",
		"State_locked",
		"State_login",
		"State_preparing",
		"State_reading_from_net",
		"State_sending_data",
		"State_sorting_result",
		"State_statistics",
		"State_updating",
		"State_writing_to_net",
		"State_none",
		"State_other",
		"Time_top_1",
		"Time_top_2",
		"Time_top_3",
		"Time_top_4",
		"Time_top_5",
		"Time_top_6",
		"Time_top_7",
		"Time_top_8",
		"Time_top_9",
		"Time_top_10",
		"Handler_commit",
		"Handler_delete",
		"Handler_discover",
		"Handler_prepare",
		"Handler_read_first",
		"Handler_read_key",
		"Handler_read_next",
		"Handler_read_prev",
		"Handler_read_rnd",
		"Handler_read_rnd_next",
		"Handler_rollback",
		"Handler_savepoint",
		"Handler_savepoint_rollback",
		"Handler_update",
		"Handler_write",
		"innodb_tables_in_use",
		"innodb_lock_wait_secs",
		"hash_index_cells_total",
		"hash_index_cells_used",
		"total_mem_alloc",
		"additional_pool_alloc",
		"uncheckpointed_bytes",
		"ibuf_used_cells",
		"ibuf_free_cells",
		"ibuf_cell_count",
		"adaptive_hash_memory",
		"page_hash_memory",
		"dictionary_cache_memory",
		"file_system_memory",
		"lock_system_memory",
		"recovery_system_memory",
		"thread_hash_memory",
		"innodb_sem_waits",
		"innodb_sem_wait_time_ms",
		"Key_buf_bytes_unflushed",
		"Key_buf_bytes_used",
		"key_buffer_size",
		"Innodb_row_lock_time",
		"Innodb_row_lock_waits",
		"Query_time_count_00",
		"Query_time_count_01",
		"Query_time_count_02",
		"Query_time_count_03",
		"Query_time_count_04",
		"Query_time_count_05",
		"Query_time_count_06",
		"Query_time_count_07",
		"Query_time_count_08",
		"Query_time_count_09",
		"Query_time_count_10",
		"Query_time_count_11",
		"Query_time_count_12",
		"Query_time_count_13",
		"Query_time_total_00",
		"Query_time_total_01",
		"Query_time_total_02",
		"Query_time_total_03",
		"Query_time_total_04",
		"Query_time_total_05",
		"Query_time_total_06",
		"Query_time_total_07",
		"Query_time_total_08",
		"Query_time_total_09",
		"Query_time_total_10",
		"Query_time_total_11",
		"Query_time_total_12",
		"Query_time_total_13",
		"wsrep_replicated_bytes",
		"wsrep_received_bytes",
		"wsrep_replicated",
		"wsrep_received",
		"wsrep_local_cert_failures",
		"wsrep_local_bf_aborts",
		"wsrep_local_send_queue",
		"wsrep_local_recv_queue",
		"wsrep_cluster_size",
		"wsrep_cert_deps_distance",
		"wsrep_apply_window",
		"wsrep_commit_window",
		"wsrep_flow_control_paused",
		"wsrep_flow_control_sent",
		"wsrep_flow_control_recv",
		"pool_reads",
		"pool_read_requests",
		"running_slave",
		"query_rt100s",
		"query_rt10s",
		"query_rt1s",
		"query_rt100ms",
		"query_rt10ms",
		"query_rt1ms",
		"query_rtavg",
		"query_avgrt",
	}

	// Return the output.
	output := make(map[string]string)
	for _, v := range key {
		// If the value isn't defined, return -1 which is lower than (most graphs')
		// minimum value of 0, so it'll be regarded as a missing value.
		if val, ok := result[v]; ok {
			output[v] = val
		} else {
			output[v] = "-1"
		}
	}

	if fp != nil {
		log.Println("write file:")
		enc := json.NewEncoder(fp)
		enc.Encode(output)
		/*		for k, v := range output {
				fmt.Fprintf(fp, "%v:%v ", k, v)
			}*/
	}

	if *items != "" {
		fields := strings.Split(*items, ",")
		for _, field := range fields {
			if val, ok := output[field]; ok {
				fmt.Println(field, ":", val)
			} else {
				fmt.Println(field + " do not exist")
			}
		}

	}

}

func printItemFromCacheFile(items string, filename string) {
	var data map[string]string
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln("printItemFromCacheFile ReadFile err:", err)
	}
	if err := json.Unmarshal(f, &data); err != nil {
		log.Fatalln("printItemFromCacheFile decode json err:", err)
	}
	fields := strings.Split(items, ",")
	for _, field := range fields {
		if val, ok := data[field]; ok {
			fmt.Println(field, ":", val)
		} else {
			fmt.Println(field + " do not exist")
		}
	}
}

func getDelayFromHeartbeat(db *sql.DB) map[string]string {
	result := make(map[string]string)
	if !*heartbeat {
		return result
	}
	var nowFunc string
	if *heartbeatUtc {
		nowFunc = "UNIX_TIMESTAMP(UTC_TIMESTAMP)"
	} else {
		nowFunc = "UNIX_TIMESTAMP()"
	}
	query := fmt.Sprintf("SELECT MAX(%s - ROUND(UNIX_TIMESTAMP(ts))) AS delay FROM %s WHERE %s = 0 OR server_id = %s", nowFunc, *heartbeatTable, *heartbeatServerId, *heartbeatServerId)
	result = collectFirstRowAsMapValue("delay_from_heartbeat", "delay", db, query)
	return result
}

func query(db *sql.DB, query string) []map[string]string {
	log.Println("exec query:", query)
	result := make([]map[string]string, 0, 500)

	rows, err := db.Query(query)
	if nil != err {
		log.Println("db.Query err:", err)
		return result
	}
	defer func(rows *sql.Rows) {
		if rows != nil {
			rows.Close()
		}
	}(rows)

	columnsName, err := rows.Columns()
	if nil != err {
		log.Println("rows.Columns err:", err)
		return result
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columnsName))
	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if nil != err {
			log.Println("rows.Next err:", err)
		}

		// Now do something with the data.
		row_map := make(map[string]string)
		for i, col := range values {
			if col == nil {
				row_map[columnsName[i]] = "NULL"
			} else {
				row_map[columnsName[i]] = string(col)
			}
		}
		result = append(result, row_map)
	}

	err = rows.Err()
	if nil != err {
		log.Println("rows.Err:", err)
	}
	return result
}

func tryQueryIfAvailable(db *sql.DB, querys ...string) []map[string]string {
	result := make([]map[string]string, 0, 500)
	for _, q := range querys {
		result = query(db, q)
		if 0 == len(result) {
			log.Println("tryQueryIfAvailable:Got nothing from sql: ", q)
			continue
		}
		return result
	}
	return result
}

// collect two columns value to map
func collectAllRowsToMap(keyColName string, valueColName string, db *sql.DB, querys ...string) map[string]string {
	result := make(map[string]string)
	for _, mp := range tryQueryIfAvailable(db, querys...) {
		// Must lowercase keys because different MySQL versions have different lettercase.
		mp = changeKeyCase(mp)
		result[mp[keyColName]] = mp[valueColName]
	}
	return result
}

// collect one column value as mapValue in the first row
func collectFirstRowAsMapValue(key string, valueColName string, db *sql.DB, querys ...string) map[string]string {
	result := make(map[string]string)
	queryResult := tryQueryIfAvailable(db, querys...)
	if 0 == len(queryResult) {
		log.Println("collectFirstRowAsMapValue:Got nothing from query: ", querys)
		return result
	}
	mp := changeKeyCase(queryResult[0])
	if _, ok := mp[valueColName]; !ok {
		log.Printf("collectFirstRowAsMapValue:Couldn't get %s from %s\n", valueColName, querys)
		return result
	}
	result[key] = mp[valueColName]
	return result
}

// collect one column value as mapValue in all rows
func collectAllRowsAsMapValue(preKey string, valueColName string, db *sql.DB, querys ...string) map[string]string {
	result := make(map[string]string)
	for i, mp := range tryQueryIfAvailable(db, querys...) {
		mp = changeKeyCase(mp)
		if _, ok := mp[valueColName]; !ok {
			log.Printf("collectAllRowsAsMapValue:Couldn't get %s from %s\n", valueColName, querys)
			return result
		}
		result[preKey+strconv.Itoa(i)] = mp[valueColName]
	}
	return result
}

// collect multi column value as mapValue in all rows
func collectMultiColumnAllRowsAsMapValue(preKeys []string, valueColNames []string, db *sql.DB, querys ...string) map[string]string {
	result := make(map[string]string)
	if len(preKeys) != len(valueColNames) {
		log.Printf("collectMultiColumnAllRowsAsMapValue:prekey length is not match with valueColName length\n")
		return result
	}
	for i, mp := range tryQueryIfAvailable(db, querys...) {
		mp = changeKeyCase(mp)
		for j, preKey := range preKeys {
			valueColName := valueColNames[j]
			if _, ok := mp[valueColName]; !ok {
				log.Printf("collectMultiColumnAllRowsAsMapValue:Couldn't get %s from %s\n", valueColName, querys)
				return result
			}
			result[preKey+strconv.Itoa(i)] = mp[valueColName]
		}
	}
	return result
}

// collect one column value as mapValue in pre lines rows
func collectRowsAsMapValue(preKey string, valueColName string, lines int, db *sql.DB, querys ...string) map[string]string {
	result := make(map[string]string)
	line := 0
	resultMap := tryQueryIfAvailable(db, querys...)
	for i, mp := range resultMap {
		if i >= lines {
			// It's possible that the number of rows returned isn't lines.
			// Don't add extra status counters.
			break
		}
		mp = changeKeyCase(mp)
		if _, ok := mp[valueColName]; !ok {
			log.Printf("collectRowsAsMapValue: Couldn't get %s from %s\n", valueColName, querys)
			return result
		}
		result[fmt.Sprintf("%s%02d", preKey, i)] = mp[valueColName]
		line++
	}

	// It's also possible that the number of rows returned is too few.
	// Don't leave any status counters unassigned; it will break graphs.
	for line < lines {
		result[fmt.Sprintf("%s%02d", preKey, line)] = "0"
		line++
	}
	return result
}

func parseInnodbStatusWithRule(status string) map[string]int64 {
	result := map[string]int64{
		"current_transactions": 0,
		"active_transactions":  0,
		"locked_transactions":  0,
	}

	// discard INDIVIDUAL BUFFER POOL INFO, only parse total BUFFER POOL AND MEMORY
	status = regIndividual.ReplaceAllString(status, "")

	txnSeen := false
	preField := ""
	fields := strings.Split(status, "\n")
	for _, field := range fields {
		field := strings.TrimSpace(field)
		log.Println("parseInnodbStatusWithRule:Field:", field)
		handleInnodbStatusField(field, preField, &txnSeen, &result)
		preField = field
	}
	return result
}

func handleInnodbStatusField(field string, preField string, txnSeen *bool, result *map[string]int64) {
	var ruleDesc = `
	// SEMAPHORES
	// Mutex spin waits 79626940, rounds 157459864, OS waits 698719
	// Mutex spin waits 0, rounds 247280272495, OS waits 316513438
	^Mutex spin waits | 3 | spin_waits | - 
	^Mutex spin waits | 5 | spin_rounds | -
	^Mutex spin waits | 8 | os_waits | -

	// RW-shared spins 3859028, OS waits 2100750; RW-excl spins 4641946, OS waits 1530310
	^RW-shared spins.*;.* | 2 | spin_waits | -
	^RW-shared spins.*;.* | 8 | spin_waits | -
	^RW-shared spins.*;.* | 5 | os_waits | -
	^RW-shared spins.*;.* | 11 | os_waits | -

	// Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
	// RW-shared spins 604733, rounds 8107431, OS waits 241268
	^RW-shared spins&&&&?!; RW-excl spins | 2 | spin_waits | -
	^RW-shared spins&&&&?!; RW-excl spins | 4 | spin_rounds | -
	^RW-shared spins&&&&?!; RW-excl spins | 7 | os_waits | -

	// Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
	// RW-excl spins 604733, rounds 8107431, OS waits 62
	^RW-excl spins&&&&?!; RW-excl spins | 2 | spin_waits | -
	^RW-excl spins&&&&?!; RW-excl spins | 4 | spin_rounds | -
	^RW-excl spins&&&&?!; RW-excl spins | 7 | os_waits | -

	// mysql 5.7: "Mutex spin" information has been split out into three kinds (RW-shared, RW-excl, RW-sx)
	// RW-sx spins 8947, rounds 258579, OS waits 8298
	^RW-sx spins | 2 | spin_waits | -
	^RW-sx spins | 4 | spin_rounds | -
	^RW-sx spins | 7 | os_waits | -

	// --Thread 907205 has waited at handler/ha_innodb.cc line 7156 for 1.00 seconds the semaphore:
	seconds the semaphore: | 9 | innodb_sem_wait_time_ms | addSemWait

	// TRANSACTIONS
	// The beginning of the TRANSACTIONS section: start counting transactions
	// Trx id counter 0 1170664159
	// Trx id counter 861144
	// percona Assume it is a hex string representation. 
	^Trx id counter\s+\d+$ | 3 | innodb_transactions | -
	^Trx id counter\s+\d+\s+\d+$ | 3 | innodb_transactions | bigAdd

	//  Purge done for trx's n:o < 0 1170663853 undo n:o < 0 0
	//  Purge done for trx's n:o < 861B135D undo n:o < 0
	//  percona Assume it is a hex string representation. 
	^Purge done for trx | 6 | unpurged_txns | purge

	// History list length 132
	^History list length | 3 | history_list | -

	// ---TRANSACTION 0, not started, process no 13510, OS thread id 1170446656
	^---TRANSACTION | 0 | current_transactions | txnSeenInc
	^---TRANSACTION&&&&ACTIVE | 0 | active_transactions | txnSeenInc

	// ------- TRX HAS BEEN WAITING 32 SEC FOR THIS LOCK TO BE GRANTED:
	^------- TRX HAS BEEN | 5 | innodb_lock_wait_secs | txnSeen-

	// 1 read views open inside InnoDB
	read views open inside InnoDB | 0 | read_views | -

	// mysql tables in use 2, locked 2
	^mysql tables in use | 4 | innodb_tables_in_use | -
	^mysql tables in use | 6 | innodb_locked_tables | - 

	// 23 lock struct(s), heap size 3024, undo log entries 27
	// LOCK WAIT 12 lock struct(s), heap size 3024, undo log entries 5
	// LOCK WAIT 2 lock struct(s), heap size 368
	lock struct\(s\)&&&&^LOCK WAIT | 2 | innodb_lock_structs | txnSeen-
	lock struct\(s\)&&&&^LOCK WAIT | 0 | locked_transactions | txnSeenInc
	lock struct\(s\)&&&&?!^LOCK WAIT | 0 | innodb_lock_structs | txnSeen-

	// FILE I/O
	// 8782182 OS file reads, 15635445 OS file writes, 947800 OS fsyncs
	OS file reads, | 0 | file_reads | -
	OS file reads, | 4 | file_writes | -
	OS file reads, | 8 | file_fsyncs | -

    // Pending normal aio reads: 0, aio writes: 0,
	// mysql 5.7: Pending normal aio reads: [0, 0, 0, 0] , aio writes: [0, 0, 0, 0] ,
	// mysql 5.7: Pending normal aio reads:, aio writes:,
	// Pending aio read/write information is still present, however the numbers are now split to show the number of waits per I/O thread 
	^Pending normal aio reads:(.*), aio writes:(.*), | 1 | pending_normal_aio_reads | NumOrEmptyOrArrayRegexp
	^Pending normal aio reads:(.*), aio writes:(.*), | 2 | pending_normal_aio_writes | NumOrEmptyOrArrayRegexp

	// ibuf aio reads: 0, log i/o's: 0, sync i/o's: 0
	// mysql 5.7: ibuf aio reads:, log i/o's:, sync i/o's:
	^ibuf aio reads:(.*), log i/o's:(.*), sync i/o's:(.*) | 1 | pending_ibuf_aio_reads | NumOrEmptyOrArrayRegexp
	^ibuf aio reads:(.*), log i/o's:(.*), sync i/o's:(.*) | 2 | pending_aio_log_ios | NumOrEmptyOrArrayRegexp
	^ibuf aio reads:(.*), log i/o's:(.*), sync i/o's:(.*) | 3 | pending_aio_sync_ios | NumOrEmptyOrArrayRegexp

	// Pending flushes (fsync) log: 0; buffer pool: 0
	^Pending flushes \(fsync\) | 4 | pending_log_flushes | -
	^Pending flushes \(fsync\) | 7 | pending_buf_pool_flushes | -

	// INSERT BUFFER AND ADAPTIVE HASH INDEX
	// Older InnoDB code seemed to be ready for an ibuf per tablespace.  It
	// had two lines in the output.  Newer has just one line, see below.
	// Ibuf for space 0: size 1, free list len 887, seg size 889, is not empty
	// Ibuf for space 0: size 1, free list len 887, seg size 889,
	^Ibuf for space 0: size&&&&?!is not empty | 5 | ibuf_used_cells | -
	^Ibuf for space 0: size&&&&?!is not empty | 9 | ibuf_free_cells | -
	^Ibuf for space 0: size&&&&?!is not empty | 12 | ibuf_cell_count | -

	// Ibuf: size 1, free list len 4634, seg size 4636,
	^Ibuf: size | 2 | ibuf_used_cells | -
	^Ibuf: size | 6 | ibuf_free_cells | -
	^Ibuf: size | 9 | ibuf_cell_count | -
	^Ibuf: size&&&&merges | 10 | ibuf_merges | -

	// Output of show engine innodb status has changed in 5.5
	// merged operations:
	// insert 593983, delete mark 387006, delete 73092
	, delete mark | 1 | ibuf_inserts | preField
	, delete mark | 1 | ibuf_merged | preField
	, delete mark | 4 | ibuf_merged | preField
	, delete mark | 6 | ibuf_merged | preField

	// 19817685 inserts, 19817684 merged recs, 3552620 merges
	merged recs, | 0 | ibuf_inserts | -
	merged recs, | 2 | ibuf_merged | -
	merged recs, | 5 | ibuf_merges | -

	// In some versions of InnoDB, the used cells is omitted.
	// Hash table size 4425293, used cells 4229064, ....
	// Hash table size 57374437, node heap has 72964 buffer(s) <-- no used , 10cells
	^Hash table size | 3 | hash_index_cells_total | -
	^Hash table size&&&&used cells | 6 | hash_index_cells_used | -

	// LOG
	// 3430041 log i/o's done, 17.44 log i/o's/second
	// 520835887 log i/o's done, 17.28 log i/o's/second, 518724686 syncs, 2980893 checkpoints
	log i/o's done, | 0 | log_writes | -

	// 0 pending log writes, 0 pending chkp writes
	pending log writes, | 0 | pending_log_writes | -
	pending log writes, | 4 | pending_chkp_writes | -

	// mysql 5.7: 1 pending log flushes, 0 pending chkp writes
	pending log flushes, | 0 | pending_log_writes | -
	pending log flushes, | 4 | pending_chkp_writes | -

	// This number is NOT printed in hex in InnoDB plugin.
    // Log sequence number 13093949495856 //plugin
	// Log sequence number 125 3934414864 //normal
	^Log sequence number\s+\d+\s+\d+$ | 3 | log_bytes_written | bigAdd
	^Log sequence number\s+\d+$ | 3 | log_bytes_written | -

	// This number is NOT printed in hex in InnoDB plugin.
	// Log flushed up to   13093948219327
	// Log flushed up to   125 3934414864
	^Log flushed up to\s+\d+\s+\d+$ | 4 | log_bytes_flushed | bigAdd
	^Log flushed up to\s+\d+$ | 4 | log_bytes_flushed | -

	// Last checkpoint at  125 3934293461
	^Last checkpoint at\s+\d+\s+\d+$ | 3 | last_checkpoint | bigAdd
	^Last checkpoint at\s+\d+$ | 3 | last_checkpoint | - 

	// BUFFER POOL AND MEMORY
	// Total memory allocated 29642194944; in additional pool allocated 0
	// Total memory allocated by read views 96
	^Total memory allocated&&&&in additional pool allocated | 3 | total_mem_alloc | -
	^Total memory allocated&&&&in additional pool allocated | 8 | additional_pool_alloc | -

	// mysql 5.7: Total large memory allocated 137428992
	^Total large memory allocated | 4 | total_mem_alloc | -
	
	// Adaptive hash index 1538240664 	(186998824 + 1351241840)
	^Adaptive hash index | 3 | adaptive_hash_memory | -

	//   Page hash           11688584
	^Page hash | 2 | page_hash_memory | -

	//   Dictionary cache    145525560 	(140250984 + 5274576)
	^Dictionary cache | 2 | dictionary_cache_memory | -

	//   File system         313848 	(82672 + 231176)
	^File system | 2 | file_system_memory | -

	//   Lock system         29232616 	(29219368 + 13248)
	^Lock system | 2 | lock_system_memory | -

	//   Recovery system     0 	(0 + 0)
	^Recovery system | 2 | recovery_system_memory | -

	//   Threads             409336 	(406936 + 2400)
	^Threads | 1 | thread_hash_memory | -

	//   innodb_io_pattern   0 	(0 + 0)
	^innodb_io_pattern | 1 | innodb_io_pattern_memory | -

	// The "\s+" after size is necessary to avoid matching the wrong line:
	// Buffer pool size        1769471
	// Buffer pool size, bytes 28991012864
	^Buffer pool size\s+ | 3 | pool_size | -

	// Free buffers            0
	^Free buffers | 2 | free_pages | -

	// Database pages          1696503
	^Database pages | 2 | database_pages | -

	// Modified db pages       160602
	^Modified db pages | 3 | modified_pages | -

	// Pages read 15240822, created 1770238, written 21705836
	^Pages read&&&&?!^Pages read ahead | 2 | pages_read | -
	^Pages read&&&&?!^Pages read ahead | 4 | pages_created | -
	^Pages read&&&&?!^Pages read ahead | 6 | pages_written | -

	// ROW OPERATIONS
	// Number of rows inserted 50678311, updated 66425915, deleted 20605903, read 454561562
	^Number of rows inserted | 4 | rows_inserted | -
	^Number of rows inserted | 6 | rows_updated | -
	^Number of rows inserted | 8 | rows_deleted | -
	^Number of rows inserted | 10 | rows_read | -

	// 0 queries inside InnoDB, 0 queries in queue
	queries inside InnoDB, | 0 | queries_inside | -
	queries inside InnoDB, | 4 | queries_queued | -`
	for _, rule := range strings.Split(ruleDesc, "\n") {
		rule = strings.TrimSpace(rule)
		if rule == "" || strings.HasPrefix(rule, "//") {
			continue
		}
		segs := strings.Split(rule, "|")
		val := strings.TrimSpace(segs[len(segs)-1])
		myRegexp := strings.TrimSpace(segs[0])
		rowIndexNum, err := strconv.Atoi(strings.TrimSpace(segs[1]))
		if nil != err {
			log.Fatalf("handleInnodbStatus: convert row index(%v) to int err:%v\n", segs[2], err)
		}
		rowKey := strings.TrimSpace(segs[2])

		if fieldRegexpJudge(field, myRegexp) {
			if strings.Contains(myRegexp, "^Trx id counter") {
				*txnSeen = true
			}
			if rowKey == "hash_index_cells_total" {
				(*result)["hash_index_cells_used"] = 0
			}
			row := regSpaces.Split(field, -1)
			if len(row) <= rowIndexNum {
				log.Fatalf("defaultHandle err: rowIndexNum(%v) out of range in slice(%v)\n", rowIndexNum, row)
			}
			switch val {
			case "-":
				doDefaultHandle(row, rowIndexNum, rowKey, result)
			case "addSemWait":
				doAddSemWaitHandle(row, rowIndexNum, rowKey, result)
			case "bigAdd":
				doBidAddHandle(row, rowIndexNum, rowKey, result)
			case "purge":
				doPurgeHandle(row, result)
			case "inc":
				doIncHandle(rowKey, result)
			case "txnSeen-":
				if *txnSeen {
					doDefaultHandle(row, rowIndexNum, rowKey, result)
				}
			case "txnSeenInc":
				if *txnSeen {
					doIncHandle(rowKey, result)
				}
			case "preField":
				if fieldRegexpJudge(preField, "merged operations:") {
					doDefaultHandle(row, rowIndexNum, rowKey, result)
				}
			case "NumOrEmptyOrArrayRegexp":
				parsedStr, err := parseRegexpSubmatch(field, myRegexp, rowIndexNum)
				if nil != err {
					log.Printf("NumOrEmptyOrArrayRegexp: parseRegexp err:(%v)", err)
					break
				}
				doNumOrEmptyOrArrayRegexpHandle(parsedStr, rowKey, result)
			}
		}
	}

	if val1, ok1 := (*result)["log_bytes_written"]; ok1 {
		if val2, ok2 := (*result)["log_bytes_flushed"]; ok2 {
			(*result)["unflushed_log"] = val1 - val2
		}

		if val2, ok2 := (*result)["last_checkpoint"]; ok2 {
			(*result)["uncheckpointed_bytes"] = val1 - val2
		}
	}
}

func doDefaultHandle(row []string, rowIndexNum int, rowKey string, result *map[string]int64) {
	i := convStrToInt64(row[rowIndexNum])
	log.Printf("doDefaultHandle: result[%v] += %v\n", rowKey, i)
	(*result)[rowKey] += i
}

func doAddSemWaitHandle(row []string, rowIndexNum int, rowKey string, result *map[string]int64) {
	(*result)["innodb_sem_waits"]++
	value := regNumbers.FindString(row[rowIndexNum])
	i, err := strconv.ParseFloat(value, 64)
	if nil != err {
		log.Printf("doAddSemWaitHandle err: parse(%v) to float64 err:%v\n", value, err)
	}
	log.Printf("doAddSemWaitHandle: result[%v] += %v * 1000\n", rowKey, value)
	(*result)[rowKey] += int64(i * 1000)
}

func doBidAddHandle(row []string, rowIndexNum int, rowKey string, result *map[string]int64) {
	i1 := convStrToInt64(row[rowIndexNum])
	i2 := convStrToInt64(row[rowIndexNum+1])
	log.Printf("doBidAddHandle: result[%v] =+ %v * (1<<32) + %v\n", rowKey, i1, i2)
	(*result)[rowKey] += i1*(1<<32) + i2
}

func doPurgeHandle(row []string, result *map[string]int64) {
	//  Purge done for trx's n:o < 0 1170663853 undo n:o < 0 0
	//  Purge done for trx's n:o < 861B135D undo n:o < 0
	var purgedTo int64
	if row[7] == "undo" {
		purgedTo = convStrToInt64(row[6]) // percona Assume it is a hex string representation.
	} else {
		purgedTo = (convStrToInt64(row[6]) * (1 << 32)) + convStrToInt64(row[7])
	}
	log.Printf("doPurgeHandle: result[\"unpurged_txns\"] = %v - %v\n", (*result)["innodb_transactions"], purgedTo)
	(*result)["unpurged_txns"] = (*result)["innodb_transactions"] - purgedTo
}

func doIncHandle(rowKey string, result *map[string]int64) {
	(*result)[rowKey]++
}

func parseRegexpSubmatch(field, reg string, regSubmatchNum int) (string, error) {
	if myReg, err := regexp.Compile(reg); nil == err {
		seg := myReg.FindStringSubmatch(field)
		if len(seg) <= regSubmatchNum {
			return "", fmt.Errorf("doRegexpHandle err: use reg(%v) FindStringSubmatch(%v) seg(%v) too short, need seg[%v]", reg, field, regSubmatchNum)
		}
		return seg[regSubmatchNum], nil
	} else {
		return "", err
	}
}

func doNumOrEmptyOrArrayRegexpHandle(parsedStr, rowKey string, result *map[string]int64) {
	parsedStr = strings.TrimSpace(parsedStr)
	// parsedStr may be "" "10" "[10, 10, 10]"
	if "" == parsedStr {
		(*result)[rowKey] = 0
	} else {
		parsedStr = strings.TrimFunc(parsedStr, func(r rune) bool {
			return r == '[' || r == ']'
		})
		// parsedStr may be "10" "10, 10, 10"
		for _, num := range strings.Split(parsedStr, ", ") {
			log.Printf("doNumOrEmptyOrArrayRegexpHandle: result[%v] add num: %v", rowKey, num)
			(*result)[rowKey] += convStrToInt64(num)
		}
	}
}

func convStrToInt64(s string) int64 {
	log.Println("convStrToInt64: string:", s)
	value := regexp.MustCompile("\\d+").FindString(s)
	i, err := strconv.ParseInt(value, 10, 64)
	if nil != err {
		log.Fatalf("convStrToInt64 err: parse(%v) to int64 err:%v\n", value, err)
	}
	return i
}

func fieldRegexpJudge(field string, myRegexp string) bool {
	regexps := strings.Split(myRegexp, "&&&&")
	for _, reg := range regexps {
		if strings.HasPrefix(reg, "?!") {
			if regexp.MustCompile(reg[2:]).FindStringIndex(field) != nil {
				return false
			}
		} else {
			if regexp.MustCompile(reg).FindStringIndex(field) == nil {
				log.Printf("not find regexps(%v) in field(%v)\n", reg, field)
				return false
			}
		}
	}
	return true
}

func checkCache(filepath string) (*os.File, error) {
	fp, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0600)
	if nil != err {
		return nil, err
	}

	err = syscall.Flock(int(fp.Fd()), syscall.LOCK_SH)
	if nil != err {
		return nil, err
	}

	fileinfo, err := fp.Stat()
	if nil != err {
		return nil, err
	}

	log.Println("fileinfo.Modtime:", fileinfo.ModTime().Unix())
	log.Println("time.Mow.Unix", time.Now().Unix())
	log.Println("fileinfo.Size", fileinfo.Size())

	if fileinfo.Size() > 0 && (fileinfo.ModTime().Unix()+int64(*pollTime/2)) > time.Now().Unix() {
		log.Println("Using the cache file")
		return nil, nil
	} else {
		log.Println("The cache file seems too small or stale")
		// Escalate the lock to exclusive, so we can write to it.
		err = syscall.Flock(int(fp.Fd()), syscall.LOCK_EX)
		if nil != err {
			return nil, err
		}

		// We might have blocked while waiting for that LOCK_EX, and
		// another process ran and updated it.  Let's see if we can just
		// return the data now:
		if fileinfo.Size() > 0 && (fileinfo.ModTime().Unix()+int64(*pollTime/2)) > time.Now().Unix() {
			log.Println("Using the cache file")
			return nil, nil
		}
		fp.Truncate(0)
	}
	return fp, nil

}

func stringMapAdd(m1 map[string]string, m2 map[string]string) {
	for k, v := range m2 {
		if _, ok := m1[k]; ok {
			log.Fatal("key conflict:", k)
		}
		m1[k] = v
	}
}

func intMapAdd(m1 map[string]string, m2 map[string]int64) {
	for k, v := range m2 {
		if _, ok := m1[k]; ok {
			log.Fatal("key conflict:", k)
		}
		m1[k] = strconv.FormatInt(v, 10)
	}
}

func changeKeyCase(m map[string]string) map[string]string {
	lowerMap := make(map[string]string)
	for k, v := range m {
		lowerMap[strings.ToLower(k)] = v
	}
	return lowerMap
}

func discoveryMysqldPort() {
	data := make([]map[string]string, 0)
	enc := json.NewEncoder(os.Stdout)
	cmd := "netstat -ntlp |awk '/\\/mysqld[ $]/{print $4}'|awk -F ':' '{print $NF}'"
	if *useSudo {
		cmd = "sudo " + cmd
	}
	log.Println("discoveryMysqldPort:find mysql port cmd:", cmd)
	out, err := exec.Command("sh", "-c", cmd).Output()
	log.Println("discoveryMysqldPort:cmd out:", string(out))
	if nil != err {
		log.Println("discoveryMysqldPort err:", err)
	}
	fields := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	for _, field := range fields {
		if "" == field {
			continue
		}
		mp := map[string]string{
			"{#MYSQLPORT}": field,
		}
		data = append(data, mp)
	}

	formatData := map[string][]map[string]string{
		"data": data,
	}
	enc.Encode(formatData)
}

func verifyMysqldPort(port string) bool {
	cmd := "netstat -ntlp |awk -F '[ ]+|/' '$4~/:" + port + "$/{print $8}'"
	if *useSudo {
		cmd = "sudo " + cmd
	}
	log.Println("verifyMysqldPort:find port Program name cmd:", cmd)
	out, err := exec.Command("sh", "-c", cmd).Output()
	log.Println("verifyMysqldPort:cmd out:", string(out))
	if nil != err {
		log.Println("verifyMysqldPort err:", err)
	}
	if string(out) == "mysqld\n" {
		return true
	}
	return false
}
