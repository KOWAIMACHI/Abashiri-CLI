package discovery

import (
	"context"
	"log"
)

type SubDomainOption struct {
	Mode string
}

type URLOption struct {
	IncudeSubDomain bool
}

type Option struct {
	Verbose         bool
	SubDomainScan   bool
	URLScan         bool
	SubDomainOption SubDomainOption
	URLOption       URLOption
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

func (es *EnumerationService) StartScan(ctx context.Context, domain string) error {
	if es.option.URLScan && !es.option.SubDomainScan {
		return es.scanURLsOnly(ctx, domain)
	}
	if !es.option.URLScan && es.option.SubDomainScan {
		return es.scanDomainsOnly(ctx, domain)
	}
	return es.scanBoth(ctx, domain)
}

func (es *EnumerationService) scanDomainsOnly(ctx context.Context, domain string) error {
	log.Println("[+] DomainOnly mode: Skipping URL enumeration")

	log.Println("[+] Starting subdomain enumeration")
	if err := es.domainEnumSrv.StartScan(ctx, domain, es.option.SubDomainOption.Mode); err != nil {
		log.Printf("[-] Subdomain enumeration failed: %v", err)
		return err
	}
	log.Println("[+] Subdomain enumeration complete")
	log.Printf("[+] Check found domains: abashiri show domain -d %v", domain)
	return nil
}

func (es *EnumerationService) scanURLsOnly(ctx context.Context, domain string) error {
	log.Println("[+] URLOnly mode: Skipping subdomain enumeration")
	domains, err := es.domainEnumSrv.domainStorage.GetSubDomainsByParent(ctx, domain)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		err := es.urlEnumSrv.StartScan(ctx, domain)
		if err != nil {
			return err
		}
	}

	log.Printf("[+] Check found links: abashiri show url -d %v", domain)
	return nil
}

func (es *EnumerationService) scanBoth(ctx context.Context, domain string) error {
	log.Println("[+] Full scan mode: Starting subdomain and URL enumeration")

	log.Println("[+] Starting subdomain enumeration")
	if err := es.domainEnumSrv.StartScan(ctx, domain, es.option.SubDomainOption.Mode); err != nil {
		log.Printf("[-] Subdomain enumeration failed: %v", err)
		return err
	}
	log.Println("[+] Subdomain enumeration complete")

	domains, err := es.domainEnumSrv.domainStorage.GetSubDomainsByParent(ctx, domain)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		err := es.urlEnumSrv.StartScan(ctx, domain)
		if err != nil {
			return err
		}
	}

	log.Printf("[+] Check found domains: abashiri show domain -d %v", domain)
	log.Printf("[+] Check found links: abashiri show url -d %v", domain)
	return nil
}
