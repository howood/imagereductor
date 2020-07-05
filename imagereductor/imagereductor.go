package main

import (
	"fmt"
	"os"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/infrastructure/custommiddleware"
	"github.com/howood/imagereductor/interfaces/service/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// DefaultPort is default port of server
var DefaultPort = "8080"

func main() {
	if os.Getenv("SERVER_PORT") != "" {
		DefaultPort = os.Getenv("SERVER_PORT")
	}
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", handler.ImageReductionHandler{}.Request)
	if os.Getenv("ADMIN_MODE") == "enable" {
		e.GET("/token", handler.TokenHandler{}.Request)
	}
	jwtconfig := middleware.JWTConfig{
		Skipper:    custommiddleware.OptionsMethodSkipper,
		Claims:     &entity.JwtClaims{},
		SigningKey: []byte(actor.TokenSecret),
	}
	e.POST("/", handler.ImageReductionHandler{}.Upload, middleware.JWTWithConfig(jwtconfig))

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", DefaultPort)))
}
