#!/bin/bash
PATH_SCRIPTS=/var/lib/zabbix/actiontech/scripts/
PATH_TEMPLATES=/var/lib/zabbix/actiontech/templates/
mkdir -p $PATH_SCRIPTS
cp actiontech_mysql_monitor $PATH_SCRIPTS
mkdir -p $PATH_TEMPLATES
cp actiontech_zabbix_agent_template_percona_mysql_server.xml $PATH_TEMPLATES
cp userparameter_actiontech_mysql.conf $PATH_TEMPLATES
