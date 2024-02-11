package services_test

import (
	"bytes"
	"encoding/json"
	"errors"
	mockservices "github.com/davidramiro/frigabun/mocks/github.com/davidramiro/frigabun/services"
	"github.com/davidramiro/frigabun/services"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"testing"
)

func setupGandiConfig() {
	viper.Set("gandi.enabled", true)
	viper.Set("gandi.baseurl", "https://api.foo.com/client/v4")
	viper.Set("gandi.apiKey", "foo")
	viper.Set("gandi.ttl", 42)
}

func TestNewGandiDnsUpdateServiceSuccess(t *testing.T) {
	setupGandiConfig()
	registrar, err := services.NewGandiDnsUpdateService(nil)
	assert.Nil(t, err)
	assert.NotNil(t, registrar)
}

func TestNewGandiDnsUpdateMissingParam(t *testing.T) {
	setupGandiConfig()
	viper.Set("gandi.baseUrl", "")
	registrar, err := services.NewGandiDnsUpdateService(nil)
	assert.ErrorIs(t, err, services.ErrMissingInfoForServiceInit)
	assert.Nil(t, registrar)
}

func TestGandiDnsUpdateService_Registrar(t *testing.T) {
	setupGandiConfig()
	registrar, _ := services.NewGandiDnsUpdateService(nil)
	assert.Equal(t, services.Registrar("gandi"), registrar.Registrar())
}

func TestGandiDnsUpdateService_UpdateRecord_RequestError(t *testing.T) {
	setupGandiConfig()
	h := mockservices.NewMockHTTPClient(t)
	h.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, errors.New("gd api request error")).Once()

	registrar, err := services.NewGandiDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	req := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(req)

	assert.Errorf(t, err, "gd api request error")
}

func TestGandiDnsUpdateService_UpdateRecord_ApiError(t *testing.T) {
	setupGandiConfig()
	h := mockservices.NewMockHTTPClient(t)

	resp := &services.CloudflareQueryResponse{
		Errors: []struct {
			Message string `json:"message"`
		}{{Message: "error"}},
		Result: nil,
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal()
	}

	h.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	registrar, err := services.NewGandiDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	req := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(req)

	assert.EqualError(t, err, "gandi rejected request: {\"errors\":[{\"message\":\"error\"}],\"result\":null}")
}

func TestGandiDnsUpdateService_UpdateRecord_Success(t *testing.T) {
	setupGandiConfig()
	h := mockservices.NewMockHTTPClient(t)

	resp := &services.CloudflareQueryResponse{
		Errors: []struct {
			Message string `json:"message"`
		}{},
		Result: []struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		}{{Name: "bar.foo.com", Id: "1"}},
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal()
	}

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	registrar, err := services.NewGandiDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	dynReq := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(dynReq)

	assert.Nil(t, err)
}
