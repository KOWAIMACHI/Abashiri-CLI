package main

import (
	"abashiri-cli/cmd"
	"database/sql"
	"fmt"
	"log"
)

func main() {
	db, err := sql.Open("sqlite3", "./abashiri.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStmt := `
CREATE TABLE IF NOT EXISTS corps (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS domains (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    corp_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (corp_id) REFERENCES corps(id)
);

CREATE TABLE IF NOT EXISTS subdomains (
    id TEXT PRIMARY KEY,
    parent_id TEXT NOT NULL,
    root_id TEXT NOT NULL,
    name TEXT NOT NULL,
    tools_detected TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES subdomains(id),
    FOREIGN KEY (root_id) REFERENCES domains(id)
);

CREATE TABLE IF NOT EXISTS links (
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL,
    domain_id TEXT,
    subdomain_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id),
    FOREIGN KEY (subdomain_id) REFERENCES subdomains(id)
);
`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Println(err)
		return
	}

	cmd.Execute()
}
