# Azure Virtual Machine Terraform Module

This Terraform module deploys an Azure Virtual Machine with the configuration specified in the AzureVirtualMachine manifest.

## Usage

```hcl
module "azure_virtual_machine" {
  source = "./iac/tf"

  metadata = {
    name = "my-vm"
  }

  spec = {
    region         = "eastus"
    resource_group = "my-rg"
    vm_size        = "Standard_D2s_v3"
    subnet_id      = "/subscriptions/.../subnets/default"
    image = {
      publisher = "Canonical"
      offer     = "0001-com-ubuntu-server-jammy"
      sku       = "22_04-lts-gen2"
    }
    ssh_public_key = "ssh-rsa AAAAB3..."
  }

  provider_config = {
    subscription_id = var.subscription_id
    tenant_id       = var.tenant_id
    client_id       = var.client_id
    client_secret   = var.client_secret
  }
}
```

## Requirements

- Terraform >= 1.0
- Azure Provider >= 3.0

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata including name | object | yes |
| spec | AzureVirtualMachine specification | object | yes |
| provider_config | Azure provider credentials | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| vm_id | The Azure resource ID of the Virtual Machine |
| vm_name | The name of the Virtual Machine |
| private_ip_address | The private IP address of the VM |
| public_ip_address | The public IP address (if enabled) |
| network_interface_id | The network interface ID |
| system_assigned_identity_principal_id | The managed identity principal ID (if enabled) |
