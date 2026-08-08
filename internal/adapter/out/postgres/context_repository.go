package postgres

import (
	"context"
	"encoding/json"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextRepository struct {
	db *pgxpool.Pool
}

func NewContextRepository(db *pgxpool.Pool) out.ContextRepository {
	return &contextRepository{db: db}
}

func (r *contextRepository) ExecuteQuery(ctx context.Context, query domain.SQLQueryType) ([]domain.Context, error) {
	rows, err := r.db.Query(ctx, string(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]domain.Context, 0)
	for rows.Next() {
		var id, ref, cust, payloadStr, ejp string
		if err := rows.Scan(&id, &ref, &cust, &payloadStr, &ejp); err != nil {
			return nil, err
		}
		uid, _ := uuid.Parse(id)
		results = append(results, domain.Context{
			ID:           domain.IDType(uid),
			Reference:    domain.ReferenceType(ref),
			Customer:     domain.CustomerType(cust),
			Payload:      domain.JSONPayloadType(payloadStr),
			EmailAddress: domain.JSONPathType(ejp),
		})
	}
	return results, nil
}

func (r *contextRepository) Save(ctx context.Context, c domain.Context) error {
	if !json.Valid([]byte(c.Payload)) {
		c.Payload = "{}"
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO contexts (id, reference_id, customer_name, payload, email_jsonpath) VALUES ($1, $2, $3, $4, $5)`,
		uuid.UUID(c.ID).String(), string(c.Reference), string(c.Customer), string(c.Payload), string(c.EmailAddress),
	)
	return err
}

func (r *contextRepository) FindAll(ctx context.Context, customer domain.CustomerType) ([]domain.Context, error) {
	query := `SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts`
	var args []interface{}
	if customer != "" {
		query += ` WHERE customer_name ILIKE $1`
		args = append(args, "%"+string(customer)+"%")
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]domain.Context, 0)
	for rows.Next() {
		var id, ref, cust, payloadStr, ejp string
		if err := rows.Scan(&id, &ref, &cust, &payloadStr, &ejp); err != nil {
			return nil, err
		}
		uid, _ := uuid.Parse(id)
		results = append(results, domain.Context{
			ID:           domain.IDType(uid),
			Reference:    domain.ReferenceType(ref),
			Customer:     domain.CustomerType(cust),
			Payload:      domain.JSONPayloadType(payloadStr),
			EmailAddress: domain.JSONPathType(ejp),
		})
	}
	return results, nil
}

func (r *contextRepository) FindByID(ctx context.Context, id domain.IDType) (domain.Context, error) {
	row := r.db.QueryRow(ctx, `SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts WHERE id = $1`, uuid.UUID(id).String())
	var dbID, ref, cust, payloadStr, ejp string
	if err := row.Scan(&dbID, &ref, &cust, &payloadStr, &ejp); err != nil {
		return domain.Context{}, err
	}
	uid, _ := uuid.Parse(dbID)
	return domain.Context{
		ID:           domain.IDType(uid),
		Reference:    domain.ReferenceType(ref),
		Customer:     domain.CustomerType(cust),
		Payload:      domain.JSONPayloadType(payloadStr),
		EmailAddress: domain.JSONPathType(ejp),
	}, nil
}

func (r *contextRepository) Update(ctx context.Context, c domain.Context) error {
	if !json.Valid([]byte(c.Payload)) {
		c.Payload = "{}"
	}
	_, err := r.db.Exec(ctx,
		`UPDATE contexts SET reference_id=$1, customer_name=$2, payload=$3, email_jsonpath=$4 WHERE id=$5`,
		string(c.Reference), string(c.Customer), string(c.Payload), string(c.EmailAddress), uuid.UUID(c.ID).String(),
	)
	return err
}
