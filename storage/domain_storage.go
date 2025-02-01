package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type DomainStorage interface {
	CreateDomainIfNotExists(context.Context, string) error
	RegisterSubDomains(context.Context, string, []string) error
	GetSubDomainsByParent(context.Context, string) ([]string, error)
	GetRootDomains(context.Context) ([]string, error)
	DeleteDomains(context.Context, string) error
}

type domainStorage struct {
	db *sql.DB
}

func NewDomainStorage(db *sql.DB) DomainStorage {
	return &domainStorage{
		db,
	}
}

// CreateDomainIfNotExistsはドメインが存在しなかった場合にそのドメインをDBに登録する
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

// RegisterSubDomainsは引数で渡されたsubDomainsをDBに登録する
func (ds *domainStorage) RegisterSubDomains(ctx context.Context, domain string, subDomains []string) error {
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	domainID, err := ds.GetDomainIDbyName(ctx, domain)
	if err != nil {
		return err
	}

	// HACKME: 急にdev.app.example.comみたいなドメインを収集した場合、app.example.comが存在しないため、parent_idが親ドメインという関係にならない
	// 機能的な問題で許容してもいいかなという気持ちがありつつ、今後どうするか考える
	query := `INSERT OR IGNORE INTO domains (domain_name, parent_id) VALUES (?, ?)`
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

func (ds *domainStorage) GetRootDomains(ctx context.Context) ([]string, error) {
	query := `SELECT domain_name FROM domains WHERE parent_id IS NULL`
	rows, err := ds.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	var results []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, domain)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return results, nil
}

/*
再起的にsubdomainを取得
id, domain_name, parent_id
1,example.com,
2,app.example.com,1
3.test.app.example.com,2

ex:
GetSubDomainByparentDomain(ctx, "example.com")
-> example.com, app.example.com, test.app.example.com
*/
func (ds *domainStorage) GetSubDomainsByParent(ctx context.Context, domain string) ([]string, error) {

	query := `
WITH RECURSIVE domain_hierarchy(id, domain_name, parent_id) AS (
	SELECT id, domain_name, parent_id FROM domains WHERE domain_name = ?

	UNION ALL

	SELECT d.id, d.domain_name, d.parent_id
	FROM domains d
	INNER JOIN domain_hierarchy dh ON dh.id = d.parent_id
)
SELECT domain_name FROM domain_hierarchy;
`
	rows, err := ds.db.QueryContext(ctx, query, domain)
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

func (ds *domainStorage) DeleteDomains(ctx context.Context, domain string) error {

	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	var domainID int
	query := `SELECT id FROM domains WHERE domain_name = ?`
	if err = tx.QueryRowContext(ctx, query, domain).Scan(&domainID); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM domains WHERE id = ?", domainID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete root domain %s: %w", domain, err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM domains WHERE parent_id = ?", domainID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete subdomains for %s: %w", domain, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil

}
