package factory

import (
	"github.com/davidramiro/frigabun/services"
	"github.com/spf13/viper"
)

type ServiceFactory interface {
	Register(services.DnsUpdateService)
	Find(services.Registrar) (services.DnsUpdateService, error)
	ListServices() []services.Registrar
}

type DnsUpdateServiceFactory struct {
	services map[services.Registrar]services.DnsUpdateService
}

func NewDnsUpdateServiceFactory() (*DnsUpdateServiceFactory, error) {
	factory := &DnsUpdateServiceFactory{
		services: make(map[services.Registrar]services.DnsUpdateService),
	}

	if viper.GetBool("cloudflare.enabled") {
		cloudflareService, err := services.NewCloudflareDnsUpdateService(nil)
		if err != nil {
			return nil, err
		}

		factory.Register(cloudflareService)
	}

	if viper.GetBool("gandi.enabled") {
		gandiService, err := services.NewGandiDnsUpdateService(nil)
		if err != nil {
			return nil, err
		}

		factory.Register(gandiService)
	}

	if viper.GetBool("porkbun.enabled") {
		porkbunService, err := services.NewPorkbunDnsUpdateService(nil)
		if err != nil {
			return nil, err
		}

		factory.Register(porkbunService)
	}

	return factory, nil
}

func (df *DnsUpdateServiceFactory) Register(service services.DnsUpdateService) {
	if service == nil {
		return
	}

	key := service.Registrar()

	df.services[key] = service
}

func (df *DnsUpdateServiceFactory) Find(registrar services.Registrar) (service services.DnsUpdateService, err error) {

	service, ok := df.services[registrar]
	if !ok {
		return nil, services.ErrRegistrarNotFound
	}

	return service, nil
}

func (df *DnsUpdateServiceFactory) ListServices() []services.Registrar {
	keys := make([]services.Registrar, len(df.services))

	i := 0
	for k := range df.services {
		keys[i] = k
		i++
	}

	return keys
}
