package porkbun

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
	testDnsInfo := &PorkbunDnsInfo{
		IP:           config.AppConfig.Test.Porkbun.IP,
		Domain:       config.AppConfig.Test.Porkbun.Domain,
		Subdomain:    config.AppConfig.Test.Porkbun.Subdomain,
		ApiKey:       config.AppConfig.Test.Porkbun.ApiKey,
		SecretApiKey: config.AppConfig.Test.Porkbun.ApiSecretKey,
	}

	err := testDnsInfo.AddRecord()

	assert.Nil(t, err)
}

func TestUpdateEndpointWithInvalidIp(t *testing.T) {
	testDnsInfo := &PorkbunDnsInfo{
		IP:           "::1",
		Domain:       config.AppConfig.Test.Porkbun.Domain,
		Subdomain:    config.AppConfig.Test.Porkbun.Subdomain,
		ApiKey:       config.AppConfig.Test.Porkbun.ApiKey,
		SecretApiKey: config.AppConfig.Test.Porkbun.ApiSecretKey,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Contains(t, err.Message, "unable to create")
}

func TestUpdateEndpointWithMissingParam(t *testing.T) {
	testDnsInfo := &PorkbunDnsInfo{
		Domain:       config.AppConfig.Test.Porkbun.Domain,
		Subdomain:    config.AppConfig.Test.Porkbun.Subdomain,
		ApiKey:       config.AppConfig.Test.Porkbun.ApiKey,
		SecretApiKey: config.AppConfig.Test.Porkbun.ApiSecretKey,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Contains(t, err.Message, "must have an answer")

}

func TestUpdateEndpointWithMissingAuth(t *testing.T) {
	testDnsInfo := &PorkbunDnsInfo{
		IP:        config.AppConfig.Test.Porkbun.IP,
		Domain:    config.AppConfig.Test.Porkbun.Domain,
		Subdomain: config.AppConfig.Test.Porkbun.Subdomain,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Contains(t, err.Message, "API key")

}

func TestUpdateEndpointWithInvalidDomain(t *testing.T) {
	testDnsInfo := &PorkbunDnsInfo{
		Domain:       "example.com",
		IP:           config.AppConfig.Test.Porkbun.IP,
		Subdomain:    config.AppConfig.Test.Porkbun.Subdomain,
		ApiKey:       config.AppConfig.Test.Porkbun.ApiKey,
		SecretApiKey: config.AppConfig.Test.Porkbun.ApiSecretKey,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Contains(t, err.Message, "Invalid domain")

}
