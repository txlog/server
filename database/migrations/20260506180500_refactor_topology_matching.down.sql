ALTER TABLE topology_patterns
  DROP COLUMN IF EXISTS tag_positions,
  DROP COLUMN IF EXISTS env_group_index,
  DROP COLUMN IF EXISTS svc_group_index,
  DROP COLUMN IF EXISTS seq_group_index;
