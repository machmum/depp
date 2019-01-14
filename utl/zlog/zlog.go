package zlog

import (
	"os"

	"github.com/labstack/echo"

	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
)

// Log represents zerolog logger
type Log struct {
	logger *zerolog.Logger
}

// New instantiates new zero logger
func New() *Log {
	z := zerolog.New(os.Stdout)
	return &Log{
		logger: &z,
	}
}

// Log logs using zerolog
func (z *Log) Log(ctx echo.Context, source, msg string, err error, params map[string]interface{}) {

	if params == nil {
		params = make(map[string]interface{})
	}

	params["source"] = source

	if id, ok := ctx.Get("id").(int); ok {
		params["id"] = id
		params["user"] = ctx.Get("username").(string)
	}

	logrus.Println(params)

	if err != nil {
		params["error"] = err
		z.logger.Error().Fields(params).Msg(msg)

		// below if logged with logrus
		//logrus.WithFields(logrus.Fields{
		//	"error":  err,
		//	"source": source,
		//	"at":     time.Now().Format("2006-01-02 15:04:05"),
		//	"code":   ctx.Response().Status,
		//	"params": params,
		//}).Error(msg)

		return
	}

	z.logger.Info().Fields(params).Msg(msg)
}
