package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/davidramiro/frigabun/internal/logger"
	"github.com/davidramiro/frigabun/pkg/gandi"
	"github.com/davidramiro/frigabun/pkg/porkbun"
	"github.com/labstack/echo/v4"
)

type StatusResponse struct {
	ApiStatus bool `json:"api_status"`
}

type ApiError struct {
	Code    int
	Message string
}

type UpdateRequest struct {
	Domain       string `query:"domain"`
	Subdomains   string `query:"subdomain"`
	IP           string `query:"ip"`
	ApiKey       string `query:"apikey"`
	ApiSecretKey string `query:"apisecretkey"`
	Registrar    string `query:"registrar"`
}

func HandleUpdateRequest(c echo.Context) error {
	var request UpdateRequest

	err := c.Bind(&request)
	if err != nil {
		logger.Log.Error().Err(err).Msg("binding request to struct failed")
		return c.String(http.StatusBadRequest, "bad request")
	}

	logger.Log.Info().Str("subdomains", request.Subdomains).Str("domain", request.Domain).Str("IP", request.IP).Msg("request received")

	apiErr := validateRequest(request.Domain, request.IP)
	if apiErr != nil {
		return c.String(apiErr.Code, apiErr.Message)
	}

	subdomains := strings.Split(request.Subdomains, ",")

	successfulUpdates := 0

	if len(subdomains) == 0 || subdomains[0] == "" {
		return c.String(http.StatusBadRequest, "missing subdomains parameter")
	}

	for i := range subdomains {
		if request.Registrar == "gandi" {
			dns := &gandi.GandiDns{
				IP:        request.IP,
				Domain:    request.Domain,
				Subdomain: subdomains[i],
				ApiKey:    request.ApiKey,
			}
			err := dns.SaveRecord()
			if err != nil {
				return c.String(err.Code, err.Message)
			}
		} else if request.Registrar == "porkbun" {
			dns := &porkbun.PorkbunDns{
				IP:           request.IP,
				Domain:       request.Domain,
				Subdomain:    subdomains[i],
				ApiKey:       request.ApiKey,
				SecretApiKey: request.ApiSecretKey,
			}
			err := dns.AddRecord()
			if err != nil {
				return c.String(err.Code, err.Message)
			}
		} else {
			return c.String(http.StatusBadRequest, "missing or invalid registrar")
		}

		successfulUpdates++
	}

	logger.Log.Info().Int("updates", successfulUpdates).Str("subdomains", request.Subdomains).Str("domain", request.Domain).Msg("successfully created")

	return c.String(http.StatusOK, fmt.Sprintf("created %d entries for subdomains %s on %s: %s", successfulUpdates, request.Subdomains, request.Domain, request.IP))

}

func HandleStatusCheck(c echo.Context) error {
	statusResponse := &StatusResponse{ApiStatus: true}
	return c.JSON(200, statusResponse)
}

func validateRequest(domain string, ip string) *ApiError {
	if !govalidator.IsIPv4(ip) {
		return &ApiError{Code: 400, Message: "missing or invalid IP address, only IPv4 allowed"}
	}

	if !govalidator.IsDNSName(domain) {
		return &ApiError{Code: 400, Message: "missing or invalid domain name"}
	}

	return nil
}
