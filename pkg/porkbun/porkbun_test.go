package porkbun

import (
	"net/http"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/davidramiro/frigabun/internal/config"
	"github.com/go-faker/faker/v4"
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
	testDnsInfo := &PorkbunDns{
		IP:           faker.IPv4(),
		Domain:       config.AppConfig.Test.Porkbun.Domain,
		Subdomain:    config.AppConfig.Test.Porkbun.Subdomain,
		ApiKey:       config.AppConfig.Test.Porkbun.ApiKey,
		SecretApiKey: config.AppConfig.Test.Porkbun.ApiSecretKey,
	}

	testDnsInfo.Domain = config.AppConfig.Test.Porkbun.Domain
	testDnsInfo.Subdomain = config.AppConfig.Test.Porkbun.Subdomain
	testDnsInfo.ApiKey = config.AppConfig.Test.Porkbun.ApiKey
	testDnsInfo.SecretApiKey = config.AppConfig.Test.Porkbun.ApiSecretKey

	dnsErr := testDnsInfo.AddRecord()

	assert.Nil(t, dnsErr)
}

func TestUpdateEndpointWithInvalidIp(t *testing.T) {
	testDnsInfo := &PorkbunDns{
		IP:           "::1",
		Domain:       config.AppConfig.Test.Porkbun.Domain,
		Subdomain:    config.AppConfig.Test.Porkbun.Subdomain,
		ApiKey:       config.AppConfig.Test.Porkbun.ApiKey,
		SecretApiKey: config.AppConfig.Test.Porkbun.ApiSecretKey,
	}

	err := testDnsInfo.AddRecord()

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Contains(t, err.Message, "unable to")
}

func TestUpdateEndpointWithMissingParam(t *testing.T) {
	testDnsInfo := &PorkbunDns{
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
	testDnsInfo := &PorkbunDns{
		IP:        faker.IPv4(),
		Domain:    config.AppConfig.Test.Porkbun.Domain,
		Subdomain: config.AppConfig.Test.Porkbun.Subdomain,
	}

	dnsErr := testDnsInfo.AddRecord()

	assert.NotNil(t, dnsErr)
	assert.Equal(t, http.StatusBadRequest, dnsErr.Code)
	assert.Contains(t, dnsErr.Message, "API key")

}

func TestUpdateEndpointWithInvalidDomain(t *testing.T) {
	testDnsInfo := &PorkbunDns{
		Domain:       faker.DomainName(),
		IP:           faker.IPv4(),
		Subdomain:    config.AppConfig.Test.Porkbun.Subdomain,
		ApiKey:       config.AppConfig.Test.Porkbun.ApiKey,
		SecretApiKey: config.AppConfig.Test.Porkbun.ApiSecretKey,
	}

	testDnsInfo.Subdomain = config.AppConfig.Test.Porkbun.Subdomain
	testDnsInfo.ApiKey = config.AppConfig.Test.Porkbun.ApiKey
	testDnsInfo.SecretApiKey = config.AppConfig.Test.Porkbun.ApiSecretKey

	dnsErr := testDnsInfo.AddRecord()

	assert.NotNil(t, dnsErr)
	assert.Equal(t, http.StatusBadRequest, dnsErr.Code)
	assert.Contains(t, dnsErr.Message, "Invalid domain")

}
