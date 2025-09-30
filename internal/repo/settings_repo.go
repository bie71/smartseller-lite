package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"smartseller-lite-starter/internal/domain"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) EnsureDefaults(ctx context.Context, brand string) error {
	if brand == "" {
		brand = "SmartSeller Lite"
	}
	const stmt = `INSERT IGNORE INTO settings (` + "`key`" + `, value, updated_at) VALUES (?, ?, ?);`
	_, err := r.db.ExecContext(ctx, stmt, "brand_name", brand, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("ensure default brand: %w", err)
	}
	return nil
}

func (r *SettingsRepository) Get(ctx context.Context) (*domain.AppSettings, error) {
	const stmt = `SELECT ` + "`key`" + `, value FROM settings WHERE ` + "`key`" + ` IN ('brand_name', 'logo_path', 'logo_hash', 'logo_width', 'logo_height', 'logo_size_bytes', 'logo_mime');`
	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("select settings: %w", err)
	}
	defer rows.Close()

	settings := &domain.AppSettings{BrandName: "SmartSeller Lite"}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		switch key {
		case "brand_name":
			if value != "" {
				settings.BrandName = value
			}
		case "logo_path":
			settings.LogoPath = value
		case "logo_hash":
			settings.LogoHash = value
		case "logo_width":
			if width, convErr := strconv.Atoi(value); convErr == nil {
				settings.LogoWidth = width
			}
		case "logo_height":
			if height, convErr := strconv.Atoi(value); convErr == nil {
				settings.LogoHeight = height
			}
		case "logo_size_bytes":
			if size, convErr := strconv.ParseInt(value, 10, 64); convErr == nil {
				settings.LogoSizeBytes = size
			}
		case "logo_mime":
			settings.LogoMime = value
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingsRepository) Update(ctx context.Context, settings domain.AppSettings) (*domain.AppSettings, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const upsert = `INSERT INTO settings (` + "`key`" + `, value, updated_at) VALUES (?, ?, ?)
	                 ON DUPLICATE KEY UPDATE value = VALUES(value), updated_at = VALUES(updated_at);`
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err = tx.ExecContext(ctx, upsert, "brand_name", settings.BrandName, now); err != nil {
		return nil, fmt.Errorf("save brand: %w", err)
	}
	if _, err = tx.ExecContext(ctx, upsert, "logo_path", settings.LogoPath, now); err != nil {
		return nil, fmt.Errorf("save logo path: %w", err)
	}
	if _, err = tx.ExecContext(ctx, upsert, "logo_hash", settings.LogoHash, now); err != nil {
		return nil, fmt.Errorf("save logo hash: %w", err)
	}
	if _, err = tx.ExecContext(ctx, upsert, "logo_width", strconv.Itoa(settings.LogoWidth), now); err != nil {
		return nil, fmt.Errorf("save logo width: %w", err)
	}
	if _, err = tx.ExecContext(ctx, upsert, "logo_height", strconv.Itoa(settings.LogoHeight), now); err != nil {
		return nil, fmt.Errorf("save logo height: %w", err)
	}
	if _, err = tx.ExecContext(ctx, upsert, "logo_size_bytes", strconv.FormatInt(settings.LogoSizeBytes, 10), now); err != nil {
		return nil, fmt.Errorf("save logo size: %w", err)
	}
	if _, err = tx.ExecContext(ctx, upsert, "logo_mime", settings.LogoMime, now); err != nil {
		return nil, fmt.Errorf("save logo mime: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *SettingsRepository) ReplaceAll(ctx context.Context, settings domain.AppSettings) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM settings;`); err != nil {
		return fmt.Errorf("clear settings: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	const insert = `INSERT INTO settings (` + "`key`" + `, value, updated_at) VALUES (?, ?, ?);`
	if _, err = tx.ExecContext(ctx, insert, "brand_name", settings.BrandName, now); err != nil {
		return fmt.Errorf("restore brand: %w", err)
	}
	if _, err = tx.ExecContext(ctx, insert, "logo_path", settings.LogoPath, now); err != nil {
		return fmt.Errorf("restore logo path: %w", err)
	}
	if _, err = tx.ExecContext(ctx, insert, "logo_hash", settings.LogoHash, now); err != nil {
		return fmt.Errorf("restore logo hash: %w", err)
	}
	if _, err = tx.ExecContext(ctx, insert, "logo_width", strconv.Itoa(settings.LogoWidth), now); err != nil {
		return fmt.Errorf("restore logo width: %w", err)
	}
	if _, err = tx.ExecContext(ctx, insert, "logo_height", strconv.Itoa(settings.LogoHeight), now); err != nil {
		return fmt.Errorf("restore logo height: %w", err)
	}
	if _, err = tx.ExecContext(ctx, insert, "logo_size_bytes", strconv.FormatInt(settings.LogoSizeBytes, 10), now); err != nil {
		return fmt.Errorf("restore logo size: %w", err)
	}
	if _, err = tx.ExecContext(ctx, insert, "logo_mime", settings.LogoMime, now); err != nil {
		return fmt.Errorf("restore logo mime: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit settings restore: %w", err)
	}
	return nil
}
