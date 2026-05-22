package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"loggraph/internal/model"
)

type WebhookTokenRepo struct{ pool *pgxpool.Pool }

func NewWebhookTokenRepo(pool *pgxpool.Pool) *WebhookTokenRepo {
	return &WebhookTokenRepo{pool: pool}
}

// Generate creates a random token, stores its SHA256 hash, and returns the plaintext.
func (r *WebhookTokenRepo) Generate(ctx context.Context, name string) (plaintext string, token *model.WebhookToken, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("rand: %w", err)
	}
	plaintext = hex.EncodeToString(raw)
	hash := sha256Hash(plaintext)

	var t model.WebhookToken
	err = r.pool.QueryRow(ctx,
		`INSERT INTO webhook_tokens (name, token_hash) VALUES ($1, $2)
		 RETURNING id, name, token_hash, created_at, expires_at`,
		name, hash,
	).Scan(&t.ID, &t.Name, &t.TokenHash, &t.CreatedAt, &t.ExpiresAt)
	if err != nil {
		return "", nil, fmt.Errorf("insert token: %w", err)
	}
	return plaintext, &t, nil
}

// List returns all tokens (never the hash).
func (r *WebhookTokenRepo) List(ctx context.Context) ([]model.WebhookToken, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, token_hash, created_at, expires_at FROM webhook_tokens ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	defer rows.Close()

	var tokens []model.WebhookToken
	for rows.Next() {
		var t model.WebhookToken
		if err := rows.Scan(&t.ID, &t.Name, &t.TokenHash, &t.CreatedAt, &t.ExpiresAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// Delete removes a token by ID.
func (r *WebhookTokenRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM webhook_tokens WHERE id=$1`, id)
	return err
}

// Validate checks whether a plaintext token matches a stored hash.
func (r *WebhookTokenRepo) Validate(ctx context.Context, plaintext string) (bool, error) {
	hash := sha256Hash(plaintext)
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM webhook_tokens
		              WHERE token_hash=$1 AND (expires_at IS NULL OR expires_at > NOW()))`,
		hash).Scan(&exists)
	return exists, err
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
