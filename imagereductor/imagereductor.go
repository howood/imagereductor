package main

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/di/uccluster"
	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/infrastructure/custommiddleware"
	"github.com/howood/imagereductor/interfaces/handler"
	"github.com/howood/imagereductor/library/utils"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	defaultPort := utils.GetOsEnv("SERVER_PORT", "8080")

	usecaseCluster := uccluster.NewUsecaseCluster()
	baseHandler := handler.BaseHandler{UcCluster: usecaseCluster}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	if os.Getenv("ADMIN_MODE") == "enable" {
		e.GET("/token", handler.TokenHandler{BaseHandler: baseHandler}.Request, custommiddleware.IPRestriction())
	}
	jwtconfig := echojwt.Config{
		Skipper: custommiddleware.OptionsMethodSkipper,
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return new(entity.JwtClaims)
		},
		SigningKey: []byte(actor.TokenSecret),
	}
	e.GET("/", handler.ImageReductionHandler{BaseHandler: baseHandler}.Request)
	e.POST("/", handler.ImageReductionHandler{BaseHandler: baseHandler}.Upload, echojwt.WithConfig(jwtconfig))
	e.GET("/files", handler.ImageReductionHandler{BaseHandler: baseHandler}.RequestFile)
	e.POST("/files", handler.ImageReductionHandler{BaseHandler: baseHandler}.UploadFile, echojwt.WithConfig(jwtconfig))
	e.GET("/streaming", handler.ImageReductionHandler{BaseHandler: baseHandler}.RequestStreaming)
	e.GET("/info", handler.ImageReductionHandler{BaseHandler: baseHandler}.RequestInfo)

	e.Logger.Fatal(e.Start(":" + defaultPort))
}
