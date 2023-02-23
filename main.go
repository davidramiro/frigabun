package main

import (
	"regexp"
	"strings"

	"github.com/davidramiro/frigabun/internal/api"
	"github.com/davidramiro/frigabun/internal/config"
	"github.com/davidramiro/frigabun/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	logger.InitLog()

	err := config.InitConfig()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("error initializing config")
	}

	initEcho()
}

func initEcho() {
	logger.Log.Info().Msg("setting up echo...")

	e := echo.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {

			uri := v.URI
			if config.AppConfig.Api.ApiKeyHidden {
				re := regexp.MustCompile(`key=([^&]*)`)
				uri = re.ReplaceAllString(v.URI, `key=**REDACTED**`)
			}

			if config.AppConfig.Api.StatusLogEnabled || !strings.Contains(v.URI, "/status") {
				logger.Log.Info().
					Str("URI", uri).
					Int("status", v.Status).
					Msg("request")
			}

			return nil
		},
	}))
	e.Use(middleware.Recover())

	a := e.Group("/api")

	a.GET("/update", api.HandleUpdateRequest)
	a.GET("/status", api.HandleStatusCheck)

	e.Logger.Fatal(e.Start(":" + config.AppConfig.Api.Port))
}
