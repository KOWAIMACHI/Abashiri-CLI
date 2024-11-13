package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type LinkStorage interface {
	RegisterLinks(context.Context, string, []string) error
	GetLinks(context.Context, string) ([]string, error)
}

type linkStorage struct {
	db *sql.DB
}

func NewLinkStorage(db *sql.DB) LinkStorage {
	return &linkStorage{
		db,
	}
}

func (ls *linkStorage) RegisterLinks(ctx context.Context, domain string, links []string) error {
	var subdomainID string
	query := `SELECT id FROM subdomains WHERE name = ?`
	err := ls.db.QueryRowContext(ctx, query, domain).Scan(&subdomainID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	tx, err := ls.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, link := range links {
		var count int
		query = `SELECT COUNT(*) FROM links WHERE url = ? AND  subdomain_id = ?`
		err := tx.QueryRowContext(ctx, query, link, subdomainID).Scan(&count)
		if err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		query = `INSERT INTO links (id, url, subdomain_id) VALUES (?, ?, ?)`
		_, err = tx.ExecContext(ctx, query, uuid.New().String(), link, subdomainID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (ls *linkStorage) GetLinks(ctx context.Context, domain string) ([]string, error) {
	query := `SELECT l.url FROM links l JOIN subdomains s ON l.subdomain_id = s.id WHERE s.name = ?`
	rows, err := ls.db.QueryContext(ctx, query, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	var results []string
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}
	return results, nil
}
