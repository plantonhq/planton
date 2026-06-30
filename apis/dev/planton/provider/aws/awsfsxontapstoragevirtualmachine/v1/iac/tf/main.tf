# ---------------------------------------------------------------------------
# FSx ONTAP Storage Virtual Machine
# ---------------------------------------------------------------------------

resource "aws_fsx_ontap_storage_virtual_machine" "this" {
  file_system_id             = var.file_system_id
  name                       = var.svm_name
  root_volume_security_style = var.root_volume_security_style
  svm_admin_password         = var.svm_admin_password != "" ? var.svm_admin_password : null

  dynamic "active_directory_configuration" {
    for_each = var.active_directory_configuration != null ? [var.active_directory_configuration] : []

    content {
      netbios_name = active_directory_configuration.value.netbios_name != "" ? active_directory_configuration.value.netbios_name : null

      self_managed_active_directory_configuration {
        domain_name                            = active_directory_configuration.value.domain_name
        dns_ips                                = active_directory_configuration.value.dns_ips
        username                               = active_directory_configuration.value.username
        password                               = active_directory_configuration.value.password
        file_system_administrators_group       = active_directory_configuration.value.file_system_administrators_group
        organizational_unit_distinguished_name = active_directory_configuration.value.organizational_unit_distinguished_name != "" ? active_directory_configuration.value.organizational_unit_distinguished_name : null
      }
    }
  }

  tags = merge(var.labels, {
    Name = var.resource_name
  })
}
