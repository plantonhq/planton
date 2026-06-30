locals {
  # Determine port - use default based on engine if not specified
  datastore_port = coalesce(
    var.spec.datastore.port,
    var.spec.datastore.engine == "postgres" ? 5432 : 3306
  )

  # Build connection string options based on engine and is_secure flag
  conn_options = var.spec.datastore.engine == "postgres" ? (
    var.spec.datastore.is_secure ? "?sslmode=require" : ""
  ) : (
    # MySQL requires parseTime=true for proper time handling
    var.spec.datastore.is_secure ? "?parseTime=true&tls=true" : "?parseTime=true"
  )

  # Check if using secret reference for password
  use_secret_ref = try(var.spec.datastore.password.secret_ref, null) != null

  # Construct URI - if using secret ref, use environment variable placeholder
  datastore_uri = local.use_secret_ref ? (
    "${var.spec.datastore.engine}://${var.spec.datastore.username}:$(OPENFGA_DATASTORE_PASSWORD)@${var.spec.datastore.host}:${local.datastore_port}/${var.spec.datastore.database}${local.conn_options}"
  ) : (
    "${var.spec.datastore.engine}://${var.spec.datastore.username}:${var.spec.datastore.password.string_value}@${var.spec.datastore.host}:${local.datastore_port}/${var.spec.datastore.database}${local.conn_options}"
  )
}

resource "helm_release" "openfga_helm_chart" {
  name             = local.resource_id
  repository       = "https://openfga.github.io/helm-charts"
  chart            = "openfga"
  version          = "0.2.12"
  namespace        = local.namespace
  create_namespace = false

  values = [
    yamlencode(merge(
      {
        fullnameOverride = local.kube_service_name
        replicaCount     = var.spec.container.replicas
        datastore = {
          engine = var.spec.datastore.engine
          uri    = local.datastore_uri
        }
        resources = {
          requests = {
            cpu    = try(var.spec.container.resources.requests.cpu, null)
            memory = try(var.spec.container.resources.requests.memory, null)
          }
          limits = {
            cpu    = try(var.spec.container.resources.limits.cpu, null)
            memory = try(var.spec.container.resources.limits.memory, null)
          }
        }
      },
      # Add extraEnvVars to inject the password from secret if using secret reference
      local.use_secret_ref ? {
        extraEnvVars = [
          {
            name = "OPENFGA_DATASTORE_PASSWORD"
            valueFrom = {
              secretKeyRef = {
                name = var.spec.datastore.password.secret_ref.name
                key  = var.spec.datastore.password.secret_ref.key
              }
            }
          }
        ]
      } : {}
    ))
  ]
}
