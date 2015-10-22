#actiontech zabbix mysql monitor

percona monitoring plugins zabbix的go语言版本

## 1. 文件列表

|文件名|对应percona版本文件名|
|------|------|
|actiontech_get_mysql_stats_wrapper.sh|get_mysql_stats_wrapper.sh|
|actiontech_mysql_monitor|ss_get_mysql_stats.php|
|userparameter_actiontech_mysql.conf|userparameter_percona_mysql.conf|
|actiontech_zabbix_agent_template_percona_mysql_server.xml|zabbix_agent_template_percona_mysql_server_ht_2.0.9-sver1.1.5.xml|
|install.sh|无|

- actiontech_get_mysql_stats_wrapper.sh 
默认路径：/var/lib/zabbix/actiontech/scripts/actiontech_get_mysql_stats_wrapper.sh
备注：辅助脚本，可修改mysql用户名，密码等配置
- actiontech_mysql_monitor
默认路径：/var/lib/zabbix/actiontech/scripts/actiontech_mysql_monitor
备注：二进制可执行文件，路径保持与actiontech_get_mysql_stats_wrapper.sh 一致即可，可单独使用（actiontech_mysql_monitor --help）
- userparameter_actiontech_mysql.conf
默认路径：/var/lib/zabbix/actiontech/templates/userparameter_actiontech_mysql.conf
备注：zabbix-agent的key配置文件
- actiontech_zabbix_agent_template_percona_mysql_server.xml
默认路径：/var/lib/zabbix/actiontech/templates/actiontech_zabbix_agent_template_percona_mysql_server.xml
备注：模板文件，可导入zabbix

## 2. 安装
1. 安装zabbix-agent
2. 解压打开tar包，执行install.sh
//install.sh作用仅为拷贝文件至默认路径，可调整
3. 创建采集mysql状态信息的用户
mysql> grant process,select,replication client on *.* to zbx@’127.0.0.1’ identified by 'zabbix';
配置用户免密码登陆
vi ~zabbix/.my.cnf
[client]
user = zbx
password = zabbix
host = 127.0.0.1
4. 修改配置
vi /var/lib/zabbix/actiontech/scripts/actiontech_get_mysql_stats_wrapper.sh
//CMD="\$DIR/actiontech_mysql_monitor --host $HOST --user=zbx --pass=zabbix"
5. 拷贝
cp /var/lib/zabbix/actiontech/templates/userparameter_actiontech_mysql.conf /etc/zabbix/zabbix_agentd.d/
//若未使用默认路径，请相应调整userparameter_actiontech_mysql.conf
重启agent
service zabbix-agent restart
6. 测试数据采集
/var/lib/zabbix/actiontech/scripts/actiontech_get_mysql_stats_wrapper.sh ix
//被监控mysql最大连接数
sudo -u zabbix -H /var/lib/zabbix/actiontech/scripts/actiontech_get_mysql_stats_wrapper.sh running-slave 
//主从复制状态
7. zabbix server导入配置模板actiontech_zabbix_agent_template_percona_mysql_server.xml， 添加主机、模板，开始监控



## 3. 所有item及其查询代码列表
可使用方法举例：
/var/lib/zabbix/actiontech/scripts/actiontech_get_mysql_stats_wrapper.sh gg 或
./actiontech_mysql_monitor -host=127.0.0.1 -user=zbx -pass=zabbix -items=gg

