# Local values and computed configuration

locals {
  # Standard OpenMCF labels merged with user-provided labels.
  standard_labels = {
    "managed-by"    = "openmcf"
    "resource"      = var.metadata.name
    "resource-kind" = "KubernetesService"
  }

  labels = merge(local.standard_labels, var.spec.labels)

  # Annotations from user-provided values.
  annotations = var.spec.annotations

  # Map protobuf enum string values to Kubernetes API strings.
  service_type_map = {
    "cluster_ip"                           = "ClusterIP"
    "node_port"                            = "NodePort"
    "load_balancer"                        = "LoadBalancer"
    "external_name"                        = "ExternalName"
    "kubernetes_service_type_unspecified"   = "ClusterIP"
  }
  service_type = lookup(local.service_type_map, var.spec.type, "ClusterIP")

  external_traffic_policy_map = {
    "cluster"                                   = "Cluster"
    "local"                                     = "Local"
    "external_traffic_policy_unspecified"        = "Cluster"
  }
  external_traffic_policy = lookup(local.external_traffic_policy_map, var.spec.external_traffic_policy, "Cluster")

  session_affinity_map = {
    "none"                            = "None"
    "client_ip"                       = "ClientIP"
    "session_affinity_unspecified"     = "None"
  }
  session_affinity = lookup(local.session_affinity_map, var.spec.session_affinity, "None")

  # Whether this is an ExternalName service.
  is_external_name = local.service_type == "ExternalName"

  # Whether this is a NodePort or LoadBalancer service (external traffic policy relevant).
  is_external = local.service_type == "NodePort" || local.service_type == "LoadBalancer"

  # Internal DNS name for the service.
  internal_dns_name = "${var.spec.name}.${var.spec.namespace}.svc.cluster.local"
}
