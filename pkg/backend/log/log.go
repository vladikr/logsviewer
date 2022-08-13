package log

import (
    "os"
    "log"
)

var Log = DefaultLogger()

func DefaultLogger() *log.Logger {
        flags := log.Lshortfile
		return log.New(os.Stdout, "", flags)
}
