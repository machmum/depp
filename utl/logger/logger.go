package logger

import (
	"strings"
	"runtime"
	"fmt"
	"log"
	"github.com/sirupsen/logrus"
)

// Log represents zerolog logger
type Log struct {
	logger *logrus.Logger
}

// New instantiates new zero logger
func New() *Log {
	l := logrus.New()
	return &Log{
		logger: l,
	}
}

func chopPath(original string) string {
	i := strings.LastIndex(original, "/")
	if i == -1 {
		return original
	} else {
		return original[i+1:]
	}
}

// Log logs using zerolog
func (z *Log) Log(source, msg string, err error) {

	depth := 3 // get real error line
	function, file, line, _ := runtime.Caller(depth)
	log.Printf("INFO: Err trace %s", fmt.Sprintf("File: %s  Function: %s Line: %d", chopPath(file), runtime.FuncForPC(function).Name(), line))

	return

	// if used logrus
	//logrus.WithFields(logrus.Fields{
	//	"error":  err,
	//	"source": source,
	//	"at":     time.Now().Format("2006-01-02 15:04:05"),
	//	//"code":   ctx.Response().Status,
	//	//"params": params,
	//}).Error(msg)
}
