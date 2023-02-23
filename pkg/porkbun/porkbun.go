package porkbun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/davidramiro/frigabun/internal/config"
	"github.com/davidramiro/frigabun/internal/logger"
)

type PorkbunDns struct {
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

type PorkbunQueryResponse struct {
	Status  string `json:"status"`
	Records []struct {
		Name string `json:"name"`
	} `json:"records"`
}

type PorkbunUpdateError struct {
	Code    int
	Message string
}

func (p *PorkbunDns) AddRecord() *PorkbunUpdateError {

	porkbunRequest := &PorkbunApiRequest{
		Subdomain:    p.Subdomain,
		IP:           p.IP,
		TTL:          config.AppConfig.Porkbun.TTL,
		Type:         "A",
		ApiKey:       p.ApiKey,
		SecretApiKey: p.SecretApiKey,
	}

	queryErr := porkbunRequest.queryRecord(p)

	if queryErr != nil && queryErr.Code == 409 {

		logger.Log.Info().Msg("record exists, updating")
		updateErr := porkbunRequest.updateRecord(p)

		if updateErr != nil {
			logger.Log.Error().Str("err", updateErr.Message).Msg("porkbun rejected updated record")
			return &PorkbunUpdateError{Code: 400, Message: updateErr.Message}
		}

	} else {
		createErr := porkbunRequest.createRecord(p)
		if createErr != nil {
			logger.Log.Error().Str("err", createErr.Message).Msg("porkbun rejected new record")
			return &PorkbunUpdateError{Code: 400, Message: createErr.Message}
		}
	}

	return nil
}

func (p *PorkbunApiRequest) queryRecord(dns *PorkbunDns) *PorkbunUpdateError {
	endpoint := fmt.Sprintf("%s/dns/retrieveByNameType/%s/A/%s", config.AppConfig.Porkbun.BaseUrl, dns.Domain, dns.Subdomain)

	logger.Log.Info().Str("subdomain", p.Subdomain).Str("endpoint", endpoint).Str("IP", p.IP).Msg("checking if record exists")

	var r PorkbunQueryResponse

	resp, updErr := executeRequest(endpoint, p)
	if updErr != nil {
		return updErr
	}

	b, _ := io.ReadAll(resp.Body)
	err := json.Unmarshal(b, &r)

	if resp.StatusCode != http.StatusOK || r.Status != "SUCCESS" || err != nil {

		logger.Log.Error().Msg("could not query record:" + string(b))
		return &PorkbunUpdateError{400, "could not query record: " + string(b)}
	}

	if r.Status == "SUCCESS" && len(r.Records) > 0 {
		for _, e := range r.Records {
			if e.Name == dns.Subdomain+"."+dns.Domain {
				return &PorkbunUpdateError{409, "record already exists"}
			}
		}
	}

	return nil
}

func (p *PorkbunApiRequest) createRecord(dns *PorkbunDns) *PorkbunUpdateError {
	endpoint := fmt.Sprintf("%s/dns/create/%s", config.AppConfig.Porkbun.BaseUrl, dns.Domain)

	logger.Log.Info().Str("subdomain", p.Subdomain).Str("endpoint", endpoint).Str("IP", p.IP).Msg("creating new record")

	resp, err := executeRequest(endpoint, p)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return &PorkbunUpdateError{400, "could not create record: " + string(b)}
	}

	return nil
}

func (p *PorkbunApiRequest) updateRecord(dns *PorkbunDns) *PorkbunUpdateError {
	endpoint := fmt.Sprintf("%s/dns/editByNameType/%s/A/%s", config.AppConfig.Porkbun.BaseUrl, dns.Domain, dns.Subdomain)

	logger.Log.Info().Str("subdomain", p.Subdomain).Str("endpoint", endpoint).Str("IP", p.IP).Msg("updating record")

	resp, err := executeRequest(endpoint, p)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return &PorkbunUpdateError{400, "could not update record: " + string(b)}
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
