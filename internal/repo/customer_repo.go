package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"smartseller-lite-starter/internal/domain"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) ListPaged(ctx context.Context, opts CustomerListOptions) (CustomerListResult, error) {
	const maxPageSize = 200

	page := opts.Page
	if page <= 0 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 0 {
		pageSize = 0
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	whereParts := make([]string, 0)
	args := make([]any, 0)

	query := strings.TrimSpace(strings.ToLower(opts.Query))
	if query != "" {
		like := "%" + query + "%"
		whereParts = append(whereParts, "(LOWER(name) LIKE ? OR LOWER(IFNULL(phone,'')) LIKE ? OR LOWER(IFNULL(email,'')) LIKE ? OR LOWER(IFNULL(address,'')) LIKE ? OR LOWER(IFNULL(city,'')) LIKE ? OR LOWER(IFNULL(province,'')) LIKE ? OR LOWER(IFNULL(postal,'')) LIKE ? OR LOWER(IFNULL(notes,'')) LIKE ?)")
		args = append(args, like, like, like, like, like, like, like, like)
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	limitClause := ""
	listArgs := append([]any{}, args...)
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		limitClause = " LIMIT ? OFFSET ?"
		listArgs = append(listArgs, pageSize, offset)
	}

	stmt := "SELECT id, type, name, phone, email, address, city, province, postal, notes, created_at, updated_at FROM customers " + whereClause + " ORDER BY name" + limitClause + ";"

	rows, err := r.db.QueryContext(ctx, stmt, listArgs...)
	if err != nil {
		return CustomerListResult{}, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Customer, 0)
	for rows.Next() {
		var c domain.Customer
		var created, updated string
		if err := rows.Scan(&c.ID, &c.Type, &c.Name, &c.Phone, &c.Email, &c.Address, &c.City, &c.Province, &c.Postal, &c.Notes, &created, &updated); err != nil {
			return CustomerListResult{}, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, created)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		items = append(items, c)
	}

	countStmt := "SELECT COUNT(*) FROM customers " + whereClause + ";"
	var total int
	if err := r.db.QueryRowContext(ctx, countStmt, args...).Scan(&total); err != nil {
		return CustomerListResult{}, fmt.Errorf("count customers: %w", err)
	}

	countsStmt := "SELECT " +
		"SUM(CASE WHEN type = 'customer' THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN type = 'marketer' THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN type = 'reseller' THEN 1 ELSE 0 END) " +
		"FROM customers " + whereClause + ";"
	var customerCount, marketerCount, resellerCount sql.NullInt64
	if err := r.db.QueryRowContext(ctx, countsStmt, args...).Scan(&customerCount, &marketerCount, &resellerCount); err != nil {
		return CustomerListResult{}, fmt.Errorf("customer type counts: %w", err)
	}

	result := CustomerListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	if pageSize <= 0 {
		result.Page = 1
		result.PageSize = len(items)
	}
	if customerCount.Valid {
		result.Counts.Customer = int(customerCount.Int64)
	}
	if marketerCount.Valid {
		result.Counts.Marketer = int(marketerCount.Int64)
	}
	if resellerCount.Valid {
		result.Counts.Reseller = int(resellerCount.Int64)
	}

	return result, nil
}

type CustomerListOptions struct {
	Query    string
	Page     int
	PageSize int
}

type CustomerListResult struct {
	Items    []domain.Customer `json:"items"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	Counts   struct {
		Customer int `json:"customer"`
		Marketer int `json:"marketer"`
		Reseller int `json:"reseller"`
	} `json:"counts"`
}

func (r *CustomerRepository) Create(ctx context.Context, c *domain.Customer) (*domain.Customer, error) {
	now := time.Now().UTC()
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	c.CreatedAt = now
	c.UpdatedAt = now

	const stmt = `INSERT INTO customers (id, type, name, phone, email, address, city, province, postal, notes, created_at, updated_at)
                  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err := r.db.ExecContext(ctx, stmt, c.ID, c.Type, c.Name, c.Phone, c.Email, c.Address, c.City, c.Province, c.Postal, c.Notes, c.CreatedAt.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("insert customer: %w", err)
	}
	return c, nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *domain.Customer) (*domain.Customer, error) {
	c.UpdatedAt = time.Now().UTC()
	const stmt = `UPDATE customers SET type = ?, name = ?, phone = ?, email = ?, address = ?, city = ?, province = ?, postal = ?, notes = ?, updated_at = ? WHERE id = ?;`
	if _, err := r.db.ExecContext(ctx, stmt, c.Type, c.Name, c.Phone, c.Email, c.Address, c.City, c.Province, c.Postal, c.Notes, c.UpdatedAt.Format(time.RFC3339), c.ID); err != nil {
		return nil, fmt.Errorf("update customer: %w", err)
	}
	return c, nil
}

func (r *CustomerRepository) List(ctx context.Context) ([]domain.Customer, error) {
	result, err := r.ListPaged(ctx, CustomerListOptions{Page: 1, PageSize: 0})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (r *CustomerRepository) Get(ctx context.Context, id string) (*domain.Customer, error) {
	const stmt = `SELECT id, type, name, phone, email, address, city, province, postal, notes, created_at, updated_at FROM customers WHERE id = ?;`
	var c domain.Customer
	var created, updated string
	if err := r.db.QueryRowContext(ctx, stmt, id).Scan(&c.ID, &c.Type, &c.Name, &c.Phone, &c.Email, &c.Address, &c.City, &c.Province, &c.Postal, &c.Notes, &created, &updated); err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &c, nil
}

func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("customer id required")
	}
	const stmt = `DELETE FROM customers WHERE id = ?;`
	if _, err := r.db.ExecContext(ctx, stmt, id); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1451 {
			return fmt.Errorf("delete customer: kontak masih digunakan oleh data order")
		}
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) ReplaceAll(ctx context.Context, items []domain.Customer) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM customers;`); err != nil {
		return fmt.Errorf("clear customers: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO customers (id, type, name, phone, email, address, city, province, postal, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("prepare customer insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, item := range items {
		id := item.ID
		if id == "" {
			id = uuid.New().String()
		}
		created := item.CreatedAt
		if created.IsZero() {
			created = now
		}
		updated := item.UpdatedAt
		if updated.IsZero() {
			updated = created
		}

		if _, err = stmt.ExecContext(ctx, id, item.Type, item.Name, item.Phone, item.Email, item.Address, item.City, item.Province, item.Postal, item.Notes, created.Format(time.RFC3339), updated.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("insert customer from backup: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit customer restore: %w", err)
	}
	return nil
}
