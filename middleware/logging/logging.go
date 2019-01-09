package logging

import (
	"time"
	"github.com/sirupsen/logrus"
	"github.com/labstack/echo"
)

func MakeLogEntry(c echo.Context) *logrus.Entry {
	if c == nil {
		return logrus.WithFields(logrus.Fields{
			"at": time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	return logrus.WithFields(logrus.Fields{
		"at": time.Now().Format("2006-01-02 15:04:05"),
		//"method": c.Request().Method,
		"code": c.Response().Status,
	})
}

func MiddlewareLogging(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logrus.WithFields(logrus.Fields{
			"at":         time.Now().Format("2006-01-02 15:04:05"),
			"method":     c.Request().Method,
			"uri":        c.Request().URL.String(),
			"ip":         c.Request().RemoteAddr,
			"host":       c.Request().Host,
			"user_agent": c.Request().UserAgent(),
			"code":       c.Response().Status,
		}).Info("Incoming request")
		return next(c)
	}
}

