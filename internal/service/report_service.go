package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"smartseller-lite-starter/internal/db"
)

type OrderExportFilters struct {
	Search  string
	Courier string
	Start   *time.Time
	End     *time.Time
}

type ReportService struct {
	store *db.Store
}

func NewReportService(store *db.Store) *ReportService {
	return &ReportService{store: store}
}

func (s *ReportService) ExportOrdersCSV(ctx context.Context, filters OrderExportFilters) ([]byte, error) {
	query, args := buildOrderExportQuery(filters)

	rows, err := s.store.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query export rows: %w", err)
	}
	defer rows.Close()

	headers := []string{
		"Order Code",
		"Order Date",
		"Updated At",
		"Buyer Name",
		"Buyer Phone",
		"Buyer Email",
		"Buyer Address",
		"Recipient Name",
		"Recipient Phone",
		"Recipient Email",
		"Recipient Address",
		"Courier",
		"Service Level",
		"Tracking Code",
		"Shipping Cost",
		"Order Discount",
		"Order Total",
		"Order Profit",
		"Order Notes",
		"Product SKU",
		"Product Name",
		"Quantity",
		"Unit Price",
		"Item Discount",
		"Item Cost",
		"Item Profit",
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for rows.Next() {
		var (
			orderCode         string
			createdRaw        string
			updatedRaw        sql.NullString
			courier           sql.NullString
			service           sql.NullString
			tracking          sql.NullString
			shipping          sql.NullFloat64
			orderDiscount     sql.NullFloat64
			orderTotal        sql.NullFloat64
			orderProfit       sql.NullFloat64
			orderNotes        sql.NullString
			buyerName         sql.NullString
			buyerPhone        sql.NullString
			buyerEmail        sql.NullString
			buyerAddress      sql.NullString
			buyerCity         sql.NullString
			buyerProvince     sql.NullString
			buyerPostal       sql.NullString
			recipientName     sql.NullString
			recipientPhone    sql.NullString
			recipientEmail    sql.NullString
			recipientAddress  sql.NullString
			recipientCity     sql.NullString
			recipientProvince sql.NullString
			recipientPostal   sql.NullString
			quantity          sql.NullInt64
			unitPrice         sql.NullFloat64
			itemDiscount      sql.NullFloat64
			itemCost          sql.NullFloat64
			itemProfit        sql.NullFloat64
			productSKU        sql.NullString
			productName       sql.NullString
		)

		if err := rows.Scan(
			&orderCode,
			&createdRaw,
			&updatedRaw,
			&courier,
			&service,
			&tracking,
			&shipping,
			&orderDiscount,
			&orderTotal,
			&orderProfit,
			&orderNotes,
			&buyerName,
			&buyerPhone,
			&buyerEmail,
			&buyerAddress,
			&buyerCity,
			&buyerProvince,
			&buyerPostal,
			&recipientName,
			&recipientPhone,
			&recipientEmail,
			&recipientAddress,
			&recipientCity,
			&recipientProvince,
			&recipientPostal,
			&quantity,
			&unitPrice,
			&itemDiscount,
			&itemCost,
			&itemProfit,
			&productSKU,
			&productName,
		); err != nil {
			return nil, fmt.Errorf("scan export row: %w", err)
		}

		record := []string{
			orderCode,
			formatTimestamp(createdRaw),
			formatTimestampNull(updatedRaw),
			valueOrEmpty(buyerName),
			valueOrEmpty(buyerPhone),
			valueOrEmpty(buyerEmail),
			joinAddress(buyerAddress, buyerCity, buyerProvince, buyerPostal),
			valueOrEmpty(recipientName),
			valueOrEmpty(recipientPhone),
			valueOrEmpty(recipientEmail),
			joinAddress(recipientAddress, recipientCity, recipientProvince, recipientPostal),
			valueOrEmpty(courier),
			valueOrEmpty(service),
			valueOrEmpty(tracking),
			formatFloat(shipping),
			formatFloat(orderDiscount),
			formatFloat(orderTotal),
			formatFloat(orderProfit),
			valueOrEmpty(orderNotes),
			valueOrEmpty(productSKU),
			valueOrEmpty(productName),
			formatInt(quantity),
			formatFloat(unitPrice),
			formatFloat(itemDiscount),
			formatFloat(itemCost),
			formatFloat(itemProfit),
		}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate export rows: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}

func buildOrderExportQuery(filters OrderExportFilters) (string, []any) {
	base := `SELECT
  o.code,
  o.created_at,
  o.updated_at,
  o.shipment_courier,
  o.shipment_service,
  o.shipment_tracking,
  o.shipment_cost,
  o.discount,
  o.total,
  o.profit,
  o.notes,
  buyer.name,
  buyer.phone,
  buyer.email,
  buyer.address,
  buyer.city,
  buyer.province,
  buyer.postal,
  recipient.name,
  recipient.phone,
  recipient.email,
  recipient.address,
  recipient.city,
  recipient.province,
  recipient.postal,
  items.quantity,
  items.unit_price,
  items.discount,
  items.cost_price,
  items.profit,
  products.sku,
  products.name
FROM orders o
LEFT JOIN customers buyer ON buyer.id = o.buyer_id
LEFT JOIN customers recipient ON recipient.id = o.recipient_id
LEFT JOIN order_items items ON items.order_id = o.id
LEFT JOIN products products ON products.id = items.product_id`

	var conditions []string
	var args []any

	if filters.Courier != "" && strings.ToLower(filters.Courier) != "all" {
		conditions = append(conditions, "COALESCE(o.shipment_courier, '') = ?")
		args = append(args, filters.Courier)
	}
	if filters.Start != nil {
		conditions = append(conditions, "o.created_at >= ?")
		args = append(args, filters.Start.Format(time.RFC3339))
	}
	if filters.End != nil {
		conditions = append(conditions, "o.created_at <= ?")
		args = append(args, filters.End.Format(time.RFC3339))
	}
	if trimmed := strings.TrimSpace(filters.Search); trimmed != "" {
		lowered := strings.ToLower(trimmed)
		like := "%" + lowered + "%"
		conditions = append(conditions, `(
  LOWER(o.code) LIKE ? OR
  LOWER(COALESCE(o.shipment_courier, '')) LIKE ? OR
  LOWER(COALESCE(o.shipment_service, '')) LIKE ? OR
  LOWER(COALESCE(o.shipment_tracking, '')) LIKE ? OR
  LOWER(COALESCE(o.notes, '')) LIKE ? OR
  LOWER(COALESCE(buyer.name, '')) LIKE ? OR
  LOWER(COALESCE(recipient.name, '')) LIKE ? OR
  LOWER(COALESCE(products.name, '')) LIKE ? OR
  LOWER(COALESCE(products.sku, '')) LIKE ?
)`)
		for i := 0; i < 9; i++ {
			args = append(args, like)
		}
	}

	if len(conditions) > 0 {
		base += "\nWHERE " + strings.Join(conditions, " AND ")
	}

	base += "\nORDER BY o.created_at DESC, o.id, items.id;"
	return base, args
}

func valueOrEmpty(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func joinAddress(addr, city, province, postal sql.NullString) string {
	parts := []string{valueOrEmpty(addr), valueOrEmpty(city), valueOrEmpty(province), valueOrEmpty(postal)}
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, ", ")
}

func formatTimestamp(value string) string {
	if value == "" {
		return ""
	}
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		return ts.Format("2006-01-02 15:04:05")
	}
	return value
}

func formatTimestampNull(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return formatTimestamp(v.String)
}

func formatFloat(v sql.NullFloat64) string {
	if !v.Valid {
		return ""
	}
	return strconv.FormatFloat(v.Float64, 'f', 2, 64)
}

func formatInt(v sql.NullInt64) string {
	if !v.Valid {
		return ""
	}
	return strconv.FormatInt(v.Int64, 10)
}
