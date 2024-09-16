package scan

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sync"

	"abashiri-cli/helpers"

	"github.com/google/uuid"
)

type Option struct {
	Verbose bool
}

type DomainEnumerationService struct {
	db     *sql.DB
	option *Option
}

func NewDomainEnumerationService(db *sql.DB, option *Option) *DomainEnumerationService {
	return &DomainEnumerationService{
		db:     db,
		option: option,
	}
}

func (ds *DomainEnumerationService) StartScan(domain string) error {
	ctx := context.Background()
	var domainID string
	err := ds.db.QueryRowContext(ctx, "SELECT id FROM domains WHERE name = ?", domain).Scan(&domainID)
	if err != nil {
		if err == sql.ErrNoRows {
			if _, err := ds.db.ExecContext(ctx, "INSERT INTO domains (id, name) VALUES (?,?)", uuid.New().String(), domain); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if err := ds.executePassiveScan(ctx, domain); err != nil {
		return err
	}
	return nil
}

func (ds *DomainEnumerationService) executePassiveScan(ctx context.Context, domain string) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := ds.executeAmassPassiveScan(domain)
		if err != nil {
			log.Println("[-] Error at executeAmassPassiveScan: ", err)
			return
		}
		if result != nil {
			if err := ds.registerSubDomain(ctx, domain, result); err != nil {
				log.Println("[-] Error at registerSubDomain: ", err)
				return
			}
		}
		log.Println("[+] Amass Passive Scan completed")
	}()

	wg.Wait()
	return nil
}

func (ds *DomainEnumerationService) executeAmassPassiveScan(domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/amass-passive-%s.txt", domain)
	cmd := exec.Command("amass", "enum", "-passive", "-d", domain, "-o", outputFile)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start subfinder: %v", err)
	}

	if ds.option.Verbose {
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				fmt.Printf("%s\n", scanner.Text())
			}
		}()
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				fmt.Printf("%s\n", scanner.Text())
			}
		}()
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("subfinder command failed: %v", err)
	}

	return extractSubdomains(outputFile, domain)
}

func (ds *DomainEnumerationService) registerSubDomain(ctx context.Context, domain string, subDomains []string) error {
	var domainID string
	err := ds.db.QueryRowContext(ctx, "SELECT id FROM domains WHERE name = ?", domain).Scan(&domainID)
	if err != nil {
		return err
	}

	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	query := `INSERT INTO subdomains (id, name, parent_id) VALUES (?, ?, ?)`
	for _, subDomain := range subDomains {
		_, err := tx.ExecContext(ctx, query, uuid.New().String(), subDomain, domainID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func extractSubdomains(filePath string, domain string) ([]string, error) {
	re := regexp.MustCompile(fmt.Sprintf(`[a-zA-Z0-9\.\-]+\.%s`, regexp.QuoteMeta(domain)))

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %v", err)
	}
	defer file.Close()

	var subdomains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindAllString(line, -1)
		subdomains = append(subdomains, matches...)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if err = os.Remove(filePath); err != nil {
		return nil, fmt.Errorf("error removing file: %v", err)
	}

	return helpers.RemoveDuplicates(subdomains), nil
}
