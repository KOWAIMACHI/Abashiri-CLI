package storage

import (
	"context"
	"database/sql"
	"fmt"
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
	var domainID int
	query := `SELECT id FROM domains WHERE domain_name = ?`
	err := us.db.QueryRowContext(ctx, query, domain).Scan(&domainID)
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
		query = `SELECT COUNT(*) FROM urls WHERE url = ? AND  domain_id = ?`
		err := tx.QueryRowContext(ctx, query, url, domainID).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		if count > 0 {
			continue
		}

		query = `INSERT INTO urls (url, domain_id) VALUES (?, ?)`
		_, err = tx.ExecContext(ctx, query, url, domainID)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}
	return tx.Commit()
}

func (us *urlStorage) GetURLs(ctx context.Context, domain string) ([]string, error) {
	query := `SELECT l.url FROM urls l JOIN domains d ON l.domain_id = d.id WHERE d.domain_name = ?`
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
