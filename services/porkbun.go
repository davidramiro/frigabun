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

type PorkbunDnsUpdateService struct {
	registrarSettings
	apiKey       string
	secretApiKey string
	client       HTTPClient
}

func NewPorkbunDnsUpdateService(client HTTPClient) (*PorkbunDnsUpdateService, error) {
	baseUrl := viper.GetString("porkbun.baseUrl")
	ttl := viper.GetInt("porkbun.ttl")
	apikey := viper.GetString("porkbun.apiKey")
	SecretApiKey := viper.GetString("porkbun.secretApiKey")

	log.Info().Msg("initializing porkbun service")

	if len(baseUrl) == 0 || ttl == 0 || len(apikey) == 0 {
		return nil, ErrMissingInfoForServiceInit
	}

	if client == nil {
		client = &http.Client{}
	}

	return &PorkbunDnsUpdateService{
		registrarSettings: registrarSettings{
			baseUrl: baseUrl,
			ttl:     ttl,
		},
		apiKey:       apikey,
		secretApiKey: SecretApiKey,
		client:       client,
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

func (p *PorkbunDnsUpdateService) UpdateRecord(request *DynDnsRequest) error {

	porkbunRequest := &PorkbunApiRequest{
		Subdomain:    request.Subdomain,
		IP:           request.IP,
		TTL:          p.ttl,
		Type:         "A",
		ApiKey:       p.apiKey,
		SecretApiKey: p.secretApiKey,
	}

	logger := log.With().Str("func", "UpdateRecord").Str("registrar", "porkbun").Str("domain", request.Domain).Str("subdomain", request.Subdomain).Logger()
	logger.Info().Msg("building update request")

	exists, err := p.queryRecordExists(request, porkbunRequest)
	if err != nil {
		logger.Err(err).Msg("error querying if record exists")
		return err
	}

	if exists {
		logger.Info().Msg("record exists, updating")
		err := p.updateRecord(request, porkbunRequest)

		if err != nil {
			logger.Error().Err(err).Msg(ErrRegistrarRejectedRequest.Error())
			return ErrRegistrarRejectedRequest
		}

	} else {
		logger.Info().Msg("record does not exist, creating")
		err := p.createRecord(request, porkbunRequest)

		if err != nil {
			logger.Error().Err(err).Msg(ErrRegistrarRejectedRequest.Error())
			return ErrRegistrarRejectedRequest
		}
	}

	logger.Info().Msg("update request successful")

	return nil
}

func (p *PorkbunDnsUpdateService) queryRecordExists(request *DynDnsRequest, porkbunRequest *PorkbunApiRequest) (bool, error) {
	endpoint := fmt.Sprintf("%s/dns/retrieveByNameType/%s/A/%s", p.baseUrl, request.Domain, request.Subdomain)

	logger := log.With().Str("func", "createRecord").Str("registrar", "porkbun").Str("subdomain", request.Subdomain).Str("endpoint", endpoint).Str("IP", request.IP).Logger()
	logger.Info().Msg("query for existing record")

	var r PorkbunQueryResponse

	resp, err := p.executeRequest(endpoint, porkbunRequest)
	if err != nil {
		return false, err
	}

	b, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(b, &r)

	if resp.StatusCode != http.StatusOK || r.Status != "SUCCESS" || err != nil {
		logger.Error().Bytes("response", b).Msg(ErrRegistrarRejectedRequest.Error())
		return false, ErrRegistrarRejectedRequest
	}

	if r.Status == "SUCCESS" && len(r.Records) > 0 {
		for _, e := range r.Records {
			logger.Info().Msg("record found")
			if e.Name == fmt.Sprintf("%s.%s", request.Subdomain, request.Domain) {
				return true, nil
			}
		}
	}

	logger.Info().Msg("record not found")
	return false, nil
}

func (p *PorkbunDnsUpdateService) createRecord(request *DynDnsRequest, porkbunRequest *PorkbunApiRequest) error {
	endpoint := fmt.Sprintf("%s/dns/create/%s", p.baseUrl, request.Domain)

	logger := log.With().Str("func", "createRecord").Str("registrar", "porkbun").Str("subdomain", request.Subdomain).Str("endpoint", endpoint).Str("IP", request.IP).Logger()
	logger.Info().Msg("creating record")

	resp, err := p.executeRequest(endpoint, porkbunRequest)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		logger.Error().Bytes("response", b).Msg(ErrRegistrarRejectedRequest.Error())
		return ErrRegistrarRejectedRequest
	}

	return nil
}

func (p *PorkbunDnsUpdateService) updateRecord(request *DynDnsRequest, porkbunRequest *PorkbunApiRequest) error {
	endpoint := fmt.Sprintf("%s/dns/editByNameType/%s/A/%s", p.baseUrl, request.Domain, request.Subdomain)

	logger := log.With().Str("func", "updateRecord").Str("registrar", "porkbun").Str("subdomain", request.Subdomain).Str("endpoint", endpoint).Str("IP", request.IP).Logger()
	logger.Info().Msg("updating record")

	resp, err := p.executeRequest(endpoint, porkbunRequest)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		logger.Error().Bytes("response", b).Msg(ErrRegistrarRejectedRequest.Error())
		return ErrRegistrarRejectedRequest
	}

	return nil
}

func (p *PorkbunDnsUpdateService) executeRequest(endpoint string, porkbunRequest *PorkbunApiRequest) (*http.Response, error) {
	logger := log.With().Str("func", "executeRequest").Str("registrar", "porkbun").Str("endpoint", endpoint).Str("subdomain", porkbunRequest.Subdomain).Logger()
	logger.Info().Msg("building update request")

	body, err := json.Marshal(porkbunRequest)
	if err != nil {
		logger.Error().Err(err).Msg(ErrBuildingRequest.Error())
		return nil, ErrBuildingRequest
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		logger.Error().Err(err).Msg(ErrBuildingRequest.Error())
		return nil, ErrBuildingRequest
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := p.client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg(ErrExecutingRequest.Error())
		return nil, ErrExecutingRequest
	}

	logger.Info().Msg("request successful")

	return resp, nil
}

func (p *PorkbunDnsUpdateService) Registrar() Registrar {
	return "porkbun"
}
