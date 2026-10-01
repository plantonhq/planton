# The Cloud Deploy API. disable_on_destroy is false: tearing down one
# target must never disable the API for every pipeline in the project.
resource "google_project_service" "clouddeploy_api" {
  project = local.project_id
  service = "clouddeploy.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The deployment target. Exactly one of gke, anthos_cluster, run,
# multi_target, or custom_target is set; location and the target ID are
# immutable, everything else updates in place. Optional fields are sent
# only when set so Cloud Deploy's defaults apply.
resource "google_clouddeploy_target" "this" {
  project           = local.project_id
  location          = var.spec.location
  name              = local.target_id
  description       = var.spec.description != "" ? var.spec.description : null
  labels            = local.final_labels
  annotations       = length(var.spec.annotations) > 0 ? var.spec.annotations : null
  require_approval  = var.spec.require_approval ? true : null
  deploy_parameters = length(var.spec.deploy_parameters) > 0 ? var.spec.deploy_parameters : null
  deletion_policy   = local.deletion_policy

  dynamic "gke" {
    for_each = var.spec.gke != null ? [var.spec.gke] : []
    content {
      cluster      = gke.value.cluster != "" ? gke.value.cluster : null
      internal_ip  = gke.value.internal_ip ? true : null
      dns_endpoint = gke.value.dns_endpoint ? true : null
      proxy_url    = gke.value.proxy_url != "" ? gke.value.proxy_url : null
    }
  }

  dynamic "anthos_cluster" {
    for_each = var.spec.anthos_cluster != null ? [var.spec.anthos_cluster] : []
    content {
      membership = anthos_cluster.value.membership != "" ? anthos_cluster.value.membership : null
    }
  }

  dynamic "run" {
    for_each = var.spec.run != null ? [var.spec.run] : []
    content {
      location = run.value.location
    }
  }

  dynamic "multi_target" {
    for_each = var.spec.multi_target != null ? [var.spec.multi_target] : []
    content {
      target_ids = multi_target.value.target_ids
    }
  }

  dynamic "custom_target" {
    for_each = var.spec.custom_target != null ? [var.spec.custom_target] : []
    content {
      custom_target_type = custom_target.value.custom_target_type
    }
  }

  # Keyed by entity_id in Google's API (the provider models the map as a
  # set), so order never matters.
  dynamic "associated_entities" {
    for_each = var.spec.associated_entities
    content {
      entity_id = associated_entities.value.entity_id

      dynamic "gke_clusters" {
        for_each = associated_entities.value.gke_clusters
        content {
          cluster     = gke_clusters.value.cluster != "" ? gke_clusters.value.cluster : null
          internal_ip = gke_clusters.value.internal_ip ? true : null
          proxy_url   = gke_clusters.value.proxy_url != "" ? gke_clusters.value.proxy_url : null
        }
      }

      dynamic "anthos_clusters" {
        for_each = associated_entities.value.anthos_clusters
        content {
          membership = anthos_clusters.value.membership != "" ? anthos_clusters.value.membership : null
        }
      }
    }
  }

  # Optional and computed: with none declared, Cloud Deploy reports its
  # defaults and the plan stays clean.
  dynamic "execution_configs" {
    for_each = var.spec.execution_configs
    content {
      usages            = execution_configs.value.usages
      worker_pool       = execution_configs.value.worker_pool != "" ? execution_configs.value.worker_pool : null
      service_account   = execution_configs.value.service_account != "" ? execution_configs.value.service_account : null
      artifact_storage  = execution_configs.value.artifact_storage != "" ? execution_configs.value.artifact_storage : null
      execution_timeout = execution_configs.value.execution_timeout != "" ? execution_configs.value.execution_timeout : null
      verbose           = execution_configs.value.verbose ? true : null

      dynamic "default_pool" {
        for_each = execution_configs.value.default_pool != null ? [execution_configs.value.default_pool] : []
        content {
          service_account  = default_pool.value.service_account != "" ? default_pool.value.service_account : null
          artifact_storage = default_pool.value.artifact_storage != "" ? default_pool.value.artifact_storage : null
        }
      }

      dynamic "private_pool" {
        for_each = execution_configs.value.private_pool != null ? [execution_configs.value.private_pool] : []
        content {
          worker_pool      = private_pool.value.worker_pool
          service_account  = private_pool.value.service_account != "" ? private_pool.value.service_account : null
          artifact_storage = private_pool.value.artifact_storage != "" ? private_pool.value.artifact_storage : null
        }
      }
    }
  }

  depends_on = [google_project_service.clouddeploy_api]
}
