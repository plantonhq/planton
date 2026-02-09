# main.tf

# Create the OpenStack compute volume attachment.
# This attaches a Cinder volume to a Nova instance, making the volume
# appear as a block device inside the instance (e.g., /dev/vdb).
resource "openstack_compute_volume_attach_v2" "main" {
  instance_id = local.instance_id
  volume_id   = local.volume_id

  # Device path (optional, Nova auto-selects if omitted).
  device = var.spec.device != "" ? var.spec.device : null

  # Region override (optional, ForceNew).
  region = var.spec.region != "" ? var.spec.region : null
}
