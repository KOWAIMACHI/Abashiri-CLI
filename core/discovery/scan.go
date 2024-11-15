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
	urlEnumSrv    *URLEnumerationService
	option        *Option
}

func NewEumerationService(des *DomainEnumerationService, ues *URLEnumerationService, option *Option) *EnumerationService {
	// TODO: optionの整理
	des.option = option
	ues.option = option

	return &EnumerationService{
		domainEnumSrv: des,
		urlEnumSrv:    ues,
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
	// ===ここ 並列処理にしたいし、recursiveな調査したい
	domains, err := es.domainEnumSrv.domainStorage.GetSubDomainsByDomain(ctx, domain)
	if err != nil {
		return err
	}

	log.Println("[+] URL Enumeration start")

	for _, domain := range domains {
		err := es.urlEnumSrv.StartScan(ctx, domain)
		if err != nil {
			return err
		}
	}
	// ===

	log.Println("[+] URL Enumeration complete")
	log.Printf("[+] check found domains : abashiri-cli show domain -d %v", domain)
	log.Printf("[+] check found links : abashiri-cli show url -d %v", domain)
	return nil
}
