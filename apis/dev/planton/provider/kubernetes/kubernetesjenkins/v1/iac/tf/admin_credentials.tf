# 1) Generate a random password
resource "random_password" "jenkins_admin_password" {
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

# 2) Base64-encode the generated password
locals {
  jenkins_admin_username            = "admin"
  jenkins_admin_password_secret_key = "admin-password"
  jenkins_admin_password_b64        = base64encode(random_password.jenkins_admin_password.result)
}

# 3) Create or update the K8s secret containing the Jenkins admin password
resource "kubernetes_secret" "jenkins_admin_secret" {
  metadata {
    # Use computed name to avoid conflicts when multiple instances share a namespace
    name      = local.admin_credentials_secret_name
    namespace = local.namespace
  }

  data = {
    (local.jenkins_admin_password_secret_key) = local.jenkins_admin_password_b64
  }
}

# 4) Output the admin credentials
output "jenkins_admin_username" {
  description = "The Jenkins admin username"
  value       = local.jenkins_admin_username
}

output "jenkins_admin_password_secret_name" {
  description = "The name of the Kubernetes secret storing the Jenkins admin password"
  value       = kubernetes_secret.jenkins_admin_secret.metadata[0].name
}

output "jenkins_admin_password_secret_key" {
  description = "The key in the Jenkins admin secret that stores the password"
  value       = local.jenkins_admin_password_secret_key
}
