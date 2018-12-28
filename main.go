package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"fmt"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"github.com/go-playground/validator"
	"time"
)

func main(){
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.GET("/index", func(c echo.Context) (err error) {
		return c.JSON(http.StatusOK, true)
	})

	logrus.WithFields(log.Fields{
		"at": time.Now().Format("2006-01-02 15:04:05"),
	})

	e.Logger.Fatal(e.Start(":9000"))
}

// CustomValidator holds custom validator
type CustomValidator struct {
	V *validator.Validate
}

// Validate validates the request
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.V.Struct(i)
}

func AuthAPI() echo.MiddlewareFunc {
	//func AuthAPI(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			//c.Error(errors.New("error"))
			//return next(c)

			fmt.Println("something 1")
			//return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": "unauthorized hehe"})
			return next(c)
			//return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": "unauthorized 2"})
		}
	}

}