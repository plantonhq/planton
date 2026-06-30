resource "random_password" "clickhouse_password" {
  length  = 20
  special = true
  numeric = true
  upper   = true
  lower   = true

  min_special = 2
  min_numeric = 3
  min_upper   = 3
  min_lower   = 3

  # IMPORTANT: Only use URL-safe special characters to avoid encoding issues.
  # Characters like +, =, /, &, ?, # cause problems when passwords are used in
  # connection strings like: tcp://host:port/?password=XXX
  # The + character is particularly problematic as it's decoded as a space.
  # See: https://github.com/Altinity/clickhouse-operator/issues/1883
  override_special = "-_"
}

resource "kubernetes_secret_v1" "clickhouse_password" {
  metadata {
    # Use computed name to avoid conflicts when multiple instances share a namespace
    name      = local.password_secret_name
    namespace = local.namespace
  }

  data = {
    "admin-password" = base64encode(random_password.clickhouse_password.result)
  }

  type = "Opaque"
}
