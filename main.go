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
CREATE TABLE IF NOT EXISTS domains (
    id SERIAL PRIMARY KEY,
    domain_name TEXT UNIQUE NOT NULL,
    parent_id TEXT ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES domains(id),
);

CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    domain_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (subdomain_id) REFERENCES subdomains(id)
);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Println(err)
		return
	}

	cmd.Execute()
}
