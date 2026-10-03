# Terraform Module: DigitalOcean Reserved IP

Reserves a static public IP (IPv4 or IPv6) and optionally assigns it to a droplet -- one module folding the provider's reserved-IP family.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_reserved_ip.ipv4` | The IPv4 reservation (created when `ip_version` is unset or `ipv4`), with the inline mutable `droplet_id` assignment |
| `digitalocean_reserved_ipv6.ipv6` | The IPv6 reservation (created when `ip_version` is `ipv6`) |
| `digitalocean_reserved_ipv6_assignment.ipv6_assignment` | The v6 droplet assignment (v6 cannot assign inline -- see below) |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanReservedIpSpec` proto: optional `ip_version`, `region` (enum value names are the provider's lowercase slugs), optional `droplet` (flattened reference string, parsed to the numeric id). Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanReservedIpOutputs` contract: `reserved_ip_address`, `urn` -- sourced from whichever family was created.

## Behavior notes

- v4 assigns through the reservation's own `droplet_id`, which the provider updates in place (assign/re-point/unassign without replacing the address).
- v6 MUST assign through the assignment resource: the provider's v6 create silently ignores `droplet_id` and the resource has no update function (an inline value would plan forever).
- The provider lowercases regions into state; the spec's enum value names are already lowercase slugs, so no normalization diff.
- The upstream v6 delete swallows non-404 errors -- verify releases with a live lookup when it matters.
- Import: `terraform import ... <address>` for either family (see `iac/import-map.yaml`); assignments never import separately.