| item名        |查询代码   | 相关出处 |
| ------------- | -----| ------ |
|		Key_read_requests|          gg|SHOW /*!50002 GLOBAL */ STATUS|
|		Key_reads|                  gh|SHOW /*!50002 GLOBAL */ STATUS|
|		Key_write_requests|         gi|SHOW /*!50002 GLOBAL */ STATUS|
|		Key_writes|                 gj|SHOW /*!50002 GLOBAL */ STATUS|
|		history_list|               gk|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_transactions|        gl|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		read_views|                 gm|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		current_transactions|       gn|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		locked_transactions|        go|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		active_transactions|        gp|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pool_size|                  gq|Innodb_buffer_pool_pages_total or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		free_pages|                 gr|Innodb_buffer_pool_pages_free or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		database_pages|             gs|Innodb_buffer_pool_pages_data or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		modified_pages|             gt|Innodb_buffer_pool_pages_dirty or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pages_read|                 gu|pages_read or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pages_created|              gv|Innodb_pages_created or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pages_written|              gw|pages_written or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		file_fsyncs|                gx|Innodb_data_fsyncs or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		file_reads|                 gy|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		file_writes|                gz|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		log_writes|                 hg|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_aio_log_ios|        hh|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_aio_sync_ios|       hi|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_buf_pool_flushes|   hj|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_chkp_writes|        hk|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_ibuf_aio_reads|     hl|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_log_flushes|        hm|Innodb_os_log_pending_fsyncs or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_log_writes|         hn|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_normal_aio_reads|   ho|Innodb_data_pending_reads or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_normal_aio_writes|  hp|Innodb_data_pending_writes or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_inserts|               hq|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_merged|                hr|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_merges|                hs|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		spin_waits|                 ht|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		spin_rounds|                hu|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		os_waits|                   hv|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		rows_inserted|              hw|Innodb_rows_inserted or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		rows_updated|               hx|Innodb_rows_updated or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		rows_deleted|               hy|Innodb_rows_deleted or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		rows_read|                  hz|Innodb_rows_read or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		Table_locks_waited|         ig|SHOW /*!50002 GLOBAL */ STATUS|
|		Table_locks_immediate|      ih|SHOW /*!50002 GLOBAL */ STATUS|
|		Slow_queries|               ii|SHOW /*!50002 GLOBAL */ STATUS|
|		Open_files|                 ij|SHOW /*!50002 GLOBAL */ STATUS|
|		Open_tables|                ik|SHOW /*!50002 GLOBAL */ STATUS|
|		Opened_tables|              il|SHOW /*!50002 GLOBAL */ STATUS|
|		innodb_open_files|          im|SHOW VARIABLES|
|		open_files_limit|           in|SHOW VARIABLES|
|		table_cache|                io|SHOW VARIABLES|
|		Aborted_clients|            ip|SHOW /*!50002 GLOBAL */ STATUS|
|		Aborted_connects|           iq|SHOW /*!50002 GLOBAL */ STATUS|
|		Max_used_connections|       ir|SHOW /*!50002 GLOBAL */ STATUS|
|		Slow_launch_threads|        is|SHOW /*!50002 GLOBAL */ STATUS|
|		Threads_cached|             it|SHOW /*!50002 GLOBAL */ STATUS|
|		Threads_connected|          iu|SHOW /*!50002 GLOBAL */ STATUS|
|		Threads_created|            iv|SHOW /*!50002 GLOBAL */ STATUS|
|		Threads_running|            iw|SHOW /*!50002 GLOBAL */ STATUS|
|		max_connections|            ix|SHOW VARIABLES|
|		thread_cache_size|          iy|SHOW VARIABLES|
|		Connections|                iz|SHOW /*!50002 GLOBAL */ STATUS|
|		slave_running|              jg| slave_lag or 0 |
|		slave_stopped|              jh| slave_log or 0
|		Slave_retried_transactions| ji|SHOW /*!50002 GLOBAL */ STATUS|
|		slave_lag|                  jj|SHOW SLAVE STATUS or percona.heartbeat|
|		Slave_open_temp_tables|     jk|SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_free_blocks|         jl|SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_free_memory|         jm|SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_hits|                jn|SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_inserts|             jo|SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_lowmem_prunes|       jp|SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_not_cached|          jq|SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_queries_in_cache|    jr|SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_total_blocks|        js|SHOW /*!50002 GLOBAL */ STATUS|
|		query_cache_size|           jt|SHOW VARIABLES|
|		Questions|                  ju|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_update|                 jv|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_insert|                 jw|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_select|                 jx|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_delete|                 jy|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_replace|                jz|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_load|                   kg|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_update_multi|           kh|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_insert_select|          ki|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_delete_multi|           kj|SHOW /*!50002 GLOBAL */ STATUS|
|		Com_replace_select|         kk|SHOW /*!50002 GLOBAL */ STATUS|
|		Select_full_join|           kl|SHOW /*!50002 GLOBAL */ STATUS|
|		Select_full_range_join|     km|SHOW /*!50002 GLOBAL */ STATUS|
|		Select_range|               kn|SHOW /*!50002 GLOBAL */ STATUS|
|		Select_range_check|         ko|SHOW /*!50002 GLOBAL */ STATUS|
|		Select_scan|                kp|SHOW /*!50002 GLOBAL */ STATUS|
|		Sort_merge_passes|          kq|SHOW /*!50002 GLOBAL */ STATUS|
|		Sort_range|                 kr|SHOW /*!50002 GLOBAL */ STATUS|
|		Sort_rows|                  ks|SHOW /*!50002 GLOBAL */ STATUS|
|		Sort_scan|                  kt|SHOW /*!50002 GLOBAL */ STATUS|
|		Created_tmp_tables|         ku|SHOW /*!50002 GLOBAL */ STATUS|
|		Created_tmp_disk_tables|    kv|SHOW /*!50002 GLOBAL */ STATUS|
|		Created_tmp_files|          kw|SHOW /*!50002 GLOBAL */ STATUS|
|		Bytes_sent|                 kx|SHOW /*!50002 GLOBAL */ STATUS|
|		Bytes_received|             ky|SHOW /*!50002 GLOBAL */ STATUS|
|		innodb_log_buffer_size|     kz|SHOW VARIABLES|
|		unflushed_log|              lg|log_bytes_written-log_bytes_flushed or innodb_log_buffer_size|
|		log_bytes_flushed|          lh|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		log_bytes_written|          li|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		relay_log_space|            lj|SHOW SLAVE STATUS|
|		binlog_cache_size|          lk|SHOW VARIABLES|
|		Binlog_cache_disk_use|      ll|SHOW /*!50002 GLOBAL */ STATUS|
|		Binlog_cache_use|           lm|SHOW /*!50002 GLOBAL */ STATUS|
|		binary_log_space|           ln|SHOW MASTER LOGS|
|		innodb_locked_tables|       lo|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_lock_structs|        lp|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		State_closing_tables|       lq|SHOW PROCESSLIST|
|		State_copying_to_tmp_table| lr|SHOW PROCESSLIST|
|		State_end|                  ls|SHOW PROCESSLIST|
|		State_freeing_items|        lt|SHOW PROCESSLIST|
|		State_init|                 lu|SHOW PROCESSLIST|
|		State_locked|               lv|SHOW PROCESSLIST|
|		State_login|                lw|SHOW PROCESSLIST|
|		State_preparing|            lx|SHOW PROCESSLIST|
|		State_reading_from_net|     ly|SHOW PROCESSLIST|
|		State_sending_data|         lz|SHOW PROCESSLIST|
|		State_sorting_result|       mg|SHOW PROCESSLIST|
|		State_statistics|           mh|SHOW PROCESSLIST|
|		State_updating|             mi|SHOW PROCESSLIST|
|		State_writing_to_net|       mj|SHOW PROCESSLIST|
|		State_none|                 mk|SHOW PROCESSLIST|
|		State_other|                ml|SHOW PROCESSLIST|
|		Handler_commit|             mm|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_delete|             mn|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_discover|           mo|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_prepare|            mp|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_first|         mq|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_key|           mr|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_next|          ms|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_prev|          mt|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_rnd|           mu|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_rnd_next|      mv|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_rollback|           mw|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_savepoint|          mx|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_savepoint_rollback| my|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_update|             mz|SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_write|              ng|SHOW /*!50002 GLOBAL */ STATUS|
|		innodb_tables_in_use|       nh|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_lock_wait_secs|      ni|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		hash_index_cells_total|     nj|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		hash_index_cells_used|      nk|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		total_mem_alloc|            nl|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		additional_pool_alloc|      nm|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		uncheckpointed_bytes|       nn|log_bytes_written-last_checkpoint|
|		ibuf_used_cells|            no|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_free_cells|            np|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_cell_count|            nq|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		adaptive_hash_memory|       nr|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		page_hash_memory|           ns|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		dictionary_cache_memory|    nt|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		file_system_memory|         nu|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		lock_system_memory|         nv|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		recovery_system_memory|     nw|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		thread_hash_memory|         nx|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_sem_waits|           ny|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_sem_wait_time_ms|    nz|SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		Key_buf_bytes_unflushed|    og|key_cache_block_size*Key_blocks_not_flushed|
|		Key_buf_bytes_used|         oh|key_buffer_size-(Key_blocks_unused*key_cache_block_size)|
|		key_buffer_size|            oi|SHOW VARIABLES|
|		Innodb_row_lock_time|       oj|SHOW /*!50002 GLOBAL */ STATUS|
|		Innodb_row_lock_waits|      ok|SHOW /*!50002 GLOBAL */ STATUS|
|		Query_time_count_00|        ol|Percona Server or MariaDB|
|		Query_time_count_01|        om|Percona Server or MariaDB|
|		Query_time_count_02|        on|Percona Server or MariaDB|
|		Query_time_count_03|        oo|Percona Server or MariaDB|
|		Query_time_count_04|        op|Percona Server or MariaDB|
|		Query_time_count_05|        oq|Percona Server or MariaDB|
|		Query_time_count_06|        or|Percona Server or MariaDB|
|		Query_time_count_07|        os|Percona Server or MariaDB|
|		Query_time_count_08|        ot|Percona Server or MariaDB|
|		Query_time_count_09|        ou|Percona Server or MariaDB|
|		Query_time_count_10|        ov|Percona Server or MariaDB|
|		Query_time_count_11|        ow|Percona Server or MariaDB|
|		Query_time_count_12|        ox|Percona Server or MariaDB|
|		Query_time_count_13|        oy|Percona Server or MariaDB|
|		Query_time_total_00|        oz|Percona Server or MariaDB|
|		Query_time_total_01|        pg|Percona Server or MariaDB|
|		Query_time_total_02|        ph|Percona Server or MariaDB|
|		Query_time_total_03|        pi|Percona Server or MariaDB|
|		Query_time_total_04|        pj|Percona Server or MariaDB|
|		Query_time_total_05|        pk|Percona Server or MariaDB|
|		Query_time_total_06|        pl|Percona Server or MariaDB|
|		Query_time_total_07|        pm|Percona Server or MariaDB|
|		Query_time_total_08|        pn|Percona Server or MariaDB|
|		Query_time_total_09|        po|Percona Server or MariaDB|
|		Query_time_total_10|        pp|Percona Server or MariaDB|
|		Query_time_total_11|        pq|Percona Server or MariaDB|
|		Query_time_total_12|        pr|Percona Server or MariaDB|
|		Query_time_total_13|        ps|Percona Server or MariaDB|
|		wsrep_replicated_bytes|     pt|Percona Server or MariaDB|
|		wsrep_received_bytes|       pu|Percona Server or MariaDB|
|		wsrep_replicated|           pv|Percona Server or MariaDB|
|		wsrep_received|             pw|Percona Server or MariaDB|
|		wsrep_local_cert_failures|  px|Percona Server or MariaDB|
|		wsrep_local_bf_aborts|      py|Percona Server or MariaDB|
|		wsrep_local_send_queue|     pz|Percona Server or MariaDB|
|		wsrep_local_recv_queue|     qg|Percona Server or MariaDB|
|		wsrep_cluster_size|         qh|Percona Server or MariaDB|
|		wsrep_cert_deps_distance|   qi|Percona Server or MariaDB|
|		wsrep_apply_window|         qj|Percona Server or MariaDB|
|		wsrep_commit_window|        qk|Percona Server or MariaDB|
|		wsrep_flow_control_paused|  ql|Percona Server or MariaDB|
|		wsrep_flow_control_sent|    qm|Percona Server or MariaDB|
|		wsrep_flow_control_recv|    qn|Percona Server or MariaDB|
|		pool_reads|                 qo|Innodb_buffer_pool_reads|
|		pool_read_requests|         qp|Innodb_buffer_pool_read_requests|
|		running_slave|              running-slave|SHOW SLAVE STATUS|

