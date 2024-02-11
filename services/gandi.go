package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"

	"github.com/davidramiro/frigabun/internal/logger"
)

type GandiDnsUpdateService struct {
	registrarSettings
	apiKey string
}

func NewGandiDnsUpdateService() (*GandiDnsUpdateService, error) {
	baseUrl := viper.GetString("gandi.baseUrl")
	ttl := viper.GetInt("gandi.ttl")
	apikey := viper.GetString("gandi.apiKey")
	if len(baseUrl) == 0 || ttl == 0 || len(apikey) == 0 {
		return nil, fmt.Errorf(ErrMissingInfoForServiceInit, "cloudflare")
	}

	return &GandiDnsUpdateService{
		registrarSettings: registrarSettings{
			baseUrl: baseUrl,
			ttl:     ttl,
		},
		apiKey: apikey,
	}, nil
}

type GandiApiRequest struct {
	Subdomain string   `json:"rrset_name"`
	Type      string   `json:"rrset_type"`
	TTL       int      `json:"rrset_ttl"`
	IPValues  []string `json:"rrset_values"`
}

type GandiUpdateError struct {
	Code    int
	Message string
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

	logger.Log.Info().Str("subdomain", gandiRequest.Subdomain).Str("endpoint", endpoint).Str("IP", gandiRequest.IPValues[0]).Msg("building update request")

	body, err := json.Marshal(gandiRequest)
	if err != nil {
		logger.Log.Error().Err(err).Msg("marshalling failed")
		return errors.New("error")
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Log.Error().Err(err).Msg("building request failed failed")
		return errors.New("could not create request for gandi")
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Apikey "+g.apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error().Err(err).Msg("executing request failed")
		return errors.New("could execute request")
	}

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msg("gandi rejected request")
		return fmt.Errorf("gandi rejected request: %s", string(b))
	}

	return nil
}

func (g *GandiDnsUpdateService) Registrar() Registrar {
	return "gandi"
}
