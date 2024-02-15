package factory

import (
	"github.com/davidramiro/frigabun/services"
	"github.com/rs/zerolog/log"
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
	log.Debug().Msg("initializing dns service factory")

	factory := &DnsUpdateServiceFactory{
		services: make(map[services.Registrar]services.DnsUpdateService),
	}

	if viper.GetBool("cloudflare.enabled") {
		log.Debug().Msg("cloudflare enabled, registering")
		cloudflareService, err := services.NewCloudflareDnsUpdateService(nil)
		if err != nil {
			return nil, err
		}

		factory.Register(cloudflareService)
	}

	if viper.GetBool("gandi.enabled") {
		log.Debug().Msg("gandi enabled, registering")
		gandiService, err := services.NewGandiDnsUpdateService(nil)
		if err != nil {
			return nil, err
		}

		factory.Register(gandiService)
	}

	if viper.GetBool("porkbun.enabled") {
		log.Debug().Msg("porkbun enabled, registering")
		porkbunService, err := services.NewPorkbunDnsUpdateService(nil)
		if err != nil {
			return nil, err
		}

		factory.Register(porkbunService)
	}

	if len(factory.services) == 0 {
		log.Fatal().Msg("no services registered, config invalid")
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
	log.Debug().Interface("registrar", registrar).Msg("fetching dns service from factory")

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
