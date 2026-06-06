# Main Python script (main.py)
resource "kubernetes_config_map" "main_py" {
  metadata {
    name      = local.main_py_configmap_name
    namespace = var.spec.create_namespace ? kubernetes_namespace.this[0].metadata[0].name : local.namespace
    labels    = local.final_labels
  }

  data = {
    "main.py" = var.spec.load_test.main_py_content
  }
}

# Additional library files (lib_files_content)
resource "kubernetes_config_map" "lib_files" {
  metadata {
    name      = local.lib_files_configmap_name
    namespace = var.spec.create_namespace ? kubernetes_namespace.this[0].metadata[0].name : local.namespace
    labels    = local.final_labels
  }

  # Because lib_files_content is a map of filename->content,
  # we can directly assign it to the data block.
  data = var.spec.load_test.lib_files_content
}

# Merge base helm values with user-provided overrides
locals {
  base_helm_values = {
    fullnameOverride = local.kube_service_name

    master = {
      replicas = var.spec.master_container.replicas
      resources = {
        requests = {
          cpu    = try(var.spec.master_container.resources.requests.cpu, null)
          memory = try(var.spec.master_container.resources.requests.memory, null)
        }
        limits = {
          cpu    = try(var.spec.master_container.resources.limits.cpu, null)
          memory = try(var.spec.master_container.resources.limits.memory, null)
        }
      }
    }

    worker = {
      replicas = var.spec.worker_container.replicas
      resources = {
        requests = {
          cpu    = try(var.spec.worker_container.resources.requests.cpu, null)
          memory = try(var.spec.worker_container.resources.requests.memory, null)
        }
        limits = {
          cpu    = try(var.spec.worker_container.resources.limits.cpu, null)
          memory = try(var.spec.worker_container.resources.limits.memory, null)
        }
      }
    }

    loadtest = {
      name                        = var.spec.load_test.name
      locust_locustfile_configmap = local.main_py_configmap_name
      locust_lib_configmap        = local.lib_files_configmap_name
    }
  }

  # Merge user-provided helm_values (if any) with base_helm_values
  merged_helm_values = merge(
    local.base_helm_values,
    try(var.spec.helm_values, {})
  )
}

resource "helm_release" "this" {
  name       = local.resource_id
  repository = "https://charts.deliveryhero.io"
  chart      = "locust"
  # set the chart version you want to deploy
  version          = "0.31.5"
  namespace        = var.spec.create_namespace ? kubernetes_namespace.this[0].metadata[0].name : local.namespace
  create_namespace = false

  values = [
    yamlencode(local.merged_helm_values)
  ]

  depends_on = [
    kubernetes_config_map.main_py,
    kubernetes_config_map.lib_files
  ]
}
