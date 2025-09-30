package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"smartseller-lite-starter/internal/domain"
)

type CourierRepository struct {
	db *sql.DB
}

func NewCourierRepository(db *sql.DB) *CourierRepository {
	return &CourierRepository{db: db}
}

func (r *CourierRepository) EnsureDefaults(ctx context.Context, defaults []domain.Courier) error {
	rows, err := r.db.QueryContext(ctx, `SELECT code FROM couriers;`)
	if err != nil {
		return fmt.Errorf("list courier codes: %w", err)
	}
	defer rows.Close()

	existing := make(map[string]struct{})
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return fmt.Errorf("scan courier code: %w", err)
		}
		existing[strings.ToUpper(strings.TrimSpace(code))] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate courier codes: %w", err)
	}

	const insertStmt = `INSERT INTO couriers (id, code, name, services, tracking_url, contact, notes, logo_path, logo_hash, logo_width, logo_height, logo_size_bytes, logo_mime, created_at, updated_at)
                         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	now := time.Now().UTC().Format(time.RFC3339)
	for _, c := range defaults {
		code := strings.ToUpper(strings.TrimSpace(c.Code))
		if code == "" {
			continue
		}
		if _, ok := existing[code]; ok {
			continue
		}
		id := c.ID
		if id == "" {
			id = uuid.New().String()
		}
		if _, err := r.db.ExecContext(ctx, insertStmt, id, code, c.Name, c.Services, c.TrackingURL, c.Contact, c.Notes, c.LogoPath, c.LogoHash, c.LogoWidth, c.LogoHeight, c.LogoSizeBytes, c.LogoMime, now, now); err != nil {
			return fmt.Errorf("insert default courier: %w", err)
		}
	}
	return nil
}

func (r *CourierRepository) List(ctx context.Context) ([]domain.Courier, error) {
	const stmt = `SELECT id, code, name, services, tracking_url, contact, notes, logo_path, logo_hash, logo_width, logo_height, logo_size_bytes, logo_mime, created_at, updated_at FROM couriers ORDER BY name;`
	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("list couriers: %w", err)
	}
	defer rows.Close()

	var items []domain.Courier
	for rows.Next() {
		var c domain.Courier
		var created, updated string
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Services, &c.TrackingURL, &c.Contact, &c.Notes, &c.LogoPath, &c.LogoHash, &c.LogoWidth, &c.LogoHeight, &c.LogoSizeBytes, &c.LogoMime, &created, &updated); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, created)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CourierRepository) Get(ctx context.Context, id string) (*domain.Courier, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("courier id required")
	}
	const stmt = `SELECT id, code, name, services, tracking_url, contact, notes, logo_path, logo_hash, logo_width, logo_height, logo_size_bytes, logo_mime, created_at, updated_at FROM couriers WHERE id = ?;`
	var c domain.Courier
	var created, updated string
	if err := r.db.QueryRowContext(ctx, stmt, id).Scan(&c.ID, &c.Code, &c.Name, &c.Services, &c.TrackingURL, &c.Contact, &c.Notes, &c.LogoPath, &c.LogoHash, &c.LogoWidth, &c.LogoHeight, &c.LogoSizeBytes, &c.LogoMime, &created, &updated); err != nil {
		return nil, fmt.Errorf("get courier: %w", err)
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &c, nil
}

func (r *CourierRepository) Save(ctx context.Context, courier *domain.Courier) (*domain.Courier, error) {
	now := time.Now().UTC()
	if courier.ID == "" {
		courier.ID = uuid.New().String()
		courier.CreatedAt = now
		courier.UpdatedAt = now
		const stmt = `INSERT INTO couriers (id, code, name, services, tracking_url, contact, notes, logo_path, logo_hash, logo_width, logo_height, logo_size_bytes, logo_mime, created_at, updated_at)
                      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
		_, err := r.db.ExecContext(ctx, stmt, courier.ID, courier.Code, courier.Name, courier.Services, courier.TrackingURL, courier.Contact, courier.Notes, courier.LogoPath, courier.LogoHash, courier.LogoWidth, courier.LogoHeight, courier.LogoSizeBytes, courier.LogoMime, courier.CreatedAt.Format(time.RFC3339), courier.UpdatedAt.Format(time.RFC3339))
		if err != nil {
			return nil, fmt.Errorf("insert courier: %w", err)
		}
		return courier, nil
	}

	courier.UpdatedAt = now
	const stmt = `UPDATE couriers SET code = ?, name = ?, services = ?, tracking_url = ?, contact = ?, notes = ?, logo_path = ?, logo_hash = ?, logo_width = ?, logo_height = ?, logo_size_bytes = ?, logo_mime = ?, updated_at = ? WHERE id = ?;`
	if _, err := r.db.ExecContext(ctx, stmt, courier.Code, courier.Name, courier.Services, courier.TrackingURL, courier.Contact, courier.Notes, courier.LogoPath, courier.LogoHash, courier.LogoWidth, courier.LogoHeight, courier.LogoSizeBytes, courier.LogoMime, courier.UpdatedAt.Format(time.RFC3339), courier.ID); err != nil {
		return nil, fmt.Errorf("update courier: %w", err)
	}
	return courier, nil
}

func (r *CourierRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("courier id required")
	}
	const stmt = `DELETE FROM couriers WHERE id = ?;`
	if _, err := r.db.ExecContext(ctx, stmt, id); err != nil {
		return fmt.Errorf("delete courier: %w", err)
	}
	return nil
}

func (r *CourierRepository) ReplaceAll(ctx context.Context, items []domain.Courier) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM couriers;`); err != nil {
		return fmt.Errorf("clear couriers: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO couriers (id, code, name, services, tracking_url, contact, notes, logo_path, logo_hash, logo_width, logo_height, logo_size_bytes, logo_mime, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("prepare courier insert: %w", err)
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

		if _, err = stmt.ExecContext(ctx, id, item.Code, item.Name, item.Services, item.TrackingURL, item.Contact, item.Notes, item.LogoPath, item.LogoHash, item.LogoWidth, item.LogoHeight, item.LogoSizeBytes, item.LogoMime, created.Format(time.RFC3339), updated.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("insert courier from backup: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit courier restore: %w", err)
	}
	return nil
}
