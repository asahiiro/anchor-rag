CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL CHECK (btrim(name) <> ''),
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE chunks (
	id UUID PRIMARY KEY,
	document_id UUID NOT NULL
	REFERENCES documents (id)
	ON DELETE CASCADE,
	content TEXT NOT NULL,
	position INTEGER NOT NULL CHECK (position >= 0),
	embedding VECTOR(1024) NOT NULL,
	UNIQUE (document_id, position)
);

CREATE INDEX chunks_document_id_idx
ON chunks (document_id);

CREATE INDEX chunks_embedding_hnsw_idx
ON chunks
USING hnsw (embedding vector_cosine_ops);
