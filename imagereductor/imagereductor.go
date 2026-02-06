package main

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/di/uccluster"
	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/infrastructure/custommiddleware"
	"github.com/howood/imagereductor/interfaces/handler"
	"github.com/howood/imagereductor/library/utils"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"golang.org/x/time/rate"
)

const (
	ipAddressRateLimitBurst        = 100
	ipAddressRateLimitCleanupTTL   = 15 * time.Minute
	ipAddressRateLimitCleanupEvery = 5 * time.Minute
)

func main() {
	defaultPort := utils.GetOsEnv("SERVER_PORT", "8080")

	usecaseCluster, err := uccluster.NewUsecaseCluster()
	if err != nil {
		panic(err)
	}
	baseHandler := handler.BaseHandler{UcCluster: usecaseCluster}
	ipLimiter := custommiddleware.NewRateLimiter(custommiddleware.RateLimitConfig{
		Rate:     rate.Every(time.Second),
		Burst:    ipAddressRateLimitBurst,
		KeyFunc:  custommiddleware.IPKeyFunc,
		ErrorMsg: "Too many requests from your IP",
		Skipper: func(c *echo.Context) bool {
			return c.RealIP() == "127.0.0.1"
		},
		CleanupTTL:   ipAddressRateLimitCleanupTTL,
		CleanupEvery: ipAddressRateLimitCleanupEvery,
	})

	e := echo.New()
	e.Use(custommiddleware.JSONRequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS("*"))
	e.Use(ipLimiter.Middleware())

	if os.Getenv("ADMIN_MODE") == "enable" {
		tokenHandler := handler.NewTokenHandler(baseHandler)
		e.GET("/token", tokenHandler.Request, custommiddleware.IPRestriction())
	}
	jwtconfig := echojwt.Config{
		Skipper: custommiddleware.OptionsMethodSkipper,
		NewClaimsFunc: func(_ *echo.Context) jwt.Claims {
			return new(entity.JwtClaims)
		},
		SigningKey: []byte(actor.TokenSecret),
	}
	imageReductorHandler := handler.NewImageReductionHandler(baseHandler)
	e.GET("/", imageReductorHandler.Request)
	e.POST("/", imageReductorHandler.Upload, echojwt.WithConfig(jwtconfig))
	e.GET("/files", imageReductorHandler.RequestFile)
	e.POST("/files", imageReductorHandler.UploadFile, echojwt.WithConfig(jwtconfig))
	e.GET("/streaming", imageReductorHandler.RequestStreaming)
	e.GET("/info", imageReductorHandler.RequestInfo)

	if err := e.Start(":" + defaultPort); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
