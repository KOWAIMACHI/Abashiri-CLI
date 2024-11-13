package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type URLStorage interface {
	RegisterURLs(context.Context, string, []string) error
	GetURLs(context.Context, string) ([]string, error)
}

type urlStorage struct {
	db *sql.DB
}

func NewURLStorage(db *sql.DB) URLStorage {
	return &urlStorage{
		db,
	}
}

func (us *urlStorage) RegisterURLs(ctx context.Context, domain string, urls []string) error {
	var subdomainID string
	query := `SELECT id FROM subdomains WHERE name = ?`
	err := us.db.QueryRowContext(ctx, query, domain).Scan(&subdomainID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	tx, err := us.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, url := range urls {
		var count int
		query = `SELECT COUNT(*) FROM urls WHERE url = ? AND  subdomain_id = ?`
		err := tx.QueryRowContext(ctx, query, url, subdomainID).Scan(&count)
		if err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		query = `INSERT INTO urls (id, url, subdomain_id) VALUES (?, ?, ?)`
		_, err = tx.ExecContext(ctx, query, uuid.New().String(), url, subdomainID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (us *urlStorage) GetURLs(ctx context.Context, domain string) ([]string, error) {
	query := `SELECT l.url FROM urls l JOIN subdomains s ON l.subdomain_id = s.id WHERE s.name = ?`
	rows, err := us.db.QueryContext(ctx, query, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	var results []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, url)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}
	return results, nil
}
