locals {
  # Topics carry no Azure tags: ARM does not support tags on Service Bus
  # entities, so the platform's identity tags live on the parent
  # namespace.

  # Gate-state wire values. The tfvars wire format carries the FULL
  # proto enum value name; unset deploys Active. Topics support only
  # Active/Disabled -- direction gating happens per subscription.
  status_map = {
    "ACTIVE"   = "Active"
    "DISABLED" = "Disabled"
  }
  status = (
    var.spec.status != null && var.spec.status != ""
    ? local.status_map[var.spec.status]
    : "Active"
  )

  # The namespace name, parsed from the resolved namespace ARM ID for
  # the output -- consumers frequently need the namespace/topic
  # name pair. The anchored regex fails the plan loudly if the ID is not
  # a Service Bus namespace ARM ID.
  namespace_name = regex("/namespaces/(?P<name>[^/]+)$", var.spec.namespace_id)["name"]
}
