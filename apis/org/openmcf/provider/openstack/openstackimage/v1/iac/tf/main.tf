# main.tf

# Create the OpenStack Glance image.
# Images are virtual disk templates used to boot compute instances or
# initialize bootable Cinder volumes.
resource "openstack_images_image_v2" "main" {
  name             = var.metadata.name
  container_format = var.spec.container_format
  disk_format      = var.spec.disk_format

  # Image source URL (optional).
  image_source_url = var.spec.image_source_url != "" ? var.spec.image_source_url : null

  # Minimum requirements (optional).
  min_disk_gb = var.spec.min_disk_gb
  min_ram_mb  = var.spec.min_ram_mb

  # Protection and visibility.
  protected  = var.spec.protected
  hidden     = var.spec.hidden
  visibility = var.spec.visibility

  # Tags (optional).
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
