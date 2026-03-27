module github.com/kidyme/nexus/offline

go 1.24.2

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/kidyme/nexus/common v0.0.0
)

require github.com/lmittmann/tint v1.0.4 // indirect

replace github.com/kidyme/nexus/common => ../common
