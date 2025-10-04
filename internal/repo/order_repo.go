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

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) ListPaged(ctx context.Context, opts OrderListOptions) (OrderListResult, error) {
	const maxPageSize = 100

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
		whereParts = append(whereParts, "(LOWER(o.code) LIKE ? OR LOWER(IFNULL(o.notes,'')) LIKE ? OR LOWER(IFNULL(o.shipment_courier,'')) LIKE ? OR LOWER(IFNULL(o.shipment_service,'')) LIKE ? OR LOWER(IFNULL(o.shipment_tracking,'')) LIKE ? OR EXISTS (SELECT 1 FROM customers cb WHERE cb.id = o.buyer_id AND LOWER(cb.name) LIKE ?) OR EXISTS (SELECT 1 FROM customers cr WHERE cr.id = o.recipient_id AND LOWER(cr.name) LIKE ?) OR EXISTS (SELECT 1 FROM order_items oi LEFT JOIN products p ON p.id = oi.product_id WHERE oi.order_id = o.id AND (LOWER(IFNULL(p.name,'')) LIKE ? OR LOWER(IFNULL(p.sku,'')) LIKE ?)))")
		args = append(args, like, like, like, like, like, like, like, like, like)
	}

	courier := strings.TrimSpace(opts.Courier)
	if courier != "" && strings.ToLower(courier) != "all" {
		whereParts = append(whereParts, "LOWER(IFNULL(o.shipment_courier,'')) = ?")
		args = append(args, strings.ToLower(courier))
	}

	if opts.DateStart != nil {
		whereParts = append(whereParts, "o.created_at >= ?")
		args = append(args, opts.DateStart.UTC().Format(time.RFC3339))
	}
	if opts.DateEnd != nil {
		end := opts.DateEnd.UTC().Add(24 * time.Hour)
		whereParts = append(whereParts, "o.created_at < ?")
		args = append(args, end.Format(time.RFC3339))
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

	stmt := "SELECT o.id, o.code, o.buyer_id, o.recipient_id, o.shipment_courier, o.shipment_service, o.shipment_tracking, o.shipment_cost, o.is_buyer_paying_shipping, o.discount_order, o.total, o.profit, o.notes, o.created_at, o.updated_at FROM orders o " + whereClause + " ORDER BY o.created_at DESC" + limitClause + ";"

	rows, err := r.db.QueryContext(ctx, stmt, listArgs...)
	if err != nil {
		return OrderListResult{}, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Order, 0)
	for rows.Next() {
		var o domain.Order
		var created, updated string
		if err := rows.Scan(&o.ID, &o.Code, &o.BuyerID, &o.RecipientID, &o.Shipment.Courier, &o.Shipment.ServiceLevel, &o.Shipment.TrackingCode, &o.Shipment.ShippingCost, &o.Shipment.ShippingByBuyer, &o.DiscountOrder, &o.Total, &o.Profit, &o.Notes, &created, &updated); err != nil {
			return OrderListResult{}, err
		}
		o.CreatedAt, _ = time.Parse(time.RFC3339, created)
		o.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		itemsList, err := r.itemsByOrder(ctx, o.ID)
		if err != nil {
			return OrderListResult{}, err
		}
		o.Items = itemsList
		items = append(items, o)
	}

	countStmt := "SELECT COUNT(*) FROM orders o " + whereClause + ";"
	var total int
	if err := r.db.QueryRowContext(ctx, countStmt, args...).Scan(&total); err != nil {
		return OrderListResult{}, fmt.Errorf("count orders: %w", err)
	}

	summary := OrderListSummary{Count: total}

	sumStmt := "SELECT COALESCE(SUM(o.total),0), COALESCE(SUM(o.profit),0) FROM orders o " + whereClause + ";"
	if err := r.db.QueryRowContext(ctx, sumStmt, args...).Scan(&summary.Revenue, &summary.Profit); err != nil {
		return OrderListResult{}, fmt.Errorf("summary totals: %w", err)
	}

	courierStmt := "SELECT IFNULL(o.shipment_courier, '') AS courier, COUNT(*) AS hits FROM orders o " + whereClause + " GROUP BY courier ORDER BY hits DESC LIMIT 1;"
	if err := r.db.QueryRowContext(ctx, courierStmt, args...).Scan(&summary.TopCourier, &summary.TopCourierHits); err != nil && err != sql.ErrNoRows {
		return OrderListResult{}, fmt.Errorf("top courier: %w", err)
	}

	productStmt := "SELECT IFNULL(p.id,''), IFNULL(p.name,''), SUM(oi.quantity) AS qty FROM order_items oi JOIN orders o ON o.id = oi.order_id LEFT JOIN products p ON p.id = oi.product_id " + whereClause + " GROUP BY p.id, p.name ORDER BY qty DESC LIMIT 1;"
	if err := r.db.QueryRowContext(ctx, productStmt, args...).Scan(&summary.TopProductID, &summary.TopProductName, &summary.TopProductQty); err != nil && err != sql.ErrNoRows {
		return OrderListResult{}, fmt.Errorf("top product: %w", err)
	}

	couriersStmt := "SELECT DISTINCT IFNULL(o.shipment_courier,'') FROM orders o " + whereClause + " ORDER BY 1;"
	courierRows, err := r.db.QueryContext(ctx, couriersStmt, args...)
	if err != nil {
		return OrderListResult{}, fmt.Errorf("list couriers: %w", err)
	}
	defer courierRows.Close()

	couriers := make([]string, 0)
	for courierRows.Next() {
		var name string
		if err := courierRows.Scan(&name); err != nil {
			return OrderListResult{}, err
		}
		couriers = append(couriers, name)
	}

	result := OrderListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Summary:  summary,
		Couriers: couriers,
	}
	if pageSize <= 0 {
		result.Page = 1
		result.PageSize = len(items)
	}

	return result, nil
}

