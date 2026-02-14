# Ubuntu 22.04 LTS with SSH Key Authentication

This preset deploys an Ubuntu 22.04 LTS Gen2 VM with SSH key authentication, no public IP, boot diagnostics enabled, and a 30 GB Premium SSD OS disk. This is the standard configuration for secure Linux workloads accessed via private network (VPN, Bastion, or peered VNet).

## When to Use

- Linux application servers, jump boxes, or development VMs
- Workloads accessed via Azure Bastion, VPN, or VNet peering (no public internet exposure)
- Teams that follow security best practices with SSH key-only authentication
- Standard compute needs (2 vCPUs, 8 GiB RAM) that don't require GPU or high memory

## Key Configuration Choices

- **Ubuntu 22.04 LTS Gen2** (`image.sku: 22_04-lts-gen2`) -- Long-term support release with Gen2 VM performance improvements
- **SSH key authentication** (`sshPublicKey`) -- No password; key-based auth is more secure and recommended for production
- **No public IP** (`network.enablePublicIp: false`) -- VM is accessible only via private network; use Azure Bastion or VPN for SSH access
- **Standard_D2s_v3** (`vmSize`) -- 2 vCPUs, 8 GiB RAM; good starting point for general-purpose Linux workloads
- **Premium SSD** (`osDisk.storageType: premium_lrs`) -- Low-latency, high-throughput storage for the OS disk
- **Boot diagnostics enabled** -- Captures serial console output to help diagnose boot issues

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<subnet-id>` | ARM resource ID of the subnet for the VM's NIC | Azure portal or `AzureVpc` / `AzureSubnet` status outputs |
| `<your-ssh-public-key>` | SSH public key (e.g., `ssh-rsa AAAAB3...`) | `~/.ssh/id_rsa.pub` or `ssh-keygen` output |

## Related Presets

- **02-windows-rdp** -- Use instead for Windows Server workloads with RDP access
