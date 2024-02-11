package services

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
	viper.Set("porkbun.apiSecretKey", "bar")
	viper.Set("porkbun.ttl", 42)
}

func TestNewDnsUpdateServiceFactory(t *testing.T) {
	factory, err := NewDnsUpdateServiceFactory()
	assert.Nil(t, err, "no error on creating factory")
	assert.Equal(t, 3, len(factory.ListServices()), "factory should contain 3 services")

	service, err := factory.Find("cloudflare")

	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "cloudflare", string(service.Registrar()))
}

func TestDnsUpdateServiceFactory_ListServices(t *testing.T) {
	factory, err := NewDnsUpdateServiceFactory()
	assert.Nil(t, err, "no error on creating factory")
	assert.Equal(t, 3, len(factory.ListServices()), "factory should contain 3 services")
}

func TestDnsUpdateServiceFactory_FindSuccess(t *testing.T) {
	factory, _ := NewDnsUpdateServiceFactory()
	service, err := factory.Find("cloudflare")

	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "cloudflare", string(service.Registrar()))

}

func TestDnsUpdateServiceFactory_FindInvalidName(t *testing.T) {
	factory, _ := NewDnsUpdateServiceFactory()
	_, err := factory.Find("cloudbun")

	assert.Error(t, err)
}
