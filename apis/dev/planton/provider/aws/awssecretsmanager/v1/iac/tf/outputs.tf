output "secret_arn_map" {
  description = "Map of secret names to their corresponding ARNs."
  value       = {
    for k, v in aws_secretsmanager_secret.secrets :
    k => v.arn
  }
}
