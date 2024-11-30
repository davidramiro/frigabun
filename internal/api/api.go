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

	logger := log.With().Str("subdomains", request.Subdomains).Str("domain", request.Domain).Str("IP", request.IP).Logger()
	logger.Info().Msg("dns update request received")

	err = validateRequest(request.Domain, request.IP)
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return c.String(400, err.Error())
	}

	subdomains := strings.Split(request.Subdomains, ",")

	if subdomains == nil {
		subdomains = []string{""}
	}

	for i := range subdomains {
		logger.Debug().Msgf("handling request %d of %d", i+1, len(subdomains))

		service, err := u.dnsServiceFactory.Find(services.Registrar(request.Registrar))
		if err != nil {
			logger.Err(err).Msg("getting registrar from factory failed")
			return c.String(400, err.Error())
		}

		request := &services.DynDnsRequest{
			IP:        request.IP,
			Domain:    request.Domain,
			Subdomain: subdomains[i],
		}

		err = service.UpdateRecord(request)
		if err != nil {
			logger.Err(err).Msg("updating record failed")
			return c.String(http.StatusInternalServerError, err.Error())
		}

	}

	logger.Info().Int("updates", len(subdomains)).Msg("successfully created")

	return c.String(http.StatusOK,
		fmt.Sprintf("created %d entries for subdomains %s on %s: %s",
			len(subdomains),
			request.Subdomains,
			request.Domain,
			request.IP),
	)
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
