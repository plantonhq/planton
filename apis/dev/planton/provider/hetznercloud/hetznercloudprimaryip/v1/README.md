# HetznerCloudPrimaryIp

The **HetznerCloudPrimaryIp** resource allocates a persistent public IP address in Hetzner Cloud — either a single IPv4 address or an IPv6 /64 network block. The IP persists independently of any server: it survives server deletion and can be assigned to a new server, making it suitable for stable endpoints like mail servers, API gateways, and any service whose public address must not change when the underlying compute is replaced.

## What It Represents

A [Hetzner Cloud Primary IP](https://docs.hetzner.cloud/#primary-ips) is a managed public IP address allocated at a specific location. For IPv4, a single address is allocated (e.g., `203.0.113.42`). For IPv6, a /64 network block is allocated (e.g., `2001:db8::/64`). The IP address is assigned by Hetzner Cloud and cannot be chosen. A Primary IP can be assigned to a server's primary public IP slot, replacing the auto-assigned IP that would otherwise be allocated and destroyed with the server.

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_primary_ip` | 1 | Always | Allocates the IP address with type, location, labels, and protection settings |
| `hcloud_rdns` | 0 or 1 | When `dnsPtr` is non-empty | Sets a reverse DNS pointer record for the allocated IP address |

The rDNS resource is bundled because it is tightly coupled to the IP address: it cannot exist without the Primary IP's allocated address, and there is no use case for managing rDNS on a Primary IP as a separate component. When `dnsPtr` is omitted, only the Primary IP resource is created.

## Key Features

### IPv4 and IPv6 Support

The `type` field selects between an IPv4 address and an IPv6 /64 block. This is an immutable choice — changing it forces replacement of the Primary IP (and a new address is allocated).

### Location-Aware Allocation

The `location` field determines where the IP is allocated (`fsn1`, `nbg1`, `hel1`, `ash`, `hil`, `sin`). A Primary IP can only be assigned to a server in the same location. This is also immutable — changing location forces replacement.

### Optional Reverse DNS

When `dnsPtr` is set, an `hcloud_rdns` resource maps the allocated IP back to the specified hostname. This is required for mail servers (SPF/DKIM verification relies on matching forward and reverse DNS) and any service where clients verify identity via reverse lookup.

### Independent Lifecycle (auto_delete = false)

The IaC modules hardcode `auto_delete = false`. In Planton's component model, resources are managed independently. If a server using this IP is deleted through HetznerCloudServer, the Primary IP remains allocated — it is not silently destroyed along with the server. This is a deliberate design choice: deleting component A must not affect component B.

### Delete Protection

When `deleteProtection` is enabled, the Primary IP cannot be deleted via the Hetzner Cloud API until protection is explicitly removed. This prevents accidental loss of an IP address that may have DNS records, email reputation, or firewall rules associated with it.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the Primary IP from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence. The rDNS resource does not support labels in the Hetzner Cloud API.

## Upstream Dependencies (What This Resource Needs)

None. `HetznerCloudPrimaryIp` is a foundation resource with no foreign key dependencies.

## Downstream Dependents (What References This Resource)

| Dependent | Field | Purpose |
|---|---|---|
| `HetznerCloudServer` | `spec` (primary IP reference) | Assign the IP to a server's primary public IP slot |

## Stack Outputs

| Output | Description |
|---|---|
| `primary_ip_id` | Hetzner Cloud numeric ID of the created Primary IP (as string). Referenced by HetznerCloudServer via StringValueOrRef. |
| `ip_address` | The allocated IP address. For IPv4, a single address (e.g., `203.0.113.42`). For IPv6, the first address in the /64 block (e.g., `2001:db8::1`). |
| `ip_network` | The allocated IPv6 /64 CIDR (e.g., `2001:db8::/64`). Empty for IPv4 Primary IPs. |

## References

- [Hetzner Cloud Primary IPs Documentation](https://docs.hetzner.cloud/#primary-ips)
- [Terraform hcloud_primary_ip Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/primary_ip)
- [Terraform hcloud_rdns Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/rdns)
- [Pulumi hcloud.PrimaryIp Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/primaryip/)
- [Pulumi hcloud.Rdns Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/rdns/)