## 4. item取值与percona版本差异

 - innodb_transactions
取值方法：SHOW /\*!50000 ENGINE*/ INNODB STATUS,可得行Trx id counter 861144，取值861144
percona将861144作为十六进制字符解析，actiontech版本将861144作为十进制字符解析
 - unpurged_txns：
取值方法：SHOW /\*!50000 ENGINE*/ INNODB STATUS,可得行Purge done for trx's n:o < 861135 undo n:o < 0，取值861135
percona将861135作为十六进制字符解析，actiontech版本将861135作为十进制字符解析

## 5. trigger列表
注：以下列表未列出trigger依赖，详情可见actiontech_zabbix_agent_template_percona_mysql_server.xml

| trigger名        |触发器表达式   |
| ------------- | -----| 
|MySQL is down on {HOST.NAME}|{proc.num[mysqld].last(0)}=0|
|MySQL connections utilization more than 95% on {HOST.NAME}|{MySQL.Threads-connected.last(0)}/{MySQL.max-connections.last(0)}>0.95|
|MySQL connections utilization more than 80% on {HOST.NAME}|{MySQL.Threads-connected.last(0)}/{MySQL.max-connections.last(0)}>0.8|
|MySQL active threads more than 100 on {HOST.NAME}|{MySQL.Threads-running.last(0)}>100|
|MySQL active threads more than 40 on {HOST.NAME}|{MySQL.Threads-running.last(0)}>40|
|MySQL slave lag more than 600 on {HOST.NAME}|{MySQL.slave-lag.last(0)}>600|
|MySQL slave lag more than 300 on {HOST.NAME}|{MySQL.slave-lag.last(0)}>300|
|Slave is stopped on {HOST.NAME}|{MySQL.running-slave.last(0)}=0|
|MySQL max-connections less than 4999 on {HOST.NAME}|{MySQL.max-connections.last(0)}<4999|
|MySQL Open-files-limit less than 65534 on {HOST.NAME}|{MySQL.Open-files-limit.last(0)}<65534|

## 6. trigger与percona版本差异
- 增加max-connections less than 4999
- 增加Open-files-limit less than 65534

## 7. 限制
- 暂不支持mysql ssl连接方式
- 不完全支持Percona Server or MariaDB

