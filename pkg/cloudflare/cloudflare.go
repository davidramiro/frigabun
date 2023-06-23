package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/davidramiro/frigabun/internal/config"
	"github.com/davidramiro/frigabun/internal/logger"
)

type CloudflareDns struct {
	IP        string
	Domain    string
	Subdomain string
	ZoneId    string
	ApiKey    string
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

type CloudflareUpdateError struct {
	Code    int
	Message string
}

func (c *CloudflareDns) SaveRecord() *CloudflareUpdateError {

	endpoint := fmt.Sprintf("%s/zones/%s/dns_records", config.AppConfig.Cloudflare.BaseUrl,
		c.ZoneId)

	req, err := http.NewRequest("GET", endpoint, nil)

	var r CloudflareQueryResponse

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return &CloudflareUpdateError{Code: 400, Message: "error getting cloudflare request"}
	}

	b, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(b, &r)

	if resp.StatusCode != http.StatusOK || len(r.Errors) > 0 || err != nil {

		logger.Log.Error().Msg("could not query record:" + string(b))
		return &CloudflareUpdateError{400, "could not query record: " + string(b)}
	}

	var id string

	if len(r.Errors) == 0 && len(r.Result) > 0 {
		for _, e := range r.Result {
			if e.Name == c.Subdomain+"."+c.Domain {
				id = e.Id
			}
		}
	}

	if len(id) == 0 {
		return c.NewRecord()
	} else {
		return c.UpdateRecord(id)
	}
}

func (c *CloudflareDns) NewRecord() *CloudflareUpdateError {
	cloudflareRequest := &CloudflareApiRequest{
		Subdomain: fmt.Sprintf("%s.%s", c.Subdomain, c.Domain),
		IP:        c.IP,
		TTL:       config.AppConfig.Cloudflare.TTL,
		Type:      "A",
	}

	endpoint := fmt.Sprintf("%s/zones/%s/dns_records", config.AppConfig.Cloudflare.BaseUrl,
		c.ZoneId)
	logger.Log.Info().Str("subdomain", cloudflareRequest.Subdomain).Str("endpoint", endpoint).Str("IP", cloudflareRequest.IP).Msg("building update request")

	body, err := json.Marshal(cloudflareRequest)
	if err != nil {
		logger.Log.Error().Err(err).Msg("marshalling failed")
		return &CloudflareUpdateError{Code: 400, Message: "could not parse request"}
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Log.Error().Err(err).Msg("building request failed failed")
		return &CloudflareUpdateError{Code: 400, Message: "could not create request for cloudflare"}
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error().Err(err).Msg("executing request failed")
		return &CloudflareUpdateError{Code: 500, Message: "could execute request"}
	}

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msg("gandi rejected request")
		return &CloudflareUpdateError{Code: resp.StatusCode, Message: "cloudflare rejected request: " + string(b)}
	}

	return nil
}

func (c *CloudflareDns) UpdateRecord(id string) *CloudflareUpdateError {
	cloudflareRequest := &CloudflareApiRequest{
		Subdomain: fmt.Sprintf("%s.%s", c.Subdomain, c.Domain),
		IP:        c.IP,
		TTL:       config.AppConfig.Cloudflare.TTL,
		Type:      "A",
	}

	endpoint := fmt.Sprintf("%s/zones/%s/dns_records/%s", config.AppConfig.Cloudflare.BaseUrl,
		c.ZoneId, id)
	logger.Log.Info().Str("subdomain", cloudflareRequest.Subdomain).Str("endpoint", endpoint).Str("IP", cloudflareRequest.IP).Msg("building update request")

	body, err := json.Marshal(cloudflareRequest)
	if err != nil {
		logger.Log.Error().Err(err).Msg("marshalling failed")
		return &CloudflareUpdateError{Code: 400, Message: "could not parse request"}
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Log.Error().Err(err).Msg("building request failed failed")
		return &CloudflareUpdateError{Code: 400, Message: "could not create request for cloudflare"}
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error().Err(err).Msg("executing request failed")
		return &CloudflareUpdateError{Code: 500, Message: "could execute request"}
	}

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msg("gandi rejected request")
		return &CloudflareUpdateError{Code: resp.StatusCode, Message: "cloudflare rejected request: " + string(b)}
	}

	return nil
}
