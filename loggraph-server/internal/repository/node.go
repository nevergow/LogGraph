package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"loggraph/internal/model"
)

type NodeRepo struct{ pool *pgxpool.Pool }

func NewNodeRepo(pool *pgxpool.Pool) *NodeRepo { return &NodeRepo{pool: pool} }

// List returns nodes, optionally filtered by type.
func (r *NodeRepo) List(ctx context.Context, nodeType *string) ([]model.Node, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if nodeType != nil && *nodeType != "" {
		rows, err = r.pool.Query(ctx,
			`SELECT id, name, type, metadata, created_at FROM nodes WHERE type=$1 ORDER BY name`, *nodeType)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, name, type, metadata, created_at FROM nodes ORDER BY type, name`)
	}
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.ID, &n.Name, &n.Type, &n.Metadata, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// Suggest returns node names matching a prefix, for autocomplete.
func (r *NodeRepo) Suggest(ctx context.Context, query string, nodeType *string, limit int) ([]model.Node, error) {
	if limit <= 0 {
		limit = 8
	}
	args := []any{query, limit}
	typeFilter := ""
	if nodeType != nil && *nodeType != "" {
		typeFilter = fmt.Sprintf("AND type=$%d", len(args)+1)
		args = append(args, *nodeType)
	}

	sql := fmt.Sprintf(
		`SELECT id, name, type, metadata, created_at
		 FROM nodes WHERE similarity(name, $1) > 0.15 %s
		 ORDER BY similarity(name, $1) DESC LIMIT $2`, typeFilter)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("suggest nodes: %w", err)
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.ID, &n.Name, &n.Type, &n.Metadata, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}
