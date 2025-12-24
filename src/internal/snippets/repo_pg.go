package snippets

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type RepoPG struct {
	pool *pgxpool.Pool
}

func NewRepoPG(pool *pgxpool.Pool) *RepoPG {
	return &RepoPG{pool: pool}
}

func (r *RepoPG) Create(ctx context.Context, s *Snippet) error {
	// MVP: IDs como texto. Aqui você pode gerar ULID/UUID no app.
	// Para simplificar, vamos usar timestamp+random depois; por enquanto, deixe o caller gerar.
	const q = `
INSERT INTO snippets (id, name, content, language, tags, visibility, creator_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING created_at, updated_at;
`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return r.pool.QueryRow(ctx, q,
		s.ID,
		s.Name,
		s.Content,
		s.Language,
		s.Tags,
		string(s.Visibility),
		s.CreatorID,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (r *RepoPG) GetByIDPublicOnly(ctx context.Context, id string) (*Snippet, error) {
	const q = `
SELECT id, name, content, language, tags, visibility, creator_id, created_at, updated_at
FROM snippets
WHERE id = $1 AND visibility = 'public'
LIMIT 1;
`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var s Snippet
	var visibility string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&s.ID,
		&s.Name,
		&s.Content,
		&s.Language,
		&s.Tags,
		&visibility,
		&s.CreatorID,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, ErrNotFound
	}
	s.Visibility = Visibility(visibility)
	return &s, nil
}

func (r *RepoPG) List(ctx context.Context, f SnippetFilter) ([]*Snippet, error) {
	where := []string{"visibility = 'public'"}
	args := []interface{}{}
	argPos := 1

	if f.Creator != "" {
		where = append(where, fmt.Sprintf("creator_id = $%d", argPos))
		args = append(args, f.Creator)
		argPos++
	}
	if f.Language != "" {
		where = append(where, fmt.Sprintf("language = $%d", argPos))
		args = append(args, f.Language)
		argPos++
	}
	if f.Query != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR content ILIKE $%d)", argPos, argPos+1))
		qstr := "%" + strings.ReplaceAll(f.Query, "%", "\\%") + "%"
		args = append(args, qstr, qstr)
		argPos += 2
	}

	limit := 100
	if f.Limit > 0 && f.Limit <= 1000 {
		limit = f.Limit
	}
	offset := 0
	if f.Offset > 0 {
		offset = f.Offset
	}

	query := fmt.Sprintf(`SELECT id, name, content, language, tags, visibility, creator_id, created_at, updated_at
FROM snippets
WHERE %s
ORDER BY created_at DESC
LIMIT %d OFFSET %d;`, strings.Join(where, " AND "), limit, offset)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []*Snippet
	for rows.Next() {
		var s Snippet
		var visibility string
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Content,
			&s.Language,
			&s.Tags,
			&visibility,
			&s.CreatorID,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.Visibility = Visibility(visibility)
		snippets = append(snippets, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snippets, nil
}

func (r *RepoPG) Update(ctx context.Context, s *Snippet) error {
	const q = `
UPDATE snippets
SET name = $1, content = $2, language = $3, tags = $4, visibility = $5, updated_at = now()
WHERE id = $6
RETURNING updated_at;
`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return r.pool.QueryRow(ctx, q,
		s.Name,
		s.Content,
		s.Language,
		s.Tags,
		string(s.Visibility),
		s.ID,
	).Scan(&s.UpdatedAt)
}

func (r *RepoPG) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM snippets WHERE id = $1;`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.pool.Exec(ctx, q, id)
	return err
}
