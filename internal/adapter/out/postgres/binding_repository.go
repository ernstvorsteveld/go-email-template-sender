package postgres

import (
	"context"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bindingRepository struct {
	db *pgxpool.Pool
}

func NewBindingRepository(db *pgxpool.Pool) out.BindingRepository {
	return &bindingRepository{db: db}
}

func (r *bindingRepository) Save(ctx context.Context, b domain.Binding) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO bindings (id, name, query, template_id) VALUES ($1, $2, $3, $4)`,
		uuid.UUID(b.ID).String(), string(b.Name), string(b.Query), uuid.UUID(b.Template).String(),
	)
	return err
}

func (r *bindingRepository) FindAll(ctx context.Context, name domain.NameType) ([]domain.Binding, error) {
	query := `SELECT id, name, query, template_id FROM bindings`
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

	results := make([]domain.Binding, 0)
	for rows.Next() {
		var id, bName, q, tid string
		if err := rows.Scan(&id, &bName, &q, &tid); err != nil {
			return nil, err
		}
		uid, _ := uuid.Parse(id)
		tuid, _ := uuid.Parse(tid)
		results = append(results, domain.Binding{
			ID:         domain.IDType(uid),
			Name:       domain.NameType(bName),
			Query:      domain.SQLQueryType(q),
			Template: domain.IDType(tuid),
		})
	}
	return results, nil
}

func (r *bindingRepository) FindByID(ctx context.Context, id domain.IDType) (domain.Binding, error) {
	row := r.db.QueryRow(ctx, `SELECT id, name, query, template_id FROM bindings WHERE id = $1`, uuid.UUID(id).String())
	var dbID, bName, q, tid string
	if err := row.Scan(&dbID, &bName, &q, &tid); err != nil {
		return domain.Binding{}, err
	}
	uid, _ := uuid.Parse(dbID)
	tuid, _ := uuid.Parse(tid)
	return domain.Binding{
		ID:         domain.IDType(uid),
		Name:       domain.NameType(bName),
		Query:      domain.SQLQueryType(q),
		Template: domain.IDType(tuid),
	}, nil
}

func (r *bindingRepository) Update(ctx context.Context, b domain.Binding) error {
	_, err := r.db.Exec(ctx,
		`UPDATE bindings SET name=$1, query=$2, template_id=$3 WHERE id=$4`,
		string(b.Name), string(b.Query), uuid.UUID(b.Template).String(), uuid.UUID(b.ID).String(),
	)
	return err
}
