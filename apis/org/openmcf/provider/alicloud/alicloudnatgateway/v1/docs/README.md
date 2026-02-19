# AlicloudNatGateway Research Documentation

## Provider Resource Analysis

### alicloud_nat_gateway (Terraform) / vpc.NatGateway (Pulumi)

The NAT Gateway is the core resource. Key findings from provider analysis:

- **Enhanced vs Normal**: The `nat_type` field supports "Enhanced" (modern, placed in a VSwitch, higher performance) and "Normal" (legacy, no VSwitch placement). Enhanced is the default and recommended type.
- **Billing models**: Two distinct billing paths:
  - `PayByLcu` (Capacity Unit) -- pay per actual usage; no specification needed
  - `PayBySpec` (Fixed Specification) -- fixed throughput tier (Small/Middle/Large/XLarge.1)
- **Deprecated fields**: `name` (use `nat_gateway_name`), `instance_charge_type` (use `payment_type`), `bandwidth_packages` (removed), `spec` (use `specification`)
- **Computed outputs**: `snat_table_ids` and `forward_table_ids` are auto-created with the gateway

### alicloud_eip_association (Terraform) / ecs.EipAssociation (Pulumi)

Binds an EIP to an instance (NAT Gateway, ECS, SLB):

- `allocation_id` -- the EIP allocation ID
- `instance_id` -- the NAT Gateway ID
- `instance_type` -- auto-detected from ID prefix ("Nat" for `ngw-*` IDs), but explicitly set for clarity

### alicloud_snat_entry (Terraform) / vpc.SnatEntry (Pulumi)

Maps private traffic sources to the EIP's public IP:

- `snat_table_id` -- from the NAT Gateway's computed output
- `snat_ip` -- the EIP's **IP address** (not ID), requiring a data source lookup
- `source_vswitch_id` and `source_cidr` are mutually exclusive (ForceNew)

## Design Rationale

### EIP IP Resolution via Data Source

SNAT entries require the EIP's IP address string, but OpenMCF users reference the EIP by its allocation ID. Rather than exposing two fields (`eip_id` + `eip_ip_address`), the IaC modules resolve the IP internally:

- **Pulumi**: `ecs.GetEipAddresses(ctx, &ecs.GetEipAddressesArgs{Ids: []string{eipId}})`
- **Terraform**: `data "alicloud_eip_addresses" "nat" { ids = [var.spec.eip_id] }`

This is safe because OpenMCF middleware guarantees the EIP is provisioned before the NAT Gateway component runs.

### Single EIP Scope

The provider supports multiple EIPs per NAT Gateway. This v1 component supports one EIP, covering the vast majority of use cases. Multi-EIP support can be added in v2 by changing `eip_id` to `repeated eip_ids`.

### Fields Excluded from Spec

The following provider fields were intentionally excluded from the v1 spec:

| Field | Reason |
| --- | --- |
| `period` | Subscription billing detail; managed outside OpenMCF |
| `dry_run` | Operational flag, not declarative infrastructure |
| `force` | Deletion behavior, not creation config |
| `network_type` | Rare; defaults to "internet" |
| `eip_bind_mode` | Rare; defaults to "MULTI_BINDED" |
| `icmp_reply_enabled` | Rare; defaults to true |
| `private_link_enabled` | Niche feature |
| `access_mode` | Niche feature |
| `eip_affinity` (SNAT) | Niche; defaults to 0 (disabled) |

### Composite Bundling (DD07)

NAT Gateway + EIP Association + SNAT Entries are bundled because:
- A NAT Gateway without an EIP association has no public IP to NAT through
- A NAT Gateway without SNAT entries doesn't route any traffic
- The three resources form a single logical unit of "outbound internet access"
