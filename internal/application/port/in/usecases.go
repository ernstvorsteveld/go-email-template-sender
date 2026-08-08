package in

import (
	"context"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
)

type ContextUseCase interface {
	CreateContext(ctx context.Context, reference domain.ReferenceType, customer domain.CustomerType, payload domain.JSONPayloadType, emailAddress domain.JSONPathType) (domain.IDType, error)
	GetContexts(ctx context.Context, customer domain.CustomerType) ([]domain.Context, error)
	GetContext(ctx context.Context, id domain.IDType) (domain.Context, error)
	UpdateContext(ctx context.Context, id domain.IDType, reference domain.ReferenceType, customer domain.CustomerType, payload domain.JSONPayloadType, emailAddress domain.JSONPathType) error
}

type StylesheetUseCase interface {
	CreateStylesheet(ctx context.Context, name domain.NameType, code domain.CodeType, content domain.CSSType) (domain.IDType, error)
	GetStylesheets(ctx context.Context, name domain.NameType) ([]domain.Stylesheet, error)
	GetStylesheet(ctx context.Context, id domain.IDType) (domain.Stylesheet, error)
	UpdateStylesheet(ctx context.Context, id domain.IDType, name domain.NameType, code domain.CodeType, content domain.CSSType) error
}

type TemplateUseCase interface {
	CreateTemplate(ctx context.Context, name domain.NameType, code domain.CodeType, content domain.HTMLType, stylesheet *domain.IDType, subject domain.SubjectType) (domain.IDType, error)
	GetTemplates(ctx context.Context, name domain.NameType) ([]domain.Template, error)
	GetTemplate(ctx context.Context, id domain.IDType) (domain.Template, error)
	UpdateTemplate(ctx context.Context, id domain.IDType, name domain.NameType, code domain.CodeType, content domain.HTMLType, stylesheet *domain.IDType, subject domain.SubjectType) error
	RenderTemplate(ctx context.Context, id domain.IDType) (domain.HTMLType, error)
}

type BindingUseCase interface {
	CreateBinding(ctx context.Context, name domain.NameType, query domain.SQLQueryType, template domain.IDType) (domain.IDType, error)
	GetBindings(ctx context.Context, name domain.NameType) ([]domain.Binding, error)
	GetBinding(ctx context.Context, id domain.IDType) (domain.Binding, error)
	UpdateBinding(ctx context.Context, id domain.IDType, name domain.NameType, query domain.SQLQueryType, template domain.IDType) error
}

type DeliveryUseCase interface {
	Dispatch(ctx context.Context, template domain.IDType, binding domain.IDType) error
}
