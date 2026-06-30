# Local values and computed configuration

locals {
  # Build combined labels
  standard_labels = {
    "managed-by"    = "planton"
    "resource"      = var.metadata.name
    "resource-kind" = "KubernetesNamespace"
  }

  # Pod security standard label
  pss_label = var.spec.pod_security_standard != null && var.spec.pod_security_standard != "pod_security_standard_unspecified" ? {
    "pod-security.kubernetes.io/enforce" = var.spec.pod_security_standard
  } : {}

  labels = merge(local.standard_labels, var.spec.labels, local.pss_label)

  # Build annotations with service mesh injection.
  # try() is used for safe nested access because HCL does not short-circuit &&.
  mesh_annotations = try(var.spec.service_mesh_config.enabled, false) ? (
    try(var.spec.service_mesh_config.mesh_type, "") == "istio" ? (
      try(var.spec.service_mesh_config.revision_tag, null) != null ? {
        "istio.io/rev" = var.spec.service_mesh_config.revision_tag
      } : {
        "istio-injection" = "enabled"
      }
    ) : try(var.spec.service_mesh_config.mesh_type, "") == "linkerd" ? {
      "linkerd.io/inject" = "enabled"
    } : try(var.spec.service_mesh_config.mesh_type, "") == "consul" ? {
      "consul.hashicorp.com/connect-inject" = "true"
    } : {}
  ) : {}

  annotations = merge(var.spec.annotations, local.mesh_annotations)

  # Resource Quota configuration
  resource_quota_enabled = var.spec.resource_profile != null

  resource_quota_preset = try(var.spec.resource_profile.preset, null) != null ? {
    "small" = {
      cpu_requests    = "2"
      cpu_limits      = "4"
      memory_requests = "4Gi"
      memory_limits   = "8Gi"
      pods            = 20
      services        = 10
      configmaps      = 50
      secrets         = 50
      pvcs            = 5
      load_balancers  = 2
    }
    "medium" = {
      cpu_requests    = "4"
      cpu_limits      = "8"
      memory_requests = "8Gi"
      memory_limits   = "16Gi"
      pods            = 50
      services        = 20
      configmaps      = 100
      secrets         = 100
      pvcs            = 10
      load_balancers  = 3
    }
    "large" = {
      cpu_requests    = "8"
      cpu_limits      = "16"
      memory_requests = "16Gi"
      memory_limits   = "32Gi"
      pods            = 100
      services        = 40
      configmaps      = 200
      secrets         = 200
      pvcs            = 20
      load_balancers  = 5
    }
    "xlarge" = {
      cpu_requests    = "16"
      cpu_limits      = "32"
      memory_requests = "32Gi"
      memory_limits   = "64Gi"
      pods            = 200
      services        = 80
      configmaps      = 400
      secrets         = 400
      pvcs            = 40
      load_balancers  = 10
    }
  }[try(var.spec.resource_profile.preset, "")] : null

  # Compute quota values
  quota_config = local.resource_quota_enabled ? (
    try(var.spec.resource_profile.preset, null) != null ? local.resource_quota_preset : {
      cpu_requests    = try(var.spec.resource_profile.custom.cpu.requests, null)
      cpu_limits      = try(var.spec.resource_profile.custom.cpu.limits, null)
      memory_requests = try(var.spec.resource_profile.custom.memory.requests, null)
      memory_limits   = try(var.spec.resource_profile.custom.memory.limits, null)
      pods            = try(var.spec.resource_profile.custom.object_counts.pods, null)
      services        = try(var.spec.resource_profile.custom.object_counts.services, null)
      configmaps      = try(var.spec.resource_profile.custom.object_counts.configmaps, null)
      secrets         = try(var.spec.resource_profile.custom.object_counts.secrets, null)
      pvcs            = try(var.spec.resource_profile.custom.object_counts.persistent_volume_claims, null)
      load_balancers  = try(var.spec.resource_profile.custom.object_counts.load_balancers, null)
    }
  ) : null

  # Build ResourceQuota hard limits
  resource_quota_hard = local.resource_quota_enabled ? merge(
    local.quota_config.cpu_requests != null ? { "requests.cpu" = local.quota_config.cpu_requests } : {},
    local.quota_config.cpu_limits != null ? { "limits.cpu" = local.quota_config.cpu_limits } : {},
    local.quota_config.memory_requests != null ? { "requests.memory" = local.quota_config.memory_requests } : {},
    local.quota_config.memory_limits != null ? { "limits.memory" = local.quota_config.memory_limits } : {},
    local.quota_config.pods != null ? { "count/pods" = tostring(local.quota_config.pods) } : {},
    local.quota_config.services != null ? { "count/services" = tostring(local.quota_config.services) } : {},
    local.quota_config.configmaps != null ? { "count/configmaps" = tostring(local.quota_config.configmaps) } : {},
    local.quota_config.secrets != null ? { "count/secrets" = tostring(local.quota_config.secrets) } : {},
    local.quota_config.pvcs != null ? { "count/persistentvolumeclaims" = tostring(local.quota_config.pvcs) } : {},
    local.quota_config.load_balancers != null ? { "count/services.loadbalancers" = tostring(local.quota_config.load_balancers) } : {}
  ) : {}

  # Limit Range configuration
  limit_range_enabled = try(var.spec.resource_profile.custom.default_limits, null) != null

  default_requests = local.limit_range_enabled ? {
    cpu    = var.spec.resource_profile.custom.default_limits.default_cpu_request
    memory = var.spec.resource_profile.custom.default_limits.default_memory_request
  } : {}

  default_limits = local.limit_range_enabled ? {
    cpu    = var.spec.resource_profile.custom.default_limits.default_cpu_limit
    memory = var.spec.resource_profile.custom.default_limits.default_memory_limit
  } : {}

  # Network policy configuration
  isolate_ingress            = var.spec.network_config != null ? var.spec.network_config.isolate_ingress : false
  restrict_egress            = var.spec.network_config != null ? var.spec.network_config.restrict_egress : false
  allowed_ingress_namespaces = var.spec.network_config != null ? var.spec.network_config.allowed_ingress_namespaces : []
  allowed_egress_cidrs       = var.spec.network_config != null ? var.spec.network_config.allowed_egress_cidrs : []

  # Service mesh configuration
  service_mesh_enabled = try(var.spec.service_mesh_config.enabled, false)
  service_mesh_type    = try(var.spec.service_mesh_config.mesh_type, "")

  # Pod security standard
  pod_security_standard = var.spec.pod_security_standard != null && var.spec.pod_security_standard != "pod_security_standard_unspecified" ? (
    var.spec.pod_security_standard
  ) : ""
}

