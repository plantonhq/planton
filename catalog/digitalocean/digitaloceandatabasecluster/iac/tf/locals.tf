locals {
  # Spec enum value names are exactly the DigitalOcean API slugs
  # (pg, mysql, redis, mongodb, kafka, opensearch, valkey / nyc3, sfo3, ...),
  # so they pass through unchanged.
  engine_slug = var.spec.engine
  region_slug = var.spec.region

  # Optional VPC UUID. References are resolved to the literal UUID before
  # the module runs, so the field arrives as a plain string.
  vpc_uuid = try(var.spec.vpc, "") != "" ? var.spec.vpc : null

  # Optional project placement.
  project_id = try(var.spec.project_id, "") != "" ? var.spec.project_id : null

  # The provider's storage_size_mib is a string holding a bare MiB count;
  # the spec carries GiB for ergonomics. Unset means "use the size slug's
  # default storage".
  storage_size_mib = var.spec.storage_gib != null && var.spec.storage_gib > 0 ? tostring(var.spec.storage_gib * 1024) : null

  # Engine-conditional arguments: null when unset so the provider's own
  # engine checks never fire for the engines they don't apply to.
  eviction_policy = try(var.spec.eviction_policy, "") != "" ? var.spec.eviction_policy : null
  sql_mode        = try(var.spec.sql_mode, "") != "" ? var.spec.sql_mode : null

  # Standard Planton labels rendered as DigitalOcean "key:value" tags —
  # the exact set and key spelling the Pulumi module applies, so both
  # provisioners tag identically.
  planton_tags = concat(
    [
      "planton-ai_resource:true",
      "planton-ai_name:${var.metadata.name}",
      "planton-ai_kind:DigitalOceanDatabaseCluster",
    ],
    try(var.metadata.org, "") != "" && var.metadata.org != null ? ["planton-ai_organization:${var.metadata.org}"] : [],
    try(var.metadata.env, "") != "" && var.metadata.env != null ? ["planton-ai_environment:${var.metadata.env}"] : [],
    try(var.metadata.id, "") != "" && var.metadata.id != null ? ["planton-ai_id:${var.metadata.id}"] : [],
  )

  tags = distinct(concat(coalesce(var.spec.tags, []), local.planton_tags))

  # DigitalOcean caps a database cluster's COMBINED tags -- the tag names
  # joined by commas -- at 255 characters (measured 2026-09-17 against
  # `POST /v2/databases`: 255 pass, 256 fail with `422 combined tags cannot
  # exceed 255 characters`; colons count as one character and the tag count
  # matters only through the separators). The six Planton label tags carry
  # metadata.name and metadata.id, so a long resource name spends the budget
  # before any spec.tags entry does. The resource's precondition checks
  # this before anything renders, with the same number and the same message
  # as the Pulumi module (its twin).
  tags_combined_budget = 255
  tags_combined        = join(",", local.tags)
  planton_tags_length  = length(join(",", local.planton_tags))
}
