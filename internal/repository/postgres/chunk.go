package postgres

import (
	"context"
	"fmt"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type ChunkRepository struct {
	pool *pgxpool.Pool
}

var _ repository.ChunkRepository = (*ChunkRepository)(nil)

func NewChunkRepository(
	pool *pgxpool.Pool,
) *ChunkRepository {
	return &ChunkRepository{
		pool: pool,
	}
}

func (r *ChunkRepository) SaveBatch(
	ctx context.Context,
	chunks []domain.Chunk,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"begin chunk transaction: %w",
			err,
		)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	for _, chunk := range chunks {
		_, err := tx.Exec(
			ctx,
			`
				INSERT INTO chunks (
					id,
					document_id,
					content,
					position,
					embedding
				)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (id)
				DO UPDATE SET
					document_id = EXCLUDED.document_id,
					content = EXCLUDED.content,
					position = EXCLUDED.position,
					embedding = EXCLUDED.embedding		
			`,
			chunk.ID,
			chunk.DocumentID,
			chunk.Content,
			chunk.Position,
			pgvector.NewVector(chunk.Embedding),
		)
		if err != nil {
			return fmt.Errorf(
				"save chunks %s: %w",
				chunk.ID,
				err,
			)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit chunks: %w",
			err,
		)
	}
	return nil
}

func (r *ChunkRepository) FindByDocumentID(
	ctx context.Context,
	documentID string,
) ([]domain.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT
			id,
			document_id,
			content,
			position,
			embedding
		FROM chunks
		WHERE document_id = $1
		ORDER BY position
		`,
		documentID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"find chunks by document ID: %w",
			err,
		)
	}

	return collectChunks(rows)
}

func (r *ChunkRepository) FindAll(
	ctx context.Context,
) ([]domain.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT
			id,
			document_id,
			content,
			position,
			embedding
		FROM chunks
		ORDER BY document_id, position
		`,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"find chunks by document ID: %w",
			err,
		)
	}

	return collectChunks(rows)
}

func collectChunks(
	rows pgx.Rows,
) ([]domain.Chunk, error) {
	defer rows.Close()

	chunks := make([]domain.Chunk, 0)

	for rows.Next() {
		var chunk domain.Chunk
		var storedEmbedding pgvector.Vector

		if err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.Content,
			&chunk.Position,
			&storedEmbedding,
		); err != nil {
			return nil, fmt.Errorf(
				"scan chunk: %w",
				err,
			)
		}

		chunk.Embedding = append(
			[]float32(nil),
			storedEmbedding.Slice()...,
		)
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate chunks: %w",
			err,
		)
	}

	return chunks, nil
}
