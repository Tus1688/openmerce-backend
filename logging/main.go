// Copyright (c) 2023. Tus1688
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
