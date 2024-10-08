package scan

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

func (ds *DomainEnumerationService) GetSubDomains(domain string) error {
	ctx := context.Background()
	query := `
    SELECT s.name FROM subdomains s
 	   JOIN domains d ON s.parent_id = d.id
    	WHERE d.name = ?`
	rows, err := ds.db.QueryContext(ctx, query, domain)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var subdomain string
		if err := rows.Scan(&subdomain); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		fmt.Println(subdomain)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error occurred during row iteration: %w", err)
	}
	return nil
}

func (ds *DomainEnumerationService) StartScan(domain, mode string) error {
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
	switch mode {
	case "passive":
		ds.executePassiveScan(ctx, domain)
	case "active":
		ds.executeActiveScan(ctx, domain)
	}

	return nil
}

func (ds *DomainEnumerationService) executePassiveScan(ctx context.Context, domain string) {
	var wg sync.WaitGroup
	f := func(method string, scanfunc func(string) ([]string, error)) {
		log.Printf("[+] %s Passive Scan started", method)
		defer wg.Done()
		result, err := scanfunc(domain)
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
		log.Printf("[+] %s Passive Scan completed", method)
	}

	scanFunctions := map[string](func(string) ([]string, error)){
		"Amass":     ds.executeAmassScan,
		"Subfinder": ds.executeSubfinderScan,
	}

	for key, scanFunc := range scanFunctions {
		wg.Add(1)
		go f(key, scanFunc)
	}

	wg.Wait()
}

func (ds *DomainEnumerationService) executeActiveScan(ctx context.Context, domain string) {
	// マルチスレッドっぽくやるとDNSリゾルバが悲鳴あげるので使わないことにした
	result, err := ds.executeDNSBruteForce(ctx, domain)
	if err != nil {
		log.Println("[-] Error at executeDNSBruteForce: ", err)
		return
	}
	if result != nil {
		if err := ds.registerSubDomain(ctx, domain, result); err != nil {
			log.Println("[-] Error at registerSubDomain: ", err)
			return
		}
	}
	log.Println("[+] Active Scan completed")
}

func (ds *DomainEnumerationService) executeDNSBruteForce(ctx context.Context, domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/dnsbrute-%s.txt", domain)
	wordlistPath := filepath.Join("./wordlists/dns", "subdomains-top1million-5000.txt")
	cmd := exec.CommandContext(ctx, "dnsx", "-d", domain, "-w", wordlistPath, "-o", outputFile)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start amass: %v", err)
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
		return nil, fmt.Errorf("amass command failed: %v", err)
	}

	return extractSubdomains(outputFile, domain)
}

func (ds *DomainEnumerationService) executeAmassScan(domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/amass-passive-%s.txt", domain)
	cmd := exec.Command("amass", "enum", "-active", "-d", domain, "-o", outputFile)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start amass: %v", err)
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
		return nil, fmt.Errorf("amass command failed: %v", err)
	}

	return extractSubdomains(outputFile, domain)
}

func (ds *DomainEnumerationService) executeSubfinderScan(domain string) ([]string, error) {
	outputFile := fmt.Sprintf("/tmp/subfinder-passive-%s.txt", domain)
	cmd := exec.Command("subfinder", "-silent", "-all", "-d", domain, "-o", outputFile)
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
	query := `
	SELECT d.id, s.name 
	FROM domains d
	LEFT JOIN subdomains s ON d.id = s.parent_id
	WHERE d.name = ?`
	rows, err := ds.db.QueryContext(ctx, query, domain)
	if err != nil {
		return err
	}
	defer rows.Close()

	var domainID string
	existingSubDomains := make(map[string]bool)

	for rows.Next() {
		var subDomainName sql.NullString
		if err := rows.Scan(&domainID, &subDomainName); err != nil {
			return err
		}
		if subDomainName.Valid {
			existingSubDomains[subDomainName.String] = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	query = `INSERT INTO subdomains (id, name, parent_id) VALUES (?, ?, ?)`
	for _, subDomain := range subDomains {
		if existingSubDomains[subDomain] {
			continue
		}
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
