# main.tf

# Create the OpenStack Cinder block storage volume.
# Volumes provide persistent block storage for compute instances.
resource "openstack_blockstorage_volume_v3" "main" {
  name = var.metadata.name
  size = var.spec.size

  # Description (optional).
  description = var.spec.description != "" ? var.spec.description : null

  # Volume type (optional, Cinder default if omitted).
  volume_type = var.spec.volume_type != "" ? var.spec.volume_type : null

  # Availability zone (optional, ForceNew).
  availability_zone = var.spec.availability_zone != "" ? var.spec.availability_zone : null

  # Source fields (mutually exclusive, all ForceNew).
  snapshot_id   = var.spec.snapshot_id != "" ? var.spec.snapshot_id : null
  source_vol_id = var.spec.source_vol_id != "" ? var.spec.source_vol_id : null
  image_id      = local.image_id

  # Metadata (optional).
  metadata = length(var.spec.metadata) > 0 ? var.spec.metadata : null

  # Region override (optional, ForceNew).
  region = var.spec.region != "" ? var.spec.region : null
}
