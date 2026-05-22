package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"loggraph/internal/model"
)

type RelationRepo struct{ pool *pgxpool.Pool }

func NewRelationRepo(pool *pgxpool.Pool) *RelationRepo { return &RelationRepo{pool: pool} }

// GraphForBlock returns the local knowledge graph centered on a block.
func (r *RelationRepo) GraphForBlock(ctx context.Context, blockID string) (*model.GraphData, error) {
	// Gather unique node IDs related to this block
	rows, err := r.pool.Query(ctx,
		`SELECT r.target_type, r.target_id, r.relation_type
		 FROM relations r
		 WHERE r.source_id = $1
		 UNION
		 SELECT r.source_type, r.source_id, r.relation_type
		 FROM relations r
		 WHERE r.target_type = 'block' AND r.target_id = $1
		 UNION
		 SELECT r.target_type, r.target_id, r.relation_type
		 FROM relations r
		 WHERE r.target_type = 'block' AND r.target_id IN (
		     SELECT target_id FROM relations WHERE source_id = $1 AND target_type = 'block'
		 )`, blockID)
	if err != nil {
		return nil, fmt.Errorf("query relations: %w", err)
	}
	defer rows.Close()

	nodeIDs := map[string]bool{}
	blockIDs := map[string]bool{blockID: true}
	var edges []model.GraphEdge

	for rows.Next() {
		var targetType, targetID, relType string
		if err := rows.Scan(&targetType, &targetID, &relType); err != nil {
			return nil, fmt.Errorf("scan relation: %w", err)
		}
		edges = append(edges, model.GraphEdge{
			Source: blockID, Target: targetID, Label: relType,
		})
		if targetType == "node" {
			nodeIDs[targetID] = true
		} else {
			blockIDs[targetID] = true
		}
	}

	// Fetch nodes
	var nodes []model.Node
	if len(nodeIDs) > 0 {
		ids := make([]string, 0, len(nodeIDs))
		for id := range nodeIDs {
			ids = append(ids, id)
		}
		nRows, err := r.pool.Query(ctx,
			`SELECT id, name, type, metadata, created_at FROM nodes WHERE id = ANY($1)`, ids)
		if err != nil {
			return nil, fmt.Errorf("query nodes: %w", err)
		}
		defer nRows.Close()
		for nRows.Next() {
			var n model.Node
			if err := nRows.Scan(&n.ID, &n.Name, &n.Type, &n.Metadata, &n.CreatedAt); err != nil {
				return nil, err
			}
			nodes = append(nodes, n)
		}
	}

	// Fetch the central block
	var block model.Block
	err = r.pool.QueryRow(ctx,
		`SELECT id, user_id, content, status, metadata, created_at, updated_at
		 FROM blocks WHERE id=$1`, blockID,
	).Scan(&block.ID, &block.UserID, &block.Content, &block.Status, &block.Metadata, &block.CreatedAt, &block.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("query block: %w", err)
	}

	return &model.GraphData{Block: block, Nodes: nodes, Edges: edges}, nil
}

