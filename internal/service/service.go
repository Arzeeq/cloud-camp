package service

import (
	"context"
	"time"
)

type TokenStorager interface {
	GetCapacity(ctx context.Context, token string) (int, error)
	SetCapacity(ctx context.Context, token string, capacity int) error
}

type TokenService struct {
	storage TokenStorager
}

func NewTokenService(storage TokenStorager) *TokenService {
	return &TokenService{storage: storage}
}

func (s *TokenService) GetCapacity(token string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.storage.GetCapacity(ctx, token)
}

func (s *TokenService) SetCapacity(token string, capacity int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.storage.SetCapacity(ctx, token, capacity)
}
