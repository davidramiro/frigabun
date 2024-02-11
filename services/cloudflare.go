package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"net/http"
)

type CloudflareDnsUpdateService struct {
	registrarSettings
	apiKey string
	zoneId string
	client HTTPClient
}

func NewCloudflareDnsUpdateService(client HTTPClient) (*CloudflareDnsUpdateService, error) {
	baseUrl := viper.GetString("cloudflare.baseUrl")
	ttl := viper.GetInt("cloudflare.ttl")
	apikey := viper.GetString("cloudflare.apiKey")
	zoneId := viper.GetString("cloudflare.zoneId")

	log.Info().Msg("initializing cloudflare service")

	if len(baseUrl) == 0 || ttl == 0 || len(apikey) == 0 || len(zoneId) == 0 {
		return nil, ErrMissingInfoForServiceInit
	}

	if client == nil {
		client = &http.Client{}
	}

	return &CloudflareDnsUpdateService{
		registrarSettings: registrarSettings{
			baseUrl: baseUrl,
			ttl:     ttl,
		},
		apiKey: apikey,
		zoneId: zoneId,
		client: client,
	}, nil
}

type CloudflareApiRequest struct {
	Subdomain string `json:"name"`
	Type      string `json:"type"`
	TTL       int    `json:"ttl"`
	IP        string `json:"content"`
}

type CloudflareQueryResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Result []struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	} `json:"result"`
}

func (c *CloudflareDnsUpdateService) UpdateRecord(request *DynDnsRequest) error {

	endpoint := fmt.Sprintf("%s/zones/%s/dns_records", c.baseUrl,
		c.zoneId)

	logger := log.With().Str("func", "UpdateRecord").Str("registrar", "cloudflare").Str("endpoint", endpoint).Str("domain", request.Domain).Str("subdomain", request.Subdomain).Logger()
	logger.Info().Msg("building update request")

	req, err := http.NewRequest("GET", endpoint, nil)

	var r CloudflareQueryResponse

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg(ErrBuildingRequest.Error())
		return ErrBuildingRequest
	}

	b, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(b, &r)

	if err != nil {
		logger.Error().Err(err).Msg(ErrParsingResponse.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK || len(r.Errors) > 0 {
		logger.Error().Interface("response", b).Msg("could not query record")
		return errors.New("could not query record: " + string(b))
	}

	var id string

	if len(r.Errors) == 0 && len(r.Result) > 0 {
		for _, e := range r.Result {
			if e.Name == fmt.Sprintf("%s.%s", request.Subdomain, request.Domain) {
				id = e.Id
			}
		}
	}

	if len(id) == 0 {
		logger.Info().Msg("entry not found, creating new")
		return c.newRecord(request)
	} else {
		logger.Info().Msg("entry found, updating")
		return c.editExistingRecord(request, id)
	}
}

func (c *CloudflareDnsUpdateService) newRecord(request *DynDnsRequest) error {
	cloudflareRequest := &CloudflareApiRequest{
		Subdomain: fmt.Sprintf("%s.%s", request.Subdomain, request.Domain),
		IP:        request.IP,
		TTL:       c.ttl,
		Type:      "A",
	}

	endpoint := fmt.Sprintf("%s/zones/%s/dns_records", c.baseUrl,
		c.zoneId)

	logger := log.With().Str("func", "newRecord").Str("registrar", "cloudflare").Str("subdomain", cloudflareRequest.Subdomain).Str("endpoint", endpoint).Str("IP", cloudflareRequest.IP).Logger()
	logger.Info().Msg("building new record request")

	body, err := json.Marshal(cloudflareRequest)
	if err != nil {
		logger.Error().Err(err).Msg(ErrBuildingRequest.Error())
		return ErrBuildingRequest
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Error().Err(err).Msg(ErrBuildingRequest.Error())
		return ErrBuildingRequest
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	logger.Info().Msg("executing request")
	resp, err := c.client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg(ErrExecutingRequest.Error())
		return ErrExecutingRequest
	}

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		logger.Error().Bytes("response", b).Msg(ErrRegistrarRejectedRequest.Error())
		return ErrRegistrarRejectedRequest
	}

	logger.Info().Msg("request for new record successful")

	return nil
}

func (c *CloudflareDnsUpdateService) editExistingRecord(request *DynDnsRequest, id string) error {
	cloudflareRequest := &CloudflareApiRequest{
		Subdomain: fmt.Sprintf("%s.%s", request.Subdomain, request.Domain),
		IP:        request.IP,
		TTL:       c.ttl,
		Type:      "A",
	}

	endpoint := fmt.Sprintf("%s/zones/%s/dns_records/%s", c.baseUrl,
		c.zoneId, id)

	logger := log.With().Str("func", "editExistingRecord").Str("registrar", "cloudflare").Str("subdomain", cloudflareRequest.Subdomain).Str("endpoint", endpoint).Str("IP", cloudflareRequest.IP).Logger()
	logger.Info().Msg("building request to edit record")

	body, err := json.Marshal(cloudflareRequest)
	if err != nil {
		logger.Error().Err(err).Msg("marshalling failed")
		return errors.New("could not parse request")
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Error().Err(err).Msg("building request failed failed")
		return errors.New("could not create request for cloudflare")
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("executing request failed")
		return errors.New("could not execute request")
	}

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		logger.Error().Bytes("response", b).Msg("cloudflare rejected request")
		return fmt.Errorf("cloudflare rejected request: %s", string(b))
	}

	return nil
}

func (c *CloudflareDnsUpdateService) Registrar() Registrar {
	return "cloudflare"
}
