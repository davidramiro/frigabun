package porkbun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/davidramiro/frigabun/internal/config"
	"github.com/davidramiro/frigabun/internal/logger"
)

type PorkbunDnsInfo struct {
	IP           string
	Domain       string
	Subdomain    string
	ApiKey       string
	SecretApiKey string
}

type PorkbunApiRequest struct {
	Subdomain    string `json:"name"`
	Type         string `json:"type"`
	TTL          int    `json:"ttl"`
	IP           string `json:"content"`
	ApiKey       string `json:"apikey"`
	SecretApiKey string `json:"secretapikey"`
}

type PorkbunUpdateError struct {
	Code    int
	Message string
}

func AddRecord(dnsInfo *PorkbunDnsInfo) *PorkbunUpdateError {

	porkbunRequest := &PorkbunApiRequest{
		Subdomain:    dnsInfo.Subdomain,
		IP:           dnsInfo.IP,
		TTL:          config.AppConfig.Porkbun.TTL,
		Type:         "A",
		ApiKey:       dnsInfo.ApiKey,
		SecretApiKey: dnsInfo.SecretApiKey,
	}

	deleteErr := deleteOldRecord(dnsInfo, porkbunRequest)
	if deleteErr != nil {
		logger.Log.Warn().Msg("deleting old porkbun request failed")
	}

	time.Sleep(2 * time.Second)

	postErr := postNewRecord(dnsInfo, porkbunRequest)
	if postErr != nil {
		logger.Log.Error().Str("err", postErr.Message).Msg("porkbun rejected new record")
		return &PorkbunUpdateError{Code: 400, Message: postErr.Message}
	}

	time.Sleep(2 * time.Second)

	return nil
}

func deleteOldRecord(dnsInfo *PorkbunDnsInfo, porkbunRequest *PorkbunApiRequest) *PorkbunUpdateError {
	endpoint := fmt.Sprintf("%s/dns/deleteByNameType/%s/A/%s", config.AppConfig.Porkbun.BaseUrl, dnsInfo.Domain, dnsInfo.Subdomain)

	logger.Log.Info().Str("subdomain", porkbunRequest.Subdomain).Str("endpoint", endpoint).Str("IP", porkbunRequest.IP).Msg("deleting old record")

	resp, err := executeRequest(endpoint, porkbunRequest)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return &PorkbunUpdateError{400, "could not delete old record"}
	}

	return nil
}

func postNewRecord(dnsInfo *PorkbunDnsInfo, porkbunRequest *PorkbunApiRequest) *PorkbunUpdateError {
	endpoint := fmt.Sprintf("%s/dns/create/%s", config.AppConfig.Porkbun.BaseUrl, dnsInfo.Domain)

	logger.Log.Info().Str("subdomain", porkbunRequest.Subdomain).Str("endpoint", endpoint).Str("IP", porkbunRequest.IP).Msg("creating new record")

	resp, err := executeRequest(endpoint, porkbunRequest)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msg("porkbun rejected request")
		return &PorkbunUpdateError{400, "could not create record: " + string(b)}
	}

	return nil
}

func executeRequest(endpoint string, porkbunRequest *PorkbunApiRequest) (*http.Response, *PorkbunUpdateError) {
	body, err := json.Marshal(porkbunRequest)
	if err != nil {
		logger.Log.Error().Err(err).Msg("marshalling failed")
		return nil, &PorkbunUpdateError{Code: 400, Message: "could not parse request"}
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Log.Error().Err(err).Msg("building request failed failed")
		return nil, &PorkbunUpdateError{Code: 400, Message: "could not create request for gandi"}
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error().Err(err).Msg("executing request failed")
		return nil, &PorkbunUpdateError{Code: 500, Message: "could execute request"}
	}

	return resp, nil
}
