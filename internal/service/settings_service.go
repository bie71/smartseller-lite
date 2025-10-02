package service

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/media"
	"smartseller-lite-starter/internal/repo"
)

type SettingsService struct {
	repo         *repo.SettingsRepository
	defaultBrand string
	media        *media.Manager
}

func NewSettingsService(repo *repo.SettingsRepository, defaultBrand string, mediaManager *media.Manager) *SettingsService {
	return &SettingsService{repo: repo, defaultBrand: defaultBrand, media: mediaManager}
}

func (s *SettingsService) Warm(ctx context.Context) {
	_ = s.repo.EnsureDefaults(ctx, s.defaultBrand)
}

func (s *SettingsService) Get(ctx context.Context) (domain.AppSettings, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return domain.AppSettings{BrandName: s.defaultBrand}, nil
	}
	if settings.BrandName == "" {
		settings.BrandName = s.defaultBrand
	}
	s.decorate(settings)
	return *settings, nil
}

func (s *SettingsService) Update(ctx context.Context, payload domain.AppSettings) (*domain.AppSettings, error) {
	if payload.BrandName == "" {
		payload.BrandName = s.defaultBrand
	}
	current, _ := s.repo.Get(ctx)
	payload.LogoData = strings.TrimSpace(payload.LogoData)
	if payload.LogoData != "" && s.media != nil {
		asset, err := s.media.SaveLogo(ctx, payload.LogoData)
		if err != nil {
			return nil, err
		}
		if current != nil {
			_ = s.media.Remove(current.LogoPath)
		}
		payload.LogoPath = asset.Path
		payload.LogoHash = asset.Hash
		payload.LogoWidth = asset.Width
		payload.LogoHeight = asset.Height
		payload.LogoSizeBytes = asset.SizeBytes
		payload.LogoMime = media.PreferredImageMIME()
	} else if current != nil {
		if payload.LogoPath == "" && payload.LogoData == "" && current.LogoPath != "" {
			if s.media != nil {
				_ = s.media.Remove(current.LogoPath)
			}
			payload.LogoHash = ""
			payload.LogoWidth = 0
			payload.LogoHeight = 0
			payload.LogoSizeBytes = 0
			payload.LogoMime = ""
		} else {
			if payload.LogoPath == "" {
				payload.LogoPath = current.LogoPath
			}
			if payload.LogoHash == "" {
				payload.LogoHash = current.LogoHash
			}
			if payload.LogoWidth == 0 {
				payload.LogoWidth = current.LogoWidth
			}
			if payload.LogoHeight == 0 {
				payload.LogoHeight = current.LogoHeight
			}
			if payload.LogoSizeBytes == 0 {
				payload.LogoSizeBytes = current.LogoSizeBytes
			}
			if payload.LogoMime == "" {
				payload.LogoMime = current.LogoMime
			}
		}
	}
	payload.LogoData = ""
	saved, err := s.repo.Update(ctx, payload)
	if err != nil {
		return nil, err
	}
	s.decorate(saved)
	return saved, nil
}

func (s *SettingsService) ReplaceAll(ctx context.Context, payload domain.AppSettings) error {
	if payload.BrandName == "" {
		payload.BrandName = s.defaultBrand
	}
	payload.LogoData = ""
	return s.repo.ReplaceAll(ctx, payload)
}

func (s *SettingsService) decorate(settings *domain.AppSettings) {
	if settings == nil {
		return
	}
	if s.media != nil {
		settings.LogoURL = s.media.PublicURL(settings.LogoPath)
	}
	if settings.LogoMime == "" && settings.LogoPath != "" {
		settings.LogoMime = media.PreferredImageMIME()
	}
}

// LoadLogoBytes returns the stored logo binary and mime type for the provided settings snapshot.
func (s *SettingsService) LoadLogoBytes(ctx context.Context, settings domain.AppSettings) ([]byte, string, error) {
	_ = ctx
	if strings.TrimSpace(settings.LogoData) != "" {
		data, err := base64.StdEncoding.DecodeString(settings.LogoData)
		if err != nil {
			return nil, "", err
		}
		mime := strings.TrimSpace(settings.LogoMime)
		if mime == "" {
			mime = http.DetectContentType(data)
		}
		return data, mime, nil
	}
	if s.media == nil {
		return nil, "", nil
	}
	if strings.TrimSpace(settings.LogoPath) == "" {
		return nil, "", nil
	}
	data, err := s.media.Read(settings.LogoPath)
	if err != nil {
		return nil, "", err
	}
	mime := strings.TrimSpace(settings.LogoMime)
	if mime == "" {
		mime = http.DetectContentType(data)
	}
	return data, mime, nil
}
