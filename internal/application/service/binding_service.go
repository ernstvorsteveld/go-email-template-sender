package service

import (
	"context"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type bindingService struct {
	repo out.BindingRepository
}

func NewBindingService(repo out.BindingRepository) in.BindingUseCase {
	return &bindingService{repo: repo}
}

func (s *bindingService) CreateBinding(ctx context.Context, name domain.NameType, query domain.SQLQueryType, template domain.IDType) (domain.IDType, error) {
	b := domain.Binding{
		ID:       domain.IDType(uuid.New()),
		Name:     name,
		Query:    query,
		Template: template,
	}
	if err := s.repo.Save(ctx, b); err != nil {
		return domain.IDType(uuid.Nil), err
	}
	return b.ID, nil
}

func (s *bindingService) GetBindings(ctx context.Context, name domain.NameType) ([]domain.Binding, error) {
	return s.repo.FindAll(ctx, name)
}

func (s *bindingService) GetBinding(ctx context.Context, id domain.IDType) (domain.Binding, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *bindingService) UpdateBinding(ctx context.Context, id domain.IDType, name domain.NameType, query domain.SQLQueryType, template domain.IDType) error {
	b := domain.Binding{
		ID:       id,
		Name:     name,
		Query:    query,
		Template: template,
	}
	return s.repo.Update(ctx, b)
}
