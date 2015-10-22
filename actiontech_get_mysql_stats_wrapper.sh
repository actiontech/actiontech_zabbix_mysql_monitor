#!/bin/sh
# The wrapper for zabbix golang script.
# It runs the script every 5 min. and parses the cache file on each following run.
# Version: 1.0.0

ITEM=$1
HOST=127.0.0.1
DIR=`dirname $0`
CMD="$DIR/actiontech_mysql_monitor --host $HOST --user=zbx --pass=zabbix"
CACHEFILE="/tmp/actiontech_$HOST-mysql_zabbix_stats.txt"

if [ "$ITEM" = "running-slave" ]; then
    # Check for running slave
    RES=`HOME=~zabbix mysql -e 'SHOW SLAVE STATUS\G' | egrep '(Slave_IO_Running|Slave_SQL_Running):' | awk -F: '{print $2}' | tr '\n' ','`
    if [ "$RES" = " Yes, Yes," ]; then
        echo 1
    else
        echo 0
    fi
    exit
elif [ -e $CACHEFILE ]; then
    # Check and run the script
    TIMEFLM=`stat -c %Y /tmp/actiontech_$HOST-mysql_zabbix_stats.txt`
    TIMENOW=`date +%s`
    if [ `expr $TIMENOW - $TIMEFLM` -gt 300 ]; then
        rm -f $CACHEFILE
        $CMD 2>&1 > /dev/null
    fi
else
    $CMD 2>&1 > /dev/null
fi

# Parse cache file
if [ -e $CACHEFILE ]; then
    cat $CACHEFILE | sed 's/ /\n/g; s/-1/0/g'| grep $ITEM | awk -F: '{print $2}'
else
    echo "ERROR: run the command manually to investigate the problem: $CMD"
fi