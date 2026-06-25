output "license_server_url" {
  description = "Public URL of the licensing server."
  value       = "https://license.${var.base_domain}"
}

output "rds_endpoint" {
  description = "RDS endpoint for the licensing database."
  value       = aws_db_instance.licensing.address
  sensitive   = true
}

output "rds_database_url" {
  description = "Full Postgres DSN for the licensing server."
  value       = "postgres://licensing:${random_password.db.result}@${aws_db_instance.licensing.address}:5432/licensing"
  sensitive   = true
}
