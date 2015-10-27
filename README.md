#actiontech zabbix mysql monitor

percona monitoring plugins zabbix的go语言版本

## 1. item取值与percona版本差异

 - innodb_transactions  
取值方法：SHOW /\*!50000 ENGINE*/ INNODB STATUS,可得行Trx id counter 861144，取值861144  
percona将861144作为十六进制字符解析，actiontech版本将861144作为十进制字符解析
 - unpurged_txns：  
取值方法：SHOW /\*!50000 ENGINE*/ INNODB STATUS,可得行Purge done for trx's n:o < 861135 undo n:o < 0，取值861135  
percona将861135作为十六进制字符解析，actiontech版本将861135作为十进制字符解析 
 - running_slave，slave_lag:  
取值方法：如果SHOW SLAVE STATUS为空，认为该mysql为master，设置running_slave=1，slave_lag=0；如果SHOW SLAVE STATUS不为空，与percona处理相同，依据slave_io_running及slave_sql_running等具体参数值设置 

## 2. trigger与percona版本差异
- 增加max-connections less than 4999
- 增加Open-files-limit less than 65534

## 3. 限制
- 暂不支持mysql ssl连接方式
- 不完全支持Percona Server or MariaDB

## 4. 安装
1. 安装zabbix-agent  
2. 执行install.sh  
//install.sh作用仅为拷贝文件至默认路径，可调整  
3. 创建采集mysql状态信息的用户  
`mysql> grant process,select,replication client on *.* to zbx@’127.0.0.1’ identified by 'zabbix';`   
4. 拷贝  
`cp /var/lib/zabbix/actiontech/templates/userparameter_actiontech_mysql.conf /etc/zabbix/zabbix_agentd.d/`  
//若改变登陆密码或默认路径，请相应调整userparameter_actiontech_mysql.conf  
重启agent  
`service zabbix-agent restart`  
5. 测试数据采集  
`sudo -u zabbix -H /var/lib/zabbix/actiontech/scripts/actiontech_mysql_monitor --host 127.0.0.1 --user zbx --pass zabbix --items max_connections`   
//被监控mysql最大连接数  
`sudo -u zabbix -H /var/lib/zabbix/actiontech/scripts/actiontech_mysql_monitor --host 127.0.0.1 --user zbx --pass zabbix --items running_slave`  
//主从复制状态  
6. zabbix server导入配置模板actiontech_zabbix_agent_template_percona_mysql_server.xml， 添加主机、模板，开始监控  


## 5. 文件列表

|文件名|对应percona版本文件名|
|------|------|
|无|get_mysql_stats_wrapper.sh|
|actiontech_mysql_monitor|ss_get_mysql_stats.php|
|userparameter_actiontech_mysql.conf|userparameter_percona_mysql.conf|
|actiontech_zabbix_agent_template_percona_mysql_server.xml|zabbix_agent_template_percona_mysql_server_ht_2.0.9-sver1.1.5.xml|
|install.sh|无|


- actiontech_mysql_monitor  
默认路径：`/var/lib/zabbix/actiontech/scripts/actiontech_mysql_monitor`  
备注：二进制可执行文件，可单独使用（actiontech_mysql_monitor --help）
- userparameter_actiontech_mysql.conf  
默认路径：`/var/lib/zabbix/actiontech/templates/userparameter_actiontech_mysql.conf`  
备注：zabbix-agent的key配置文件  
- actiontech_zabbix_agent_template_percona_mysql_server.xml  
默认路径：`/var/lib/zabbix/actiontech/templates/actiontech_zabbix_agent_template_percona_mysql_server.xml`  
备注：模板文件，可导入zabbix  




## 6. 所有item及其出处列表

