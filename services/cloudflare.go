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
}

func NewCloudflareDnsUpdateService() (*CloudflareDnsUpdateService, error) {
	baseUrl := viper.GetString("cloudflare.baseUrl")
	ttl := viper.GetInt("cloudflare.ttl")
	apikey := viper.GetString("cloudflare.apiKey")
	zoneId := viper.GetString("cloudflare.zoneId")
	if len(baseUrl) == 0 || ttl == 0 || len(apikey) == 0 || len(zoneId) == 0 {
		return nil, fmt.Errorf(ErrMissingInfoForServiceInit, "cloudflare")
	}

	return &CloudflareDnsUpdateService{
		registrarSettings: registrarSettings{
			baseUrl: baseUrl,
			ttl:     ttl,
		},
		apiKey: apikey,
		zoneId: zoneId,
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

	req, err := http.NewRequest("GET", endpoint, nil)

	var r CloudflareQueryResponse

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("error getting cloudflare request")
	}

	b, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(b, &r)

	if resp.StatusCode != http.StatusOK || len(r.Errors) > 0 || err != nil {
		log.Error().Msg("could not query record:" + string(b))
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
		return c.newRecord(request)
	} else {
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
	log.Info().Str("subdomain", cloudflareRequest.Subdomain).Str("endpoint", endpoint).Str("IP", cloudflareRequest.IP).Msg("building update request")

	body, err := json.Marshal(cloudflareRequest)
	if err != nil {
		log.Error().Err(err).Msg("marshalling failed")
		return errors.New("could not parse request")
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("building request failed failed")
		return errors.New("could not create request for cloudflare")
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("executing request failed")
		return errors.New("could execute request")
	}

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		log.Error().Msg("gandi rejected request")
		return fmt.Errorf("cloudflare rejected request: %s", string(b))
	}

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
	log.Info().Str("subdomain", cloudflareRequest.Subdomain).Str("endpoint", endpoint).Str("IP", cloudflareRequest.IP).Msg("building update request")

	body, err := json.Marshal(cloudflareRequest)
	if err != nil {
		log.Error().Err(err).Msg("marshalling failed")
		return errors.New("could not parse request")
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("building request failed failed")
		return errors.New("could not create request for cloudflare")
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("executing request failed")
		return errors.New("could execute request")
	}

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		log.Error().Msg("gandi rejected request")
		return fmt.Errorf("cloudflare rejected request: %s", string(b))
	}

	return nil
}

func (c *CloudflareDnsUpdateService) Registrar() Registrar {
	return "cloudflare"
}
