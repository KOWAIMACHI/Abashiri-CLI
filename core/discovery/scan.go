package discovery

import (
	"abashiri-cli/storage"
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

type ScanService interface {
	Execute(ctx context.Context, domain string) error
}

type EnumerationService struct {
	storageServie     *storage.StorageService
	subdomainmScanSrv ScanService
	urlEnumSrv        ScanService
	option            *Option
}

func NewEumerationService(ss *storage.StorageService, option *Option) *EnumerationService {
	return &EnumerationService{
		storageServie:     ss,
		subdomainmScanSrv: NewSubdomainScanService(ss, option),
		urlEnumSrv:        NewURLEumerationService(ss),
		option:            option,
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
	if err := es.subdomainmScanSrv.Execute(ctx, domain); err != nil {
		log.Printf("[-] Subdomain enumeration failed: %v", err)
		return err
	}
	log.Println("[+] Subdomain enumeration complete")
	log.Printf("[+] Check found domains: abashiri show domain -d %v", domain)
	return nil
}

func (es *EnumerationService) scanURLsOnly(ctx context.Context, domain string) error {
	log.Println("[+] URLOnly mode: Skipping subdomain enumeration")
	domains, err := es.storageServie.DomainStorage.GetSubDomainsByParent(ctx, domain)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		err := es.urlEnumSrv.Execute(ctx, domain)
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
	if err := es.subdomainmScanSrv.Execute(ctx, domain); err != nil {
		log.Printf("[-] Subdomain enumeration failed: %v", err)
		return err
	}
	log.Println("[+] Subdomain enumeration complete")

	domains, err := es.storageServie.DomainStorage.GetSubDomainsByParent(ctx, domain)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		err := es.urlEnumSrv.Execute(ctx, domain)
		if err != nil {
			return err
		}
	}

	log.Printf("[+] Check found domains: abashiri show domain -d %v", domain)
	log.Printf("[+] Check found links: abashiri show url -d %v", domain)
	return nil
}