| item名         | 相关出处 |
| ------------- | -----| ------ |
|		Key_read_requests|          SHOW /*!50002 GLOBAL */ STATUS|
|		Key_reads|                  SHOW /*!50002 GLOBAL */ STATUS|
|		Key_write_requests|         SHOW /*!50002 GLOBAL */ STATUS|
|		Key_writes|                 SHOW /*!50002 GLOBAL */ STATUS|
|		history_list|               SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_transactions|        SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		read_views|                 SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		current_transactions|       SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		locked_transactions|        SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		active_transactions|        SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pool_size|                  Innodb_buffer_pool_pages_total or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		free_pages|                 Innodb_buffer_pool_pages_free or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		database_pages|            Innodb_buffer_pool_pages_data or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		modified_pages|             Innodb_buffer_pool_pages_dirty or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pages_read|                 pages_read or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pages_created|              Innodb_pages_created or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pages_written|              pages_written or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		file_fsyncs|                Innodb_data_fsyncs or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		file_reads|                 SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		file_writes|                SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		log_writes|                 SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_aio_log_ios|        SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_aio_sync_ios|       SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_buf_pool_flushes|   SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_chkp_writes|        SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_ibuf_aio_reads|     SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_log_flushes|        Innodb_os_log_pending_fsyncs or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_log_writes|         SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_normal_aio_reads|   Innodb_data_pending_reads or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		pending_normal_aio_writes|  Innodb_data_pending_writes or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_inserts|               SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_merged|                SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_merges|                SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		spin_waits|                 SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		spin_rounds|                SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		os_waits|                   SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		rows_inserted|              Innodb_rows_inserted or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		rows_updated|               Innodb_rows_updated or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		rows_deleted|               Innodb_rows_deleted or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		rows_read|                  Innodb_rows_read or SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		Table_locks_waited|         SHOW /*!50002 GLOBAL */ STATUS|
|		Table_locks_immediate|      SHOW /*!50002 GLOBAL */ STATUS|
|		Slow_queries|               SHOW /*!50002 GLOBAL */ STATUS|
|		Open_files|                 SHOW /*!50002 GLOBAL */ STATUS|
|		Open_tables|                SHOW /*!50002 GLOBAL */ STATUS|
|		Opened_tables|              SHOW /*!50002 GLOBAL */ STATUS|
|		innodb_open_files|          SHOW VARIABLES|
|		open_files_limit|           SHOW VARIABLES|
|		table_cache|                SHOW VARIABLES|
|		Aborted_clients|            SHOW /*!50002 GLOBAL */ STATUS|
|		Aborted_connects|           SHOW /*!50002 GLOBAL */ STATUS|
|		Max_used_connections|       SHOW /*!50002 GLOBAL */ STATUS|
|		Slow_launch_threads|        SHOW /*!50002 GLOBAL */ STATUS|
|		Threads_cached|             SHOW /*!50002 GLOBAL */ STATUS|
|		Threads_connected|          SHOW /*!50002 GLOBAL */ STATUS|
|		Threads_created|            SHOW /*!50002 GLOBAL */ STATUS|
|		Threads_running|            SHOW /*!50002 GLOBAL */ STATUS|
|		max_connections|            SHOW VARIABLES|
|		thread_cache_size|          SHOW VARIABLES|
|		Connections|                SHOW /*!50002 GLOBAL */ STATUS|
|		slave_running|              slave_lag or 0 |
|		slave_stopped|              slave_lag or 0|
|		Slave_retried_transactions| SHOW /*!50002 GLOBAL */ STATUS|
|		slave_lag|                  SHOW SLAVE STATUS or percona.heartbeat|
|		Slave_open_temp_tables|     SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_free_blocks|         SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_free_memory|         SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_hits|                SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_inserts|             SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_lowmem_prunes|       SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_not_cached|          SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_queries_in_cache|    SHOW /*!50002 GLOBAL */ STATUS|
|		Qcache_total_blocks|        SHOW /*!50002 GLOBAL */ STATUS|
|		query_cache_size|           SHOW VARIABLES|
|		Questions|                  SHOW /*!50002 GLOBAL */ STATUS|
|		Com_update|                 SHOW /*!50002 GLOBAL */ STATUS|
|		Com_insert|                 SHOW /*!50002 GLOBAL */ STATUS|
|		Com_select|                 SHOW /*!50002 GLOBAL */ STATUS|
|		Com_delete|                 SHOW /*!50002 GLOBAL */ STATUS|
|		Com_replace|                SHOW /*!50002 GLOBAL */ STATUS|
|		Com_load|                   SHOW /*!50002 GLOBAL */ STATUS|
|		Com_update_multi|           SHOW /*!50002 GLOBAL */ STATUS|
|		Com_insert_select|          SHOW /*!50002 GLOBAL */ STATUS|
|		Com_delete_multi|           SHOW /*!50002 GLOBAL */ STATUS|
|		Com_replace_select|         SHOW /*!50002 GLOBAL */ STATUS|
|		Select_full_join|           SHOW /*!50002 GLOBAL */ STATUS|
|		Select_full_range_join|     SHOW /*!50002 GLOBAL */ STATUS|
|		Select_range|               SHOW /*!50002 GLOBAL */ STATUS|
|		Select_range_check|         SHOW /*!50002 GLOBAL */ STATUS|
|		Select_scan|                SHOW /*!50002 GLOBAL */ STATUS|
|		Sort_merge_passes|          SHOW /*!50002 GLOBAL */ STATUS|
|		Sort_range|                 SHOW /*!50002 GLOBAL */ STATUS|
|		Sort_rows|                  SHOW /*!50002 GLOBAL */ STATUS|
|		Sort_scan|                  SHOW /*!50002 GLOBAL */ STATUS|
|		Created_tmp_tables|         SHOW /*!50002 GLOBAL */ STATUS|
|		Created_tmp_disk_tables|    SHOW /*!50002 GLOBAL */ STATUS|
|		Created_tmp_files|          SHOW /*!50002 GLOBAL */ STATUS|
|		Bytes_sent|                 SHOW /*!50002 GLOBAL */ STATUS|
|		Bytes_received|             SHOW /*!50002 GLOBAL */ STATUS|
|		innodb_log_buffer_size|     SHOW VARIABLES|
|		unflushed_log|              log_bytes_written-log_bytes_flushed or innodb_log_buffer_size|
|		log_bytes_flushed|          SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		log_bytes_written|          SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		relay_log_space|            SHOW SLAVE STATUS|
|		binlog_cache_size|          SHOW VARIABLES|
|		Binlog_cache_disk_use|      SHOW /*!50002 GLOBAL */ STATUS|
|		Binlog_cache_use|           SHOW /*!50002 GLOBAL */ STATUS|
|		binary_log_space|           SHOW MASTER LOGS|
|		innodb_locked_tables|       SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_lock_structs|        SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		State_closing_tables|       SHOW PROCESSLIST|
|		State_copying_to_tmp_table| SHOW PROCESSLIST|
|		State_end|                  SHOW PROCESSLIST|
|		State_freeing_items|        SHOW PROCESSLIST|
|		State_init|                 SHOW PROCESSLIST|
|		State_locked|               SHOW PROCESSLIST|
|		State_login|                SHOW PROCESSLIST|
|		State_preparing|            SHOW PROCESSLIST|
|		State_reading_from_net|     SHOW PROCESSLIST|
|		State_sending_data|         SHOW PROCESSLIST|
|		State_sorting_result|       SHOW PROCESSLIST|
|		State_statistics|           SHOW PROCESSLIST|
|		State_updating|             SHOW PROCESSLIST|
|		State_writing_to_net|       SHOW PROCESSLIST|
|		State_none|                 SHOW PROCESSLIST|
|		State_other|                SHOW PROCESSLIST|
|		Handler_commit|             SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_delete|             SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_discover|           SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_prepare|            SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_first|         SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_key|           SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_next|          SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_prev|          SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_rnd|           SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_read_rnd_next|      SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_rollback|           SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_savepoint|          SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_savepoint_rollback| SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_update|             SHOW /*!50002 GLOBAL */ STATUS|
|		Handler_write|              SHOW /*!50002 GLOBAL */ STATUS|
|		innodb_tables_in_use|       SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_lock_wait_secs|      SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		hash_index_cells_total|     SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		hash_index_cells_used|      SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		total_mem_alloc|            SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		additional_pool_alloc|      SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		uncheckpointed_bytes|       log_bytes_written-last_checkpoint|
|		ibuf_used_cells|            SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_free_cells|            SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		ibuf_cell_count|            SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		adaptive_hash_memory|       SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		page_hash_memory|           SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		dictionary_cache_memory|    SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		file_system_memory|         SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		lock_system_memory|         SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		recovery_system_memory|     SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		thread_hash_memory|         SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_sem_waits|           SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		innodb_sem_wait_time_ms|    SHOW /\*!50000 ENGINE*/ INNODB STATUS|
|		Key_buf_bytes_unflushed|    key_cache_block_size*Key_blocks_not_flushed|
|		Key_buf_bytes_used|         key_buffer_size-(Key_blocks_unused*key_cache_block_size)|
|		key_buffer_size|            SHOW VARIABLES|
|		Innodb_row_lock_time|       SHOW /*!50002 GLOBAL */ STATUS|
|		Innodb_row_lock_waits|      SHOW /*!50002 GLOBAL */ STATUS|
|		Query_time_count_00|        Percona Server or MariaDB|
|		Query_time_count_01|        Percona Server or MariaDB|
|		Query_time_count_02|        Percona Server or MariaDB|
|		Query_time_count_03|        Percona Server or MariaDB|
|		Query_time_count_04|        Percona Server or MariaDB|
|		Query_time_count_05|        Percona Server or MariaDB|
|		Query_time_count_06|        Percona Server or MariaDB|
|		Query_time_count_07|        Percona Server or MariaDB|
|		Query_time_count_08|        Percona Server or MariaDB|
|		Query_time_count_09|        Percona Server or MariaDB|
|		Query_time_count_10|        Percona Server or MariaDB|
|		Query_time_count_11|        Percona Server or MariaDB|
|		Query_time_count_12|        Percona Server or MariaDB|
|		Query_time_count_13|        Percona Server or MariaDB|
|		Query_time_total_00|        Percona Server or MariaDB|
|		Query_time_total_01|        Percona Server or MariaDB|
|		Query_time_total_02|        Percona Server or MariaDB|
|		Query_time_total_03|        Percona Server or MariaDB|
|		Query_time_total_04|        Percona Server or MariaDB|
|		Query_time_total_05|        Percona Server or MariaDB|
|		Query_time_total_06|        Percona Server or MariaDB|
|		Query_time_total_07|        Percona Server or MariaDB|
|		Query_time_total_08|        Percona Server or MariaDB|
|		Query_time_total_09|        Percona Server or MariaDB|
|		Query_time_total_10|        Percona Server or MariaDB|
|		Query_time_total_11|        Percona Server or MariaDB|
|		Query_time_total_12|        Percona Server or MariaDB|
|		Query_time_total_13|        Percona Server or MariaDB|
|		wsrep_replicated_bytes|     Percona Server or MariaDB|
|		wsrep_received_bytes|       Percona Server or MariaDB|
|		wsrep_replicated|           Percona Server or MariaDB|
|		wsrep_received|             Percona Server or MariaDB|
|		wsrep_local_cert_failures|  Percona Server or MariaDB|
|		wsrep_local_bf_aborts|      Percona Server or MariaDB|
|		wsrep_local_send_queue|     Percona Server or MariaDB|
|		wsrep_local_recv_queue|     Percona Server or MariaDB|
|		wsrep_cluster_size|         Percona Server or MariaDB|
|		wsrep_cert_deps_distance|   Percona Server or MariaDB|
|		wsrep_apply_window|         Percona Server or MariaDB|
|		wsrep_commit_window|        Percona Server or MariaDB|
|		wsrep_flow_control_paused|  Percona Server or MariaDB|
|		wsrep_flow_control_sent|    Percona Server or MariaDB|
|		wsrep_flow_control_recv|    Percona Server or MariaDB|
|		pool_reads|                 Innodb_buffer_pool_reads|
|		pool_read_requests|         Innodb_buffer_pool_read_requests|
|		running_slave|              SHOW SLAVE STATUS|



## 7. trigger列表
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


## 8. issues
- 欢迎在issues中提出任何使用问题
