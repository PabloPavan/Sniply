-- Extensões (fuzzy no name via trigram)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Enum de visibilidade
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'snippet_visibility') THEN
    CREATE TYPE snippet_visibility AS ENUM ('public', 'private');
  END IF;
END$$;

-- Usuários
CREATE TABLE IF NOT EXISTS users (
  id            TEXT PRIMARY KEY,
  email         TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Snippets
CREATE TABLE IF NOT EXISTS snippets (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  content     TEXT NOT NULL,
  language    TEXT NOT NULL DEFAULT 'txt',
  tags        TEXT[] NOT NULL DEFAULT '{}',
  visibility  snippet_visibility NOT NULL DEFAULT 'private',
  creator_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  search_tsv  TSVECTOR
);

-- Trigger para updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION snippets_set_search_tsv()
RETURNS TRIGGER AS $$
BEGIN
  NEW.search_tsv :=
    to_tsvector(
      'simple',
      coalesce(NEW.name,'') || ' ' ||
      coalesce(NEW.content,'') || ' ' ||
      array_to_string(NEW.tags,' ')
    );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_snippets_search_tsv ON snippets;
CREATE TRIGGER trg_snippets_search_tsv
BEFORE INSERT OR UPDATE OF name, content, tags ON snippets
FOR EACH ROW
EXECUTE FUNCTION snippets_set_search_tsv();


DROP TRIGGER IF EXISTS trg_snippets_updated_at ON snippets;
CREATE TRIGGER trg_snippets_updated_at
BEFORE UPDATE ON snippets
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Índices para listagens/filtros
CREATE INDEX IF NOT EXISTS idx_snippets_creator_updated
  ON snippets (creator_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_snippets_visibility
  ON snippets (visibility);

CREATE INDEX IF NOT EXISTS idx_snippets_language
  ON snippets (language);

-- Fuzzy para name (trigram)
CREATE INDEX IF NOT EXISTS idx_snippets_name_trgm
  ON snippets USING GIN (name gin_trgm_ops);

-- Full-text search (conteúdo + nome + tags)
CREATE INDEX IF NOT EXISTS idx_snippets_fts
  ON snippets USING GIN (search_tsv);

-- Usuário demo
INSERT INTO users (id, email, password_hash)
VALUES ('usr_demo', 'demo@local', 'x')
ON CONFLICT (id) DO NOTHING;