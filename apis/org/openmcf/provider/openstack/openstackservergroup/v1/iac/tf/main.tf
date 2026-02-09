# main.tf

# Create the OpenStack Compute server group.
# Server groups control instance placement on hypervisors via affinity policies.
# The OpenStack API accepts policies as a list, but only one policy is allowed
# per server group. We wrap the singular policy string into a list here.
resource "openstack_compute_servergroup_v2" "main" {
  name     = var.metadata.name
  policies = [var.spec.policy]

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
