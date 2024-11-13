package discovery

import (
	"context"
	"log"
)

type Option struct {
	Verbose bool
}

type EnumerationService struct {
	domainEnumSrv *DomainEnumerationService
	linkEnumSrv   *LinkEnumerationService
	option        *Option
}

func NewEumerationService(des *DomainEnumerationService, les *LinkEnumerationService, option *Option) *EnumerationService {
	// TODO: optionの整理
	des.option = option
	les.option = option

	return &EnumerationService{
		domainEnumSrv: des,
		linkEnumSrv:   les,
		option:        option,
	}
}

func (es *EnumerationService) StartScan(ctx context.Context, domain string, mode string) error {
	log.Println("[+] SubDomain Enumeration start")
	if err := es.domainEnumSrv.StartScan(ctx, domain, mode); err != nil {
		return err
	}
	log.Println("[+] SubDomain Enumeration complete")

	// iterate domains
	domains, err := es.domainEnumSrv.domainStorage.GetSubDomains(ctx, domain)
	if err != nil {
		return err
	}

	log.Println("[+] Link Enumeration start")

	for _, domain := range domains {
		es.linkEnumSrv.StartScan(ctx, domain)
	}

	log.Println("[+] Link Enumeration complete")
	log.Printf("[+] You can confirm the result : abashili show -d %v", domain)

	return nil
}
