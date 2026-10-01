# DigitalOcean block storage volume.
#
# Attachment to Droplets is a property of the Droplet (its volume_ids list),
# never of the volume. Size can only be EXPANDED after creation -- the
# provider rejects a shrink at plan time. name, region, and description are
# create-only and replace the volume when changed; the three initialization
# arguments are create-only too but IGNORED after creation (see lifecycle).
resource "digitalocean_volume" "this" {
  name = var.spec.volume_name

  # Create-only at the current provider pin: a description change REPLACES
  # the volume. Empty stays unset (the provider rejects empty strings).
  description = var.spec.description != "" ? var.spec.description : null

  # Volumes attach only to Droplets in the same region.
  region = var.spec.region

  size = var.spec.size_gib

  # Formatting happens once at creation; the API never reports these
  # arguments back (the resulting filesystem is observable through separate
  # computed attributes).
  initial_filesystem_type  = local.filesystem_type
  initial_filesystem_label = local.filesystem_label

  # When set, the volume is created from the snapshot, inheriting its region
  # and minimum size. Create-only, never reported back by the API.
  snapshot_id = var.spec.snapshot_id != "" ? var.spec.snapshot_id : null

  tags = local.tags

  # The three initialization arguments are ForceNew at the provider and the
  # API never returns them, so without this block any later edit -- or an
  # adoption (import) whose manifest still describes how the volume was
  # formatted -- plans as a destroy-and-recreate of a volume holding data.
  # They have no meaning after creation (the volume is already formatted, or
  # already cloned from its snapshot), so ignoring them loses nothing and
  # makes import-then-apply a no-op. Mirrored by IgnoreChanges in the Pulumi
  # module; the spec field comments tell manifest authors the same.
  lifecycle {
    ignore_changes = [initial_filesystem_type, initial_filesystem_label, snapshot_id]
  }
}
