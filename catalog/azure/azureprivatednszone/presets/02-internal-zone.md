# Custom Internal DNS Zone

This preset creates a custom internal zone (e.g. `corp.internal`) for
private name resolution, with the two SOA fields worth pinning: a real
DNS-admin contact and a 30-second negative-caching TTL so newly-created
records become visible quickly to clients that asked before the record
existed.

To make VMs discoverable by hostname automatically, pair the zone with an
`AzurePrivateDnsZoneVirtualNetworkLink` that sets
`registrationEnabled: true` (one such link per network) -- see that
kind's presets.

## When to Use

- Internal service discovery domains for VMs and appliances
- Lift-and-shift estates that relied on on-premises internal DNS

## Key Configuration Choices

- **`minimumTtl: 30`** -- the default (10) is aggressive for zones edited
  by automation; 30-60 seconds trades a little propagation speed for less
  resolver churn. Remove the whole `soaRecord` block to take Azure's
  defaults.
- **SOA email uses host format** -- dots instead of `@`
  (`dnsadmin.mycompany.com` for dnsadmin@mycompany.com)
- **The SOA is written at creation** -- changing it later replaces the
  zone, so decide before records accumulate

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<resource-group-name>` | The resource group to create the zone in | The resource group's `status.outputs.resource_group_name` |
| `corp.internal` | Your internal domain (e.g. `corp.internal`) | Your DNS architecture |
| `<dns-admin-email-soa-format>` | DNS contact in SOA host format | Your ops contact conventions |
| `<cost-center>` | Your org's cost-attribution tag value | Your tagging convention |
