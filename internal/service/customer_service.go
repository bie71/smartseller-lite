package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/repo"
)

type CustomerService struct {
	repo *repo.CustomerRepository
}

func NewCustomerService(repo *repo.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) Warm(ctx context.Context) {
	_, _ = s.repo.List(ctx)
}

func (s *CustomerService) Create(ctx context.Context, c domain.Customer) (*domain.Customer, error) {
	c.Name = strings.TrimSpace(c.Name)
	c.Email = strings.TrimSpace(c.Email)
	c.Address = strings.TrimSpace(c.Address)
	c.City = strings.TrimSpace(c.City)
	c.Province = strings.TrimSpace(c.Province)
	c.Postal = strings.TrimSpace(c.Postal)
	c.Notes = strings.TrimSpace(c.Notes)
	c.Type = domain.CustomerType(strings.TrimSpace(strings.ToLower(string(c.Type))))
	normalised, err := normalisePhone(c.Phone)
	if err != nil {
		return nil, err
	}
	c.Phone = normalised
	if err := validateCustomer(c); err != nil {
		return nil, err
	}
	if c.Type == "" {
		c.Type = domain.CustomerTypeCustomer
	}
	c.ID = ""
	return s.repo.Create(ctx, &c)
}

func (s *CustomerService) Update(ctx context.Context, c domain.Customer) (*domain.Customer, error) {
	if c.ID == "" {
		return nil, errors.New("customer id required")
	}
	c.Name = strings.TrimSpace(c.Name)
	c.Email = strings.TrimSpace(c.Email)
	c.Address = strings.TrimSpace(c.Address)
	c.City = strings.TrimSpace(c.City)
	c.Province = strings.TrimSpace(c.Province)
	c.Postal = strings.TrimSpace(c.Postal)
	c.Notes = strings.TrimSpace(c.Notes)
	c.Type = domain.CustomerType(strings.TrimSpace(strings.ToLower(string(c.Type))))
	normalised, err := normalisePhone(c.Phone)
	if err != nil {
		return nil, err
	}
	c.Phone = normalised
	if err := validateCustomer(c); err != nil {
		return nil, err
	}
	if c.Type == "" {
		c.Type = domain.CustomerTypeCustomer
	}
	return s.repo.Update(ctx, &c)
}

func (s *CustomerService) List(ctx context.Context) ([]domain.Customer, error) {
	return s.repo.List(ctx)
}

func (s *CustomerService) Get(ctx context.Context, id string) (*domain.Customer, error) {
	if id == "" {
		return nil, errors.New("customer id required")
	}
	return s.repo.Get(ctx, id)
}

func (s *CustomerService) ReplaceAll(ctx context.Context, items []domain.Customer) error {
	return s.repo.ReplaceAll(ctx, items)
}

func validateCustomer(c domain.Customer) error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("customer name is required")
	}
	if c.Type != "" && c.Type != domain.CustomerTypeCustomer && c.Type != domain.CustomerTypeMarketer && c.Type != domain.CustomerTypeReseller {
		return errors.New("invalid customer type")
	}
	if strings.TrimSpace(c.Phone) != "" && !strings.HasPrefix(c.Phone, "+") {
		return errors.New("phone number must be stored in international format")
	}
	return nil
}

func normalisePhone(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	hasPlus := strings.HasPrefix(trimmed, "+")
	digits := strings.Builder{}
	for _, r := range trimmed {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	cleaned := digits.String()
	if cleaned == "" {
		return "", fmt.Errorf("nomor HP tidak valid")
	}
	if len(cleaned) < 8 || len(cleaned) > 15 {
		return "", fmt.Errorf("nomor HP harus 8-15 digit")
	}
	if hasPlus {
		return "+" + cleaned, nil
	}
	if strings.HasPrefix(cleaned, "0") {
		rest := strings.TrimLeft(cleaned[1:], "0")
		if rest == "" {
			rest = "0"
		}
		return "+62" + rest, nil
	}
	return "+" + cleaned, nil
}
