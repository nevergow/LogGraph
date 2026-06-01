package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"loggraph/internal/model"
	"loggraph/internal/parser"
)

type BlockRepo struct{ pool *pgxpool.Pool }

func NewBlockRepo(pool *pgxpool.Pool) *BlockRepo { return &BlockRepo{pool: pool} }

// Create inserts a block and all derived relations in a single transaction.
// Only links nodes that already exist — does not auto-create.
func (r *BlockRepo) Create(ctx context.Context, input model.CreateBlockInput) (*model.Block, error) {
	parsed := parser.Parse(input.Content)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var block model.Block
	err = tx.QueryRow(ctx,
		`INSERT INTO blocks (user_id, content, status, metadata)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, content, status, metadata, created_at, updated_at`,
		input.UserID, input.Content, string(parsed.Status), input.Metadata,
	).Scan(&block.ID, &block.UserID, &block.Content, &block.Status, &block.Metadata, &block.CreatedAt, &block.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert block: %w", err)
	}

	// Link existing project nodes from &tags (only if node already exists)
	for _, name := range parsed.Tags {
		if err := r.insertRelation(ctx, tx, block.ID, "node", name, model.NodeTypeProject, model.RelationMentions); err != nil {
			return nil, err
		}
	}

	// Link existing person nodes from @mentions (only if node already exists)
	for _, name := range parsed.Mentions {
		if err := r.insertRelation(ctx, tx, block.ID, "node", name, model.NodeTypePerson, model.RelationMentions); err != nil {
			return nil, err
		}
	}

	// Insert references to other blocks
	for _, refID := range parsed.References {
		if err := r.insertRelationByID(ctx, tx, block.ID, "block", refID, model.RelationReferences); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &block, nil
}

// Update replaces a block's content and/or status.
func (r *BlockRepo) Update(ctx context.Context, id string, input model.UpdateBlockInput) (*model.Block, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var block model.Block

	if input.Content != nil && *input.Content != "" {
		parsed := parser.Parse(*input.Content)
		err = tx.QueryRow(ctx,
			`UPDATE blocks SET content=$1, status=$2, metadata=COALESCE($3, metadata), updated_at=NOW()
			 WHERE id=$4
			 RETURNING id, user_id, content, status, metadata, created_at, updated_at`,
			*input.Content, string(parsed.Status), input.Metadata, id,
		).Scan(&block.ID, &block.UserID, &block.Content, &block.Status, &block.Metadata, &block.CreatedAt, &block.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("update block: %w", err)
		}

		// Rebuild relations from parsed content (only links existing nodes)
		if _, err := tx.Exec(ctx, `DELETE FROM relations WHERE source_id=$1`, id); err != nil {
			return nil, fmt.Errorf("delete old relations: %w", err)
		}
		for _, name := range parsed.Tags {
			if err := r.insertRelation(ctx, tx, block.ID, "node", name, model.NodeTypeProject, model.RelationMentions); err != nil {
				return nil, err
			}
		}
		for _, name := range parsed.Mentions {
			if err := r.insertRelation(ctx, tx, block.ID, "node", name, model.NodeTypePerson, model.RelationMentions); err != nil {
				return nil, err
			}
		}
		for _, refID := range parsed.References {
			if err := r.insertRelationByID(ctx, tx, block.ID, "block", refID, model.RelationReferences); err != nil {
				return nil, err
			}
		}
	} else if input.Status != nil && *input.Status != "" {
		err = tx.QueryRow(ctx,
			`UPDATE blocks SET status=$1, updated_at=NOW()
			 WHERE id=$2
			 RETURNING id, user_id, content, status, metadata, created_at, updated_at`,
			*input.Status, id,
		).Scan(&block.ID, &block.UserID, &block.Content, &block.Status, &block.Metadata, &block.CreatedAt, &block.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("update block status: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no content or status provided")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &block, nil
}

// GetByID returns a single block.
func (r *BlockRepo) GetByID(ctx context.Context, id string) (*model.Block, error) {
	var b model.Block
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, content, status, metadata, created_at, updated_at
		 FROM blocks WHERE id=$1`, id,
	).Scan(&b.ID, &b.UserID, &b.Content, &b.Status, &b.Metadata, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get block: %w", err)
	}
	return &b, nil
}

// List returns blocks with cursor-based pagination and optional filters.
func (r *BlockRepo) List(ctx context.Context, q model.BlockListQuery) (*model.CursorPage[model.Block], error) {
	limit := q.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	args := []any{}
	conds := []string{}

	if q.Project != nil && *q.Project != "" {
		conds = append(conds, fmt.Sprintf(
			`id IN (SELECT source_id FROM relations r
			        JOIN nodes n ON r.target_id=n.id AND r.target_type='node'
			        WHERE n.name=$%d AND n.type='project' AND r.relation_type='mentions')`, len(args)+1))
		args = append(args, *q.Project)
	}
	if q.Person != nil && *q.Person != "" {
		conds = append(conds, fmt.Sprintf(
			`id IN (SELECT source_id FROM relations r
			        JOIN nodes n ON r.target_id=n.id
			        WHERE n.name=$%d AND n.type='person' AND r.relation_type='mentions')`, len(args)+1))
		args = append(args, *q.Person)
	}
	if q.Status != nil && *q.Status != "" {
		conds = append(conds, fmt.Sprintf("status=$%d", len(args)+1))
		args = append(args, *q.Status)
	}
	if q.Query != nil && *q.Query != "" {
		conds = append(conds, fmt.Sprintf("to_tsvector('simple', content) @@ plainto_tsquery('simple', $%d)", len(args)+1))
		args = append(args, *q.Query)
	}
	if q.Cursor != nil && *q.Cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, *q.Cursor)
		if err == nil {
			conds = append(conds, fmt.Sprintf("created_at < $%d", len(args)+1))
			args = append(args, t)
		}
	}
	if q.Since != nil && *q.Since != "" {
		t, err := time.Parse(time.RFC3339, *q.Since)
		if err == nil {
			conds = append(conds, fmt.Sprintf("created_at >= $%d", len(args)+1))
			args = append(args, t)
		}
	}
	if q.Until != nil && *q.Until != "" {
		t, err := time.Parse(time.RFC3339, *q.Until)
		if err == nil {
			conds = append(conds, fmt.Sprintf("created_at <= $%d", len(args)+1))
			args = append(args, t)
		}
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, content, status, metadata, created_at, updated_at
		 FROM blocks %s ORDER BY created_at DESC LIMIT $%d`, where, len(args)+1)
	args = append(args, limit+1) // fetch one extra to detect has_more

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list blocks: %w", err)
	}
	defer rows.Close()

	var blocks []model.Block
	for rows.Next() {
		var b model.Block
		if err := rows.Scan(&b.ID, &b.UserID, &b.Content, &b.Status, &b.Metadata, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan block: %w", err)
		}
		blocks = append(blocks, b)
	}

	hasMore := len(blocks) > limit
	if hasMore {
		blocks = blocks[:limit]
	}

	page := &model.CursorPage[model.Block]{Data: blocks, HasMore: hasMore}
	if hasMore && len(blocks) > 0 {
		page.NextCursor = blocks[len(blocks)-1].CreatedAt.Format(time.RFC3339Nano)
	}
	return page, nil
}

// Delete removes a block and its relations.
func (r *BlockRepo) Delete(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM relations WHERE source_id=$1 OR (target_type='block' AND target_id=$1)`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM blocks WHERE id=$1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ReplaceProjectInContent replaces all occurrences of &oldName with &newName
// in block content. Used when renaming a project/person node with cascade.
func (r *BlockRepo) ReplaceProjectInContent(ctx context.Context, oldName, newName string) (int64, error) {
	escapedOld := escapeRegex(oldName)
	pattern := `(^|\s)&` + escapedOld + `(\s|$)`
	replacement := `\1&` + newName + `\2`

	tag, err := r.pool.Exec(ctx,
		`UPDATE blocks
		 SET content = regexp_replace(content, $1, $2, 'g'),
		     updated_at = NOW()
		 WHERE content ~ $3`,
		pattern, replacement, `(^|\s)&`+escapedOld+`(\s|$)`)
	if err != nil {
		return 0, fmt.Errorf("replace project in content: %w", err)
	}
	return tag.RowsAffected(), nil
}

// escapeRegex escapes special characters for PostgreSQL regex (~ operator).
func escapeRegex(s string) string {
	special := []string{`\`, `.`, `*`, `+`, `?`, `[`, `]`, `(`, `)`, `{`, `}`, `|`, `^`, `$`}
	for _, c := range special {
		s = strings.ReplaceAll(s, c, `\`+c)
	}
	return s
}

// BackfillRelationsForNode scans existing blocks for matching &name / #name / @name
// text and creates relations. Call after manually creating a node so existing
// blocks that already mention this node get linked.
func (r *BlockRepo) BackfillRelationsForNode(ctx context.Context, nodeName string, nodeType model.NodeType) (int64, error) {
	escaped := escapeRegex(nodeName)
	var pattern string
	if nodeType == model.NodeTypePerson {
		pattern = `(^|\s)@` + escaped + `(\s|$)`
	} else {
		pattern = `(^|\s)(&|#)` + escaped + `(\s|$)`
	}

	relType := string(model.RelationMentions)

	tag, err := r.pool.Exec(ctx,
		`INSERT INTO relations (source_type, source_id, target_type, target_id, relation_type)
		 SELECT 'block', b.id, 'node', n.id, $1
		 FROM blocks b, nodes n
		 WHERE n.name = $2 AND n.type = $3
		   AND b.content ~ $4
		   AND NOT EXISTS (
		     SELECT 1 FROM relations r
		     WHERE r.source_id = b.id AND r.target_id = n.id AND r.relation_type = $5
		   )`,
		relType, nodeName, string(nodeType), pattern, relType)
	if err != nil {
		return 0, fmt.Errorf("backfill relations for node: %w", err)
	}
	return tag.RowsAffected(), nil
}

// ── helpers ──

func (r *BlockRepo) insertRelation(ctx context.Context, tx pgx.Tx, sourceID, targetType, nodeName string, nodeType model.NodeType, relType model.RelationType) error {
	// Auto-create node if it doesn't exist, so &NewProject in content
	// creates both the node and the relation in a single operation.
	_, err := tx.Exec(ctx,
		`INSERT INTO nodes (name, type) VALUES ($1, $2) ON CONFLICT (name, type) DO NOTHING`,
		nodeName, string(nodeType))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO relations (source_type, source_id, target_type, target_id, relation_type)
		 SELECT 'block', $1, $2, id, $3 FROM nodes WHERE name=$4 AND type=$5`,
		sourceID, targetType, string(relType), nodeName, string(nodeType))
	return err
}

func (r *BlockRepo) insertRelationByID(ctx context.Context, tx pgx.Tx, sourceID, targetType, targetID string, relType model.RelationType) error {
	// Verify target exists
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM blocks WHERE id=$1)`, targetID).Scan(&exists); err != nil || !exists {
		return nil // silently skip invalid references
	}
	_, err := tx.Exec(ctx,
		`INSERT INTO relations (source_type, source_id, target_type, target_id, relation_type)
		 VALUES ('block', $1, $2, $3, $4)`,
		sourceID, targetType, targetID, string(relType))
	return err
}
