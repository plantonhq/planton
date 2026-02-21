# Standard IPv4 Primary IP

This preset allocates a persistent public IPv4 address in Hetzner Cloud's Falkenstein datacenter. The IP exists independently of any server -- it survives server deletion and can be reassigned, making it suitable for any endpoint that needs a stable address (web servers, API gateways, game servers, VPN endpoints, etc.).

The IaC module hardcodes `auto_delete: false` and `assignee_type: "server"`. Server assignment is handled by the HetznerCloudServer component, not by this resource.

## When to Use

- Any server that needs a public IPv4 address that outlives server rebuilds or migrations
- Services where DNS A records point to a fixed IP and downtime from IP changes is unacceptable
- Environments where servers are frequently rebuilt (e.g., immutable infrastructure) but the public IP must remain constant

## Key Configuration Choices

- **IPv4** (`type: ipv4`) -- allocates a single public address; the most common choice since most internet clients still require IPv4 reachability
- **Falkenstein location** (`location: fsn1`) -- Hetzner's largest datacenter in the eu-central network zone; change to `nbg1` (Nuremberg), `hel1` (Helsinki), `ash` (Ashburn), `hil` (Hillsboro), or `sin` (Singapore) to match your server location
- **No rDNS** -- omitted because most server workloads do not require reverse DNS; add `dnsPtr` if your use case needs it (see the `02-mail-server-ipv4` preset)
- **No delete protection** -- allows easy teardown during development; set `deleteProtection: true` before promoting to production

## Placeholders to Replace

No placeholders -- this preset is ready to deploy after setting `metadata.name` to the desired resource name.

## Related Presets

- **02-mail-server-ipv4** -- adds reverse DNS and delete protection for mail servers and services that require rDNS verification
