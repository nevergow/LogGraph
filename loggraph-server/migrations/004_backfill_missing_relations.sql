-- 004: Backfill missing project/person nodes and relations from existing block content.
-- Before this fix, writing &NewProject in a block would not auto-create the node or relation.
-- This migration scans existing blocks and creates any missing nodes + relations.

-- Helper: escape regex-special characters in a string
CREATE OR REPLACE FUNCTION escape_regex(text) RETURNS text AS $$
  SELECT regexp_replace($1, '([.^$*+?()[{\]|\\])', '\\\1', 'g');
$$ LANGUAGE sql IMMUTABLE;

-- Step 1: Create missing project nodes from &name and #name patterns in block content
INSERT INTO nodes (name, type)
SELECT DISTINCT m[1], 'project'
FROM (
  SELECT regexp_matches(content, '(?:^|\s)[&#]([^\s#&@^~\[\].,;:!?，。；：！？]+)', 'g') AS m
  FROM blocks
) t
ON CONFLICT (name, type) DO NOTHING;

-- Step 2: Create missing person nodes from @name patterns
INSERT INTO nodes (name, type)
SELECT DISTINCT m[1], 'person'
FROM (
  SELECT regexp_matches(content, '(?:^|\s)@([^\s#&@^~\[\].,;:!?，。；：！？]+)', 'g') AS m
  FROM blocks
) t
ON CONFLICT (name, type) DO NOTHING;

-- Step 3: Backfill project relations where missing
INSERT INTO relations (source_type, source_id, target_type, target_id, relation_type)
SELECT 'block', b.id, 'node', n.id, 'mentions'
FROM blocks b
JOIN nodes n ON n.type = 'project' AND b.content ~ ('(^|\s)[&#]' || escape_regex(n.name) || '(\s|$)')
WHERE NOT EXISTS (
  SELECT 1 FROM relations r
  WHERE r.source_id = b.id AND r.target_id = n.id AND r.relation_type = 'mentions'
);

-- Step 4: Backfill person relations where missing
INSERT INTO relations (source_type, source_id, target_type, target_id, relation_type)
SELECT 'block', b.id, 'node', n.id, 'mentions'
FROM blocks b
JOIN nodes n ON n.type = 'person' AND b.content ~ ('(^|\s)@' || escape_regex(n.name) || '(\s|$)')
WHERE NOT EXISTS (
  SELECT 1 FROM relations r
  WHERE r.source_id = b.id AND r.target_id = n.id AND r.relation_type = 'mentions'
);

-- Cleanup
DROP FUNCTION IF EXISTS escape_regex;
