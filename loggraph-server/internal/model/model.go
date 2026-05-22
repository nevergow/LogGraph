package model

import (
	"time"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusBlocked   Status = "blocked"
)

type NodeType string

const (
	NodeTypeProject  NodeType = "project"
	NodeTypePerson   NodeType = "person"
	NodeTypeStandard NodeType = "standard"
	NodeTypeCustom   NodeType = "custom"
)

type RelationType string

const (
	RelationMentions   RelationType = "mentions"
	RelationBlocks     RelationType = "blocks"
	RelationReferences RelationType = "references"
)

// ── Domain models ──

type Block struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Content   string     `json:"content"`
	Status    Status     `json:"status"`
	Metadata  JSONMap    `json:"metadata,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Node struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      NodeType   `json:"type"`
	Metadata  JSONMap    `json:"metadata,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Relation struct {
	ID           string       `json:"id"`
	SourceType   string       `json:"source_type"`
	SourceID     string       `json:"source_id"`
	TargetType   string       `json:"target_type"`
	TargetID     string       `json:"target_id"`
	RelationType RelationType `json:"relation_type"`
	CreatedAt    time.Time    `json:"created_at"`
}

type WebhookToken struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	TokenHash string     `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ── API DTOs ──

type CreateBlockInput struct {
	UserID   string  `json:"user_id"`
	Content  string  `json:"content"`
	Metadata JSONMap `json:"metadata,omitempty"`
}

type UpdateBlockInput struct {
	Content  *string `json:"content"`
	Status   *string `json:"status"`
	Metadata JSONMap `json:"metadata,omitempty"`
}

type BlockListQuery struct {
	Cursor  *string `json:"cursor"`
	Limit   int     `json:"limit"`
	Project *string `json:"project"`
	Person  *string `json:"person"`
	Status  *string `json:"status"`
	Query   *string `json:"q"`
	Since   *string `json:"since"`
	Until   *string `json:"until"`
}

type WebhookLogInput struct {
	UserID   string  `json:"user_id"`
	Content  string  `json:"content"`
	Metadata JSONMap  `json:"metadata,omitempty"`
}

type PresignInput struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

type PresignOutput struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
}

type GraphData struct {
	Block Block             `json:"block"`
	Nodes []Node            `json:"nodes"`
	Edges []GraphEdge       `json:"edges"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

type CursorPage[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// JSONMap is a convenience alias for JSONB fields.
type JSONMap map[string]any
