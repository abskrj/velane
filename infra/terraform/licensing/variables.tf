variable "region" {
  description = "AWS region."
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "EKS cluster name (must match the cluster in infra/terraform/aws-eks)."
  type        = string
}

variable "license_server_image" {
  description = "Container image for the licensing server (e.g. ghcr.io/abskrj/velane-licensing:latest)."
  type        = string
}

variable "private_key_pem" {
  description = "Ed25519 private key PEM — output of 'license keygen'. Store in a secret manager; never commit the actual value."
  type        = string
  sensitive   = true
}

variable "acm_certificate_arn" {
  description = "ACM certificate ARN for TLS. Use the same cert as the main Velane ALB."
  type        = string
  default     = ""
}

variable "base_domain" {
  description = "Root domain (e.g. velane.sh). The server is exposed at license.<base_domain>."
  type        = string
}

variable "db_instance_class" {
  description = "RDS instance class for the licensing database."
  type        = string
  default     = "db.t4g.micro"
}

variable "db_allocated_storage" {
  description = "Storage in GiB for the licensing RDS instance."
  type        = number
  default     = 20
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig for the target EKS cluster."
  type        = string
  default     = "~/.kube/config"
}

variable "kubeconfig_context" {
  description = "Optional kubeconfig context name."
  type        = string
  default     = ""
}

variable "image_pull_policy" {
  description = "Image pull policy."
  type        = string
  default     = "IfNotPresent"
}

variable "tags" {
  description = "Extra tags applied to all AWS resources."
  type        = map(string)
  default     = {}
}
