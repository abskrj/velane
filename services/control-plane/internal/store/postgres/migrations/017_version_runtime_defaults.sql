-- Default runtime limits for new workflow versions (when columns use DB defaults).
ALTER TABLE snippet_versions
  ALTER COLUMN timeout_ms SET DEFAULT 60000,
  ALTER COLUMN max_memory_mb SET DEFAULT 200,
  ALTER COLUMN max_cpu_percent SET DEFAULT 10;
