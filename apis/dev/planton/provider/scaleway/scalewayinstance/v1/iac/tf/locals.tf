locals {
  # Resource identity
  instance_name = var.metadata.name

  # Spec fields
  zone               = var.spec.zone
  instance_type      = var.spec.type
  image              = var.spec.image
  security_group_id  = var.spec.security_group_id
  private_network_id = var.spec.private_network_id
  cloud_init         = var.spec.cloud_init
  state              = var.spec.state
  protected          = var.spec.protected

  # Public IP: create when spec.public_ip is not null
  create_public_ip = var.spec.public_ip != null
  reverse_dns      = var.spec.public_ip != null ? var.spec.public_ip.reverse_dns : ""

  # Root volume configuration (safely unwrap optional nested object)
  root_volume_size_in_gb            = var.spec.root_volume != null ? var.spec.root_volume.size_in_gb : null
  root_volume_type                  = var.spec.root_volume != null ? var.spec.root_volume.volume_type : null
  root_volume_delete_on_termination = var.spec.root_volume != null ? var.spec.root_volume.delete_on_termination : true
  root_volume_sbs_iops              = var.spec.root_volume != null ? var.spec.root_volume.sbs_iops : null

  # Additional volumes: build an indexed map for for_each
  additional_volumes_map = {
    for i, vol in var.spec.additional_volumes : tostring(i) => {
      name        = vol.name != "" ? vol.name : "${var.metadata.name}-vol-${i}"
      volume_type = vol.volume_type
      size_in_gb  = vol.size_in_gb
    }
  }

  # Standard Planton tags formatted as flat strings for Scaleway
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayInstance",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
