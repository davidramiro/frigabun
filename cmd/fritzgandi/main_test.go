package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func init() {
	// make sure we're in project root for tests
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	initConfig()
}

func TestInitConfig(t *testing.T) {
	assert.NotNil(t, configuration, "initConfig should fill config object")
}

func TestValidDomainAndIp(t *testing.T) {
	err := validateRequest("domain.com", "1.1.1.1")

	assert.Nil(t, err, "valid domain and ip should not return error")
}

func TestInvalidIPv4(t *testing.T) {
	err := validateRequest("domain.com", "::1")

	assert.Equal(t, 400, err.Code, "invalid IPv4 should return error")
}

func TestInvalidDomain(t *testing.T) {
	err := validateRequest("domain .com", "1.1.1.1")

	assert.Equal(t, 400, err.Code, "invalid domain should return error")
}

func TestUpdateEndpointWithInvalidApiKey(t *testing.T) {
	initConfig()

	// Setup
	q := make(url.Values)
	q.Set("ip", "1.2.3.4")
	q.Set("domain", "domain.com")
	q.Set("subdomain", "test1,test2")
	q.Set("apiKey", "foo")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, handleUpdateRequest(c)) {
		assert.Equal(t, http.StatusForbidden, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "gandi rejected request")
	}
}

func TestUpdateEndpointWithInvalidIp(t *testing.T) {
	q := make(url.Values)
	q.Set("ip", "::1")
	q.Set("domain", "domain.com")
	q.Set("subdomain", "test1,test2")
	q.Set("apiKey", "foo")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, handleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "invalid IP address")
	}
}

func TestUpdateEndpointWithMissingParam(t *testing.T) {
	q := make(url.Values)
	q.Set("ip", "1.2.3.4")
	q.Set("subdomain", "test1,test2")
	q.Set("apiKey", "foo")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, handleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "missing or invalid domain name")
	}
}

func TestUpdateEndpointWithInvalidMissingSubdomains(t *testing.T) {
	q := make(url.Values)
	q.Set("ip", "1.2.3.4")
	q.Set("domain", "domain.com")
	q.Set("apiKey", "foo")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, handleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "missing subdomains parameter")
	}
}

func TestStatusEndpoint(t *testing.T) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, handleStatusCheck(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "\"api_status\":true")
	}
}
