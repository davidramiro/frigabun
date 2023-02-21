package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/davidramiro/frigabun/internal/config"
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

	config.InitConfig()
}

func TestStatusEndpoint(t *testing.T) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, HandleStatusCheck(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "\"api_status\":true")
	}
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

func TestGandiUpdateWithValidRequest(t *testing.T) {
	q := make(url.Values)
	q.Set("ip", config.AppConfig.Test.Gandi.IP)
	q.Set("domain", config.AppConfig.Test.Gandi.Domain)
	q.Set("subdomain", config.AppConfig.Test.Gandi.Subdomain+"2")
	q.Set("apiKey", config.AppConfig.Test.Gandi.ApiKey)
	q.Set("registrar", "gandi")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "created")
	}
}

func TestPorkbunUpdateWithValidRequest(t *testing.T) {
	q := make(url.Values)
	q.Set("ip", config.AppConfig.Test.Porkbun.IP)
	q.Set("domain", config.AppConfig.Test.Porkbun.Domain)
	q.Set("subdomain", config.AppConfig.Test.Porkbun.Subdomain+"2")
	q.Set("apikey", config.AppConfig.Test.Porkbun.ApiKey)
	q.Set("apisecretkey", config.AppConfig.Test.Porkbun.ApiSecretKey)
	q.Set("registrar", "porkbun")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "created")
	}
}

func TestGandiUpdateWithInvalidIp(t *testing.T) {
	q := make(url.Values)
	q.Set("ip", "::1")
	q.Set("domain", "domain.com")
	q.Set("subdomain", "test1,test2")
	q.Set("apiKey", "foo")
	q.Set("registrar", "gandi")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "invalid IP address")
	}
}

func TestGandiUpdateWithMissingParam(t *testing.T) {
	q := make(url.Values)
	q.Set("ip", "1.2.3.4")
	q.Set("subdomain", "test1,test2")
	q.Set("apiKey", "foo")
	q.Set("registrar", "gandi")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "missing or invalid domain name")
	}
}

func TestGandiUpdateWithInvalidMissingSubdomains(t *testing.T) {
	q := make(url.Values)
	q.Set("ip", "1.2.3.4")
	q.Set("domain", "domain.com")
	q.Set("apiKey", "foo")
	q.Set("registrar", "gandi")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "missing subdomains parameter")
	}
}

func TestGandiUpdateWithInvalidApiKey(t *testing.T) {
	// Setup
	q := make(url.Values)
	q.Set("ip", "1.2.3.4")
	q.Set("domain", "domain.com")
	q.Set("subdomain", "test1,test2")
	q.Set("apiKey", "foo")
	q.Set("registrar", "gandi")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusForbidden, rec.Code)
		b, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(b), "gandi rejected request")
	}
}
