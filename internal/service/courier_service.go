package service

import (
	"context"
	"errors"
	"strings"

	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/media"
	"smartseller-lite-starter/internal/repo"
)

type CourierService struct {
	repo  *repo.CourierRepository
	media *media.Manager
}

type CourierListOptions struct {
	Query    string `json:"query"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type CourierListResult struct {
	Items    []domain.Courier `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

func NewCourierService(repo *repo.CourierRepository, mediaManager *media.Manager) *CourierService {
	return &CourierService{repo: repo, media: mediaManager}
}

func defaultCouriers() []domain.Courier {
	return []domain.Courier{
		{Code: "JNE", Name: "Jalur Nugraha Ekakurir", Services: "REG · YES", TrackingURL: "https://www.jne.co.id/id/tracking/trace", Contact: "(021) 2927 8888"},
		{Code: "JNT", Name: "J&T Express", Services: "EZ · J&T Super", TrackingURL: "https://www.jet.co.id/track", Contact: "(021) 8066 1888"},
		{Code: "SICEPAT", Name: "SiCepat Ekspres", Services: "REG · BEST", TrackingURL: "https://www.sicepat.com/checkAwb", Contact: "(021) 5020 0050"},
		{Code: "NINJA", Name: "Ninja Xpress", Services: "Standard · Same Day", TrackingURL: "https://www.ninjaxpress.co/id-id/tracking", Contact: "support_id@ninjavan.co"},
		{Code: "ANTERAJA", Name: "AnterAja", Services: "REG · Next Day", TrackingURL: "https://anteraja.id/tracking", Contact: "(021) 5060 3333"},
		{Code: "POS", Name: "Pos Indonesia", Services: "Paket Kilat Khusus · Q9 Sameday", TrackingURL: "https://www.posindonesia.co.id/id/tracking", Contact: "1500161"},
		{Code: "TIKI", Name: "Titipan Kilat", Services: "REG · ONS", TrackingURL: "https://www.tiki.id/id/tracking", Contact: "1500125"},
		{Code: "WAHANA", Name: "Wahana Prestasi Logistik", Services: "Service Next Day · Service Regular", TrackingURL: "https://www.wahana.com/lacak-kiriman", Contact: "(021) 7341688"},
		{Code: "LION", Name: "Lion Parcel", Services: "REGPACK · ONEPACK", TrackingURL: "https://lionparcel.com/tracking", Contact: "(021) 80820072"},
		{Code: "SAP", Name: "SAP Express", Services: "Regular · Sameday", TrackingURL: "https://www.sap-express.id/track", Contact: "(021) 2280 3663"},
	}
}

func (s *CourierService) ensureDefaults(ctx context.Context) {
	_ = s.repo.EnsureDefaults(ctx, defaultCouriers())
}

func (s *CourierService) Warm(ctx context.Context) {
	s.ensureDefaults(ctx)
}

func (s *CourierService) List(ctx context.Context) ([]domain.Courier, error) {
	s.ensureDefaults(ctx)
	result, err := s.repo.ListPaged(ctx, repo.CourierListOptions{Page: 1, PageSize: 0})
	if err != nil {
		return nil, err
	}
	for i := range result.Items {
		s.decorate(&result.Items[i])
	}
	return result.Items, nil
}

func (s *CourierService) ListPaged(ctx context.Context, opts CourierListOptions) (CourierListResult, error) {
	s.ensureDefaults(ctx)
	repoResult, err := s.repo.ListPaged(ctx, repo.CourierListOptions{
		Query:    opts.Query,
		Page:     opts.Page,
		PageSize: opts.PageSize,
	})
	if err != nil {
		return CourierListResult{}, err
	}
	for i := range repoResult.Items {
		s.decorate(&repoResult.Items[i])
	}
	return CourierListResult{
		Items:    repoResult.Items,
		Total:    repoResult.Total,
		Page:     repoResult.Page,
		PageSize: repoResult.PageSize,
	}, nil
}

func (s *CourierService) Save(ctx context.Context, courier domain.Courier) (*domain.Courier, error) {
	courier.Code = strings.ToUpper(strings.TrimSpace(courier.Code))
	courier.Name = strings.TrimSpace(courier.Name)
	courier.TrackingURL = strings.TrimSpace(courier.TrackingURL)
	courier.LogoData = strings.TrimSpace(courier.LogoData)
	courier.LogoMime = strings.TrimSpace(courier.LogoMime)

	var existing *domain.Courier
	if courier.ID != "" {
		var err error
		existing, err = s.repo.Get(ctx, courier.ID)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, err
		}
	}

	if courier.LogoData != "" && s.media != nil {
		asset, err := s.media.SaveLogo(ctx, courier.LogoData)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			_ = s.media.Remove(existing.LogoPath)
		}
		courier.LogoPath = asset.Path
		courier.LogoHash = asset.Hash
		courier.LogoWidth = asset.Width
		courier.LogoHeight = asset.Height
		courier.LogoSizeBytes = asset.SizeBytes
		courier.LogoMime = media.PreferredImageMIME()
	} else if existing != nil {
		if courier.LogoPath == "" && courier.LogoData == "" && existing.LogoPath != "" {
			if s.media != nil {
				_ = s.media.Remove(existing.LogoPath)
			}
			courier.LogoHash = ""
			courier.LogoWidth = 0
			courier.LogoHeight = 0
			courier.LogoSizeBytes = 0
			courier.LogoMime = ""
		} else {
			if courier.LogoPath == "" {
				courier.LogoPath = existing.LogoPath
			}
			if courier.LogoHash == "" {
				courier.LogoHash = existing.LogoHash
			}
			if courier.LogoWidth == 0 {
				courier.LogoWidth = existing.LogoWidth
			}
			if courier.LogoHeight == 0 {
				courier.LogoHeight = existing.LogoHeight
			}
			if courier.LogoSizeBytes == 0 {
				courier.LogoSizeBytes = existing.LogoSizeBytes
			}
			if courier.LogoMime == "" {
				courier.LogoMime = existing.LogoMime
			}
		}
	}
	courier.LogoData = ""

	saved, err := s.repo.Save(ctx, &courier)
	if err != nil {
		return nil, err
	}
	s.decorate(saved)
	return saved, nil
}

func (s *CourierService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("courier id required")
	}
	var existing *domain.Courier
	if s.repo != nil {
		existing, _ = s.repo.Get(ctx, id)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if existing != nil && s.media != nil {
		_ = s.media.Remove(existing.LogoPath)
	}
	return nil
}

func (s *CourierService) ReplaceAll(ctx context.Context, items []domain.Courier) error {
	for i := range items {
		items[i].LogoData = ""
	}
	return s.repo.ReplaceAll(ctx, items)
}

func (s *CourierService) decorate(courier *domain.Courier) {
	if courier == nil {
		return
	}
	courier.LogoData = ""
	if s.media != nil {
		courier.LogoURL = s.media.PublicURL(courier.LogoPath)
	}
	if courier.LogoMime == "" && courier.LogoPath != "" {
		courier.LogoMime = media.PreferredImageMIME()
	}
}
