package logging

import "github.com/Tus1688/openmerce-backend/database"

const (
	WARN    = "WARN"
	INFO    = "INFO"
	ERROR   = "ERROR"
	UNKNOWN = "UNKNOWN"
)

// InsertLog is used to insert log to database and send to logstash
func InsertLog(logLevel, info string) {
	// check if logLevel is valid
	if logLevel != WARN && logLevel != INFO && logLevel != ERROR {
		logLevel = UNKNOWN
	}
	// insert the log to database
	//	we don't care if the insert is failed or not as might be the database is also down
	_, _ = database.MysqlInstance.Exec("INSERT INTO logs (log_level, info) VALUES (?, ?)", logLevel, info)
}
