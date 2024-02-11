package api

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
