package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	pool *pgxpool.Pool
}

var _ repository.DocumentRepository = (*DocumentRepository)(nil)

func NewDocumentRepository(
	pool *pgxpool.Pool,
) *DocumentRepository {
	return &DocumentRepository{
		pool: pool,
	}
}

func (r *DocumentRepository) Save(
	ctx context.Context,
	doc domain.Document,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := r.pool.Exec(
		ctx,
		`

			INSERT INTO documents (
				id,
				name,
				content,
				created_at
			)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id)
			DO UPDATE SET
				name = EXCLUDED.name,
				content = EXCLUDED.content,
				created_at = EXCLUDED.created_at
		`,
		doc.ID,
		doc.Name,
		doc.Content,
		doc.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"save document: %w",
			err,
		)
	}

	return nil
}

func (r *DocumentRepository) FindByID(
	ctx context.Context,
	id string,
) (domain.Document, error) {
	if err := ctx.Err(); err != nil {
		return domain.Document{}, err
	}

	var doc domain.Document

	err := r.pool.QueryRow(
		ctx,
		`

		SELECT
		id,
		name,
		content,
		created_at
		FROM documents
		WHERE id = $1
		`,
		id,
	).Scan(
		&doc.ID,
		&doc.Name,
		&doc.Content,
		&doc.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Document{},
			repository.ErrDocumentNotFound
	}

	if err != nil {
		return domain.Document{}, fmt.Errorf(
			"find document by ID: %w",
			err,
		)
	}

	return doc, nil
}
