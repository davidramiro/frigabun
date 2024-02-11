package services

import (
	"fmt"
	"github.com/spf13/viper"
)

type ServiceFactory interface {
	Register(DnsUpdateService)
	Find(Registrar) (DnsUpdateService, error)
	ListServices() []Registrar
}

type DnsUpdateServiceFactory struct {
	services map[Registrar]DnsUpdateService
}

func NewDnsUpdateServiceFactory() (*DnsUpdateServiceFactory, error) {
	factory := &DnsUpdateServiceFactory{
		services: make(map[Registrar]DnsUpdateService),
	}

	if viper.GetBool("cloudflare.enabled") {
		cloudflareService, err := NewCloudflareDnsUpdateService()
		if err != nil {
			return nil, err
		}

		factory.Register(cloudflareService)
	}

	if viper.GetBool("gandi.enabled") {
		gandiService, err := NewGandiDnsUpdateService()
		if err != nil {
			return nil, err
		}

		factory.Register(gandiService)
	}

	if viper.GetBool("porkbun.enabled") {
		porkbunService, err := NewPorkbunDnsUpdateService()
		if err != nil {
			return nil, err
		}

		factory.Register(porkbunService)
	}

	return factory, nil
}

func (df *DnsUpdateServiceFactory) Register(service DnsUpdateService) {
	if service == nil {
		return
	}

	key := service.Registrar()

	df.services[key] = service
}

func (df *DnsUpdateServiceFactory) Find(registrar Registrar) (service DnsUpdateService, err error) {

	service, ok := df.services[registrar]
	if !ok {
		return nil, fmt.Errorf(
			"no dns update service found for registrar %s", registrar)
	}

	return service, nil
}

func (df *DnsUpdateServiceFactory) ListServices() []Registrar {
	keys := make([]Registrar, len(df.services))

	i := 0
	for k := range df.services {
		keys[i] = k
		i++
	}

	return keys
}
