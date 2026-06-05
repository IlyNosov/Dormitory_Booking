package room

import (
	"context"
	"errors"
	"fmt"

	roomdomain "Dormitory_Booking/internal/domain/room"
)

type Service struct {
	repo roomdomain.Repository
}

func NewService(repo roomdomain.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]roomdomain.Room, error) {
	return s.repo.List(ctx)
}

func (s *Service) ActiveRooms(ctx context.Context) ([]roomdomain.Room, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var out []roomdomain.Room
	for _, r := range all {
		if r.IsActive {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *Service) IsActive(ctx context.Context, number int) (bool, error) {
	r, err := s.repo.Get(ctx, number)
	if err != nil {
		return false, err
	}
	return r.IsActive, nil
}

func (s *Service) GetSchedule(ctx context.Context, number int) (roomdomain.Room, error) {
	return s.repo.Get(ctx, number)
}

func (s *Service) Create(ctx context.Context, r roomdomain.Room) (roomdomain.Room, error) {
	if r.Number <= 0 || r.Number > 9999 {
		return roomdomain.Room{}, errors.New("номер комнаты должен быть от 1 до 9999")
	}
	if r.Name == "" {
		r.Name = fmt.Sprintf("Комната %d", r.Number)
	}
	if r.WeekdayOpen == 0 && r.WeekdayClose == 0 {
		r.WeekdayOpen, r.WeekdayClose = 6, 23
		r.FriSatOpen, r.FriSatClose = 6, 25
		r.SunOpen, r.SunClose = 6, 23
	}
	r.IsActive = true
	return s.repo.Create(ctx, r)
}

func (s *Service) Delete(ctx context.Context, number int) error {
	return s.repo.Delete(ctx, number)
}
