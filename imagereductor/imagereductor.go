package main

import (
	"fmt"
	"os"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/infrastructure/custommiddleware"
	"github.com/howood/imagereductor/interfaces/service/handler"
	"github.com/howood/imagereductor/library/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// DefaultPort is default port of server
var DefaultPort = utils.GetOsEnv("SERVER_PORT", "8080")

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	if os.Getenv("ADMIN_MODE") == "enable" {
		e.GET("/token", handler.TokenHandler{}.Request, custommiddleware.IPRestriction())
	}
	jwtconfig := middleware.JWTConfig{
		Skipper:    custommiddleware.OptionsMethodSkipper,
		Claims:     &entity.JwtClaims{},
		SigningKey: []byte(actor.TokenSecret),
	}
	e.GET("/", handler.ImageReductionHandler{}.Request)
	e.POST("/", handler.ImageReductionHandler{}.Upload, middleware.JWTWithConfig(jwtconfig))
	e.GET("/files", handler.ImageReductionHandler{}.RequestFile)
	e.POST("/files", handler.ImageReductionHandler{}.UploadFile, middleware.JWTWithConfig(jwtconfig))

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", DefaultPort)))
}
