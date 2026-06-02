variable "kubeconfig_path" {
  description = "Path to kubeconfig for the target cluster."
  type        = string
  default     = "~/.kube/config"
}

variable "kubeconfig_context" {
  description = "Optional kubeconfig context name."
  type        = string
  default     = ""
}

variable "namespace" {
  description = "Kubernetes namespace for Velane workloads."
  type        = string
  default     = "velane"
}

variable "create_namespace" {
  description = "Whether to create the namespace."
  type        = bool
  default     = true
}

variable "image_pull_policy" {
  description = "Image pull policy for all containers."
  type        = string
  default     = "IfNotPresent"
}

variable "control_plane_image" {
  description = "Container image for control-plane."
  type        = string
}

variable "bun_executor_image" {
  description = "Container image for Bun executor."
  type        = string
}

variable "python_executor_image" {
  description = "Container image for Python executor."
  type        = string
}

variable "admin_image" {
  description = "Container image for admin UI."
  type        = string
}

variable "mcp_server_image" {
  description = "Container image for MCP server."
  type        = string
}

variable "control_plane_replicas" {
  description = "Replica count for control-plane."
  type        = number
  default     = 2
}

variable "bun_executor_replicas" {
  description = "Replica count for Bun executor."
  type        = number
  default     = 2
}

variable "python_executor_replicas" {
  description = "Replica count for Python executor."
  type        = number
  default     = 2
}

variable "admin_replicas" {
  description = "Replica count for admin UI."
  type        = number
  default     = 2
}

variable "mcp_server_replicas" {
  description = "Replica count for MCP server."
  type        = number
  default     = 1
}

variable "admin_service_type" {
  description = "Service type for admin UI."
  type        = string
  default     = "LoadBalancer"
}

variable "control_plane_service_type" {
  description = "Service type for control-plane API."
  type        = string
  default     = "LoadBalancer"
}

variable "mcp_server_service_type" {
  description = "Service type for MCP server."
  type        = string
  default     = "LoadBalancer"
}

variable "database_url" {
  description = "Postgres DSN for control-plane."
  type        = string
  sensitive   = true
}

variable "redis_url" {
  description = "Redis address for control-plane."
  type        = string
}

variable "encryption_key" {
  description = "64-char hex AES key used to encrypt secrets."
  type        = string
  sensitive   = true
}

variable "jwt_private_key_pem" {
  description = "RS256 private key PEM for session tokens."
  type        = string
  sensitive   = true
}

variable "worker_count" {
  description = "Async worker concurrency for control-plane."
  type        = number
  default     = 5
}

variable "executor_type" {
  description = "Executor type for control-plane (process or firecracker)."
  type        = string
  default     = "process"
}

variable "bootstrap_email" {
  description = "Optional bootstrap admin email (first boot only)."
  type        = string
  default     = ""
}

variable "bootstrap_password" {
  description = "Optional bootstrap admin password (first boot only)."
  type        = string
  default     = ""
  sensitive   = true
}

variable "bootstrap_tenant" {
  description = "Optional bootstrap tenant slug (first boot only)."
  type        = string
  default     = "default"
}

variable "nango_internal_url" {
  description = "Internal Nango API URL, if Nango is deployed separately."
  type        = string
  default     = "http://nango:3003"
}

variable "nango_connect_url" {
  description = "Browser-facing Nango Connect URL."
  type        = string
  default     = ""
}

variable "nango_api_url" {
  description = "Browser-facing Nango API URL."
  type        = string
  default     = ""
}

variable "nango_secret_key" {
  description = "Nango secret key."
  type        = string
  default     = ""
  sensitive   = true
}

variable "nango_public_key" {
  description = "Nango public key for frontend connect."
  type        = string
  default     = ""
  sensitive   = true
}

variable "nango_webhook_secret" {
  description = "Nango webhook signing secret."
  type        = string
  default     = ""
  sensitive   = true
}

variable "clickhouse_dsn" {
  description = "Optional ClickHouse DSN."
  type        = string
  default     = ""
}

variable "logs_bucket" {
  description = "Optional logs bucket name."
  type        = string
  default     = ""
}

variable "replay_bucket" {
  description = "Optional replay bucket name."
  type        = string
  default     = ""
}

# ==================== Ingress ====================

variable "enable_ingress" {
  description = "Create an Ingress resource for subdomain-based routing."
  type        = bool
  default     = true
}

variable "ingress_class_name" {
  description = "Ingress class name. Use 'alb' for AWS Application Load Balancer (requires aws-load-balancer-controller installed on the cluster), 'nginx' for NGINX Ingress Controller, or your cloud's equivalent."
  type        = string
  default     = "nginx"
}

variable "base_domain" {
  description = "Base domain for the ingress (e.g. 'yourdomain.com'). Subdomains will be created under this."
  type        = string
  default     = "example.com"
}

variable "admin_subdomain" {
  description = "Subdomain prefix for the admin UI."
  type        = string
  default     = "admin"
}

variable "api_subdomain" {
  description = "Subdomain prefix for the control-plane API."
  type        = string
  default     = "api"
}

variable "mcp_subdomain" {
  description = "Subdomain prefix for the MCP server."
  type        = string
  default     = "mcp"
}

variable "nango_connect_subdomain" {
  description = "Subdomain prefix for the Nango Connect UI."
  type        = string
  default     = "connect"
}

variable "nango_api_subdomain" {
  description = "Subdomain prefix for the Nango API."
  type        = string
  default     = "nango"
}

variable "ingress_annotations" {
  description = "Extra annotations to put on the Ingress (useful for ALB: alb.ingress.kubernetes.io/scheme, alb.ingress.kubernetes.io/target-type, etc.)."
  type        = map(string)
  default     = {}
}

# ==================== Nango (in-cluster) ====================

variable "deploy_nango" {
  description = "Deploy Nango server inside the cluster (recommended for full integrations support)."
  type        = bool
  default     = true
}

variable "nango_image" {
  description = "Nango server container image."
  type        = string
  default     = "nangohq/nango-server:hosted"
}

variable "nango_replicas" {
  description = "Replica count for Nango."
  type        = number
  default     = 1
}

variable "nango_encryption_key" {
  description = "Encryption key for Nango (32 bytes base64). If empty, a default dev key is used (not for production)."
  type        = string
  default     = "6ProXeOGZC0HLT+Kd+2TfneHJmyqcMviCkH8aqwdF4I="
  sensitive   = true
}

variable "nango_secret_key" {
  description = "NANGO_SECRET_KEY for this environment."
  type        = string
  default     = ""
  sensitive   = true
}

variable "nango_public_key" {
  description = "NANGO_PUBLIC_KEY (usually same as secret for self-hosted)."
  type        = string
  default     = ""
  sensitive   = true
}

variable "nango_webhook_secret" {
  description = "Signing secret for Nango webhooks."
  type        = string
  default     = ""
  sensitive   = true
}

# ==================== Nango Database ====================

variable "nango_database_url" {
  description = "Full Postgres DSN for Nango (e.g. postgres://user:pass@host:5432/nango?sslmode=require). If empty and create_nango_database=true, a separate 'nango' database will be created on the same Postgres server used by Velane."
  type        = string
  default     = ""
  sensitive   = true
}

variable "create_nango_database" {
  description = "If nango_database_url is not provided, automatically create a dedicated 'nango' database on the Velane Postgres instance (one-time Job)."
  type        = bool
  default     = true
}

variable "nango_db_name" {
  description = "Name of the separate database to use/create for Nango when nango_database_url is not supplied."
  type        = string
  default     = "nango"
}
