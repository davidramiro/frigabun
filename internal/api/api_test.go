package api

import (
	"encoding/json"
	"errors"
	"fmt"
	mock_services "github.com/davidramiro/frigabun/mocks/github.com/davidramiro/frigabun/services"
	mock_factory "github.com/davidramiro/frigabun/mocks/github.com/davidramiro/frigabun/services/factory"
	"github.com/davidramiro/frigabun/services"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var updateApi *UpdateApi

func init() {

}

func TestStatusEndpointOk(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	sf := mock_factory.NewMockServiceFactory(t)
	sf.On("ListServices").Return([]services.Registrar{"cloudflare", "gandi"}).Once()

	updateApi = NewUpdateApi(sf)

	if assert.NoError(t, updateApi.HandleStatusCheck(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var status StatusResponse
		err := json.Unmarshal(rec.Body.Bytes(), &status)

		assert.Nil(t, err)
		assert.Equal(t, true, status.ApiStatus)
		assert.Equal(t, 2, len(status.ActiveServices))
	}
}

func TestUpdateEndpointMissingSubdomain(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("domain", "foo.com")
	q.Set("subdomain", "")
	q.Set("ip", "10.0.0.1")
	q.Set("registrar", "porkbun")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", q.Encode()), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	sf := mock_factory.NewMockServiceFactory(t)
	updateApi = NewUpdateApi(sf)

	if assert.NoError(t, updateApi.HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, ErrMissingParameter.Error(), rec.Body.String())
	}
}

func TestUpdateEndpointInvalidIP(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("domain", "foo.com")
	q.Set("subdomain", "")
	q.Set("ip", "10.0,0.1")
	q.Set("registrar", "porkbun")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", q.Encode()), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	sf := mock_factory.NewMockServiceFactory(t)
	updateApi = NewUpdateApi(sf)

	if assert.NoError(t, updateApi.HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, ErrInvalidIP.Error(), rec.Body.String())
	}
}

func TestUpdateEndpointInvalidRegistrar(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("domain", "foo.com")
	q.Set("subdomain", "bar")
	q.Set("ip", "10.0.0.1")
	q.Set("registrar", "porkbun")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", q.Encode()), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	sf := mock_factory.NewMockServiceFactory(t)
	sf.On("Find", services.Registrar("porkbun")).Return(nil, errors.New("cannot find registrar porkbun"))

	updateApi = NewUpdateApi(sf)

	if assert.NoError(t, updateApi.HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "cannot find registrar porkbun", rec.Body.String())
	}
}

func TestUpdateEndpointFailureInService(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("domain", "foo.com")
	q.Set("subdomain", "bar")
	q.Set("ip", "10.0.0.1")
	q.Set("registrar", "cloudflare")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", q.Encode()), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cs := mock_services.NewMockDnsUpdateService(t)
	cs.On("UpdateRecord", mock.Anything).Return(errors.New("failed to update")).Once()

	sf := mock_factory.NewMockServiceFactory(t)
	sf.On("Find", services.Registrar("cloudflare")).Return(cs, nil).Once()

	updateApi = NewUpdateApi(sf)

	if assert.NoError(t, updateApi.HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "failed to update", rec.Body.String())
	}
}

func TestUpdateEndpointSuccessSingleSubdomain(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("domain", "foo.com")
	q.Set("subdomain", "bar")
	q.Set("ip", "10.0.0.1")
	q.Set("registrar", "cloudflare")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", q.Encode()), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cs := mock_services.NewMockDnsUpdateService(t)
	cs.On("UpdateRecord", mock.Anything).Return(nil).Once()

	sf := mock_factory.NewMockServiceFactory(t)
	sf.On("Find", services.Registrar("cloudflare")).Return(cs, nil).Once()

	updateApi = NewUpdateApi(sf)

	if assert.NoError(t, updateApi.HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "created 1 entries for subdomains bar on foo.com: 10.0.0.1", rec.Body.String())
	}
}

func TestUpdateEndpointSuccessThreeSubdomains(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("domain", "foo.com")
	q.Set("subdomain", "foo,bar,baz")
	q.Set("ip", "10.0.0.1")
	q.Set("registrar", "cloudflare")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", q.Encode()), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cs := mock_services.NewMockDnsUpdateService(t)
	cs.On("UpdateRecord", mock.Anything).Return(nil).Times(3)

	sf := mock_factory.NewMockServiceFactory(t)
	sf.On("Find", services.Registrar("cloudflare")).Return(cs, nil).Once().Times(3)

	updateApi = NewUpdateApi(sf)

	if assert.NoError(t, updateApi.HandleUpdateRequest(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "created 3 entries for subdomains foo,bar,baz on foo.com: 10.0.0.1", rec.Body.String())
	}
}

func TestValidDomainAndIp(t *testing.T) {
	err := validateRequest("domain.com", "1.1.1.1")
	assert.Nil(t, err, "valid domain and ip should not return error")
}

func TestInvalidIPv4(t *testing.T) {
	err := validateRequest("domain.com", "::1")
	assert.Equal(t, ErrInvalidIP, err, "invalid IPv4 should return error")
}

func TestInvalidDomain(t *testing.T) {
	err := validateRequest("domain .com", "1.1.1.1")
	assert.Equal(t, ErrInvalidDomain, err, "invalid domain should return error")
}
