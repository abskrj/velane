output "namespace" {
  description = "Namespace where Velane is deployed."
  value       = local.namespace
}

output "admin_service_name" {
  description = "Kubernetes service name for admin."
  value       = kubernetes_service_v1.admin.metadata[0].name
}

output "control_plane_service_name" {
  description = "Kubernetes service name for control-plane."
  value       = kubernetes_service_v1.control_plane.metadata[0].name
}

output "mcp_server_service_name" {
  description = "Kubernetes service name for mcp-server."
  value       = kubernetes_service_v1.mcp_server.metadata[0].name
}

output "admin_service_type" {
  description = "Service type used for admin."
  value       = kubernetes_service_v1.admin.spec[0].type
}

output "admin_load_balancer_hostname" {
  description = "External hostname for the admin LoadBalancer service, when assigned."
  value       = try(kubernetes_service_v1.admin.status[0].load_balancer[0].ingress[0].hostname, null)
}

output "control_plane_service_type" {
  description = "Service type used for control-plane."
  value       = kubernetes_service_v1.control_plane.spec[0].type
}

output "control_plane_load_balancer_hostname" {
  description = "External hostname for the control-plane LoadBalancer service, when assigned."
  value       = try(kubernetes_service_v1.control_plane.status[0].load_balancer[0].ingress[0].hostname, null)
}

output "mcp_server_service_type" {
  description = "Service type used for mcp-server."
  value       = kubernetes_service_v1.mcp_server.spec[0].type
}

output "mcp_server_load_balancer_hostname" {
  description = "External hostname for the MCP server LoadBalancer service, when assigned."
  value       = try(kubernetes_service_v1.mcp_server.status[0].load_balancer[0].ingress[0].hostname, null)
}

output "ingress_name" {
  description = "Name of the created Ingress (if enabled)."
  value       = var.enable_ingress ? kubernetes_ingress_v1.velane[0].metadata[0].name : null
}

output "nango_deployed" {
  description = "Whether an in-cluster Nango deployment was created."
  value       = var.deploy_nango
}

output "effective_nango_database_url" {
  description = "The Nango database URL that will actually be used (computed when not explicitly provided)."
  value       = local.effective_nango_database_url
  sensitive   = true
}
