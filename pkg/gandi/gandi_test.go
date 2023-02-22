package gandi

import (
	"net/http"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/davidramiro/frigabun/internal/config"
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

func TestUpdateEndpointWithValidRequest(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		IP:        config.AppConfig.Test.Gandi.IP,
		Domain:    config.AppConfig.Test.Gandi.Domain,
		Subdomain: config.AppConfig.Test.Gandi.Subdomain,
		ApiKey:    config.AppConfig.Test.Gandi.ApiKey,
	}

	err := testDnsInfo.AddRecord()

	assert.Nil(t, err)
}

func TestUpdateEndpointWithInvalidIp(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		IP:        "::1",
		Domain:    config.AppConfig.Test.Gandi.Domain,
		Subdomain: config.AppConfig.Test.Gandi.Subdomain,
		ApiKey:    config.AppConfig.Test.Gandi.ApiKey,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Contains(t, err.Message, "IPv4")
}

func TestUpdateEndpointWithMissingParam(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		Domain:    config.AppConfig.Test.Gandi.Domain,
		Subdomain: config.AppConfig.Test.Gandi.Subdomain,
		ApiKey:    config.AppConfig.Test.Gandi.ApiKey,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Contains(t, err.Message, "rrset_values.0")

}

func TestUpdateEndpointWithMissingAuth(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		IP:        config.AppConfig.Test.Gandi.IP,
		Domain:    config.AppConfig.Test.Gandi.Domain,
		Subdomain: config.AppConfig.Test.Gandi.Subdomain,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusForbidden, err.Code)
	assert.Contains(t, err.Message, "rejected")

}

func TestUpdateEndpointWithInvalidDomain(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		Domain:    "example.com",
		IP:        config.AppConfig.Test.Gandi.IP,
		Subdomain: config.AppConfig.Test.Gandi.Subdomain,
		ApiKey:    config.AppConfig.Test.Gandi.ApiKey,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, err.Code)
	assert.Contains(t, err.Message, "rejected")

}
