# Pulumi Module: DigitalOcean Reserved IP

Reserves a static public IP (IPv4 or IPv6) and optionally assigns it to a droplet -- one module folding the provider's reserved-IP family. Behavioral parity with the Terraform module is the contract.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean.ReservedIp` | The IPv4 reservation (when `ip_version` is unset or `ipv4`), with the inline mutable droplet assignment |
| `digitalocean.ReservedIpv6` | The IPv6 reservation (when `ip_version` is `ipv6`) |
| `digitalocean.ReservedIpv6Assignment` | The v6 droplet assignment (v6 cannot assign inline) |

## Inputs

`DigitalOceanReservedIpIacInput`: the target `DigitalOceanReservedIp` resource and the DigitalOcean provider config (API token). Droplet references resolve to the literal numeric id before the module runs (parsed with `strconv`).

## Outputs

Exactly the `DigitalOceanReservedIpOutputs` contract: `reserved_ip_address`, `urn` -- sourced from whichever family was created.

## Behavior notes

- The provider's `urn` attribute is renamed by the Pulumi bridge (`ReservedIpUrn` / `ReservedIpv6Urn`) because it collides with Pulumi's own resource URN; both export as the `urn` output.
- v6 assignment goes through the separate assignment resource -- the upstream v6 inline `droplet_id` is dead on create and has no update path.
