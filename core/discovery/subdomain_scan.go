package discovery

import (
	"abashiri-cli/helper"
	"abashiri-cli/storage"
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

type DomainEnumerationService struct {
	domainStorage storage.DomainStorage
	option        *Option
}

func NewDomainEnumerationService(ds storage.DomainStorage, option *Option) *DomainEnumerationService {
	return &DomainEnumerationService{
		domainStorage: ds,
		option:        option,
	}
}

func (ds *DomainEnumerationService) StartScan(ctx context.Context, domain, mode string) error {
	err := ds.domainStorage.CreateDomainIfNotExists(ctx, domain)
	if err != nil {
		return err
	}
	var result []string

	switch mode {
	case "passive":
		result, err = ds.executePassiveScan(ctx, domain)
	case "active":
		result, err = ds.executeActiveScan(ctx, domain)
	default:
		return fmt.Errorf("unknown scan mode: %s", mode)
	}

	if err != nil {
		return err
	}

	return ds.domainStorage.RegisterSubDomains(ctx, domain, result)
}

func (ds *DomainEnumerationService) executePassiveScan(ctx context.Context, domain string) ([]string, error) {
	scanFunctions := map[string](func(string) ([]string, error)){
		"Subfinder": ds.executeSubfinderScan,
		// "Amass":     ds.executeAmassScan,
	}

	var results []string
	for method, scanfunc := range scanFunctions {
		log.Printf("[+] %s Passive Scan started", method)
		result, err := scanfunc(domain)
		if err != nil {
			log.Printf("[-] Error at %s Passive Scan for %s: %v", method, domain, err)
			continue
		}
		results = append(results, result...)
		log.Printf("[+] %s Passive Scan completed", method)
	}
	return results, nil
}

func (ds *DomainEnumerationService) executeActiveScan(ctx context.Context, domain string) ([]string, error) {
	return ds.executeDNSBruteForce(ctx, domain)
}

func (ds *DomainEnumerationService) executeScanCmd(cmdName string, args []string, domain string) error {
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
func (ds *DomainEnumerationService) executeAmassScan(domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/amass-passive-%s.txt", domain)
	args := []string{"enum", "-active", "-d", domain, "-o", outputFile}
	if err := ds.executeScanCmd("amass", args, domain); err != nil {
		return nil, err
	}
	return helper.ExtractSubdomains(outputFile, domain)
}

func (ds *DomainEnumerationService) executeSubfinderScan(domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/subfinder-passive-%s.txt", domain)
	args := []string{"-silent", "-all", "-d", domain, "-o", outputFile}
	if err := ds.executeScanCmd("subfinder", args, domain); err != nil {
		return nil, err
	}
	return helper.ExtractSubdomains(outputFile, domain)
}

func (ds *DomainEnumerationService) executeDNSBruteForce(ctx context.Context, domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/dnsbrute-%s.txt", domain)
	wordlistPath := filepath.Join("./wordlists/dns", "subdomains-top1million-5000.txt")
	args := []string{"-d", domain, "-w", wordlistPath, "-o", outputFile}
	if err := ds.executeScanCmd("dnsx", args, domain); err != nil {
		return nil, err
	}
	return helper.ExtractSubdomains(outputFile, domain)
}
