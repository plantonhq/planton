locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurenetworkwatcherflowlog"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over them:
  # user tags deliberately win so an org's governance conventions (cost
  # center, owner) can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The regional Network Watcher the flow log attaches to. Unset spec
  # fields resolve to the AUTO-CREATED singleton -- Azure names it
  # "NetworkWatcher_{region}" and homes it in "NetworkWatcherRG" the
  # moment the region hosts a virtual network (one watcher per region
  # per subscription). Both fields override together (spec-validated)
  # for subscriptions running a self-managed watcher.
  network_watcher_name = (
    var.spec.network_watcher_name != ""
    ? var.spec.network_watcher_name
    : "NetworkWatcher_${var.spec.region}"
  )
  network_watcher_resource_group = (
    var.spec.network_watcher_resource_group != ""
    ? var.spec.network_watcher_resource_group
    : "NetworkWatcherRG"
  )
}
