resource "oci_load_balancer_hostname" "this" {
  for_each = local.hostnames_map

  load_balancer_id = oci_load_balancer_load_balancer.this.id
  name             = each.value.name
  hostname         = each.value.hostname
}
