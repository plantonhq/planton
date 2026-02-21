# HetznerCloudFloatingIp

The **HetznerCloudFloatingIp** resource allocates a reassignable public IP address in Hetzner Cloud — either a single IPv4 address or an IPv6 /64 network block. Unlike a Primary IP (which occupies a server's primary public IP slot), a Floating IP is a secondary address that can be moved between servers in the same location at any time. This makes Floating IPs the standard building block for failover, high availability, and rolling deployment patterns where a stable public endpoint must survive server replacement without DNS propagation delay.

## What It Represents

A [Hetzner Cloud Floating IP](https://docs.hetzner.cloud/#floating-ips) is a managed public IP address homed to a specific location. For IPv4, a single address is allocated (e.g., `203.0.113.42`). For IPv6, a /64 network block is allocated (e.g., `2001:db8::/64`). The address is assigned by Hetzner Cloud and cannot be chosen. A Floating IP can be assigned to a server, unassigned, and reassigned to a different server — all without changing the IP address itself. The server must configure an IP alias on its network interface to accept traffic on the Floating IP address.

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_floating_ip` | 1 | Always | Allocates the IP address with type, location, optional server assignment, labels, and protection settings |
| `hcloud_rdns` | 0 or 1 | When `dnsPtr` is non-empty | Sets a reverse DNS pointer record for the allocated IP address |

The rDNS resource is bundled because it is tightly coupled to the IP address: it cannot exist without the Floating IP's allocated address, and there is no use case for managing rDNS on a Floating IP as a separate component. When `dnsPtr` is omitted, only the Floating IP resource is created.

Server assignment uses the inline `server_id` attribute on the `hcloud_floating_ip` resource rather than a separate `hcloud_floating_ip_assignment` resource — this covers the common case without introducing a third resource type.

## Key Features

### IPv4 and IPv6 Support

The `type` field selects between an IPv4 address and an IPv6 /64 block. This is an immutable choice — changing it forces replacement of the Floating IP (and a new address is allocated).

### Location-Aware Allocation

The `homeLocation` field determines where the IP is allocated (`fsn1`, `nbg1`, `hel1`, `ash`, `hil`, `sin`). A Floating IP can only be assigned to a server in the same location. This is also immutable — changing location forces replacement.

### Optional Server Assignment

The `serverId` field optionally assigns the Floating IP to a server at creation time. It accepts a literal Hetzner Cloud server ID (as a string) or a reference to a `HetznerCloudServer` resource's output via `valueFrom`. If omitted, the Floating IP is created unassigned (reserved) and can be assigned later by updating the spec. Assignment changes are in-place updates — they do not trigger Floating IP replacement.

### Optional Reverse DNS

When `dnsPtr` is set, an `hcloud_rdns` resource maps the allocated IP back to the specified hostname. This is required for mail servers (SPF/DKIM verification relies on matching forward and reverse DNS) and any service where clients verify identity via reverse lookup.

### Delete Protection

When `deleteProtection` is enabled, the Floating IP cannot be deleted via the Hetzner Cloud API until protection is explicitly removed. This prevents accidental loss of an IP address that may have DNS records, email reputation, OS configurations, or failover scripts associated with it.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the Floating IP from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence. The rDNS resource does not support labels in the Hetzner Cloud API.

## Upstream Dependencies (What This Resource Needs)

| Dependency | Field | Required | Purpose |
|---|---|---|---|
| `HetznerCloudServer` | `spec.serverId` | No | Server to assign the Floating IP to. Only needed if assigning at creation time. |

`HetznerCloudFloatingIp` can be created without any dependencies. The server assignment is optional.

## Downstream Dependents (What References This Resource)

No other components in the current catalog directly reference `HetznerCloudFloatingIp` outputs. The `floating_ip_id` and `ip_address` outputs are available for external automation (monitoring, DNS record creation, firewall rule configuration).

## Stack Outputs

| Output | Description |
|---|---|
| `floating_ip_id` | Hetzner Cloud numeric ID of the created Floating IP (as string). Can be used for monitoring or external automation. |
| `ip_address` | The allocated IP address. For IPv4, a single address (e.g., `203.0.113.42`). For IPv6, the first address in the /64 block (e.g., `2001:db8::1`). |
| `ip_network` | The allocated IPv6 /64 CIDR (e.g., `2001:db8::/64`). Empty for IPv4 Floating IPs. |

## References

- [Hetzner Cloud Floating IPs Documentation](https://docs.hetzner.cloud/#floating-ips)
- [Terraform hcloud_floating_ip Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/floating_ip)
- [Terraform hcloud_rdns Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/rdns)
- [Pulumi hcloud.FloatingIp Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/floatingip/)
- [Pulumi hcloud.Rdns Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/rdns/)
