locals {
  # Volumes carry TWO names: the ONTAP-internal spec.name (the volume identity
  # in junction paths, SnapMirror, and the ONTAP CLI — underscore-only
  # charset) and the Infra Component's metadata.name, which becomes the Name
  # tag so the AWS console shows the same identity both engines pin.
  resource_name = var.metadata.name

  # Resource-identity tags match the Pulumi module key-for-key (the canonical
  # six-key identity map -- user labels never merge into cloud tags).
  aws_tags = {
    "Name"                     = local.resource_name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsFsxOntapVolume"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # The provider models size_in_bytes as a string-typed nullable integer (it
  # must represent values past 2^31); the spec's int64 arm converts here.
  size_in_bytes = var.spec.size_in_bytes != null ? tostring(var.spec.size_in_bytes) : null

  # Empty strings become null so unset stays indistinguishable from the
  # ONTAP-side defaults (security style inherits from the SVM; snapshot
  # policy defaults to "default"; no junction path means "not mounted").
  junction_path   = var.spec.junction_path != "" ? var.spec.junction_path : null
  security_style  = var.spec.security_style != "" ? var.spec.security_style : null
  snapshot_policy = var.spec.snapshot_policy != "" ? var.spec.snapshot_policy : null

  # final_backup_tags only matter when a final backup is actually taken.
  final_backup_tags = length(var.spec.final_backup_tags) > 0 ? var.spec.final_backup_tags : null
}
