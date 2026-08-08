package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type templateRepository struct {
	db *pgxpool.Pool
}

func NewTemplateRepository(db *pgxpool.Pool) out.TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Save(ctx context.Context, t domain.Template) error {
	var sID *string
	if t.Stylesheet != nil {
		val := uuid.UUID(*t.Stylesheet).String()
		sID = &val
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO templates (id, name, code, version, stylesheet_id, html_content, subject) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.UUID(t.ID).String(), string(t.Name), string(t.Code), int(t.Version), sID, string(t.Content), string(t.Subject),
	)
	if err != nil {
		if strings.Contains(err.Error(), "23503") || strings.Contains(err.Error(), "templates_stylesheet_id_fkey") {
			return errors.New("invalid stylesheet_id: the referenced stylesheet does not exist")
		}
		return err
	}
	return nil
}

func (r *templateRepository) FindAll(ctx context.Context, name domain.NameType) ([]domain.Template, error) {
	query := `SELECT id, name, code, version, stylesheet_id, html_content, subject FROM templates`
	var args []interface{}
	if name != "" {
		query += ` WHERE name ILIKE $1`
		args = append(args, "%"+string(name)+"%")
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Template
	for rows.Next() {
		var id, tName, code, content, subject string
		var version int
		var sID *string

		if err := rows.Scan(&id, &tName, &code, &version, &sID, &content, &subject); err != nil {
			return nil, err
		}

		uid, _ := uuid.Parse(id)
		tmpl := domain.Template{
			ID:      domain.IDType(uid),
			Name:    domain.NameType(tName),
			Code:    domain.CodeType(code),
			Version: domain.VersionType(version),
			Content: domain.HTMLType(content),
			Subject: domain.SubjectType(subject),
		}
		if sID != nil {
			suid, _ := uuid.Parse(*sID)
			sidType := domain.IDType(suid)
			tmpl.Stylesheet = &sidType
		}
		results = append(results, tmpl)
	}

	// ensure it returns an empty slice rather than nil so JSON encodes to []
	if results == nil {
		results = make([]domain.Template, 0)
	}
	
	return results, nil
}

func (r *templateRepository) FindByID(ctx context.Context, id domain.IDType) (domain.Template, error) {
	row := r.db.QueryRow(ctx, `SELECT id, name, code, version, stylesheet_id, html_content, subject FROM templates WHERE id = $1`, uuid.UUID(id).String())

	var dbID, tName, code, content, subject string
	var version int
	var sID *string

	if err := row.Scan(&dbID, &tName, &code, &version, &sID, &content, &subject); err != nil {
		return domain.Template{}, err
	}

	uid, _ := uuid.Parse(dbID)
	tmpl := domain.Template{
		ID:      domain.IDType(uid),
		Name:    domain.NameType(tName),
		Code:    domain.CodeType(code),
		Version: domain.VersionType(version),
		Content: domain.HTMLType(content),
		Subject: domain.SubjectType(subject),
	}
	if sID != nil {
		suid, _ := uuid.Parse(*sID)
		sidType := domain.IDType(suid)
		tmpl.Stylesheet = &sidType
	}
	return tmpl, nil
}

func (r *templateRepository) Update(ctx context.Context, t domain.Template) error {
	var sID *string
	if t.Stylesheet != nil {
		val := uuid.UUID(*t.Stylesheet).String()
		sID = &val
	}
	_, err := r.db.Exec(ctx,
		`UPDATE templates SET name=$1, code=$2, version=$3, stylesheet_id=$4, html_content=$5, subject=$6 WHERE id=$7`,
		string(t.Name), string(t.Code), int(t.Version), sID, string(t.Content), string(t.Subject), uuid.UUID(t.ID).String(),
	)
	if err != nil {
		if strings.Contains(err.Error(), "23503") || strings.Contains(err.Error(), "templates_stylesheet_id_fkey") {
			return errors.New("invalid stylesheet_id: the referenced stylesheet does not exist")
		}
		return err
	}
	return nil
}
