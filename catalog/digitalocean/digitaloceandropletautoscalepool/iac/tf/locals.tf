locals {
  # Exactly one scaling branch is set (spec oneof); the config block below
  # renders only that branch's leaves. Null attributes are omitted from the
  # request, matching the API's zero-means-unset wire behavior.
  is_dynamic = var.spec.dynamic != null

  # Optional template references resolve to literal ids before the module
  # runs; empty means unset.
  configured_vpc_uuid = try(var.spec.droplet_template.vpc, "") != "" ? var.spec.droplet_template.vpc : null
  project_id          = try(var.spec.droplet_template.project_id, "") != "" ? var.spec.droplet_template.project_id : null

  # The VPC is always sent explicitly. When the spec leaves it unset,
  # DigitalOcean places members in the region's default VPC and reports that
  # UUID back on every read; the provider's `vpc_uuid` is Optional but not
  # Computed, so an unset value beside a populated read-back would re-plan
  # on every apply (measured live: an unset vpc_uuid read back as the
  # default-<region> VPC). Sending the API's own default makes the manifest
  # and the cloud agree. The Pulumi module resolves it the same way.
  vpc_uuid = coalesce(local.configured_vpc_uuid, one(data.digitalocean_vpc.region_default[*].id))

  # Standard Planton labels rendered as DigitalOcean "key:value" tags —
  # the exact set and key spelling the Pulumi module applies, so both
  # provisioners tag identically. Tags land on every member droplet.
  planton_tags = concat(
    [
      "planton-ai_resource:true",
      "planton-ai_name:${var.metadata.name}",
      "planton-ai_kind:DigitalOceanDropletAutoscalePool",
    ],
    try(var.metadata.org, "") != "" && var.metadata.org != null ? ["planton-ai_organization:${var.metadata.org}"] : [],
    try(var.metadata.env, "") != "" && var.metadata.env != null ? ["planton-ai_environment:${var.metadata.env}"] : [],
    try(var.metadata.id, "") != "" && var.metadata.id != null ? ["planton-ai_id:${var.metadata.id}"] : [],
  )

  tags = distinct(concat(coalesce(var.spec.droplet_template.tags, []), local.planton_tags))
}
