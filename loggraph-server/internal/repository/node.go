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

// GetByNameAndType finds a node by exact name and type match. Returns nil if not found.
func (r *NodeRepo) GetByNameAndType(ctx context.Context, name string, nodeType model.NodeType) (*model.Node, error) {
	var n model.Node
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, type, metadata, created_at FROM nodes WHERE name=$1 AND type=$2`,
		name, string(nodeType),
	).Scan(&n.ID, &n.Name, &n.Type, &n.Metadata, &n.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get node by name and type: %w", err)
	}
	return &n, nil
}

func (r *NodeRepo) GetByID(ctx context.Context, id string) (*model.Node, error) {
	var n model.Node
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, type, metadata, created_at FROM nodes WHERE id=$1`, id,
	).Scan(&n.ID, &n.Name, &n.Type, &n.Metadata, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get node by id: %w", err)
	}
	return &n, nil
}

func (r *NodeRepo) Create(ctx context.Context, name string, nodeType model.NodeType) (model.Node, error) {
	var n model.Node
	err := r.pool.QueryRow(ctx,
		`INSERT INTO nodes (name, type) VALUES ($1, $2)
		 ON CONFLICT (name, type) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id, name, type, metadata, created_at`,
		name, nodeType,
	).Scan(&n.ID, &n.Name, &n.Type, &n.Metadata, &n.CreatedAt)
	if err != nil {
		return model.Node{}, fmt.Errorf("create node: %w", err)
	}
	return n, nil
}

func (r *NodeRepo) Update(ctx context.Context, id, name string) (model.Node, error) {
	var n model.Node
	err := r.pool.QueryRow(ctx,
		`UPDATE nodes SET name=$1 WHERE id=$2
		 RETURNING id, name, type, metadata, created_at`,
		name, id,
	).Scan(&n.ID, &n.Name, &n.Type, &n.Metadata, &n.CreatedAt)
	if err != nil {
		return model.Node{}, fmt.Errorf("update node: %w", err)
	}
	return n, nil
}

func (r *NodeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM nodes WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete node: %w", err)
	}
	return nil
}
