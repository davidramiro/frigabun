package main

import (
	"fmt"
	"github.com/davidramiro/frigabun/internal/api"
	"github.com/davidramiro/frigabun/services/factory"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

func main() {

	_, err := os.Stat("/data/options.json")
	if err != nil {
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
	} else {
		viper.AddConfigPath("/data")
		viper.SetConfigName("options")
	}

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("could not read config file")
	}

	if viper.GetBool("api.prettyLog") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	switch viper.GetString("api.logLevel") {
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msg("starting frigabun")

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	enableStatusLog := viper.GetBool("api.enableStatusLog")

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			uri := v.URI
			if enableStatusLog || !strings.Contains(v.URI, "/status") {
				log.Info().
					Str("URI", uri).
					Int("status", v.Status).
					Msg("request")
			}

			return nil
		},
	}))
	e.Use(middleware.Recover())

	serviceFactory, err := factory.NewDnsUpdateServiceFactory()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot init service serviceFactory")
	}

	updateApi := api.NewUpdateApi(serviceFactory)
	g := e.Group("/api")
	g.GET("/update", updateApi.HandleUpdateRequest)
	g.GET("/status", updateApi.HandleStatusCheck)

	endpoint := fmt.Sprintf(":%d", viper.GetInt("api.port"))
	log.Info().Str("port", endpoint).Msg("starting server")

	log.Fatal().Err(e.Start(endpoint)).Msg("server error")
}
