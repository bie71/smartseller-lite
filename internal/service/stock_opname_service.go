package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/repo"
)

// PerformStockOpnameItem describes a counted product during stock take.
type PerformStockOpnameItem struct {
	ProductID string `json:"productId"`
	Counted   int    `json:"counted"`
}

// PerformStockOpnameInput is the payload accepted by the stock opname workflow.
type PerformStockOpnameInput struct {
	Note  string                   `json:"note"`
	User  string                   `json:"user"`
	Items []PerformStockOpnameItem `json:"items"`
}

// StockOpnameService coordinates stock taking sessions and resulting adjustments.
type StockOpnameService struct {
	repo     *repo.StockOpnameRepository
	products *ProductService
}

func NewStockOpnameService(repo *repo.StockOpnameRepository, products *ProductService) *StockOpnameService {
	return &StockOpnameService{repo: repo, products: products}
}

func (s *StockOpnameService) Warm(ctx context.Context) {
	_, _ = s.repo.List(ctx, 5)
}

func (s *StockOpnameService) Perform(ctx context.Context, input PerformStockOpnameInput) (*domain.StockOpname, error) {
	if len(input.Items) == 0 {
		return nil, fmt.Errorf("stock opname requires at least one item")
	}
	actor := strings.TrimSpace(input.User)
	if actor == "" {
		return nil, fmt.Errorf("nama petugas opname wajib diisi")
	}
	opname := &domain.StockOpname{ID: uuid.New().String(), Note: strings.TrimSpace(input.Note), PerformedBy: actor}
	opname.PerformedAt = time.Now().UTC()

	for _, entry := range input.Items {
		if entry.ProductID == "" {
			return nil, fmt.Errorf("product id is required")
		}
		if entry.Counted < 0 {
			return nil, fmt.Errorf("counted stock cannot be negative")
		}
		product, err := s.products.Get(ctx, entry.ProductID)
		if err != nil {
			return nil, err
		}
		diff := entry.Counted - product.Stock
		opname.Items = append(opname.Items, domain.StockOpnameItem{
			ProductID:     product.ID,
			ProductName:   product.Name,
			ProductSKU:    product.SKU,
			Counted:       entry.Counted,
			PreviousStock: product.Stock,
			Difference:    diff,
		})
	}

	saved, err := s.repo.Create(ctx, opname)
	if err != nil {
		return nil, err
	}

	for _, item := range saved.Items {
		if item.Difference == 0 {
			continue
		}
		reason := fmt.Sprintf("stock_opname:%s", saved.ID)
		if err := s.products.AdjustStock(ctx, item.ProductID, item.Difference, reason); err != nil {
			return nil, err
		}
	}

	// Refresh item details after adjustments to reflect latest stock in the response.
	for idx, item := range saved.Items {
		product, err := s.products.Get(ctx, item.ProductID)
		if err == nil {
			saved.Items[idx].ProductName = product.Name
			saved.Items[idx].ProductSKU = product.SKU
		}
	}

	return saved, nil
}

func (s *StockOpnameService) List(ctx context.Context, limit int) ([]domain.StockOpname, error) {
	return s.repo.List(ctx, limit)
}

func (s *StockOpnameService) ListAll(ctx context.Context) ([]domain.StockOpname, error) {
	return s.repo.ListAll(ctx)
}

func (s *StockOpnameService) ReplaceAll(ctx context.Context, opnames []domain.StockOpname) error {
	return s.repo.ReplaceAll(ctx, opnames)
}
