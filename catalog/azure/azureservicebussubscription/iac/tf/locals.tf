locals {
  # Subscriptions carry no Azure tags: ARM does not support tags on
  # Service Bus entities, so the platform's identity tags live on the
  # parent namespace.

  # Gate-state wire values. The tfvars wire format carries the FULL
  # proto enum value name; unset deploys Active.
  status_map = {
    "ACTIVE"           = "Active"
    "DISABLED"         = "Disabled"
    "RECEIVE_DISABLED" = "ReceiveDisabled"
  }
  status = (
    var.spec.status != null && var.spec.status != ""
    ? local.status_map[var.spec.status]
    : "Active"
  )

  # Filter-type wire values -- the provider validates these
  # case-sensitively.
  filter_type_map = {
    "SQL_FILTER"         = "SqlFilter"
    "CORRELATION_FILTER" = "CorrelationFilter"
  }

  # The topic and namespace names, parsed from the resolved topic ARM ID
  # for the outputs -- consumers frequently need the
  # namespace/topic/subscription triple. The anchored regexes fail the
  # plan loudly if the ID is not a Service Bus topic ARM ID.
  topic_name     = regex("/topics/(?P<name>[^/]+)$", var.spec.topic_id)["name"]
  namespace_name = regex("/namespaces/(?P<name>[^/]+)/topics/", var.spec.topic_id)["name"]
}
