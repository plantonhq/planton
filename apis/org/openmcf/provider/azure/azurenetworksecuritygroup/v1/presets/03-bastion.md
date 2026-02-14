# Bastion NSG

This preset creates a Network Security Group for bastion or jump-host subnets, allowing SSH and RDP access only from trusted IP ranges. All other internet traffic is explicitly denied. This is the standard NSG for secure remote administration access points.

## When to Use

- Subnets hosting bastion or jump-host VMs for remote administration
- Secure entry points into the VNet for SSH and RDP sessions
- Environments that require controlled, auditable remote access

## Key Configuration Choices

- **SSH from trusted IPs** (`priority: 100, destinationPortRange: "22"`) -- Linux remote access restricted to your office, VPN, or CI/CD runner IP ranges
- **RDP from trusted IPs** (`priority: 110, destinationPortRange: "3389"`) -- Windows remote access with the same IP restriction. Remove if only Linux bastion hosts are used
- **Deny all internet** (`priority: 4000, access: Deny`) -- Explicit catch-all deny for any internet traffic not matching SSH/RDP rules
- **Trusted CIDR restriction** -- Replace `<your-trusted-cidr>` with your office IP range (e.g., `203.0.113.0/24`) or VPN gateway CIDR. Never use `*` or `Internet` for bastion access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the associated subnet) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-nsg-name>` | Name for the NSG (unique within resource group) | Your naming convention |
| `<your-trusted-cidr>` | CIDR range of trusted source IPs (e.g., `203.0.113.0/24`) | Your network team or VPN configuration |

## Related Presets

- **01-web-tier** -- Use for internet-facing subnets (allows HTTP/HTTPS inbound)
- **02-database-tier** -- Use for database subnets (allows only VNet-internal traffic)
