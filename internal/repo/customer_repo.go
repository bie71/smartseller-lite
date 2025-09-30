package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"smartseller-lite-starter/internal/domain"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
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
	const stmt = `SELECT id, type, name, phone, email, address, city, province, postal, notes, created_at, updated_at FROM customers ORDER BY name;`
	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()

	items := []domain.Customer{}
	for rows.Next() {
		var c domain.Customer
		var created, updated string
		if err := rows.Scan(&c.ID, &c.Type, &c.Name, &c.Phone, &c.Email, &c.Address, &c.City, &c.Province, &c.Postal, &c.Notes, &created, &updated); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, created)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		items = append(items, c)
	}
	return items, nil
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
