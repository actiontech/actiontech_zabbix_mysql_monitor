#!/bin/bash
PATH_SCRIPTS=/var/lib/zabbix/actiontech/scripts/
PATH_TEMPLATES=/var/lib/zabbix/actiontech/templates/
mkdir -p $PATH_SCRIPTS
cp actiontech_mysql_monitor $PATH_SCRIPTS
mkdir -p $PATH_TEMPLATES
cp actiontech_zabbix_agent_template_mysql_server.xml $PATH_TEMPLATES
cp userparameter_actiontech_mysql.conf $PATH_TEMPLATES
echo -e 'Defaults:zabbix !requiretty\nzabbix ALL=(ALL) NOPASSWD: /bin/netstat' > /etc/sudoers.d/actiontech-MYSQL-MONITOR 
