package storage

import (
	"abashiri-cli/helper"
	"context"
	"database/sql"
	"fmt"
)

type DomainStorage interface {
	CreateDomainIfNotExists(context.Context, string) error
	RegisterSubDomains(context.Context, string, []string) error
	GetSubDomainsByParentDomain(context.Context, string) ([]string, error)
}

type domainStorage struct {
	db *sql.DB
}

func NewDomainStorage(db *sql.DB) DomainStorage {
	return &domainStorage{
		db,
	}
}

func (ds *domainStorage) CreateDomainIfNotExists(ctx context.Context, domain string) error {
	var domainID int
	err := ds.db.QueryRowContext(ctx, "SELECT id FROM domains WHERE domain_name = ?", domain).Scan(&domainID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx, err := ds.db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO domains (domain_name) VALUES (?)", domain); err != nil {
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
	_subDomains, err := ds.GetSubDomainsByParentDomain(ctx, domain)
	if err != nil {
		return err
	}
	subDomains = helper.RemoveDuplicatesBetweenArrays(subDomains, _subDomains)
	if len(subDomains) == 0 {
		return nil
	}
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	domainID, err := ds.GetDomainIDbyName(ctx, domain)
	if err != nil {
		return err
	}
	query := `INSERT INTO domains (domain_name, parent_id) VALUES (?, ?)`
	for _, subDomain := range subDomains {
		_, err := tx.ExecContext(ctx, query, subDomain, domainID)
		if err != nil {
			return fmt.Errorf("%s: %v", subDomain, err)
		}
	}
	return tx.Commit()
}

func (ds *domainStorage) GetDomainIDbyName(ctx context.Context, domain string) (int, error) {
	var domainID int
	query := `SELECT id FROM domains WHERE domain_name = ?`
	err := ds.db.QueryRowContext(ctx, query, domain).Scan(&domainID)
	if err != nil {
		return -1, err
	}
	return domainID, nil
}

func (ds *domainStorage) GetSubDomainsByParentDomain(ctx context.Context, domain string) ([]string, error) {
	// 指定したdomain自体も含める。問題があれば考える
	query := `SELECT domain_name FROM domains WHERE parent_id = (SELECT id FROM domains WHERE domain_name = ?) OR domain_name = ?`
	rows, err := ds.db.QueryContext(ctx, query, domain, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query %v", err)
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
