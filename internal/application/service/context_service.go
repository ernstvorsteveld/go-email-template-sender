package service

import (
	"context"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type contextService struct {
	repo out.ContextRepository
}

// NewContextService creates a new Context application service.
func NewContextService(repo out.ContextRepository) in.ContextUseCase {
	return &contextService{repo: repo}
}

func (s *contextService) CreateContext(ctx context.Context, reference domain.ReferenceType, customer domain.CustomerType, payload domain.JSONPayloadType, emailAddress domain.JSONPathType) (domain.IDType, error) {
	newContext := domain.Context{
		ID:           domain.IDType(uuid.New()),
		Reference:    reference,
		Customer:     customer,
		Payload:      payload,
		EmailAddress: emailAddress,
	}

	if err := s.repo.Save(ctx, newContext); err != nil {
		return domain.IDType(uuid.Nil), err
	}
	return newContext.ID, nil
}

func (s *contextService) GetContexts(ctx context.Context, customer domain.CustomerType) ([]domain.Context, error) {
	return s.repo.FindAll(ctx, customer)
}

func (s *contextService) GetContext(ctx context.Context, id domain.IDType) (domain.Context, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *contextService) UpdateContext(ctx context.Context, id domain.IDType, reference domain.ReferenceType, customer domain.CustomerType, payload domain.JSONPayloadType, emailAddress domain.JSONPathType) error {
	updatedContext := domain.Context{
		ID:           id,
		Reference:    reference,
		Customer:     customer,
		Payload:      payload,
		EmailAddress: emailAddress,
	}
	return s.repo.Update(ctx, updatedContext)
}
