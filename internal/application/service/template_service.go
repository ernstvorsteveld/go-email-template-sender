package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type templateService struct {
	templateRepo   out.TemplateRepository
	stylesheetRepo out.StylesheetRepository
}

// NewTemplateService creates a new Template application service.
func NewTemplateService(templateRepo out.TemplateRepository, stylesheetRepo out.StylesheetRepository) in.TemplateUseCase {
	return &templateService{
		templateRepo:   templateRepo,
		stylesheetRepo: stylesheetRepo,
	}
}

func (s *templateService) CreateTemplate(ctx context.Context, name domain.NameType, code domain.CodeType, content domain.HTMLType, stylesheet *domain.IDType, subject domain.SubjectType) (domain.IDType, error) {
	newTemplate := domain.Template{
		ID:         domain.IDType(uuid.New()),
		Name:       name,
		Code:       code,
		Version:    domain.VersionType(1),
		Stylesheet: stylesheet,
		Content:    content,
		Subject:    subject,
	}
	if err := s.templateRepo.Save(ctx, newTemplate); err != nil {
		return domain.IDType(uuid.Nil), err
	}
	return newTemplate.ID, nil
}

func (s *templateService) GetTemplates(ctx context.Context, name domain.NameType) ([]domain.Template, error) {
	return s.templateRepo.FindAll(ctx, name)
}

func (s *templateService) GetTemplate(ctx context.Context, id domain.IDType) (domain.Template, error) {
	return s.templateRepo.FindByID(ctx, id)
}

func (s *templateService) UpdateTemplate(ctx context.Context, id domain.IDType, name domain.NameType, code domain.CodeType, content domain.HTMLType, stylesheet *domain.IDType, subject domain.SubjectType) error {
	tmpl, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	tmpl.Name = name
	tmpl.Code = code
	tmpl.Content = content
	tmpl.Stylesheet = stylesheet
	tmpl.Subject = subject
	tmpl.Version++ // Increment version on update

	return s.templateRepo.Update(ctx, tmpl)
}

func (s *templateService) RenderTemplate(ctx context.Context, id domain.IDType) (domain.HTMLType, error) {
	tmpl, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	htmlContent := string(tmpl.Content)

	if tmpl.Stylesheet != nil {
		style, err := s.stylesheetRepo.FindByID(ctx, *tmpl.Stylesheet)
		if err != nil {
			return "", fmt.Errorf("failed to load linked stylesheet: %w", err)
		}

		// Use goquery to parse the HTML string
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			return "", fmt.Errorf("failed to parse HTML with goquery: %w", err)
		}

		// Safely append a <style> tag containing the CSS into the <head> section
		// We use an inline <style> rather than <link> to ensure email clients render it correctly,
		// though a <link> could also be injected here if the stylesheet is hosted publicly.
		styleTag := fmt.Sprintf("<style>\n%s\n</style>", style.Content)
		doc.Find("head").AppendHtml(styleTag)

		var buf bytes.Buffer
		if err := goquery.Render(&buf, doc.Selection); err != nil {
			return "", fmt.Errorf("failed to render HTML with goquery: %w", err)
		}
		htmlContent = buf.String()
	}

	return domain.HTMLType(htmlContent), nil
}
