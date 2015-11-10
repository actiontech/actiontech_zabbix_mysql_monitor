go build -ldflags "-X main.Version='`git rev-parse --abbrev-ref HEAD`++`git rev-parse HEAD`'" actiontech_mysql_monitor.go
