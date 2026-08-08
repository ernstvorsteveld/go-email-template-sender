package postgres

import (
	"context"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type stylesheetRepository struct {
	db *pgxpool.Pool
}

func NewStylesheetRepository(db *pgxpool.Pool) out.StylesheetRepository {
	return &stylesheetRepository{db: db}
}

func (r *stylesheetRepository) Save(ctx context.Context, s domain.Stylesheet) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO stylesheets (id, name, code, css_content) VALUES ($1, $2, $3, $4)`,
		uuid.UUID(s.ID).String(), string(s.Name), string(s.Code), string(s.Content),
	)
	return err
}

func (r *stylesheetRepository) FindAll(ctx context.Context, name domain.NameType) ([]domain.Stylesheet, error) {
	query := `SELECT id, name, code, css_content FROM stylesheets`
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

	results := make([]domain.Stylesheet, 0)
	for rows.Next() {
		var id, sName, code, content string
		if err := rows.Scan(&id, &sName, &code, &content); err != nil {
			return nil, err
		}
		uid, _ := uuid.Parse(id)
		results = append(results, domain.Stylesheet{
			ID:      domain.IDType(uid),
			Name:    domain.NameType(sName),
			Code:    domain.CodeType(code),
			Content: domain.CSSType(content),
		})
	}
	return results, nil
}

func (r *stylesheetRepository) FindByID(ctx context.Context, id domain.IDType) (domain.Stylesheet, error) {
	row := r.db.QueryRow(ctx, `SELECT id, name, code, css_content FROM stylesheets WHERE id = $1`, uuid.UUID(id).String())
	var dbID, sName, code, content string
	if err := row.Scan(&dbID, &sName, &code, &content); err != nil {
		return domain.Stylesheet{}, err
	}
	uid, _ := uuid.Parse(dbID)
	return domain.Stylesheet{
		ID:      domain.IDType(uid),
		Name:    domain.NameType(sName),
		Code:    domain.CodeType(code),
		Content: domain.CSSType(content),
	}, nil
}

func (r *stylesheetRepository) Update(ctx context.Context, s domain.Stylesheet) error {
	_, err := r.db.Exec(ctx,
		`UPDATE stylesheets SET name=$1, code=$2, css_content=$3 WHERE id=$4`,
		string(s.Name), string(s.Code), string(s.Content), uuid.UUID(s.ID).String(),
	)
	return err
}
