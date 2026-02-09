resource "random_password" "admin_password" {
  length      = 12
  special     = true
  numeric     = true
  upper       = true
  lower       = true
  min_special = 3
  min_numeric = 2
  min_upper   = 2
  min_lower   = 2
}

locals {
  admin_password_b64 = base64encode(random_password.admin_password.result)
}

resource "kubernetes_secret" "redis_admin_secret" {
  metadata {
    name      = local.password_secret_name
    namespace = local.namespace
  }

  data = {
    "password" = local.admin_password_b64
  }

  depends_on = [kubernetes_namespace.redis_namespace]
}

output "redis_username" {
  description = "The Redis username"
  value       = "default"
}

output "redis_password_secret_name" {
  description = "The name of the Kubernetes secret containing the admin password"
  value       = kubernetes_secret.redis_admin_secret.metadata[0].name
}

output "redis_password_secret_key" {
  description = "The key in the secret that stores the password"
  value       = "password"
}
