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

func setupCloudflareConfig() {
	viper.Set("cloudflare.enabled", true)
	viper.Set("cloudflare.baseurl", "https://api.foo.com/client/v4")
	viper.Set("cloudflare.apiKey", "foo")
	viper.Set("cloudflare.zoneId", "bar")
	viper.Set("cloudflare.ttl", 42)
}

func TestNewCloudflareDnsUpdateServiceSuccess(t *testing.T) {
	setupCloudflareConfig()
	registrar, err := services.NewCloudflareDnsUpdateService(nil)
	assert.Nil(t, err)
	assert.NotNil(t, registrar)
}

func TestNewCloudflareDnsUpdateMissingParam(t *testing.T) {
	setupCloudflareConfig()
	viper.Set("cloudflare.baseUrl", "")
	registrar, err := services.NewCloudflareDnsUpdateService(nil)
	assert.ErrorIs(t, err, services.ErrMissingInfoForServiceInit)
	assert.Nil(t, registrar)
}

func TestCloudflareDnsUpdateService_Registrar(t *testing.T) {
	setupCloudflareConfig()
	registrar, _ := services.NewCloudflareDnsUpdateService(nil)
	assert.Equal(t, services.Registrar("cloudflare"), registrar.Registrar())
}

func TestCloudflareDnsUpdateService_UpdateRecord_RequestError(t *testing.T) {
	setupCloudflareConfig()
	h := mockservices.NewMockHTTPClient(t)
	h.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, errors.New("cf api request error")).Once()

	registrar, err := services.NewCloudflareDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	req := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(req)

	assert.Errorf(t, err, "cf api request error")
}

func TestCloudflareDnsUpdateService_UpdateRecord_QueryError(t *testing.T) {
	setupCloudflareConfig()
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

	registrar, err := services.NewCloudflareDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	req := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(req)

	assert.EqualError(t, err, "could not query record: {\"errors\":[{\"message\":\"error\"}],\"result\":null}")
}

func TestCloudflareDnsUpdateService_UpdateRecord_ExistingRecord(t *testing.T) {
	setupCloudflareConfig()
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
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       nil,
	}, nil).Once()

	registrar, err := services.NewCloudflareDnsUpdateService(h)
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

func TestCloudflareDnsUpdateService_UpdateRecord_NewRecord(t *testing.T) {
	setupCloudflareConfig()
	h := mockservices.NewMockHTTPClient(t)

	resp := &services.CloudflareQueryResponse{
		Errors: []struct {
			Message string `json:"message"`
		}{},
		Result: []struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		}{},
	}

	jsonBytes, err := json.Marshal(resp)
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

	registrar, err := services.NewCloudflareDnsUpdateService(h)
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

func TestCloudflareDnsUpdateService_UpdateRecord_NewRecord_ApiError(t *testing.T) {
	setupCloudflareConfig()
	h := mockservices.NewMockHTTPClient(t)

	resp := &services.CloudflareQueryResponse{
		Errors: []struct {
			Message string `json:"message"`
		}{},
		Result: []struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		}{},
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal()
	}

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("api error")),
	}, nil).Once()

	registrar, err := services.NewCloudflareDnsUpdateService(h)
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

func TestCloudflareDnsUpdateService_UpdateRecord_ExistingRecord_ApiError(t *testing.T) {
	setupCloudflareConfig()
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
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("api error")),
	}, nil).Once()

	registrar, err := services.NewCloudflareDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	dynReq := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(dynReq)

	assert.EqualError(t, err, "cloudflare rejected request: api error")
}

func TestCloudflareDnsUpdateService_UpdateRecord_ExistingRecord_RequestError(t *testing.T) {
	setupCloudflareConfig()
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
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	h.On("Do", mock.Anything).Return(nil, errors.New("error on request")).Once()

	registrar, err := services.NewCloudflareDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	dynReq := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(dynReq)

	assert.EqualError(t, err, "could not execute request")
}

func TestCloudflareDnsUpdateService_UpdateRecord_NewRecord_RequestError(t *testing.T) {
	setupCloudflareConfig()
	h := mockservices.NewMockHTTPClient(t)

	resp := &services.CloudflareQueryResponse{
		Errors: []struct {
			Message string `json:"message"`
		}{},
		Result: []struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		}{},
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal()
	}

	h.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
	}, nil).Once()

	h.On("Do", mock.Anything).Return(nil, errors.New("error on request")).Once()

	registrar, err := services.NewCloudflareDnsUpdateService(h)
	if err != nil {
		t.Fatal(err)
	}

	dynReq := &services.DynDnsRequest{
		Subdomain: "bar",
		Domain:    "foo.com",
		IP:        "1.2.3.4",
	}

	err = registrar.UpdateRecord(dynReq)

	assert.EqualError(t, err, services.ErrExecutingRequest.Error())
}
