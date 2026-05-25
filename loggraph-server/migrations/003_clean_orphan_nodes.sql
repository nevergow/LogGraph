-- 003: Remove orphan nodes (no relations) that were auto-created by the old
-- upsertNode behavior. After switching to whitelist-only recognition, these
-- phantom entries are no longer created, but existing orphans need cleanup.
DELETE FROM nodes
WHERE id NOT IN (
    SELECT DISTINCT target_id FROM relations WHERE target_type = 'node'
)
AND type IN ('project', 'person');
