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

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) ListPaged(ctx context.Context, opts ProductListOptions) (ProductListResult, error) {
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
	if !opts.IncludeArchived {
		whereParts = append(whereParts, "deleted_at IS NULL")
	}
	args := make([]any, 0)

	query := strings.TrimSpace(strings.ToLower(opts.Query))
	if query != "" {
		like := "%" + query + "%"
		whereParts = append(whereParts, "(LOWER(name) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(IFNULL(category,'')) LIKE ?)")
		args = append(args, like, like, like)
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

	stmt := "SELECT id, name, sku, cost_price, sale_price, stock, category, low_stock_threshold, description, image_path, thumb_path, image_hash, image_width, image_height, image_size_bytes, thumb_width, thumb_height, thumb_size_bytes, deleted_at, created_at, updated_at FROM products " + whereClause + " ORDER BY name" + limitClause + ";"

	rows, err := r.db.QueryContext(ctx, stmt, listArgs...)
	if err != nil {
		return ProductListResult{}, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Product, 0)
	for rows.Next() {
		var p domain.Product
		var created, updated string
		var deleted sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.CostPrice, &p.SalePrice, &p.Stock, &p.Category, &p.LowStockThreshold, &p.Description, &p.ImagePath, &p.ThumbPath, &p.ImageHash, &p.ImageWidth, &p.ImageHeight, &p.ImageSizeBytes, &p.ThumbWidth, &p.ThumbHeight, &p.ThumbSizeBytes, &deleted, &created, &updated); err != nil {
			return ProductListResult{}, err
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, created)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		if deleted.Valid {
			ts, _ := time.Parse(time.RFC3339, deleted.String)
			p.DeletedAt = &ts
		}
		if p.LowStockThreshold <= 0 {
			p.LowStockThreshold = 5
		}
		items = append(items, p)
	}

	countStmt := "SELECT COUNT(*) FROM products " + whereClause + ";"
	var total int
	if err := r.db.QueryRowContext(ctx, countStmt, args...).Scan(&total); err != nil {
		return ProductListResult{}, fmt.Errorf("count products: %w", err)
	}

	// counts for out of stock and warning stock (stock >0 && <= threshold)
	thresholdExpr := "COALESCE(NULLIF(low_stock_threshold,0),5)"

	outClause := whereClause
	if outClause != "" {
		outClause += " AND "
	} else {
		outClause = "WHERE "
	}
	outClause += "stock <= 0"
	outStmt := "SELECT COUNT(*) FROM products " + outClause + ";"
	var outOfStock int
	if err := r.db.QueryRowContext(ctx, outStmt, args...).Scan(&outOfStock); err != nil {
		return ProductListResult{}, fmt.Errorf("count out of stock: %w", err)
	}

	warnClause := whereClause
	if warnClause != "" {
		warnClause += " AND "
	} else {
		warnClause = "WHERE "
	}
	warnClause += fmt.Sprintf("stock > 0 AND stock <= %s", thresholdExpr)
	warnStmt := "SELECT COUNT(*) FROM products " + warnClause + ";"
	var warning int
	if err := r.db.QueryRowContext(ctx, warnStmt, args...).Scan(&warning); err != nil {
		return ProductListResult{}, fmt.Errorf("count warning stock: %w", err)
	}

	// highlights - up to 5 products low stock (including zero)
	highlightClause := whereClause
	if highlightClause != "" {
		highlightClause += " AND "
	} else {
		highlightClause = "WHERE "
	}
	highlightClause += fmt.Sprintf("stock <= %s", thresholdExpr)
	highlightStmt := "SELECT id, name, sku, cost_price, sale_price, stock, category, low_stock_threshold, description, image_path, thumb_path, image_hash, image_width, image_height, image_size_bytes, thumb_width, thumb_height, thumb_size_bytes, deleted_at, created_at, updated_at FROM products " + highlightClause + " ORDER BY stock ASC, name ASC LIMIT 5;"
	highlightRows, err := r.db.QueryContext(ctx, highlightStmt, args...)
	if err != nil {
		return ProductListResult{}, fmt.Errorf("highlight low stock: %w", err)
	}
	defer highlightRows.Close()

	highlights := make([]domain.Product, 0)
	for highlightRows.Next() {
		var p domain.Product
		var created, updated string
		var deleted sql.NullString
		if err := highlightRows.Scan(&p.ID, &p.Name, &p.SKU, &p.CostPrice, &p.SalePrice, &p.Stock, &p.Category, &p.LowStockThreshold, &p.Description, &p.ImagePath, &p.ThumbPath, &p.ImageHash, &p.ImageWidth, &p.ImageHeight, &p.ImageSizeBytes, &p.ThumbWidth, &p.ThumbHeight, &p.ThumbSizeBytes, &deleted, &created, &updated); err != nil {
			return ProductListResult{}, err
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, created)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		if deleted.Valid {
			ts, _ := time.Parse(time.RFC3339, deleted.String)
			p.DeletedAt = &ts
		}
		if p.LowStockThreshold <= 0 {
			p.LowStockThreshold = 5
		}
		highlights = append(highlights, p)
	}

	result := ProductListResult{
		Items:              items,
		Total:              total,
		Page:               page,
		PageSize:           pageSize,
		OutOfStockCount:    outOfStock,
		WarningStockCount:  warning,
		LowStockHighlights: highlights,
	}

	if pageSize <= 0 {
		result.Page = 1
		result.PageSize = len(items)
	}

	return result, nil
}

