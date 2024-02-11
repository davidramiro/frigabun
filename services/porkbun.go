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

type PorkbunDnsUpdateService struct {
	registrarSettings
	apiKey       string
	apiSecretKey string
}

func NewPorkbunDnsUpdateService() (*PorkbunDnsUpdateService, error) {
	baseUrl := viper.GetString("porkbun.baseUrl")
	ttl := viper.GetInt("gandi.ttl")
	apikey := viper.GetString("porkbun.apiKey")
	apiSecretkey := viper.GetString("porkbun.apiSecretKey")
	if len(baseUrl) == 0 || ttl == 0 || len(apikey) == 0 {
		return nil, fmt.Errorf(ErrMissingInfoForServiceInit, "cloudflare")
	}

	return &PorkbunDnsUpdateService{
		registrarSettings: registrarSettings{
			baseUrl: baseUrl,
			ttl:     ttl,
		},
		apiKey:       apikey,
		apiSecretKey: apiSecretkey,
	}, nil
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

func (p *PorkbunDnsUpdateService) UpdateRecord(request *DynDnsRequest) error {

	porkbunRequest := &PorkbunApiRequest{
		Subdomain:    request.Subdomain,
		IP:           request.IP,
		TTL:          p.ttl,
		Type:         "A",
		ApiKey:       p.apiKey,
		SecretApiKey: p.apiSecretKey,
	}

	exists, err := p.queryRecord(request, porkbunRequest)

	if err != nil {
		return err
	}

	if exists {

		log.Info().Msg("record exists, updating")
		err := p.updateRecord(request, porkbunRequest)

		if err != nil {
			log.Error().Err(err).Msg("porkbun rejected updated record")
			return errors.New("porkbun rejected updated record")
		}

	} else {
		err := p.createRecord(request, porkbunRequest)

		if err != nil {
			log.Error().Err(err).Msg("porkbun rejected new record")
			return errors.New("porkbun rejected new record")
		}
	}

	return nil
}

func (p *PorkbunDnsUpdateService) queryRecord(request *DynDnsRequest, porkbunRequest *PorkbunApiRequest) (bool, error) {
	endpoint := fmt.Sprintf("%s/dns/retrieveByNameType/%s/A/%s", p.baseUrl, request.Domain, request.Subdomain)

	log.Info().Str("subdomain", request.Subdomain).Str("endpoint", endpoint).Str("IP", request.IP).Msg("checking if record exists")

	var r PorkbunQueryResponse

	resp, err := p.executeRequest(endpoint, porkbunRequest)
	if err != nil {
		return false, err
	}

	b, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(b, &r)

	if resp.StatusCode != http.StatusOK || r.Status != "SUCCESS" || err != nil {
		log.Error().Msg("could not query record:" + string(b))
		return false, fmt.Errorf("could not query record: %s", string(b))
	}

	if r.Status == "SUCCESS" && len(r.Records) > 0 {
		for _, e := range r.Records {
			if e.Name == fmt.Sprintf("%s.%s", request.Subdomain, request.Domain) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (p *PorkbunDnsUpdateService) createRecord(request *DynDnsRequest, porkbunRequest *PorkbunApiRequest) error {
	endpoint := fmt.Sprintf("%s/dns/create/%s", p.baseUrl, request.Domain)

	log.Info().Str("subdomain", request.Subdomain).Str("endpoint", endpoint).Str("IP", request.IP).Msg("creating new record")

	resp, err := p.executeRequest(endpoint, porkbunRequest)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("could not create record: %s", string(b))
	}

	return nil
}

func (p *PorkbunDnsUpdateService) updateRecord(request *DynDnsRequest, porkbunRequest *PorkbunApiRequest) error {
	endpoint := fmt.Sprintf("%s/dns/editByNameType/%s/A/%s", p.baseUrl, request.Domain, request.Subdomain)

	log.Info().Str("subdomain", request.Subdomain).Str("endpoint", endpoint).Str("IP", request.IP).Msg("updating record")

	resp, err := p.executeRequest(endpoint, porkbunRequest)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("could not update record: %s", string(b))
	}

	return nil
}

func (p *PorkbunDnsUpdateService) executeRequest(endpoint string, porkbunRequest *PorkbunApiRequest) (*http.Response, error) {
	body, err := json.Marshal(porkbunRequest)
	if err != nil {
		log.Error().Err(err).Msg("marshalling failed")
		return nil, errors.New("could not parse request")
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("building request failed failed")
		return nil, errors.New("could not create request for porkbun")
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("executing request failed")
		return nil, errors.New("could not execute request")
	}

	return resp, nil
}

func (p *PorkbunDnsUpdateService) Registrar() Registrar {
	return "porkbun"
}
