# Pulumi Module: DigitalOcean VPC Peering

Provisions a private-network peering between exactly two VPCs -- the complete `digitalocean_vpc_peering` resource surface. Behavioral parity with the Terraform module is the contract.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean.VpcPeering` | The peering connection: name + the two-VPC set |

## Inputs

`DigitalOceanVpcPeeringStackInput`: the target `DigitalOceanVpcPeering` resource and the DigitalOcean provider config (API token).

## Outputs

Exactly the `DigitalOceanVpcPeeringStackOutputs` contract: `peering_id` (Pulumi's resource id). The SDK's `Status` property is deliberately not exported (the resource waits for ACTIVE, so a stored status could only ever say ACTIVE and would go stale; live status belongs to whoever reads the API).

## Behavior notes

- The bridge flattens the provider's unordered `vpc_ids` set to an array; the pair's order is irrelevant to DigitalOcean.
- Only `name` updates in place; changing either VPC replaces the peering.
- The provider's declared `id` attribute is absorbed into Pulumi's built-in resource id -- exported here as `peering_id`.
