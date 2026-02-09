resource "openstack_containerinfra_clustertemplate_v1" "main" {
  name  = local.template_name
  coe   = var.spec.coe
  image = local.image

  keypair_id          = local.keypair
  external_network_id = local.external_network
  fixed_network       = local.fixed_network
  fixed_subnet        = local.fixed_subnet

  network_driver     = var.spec.network_driver != "" ? var.spec.network_driver : null
  volume_driver      = var.spec.volume_driver != "" ? var.spec.volume_driver : null
  dns_nameserver     = var.spec.dns_nameserver != "" ? var.spec.dns_nameserver : null
  docker_volume_size = var.spec.docker_volume_size
  flavor             = var.spec.flavor != "" ? var.spec.flavor : null
  master_flavor      = var.spec.master_flavor != "" ? var.spec.master_flavor : null
  floating_ip_enabled = var.spec.floating_ip_enabled
  master_lb_enabled   = var.spec.master_lb_enabled
  tls_disabled        = var.spec.tls_disabled
  labels             = length(var.spec.labels) > 0 ? var.spec.labels : null
  region             = var.spec.region != "" ? var.spec.region : null
}
