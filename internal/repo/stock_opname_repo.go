package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"smartseller-lite-starter/internal/domain"
)

// StockOpnameRepository persists stock take sessions and their line items.
type StockOpnameRepository struct {
	db *sql.DB
}

func NewStockOpnameRepository(db *sql.DB) *StockOpnameRepository {
	return &StockOpnameRepository{db: db}
}

func (r *StockOpnameRepository) Create(ctx context.Context, opname *domain.StockOpname) (*domain.StockOpname, error) {
	if opname == nil {
		return nil, fmt.Errorf("stock opname payload is nil")
	}
	if len(opname.Items) == 0 {
		return nil, fmt.Errorf("stock opname requires at least one item")
	}
	now := time.Now().UTC()
	if opname.ID == "" {
		opname.ID = uuid.New().String()
	}
	if opname.PerformedAt.IsZero() {
		opname.PerformedAt = now
	}
	opname.CreatedAt = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const headerStmt = `INSERT INTO stock_opnames (id, note, performed_by, performed_at, created_at) VALUES (?, ?, ?, ?, ?);`
	if _, err = tx.ExecContext(ctx, headerStmt, opname.ID, opname.Note, opname.PerformedBy, opname.PerformedAt.Format(time.RFC3339), opname.CreatedAt.Format(time.RFC3339)); err != nil {
		return nil, fmt.Errorf("insert stock opname: %w", err)
	}

	const itemStmt = `INSERT INTO stock_opname_items (id, stock_opname_id, product_id, counted, previous_stock, difference) VALUES (?, ?, ?, ?, ?, ?);`
	for i := range opname.Items {
		item := &opname.Items[i]
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.StockOpnameID = opname.ID
		if _, err = tx.ExecContext(ctx, itemStmt, item.ID, item.StockOpnameID, item.ProductID, item.Counted, item.PreviousStock, item.Difference); err != nil {
			return nil, fmt.Errorf("insert stock opname item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit stock opname: %w", err)
	}
	return opname, nil
}

func (r *StockOpnameRepository) List(ctx context.Context, limit int) ([]domain.StockOpname, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	const stmt = `SELECT id, note, performed_by, performed_at, created_at FROM stock_opnames ORDER BY performed_at DESC LIMIT ?;`
	rows, err := r.db.QueryContext(ctx, stmt, limit)
	if err != nil {
		return nil, fmt.Errorf("list stock opnames: %w", err)
	}
	defer rows.Close()

	var items []domain.StockOpname
	for rows.Next() {
		var header domain.StockOpname
		var performed, created string
		if err := rows.Scan(&header.ID, &header.Note, &header.PerformedBy, &performed, &created); err != nil {
			return nil, err
		}
		header.PerformedAt, _ = time.Parse(time.RFC3339, performed)
		header.CreatedAt, _ = time.Parse(time.RFC3339, created)
		details, err := r.itemsByOpname(ctx, header.ID)
		if err != nil {
			return nil, err
		}
		header.Items = details
		items = append(items, header)
	}
	if items == nil {
		items = []domain.StockOpname{}
	}
	return items, nil
}

func (r *StockOpnameRepository) ListAll(ctx context.Context) ([]domain.StockOpname, error) {
	const stmt = `SELECT id, note, performed_by, performed_at, created_at FROM stock_opnames ORDER BY performed_at DESC;`
	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("list all stock opnames: %w", err)
	}
	defer rows.Close()

	var items []domain.StockOpname
	for rows.Next() {
		var header domain.StockOpname
		var performed, created string
		if err := rows.Scan(&header.ID, &header.Note, &header.PerformedBy, &performed, &created); err != nil {
			return nil, err
		}
		header.PerformedAt, _ = time.Parse(time.RFC3339, performed)
		header.CreatedAt, _ = time.Parse(time.RFC3339, created)
		details, err := r.itemsByOpname(ctx, header.ID)
		if err != nil {
			return nil, err
		}
		header.Items = details
		items = append(items, header)
	}
	if items == nil {
		items = []domain.StockOpname{}
	}
	return items, nil
}

func (r *StockOpnameRepository) itemsByOpname(ctx context.Context, opnameID string) ([]domain.StockOpnameItem, error) {
	const stmt = `SELECT i.id, i.stock_opname_id, i.product_id, p.name, p.sku, i.counted, i.previous_stock, i.difference
                FROM stock_opname_items i
                LEFT JOIN products p ON p.id = i.product_id
                WHERE i.stock_opname_id = ?
                ORDER BY p.name;`
	rows, err := r.db.QueryContext(ctx, stmt, opnameID)
	if err != nil {
		return nil, fmt.Errorf("list stock opname items: %w", err)
	}
	defer rows.Close()

	var items []domain.StockOpnameItem
	for rows.Next() {
		var item domain.StockOpnameItem
		if err := rows.Scan(&item.ID, &item.StockOpnameID, &item.ProductID, &item.ProductName, &item.ProductSKU, &item.Counted, &item.PreviousStock, &item.Difference); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []domain.StockOpnameItem{}
	}
	return items, nil
}

func (r *StockOpnameRepository) ReplaceAll(ctx context.Context, opnames []domain.StockOpname) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM stock_opname_items;`); err != nil {
		return fmt.Errorf("clear stock opname items: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM stock_opnames;`); err != nil {
		return fmt.Errorf("clear stock opnames: %w", err)
	}

	headerStmt, err := tx.PrepareContext(ctx, `INSERT INTO stock_opnames (id, note, performed_by, performed_at, created_at) VALUES (?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("prepare stock opname insert: %w", err)
	}
	defer headerStmt.Close()

	itemStmt, err := tx.PrepareContext(ctx, `INSERT INTO stock_opname_items (id, stock_opname_id, product_id, counted, previous_stock, difference) VALUES (?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("prepare stock opname item insert: %w", err)
	}
	defer itemStmt.Close()

	for _, opname := range opnames {
		performed := opname.PerformedAt
		if performed.IsZero() {
			performed = time.Now().UTC()
		}
		created := opname.CreatedAt
		if created.IsZero() {
			created = performed
		}
		if opname.ID == "" {
			opname.ID = uuid.New().String()
		}
		if _, err = headerStmt.ExecContext(ctx, opname.ID, opname.Note, opname.PerformedBy, performed.Format(time.RFC3339), created.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("insert stock opname header: %w", err)
		}

		for _, item := range opname.Items {
			itemID := item.ID
			if itemID == "" {
				itemID = uuid.New().String()
			}
			if _, err = itemStmt.ExecContext(ctx, itemID, opname.ID, item.ProductID, item.Counted, item.PreviousStock, item.Difference); err != nil {
				return fmt.Errorf("insert stock opname item: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit stock opname restore: %w", err)
	}
	return nil
}
