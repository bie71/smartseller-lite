package repo

import (
	"context"
	"database/sql"
	"fmt"
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
        id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost, discount, total, profit, notes, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err = tx.ExecContext(ctx, orderStmt,
		o.ID, o.Code, o.BuyerID, o.RecipientID,
		o.Shipment.Courier, o.Shipment.ServiceLevel, o.Shipment.TrackingCode, o.Shipment.ShippingCost,
		o.Discount, o.Total, o.Profit, o.Notes,
		o.CreatedAt.Format(time.RFC3339), o.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	const itemStmt = `INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, discount, cost_price, profit) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
	for i := range o.Items {
		item := &o.Items[i]
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.OrderID = o.ID
		if _, err = tx.ExecContext(ctx, itemStmt, item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPrice, item.Discount, item.CostPrice, item.Profit); err != nil {
			return nil, fmt.Errorf("insert order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit order: %w", err)
	}
	return o, nil
}

func (r *OrderRepository) List(ctx context.Context, limit int) ([]domain.Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	const stmt = `SELECT id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost, discount, total, profit, notes, created_at, updated_at
                 FROM orders ORDER BY created_at DESC LIMIT ?;`

	rows, err := r.db.QueryContext(ctx, stmt, limit)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := []domain.Order{}
	for rows.Next() {
		var o domain.Order
		var created, updated string
		if err := rows.Scan(&o.ID, &o.Code, &o.BuyerID, &o.RecipientID, &o.Shipment.Courier, &o.Shipment.ServiceLevel, &o.Shipment.TrackingCode, &o.Shipment.ShippingCost, &o.Discount, &o.Total, &o.Profit, &o.Notes, &created, &updated); err != nil {
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

func (r *OrderRepository) ListAll(ctx context.Context) ([]domain.Order, error) {
	const stmt = `SELECT id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost, discount, total, profit, notes, created_at, updated_at FROM orders ORDER BY created_at DESC;`

	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("list all orders: %w", err)
	}
	defer rows.Close()

	orders := []domain.Order{}
	for rows.Next() {
		var o domain.Order
		var created, updated string
		if err := rows.Scan(&o.ID, &o.Code, &o.BuyerID, &o.RecipientID, &o.Shipment.Courier, &o.Shipment.ServiceLevel, &o.Shipment.TrackingCode, &o.Shipment.ShippingCost, &o.Discount, &o.Total, &o.Profit, &o.Notes, &created, &updated); err != nil {
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
	const stmt = `SELECT id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost, discount, total, profit, notes, created_at, updated_at
                  FROM orders WHERE id = ?;`
	var o domain.Order
	var created, updated string
	if err := r.db.QueryRowContext(ctx, stmt, id).Scan(&o.ID, &o.Code, &o.BuyerID, &o.RecipientID, &o.Shipment.Courier, &o.Shipment.ServiceLevel, &o.Shipment.TrackingCode, &o.Shipment.ShippingCost, &o.Discount, &o.Total, &o.Profit, &o.Notes, &created, &updated); err != nil {
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

func (r *OrderRepository) itemsByOrder(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	const stmt = `SELECT i.id, i.order_id, i.product_id, p.sku, i.quantity, i.unit_price, i.discount, i.cost_price, i.profit
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
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &sku, &item.Quantity, &item.UnitPrice, &item.Discount, &item.CostPrice, &item.Profit); err != nil {
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

	orderStmt, err := tx.PrepareContext(ctx, `INSERT INTO orders (id, code, buyer_id, recipient_id, shipment_courier, shipment_service, shipment_tracking, shipment_cost, discount, total, profit, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("prepare order insert: %w", err)
	}
	defer orderStmt.Close()

	itemStmt, err := tx.PrepareContext(ctx, `INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, discount, cost_price, profit) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`)
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

		if _, err = orderStmt.ExecContext(ctx, id, code, order.BuyerID, order.RecipientID, order.Shipment.Courier, order.Shipment.ServiceLevel, order.Shipment.TrackingCode, order.Shipment.ShippingCost, order.Discount, order.Total, order.Profit, order.Notes, created.Format(time.RFC3339), updated.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("insert order from backup: %w", err)
		}

		for _, item := range order.Items {
			itemID := item.ID
			if itemID == "" {
				itemID = uuid.New().String()
			}
			if _, err = itemStmt.ExecContext(ctx, itemID, id, item.ProductID, item.Quantity, item.UnitPrice, item.Discount, item.CostPrice, item.Profit); err != nil {
				return fmt.Errorf("insert order item from backup: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit order restore: %w", err)
	}
	return nil
}
