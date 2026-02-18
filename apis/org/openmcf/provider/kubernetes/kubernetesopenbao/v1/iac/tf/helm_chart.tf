# Deploy OpenBao using the official Helm chart
resource "helm_release" "openbao" {
  name       = var.metadata.name
  namespace  = local.namespace
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.helm_chart_version

  # Wait for namespace to be created if create_namespace is true
  depends_on = [kubernetes_namespace.openbao_namespace]

  # Global configuration
  set {
    name  = "fullnameOverride"
    value = var.metadata.name
  }

  set {
    name  = "global.enabled"
    value = "true"
  }

  set {
    name  = "global.tlsDisable"
    value = tostring(!local.tls_enabled)
  }

  # Server data storage
  set {
    name  = "server.dataStorage.enabled"
    value = "true"
  }

  set {
    name  = "server.dataStorage.size"
    value = var.spec.server_container.data_storage_size
  }

  # Server resources
  set {
    name  = "server.resources.requests.cpu"
    value = var.spec.server_container.resources.requests.cpu
  }

  set {
    name  = "server.resources.requests.memory"
    value = var.spec.server_container.resources.requests.memory
  }

  set {
    name  = "server.resources.limits.cpu"
    value = var.spec.server_container.resources.limits.cpu
  }

  set {
    name  = "server.resources.limits.memory"
    value = var.spec.server_container.resources.limits.memory
  }

  # HA mode configuration
  set {
    name  = "server.ha.enabled"
    value = tostring(local.ha_enabled)
  }

  dynamic "set" {
    for_each = local.ha_enabled ? [1] : []
    content {
      name  = "server.ha.replicas"
      value = tostring(local.ha_replicas)
    }
  }

  dynamic "set" {
    for_each = local.ha_enabled ? [1] : []
    content {
      name  = "server.ha.raft.enabled"
      value = "true"
    }
  }

  dynamic "set" {
    for_each = local.ha_enabled ? [1] : []
    content {
      name  = "server.ha.raft.setNodeId"
      value = "true"
    }
  }

  # Standalone mode (when HA is disabled)
  set {
    name  = "server.standalone.enabled"
    value = tostring(!local.ha_enabled)
  }

  # UI configuration
  set {
    name  = "ui.enabled"
    value = tostring(local.ui_enabled)
  }

  # Injector configuration
  set {
    name  = "injector.enabled"
    value = tostring(local.injector_enabled)
  }

  dynamic "set" {
    for_each = local.injector_enabled ? [1] : []
    content {
      name  = "injector.replicas"
      value = tostring(local.injector_replicas)
    }
  }

}
