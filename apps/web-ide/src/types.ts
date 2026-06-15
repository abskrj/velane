export interface Snippet {
  id: string
  tenant_id: string
  name: string
  slug: string
  language: 'bun' | 'python'
  created_at: string
  created_by: string
}

export interface SnippetEnvironment {
  snippet_id: string
  env: 'dev' | 'staging' | 'prod'
  active_version_id: string | null
  min_instances: number
  canary_version_id: string | null
  canary_pct: number
}

export interface SnippetVersion {
  id: string
  snippet_id: string
  version_number: number
  code: string
  input_schema: string
  output_schema: string
  timeout_ms: number
  max_memory_mb: number
  max_cpu_percent: number
  status: 'draft' | 'published' | 'archived'
  created_at: string
  created_by: string
}

export interface InvocationResult {
  output: unknown
  invocation_id: string
  duration_ms: number
  status: string
  error: string
  stderr: string
}

export interface CreateSnippetInput {
  name: string
  language: 'bun' | 'python'
}
