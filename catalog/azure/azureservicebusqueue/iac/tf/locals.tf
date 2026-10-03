locals {
  # Queues carry no Azure tags: ARM does not support tags on Service Bus
  # entities (namespaces/queues/topics), so the platform's identity tags
  # live on the parent namespace.

  # Gate-state wire values. The tfvars wire format carries the FULL
  # proto enum value name; unset deploys Active.
  status_map = {
    "ACTIVE"           = "Active"
    "DISABLED"         = "Disabled"
    "SEND_DISABLED"    = "SendDisabled"
    "RECEIVE_DISABLED" = "ReceiveDisabled"
  }
  status = (
    var.spec.status != null && var.spec.status != ""
    ? local.status_map[var.spec.status]
    : "Active"
  )

  # The namespace name, parsed from the resolved namespace ARM ID for
  # the output -- consumers frequently need the namespace/queue
  # name pair. The anchored regex fails the plan loudly if the ID is not
  # a Service Bus namespace ARM ID.
  namespace_name = regex("/namespaces/(?P<name>[^/]+)$", var.spec.namespace_id)["name"]
}
