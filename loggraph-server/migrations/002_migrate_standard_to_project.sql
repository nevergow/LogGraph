-- 002: Merge legacy standard nodes into project type.
-- After removing NodeTypeStandard, existing standard-typed nodes
-- must be converted so the project filter (n.type='project') finds them.

-- Step 1: For standard nodes that share a name with an existing project
-- node, reroute relations to the project node, then delete the standard node.
WITH conflicts AS (
    SELECT s.id AS std_id, p.id AS proj_id
    FROM nodes s
    JOIN nodes p ON p.name = s.name AND p.type = 'project'
    WHERE s.type = 'standard'
)
UPDATE relations
SET target_id = c.proj_id
FROM conflicts c
WHERE target_id = c.std_id AND target_type = 'node';

DELETE FROM nodes
WHERE type = 'standard'
AND id IN (
    SELECT s.id
    FROM nodes s
    JOIN nodes p ON p.name = s.name AND p.type = 'project'
    WHERE s.type = 'standard'
);

-- Step 2: Remaining standard nodes (no name conflict) — flip type to project.
UPDATE nodes SET type = 'project' WHERE type = 'standard';
