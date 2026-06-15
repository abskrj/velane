ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS runtime_limits JSONB NOT NULL DEFAULT '{
    "max_timeout_ms": 900000,
    "max_memory_mb": 2048,
    "max_cpu_percent": 100
  }'::jsonb;