// GraphForNode returns a graph centered on a node (project/person/standard).
func (r *RelationRepo) GraphForNode(ctx context.Context, nodeID string) (*model.GraphData, error) {
	// Fetch the node
	var node model.Node
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, type, metadata, created_at FROM nodes WHERE id=$1`, nodeID,
	).Scan(&node.ID, &node.Name, &node.Type, &node.Metadata, &node.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("query node: %w", err)
	}

	// Fetch all blocks that mention this node (as source)
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.user_id, b.content, b.status, b.metadata, b.created_at, b.updated_at,
		        r.relation_type
		 FROM blocks b
		 JOIN relations r ON r.source_id = b.id AND r.source_type = 'block'
		 WHERE r.target_type = 'node' AND r.target_id = $1
		 ORDER BY b.created_at DESC
		 LIMIT 50`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("query blocks for node: %w", err)
	}
	defer rows.Close()

	var blocks []model.Block
	type edge struct {
		Source string
		Target string
		Label  string
	}
	var edgeSet = map[string]bool{}
	var edges []model.GraphEdge
	relatedNodeIDs := map[string]bool{}

	// Collect block IDs
	blockIDs := []string{}
	for rows.Next() {
		var b model.Block
		var relType string
		if err := rows.Scan(&b.ID, &b.UserID, &b.Content, &b.Status, &b.Metadata, &b.CreatedAt, &b.UpdatedAt, &relType); err != nil {
			return nil, fmt.Errorf("scan block: %w", err)
		}
		blocks = append(blocks, b)
		blockIDs = append(blockIDs, b.ID)

		key := b.ID + "-" + nodeID + "-" + relType
		if !edgeSet[key] {
			edgeSet[key] = true
			edges = append(edges, model.GraphEdge{Source: b.ID, Target: nodeID, Label: relType})
		}
	}

	// Fetch relations between these blocks and other nodes
	if len(blockIDs) > 0 {
		r2, err := r.pool.Query(ctx,
			`SELECT r.source_id, r.target_id, r.relation_type, n.id, n.name, n.type
			 FROM relations r
			 JOIN nodes n ON r.target_id = n.id AND r.target_type = 'node'
			 WHERE r.source_id = ANY($1) AND r.target_id != $2`, blockIDs, nodeID)
		if err != nil {
			return nil, fmt.Errorf("query secondary relations: %w", err)
		}
		defer r2.Close()
		for r2.Next() {
			var srcID, tgtID, relType, nID, nName, nType string
			if err := r2.Scan(&srcID, &tgtID, &relType, &nID, &nName, &nType); err != nil {
				return nil, err
			}
			relatedNodeIDs[nID] = true
			key := srcID + "-" + tgtID
			if !edgeSet[key] {
				edgeSet[key] = true
				edges = append(edges, model.GraphEdge{Source: srcID, Target: tgtID, Label: relType})
			}
		}
	}

	// Also fetch references between the blocks
	if len(blockIDs) > 0 {
		r3, err := r.pool.Query(ctx,
			`SELECT source_id, target_id, relation_type
			 FROM relations
			 WHERE source_id = ANY($1) AND target_type = 'block'`, blockIDs)
		if err != nil {
			return nil, fmt.Errorf("query block refs: %w", err)
		}
		defer r3.Close()
		for r3.Next() {
			var srcID, tgtID, relType string
			if err := r3.Scan(&srcID, &tgtID, &relType); err != nil {
				return nil, err
			}
			key := srcID + "-" + tgtID
			if !edgeSet[key] {
				edgeSet[key] = true
				edges = append(edges, model.GraphEdge{Source: srcID, Target: tgtID, Label: relType})
			}
		}
	}

	// Fetch related nodes
	var nodes []model.Node
	nodes = append(nodes, node) // center node
	if len(relatedNodeIDs) > 0 {
		ids := make([]string, 0, len(relatedNodeIDs))
		for id := range relatedNodeIDs {
			ids = append(ids, id)
		}
		nRows, err := r.pool.Query(ctx,
			`SELECT id, name, type, metadata, created_at FROM nodes WHERE id = ANY($1)`, ids)
		if err != nil {
			return nil, fmt.Errorf("query related nodes: %w", err)
		}
		defer nRows.Close()
		for nRows.Next() {
			var n model.Node
			if err := nRows.Scan(&n.ID, &n.Name, &n.Type, &n.Metadata, &n.CreatedAt); err != nil {
				return nil, err
			}
			nodes = append(nodes, n)
		}
	}

	if blocks == nil {
		blocks = []model.Block{}
	}

	return &model.GraphData{Block: model.Block{
		ID:      node.ID,
		UserID:  "system",
		Content: node.Name,
		Status:  "active",
	}, Nodes: nodes, Edges: edges}, nil
}
