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
	"strings"
	"testing"
)

func setupPorkbunConfig() {
	viper.Set("porkbun.enabled", true)
	viper.Set("porkbun.baseurl", "https://api.foo.com/client/v4")
	viper.Set("porkbun.apiKey", "foo")
	viper.Set("porkbun.secretApiKey", "bar")
	viper.Set("porkbun.ttl", 42)
}

func TestNewPorkbunDnsUpdateServiceSuccess(t *testing.T) {
	setupPorkbunConfig()
	registrar, err := services.NewPorkbunDnsUpdateService(nil)
	assert.Nil(t, err)
	assert.NotNil(t, registrar)
}

func TestNewPorkbunDnsUpdateMissingParam(t *testing.T) {
	setupPorkbunConfig()
	viper.Set("porkbun.baseUrl", "")
	registrar, err := services.NewPorkbunDnsUpdateService(nil)
	assert.ErrorIs(t, err, services.ErrMissingInfoForServiceInit)
	assert.Nil(t, registrar)
}

func TestPorkbunDnsUpdateService_Registrar(t *testing.T) {
	setupPorkbunConfig()
	registrar, _ := services.NewPorkbunDnsUpdateService(nil)
	assert.Equal(t, services.Registrar("porkbun"), registrar.Registrar())
}

func TestPorkbunDnsUpdateService_UpdateRecord_RequestError(t *testing.T) {
	setupPorkbunConfig()
	h := mockservices.NewMockHTTPClient(t)
	h.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, errors.New("pb api request error")).Once()

	registrar, err := services.NewPorkbunDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	req := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(req)

	assert.Errorf(t, err, "pb api request error")
}

func TestPorkbunDnsUpdateService_UpdateRecord_ApiError(t *testing.T) {
	setupPorkbunConfig()
	h := mockservices.NewMockHTTPClient(t)

	resp := &services.PorkbunQueryResponse{
		Status: "ERROR",
		Records: []struct {
			Name string `json:"name"`
		}{},
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal()
	}

	h.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	registrar, err := services.NewPorkbunDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	req := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(req)

	assert.EqualError(t, err, services.ErrRegistrarRejectedRequest.Error())
}

func TestPorkbunDnsUpdateService_UpdateRecord_Exists_Success(t *testing.T) {
	setupPorkbunConfig()
	h := mockservices.NewMockHTTPClient(t)

	queryResp := &services.PorkbunQueryResponse{
		Status: "SUCCESS",
		Records: []struct {
			Name string `json:"name"`
		}{{Name: "bar.foo.com"}},
	}

	jsonBytes, err := json.Marshal(queryResp)
	if err != nil {
		t.Fatal()
	}

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       nil,
	}, nil).Once()

	registrar, err := services.NewPorkbunDnsUpdateService(h)
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

func TestPorkbunDnsUpdateService_UpdateRecord_Exists_Failure_On_Update(t *testing.T) {
	setupPorkbunConfig()
	h := mockservices.NewMockHTTPClient(t)

	queryResp := &services.PorkbunQueryResponse{
		Status: "SUCCESS",
		Records: []struct {
			Name string `json:"name"`
		}{{Name: "bar.foo.com"}},
	}

	jsonBytes, err := json.Marshal(queryResp)
	if err != nil {
		t.Fatal()
	}

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("error updating")),
	}, nil).Once()

	registrar, err := services.NewPorkbunDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	dynReq := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(dynReq)

	assert.EqualError(t, err, services.ErrRegistrarRejectedRequest.Error())
}

func TestPorkbunDnsUpdateService_UpdateRecord_NotExists_Failure_On_Create(t *testing.T) {
	setupPorkbunConfig()
	h := mockservices.NewMockHTTPClient(t)

	queryResp := &services.PorkbunQueryResponse{
		Status: "SUCCESS",
		Records: []struct {
			Name string `json:"name"`
		}{},
	}

	jsonBytes, err := json.Marshal(queryResp)
	if err != nil {
		t.Fatal()
	}

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("error creating")),
	}, nil).Once()

	registrar, err := services.NewPorkbunDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	dynReq := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(dynReq)

	assert.EqualError(t, err, services.ErrRegistrarRejectedRequest.Error())
}
