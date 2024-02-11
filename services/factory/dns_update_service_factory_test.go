package factory

import (
	"github.com/davidramiro/frigabun/services"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var factory *DnsUpdateServiceFactory

func init() {
	viper.Set("cloudflare.enabled", true)
	viper.Set("cloudflare.baseurl", "https://api.foo.com/client/v4")
	viper.Set("cloudflare.apiKey", "foo")
	viper.Set("cloudflare.zoneId", "bar")
	viper.Set("cloudflare.ttl", 42)
	viper.Set("gandi.enabled", true)
	viper.Set("gandi.baseurl", "https://api.foo.com/client/v4")
	viper.Set("gandi.apiKey", "foo")
	viper.Set("gandi.ttl", 42)
	viper.Set("porkbun.enabled", true)
	viper.Set("porkbun.baseurl", "https://api.foo.com/client/v4")
	viper.Set("porkbun.apiKey", "foo")
	viper.Set("porkbun.secretApiKey", "bar")
	viper.Set("porkbun.ttl", 42)

	factory, _ = NewDnsUpdateServiceFactory()
}

func TestNewDnsUpdateServiceFactory(t *testing.T) {

	assert.Equal(t, 3, len(factory.ListServices()), "factory should contain 3 services")

	service, err := factory.Find("cloudflare")

	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "cloudflare", string(service.Registrar()))
}

func TestDnsUpdateServiceFactory_ListServices(t *testing.T) {
	assert.Equal(t, 3, len(factory.ListServices()), "factory should contain 3 services")
}

func TestDnsUpdateServiceFactory_FindSuccess(t *testing.T) {
	service, err := factory.Find("cloudflare")

	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "cloudflare", string(service.Registrar()))

}

func TestDnsUpdateServiceFactory_FindInvalidName(t *testing.T) {
	_, err := factory.Find("cloudbun")
	assert.ErrorIs(t, err, services.ErrRegistrarNotFound)
}

func TestNewDnsUpdateServiceFactory_MissingParamForPorkbun(t *testing.T) {
	viper.Set("porkbun.baseUrl", "")
	f, err := NewDnsUpdateServiceFactory()
	assert.ErrorIs(t, err, services.ErrMissingInfoForServiceInit)
	assert.Nil(t, f)
}

func TestNewDnsUpdateServiceFactory_MissingParamForCloudflare(t *testing.T) {
	viper.Set("cloudflare.apiKey", "")
	f, err := NewDnsUpdateServiceFactory()
	assert.ErrorIs(t, err, services.ErrMissingInfoForServiceInit)
	assert.Nil(t, f)
}

func TestNewDnsUpdateServiceFactory_MissingParamForGandi(t *testing.T) {
	viper.Set("gandi.ttl", 0)
	f, err := NewDnsUpdateServiceFactory()
	assert.ErrorIs(t, err, services.ErrMissingInfoForServiceInit)
	assert.Nil(t, f)
}
