package main

import (
	"os"

	"github.com/howood/imagereductor/interfaces/service/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", handler.ImageReductionHandler{}.Request)
	if os.Getenv("VERIFY_MODE") == "enable" {
		e.POST("/", handler.ImageReductionHandler{}.Upload)
	}
	e.Logger.Fatal(e.Start(":8080"))
}
