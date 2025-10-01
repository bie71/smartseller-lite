package service

import (
	"context"
	"errors"
	"strings"

	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/media"
	"smartseller-lite-starter/internal/repo"
)

// ProductService encapsulates business rules for products and stock.
type ProductService struct {
    repo  *repo.ProductRepository
    media *media.Manager
}

type ProductListOptions struct {
    Query    string `json:"query"`
    Page     int    `json:"page"`
    PageSize int    `json:"pageSize"`
}

type ProductListResult struct {
    Items              []domain.Product `json:"items"`
    Total              int              `json:"total"`
    Page               int              `json:"page"`
    PageSize           int              `json:"pageSize"`
    OutOfStockCount    int              `json:"outOfStockCount"`
    WarningStockCount  int              `json:"warningStockCount"`
    LowStockHighlights []domain.Product `json:"lowStockHighlights"`
}

func NewProductService(repo *repo.ProductRepository, mediaManager *media.Manager) *ProductService {
    return &ProductService{repo: repo, media: mediaManager}
}

func (s *ProductService) Warm(ctx context.Context) {
	_, _ = s.repo.List(ctx)
}

func (s *ProductService) upsert(ctx context.Context, existing *domain.Product, p domain.Product) (*domain.Product, error) {
	if strings.TrimSpace(p.Name) == "" {
		return nil, errors.New("product name is required")
	}
	if strings.TrimSpace(p.SKU) == "" {
		return nil, errors.New("product SKU is required")
	}
	p.Category = strings.TrimSpace(p.Category)
	if p.LowStockThreshold <= 0 {
		p.LowStockThreshold = 5
	}
	p.ImageData = strings.TrimSpace(p.ImageData)
	if p.ImageData != "" && s.media != nil {
		asset, err := s.media.SaveProductImage(ctx, p.ImageData)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			_ = s.media.Remove(existing.ImagePath, existing.ThumbPath)
		}
		p.ImagePath = asset.Path
		p.ThumbPath = asset.ThumbPath
		p.ImageHash = asset.Hash
		p.ImageWidth = asset.Width
		p.ImageHeight = asset.Height
		p.ImageSizeBytes = asset.SizeBytes
		p.ThumbWidth = asset.ThumbWidth
		p.ThumbHeight = asset.ThumbHeight
		p.ThumbSizeBytes = asset.ThumbSizeBytes
	} else if existing != nil {
		p.ImagePath = existing.ImagePath
		p.ThumbPath = existing.ThumbPath
		p.ImageHash = existing.ImageHash
		p.ImageWidth = existing.ImageWidth
		p.ImageHeight = existing.ImageHeight
		p.ImageSizeBytes = existing.ImageSizeBytes
		p.ThumbWidth = existing.ThumbWidth
		p.ThumbHeight = existing.ThumbHeight
		p.ThumbSizeBytes = existing.ThumbSizeBytes
	}
	p.ImageData = ""
	if p.ID == "" {
		return s.repo.Create(ctx, &p)
	}
	return s.repo.Update(ctx, &p)
}

func (s *ProductService) Create(ctx context.Context, p domain.Product) (*domain.Product, error) {
	p.ID = ""
	created, err := s.upsert(ctx, nil, p)
	if err != nil {
		return nil, err
	}
	s.decorate(created)
	return created, nil
}

func (s *ProductService) Update(ctx context.Context, p domain.Product) (*domain.Product, error) {
	if p.ID == "" {
		return nil, errors.New("product ID is required for update")
	}
	var existing *domain.Product
	if s.repo != nil {
		var err error
		existing, err = s.repo.Get(ctx, p.ID)
		if err != nil {
			return nil, err
		}
	}
	updated, err := s.upsert(ctx, existing, p)
	if err != nil {
		return nil, err
	}
	s.decorate(updated)
	return updated, nil
}

func (s *ProductService) AdjustStock(ctx context.Context, productID string, delta int, reason string) error {
	if productID == "" {
		return errors.New("productID required")
	}
	if delta == 0 {
		return nil
	}
	if reason == "" {
		reason = "manual"
	}
	return s.repo.AdjustStock(ctx, productID, delta, reason)
}

func (s *ProductService) List(ctx context.Context) ([]domain.Product, error) {
    result, err := s.ListPaged(ctx, ProductListOptions{Page: 1, PageSize: 0})
    if err != nil {
        return nil, err
    }
    return result.Items, nil
}

func (s *ProductService) ListIncludingArchived(ctx context.Context) ([]domain.Product, error) {
    items, err := s.repo.ListIncludingArchived(ctx)
    if err != nil {
        return nil, err
	}
	for i := range items {
		s.decorate(&items[i])
	}
	return items, nil
}

func (s *ProductService) ListPaged(ctx context.Context, opts ProductListOptions) (ProductListResult, error) {
    repoResult, err := s.repo.ListPaged(ctx, repo.ProductListOptions{
        Query:           opts.Query,
        Page:            opts.Page,
        PageSize:        opts.PageSize,
        IncludeArchived: false,
    })
    if err != nil {
        return ProductListResult{}, err
    }

    for i := range repoResult.Items {
        s.decorate(&repoResult.Items[i])
    }
    for i := range repoResult.LowStockHighlights {
        s.decorate(&repoResult.LowStockHighlights[i])
    }

    return ProductListResult{
        Items:              repoResult.Items,
        Total:              repoResult.Total,
        Page:               repoResult.Page,
        PageSize:           repoResult.PageSize,
        OutOfStockCount:    repoResult.OutOfStockCount,
        WarningStockCount:  repoResult.WarningStockCount,
        LowStockHighlights: repoResult.LowStockHighlights,
    }, nil
}

func (s *ProductService) Get(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, errors.New("product id required")
	}
	product, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	s.decorate(product)
	return product, nil
}

func (s *ProductService) Archive(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("product id required")
	}
	return s.repo.Archive(ctx, id)
}

func (s *ProductService) ReplaceAll(ctx context.Context, items []domain.Product) error {
	for i := range items {
		items[i].ImageData = ""
	}
	return s.repo.ReplaceAll(ctx, items)
}

func (s *ProductService) decorate(product *domain.Product) {
	if product == nil {
		return
	}
	product.ImageData = ""
	if s.media != nil {
		product.ImageURL = s.media.PublicURL(product.ImagePath)
		product.ThumbURL = s.media.PublicURL(product.ThumbPath)
	}
}
