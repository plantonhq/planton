# A Record

This preset creates a DNS A record that maps a fully qualified domain name to an IPv4 address. A records are the most fundamental DNS record type and the starting point for pointing a domain at any OCI resource with a public or private IP -- load balancers, compute instances, NAT gateways, or network firewalls. The record set is managed atomically: updates replace all items in a single operation.

## When to Use

- Pointing a domain at a load balancer's public IP address
- Mapping a hostname to a compute instance for direct access
- Creating DNS entries for network appliances (firewalls, NAT gateways) with reserved IPs
- Any scenario where a domain name needs to resolve to one or more IPv4 addresses

## Key Configuration Choices

- **A record type** (`rtype: A`) -- resolves a domain name to an IPv4 address. For IPv6, use `AAAA` instead.
- **300-second TTL** (`ttl: 300`) -- 5-minute cache lifetime balances fast propagation during changes with reasonable resolver caching. Lower values (60s) are appropriate during migrations; higher values (3600s) reduce query volume for stable records.
- **Single record item** -- the preset includes one IP address. Add additional items to the `items` list for round-robin DNS across multiple endpoints.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<zone-name-or-ocid>` | Name or OCID of the target DNS zone (e.g., `example.com` or zone OCID) | OCI Console > DNS Management > Zones, or `OciDnsZone` status outputs (`zoneId`) |
| `<ipv4-address>` | IPv4 address the record resolves to (e.g., `203.0.113.10`) | OCI Console > target resource details, or `OciPublicIp` / `OciApplicationLoadBalancer` status outputs |

## Related Presets

- **02-cname-alias** -- use instead when you want to alias a domain to another domain name rather than a specific IP address
