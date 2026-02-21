# DNS Proxy Enabled

This preset creates an OCI Bastion with DNS proxy and SOCKS5 support enabled, allowing sessions to target resources using fully qualified domain names (FQDNs) instead of IP addresses. This is essential for environments where target resource IPs are dynamic (auto-scaling groups, container instances) or where DNS-based service discovery is the standard access pattern.

## When to Use

- Environments where target resources have dynamic IP addresses that change on restart or scaling events
- Access to OCI managed services that expose private DNS names (Autonomous Database private endpoints, MySQL HeatWave, OKE private API endpoints)
- Teams using SOCKS5 dynamic port forwarding to browse multiple private web UIs through a single session
- Container-based environments where individual containers are addressed by service discovery DNS rather than static IPs

## Key Configuration Choices

- **DNS proxy enabled** (`isDnsProxyEnabled: true`) -- the bastion resolves FQDNs on behalf of the client, allowing sessions to target resources by DNS name (e.g., `mydb.subnet1.vcn1.oraclevcn.com`) instead of requiring the client to know the target IP. This also enables SOCKS5 dynamic port forwarding, where the bastion acts as a proxy for any TCP connection. This setting is immutable after creation.
- **Client CIDR allow list** (`clientCidrBlockAllowList`) -- same as the standard preset. Restricts which source IPs can create and connect to sessions.
- **3-hour max session TTL** (`maxSessionTtlInSeconds: 10800`) -- same as the standard preset. Sessions are time-limited regardless of DNS proxy status.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the bastion will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<private-subnet-ocid>` | OCID of the private subnet the bastion connects to | OCI Console > Networking > Subnets, or `OciSubnet` status outputs |
| `<allowed-cidr>` | CIDR block(s) allowed to connect (e.g., `10.0.0.0/16` for VPN, `203.0.113.5/32` for a single IP) | Your network team or VPN provider documentation |

## Related Presets

- **01-standard-ssh-gateway** -- Use instead when targets have stable IP addresses and DNS proxy overhead is unnecessary
