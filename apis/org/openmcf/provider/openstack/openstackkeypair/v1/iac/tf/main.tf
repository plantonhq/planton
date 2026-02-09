# main.tf

# Create the OpenStack compute keypair.
# If public_key is provided, it imports the existing key.
# If public_key is omitted, OpenStack generates a new keypair.
resource "openstack_compute_keypair_v2" "main" {
  name = local.keypair_name

  # Only set public_key when importing an existing key
  public_key = local.is_import ? var.spec.public_key : null

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
