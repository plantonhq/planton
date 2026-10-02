# Enable the Datastream API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one profile must never
# disable the API for every other profile and stream in the project.
resource "google_project_service" "datastream_api" {
  project = local.project_id
  service = "datastream.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The connection profile: exactly one profile type (the spec's rule), with
# public connectivity, a private connection, or an SSH tunnel. Empty
# optional strings and zero ports are sent as null so the provider's
# defaults (the engine ports) apply; passwords and keys are sensitive in the
# provider.
resource "google_datastream_connection_profile" "this" {
  project               = local.project_id
  location              = var.spec.location
  connection_profile_id = local.connection_profile_id
  display_name          = local.display_name

  create_without_validation = var.spec.create_without_validation

  # Google's BigQuery profile has no settings; the spec's bool emits the
  # empty marker block.
  dynamic "bigquery_profile" {
    for_each = var.spec.bigquery_profile ? [true] : []
    content {}
  }

  dynamic "gcs_profile" {
    for_each = var.spec.gcs_profile != null ? [var.spec.gcs_profile] : []
    content {
      bucket    = gcs_profile.value.bucket
      root_path = gcs_profile.value.root_path != "" ? gcs_profile.value.root_path : null
    }
  }

  dynamic "mysql_profile" {
    for_each = var.spec.mysql_profile != null ? [var.spec.mysql_profile] : []
    content {
      hostname                       = mysql_profile.value.hostname
      port                           = mysql_profile.value.port > 0 ? mysql_profile.value.port : null
      username                       = mysql_profile.value.username
      password                       = mysql_profile.value.password != "" ? mysql_profile.value.password : null
      secret_manager_stored_password = mysql_profile.value.secret_manager_stored_password != "" ? mysql_profile.value.secret_manager_stored_password : null

      # Declared (even empty) means sent: an empty ssl_config turns TLS on
      # without certificate verification.
      dynamic "ssl_config" {
        for_each = mysql_profile.value.ssl_config != null ? [mysql_profile.value.ssl_config] : []
        content {
          ca_certificate     = ssl_config.value.ca_certificate != "" ? ssl_config.value.ca_certificate : null
          client_certificate = ssl_config.value.client_certificate != "" ? ssl_config.value.client_certificate : null
          client_key         = ssl_config.value.client_key != "" ? ssl_config.value.client_key : null
        }
      }
    }
  }

  dynamic "postgresql_profile" {
    for_each = var.spec.postgresql_profile != null ? [var.spec.postgresql_profile] : []
    content {
      hostname                       = postgresql_profile.value.hostname
      port                           = postgresql_profile.value.port > 0 ? postgresql_profile.value.port : null
      username                       = postgresql_profile.value.username
      password                       = postgresql_profile.value.password != "" ? postgresql_profile.value.password : null
      secret_manager_stored_password = postgresql_profile.value.secret_manager_stored_password != "" ? postgresql_profile.value.secret_manager_stored_password : null
      database                       = postgresql_profile.value.database

      dynamic "ssl_config" {
        for_each = postgresql_profile.value.ssl_config != null ? [postgresql_profile.value.ssl_config] : []
        content {
          dynamic "server_verification" {
            for_each = ssl_config.value.server_verification != null ? [ssl_config.value.server_verification] : []
            content {
              ca_certificate = server_verification.value.ca_certificate
            }
          }

          dynamic "server_and_client_verification" {
            for_each = ssl_config.value.server_and_client_verification != null ? [ssl_config.value.server_and_client_verification] : []
            content {
              ca_certificate     = server_and_client_verification.value.ca_certificate
              client_certificate = server_and_client_verification.value.client_certificate
              client_key         = server_and_client_verification.value.client_key
            }
          }
        }
      }
    }
  }

  dynamic "oracle_profile" {
    for_each = var.spec.oracle_profile != null ? [var.spec.oracle_profile] : []
    content {
      hostname                       = oracle_profile.value.hostname
      port                           = oracle_profile.value.port > 0 ? oracle_profile.value.port : null
      username                       = oracle_profile.value.username
      password                       = oracle_profile.value.password != "" ? oracle_profile.value.password : null
      secret_manager_stored_password = oracle_profile.value.secret_manager_stored_password != "" ? oracle_profile.value.secret_manager_stored_password : null
      database_service               = oracle_profile.value.database_service
      connection_attributes          = length(oracle_profile.value.connection_attributes) > 0 ? oracle_profile.value.connection_attributes : null
    }
  }

  dynamic "sql_server_profile" {
    for_each = var.spec.sql_server_profile != null ? [var.spec.sql_server_profile] : []
    content {
      hostname                       = sql_server_profile.value.hostname
      port                           = sql_server_profile.value.port > 0 ? sql_server_profile.value.port : null
      username                       = sql_server_profile.value.username
      password                       = sql_server_profile.value.password != "" ? sql_server_profile.value.password : null
      secret_manager_stored_password = sql_server_profile.value.secret_manager_stored_password != "" ? sql_server_profile.value.secret_manager_stored_password : null
      database                       = sql_server_profile.value.database
    }
  }

  dynamic "mongodb_profile" {
    for_each = var.spec.mongodb_profile != null ? [var.spec.mongodb_profile] : []
    content {
      username                       = mongodb_profile.value.username
      password                       = mongodb_profile.value.password != "" ? mongodb_profile.value.password : null
      secret_manager_stored_password = mongodb_profile.value.secret_manager_stored_password != "" ? mongodb_profile.value.secret_manager_stored_password : null
      replica_set                    = mongodb_profile.value.replica_set != "" ? mongodb_profile.value.replica_set : null
      additional_options             = length(mongodb_profile.value.additional_options) > 0 ? mongodb_profile.value.additional_options : null

      dynamic "host_addresses" {
        for_each = mongodb_profile.value.host_addresses
        content {
          hostname = host_addresses.value.hostname
          port     = host_addresses.value.port > 0 ? host_addresses.value.port : null
        }
      }

      # Google's SRV block has no settings; the spec's bool emits the empty
      # marker block.
      dynamic "srv_connection_format" {
        for_each = mongodb_profile.value.srv_connection_format ? [true] : []
        content {}
      }

      dynamic "standard_connection_format" {
        for_each = mongodb_profile.value.standard_connection_format != null ? [mongodb_profile.value.standard_connection_format] : []
        content {
          direct_connection = standard_connection_format.value.direct_connection ? true : null
        }
      }

      dynamic "ssl_config" {
        for_each = mongodb_profile.value.ssl_config != null ? [mongodb_profile.value.ssl_config] : []
        content {
          ca_certificate                   = ssl_config.value.ca_certificate != "" ? ssl_config.value.ca_certificate : null
          client_certificate               = ssl_config.value.client_certificate != "" ? ssl_config.value.client_certificate : null
          client_key                       = ssl_config.value.client_key != "" ? ssl_config.value.client_key : null
          secret_manager_stored_client_key = ssl_config.value.secret_manager_stored_client_key != "" ? ssl_config.value.secret_manager_stored_client_key : null
        }
      }
    }
  }

  # The spec lifts the private-connectivity block's one leaf; the block is
  # sent only when a private connection is named.
  dynamic "private_connectivity" {
    for_each = local.private_connection != null ? [local.private_connection] : []
    content {
      private_connection = private_connectivity.value
    }
  }

  dynamic "forward_ssh_connectivity" {
    for_each = var.spec.forward_ssh_connectivity != null ? [var.spec.forward_ssh_connectivity] : []
    content {
      hostname    = forward_ssh_connectivity.value.hostname
      port        = forward_ssh_connectivity.value.port > 0 ? forward_ssh_connectivity.value.port : null
      username    = forward_ssh_connectivity.value.username
      password    = forward_ssh_connectivity.value.password != "" ? forward_ssh_connectivity.value.password : null
      private_key = forward_ssh_connectivity.value.private_key != "" ? forward_ssh_connectivity.value.private_key : null
    }
  }

  labels = local.final_labels

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.datastream_api]
}
