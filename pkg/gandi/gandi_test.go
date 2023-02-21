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

func TestUpdateEndpointWithInvalidIp(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		IP:        config.AppConfig.Test.IP,
		Domain:    config.AppConfig.Test.Domain,
		Subdomain: config.AppConfig.Test.Subdomain,
		ApiKey:    config.AppConfig.Test.ApiKey,
	}

	err := AddRecord(testDnsInfo)

	assert.Nil(t, err)
}

func TestUpdateEndpointWithMissingParam(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		Domain:    config.AppConfig.Test.Domain,
		Subdomain: config.AppConfig.Test.Subdomain,
		ApiKey:    config.AppConfig.Test.ApiKey,
	}

	err := AddRecord(testDnsInfo)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Contains(t, err.Message, "rrset_values.0")

}

func TestUpdateEndpointWithMissingAuth(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		IP:        config.AppConfig.Test.IP,
		Domain:    config.AppConfig.Test.Domain,
		Subdomain: config.AppConfig.Test.Subdomain,
	}

	err := AddRecord(testDnsInfo)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusForbidden, err.Code)
	assert.Contains(t, err.Message, "rejected")

}

func TestUpdateEndpointWithInvalidDomain(t *testing.T) {
	testDnsInfo := &GandiDnsInfo{
		Domain:    "example.com",
		IP:        config.AppConfig.Test.IP,
		Subdomain: config.AppConfig.Test.Subdomain,
		ApiKey:    config.AppConfig.Test.ApiKey,
	}

	err := AddRecord(testDnsInfo)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, err.Code)
	assert.Contains(t, err.Message, "rejected")

}
