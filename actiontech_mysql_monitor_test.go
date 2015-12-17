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
	Pending normal aio reads: 1, aio writes: 2,
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
		"spin_rounds":               247453947221,
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
		"pending_normal_aio_reads":  1,
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
		"pending_normal_aio_writes": 2,
		"modified_pages":            160602,
		"dictionary_cache_memory":   145525560,
		"free_pages":                0,
	}
	mpB := parseInnodbStatusWithRule(str)
	for k, v := range mpA {
		if mpB[k] != v {
			t.Errorf("parseInnodbStatusWithRule failed.\n At key (%v).\nresult map value(%v) != test value(%v).\n", k, v, mpB[k])
		}
	}

	for k, v := range mpB {
		if mpA[k] != v {
			t.Errorf("parseInnodbStatusWithRule failed.\n At key (%v).\nresult map value(%v) != test value(%v).\n", k, mpA[k], v)
		}
	}
}

// mysql 5.7
func TestParseInnodbStatusWithRule2(t *testing.T) {
	str := `=====================================
2015-12-13 16:34:43 0x7fe9782b4700 INNODB MONITOR OUTPUT
=====================================
Per second averages calculated from the last 15 seconds
-----------------
BACKGROUND THREAD
-----------------
srv_master_thread loops: 852474 srv_active, 0 srv_shutdown, 17588 srv_idle
srv_master_thread log flush and writes: 870062
----------
SEMAPHORES
----------
OS WAIT ARRAY INFO: reservation count 2695766
OS WAIT ARRAY INFO: signal count 2861714
RW-shared spins 0, rounds 1341306, OS waits 183508
RW-excl spins 0, rounds 553078, OS waits 13765
RW-sx spins 4184, rounds 119993, OS waits 3770
Spin rounds per wait: 1341306.00 RW-shared, 553078.00 RW-excl, 28.68 RW-sx
------------
TRANSACTIONS
------------
Trx id counter 784052289
Purge done for trx's n:o < 784052289 undo n:o < 0 state: running but idle
History list length 1758
LIST OF TRANSACTIONS FOR EACH SESSION:
---TRANSACTION 422152722650960, not started
0 lock struct(s), heap size 1136, 0 row lock(s)
---TRANSACTION 422152722654608, not started
0 lock struct(s), heap size 1136, 0 row lock(s)
---TRANSACTION 422152722653696, not started
0 lock struct(s), heap size 1136, 0 row lock(s)
---TRANSACTION 422152722652784, not started
0 lock struct(s), heap size 1136, 0 row lock(s)
---TRANSACTION 422152722651872, not started
0 lock struct(s), heap size 1136, 0 row lock(s)
---TRANSACTION 422152722657344, not started
0 lock struct(s), heap size 1136, 0 row lock(s)
--------
FILE I/O
--------
I/O thread 0 state: waiting for completed aio requests (insert buffer thread)
I/O thread 1 state: waiting for completed aio requests (log thread)
I/O thread 2 state: waiting for completed aio requests (read thread)
I/O thread 3 state: waiting for completed aio requests (read thread)
I/O thread 4 state: waiting for completed aio requests (read thread)
I/O thread 5 state: waiting for completed aio requests (read thread)
I/O thread 6 state: waiting for completed aio requests (read thread)
I/O thread 7 state: waiting for completed aio requests (read thread)
I/O thread 8 state: waiting for completed aio requests (read thread)
I/O thread 9 state: waiting for completed aio requests (read thread)
I/O thread 10 state: waiting for completed aio requests (read thread)
I/O thread 11 state: waiting for completed aio requests (read thread)
I/O thread 12 state: waiting for completed aio requests (read thread)
I/O thread 13 state: waiting for completed aio requests (read thread)
I/O thread 14 state: waiting for completed aio requests (write thread)
I/O thread 15 state: waiting for completed aio requests (write thread)
I/O thread 16 state: waiting for completed aio requests (write thread)
I/O thread 17 state: waiting for completed aio requests (write thread)
I/O thread 18 state: waiting for completed aio requests (write thread)
I/O thread 19 state: waiting for completed aio requests (write thread)
I/O thread 20 state: waiting for completed aio requests (write thread)
I/O thread 21 state: waiting for completed aio requests (write thread)
I/O thread 22 state: waiting for completed aio requests (write thread)
I/O thread 23 state: waiting for completed aio requests (write thread)
I/O thread 24 state: waiting for completed aio requests (write thread)
I/O thread 25 state: waiting for completed aio requests (write thread)
Pending normal aio reads: [1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 5] , aio writes: [1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0] ,
 ibuf aio reads:, log i/o's:, sync i/o's:
Pending flushes (fsync) log: 0; buffer pool: 0
1845120 OS file reads, 322178887 OS file writes, 8595398 OS fsyncs
0.13 reads/s, 16384 avg bytes/read, 692.42 writes/s, 20.87 fsyncs/s
-------------------------------------
INSERT BUFFER AND ADAPTIVE HASH INDEX
-------------------------------------
Ibuf: size 1, free list len 78, seg size 80, 221065 merges
merged operations:
 insert 410367, delete mark 0, delete 0
discarded operations:
 insert 0, delete mark 0, delete 0
Hash table size 9461113, node heap has 1 buffer(s)
Hash table size 9461113, node heap has 43 buffer(s)
Hash table size 9461113, node heap has 1 buffer(s)
Hash table size 9461113, node heap has 7 buffer(s)
Hash table size 9461113, node heap has 2 buffer(s)
Hash table size 9461113, node heap has 1 buffer(s)
Hash table size 9461113, node heap has 5028 buffer(s)
Hash table size 9461113, node heap has 19 buffer(s)
701.82 hash searches/s, 481.37 non-hash searches/s
---
LOG
---
Log sequence number 338086806104
Log flushed up to   338086701989
Pages flushed up to 337893695348
Last checkpoint at  337892779162
1 pending log flushes, 5 pending chkp writes
287170691 log i/o's done, 651.60 log i/o's/second
----------------------
BUFFER POOL AND MEMORY
----------------------
Total large memory allocated 35181821952
Dictionary memory allocated 845480
Buffer pool size   2097024
Free buffers       8192
Database pages     2083730
Old database pages 769025
Modified db pages  14765
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 698403, not young 367390
0.07 youngs/s, 0.40 non-youngs/s
Pages read 1845072, created 5545638, written 31512819
0.13 reads/s, 1.60 creates/s, 31.66 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 2083730, unzip_LRU len: 0
I/O sum[15632]:cur[0], unzip sum[0]:cur[0]
----------------------
INDIVIDUAL BUFFER POOL INFO
----------------------
---BUFFER POOL 0
Buffer pool size   262144
Free buffers       1024
Database pages     260480
Old database pages 96133
Modified db pages  1779
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 89745, not young 52407
0.00 youngs/s, 0.00 non-youngs/s
Pages read 226913, created 696652, written 3928743
0.00 reads/s, 0.60 creates/s, 3.93 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 260480, unzip_LRU len: 0
I/O sum[1954]:cur[0], unzip sum[0]:cur[0]
---BUFFER POOL 1
Buffer pool size   262112
Free buffers       1024
Database pages     260445
Old database pages 96120
Modified db pages  1094
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 96934, not young 63327
0.00 youngs/s, 0.20 non-youngs/s
Pages read 235207, created 682482, written 3910248
0.07 reads/s, 0.00 creates/s, 3.93 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 260445, unzip_LRU len: 0
I/O sum[1954]:cur[0], unzip sum[0]:cur[0]
---BUFFER POOL 2
Buffer pool size   262144
Free buffers       1024
Database pages     260482
Old database pages 96134
Modified db pages  1906
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 88308, not young 23120
0.00 youngs/s, 0.00 non-youngs/s
Pages read 227545, created 685713, written 3912381
0.00 reads/s, 0.07 creates/s, 4.00 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 260482, unzip_LRU len: 0
I/O sum[1954]:cur[0], unzip sum[0]:cur[0]
---BUFFER POOL 3
Buffer pool size   262112
Free buffers       1024
Database pages     260448
Old database pages 96121
Modified db pages  2463
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 92299, not young 47113
0.00 youngs/s, 0.00 non-youngs/s
Pages read 229238, created 695905, written 3987887
0.00 reads/s, 0.13 creates/s, 4.00 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 260448, unzip_LRU len: 0
I/O sum[1954]:cur[0], unzip sum[0]:cur[0]
---BUFFER POOL 4
Buffer pool size   262144
Free buffers       1024
Database pages     260484
Old database pages 96135
Modified db pages  1931
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 99976, not young 58704
0.00 youngs/s, 0.00 non-youngs/s
Pages read 234007, created 695549, written 3940433
0.00 reads/s, 0.60 creates/s, 4.00 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 260484, unzip_LRU len: 0
I/O sum[1954]:cur[0], unzip sum[0]:cur[0]
---BUFFER POOL 5
Buffer pool size   262112
Free buffers       1024
Database pages     260453
Old database pages 96123
Modified db pages  1954
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 26168, not young 19195
0.00 youngs/s, 0.20 non-youngs/s
Pages read 223661, created 694001, written 3942893
0.07 reads/s, 0.20 creates/s, 3.93 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 260453, unzip_LRU len: 0
I/O sum[1954]:cur[0], unzip sum[0]:cur[0]
---BUFFER POOL 6
Buffer pool size   262144
Free buffers       1024
Database pages     260481
Old database pages 96134
Modified db pages  1644
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 99554, not young 56165
0.00 youngs/s, 0.00 non-youngs/s
Pages read 233919, created 697369, written 3960427
0.00 reads/s, 0.00 creates/s, 3.93 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 260481, unzip_LRU len: 0
I/O sum[1954]:cur[0], unzip sum[0]:cur[0]
---BUFFER POOL 7
Buffer pool size   262112
Free buffers       1024
Database pages     260457
Old database pages 96125
Modified db pages  1994
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages made young 105419, not young 47359
0.07 youngs/s, 0.00 non-youngs/s
Pages read 234582, created 697967, written 3929807
0.00 reads/s, 0.00 creates/s, 3.93 writes/s
Buffer pool hit rate 1000 / 1000, young-making rate 0 / 1000 not 0 / 1000
Pages read ahead 0.00/s, evicted without access 0.00/s, Random read ahead 0.00/s
LRU len: 260457, unzip_LRU len: 0
I/O sum[1954]:cur[0], unzip sum[0]:cur[0]
--------------
ROW OPERATIONS
--------------
0 queries inside InnoDB, 0 queries in queue
0 read views open inside InnoDB
Process ID=122556, Main thread ID=140640873428736, state: sleeping
Number of rows inserted 283194386, updated 292299839, deleted 131, read 839949762
79.33 inserts/s, 651.89 updates/s, 0.00 deletes/s, 651.89 reads/s
----------------------------
END OF INNODB MONITOR OUTPUT
============================`

	mpA := map[string]int64{
		"innodb_transactions":       784052289,
		"ibuf_used_cells":           1,
		"log_bytes_flushed":         338086701989,
		"additional_pool_alloc":     0,
		"unflushed_log":             104115,
		"last_checkpoint":           337892779162,
		"pages_written":             31512819,
		"read_views":                0,
		"pending_aio_sync_ios":      0,
		"pending_chkp_writes":       5,
		"pending_ibuf_aio_reads":    0,
		"log_bytes_written":         338086806104,
		"total_mem_alloc":           35181821952,
		"database_pages":            2083730,
		"rows_read":                 839949762,
		"innodb_sem_waits":          0,
		"pending_log_flushes":       0,
		"pending_log_writes":        1,
		"spin_rounds":               2014377,
		"hash_index_cells_total":    75688904,
		"spin_waits":                4184,
		"history_list":              1758,
		"file_fsyncs":               8595398,
		"pending_buf_pool_flushes":  0,
		"ibuf_free_cells":           78,
		"pages_read":                1845072,
		"rows_updated":              292299839,
		"os_waits":                  201043,
		"current_transactions":      6,
		"file_writes":               322178887,
		"hash_index_cells_used":     0,
		"log_writes":                287170691,
		"rows_deleted":              131,
		"unpurged_txns":             0,
		"pending_aio_log_ios":       0,
		"uncheckpointed_bytes":      194026942,
		"queries_queued":            0,
		"innodb_lock_structs":       0,
		"pending_normal_aio_reads":  15,
		"ibuf_cell_count":           80,
		"ibuf_inserts":              410367,
		"ibuf_merged":               410367,
		"queries_inside":            0,
		"active_transactions":       0,
		"locked_transactions":       0,
		"pool_size":                 2097024,
		"rows_inserted":             283194386,
		"pages_created":             5545638,
		"file_reads":                1845120,
		"pending_normal_aio_writes": 1,
		"modified_pages":            14765,
		"free_pages":                8192,
		"ibuf_merges":               221065,
	}
	mpB := parseInnodbStatusWithRule(str)
	for k, v := range mpA {
		if mpB[k] != v {
			t.Errorf("parseInnodbStatusWithRule failed.\n At key (%v).\nresult map value(%v) != test value(%v).\n", k, v, mpB[k])
		}
	}

	for k, v := range mpB {
		if mpA[k] != v {
			t.Errorf("parseInnodbStatusWithRule failed.\n At key (%v).\nresult map value(%v) != test value(%v).\n", k, mpA[k], v)
		}
	}
}

// mysql 5.7
func TestParseInnodbStatusWithRule3(t *testing.T) {
	str := `=====================================
Pending normal aio reads:, aio writes:,
 ibuf aio reads: 1, log i/o's: 2, sync i/o's: 3
============================`

	mpA := map[string]int64{
		"pending_aio_sync_ios":      3,
		"pending_ibuf_aio_reads":    1,
		"pending_aio_log_ios":       2,
		"pending_normal_aio_reads":  0,
		"pending_normal_aio_writes": 0,
		"current_transactions":      0,
		"active_transactions":       0,
		"locked_transactions":       0,
	}
	mpB := parseInnodbStatusWithRule(str)
	for k, v := range mpA {
		if mpB[k] != v {
			t.Errorf("parseInnodbStatusWithRule failed.\n At key (%v).\nresult map value(%v) != test value(%v).\n", k, v, mpB[k])
		}
	}

	for k, v := range mpB {
		if mpA[k] != v {
			t.Errorf("parseInnodbStatusWithRule failed.\n At key (%v).\nresult map value(%v) != test value(%v).\n", k, mpA[k], v)
		}
	}
}
