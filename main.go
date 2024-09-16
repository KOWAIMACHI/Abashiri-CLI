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
	    id TEXT PRIMARY KEY,
        name TEXT UNIQUE
    );
    CREATE TABLE IF NOT EXISTS subdomains (
	    id TEXT PRIMARY KEY,
        name TEXT UNIQUE,
		parent_id TEXT NOT NULL
    );`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Println(err)
		return
	}

	cmd.Execute()
}
