package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"net/http"
)

type GandiDnsUpdateService struct {
	registrarSettings
	apiKey string
	client HTTPClient
}

func NewGandiDnsUpdateService(client HTTPClient) (*GandiDnsUpdateService, error) {
	baseUrl := viper.GetString("gandi.baseUrl")
	ttl := viper.GetInt("gandi.ttl")
	apikey := viper.GetString("gandi.apiKey")

	log.Info().Msg("initializing gandi service")

	if len(baseUrl) == 0 || ttl == 0 || len(apikey) == 0 {
		return nil, ErrMissingInfoForServiceInit
	}

	if client == nil {
		client = &http.Client{}
	}

	return &GandiDnsUpdateService{
		registrarSettings: registrarSettings{
			baseUrl: baseUrl,
			ttl:     ttl,
		},
		apiKey: apikey,
		client: client,
	}, nil
}

type GandiApiRequest struct {
	Subdomain string   `json:"rrset_name"`
	Type      string   `json:"rrset_type"`
	TTL       int      `json:"rrset_ttl"`
	IPValues  []string `json:"rrset_values"`
}

func (g *GandiDnsUpdateService) UpdateRecord(request *DynDnsRequest) error {

	gandiRequest := &GandiApiRequest{
		Subdomain: request.Subdomain,
		IPValues:  []string{request.IP},
		TTL:       g.ttl,
		Type:      "A",
	}

	endpoint := fmt.Sprintf("%s/domains/%s/records/%s/A", g.baseUrl,
		request.Domain, gandiRequest.Subdomain)

	logger := log.With().Str("func", "UpdateRecord").Str("registrar", "gandi").Str("endpoint", endpoint).Str("domain", request.Domain).Str("subdomain", request.Subdomain).Logger()
	logger.Info().Msg("building update request")

	body, err := json.Marshal(gandiRequest)
	if err != nil {
		logger.Error().Err(err).Msg(ErrBuildingRequest.Error())
		return ErrBuildingRequest
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Error().Err(err).Msg(ErrBuildingRequest.Error())
		return ErrBuildingRequest
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Apikey "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg(ErrExecutingRequest.Error())
		return ErrExecutingRequest
	}

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		logger.Error().Bytes("response", b).Msg(ErrRegistrarRejectedRequest.Error())
		return ErrRegistrarRejectedRequest
	}

	logger.Info().Msg("update request successful")

	return nil
}

func (g *GandiDnsUpdateService) Registrar() Registrar {
	return "gandi"
}
