package gandi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/davidramiro/frigabun/internal/config"
	"github.com/davidramiro/frigabun/internal/logger"
)

type GandiDnsInfo struct {
	IP        string
	Domain    string
	Subdomain string
	ApiKey    string
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

func AddRecord(updateRequest *GandiDnsInfo) *GandiUpdateError {

	gandiRequest := &GandiApiRequest{
		Subdomain: updateRequest.Subdomain,
		IPValues:  []string{updateRequest.IP},
		TTL:       config.AppConfig.Gandi.TTL,
		Type:      "A",
	}

	endpoint := fmt.Sprintf("%s/domains/%s/records/%s/A", config.AppConfig.Gandi.BaseUrl,
		updateRequest.Domain, gandiRequest.Subdomain)

	logger.Log.Info().Str("subdomain", gandiRequest.Subdomain).Str("endpoint", endpoint).Str("IP", gandiRequest.IPValues[0]).Msg("building update request")

	body, err := json.Marshal(gandiRequest)
	if err != nil {
		logger.Log.Error().Err(err).Msg("marshalling failed")
		return &GandiUpdateError{Code: 400, Message: "could not parse request"}
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Log.Error().Err(err).Msg("building request failed failed")
		return &GandiUpdateError{Code: 400, Message: "could not create request for gandi"}
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Apikey "+updateRequest.ApiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error().Err(err).Msg("executing request failed")
		return &GandiUpdateError{Code: 500, Message: "could execute request"}
	}

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msg("gandi rejected request")
		return &GandiUpdateError{Code: resp.StatusCode, Message: "gandi rejected request: " + string(b)}
	}

	return nil
}
