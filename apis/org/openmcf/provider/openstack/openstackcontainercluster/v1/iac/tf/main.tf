resource "openstack_containerinfra_cluster_v1" "main" {
  name                = local.cluster_name
  cluster_template_id = local.cluster_template

  master_count        = var.spec.master_count
  node_count          = var.spec.node_count
  keypair             = local.keypair
  flavor              = var.spec.flavor != "" ? var.spec.flavor : null
  master_flavor       = var.spec.master_flavor != "" ? var.spec.master_flavor : null
  docker_volume_size  = var.spec.docker_volume_size
  labels              = length(var.spec.labels) > 0 ? var.spec.labels : null
  create_timeout      = var.spec.create_timeout
  floating_ip_enabled = var.spec.floating_ip_enabled
  region              = var.spec.region != "" ? var.spec.region : null
}
