# Basic Host

This preset deploys the default Bastion shape: dedicated infrastructure at fixed capacity, browser-based RDP/SSH sessions to every VM the network can reach, no public IPs on the machines themselves.

## When to Use

- Production or staging networks that need audited VM access without per-machine exposure
- Teams comfortable with browser sessions who do not need native-client tunneling, file copy, or scaling

## Key Configuration Choices

- **BASIC is fixed at 2 scale units** (~40 concurrent sessions) with no feature knobs -- upgrade in-place to Standard when you need tunneling, file copy, or capacity
- **The subnet must be named exactly `AzureBastionSubnet`** (/26+) -- carve it when you design the network
- **The public IP is the host's alone** (Standard SKU, static) -- never share it with another consumer

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group>` | The AzureResourceGroup the host is created in | The resource-group component's name |
| `<your-bastion-subnet>` | The AzureSubnet whose ARM name is `AzureBastionSubnet` | The subnet component's name |
| `<your-bastion-pip>` | The AzurePublicIp the host binds exclusively | The address component's name |

Basic bills hourly from provisioning (~10-minute create), sessions or not — the verified figure lives in the kind's generated estimate at `catalog/_pricing/estimates/azurebastionhost.yaml`. Upgrades to Standard/Premium are in-place; downgrades replace the host.
