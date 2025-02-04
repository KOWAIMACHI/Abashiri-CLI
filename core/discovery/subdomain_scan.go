package discovery

import (
	"abashiri-cli/helper"
	"abashiri-cli/storage"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/buger/jsonparser"
)

type subdomainScanService struct {
	storage    storage.DomainStorage
	httpClient *HTTPClient
	option     *Option
}

func NewSubdomainScanService(ss storage.DomainStorage, option *Option) ScanService {
	return &subdomainScanService{
		storage:    ss,
		httpClient: newHTTPClient(),
		option:     option,
	}
}

func (ds *subdomainScanService) Execute(ctx context.Context, domain string) error {
	err := ds.storage.CreateDomainIfNotExists(ctx, domain)
	if err != nil {
		return err
	}
	var result []string

	result, err = ds.executePassiveScan(ctx, domain)
	if err != nil {
		return err
	}

	if ds.option.SubDomainOption.Mode == "active" {
		res, err := ds.executeActiveScan(ctx, domain)
		if err != nil {
			return err
		}
		result = append(result, res...)
	}

	return ds.storage.RegisterSubDomains(ctx, domain, helper.RemoveDuplicatesFromArray(result))
}

func (ds *subdomainScanService) executePassiveScan(ctx context.Context, domain string) ([]string, error) {
	scanFunctions := map[string](func(string) ([]string, error)){
		"Subfinder": ds.executeSubfinderScan,
		// "Amass":     ds.executeAmassScan, //クッソ遅いので保留
		"AlienVault OTX": ds.enumDomainFromAlienVaultOTX,
		// "bevigil":        ds.enumURLFromBevigil, // APIキー必要だし、無料だと月50回しかクエリできないし、対して精度良くないので作ったけど使わない
	}

	var results []string
	for method, scanfunc := range scanFunctions {
		log.Printf("[+] %s start: %s", method, domain)
		result, err := scanfunc(domain)
		if err != nil {
			log.Printf("[-] Error at %s Passive Scan for %s: %v", method, domain, err)
			continue
		}
		results = append(results, result...)
		log.Printf("[+] %s complete: %s", method, domain)
	}
	return results, nil
}

func (ds *subdomainScanService) executeActiveScan(ctx context.Context, domain string) ([]string, error) {
	return ds.executeDNSBruteForce(ctx, domain)
}

func (ds *subdomainScanService) executeScanCmd(cmdName string, args []string) error {
	cmd := exec.Command(cmdName, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start amass: %v", err)
	}

	if ds.option.Verbose {
		scannerStdout := bufio.NewScanner(stdout)
		go func() {
			for scannerStdout.Scan() {
				fmt.Printf("%s\n", scannerStdout.Text())
			}
		}()
		scannerStderr := bufio.NewScanner(stderr)
		go func() {
			for scannerStderr.Scan() {
				fmt.Printf("%s\n", scannerStderr.Text())
			}
		}()
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("[%s] failed: %v", cmdName, err)
	}
	return nil
}

// Amass遅すぎ問題
func (ds *subdomainScanService) executeAmassScan(domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/amass-passive-%s.txt", domain)
	args := []string{"enum", "-active", "-d", domain, "-o", outputFile}
	if err := ds.executeScanCmd("amass", args); err != nil {
		return nil, err
	}
	return helper.ExtractSubdomains(outputFile, domain)
}

func (ds *subdomainScanService) executeSubfinderScan(domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/subfinder-passive-%s.txt", domain)
	args := []string{"-silent", "-all", "-d", domain, "-o", outputFile}
	if err := ds.executeScanCmd("subfinder", args); err != nil {
		return nil, err
	}
	return helper.ExtractSubdomains(outputFile, domain)
}

// TODO: もうちょい実装DRYにしたい
func (ds *subdomainScanService) enumDomainFromAlienVaultOTX(domain string) ([]string, error) {
	apiURL := fmt.Sprintf("https://otx.alienvault.com/otxapi/indicators/domain/passive_dns/%s", domain)
	resp, err := ds.httpClient.GET(apiURL, nil)
	if err != nil {
		log.Printf("[-] faled to request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[-] (%s) failed to parse response: %v", domain, err)
		return nil, err
	}

	var results []string

	passiveDnsData, _, _, err := jsonparser.Get(body, "passive_dns")
	if err != nil {
		return nil, err
	}

	jsonparser.ArrayEach(passiveDnsData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			log.Printf("[-] Error parsing array element: %v", err)
		}
		hostname, err := jsonparser.GetString(value, "hostname")
		if err != nil {
			log.Printf("[-] Error parsing array element: %v", err)
		}
		results = append(results, hostname)
	})

	return results, nil
}

// enumURLFromBevigil
func (ds *subdomainScanService) enumURLFromBevigil(domain string) ([]string, error) {
	apiURL := fmt.Sprintf("https://osint.bevigil.com/api/%s/subdomains/", domain)
	header := http.Header{
		// TODO: configから読む
		"X-Access-Token": {"xxxxx"},
	}
	resp, err := ds.httpClient.GET(apiURL, header)
	if err != nil {
		log.Printf("faled to request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to parse response: %v", err)
		return nil, err
	}

	var results []string

	subdomains, _, _, err := jsonparser.Get(body, "subdomains")
	if err != nil {
		log.Fatalf("Error getting passive_dns: %v", err)
	}

	jsonparser.ArrayEach(subdomains, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			log.Fatalf("Error parsing array element: %v", err)
		}
		results = append(results, string(value))
	})

	return results, nil
}

func (ds *subdomainScanService) executeDNSBruteForce(ctx context.Context, domain string) ([]string, error) {
	log.Printf("[+] DNS bruteforce start: %s", domain)
	outputFile := fmt.Sprintf("/tmp/dnsbrute-%s.txt", domain)
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	wordlistPath := fmt.Sprintf("%s/.abashiri/subdomains-top1million-20000.txt", dir)
	args := []string{"-d", domain, "-w", wordlistPath, "-o", outputFile}
	if err := ds.executeScanCmd("dnsx", args); err != nil {
		return nil, err
	}
	log.Printf("[+] DNS bruteforce end: %s", domain)
	return helper.ExtractSubdomains(outputFile, domain)
}
