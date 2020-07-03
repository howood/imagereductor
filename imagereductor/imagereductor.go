package main

import (
	"os"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/infrastructure/custommiddleware"
	"github.com/howood/imagereductor/interfaces/service/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", handler.ImageReductionHandler{}.Request)
	if os.Getenv("ADMIN_MODE") == "enable" {
		e.GET("/token", handler.TokenHandler{}.Request)
	}
	jwtconfig := middleware.JWTConfig{
		Skipper:    custommiddleware.OptionsMethodSkipper,
		Claims:     &actor.JWTClaims{},
		SigningKey: []byte(actor.TokenSecret),
	}
	e.POST("/", handler.ImageReductionHandler{}.Upload, middleware.JWTWithConfig(jwtconfig))
	e.Logger.Fatal(e.Start(":8080"))
}
