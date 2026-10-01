# Terraform Module: DigitalOcean VPC Peering

Provisions a private-network peering between exactly two VPCs -- the complete `digitalocean_vpc_peering` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_vpc_peering.peering` | The peering connection: name + the two-VPC set |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanVpcPeeringSpec` proto: `peering_name` and the two flattened VPC reference strings `vpc_1` / `vpc_2`. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanVpcPeeringStackOutputs` contract: `peering_id`. The provider's `status` attribute is deliberately not exported (the resource waits for ACTIVE, so a stored status could only ever say ACTIVE and would go stale; live status belongs to whoever reads the API).

## Behavior notes

- The provider's `vpc_ids` is an unordered set -- member order never diffs; the module assembles it from the spec's two named references (exactly-two is unrepresentable-wrong at validation).
- Only `name` updates in place; changing either VPC replaces the peering.
- Create waits for ACTIVE; delete retries through DigitalOcean's transient 403s while the peering settles (2-minute provider timeouts).
- Import: `terraform import ... <peering_id>` (see `iac/import-map.yaml`).
