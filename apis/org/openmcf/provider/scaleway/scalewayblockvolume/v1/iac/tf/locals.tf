locals {
  # Volume name from metadata
  volume_name = var.metadata.name

  # Zone from spec
  zone = var.spec.zone

  # Size in GB
  size_gb = var.spec.size_gb

  # Map performance tier enum string to Scaleway IOPS integer.
  # The proto enum values are: "sbs_5k" (5,000 IOPS), "sbs_15k" (15,000 IOPS).
  performance_tier_to_iops = {
    "sbs_5k"  = 5000
    "sbs_15k" = 15000
  }

  iops = local.performance_tier_to_iops[var.spec.performance_tier]

  # Whether to create from a snapshot
  from_snapshot = var.spec.snapshot_id != null && var.spec.snapshot_id != ""

  # Standard OpenMCF tags formatted as flat strings for Scaleway
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayBlockVolume",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