type OrderListOptions struct {
	Query     string
	Courier   string
	DateStart *time.Time
	DateEnd   *time.Time
	Page      int
	PageSize  int
}

type OrderListSummary struct {
	Count          int     `json:"count"`
	Revenue        float64 `json:"revenue"`
	Profit         float64 `json:"profit"`
	TopCourier     string  `json:"topCourier"`
	TopCourierHits int     `json:"topCourierHits"`
	TopProductID   string  `json:"topProductId"`
	TopProductName string  `json:"topProductName"`
	TopProductQty  int     `json:"topProductQty"`
}

type OrderListResult struct {
	Items    []domain.Order   `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Summary  OrderListSummary `json:"summary"`
	Couriers []string         `json:"couriers"`
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) (*domain.Order, error) {
	now := time.Now().UTC()
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	if o.Code == "" {
		o.Code = fmt.Sprintf("ORD-%s", now.Format("200601021504"))
	}
	o.CreatedAt = now
	o.UpdatedAt = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const orderStmt = `INSERT INTO orders (
        id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost, is_buyer_paying_shipping, discount_order, total, profit, notes, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err = tx.ExecContext(ctx, orderStmt,
		o.ID, o.Code, o.BuyerID, o.RecipientID,
		o.Shipment.Courier, o.Shipment.ServiceLevel, o.Shipment.TrackingCode, o.Shipment.ShippingCost, o.Shipment.ShippingByBuyer,
		o.DiscountOrder, o.Total, o.Profit, o.Notes,
		o.CreatedAt.Format(time.RFC3339), o.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	const itemStmt = `INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, discount_item, cost_price, profit) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
	for i := range o.Items {
		item := &o.Items[i]
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.OrderID = o.ID
		if _, err = tx.ExecContext(ctx, itemStmt, item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPrice, item.DiscountItem, item.CostPrice, item.Profit); err != nil {
			return nil, fmt.Errorf("insert order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit order: %w", err)
	}
	return o, nil
}

func (r *OrderRepository) List(ctx context.Context, limit int) ([]domain.Order, error) {
	result, err := r.ListPaged(ctx, OrderListOptions{Page: 1, PageSize: limit})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (r *OrderRepository) ListAll(ctx context.Context) ([]domain.Order, error) {
	const stmt = `SELECT id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost, is_buyer_paying_shipping, discount_order, total, profit, notes, created_at, updated_at FROM orders ORDER BY created_at DESC;`

	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("list all orders: %w", err)
	}
	defer rows.Close()

	orders := []domain.Order{}
	for rows.Next() {
		var o domain.Order
		var created, updated string
		if err := rows.Scan(&o.ID, &o.Code, &o.BuyerID, &o.RecipientID, &o.Shipment.Courier, &o.Shipment.ServiceLevel, &o.Shipment.TrackingCode, &o.Shipment.ShippingCost, &o.Shipment.ShippingByBuyer, &o.DiscountOrder, &o.Total, &o.Profit, &o.Notes, &created, &updated); err != nil {
			return nil, err
		}
		o.CreatedAt, _ = time.Parse(time.RFC3339, created)
		o.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		items, err := r.itemsByOrder(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepository) Get(ctx context.Context, id string) (*domain.Order, error) {
	const stmt = `SELECT id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost,is_buyer_paying_shipping, discount_order, total, profit, notes, created_at, updated_at
                  FROM orders WHERE id = ?;`
	var o domain.Order
	var created, updated string
	if err := r.db.QueryRowContext(ctx, stmt, id).Scan(&o.ID, &o.Code, &o.BuyerID, &o.RecipientID, &o.Shipment.Courier, &o.Shipment.ServiceLevel, &o.Shipment.TrackingCode, &o.Shipment.ShippingCost, &o.Shipment.ShippingByBuyer, &o.DiscountOrder, &o.Total, &o.Profit, &o.Notes, &created, &updated); err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	o.CreatedAt, _ = time.Parse(time.RFC3339, created)
	o.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	items, err := r.itemsByOrder(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

func (r *OrderRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = ?;`, id); err != nil {
		return fmt.Errorf("delete order items: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM orders WHERE id = ?;`, id); err != nil {
		return fmt.Errorf("delete order: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit delete order: %w", err)
	}

	return nil
}

func (r *OrderRepository) itemsByOrder(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	const stmt = `SELECT i.id, i.order_id, i.product_id, p.sku, i.quantity, i.unit_price, i.discount_item, i.cost_price, i.profit
                FROM order_items i
                LEFT JOIN products p ON p.id = i.product_id
                WHERE i.order_id = ?;`
	rows, err := r.db.QueryContext(ctx, stmt, orderID)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		var sku sql.NullString
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &sku, &item.Quantity, &item.UnitPrice, &item.DiscountItem, &item.CostPrice, &item.Profit); err != nil {
			return nil, err
		}
		if sku.Valid {
			item.SKU = sku.String
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *OrderRepository) ReplaceAll(ctx context.Context, orders []domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM order_items;`); err != nil {
		return fmt.Errorf("clear order items: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM orders;`); err != nil {
		return fmt.Errorf("clear orders: %w", err)
	}

	orderStmt, err := tx.PrepareContext(ctx, `INSERT INTO orders (id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost, is_buyer_paying_shipping,  discount_order, total, profit, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("prepare order insert: %w", err)
	}
	defer orderStmt.Close()

	itemStmt, err := tx.PrepareContext(ctx, `INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, discount_item, cost_price, profit) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("prepare order item insert: %w", err)
	}
	defer itemStmt.Close()

	now := time.Now().UTC()
	for _, order := range orders {
		id := order.ID
		if id == "" {
			id = uuid.New().String()
		}
		created := order.CreatedAt
		if created.IsZero() {
			created = now
		}
		updated := order.UpdatedAt
		if updated.IsZero() {
			updated = created
		}
		code := order.Code
		if code == "" {
			code = fmt.Sprintf("ORD-%s", created.Format("200601021504"))
		}

		if _, err = orderStmt.ExecContext(ctx, id, code, order.BuyerID, order.RecipientID, order.Shipment.Courier, order.Shipment.ServiceLevel, order.Shipment.TrackingCode, order.Shipment.ShippingCost, order.Shipment.ShippingByBuyer, order.DiscountOrder, order.Total, order.Profit, order.Notes, created.Format(time.RFC3339), updated.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("insert order from backup: %w", err)
		}

		for _, item := range order.Items {
			itemID := item.ID
			if itemID == "" {
				itemID = uuid.New().String()
			}
			if _, err = itemStmt.ExecContext(ctx, itemID, id, item.ProductID, item.Quantity, item.UnitPrice, item.DiscountItem, item.CostPrice, item.Profit); err != nil {
				return fmt.Errorf("insert order item from backup: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit order restore: %w", err)
	}
	return nil
}
