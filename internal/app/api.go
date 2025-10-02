package app

import (
	"context"
	"encoding/base64"
	"fmt"

	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/service"
)

// API exposes the domain services to the HTTP transport.
type API struct {
	core *Core
}

func NewAPI(core *Core) *API {
	return &API{core: core}
}

func (a *API) ListProducts(ctx context.Context, opts service.ProductListOptions) (service.ProductListResult, error) {
	return a.core.ProductService.ListPaged(ctx, opts)
}

func (a *API) SaveProduct(ctx context.Context, payload domain.Product) (*domain.Product, error) {
	if payload.ID == "" {
		return a.core.ProductService.Create(ctx, payload)
	}
	return a.core.ProductService.Update(ctx, payload)
}

func (a *API) ArchiveProduct(ctx context.Context, id string) error {
	return a.core.ProductService.Archive(ctx, id)
}

func (a *API) DeleteProduct(ctx context.Context, id string) error {
	return a.core.ProductService.Delete(ctx, id)
}

func (a *API) AdjustStock(ctx context.Context, productID string, delta int, reason string) error {
	return a.core.ProductService.AdjustStock(ctx, productID, delta, reason)
}

func (a *API) ListCustomers(ctx context.Context, opts service.CustomerListOptions) (service.CustomerListResult, error) {
	return a.core.CustomerService.ListPaged(ctx, opts)
}

func (a *API) SaveCustomer(ctx context.Context, payload domain.Customer) (*domain.Customer, error) {
	if payload.ID == "" {
		return a.core.CustomerService.Create(ctx, payload)
	}
	return a.core.CustomerService.Update(ctx, payload)
}

func (a *API) DeleteCustomer(ctx context.Context, id string) error {
	return a.core.CustomerService.Delete(ctx, id)
}

func (a *API) ListOrders(ctx context.Context, opts service.OrderListOptions) (service.OrderListResult, error) {
	return a.core.OrderService.ListPaged(ctx, opts)
}

func (a *API) CreateOrder(ctx context.Context, payload service.CreateOrderInput) (*domain.Order, error) {
	return a.core.OrderService.Create(ctx, payload)
}

func (a *API) DeleteOrder(ctx context.Context, id string) error {
	return a.core.OrderService.Delete(ctx, id)
}

func (a *API) GenerateLabel(ctx context.Context, orderID string) (string, error) {
	pdfBytes, err := a.core.OrderService.GenerateLabelPDF(ctx, orderID)
	if err != nil {
		return "", err
	}
	if len(pdfBytes) == 0 {
		return "", fmt.Errorf("empty pdf generated")
	}
	return base64.StdEncoding.EncodeToString(pdfBytes), nil
}

func (a *API) GetSettings(ctx context.Context) (domain.AppSettings, error) {
	return a.core.SettingsService.Get(ctx)
}

func (a *API) UpdateSettings(ctx context.Context, payload domain.AppSettings) (*domain.AppSettings, error) {
	return a.core.SettingsService.Update(ctx, payload)
}

func (a *API) ListCouriers(ctx context.Context, opts service.CourierListOptions) (service.CourierListResult, error) {
	return a.core.CourierService.ListPaged(ctx, opts)
}

func (a *API) SaveCourier(ctx context.Context, payload domain.Courier) (*domain.Courier, error) {
	return a.core.CourierService.Save(ctx, payload)
}

func (a *API) DeleteCourier(ctx context.Context, id string) error {
	return a.core.CourierService.Delete(ctx, id)
}

func (a *API) CreateBackup(ctx context.Context, opts domain.BackupOptions) (string, error) {
	return a.core.BackupService.Create(ctx, opts)
}

func (a *API) RestoreBackup(ctx context.Context, payload string, opts domain.RestoreOptions) (domain.RestoreResult, error) {
	return a.core.BackupService.Restore(ctx, payload, opts)
}

func (a *API) ExportOrdersCSV(ctx context.Context, filters service.OrderExportFilters) ([]byte, error) {
	if a.core.ReportService == nil {
		return nil, fmt.Errorf("report service unavailable")
	}
	return a.core.ReportService.ExportOrdersCSV(ctx, filters)
}

func (a *API) ListStockOpnames(ctx context.Context, limit int) ([]domain.StockOpname, error) {
	if a.core.StockOpnameService == nil {
		return []domain.StockOpname{}, nil
	}
	return a.core.StockOpnameService.List(ctx, limit)
}

func (a *API) PerformStockOpname(ctx context.Context, payload service.PerformStockOpnameInput) (*domain.StockOpname, error) {
	if a.core.StockOpnameService == nil {
		return nil, fmt.Errorf("stock opname service unavailable")
	}
	return a.core.StockOpnameService.Perform(ctx, payload)
}
