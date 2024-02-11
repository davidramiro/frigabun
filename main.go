package main

import (
	"fmt"
	"github.com/spf13/viper"
	"regexp"
	"strings"

	"github.com/davidramiro/frigabun/internal/api"
	"github.com/davidramiro/frigabun/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	logger.InitLog()

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
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
			if viper.GetBool("api.hideApiKeyInLogs") {
				re := regexp.MustCompile(`key=([^&]*)`)
				uri = re.ReplaceAllString(v.URI, `key=**REDACTED**`)
			}

			if viper.GetBool("api.enableStatusLog") || !strings.Contains(v.URI, "/status") {
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

	endpoint := fmt.Sprintf(":%d", viper.GetInt("api.port"))
	e.Logger.Fatal(e.Start(endpoint))
}
