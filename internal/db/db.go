package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"smartseller-lite-starter/internal/repo"
)

// Store wraps the MySQL connection and provides repositories on demand.
type Store struct {
	db              *sql.DB
	cfg             *mysql.Config
	productRepo     *repo.ProductRepository
	customerRepo    *repo.CustomerRepository
	orderRepo       *repo.OrderRepository
	settingsRepo    *repo.SettingsRepository
	courierRepo     *repo.CourierRepository
	stockOpnameRepo *repo.StockOpnameRepository
}

// NewStore initialises a new Store using the provided MySQL DSN.
func NewStore(dsn string) (*Store, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("mysql dsn is required")
	}

	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse mysql dsn: %w", err)
	}
	if cfg.DBName == "" {
		return nil, errors.New("mysql dsn must include database name")
	}
	if cfg.Params == nil {
		cfg.Params = map[string]string{}
	}
	if cfg.Net == "" {
		cfg.Net = "tcp"
	}
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:3306"
	}
	cfg.ParseTime = true
	cfg.MultiStatements = true
	if _, ok := cfg.Params["charset"]; !ok {
		cfg.Params["charset"] = "utf8mb4"
	}
	if cfg.Loc == nil {
		cfg.Loc = time.UTC
	}

	conn, err := openMySQLWithRetry(cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	store := &Store{db: conn, cfg: cfg}
	if err := store.initSchema(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS products (
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            sku VARCHAR(191) NOT NULL,
            cost_price DOUBLE NOT NULL DEFAULT 0,
            sale_price DOUBLE NOT NULL DEFAULT 0,
            stock INT NOT NULL DEFAULT 0,
            category VARCHAR(191),
            low_stock_threshold INT NOT NULL DEFAULT 5,
            description TEXT,
            image_path VARCHAR(255),
            thumb_path VARCHAR(255),
            image_hash CHAR(64),
            image_width INT,
            image_height INT,
            image_size_bytes BIGINT,
            thumb_width INT,
            thumb_height INT,
            thumb_size_bytes BIGINT,
            deleted_at VARCHAR(64),
            created_at VARCHAR(64) NOT NULL,
            updated_at VARCHAR(64) NOT NULL,
            UNIQUE KEY idx_products_sku (sku)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS customers (
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            type VARCHAR(32) NOT NULL,
            name VARCHAR(255) NOT NULL,
            phone VARCHAR(64),
            email VARCHAR(255),
            address TEXT,
            city VARCHAR(191),
            province VARCHAR(191),
            postal VARCHAR(32),
            notes TEXT,
            created_at VARCHAR(64) NOT NULL,
            updated_at VARCHAR(64) NOT NULL
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS orders (
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            code VARCHAR(64) NOT NULL,
            buyer_id VARCHAR(36) NOT NULL,
            recipient_id VARCHAR(36) NOT NULL,
            shipment_courier VARCHAR(191) NOT NULL,
            shipment_service VARCHAR(191),
            shipment_tracking VARCHAR(191),
            shipment_cost DOUBLE NOT NULL DEFAULT 0,
            discount DOUBLE NOT NULL DEFAULT 0,
            total DOUBLE NOT NULL DEFAULT 0,
            profit DOUBLE NOT NULL DEFAULT 0,
            notes TEXT,
            created_at VARCHAR(64) NOT NULL,
            updated_at VARCHAR(64) NOT NULL,
            UNIQUE KEY idx_orders_code (code),
            KEY idx_orders_created_at (created_at),
            CONSTRAINT fk_orders_buyer FOREIGN KEY (buyer_id) REFERENCES customers(id),
            CONSTRAINT fk_orders_recipient FOREIGN KEY (recipient_id) REFERENCES customers(id)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS order_items (
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            order_id VARCHAR(36) NOT NULL,
            product_id VARCHAR(36) NOT NULL,
            quantity INT NOT NULL,
            unit_price DOUBLE NOT NULL,
            discount DOUBLE NOT NULL DEFAULT 0,
            cost_price DOUBLE NOT NULL DEFAULT 0,
            profit DOUBLE NOT NULL DEFAULT 0,
            KEY idx_order_items_order (order_id),
            CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
            CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES products(id)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS stock_mutations (
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            product_id VARCHAR(36) NOT NULL,
            delta INT NOT NULL,
            reason VARCHAR(255) NOT NULL,
            created_at VARCHAR(64) NOT NULL,
            KEY idx_stock_mutations_product (product_id),
            CONSTRAINT fk_stock_mutations_product FOREIGN KEY (product_id) REFERENCES products(id)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS stock_opnames (
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            note TEXT,
            performed_by VARCHAR(191),
            performed_at VARCHAR(64) NOT NULL,
            created_at VARCHAR(64) NOT NULL
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS stock_opname_items (
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            stock_opname_id VARCHAR(36) NOT NULL,
            product_id VARCHAR(36) NOT NULL,
            counted INT NOT NULL,
            previous_stock INT NOT NULL,
            difference INT NOT NULL,
            KEY idx_stock_opname_items_opname (stock_opname_id),
            CONSTRAINT fk_stock_opname_items_opname FOREIGN KEY (stock_opname_id) REFERENCES stock_opnames(id) ON DELETE CASCADE,
            CONSTRAINT fk_stock_opname_items_product FOREIGN KEY (product_id) REFERENCES products(id)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS settings (
			` + "`key`" + ` VARCHAR(191) NOT NULL PRIMARY KEY,
			value LONGTEXT NOT NULL,
			updated_at VARCHAR(64) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS couriers (
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            code VARCHAR(64) NOT NULL,
            name VARCHAR(191) NOT NULL,
            services TEXT,
            tracking_url VARCHAR(255),
            contact VARCHAR(191),
            notes TEXT,
            logo_path VARCHAR(255),
            logo_hash CHAR(64),
            logo_width INT,
            logo_height INT,
            logo_size_bytes BIGINT,
            logo_mime VARCHAR(64),
            created_at VARCHAR(64) NOT NULL,
            updated_at VARCHAR(64) NOT NULL,
            KEY idx_couriers_code (code)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for _, stmt := range statements {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	migrations := []string{
		`ALTER TABLE products ADD COLUMN image_path VARCHAR(255);`,
		`ALTER TABLE products ADD COLUMN thumb_path VARCHAR(255);`,
		`ALTER TABLE products ADD COLUMN image_hash CHAR(64);`,
		`ALTER TABLE products ADD COLUMN image_width INT;`,
		`ALTER TABLE products ADD COLUMN image_height INT;`,
		`ALTER TABLE products ADD COLUMN image_size_bytes BIGINT;`,
		`ALTER TABLE products ADD COLUMN thumb_width INT;`,
		`ALTER TABLE products ADD COLUMN thumb_height INT;`,
		`ALTER TABLE products ADD COLUMN thumb_size_bytes BIGINT;`,
		`ALTER TABLE products ADD COLUMN category VARCHAR(191);`,
		`ALTER TABLE products ADD COLUMN low_stock_threshold INT NOT NULL DEFAULT 5;`,
		`ALTER TABLE products ADD COLUMN deleted_at VARCHAR(64);`,
		`ALTER TABLE couriers ADD COLUMN logo_path VARCHAR(255);`,
		`ALTER TABLE couriers ADD COLUMN logo_hash CHAR(64);`,
		`ALTER TABLE couriers ADD COLUMN logo_width INT;`,
		`ALTER TABLE couriers ADD COLUMN logo_height INT;`,
		`ALTER TABLE couriers ADD COLUMN logo_size_bytes BIGINT;`,
		`ALTER TABLE couriers ADD COLUMN logo_mime VARCHAR(64);`,
		`ALTER TABLE stock_opnames ADD COLUMN performed_by VARCHAR(191);`,
	}

	for _, stmt := range migrations {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			lower := strings.ToLower(err.Error())
			if !strings.Contains(lower, "duplicate column") && !strings.Contains(lower, "exists") {
				return fmt.Errorf("migrate: %w", err)
			}
		}
	}

	return nil
}

// Close releases the underlying database connection.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return errors.New("store not initialised")
	}
	return s.db.Close()
}

func (s *Store) ProductRepository() *repo.ProductRepository {
	if s.productRepo == nil {
		s.productRepo = repo.NewProductRepository(s.db)
	}
	return s.productRepo
}

func (s *Store) CustomerRepository() *repo.CustomerRepository {
	if s.customerRepo == nil {
		s.customerRepo = repo.NewCustomerRepository(s.db)
	}
	return s.customerRepo
}

func (s *Store) OrderRepository() *repo.OrderRepository {
	if s.orderRepo == nil {
		s.orderRepo = repo.NewOrderRepository(s.db)
	}
	return s.orderRepo
}

func (s *Store) SettingsRepository() *repo.SettingsRepository {
	if s.settingsRepo == nil {
		s.settingsRepo = repo.NewSettingsRepository(s.db)
	}
	return s.settingsRepo
}

func (s *Store) CourierRepository() *repo.CourierRepository {
	if s.courierRepo == nil {
		s.courierRepo = repo.NewCourierRepository(s.db)
	}
	return s.courierRepo
}

func (s *Store) StockOpnameRepository() *repo.StockOpnameRepository {
	if s.stockOpnameRepo == nil {
		s.stockOpnameRepo = repo.NewStockOpnameRepository(s.db)
	}
	return s.stockOpnameRepo
}

// DB exposes the raw database connection for advanced use cases.
func (s *Store) DB() *sql.DB {
	return s.db
}

// Config returns a copy of the MySQL configuration used to create the store.
func (s *Store) Config() mysql.Config {
	if s.cfg == nil {
		return mysql.Config{}
	}
	return *s.cfg
}

// RedactedDSN returns the DSN with the password hidden for logging purposes.
func (s *Store) RedactedDSN() string {
	if s.cfg == nil {
		return ""
	}
	clone := *s.cfg
	if clone.Passwd != "" {
		clone.Passwd = "****"
	}
	return clone.FormatDSN()
}
