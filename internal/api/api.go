package api

import (
	"fmt"
	"github.com/davidramiro/frigabun/services"
	"github.com/davidramiro/frigabun/services/factory"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
)

type UpdateApi struct {
	dnsServiceFactory factory.ServiceFactory
}

type StatusResponse struct {
	ApiStatus      bool                 `json:"api_status"`
	ActiveServices []services.Registrar `json:"active_services"`
}

type UpdateRequest struct {
	Domain     string `query:"domain"`
	Subdomains string `query:"subdomain"`
	IP         string `query:"ip"`
	Registrar  string `query:"registrar"`
}

func NewUpdateApi(dnsServiceFactory factory.ServiceFactory) *UpdateApi {
	return &UpdateApi{dnsServiceFactory: dnsServiceFactory}
}

func (u *UpdateApi) HandleUpdateRequest(c echo.Context) error {
	var request UpdateRequest

	err := c.Bind(&request)
	if err != nil {
		log.Error().Err(err).Msg(ErrCannotParseRequest.Error())
		return c.String(http.StatusBadRequest, ErrCannotParseRequest.Error())
	}

	log.Info().Str("subdomains", request.Subdomains).Str("domain", request.Domain).Str("IP", request.IP).Msg("request received")

	err = validateRequest(request.Domain, request.IP)
	if err != nil {
		return c.String(400, err.Error())
	}

	subdomains := strings.Split(request.Subdomains, ",")

	successfulUpdates := 0

	if len(subdomains) == 0 || subdomains[0] == "" {
		return c.String(http.StatusBadRequest, ErrMissingParameter.Error())
	}

	for i := range subdomains {

		service, err := u.dnsServiceFactory.Find(services.Registrar(request.Registrar))
		if err != nil {
			return c.String(400, err.Error())
		}

		request := &services.DynDnsRequest{
			IP:        request.IP,
			Domain:    request.Domain,
			Subdomain: subdomains[i],
		}

		err = service.UpdateRecord(request)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		successfulUpdates++
	}

	log.Info().Int("updates", successfulUpdates).Str("subdomains", request.Subdomains).Str("domain", request.Domain).Msg("successfully created")

	return c.String(http.StatusOK, fmt.Sprintf("created %d entries for subdomains %s on %s: %s", successfulUpdates, request.Subdomains, request.Domain, request.IP))

}

func (u *UpdateApi) HandleStatusCheck(c echo.Context) error {
	listServices := u.dnsServiceFactory.ListServices()
	statusResponse := &StatusResponse{ApiStatus: true, ActiveServices: listServices}

	return c.JSON(200, statusResponse)
}

func validateRequest(domain string, ip string) error {
	if !govalidator.IsIPv4(ip) {
		return ErrInvalidIP
	}

	if !govalidator.IsDNSName(domain) {
		return ErrInvalidDomain
	}

	return nil
}
