package out

import (
	"context"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
)

type ContextRepository interface {
	Save(ctx context.Context, c domain.Context) error
	FindAll(ctx context.Context, customer domain.CustomerType) ([]domain.Context, error)
	FindByID(ctx context.Context, id domain.IDType) (domain.Context, error)
	Update(ctx context.Context, c domain.Context) error
	ExecuteQuery(ctx context.Context, query domain.SQLQueryType) ([]domain.Context, error)
}

type StylesheetRepository interface {
	Save(ctx context.Context, s domain.Stylesheet) error
	FindAll(ctx context.Context, name domain.NameType) ([]domain.Stylesheet, error)
	FindByID(ctx context.Context, id domain.IDType) (domain.Stylesheet, error)
	Update(ctx context.Context, s domain.Stylesheet) error
}

type TemplateRepository interface {
	Save(ctx context.Context, t domain.Template) error
	FindAll(ctx context.Context, name domain.NameType) ([]domain.Template, error)
	FindByID(ctx context.Context, id domain.IDType) (domain.Template, error)
	Update(ctx context.Context, t domain.Template) error
}

type BindingRepository interface {
	Save(ctx context.Context, b domain.Binding) error
	FindAll(ctx context.Context, name domain.NameType) ([]domain.Binding, error)
	FindByID(ctx context.Context, id domain.IDType) (domain.Binding, error)
	Update(ctx context.Context, b domain.Binding) error
}
