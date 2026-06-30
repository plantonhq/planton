# HetznerCloudFirewall

The **HetznerCloudFirewall** resource creates a firewall in a Hetzner Cloud account with inline rules that control inbound and outbound network traffic for servers. When applied to a server (via the server's `firewallIds` field), the firewall enforces a deny-by-default inbound policy — all incoming packets not matching at least one rule are dropped, while outbound traffic is allowed unless explicitly restricted.

## What It Represents

A [Hetzner Cloud Firewall](https://docs.hetzner.cloud/#firewalls) is an account-level, stateful packet filter applied to servers at creation time. It supports up to 50 rules, each specifying a direction, protocol, optional port (for TCP/UDP), and source or destination CIDR blocks. Return traffic for established connections is automatically permitted.

## Bundled Resources

| Terraform Resource | Created When | Purpose |
|---|---|---|
| `hcloud_firewall` | Always | Creates the firewall with inline rules |

This is a single-resource component — rules are defined inline as part of the firewall, not as separate resources.

## Key Features

### Inline Rule Definitions

Rules are declared as a repeated `Rule` message in the spec. Each rule specifies direction (`in`/`out`), protocol (`icmp`/`tcp`/`udp`/`esp`/`gre`), optional port, and CIDR blocks. This keeps the entire security policy in a single manifest.

### Deny-by-Default Inbound

When a firewall is applied to a server, all inbound traffic not matching a rule is dropped. Outbound traffic is allowed by default. An empty rules list creates a firewall that blocks all inbound and allows all outbound — useful as a lockdown configuration.

### Proto-Level Validation

Three CEL validation rules catch common misconfiguration errors at manifest validation time, before any cloud API call:
- Port is required when protocol is `tcp` or `udp`
- `sourceIps` is required when direction is `in`
- `destinationIps` is required when direction is `out`

### Enum-Based Direction and Protocol

Direction and protocol are defined as proto enums rather than free-form strings. Invalid values like `"inbound"` or `"TCP"` are caught by the schema, not at the cloud API level.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the Hetzner Cloud firewall from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence.

## Upstream Dependencies (What This Resource Needs)

None. `HetznerCloudFirewall` is a foundation resource with no foreign key dependencies.

## Downstream Dependents (What References This Resource)

| Dependent | Field | Purpose |
|---|---|---|
| `HetznerCloudServer` | `spec.firewallIds` | Apply firewall rules to server at creation |

## Stack Outputs

| Output | Description |
|---|---|
| `firewall_id` | Hetzner Cloud numeric ID of the created firewall (as string) |

## References

- [Hetzner Cloud Firewalls Documentation](https://docs.hetzner.cloud/#firewalls)
- [Terraform hcloud_firewall Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/firewall)
- [Pulumi hcloud.Firewall Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/firewall/)
