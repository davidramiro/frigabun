package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/davidramiro/fritzgandi/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type DnsRequest struct {
	Domain     string `query:"domain"`
	Subdomains string `query:"subdomain"`
	IP         string `query:"ip"`
	ApiKey     string `query:"apikey"`
}

type GandiRequest struct {
	Subdomain string   `json:"rrset_name"`
	Type      string   `json:"rrset_type"`
	TTL       int      `json:"rrset_ttl"`
	IPValues  []string `json:"rrset_values"`
}

type StatusResponse struct {
	ApiStatus bool `json:"api_status"`
}

type HTTPError struct {
	Code    int
	Message string
}

var logger zerolog.Logger
var configuration *config.Config

func main() {
	logger = zerolog.New(os.Stdout)

	initConfig()

	initEcho()
}

func initConfig() {
	logger.Info().Msg("reading config...")

	confExists, err := config.CheckPresent()
	if err != nil || !confExists {
		logger.Panic().Err(err).Msg("config.yml not found")
	}

	configuration, err = config.GetConfig()
	if err != nil {
		logger.Panic().Err(err).Msg("Error reading config")
	}
}

func initEcho() {
	logger.Info().Msg("setting up echo...")

	e := echo.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {

			uri := v.URI
			if configuration.Api.ApiKeyHidden {
				re := regexp.MustCompile(`apikey=([^&]*)`)
				uri = re.ReplaceAllString(v.URI, `apikey=**REDACTED**`)
			}

			if configuration.Api.StatusLogEnabled || !strings.Contains(v.URI, "/status") {
				logger.Info().
					Str("URI", uri).
					Int("status", v.Status).
					Msg("request")
			}

			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.GET("/api/update", handleUpdateRequest)
	e.GET("/api/status", handleStatusCheck)

	e.Logger.Fatal(e.Start(":" + configuration.Api.Port))
}

func handleUpdateRequest(c echo.Context) error {
	var request DnsRequest

	err := c.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("binding request to struct failed")
		return c.String(http.StatusBadRequest, "bad request")
	}

	logger.Info().Str("subdomains", request.Subdomains).Str("domain", request.Domain).Str("IP", request.IP).Msg("request received")

	httpErr := validateRequest(request.Domain, request.IP)
	if httpErr != nil {
		return c.String(httpErr.Code, httpErr.Message)
	}

	subdomains := strings.Split(request.Subdomains, ",")

	successfulUpdates := 0

	if len(subdomains) == 0 || subdomains[0] == "" {
		return c.String(http.StatusBadRequest, "missing subdomains parameter")
	}

	for i := range subdomains {
		err := putUpdate(&request, subdomains[i])
		if err != nil {
			return c.String(err.Code, err.Message)
		}

		successfulUpdates++
	}

	logger.Info().Int("updates", successfulUpdates).Str("subdomains", request.Subdomains).Str("domain", request.Domain).Msg("successfully created")

	return c.String(http.StatusOK, fmt.Sprintf("created %d entries for subdomains %s on %s: %s", successfulUpdates, request.Subdomains, request.Domain, request.IP))
}

func handleStatusCheck(c echo.Context) error {
	statusResponse := &StatusResponse{ApiStatus: true}
	return c.JSON(200, statusResponse)
}

func validateRequest(domain string, ip string) *HTTPError {
	if !govalidator.IsIPv4(ip) {
		return &HTTPError{Code: 400, Message: "missing or invalid IP address, only IPv4 allowed"}
	}

	if !govalidator.IsDNSName(domain) {
		return &HTTPError{Code: 400, Message: "missing or invalid domain name"}
	}

	return nil
}

func putUpdate(updateRequest *DnsRequest, subdomain string) *HTTPError {

	gandiRequest := &GandiRequest{
		Subdomain: subdomain,
		IPValues:  []string{updateRequest.IP},
		TTL:       configuration.Gandi.TTL,
		Type:      "A",
	}

	endpoint := fmt.Sprintf("%s/domains/%s/records/%s/A", configuration.Gandi.BaseUrl,
		updateRequest.Domain, gandiRequest.Subdomain)

	logger.Info().Str("subdomain", gandiRequest.Subdomain).Str("endpoint", endpoint).Str("IP", gandiRequest.IPValues[0]).Msg("building update request")

	body, err := json.Marshal(gandiRequest)
	if err != nil {
		logger.Error().Err(err).Msg("marshalling failed")
		return &HTTPError{Code: 400, Message: "could not parse request"}
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Error().Err(err).Msg("building request failed failed")
		return &HTTPError{Code: 400, Message: "could not create request for gandi"}
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Apikey "+updateRequest.ApiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("executing request failed")
		return &HTTPError{Code: 500, Message: "could execute request"}
	}

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		logger.Error().Err(err).Msg("gandi rejected request")
		return &HTTPError{Code: resp.StatusCode, Message: "gandi rejected request: " + string(b)}
	}

	return nil
}
