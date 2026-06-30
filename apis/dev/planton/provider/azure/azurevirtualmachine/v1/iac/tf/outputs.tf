output "vm_id" {
  description = "The Azure resource ID of the Virtual Machine"
  value       = local.use_ssh_key ? azurerm_linux_virtual_machine.this[0].id : azurerm_windows_virtual_machine.this[0].id
}

output "vm_name" {
  description = "The name of the Virtual Machine"
  value       = local.name
}

output "private_ip_address" {
  description = "The private IP address assigned to the VM's primary network interface"
  value       = azurerm_network_interface.this.private_ip_address
}

output "public_ip_address" {
  description = "The public IP address assigned to the VM (if public IP is enabled)"
  value       = local.enable_public_ip ? azurerm_public_ip.this[0].ip_address : null
}

output "public_ip_fqdn" {
  description = "The FQDN of the public IP (if public IP is enabled with DNS label)"
  value       = local.enable_public_ip ? try(azurerm_public_ip.this[0].fqdn, null) : null
}

output "computer_name" {
  description = "The computer name (hostname) of the Virtual Machine"
  value       = local.name
}

output "system_assigned_identity_principal_id" {
  description = "The principal ID of the system-assigned managed identity (if enabled)"
  value = (
    var.spec.enable_system_assigned_identity ?
    (local.use_ssh_key ? azurerm_linux_virtual_machine.this[0].identity[0].principal_id : azurerm_windows_virtual_machine.this[0].identity[0].principal_id) :
    null
  )
}

output "network_interface_id" {
  description = "The Azure resource ID of the primary network interface"
  value       = azurerm_network_interface.this.id
}

output "availability_zone" {
  description = "The availability zone where the VM is deployed"
  value       = var.spec.availability_zone
}
