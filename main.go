package main

import (
	"abashiri-cli/cmd"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func init() {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if err = os.MkdirAll(fmt.Sprintf("%s/.abashiri", dir), 0755); err != nil {
		log.Fatal(err)
	}
	dnsWordlistPath := fmt.Sprintf("%s/.abashiri/subdomains-top1million-20000.txt", dir)
	if _, err := os.Stat(dnsWordlistPath); os.IsNotExist(err) {
		url := "https://raw.githubusercontent.com/danielmiessler/SecLists/refs/heads/master/Discovery/DNS/subdomains-top1million-20000.txt"
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("failed to download wordlists: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed to download wordlists: %v", resp.Status)
		}

		file, err := os.Create(dnsWordlistPath)
		if err != nil {
			log.Fatalf("failed to create file: %v", err)
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Fatalf("failed to copy: %v", err)
		}
	}
}

func main() {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/.abashiri/abashiri.db", dir))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStmt := `
CREATE TABLE IF NOT EXISTS domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_name TEXT UNIQUE NOT NULL,
    parent_id TEXT, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES domains(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL,
    domain_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE
);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Println(err)
		return
	}

	cmd.Execute()
}
