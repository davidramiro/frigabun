package services

type Registrar string

type DnsUpdateService interface {
	UpdateRecord(*DynDnsRequest) error
	Registrar() Registrar
}

type DynDnsRequest struct {
	Subdomain string
	Domain    string
	IP        string
}

type registrarSettings struct {
	baseUrl string
	ttl     int
}
