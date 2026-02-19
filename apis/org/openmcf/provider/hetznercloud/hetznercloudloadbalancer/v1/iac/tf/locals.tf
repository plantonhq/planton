locals {
  lb_name = var.metadata.name

  standard_labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "planton-ai_resource" = "true"
      "planton-ai_name"     = var.metadata.name
      "planton-ai_kind"     = "HetznerCloudLoadBalancer"
      "planton-ai_org"      = var.metadata.org != null ? var.metadata.org : ""
      "planton-ai_env"      = var.metadata.env != null ? var.metadata.env : ""
      "planton-ai_id"       = var.metadata.id != null ? var.metadata.id : ""
    },
  )

  algorithm = (
    var.spec.algorithm != null && var.spec.algorithm != "" && var.spec.algorithm != "algorithm_unspecified"
    ? var.spec.algorithm
    : "round_robin"
  )

  services = { for svc in var.spec.services : tostring(
    svc.listen_port != null
    ? svc.listen_port
    : (svc.protocol == "http" ? 80 : (svc.protocol == "https" ? 443 : 0))
  ) => merge(svc, {
    effective_listen_port = (
      svc.listen_port != null
      ? svc.listen_port
      : (svc.protocol == "http" ? 80 : (svc.protocol == "https" ? 443 : 0))
    )
    effective_destination_port = (
      svc.destination_port != null
      ? svc.destination_port
      : (svc.listen_port != null
        ? svc.listen_port
        : (svc.protocol == "http" ? 80 : (svc.protocol == "https" ? 443 : 0)))
    )
  }) }
}