type ProductListOptions struct {
	Query           string
	Page            int
	PageSize        int
	IncludeArchived bool
}

type ProductListResult struct {
	Items              []domain.Product
	Total              int
	Page               int
	PageSize           int
	OutOfStockCount    int
	WarningStockCount  int
	LowStockHighlights []domain.Product
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	now := time.Now().UTC()
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.LowStockThreshold <= 0 {
		p.LowStockThreshold = 5
	}
	p.CreatedAt = now
	p.UpdatedAt = now

	const stmt = `INSERT INTO products (id, name, sku, cost_price, sale_price, stock, category, low_stock_threshold, description, image_path, thumb_path, image_hash, image_width, image_height, image_size_bytes, thumb_width, thumb_height, thumb_size_bytes, deleted_at, created_at, updated_at)
                  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	var deleted interface{}
	if p.DeletedAt != nil {
		deleted = p.DeletedAt.Format(time.RFC3339)
	}
	_, err := r.db.ExecContext(ctx, stmt, p.ID, p.Name, p.SKU, p.CostPrice, p.SalePrice, p.Stock, p.Category, p.LowStockThreshold, p.Description, p.ImagePath, p.ThumbPath, p.ImageHash, p.ImageWidth, p.ImageHeight, p.ImageSizeBytes, p.ThumbWidth, p.ThumbHeight, p.ThumbSizeBytes, deleted, p.CreatedAt.Format(time.RFC3339), p.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("insert product: %w", err)
	}
	return p, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	p.UpdatedAt = time.Now().UTC()
	if p.LowStockThreshold <= 0 {
		p.LowStockThreshold = 5
	}
	const stmt = `UPDATE products SET name = ?, sku = ?, cost_price = ?, sale_price = ?, stock = ?, category = ?, low_stock_threshold = ?, description = ?, image_path = ?, thumb_path = ?, image_hash = ?, image_width = ?, image_height = ?, image_size_bytes = ?, thumb_width = ?, thumb_height = ?, thumb_size_bytes = ?, deleted_at = ?, updated_at = ? WHERE id = ?;`
	var deleted interface{}
	if p.DeletedAt != nil {
		deleted = p.DeletedAt.Format(time.RFC3339)
	}
	if _, err := r.db.ExecContext(ctx, stmt, p.Name, p.SKU, p.CostPrice, p.SalePrice, p.Stock, p.Category, p.LowStockThreshold, p.Description, p.ImagePath, p.ThumbPath, p.ImageHash, p.ImageWidth, p.ImageHeight, p.ImageSizeBytes, p.ThumbWidth, p.ThumbHeight, p.ThumbSizeBytes, deleted, p.UpdatedAt.Format(time.RFC3339), p.ID); err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}
	return p, nil
}

func (r *ProductRepository) AdjustStock(ctx context.Context, productID string, delta int, reason string) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("commit stock adjustment: %w", commitErr)
		}
	}()

	const selectStmt = `SELECT stock FROM products WHERE id = ? AND deleted_at IS NULL;`
	var stock int
	if err = tx.QueryRowContext(ctx, selectStmt, productID).Scan(&stock); err != nil {
		err = fmt.Errorf("select stock: %w", err)
		return err
	}
	stock += delta
	if stock < 0 {
		err = fmt.Errorf("insufficient stock for product %s", productID)
		return err
	}

	const updateStmt = `UPDATE products SET stock = ?, updated_at = ? WHERE id = ?;`
	if _, err = tx.ExecContext(ctx, updateStmt, stock, time.Now().UTC().Format(time.RFC3339), productID); err != nil {
		err = fmt.Errorf("update stock: %w", err)
		return err
	}

	const mutationStmt = `INSERT INTO stock_mutations (id, product_id, delta, reason, created_at) VALUES (?, ?, ?, ?, ?);`
	if _, err = tx.ExecContext(ctx, mutationStmt, uuid.New().String(), productID, delta, reason, time.Now().UTC().Format(time.RFC3339)); err != nil {
		err = fmt.Errorf("insert stock mutation: %w", err)
		return err
	}

	return nil
}

func (r *ProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	res, err := r.ListPaged(ctx, ProductListOptions{Page: 1, PageSize: 0, IncludeArchived: false})
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (r *ProductRepository) ListIncludingArchived(ctx context.Context) ([]domain.Product, error) {
	res, err := r.ListPaged(ctx, ProductListOptions{Page: 1, PageSize: 0, IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (r *ProductRepository) Get(ctx context.Context, id string) (*domain.Product, error) {
	const stmt = `SELECT id, name, sku, cost_price, sale_price, stock, category, low_stock_threshold, description, image_path, thumb_path, image_hash, image_width, image_height, image_size_bytes, thumb_width, thumb_height, thumb_size_bytes, deleted_at, created_at, updated_at FROM products WHERE id = ?;`
	var p domain.Product
	var created, updated string
	var deleted sql.NullString
	if err := r.db.QueryRowContext(ctx, stmt, id).Scan(&p.ID, &p.Name, &p.SKU, &p.CostPrice, &p.SalePrice, &p.Stock, &p.Category, &p.LowStockThreshold, &p.Description, &p.ImagePath, &p.ThumbPath, &p.ImageHash, &p.ImageWidth, &p.ImageHeight, &p.ImageSizeBytes, &p.ThumbWidth, &p.ThumbHeight, &p.ThumbSizeBytes, &deleted, &created, &updated); err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	p.CreatedAt, _ = time.Parse(time.RFC3339, created)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	if deleted.Valid {
		ts, _ := time.Parse(time.RFC3339, deleted.String)
		p.DeletedAt = &ts
	}
	if p.LowStockThreshold <= 0 {
		p.LowStockThreshold = 5
	}
	return &p, nil
}

func (r *ProductRepository) Archive(ctx context.Context, id string) error {
	const stmt = `UPDATE products SET deleted_at = ?, updated_at = ? WHERE id = ?;`
	now := time.Now().UTC()
	if _, err := r.db.ExecContext(ctx, stmt, now.Format(time.RFC3339), now.Format(time.RFC3339), id); err != nil {
		return fmt.Errorf("archive product: %w", err)
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product id required")
	}
	const stmt = `DELETE FROM products WHERE id = ?;`
	if _, err := r.db.ExecContext(ctx, stmt, id); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1451 {
			return fmt.Errorf("delete product: produk masih digunakan pada transaksi")
		}
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

func (r *ProductRepository) ReplaceAll(ctx context.Context, items []domain.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM stock_mutations;`); err != nil {
		return fmt.Errorf("clear stock mutations: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM products;`); err != nil {
		return fmt.Errorf("clear products: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO products (id, name, sku, cost_price, sale_price, stock, category, low_stock_threshold, description, image_path, thumb_path, image_hash, image_width, image_height, image_size_bytes, thumb_width, thumb_height, thumb_size_bytes, deleted_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("prepare product insert: %w", err)
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

		var deleted interface{}
		if item.DeletedAt != nil {
			deleted = item.DeletedAt.Format(time.RFC3339)
		}
		threshold := item.LowStockThreshold
		if threshold <= 0 {
			threshold = 5
		}
		if _, err = stmt.ExecContext(ctx, id, item.Name, item.SKU, item.CostPrice, item.SalePrice, item.Stock, item.Category, threshold, item.Description, item.ImagePath, item.ThumbPath, item.ImageHash, item.ImageWidth, item.ImageHeight, item.ImageSizeBytes, item.ThumbWidth, item.ThumbHeight, item.ThumbSizeBytes, deleted, created.Format(time.RFC3339), updated.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("insert product from backup: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit product restore: %w", err)
	}
	return nil
}
