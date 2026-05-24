CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Block: the core log entry
CREATE TABLE IF NOT EXISTS blocks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     VARCHAR(64)  NOT NULL DEFAULT 'system',
    content     TEXT         NOT NULL,
    status      VARCHAR(16)  NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'completed', 'blocked')),
    metadata    JSONB        DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blocks_created_at ON blocks (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_blocks_status     ON blocks (status);
CREATE INDEX IF NOT EXISTS idx_blocks_user_id    ON blocks (user_id);
CREATE INDEX IF NOT EXISTS idx_blocks_content_gin ON blocks
    USING GIN (to_tsvector('simple', content));

-- Node: extracted tags / entities (#project, @person, etc.)
CREATE TABLE IF NOT EXISTS nodes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(32)  NOT NULL
        CHECK (type IN ('project', 'person', 'custom')),
    metadata    JSONB        DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (name, type)
);

CREATE INDEX IF NOT EXISTS idx_nodes_type       ON nodes (type);
CREATE INDEX IF NOT EXISTS idx_nodes_name_trgm  ON nodes USING GIN (name gin_trgm_ops);

-- Relation: directed edges for the knowledge graph
CREATE TABLE IF NOT EXISTS relations (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_type   VARCHAR(16) NOT NULL DEFAULT 'block'
        CHECK (source_type IN ('block')),
    source_id     UUID        NOT NULL,
    target_type   VARCHAR(16) NOT NULL
        CHECK (target_type IN ('block', 'node')),
    target_id     UUID        NOT NULL,
    relation_type VARCHAR(32) NOT NULL
        CHECK (relation_type IN ('mentions', 'blocks', 'references')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_relations_source ON relations (source_id);
CREATE INDEX IF NOT EXISTS idx_relations_target ON relations (target_id, target_type);
CREATE INDEX IF NOT EXISTS idx_relations_type   ON relations (relation_type);

-- Webhook token registry
CREATE TABLE IF NOT EXISTS webhook_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    token_hash  VARCHAR(64)  NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ
);
