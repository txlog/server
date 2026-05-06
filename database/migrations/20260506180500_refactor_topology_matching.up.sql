ALTER TABLE topology_patterns
  ADD COLUMN tag_positions JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN env_group_index INT,
  ADD COLUMN svc_group_index INT,
  ADD COLUMN seq_group_index INT;

COMMENT ON COLUMN topology_patterns.tag_positions IS 'JSON array with the ordered list of tags in the template, e.g. [":env",":any",":svc",":any",":seq"]';
COMMENT ON COLUMN topology_patterns.env_group_index IS 'Capture group index of :env in compiled_pattern (1-based). NULL if :env not present.';
COMMENT ON COLUMN topology_patterns.svc_group_index IS 'Capture group index of :svc in compiled_pattern (1-based). NULL if :svc not present.';
COMMENT ON COLUMN topology_patterns.seq_group_index IS 'Capture group index of :seq in compiled_pattern (1-based). NULL if :seq not present.';
