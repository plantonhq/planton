# Enable the BigQuery Connection API first so a fresh project works on the
# first deploy. disable_on_destroy is false: tearing down one connection
# must never disable the API for every other connection in the project.
resource "google_project_service" "bigqueryconnection_api" {
  project = local.project_id
  service = "bigqueryconnection.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The connection. Exactly one arm block is emitted -- the spec's CEL rule
# guarantees exactly one is declared. project, location, connection_id,
# and the connector id are immutable; everything else updates in place.
resource "google_bigquery_connection" "this" {
  project       = local.project_id
  location      = local.location
  connection_id = local.connection_id
  friendly_name = local.friendly_name
  description   = local.description
  kms_key_name  = local.kms_key_name

  # The spec lifts the access_role wrapper's one leaf.
  dynamic "aws" {
    for_each = var.spec.aws != null ? [var.spec.aws] : []
    content {
      access_role {
        iam_role_id = aws.value.iam_role_id
      }
    }
  }

  dynamic "azure" {
    for_each = var.spec.azure != null ? [var.spec.azure] : []
    content {
      customer_tenant_id              = azure.value.customer_tenant_id
      federated_application_client_id = azure.value.federated_application_client_id != "" ? azure.value.federated_application_client_id : null
    }
  }

  # Google's arm has no settings -- the spec's bool emits the empty marker
  # block, and Google answers with the service account it created.
  dynamic "cloud_resource" {
    for_each = var.spec.cloud_resource ? [true] : []
    content {}
  }

  # The flags are sent only when true, so Google's defaults stay in charge
  # otherwise.
  dynamic "cloud_spanner" {
    for_each = var.spec.cloud_spanner != null ? [var.spec.cloud_spanner] : []
    content {
      database        = cloud_spanner.value.database
      database_role   = cloud_spanner.value.database_role != "" ? cloud_spanner.value.database_role : null
      use_parallelism = cloud_spanner.value.use_parallelism ? true : null
      use_data_boost  = cloud_spanner.value.use_data_boost ? true : null
      max_parallelism = cloud_spanner.value.max_parallelism > 0 ? cloud_spanner.value.max_parallelism : null
    }
  }

  dynamic "cloud_sql" {
    for_each = var.spec.cloud_sql != null ? [var.spec.cloud_sql] : []
    content {
      instance_id = cloud_sql.value.instance_id
      database    = cloud_sql.value.database
      type        = cloud_sql.value.type

      credential {
        username = cloud_sql.value.credential.username
        password = cloud_sql.value.credential.password
      }
    }
  }

  # The spec lifts the authentication, password, endpoint, and network
  # wrappers; each nested block is sent only when its value is declared.
  dynamic "configuration" {
    for_each = var.spec.configuration != null ? [var.spec.configuration] : []
    content {
      connector_id = configuration.value.connector_id

      asset {
        database              = configuration.value.asset.database != "" ? configuration.value.asset.database : null
        google_cloud_resource = configuration.value.asset.google_cloud_resource != "" ? configuration.value.asset.google_cloud_resource : null
      }

      dynamic "authentication" {
        for_each = configuration.value.username_password != null ? [configuration.value.username_password] : []
        content {
          username_password {
            username = authentication.value.username
            password {
              plaintext = authentication.value.password
            }
          }
        }
      }

      dynamic "endpoint" {
        for_each = configuration.value.host_port != "" ? [configuration.value.host_port] : []
        content {
          host_port = endpoint.value
        }
      }

      dynamic "network" {
        for_each = configuration.value.network_attachment != "" ? [configuration.value.network_attachment] : []
        content {
          private_service_connect {
            network_attachment = network.value
          }
        }
      }
    }
  }

  # The spec lifts both one-leaf wrappers; each is sent only when set.
  dynamic "spark" {
    for_each = var.spec.spark != null ? [var.spec.spark] : []
    content {
      dynamic "metastore_service_config" {
        for_each = spark.value.metastore_service != "" ? [spark.value.metastore_service] : []
        content {
          metastore_service = metastore_service_config.value
        }
      }

      dynamic "spark_history_server_config" {
        for_each = spark.value.history_server_dataproc_cluster != "" ? [spark.value.history_server_dataproc_cluster] : []
        content {
          dataproc_cluster = spark_history_server_config.value
        }
      }
    }
  }

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.bigqueryconnection_api]
}
