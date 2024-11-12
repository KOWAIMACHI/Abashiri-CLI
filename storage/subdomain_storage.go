package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type DomainStorage interface {
	CreateDomainIfNotExists(context.Context, string) error
	RegisterSubDomains(context.Context, string, []string) error
	GetSubDomains(string) ([]string, error)
}

type domainStorage struct {
	db *sql.DB
}

func NewDomainStorage(db *sql.DB) *domainStorage {
	return &domainStorage{
		db,
	}
}

func (ds *domainStorage) CreateDomainIfNotExists(ctx context.Context, domain string) error {
	var domainID string
	err := ds.db.QueryRowContext(ctx, "SELECT id FROM domains WHERE name = ?", domain).Scan(&domainID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx, err := ds.db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			domainID = uuid.New().String()
			if _, err := tx.ExecContext(ctx, "INSERT INTO domains (id, name) VALUES (?,?)", domainID, domain); err != nil {
				tx.Rollback()
				return err
			}
			if err = tx.Commit(); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (ds *domainStorage) RegisterSubDomains(ctx context.Context, domain string, subDomains []string) error {
	//===========
	// memo: 別の関数に切り分けるべきだと思う
	query := `SELECT d.id, s.name FROM domains d LEFT JOIN subdomains s ON d.id = s.parent_id WHERE d.name = ?`
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
	//===========

	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query = `INSERT INTO subdomains (id, name, parent_id) VALUES (?, ?, ?)`
	for _, subDomain := range subDomains {
		if existingSubDomains[subDomain] {
			continue
		}
		_, err := tx.ExecContext(ctx, query, uuid.New().String(), subDomain, domainID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (ds *domainStorage) GetSubDomains(domain string) ([]string, error) {
	ctx := context.Background()
	query := `SELECT s.name FROM subdomains s JOIN domains d ON s.parent_id = d.id WHERE d.name = ?`
	rows, err := ds.db.QueryContext(ctx, query, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	var results []string
	for rows.Next() {
		var subdomain string
		if err := rows.Scan(&subdomain); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, subdomain)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}
	return results, nil
}
