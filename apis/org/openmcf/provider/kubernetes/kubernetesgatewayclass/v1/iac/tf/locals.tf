locals {
  # GatewayClass is cluster-scoped; its Kubernetes resource name is the
  # OpenMCF resource name.
  gateway_class_name = var.metadata.name
  controller_name    = var.spec.controller_name

  labels = {
    "app.kubernetes.io/name"       = "gateway-class"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "gateway-class"
  }

  has_parameters_ref = var.spec.parameters_ref != null

  # Build the parametersRef CRD object, including namespace only when provided
  # (upstream requires namespace to be unset for cluster-scoped resources).
  parameters_ref = local.has_parameters_ref ? merge(
    {
      group = var.spec.parameters_ref.group
      kind  = var.spec.parameters_ref.kind
      name  = var.spec.parameters_ref.name
    },
    var.spec.parameters_ref.namespace != null ? { namespace = var.spec.parameters_ref.namespace } : {}
  ) : null
}
