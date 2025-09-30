package app

import (
	"context"

	"smartseller-lite-starter/internal/db"
	"smartseller-lite-starter/internal/media"
	"smartseller-lite-starter/internal/service"
)

// CoreConfig defines runtime configuration defaults for the app domain services.
type CoreConfig struct {
	DefaultBrandName string
	MediaManager     *media.Manager
}

// Core wires repositories, services, and lifecycle events together.
type Core struct {
	store              *db.Store
	CustomerService    *service.CustomerService
	OrderService       *service.OrderService
	ProductService     *service.ProductService
	SettingsService    *service.SettingsService
	CourierService     *service.CourierService
	BackupService      *service.BackupService
	StockOpnameService *service.StockOpnameService
	ReportService      *service.ReportService
}

func NewCore(store *db.Store, cfg CoreConfig) *Core {
	productRepo := store.ProductRepository()
	customerRepo := store.CustomerRepository()
	orderRepo := store.OrderRepository()
	settingsRepo := store.SettingsRepository()
	courierRepo := store.CourierRepository()
	stockOpnameRepo := store.StockOpnameRepository()

	productSvc := service.NewProductService(productRepo, cfg.MediaManager)
	customerSvc := service.NewCustomerService(customerRepo)
	settingsSvc := service.NewSettingsService(settingsRepo, cfg.DefaultBrandName, cfg.MediaManager)
	courierSvc := service.NewCourierService(courierRepo, cfg.MediaManager)
	orderSvc := service.NewOrderService(orderRepo, productSvc, customerSvc, settingsSvc)
	stockOpnameSvc := service.NewStockOpnameService(stockOpnameRepo, productSvc)
	backupSvc := service.NewBackupService(store, cfg.MediaManager)
	reportSvc := service.NewReportService(store)

	return &Core{
		store:              store,
		CustomerService:    customerSvc,
		OrderService:       orderSvc,
		ProductService:     productSvc,
		SettingsService:    settingsSvc,
		CourierService:     courierSvc,
		BackupService:      backupSvc,
		StockOpnameService: stockOpnameSvc,
		ReportService:      reportSvc,
	}
}

// Warm primes repositories and caches so the first user interaction feels responsive.
func (c *Core) Warm(ctx context.Context) {
	c.ProductService.Warm(ctx)
	c.CustomerService.Warm(ctx)
	c.OrderService.Warm(ctx)
	c.SettingsService.Warm(ctx)
	c.CourierService.Warm(ctx)
	if c.StockOpnameService != nil {
		c.StockOpnameService.Warm(ctx)
	}
}

// Close releases the underlying store connection.
func (c *Core) Close() error {
	if c == nil {
		return nil
	}
	return c.store.Close()
}
