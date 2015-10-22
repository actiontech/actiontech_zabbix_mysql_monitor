package main

import (
	"testing"
)

func TestParseInnodbStatusWithRule(t *testing.T) {
	str := `Mutex spin waits 79626940, rounds 157459864, OS waits 698719
	Mutex spin waits 0, rounds 247280272495, OS waits 316513438
	RW-shared spins 3859028, OS waits 2100750; RW-excl spins 4641946, OS waits 1530310
	RW-shared spins 604733, rounds 8107431, OS waits 241268
	RW-excl spins 604733, rounds 8107431, OS waits 241268
	--Thread 907205 has waited at handler/ha_innodb.cc line 7156 for 1.00 seconds the semaphore:
	Trx id counter 0 1170664159
	Purge done for trx's n:o < 0 1170663853 undo n:o < 0 0
	History list length 132
	---TRANSACTION 0, not started, process no 13510, OS thread id 1170446656
	---TRANSACTION 0 42313619, ACTIVE 49 sec, process no 10099, OS thread id 3771312 starting index read
	------- TRX HAS BEEN WAITING 32 SEC FOR THIS LOCK TO BE GRANTED:
	1 read views open inside InnoDB
	mysql tables in use 2, locked 2
	23 lock struct(s), heap size 3024, undo log entries 27
	LOCK WAIT 12 lock struct(s), heap size 3024, undo log entries 5
	LOCK WAIT 2 lock struct(s), heap size 368
	8782182 OS file reads, 15635445 OS file writes, 947800 OS fsyncs
	Pending normal aio reads: 0, aio writes: 0,
	ibuf aio reads: 0, log i/o's: 0, sync i/o's: 0
	Pending flushes (fsync) log: 0; buffer pool: 0
	Ibuf for space 0: size 1, free list len 887, seg size 889, is not empty
	Ibuf for space 0: size 1, free list len 887, seg size 889,
	merged operations:
	insert 593983, delete mark 387006, delete 73092
	Hash table size 4425293, used cells 4229064, ....
	3430041 log i/o's done, 17.44 log i/o's/second
	0 pending log writes, 0 pending chkp writes
	Log sequence number 125 3934414864
	Log flushed up to   125 3934414864
	Last checkpoint at  125 3934293461
	Total memory allocated 29642194944; in additional pool allocated 0
	Total memory allocated by read views 96
	Adaptive hash index 1538240664 	(186998824 + 1351241840)
	Page hash           11688584
	Dictionary cache    145525560 	(140250984 + 5274576)
	File system         313848 	(82672 + 231176)
	Lock system         29232616 	(29219368 + 13248)
	Recovery system     0 	(0 + 0)
	Threads             409336 	(406936 + 2400)
	innodb_io_pattern   0 	(0 + 0)
	Buffer pool size        1769471
	Buffer pool size, bytes 28991012864
	Free buffers            0
	Database pages          1696503
	Modified db pages       160602
	Pages read 15240822, created 1770238, written 21705836
	Number of rows inserted 50678311, updated 66425915, deleted 20605903, read 454561562
	0 queries inside InnoDB, 0 queries in queue`
	mpA := map[string]int64{
		"innodb_transactions":       1170664159,
		"ibuf_used_cells":           1,
		"log_bytes_flushed":         540805326864,
		"additional_pool_alloc":     0,
		"unflushed_log":             0,
		"last_checkpoint":           540805205461,
		"pages_written":             21705836,
		"read_views":                1,
		"pending_aio_sync_ios":      0,
		"pending_chkp_writes":       0,
		"lock_system_memory":        29232616,
		"recovery_system_memory":    0,
		"pending_ibuf_aio_reads":    0,
		"log_bytes_written":         540805326864,
		"total_mem_alloc":           29642194944,
		"database_pages":            1696503,
		"rows_read":                 454561562,
		"innodb_sem_waits":          1,
		"innodb_locked_tables":      2,
		"pending_log_flushes":       0,
		"pending_log_writes":        0,
		"spin_rounds":               247437732359,
		"innodb_tables_in_use":      2,
		"hash_index_cells_total":    4425293,
		"file_system_memory":        313848,
		"innodb_io_pattern_memory":  0,
		"spin_waits":                89337380,
		"history_list":              132,
		"file_fsyncs":               947800,
		"pending_buf_pool_flushes":  0,
		"ibuf_free_cells":           887,
		"thread_hash_memory":        409336,
		"pages_read":                15240822,
		"rows_updated":              66425915,
		"page_hash_memory":          11688584,
		"os_waits":                  321325753,
		"current_transactions":      2,
		"file_writes":               15635445,
		"hash_index_cells_used":     4229064,
		"log_writes":                3430041,
		"rows_deleted":              20605903,
		"unpurged_txns":             306,
		"pending_aio_log_ios":       0,
		"uncheckpointed_bytes":      121403,
		"queries_queued":            0,
		"innodb_lock_structs":       37,
		"pending_normal_aio_reads":  0,
		"ibuf_cell_count":           889,
		"ibuf_inserts":              593983,
		"ibuf_merged":               1054081,
		"queries_inside":            0,
		"active_transactions":       1,
		"locked_transactions":       2,
		"pool_size":                 1769471,
		"rows_inserted":             50678311,
		"innodb_sem_wait_time_ms":   1000,
		"innodb_lock_wait_secs":     32,
		"adaptive_hash_memory":      1538240664,
		"pages_created":             1770238,
		"file_reads":                8782182,
		"pending_normal_aio_writes": 0,
		"modified_pages":            160602,
		"dictionary_cache_memory":   145525560,
		"free_pages":                0,
	}
	mpB := parseInnodbStatusWithRule(str)
	for k, v := range mpA {
		if mpB[k] != v {
			t.Errorf("parseInnodbStatusWithRule failed.\n At key (%v).\nresult map value(%v) != test value(%v).\n", k, mpB[k], v)
		}
	}

	for k, v := range mpB {
		if mpA[k] != v {
			t.Errorf("parseInnodbStatusWithRule failed.\n At key (%v).\nresult map value(%v) != test value(%v).\n", k, mpA[k], v)
		}
	}
}
