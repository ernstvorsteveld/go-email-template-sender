package service

import (
	"context"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type stylesheetService struct {
	repo out.StylesheetRepository
}

func NewStylesheetService(repo out.StylesheetRepository) in.StylesheetUseCase {
	return &stylesheetService{repo: repo}
}

func (s *stylesheetService) CreateStylesheet(ctx context.Context, name domain.NameType, code domain.CodeType, content domain.CSSType) (domain.IDType, error) {
	newStyle := domain.Stylesheet{
		ID:      domain.IDType(uuid.New()),
		Name:    name,
		Code:    code,
		Content: content,
	}
	if err := s.repo.Save(ctx, newStyle); err != nil {
		return domain.IDType(uuid.Nil), err
	}
	return newStyle.ID, nil
}

func (s *stylesheetService) GetStylesheets(ctx context.Context, name domain.NameType) ([]domain.Stylesheet, error) {
	return s.repo.FindAll(ctx, name)
}

func (s *stylesheetService) GetStylesheet(ctx context.Context, id domain.IDType) (domain.Stylesheet, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *stylesheetService) UpdateStylesheet(ctx context.Context, id domain.IDType, name domain.NameType, code domain.CodeType, content domain.CSSType) error {
	updated := domain.Stylesheet{
		ID:      id,
		Name:    name,
		Code:    code,
		Content: content,
	}
	return s.repo.Update(ctx, updated)
}
