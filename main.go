package main

import (
	"github.com/labstack/echo"
	"fmt"
	"github.com/labstack/echo/middleware"
	"net/http"
)

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.GET("/index", func(c echo.Context) (err error) {
		return c.JSON(http.StatusOK, true)
	})

	e.Logger.Fatal(e.Start(":9000"))
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
