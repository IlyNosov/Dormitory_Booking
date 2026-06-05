package admin

import (
	"context"
	"errors"
	"strings"

	admindomain "Dormitory_Booking/internal/domain/admin"
	"Dormitory_Booking/internal/application/auth"
	"Dormitory_Booking/internal/infrastructure/postgres"
)

type Service struct {
	repo *postgres.AdminRepo
}

func NewService(repo *postgres.AdminRepo) *Service { return &Service{repo: repo} }

type AdminDTO struct {
	Email     string `json:"email"`
	AddedBy   string `json:"addedBy"`
	CreatedAt string `json:"createdAt"`
}

func (s *Service) List(ctx context.Context) ([]admindomain.Admin, error) {
	entries, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	// преобразуем в доменный тип
	var out []admindomain.Admin
	for _, e := range entries {
		if admindomain.IsSuperAdmin(e.Email) {
			continue
		}
		out = append(out, admindomain.Admin{
			Email:     e.Email,
			AddedBy:   e.AddedBy,
			CreatedAt: e.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) IsAdmin(ctx context.Context, email string) (bool, error) {
	return s.repo.IsAdmin(ctx, email)
}

func (s *Service) Add(ctx context.Context, email, byEmail string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := auth.ValidateHSEEmail(email); err != nil {
		return err
	}
	if admindomain.IsSuperAdmin(email) {
		return errors.New("нельзя добавить суперадмина повторно")
	}
	return s.repo.Add(ctx, email, byEmail)
}

func (s *Service) Remove(ctx context.Context, email, byEmail string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if admindomain.IsSuperAdmin(email) {
		return errors.New("нельзя удалить суперадмина")
	}
	return s.repo.Remove(ctx, email)
}
