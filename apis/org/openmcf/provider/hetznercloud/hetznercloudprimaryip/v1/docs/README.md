# Hetzner Cloud Primary IP — Research Documentation

## Introduction

A Hetzner Cloud Primary IP is a managed public IP address — either a single IPv4 address or an IPv6 /64 network block — that persists independently of any server. When a server is created with a Primary IP, the server uses that address as its public interface. When the server is deleted, the Primary IP remains allocated to the account and can be assigned to a new server. This decoupling of IP lifecycle from server lifecycle is the core value: stable endpoints that survive infrastructure churn.

The `HetznerCloudPrimaryIp` component provisions a Primary IP at a specific Hetzner Cloud location with an optional reverse DNS (rDNS) record. It is a **foundation resource** with no upstream dependencies. Its `primary_ip_id` output is referenced by `HetznerCloudServer` via `StringValueOrRef` to assign the IP at server creation time. The component does not handle assignment itself — that responsibility belongs to the server component, keeping each resource's lifecycle independent.

OpenMCF bundles the Primary IP and its optional rDNS record into a single component because rDNS is tightly coupled to the IP address: you cannot set a reverse DNS pointer without knowing the allocated address, and there is no use case for managing rDNS on a Primary IP as a separate resource. The conditional creation (rDNS is only created when `dnsPtr` is non-empty) keeps the simple case simple while supporting the mail-server use case in the same manifest.

## Historical Context

Public IP management in cloud infrastructure has progressed through several generations, each addressing the same fundamental problem: IP addresses are external identifiers that clients, DNS records, firewall rules, and reputation systems depend on, but server instances are disposable.

**The ephemeral IP era:** Early cloud instances received a public IP from a pool when they booted and returned it when they stopped. Every server restart meant a new IP. DNS records had to be updated manually. Firewall allow-lists broke. Email reputation (built over months of clean sending behavior) was lost instantly. Operators worked around this with Elastic DNS services, short TTLs, and external load balancers — all adding complexity to solve a problem that the platform should own.

**Reserved IP era (AWS Elastic IP, 2008):** AWS introduced Elastic IPs — static IPv4 addresses that could be allocated to an account and moved between instances. This was the first cloud-native answer to IP persistence. GCP followed with Static External IPs, Azure with Public IP Addresses (static allocation method). The pattern was the same: allocate an IP independently, attach it to compute, detach and reattach as needed. The key insight was that IP lifecycle and compute lifecycle are different concerns.

**Hetzner Cloud's Primary IP model:** Hetzner Cloud introduced Primary IPs as their version of reserved IPs, but with a twist. Every Hetzner Cloud server has a "primary" public IP slot for each address family (IPv4 and IPv6). When you create a server without specifying a Primary IP, Hetzner auto-assigns one from the pool — and by default, that auto-assigned IP is deleted when the server is deleted. A managed Primary IP, created explicitly, replaces this auto-assigned IP: it occupies the same primary slot but has an independent lifecycle. The naming reflects this: it is the server's "primary" IP, but it is not owned by the server.

**The rDNS dimension:** Reverse DNS adds a layer that is unique to IP management. Forward DNS maps names to IPs; reverse DNS maps IPs back to names. For most services, rDNS is optional. For email, it is essential — receiving mail servers check that the sending IP's rDNS record matches the sender's domain (SPF, DKIM, and DMARC all rely on this chain of trust). In Hetzner Cloud, rDNS is set per-IP, not per-server, making it a property of the Primary IP rather than the server that uses it.

**IaC management:** Terraform and Pulumi both support `hcloud_primary_ip` / `hcloud.PrimaryIp` as separate resources, with `hcloud_rdns` / `hcloud.Rdns` as a second resource for reverse DNS. This two-resource model is correct but introduces wiring: the rDNS resource needs the Primary IP's numeric ID and its allocated IP address, both of which are only known after creation. OpenMCF's single-manifest approach handles this wiring internally.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Primary IPs** in the left sidebar (under Networking)
3. Click **Create Primary IP**
4. Select address type: IPv4 or IPv6
5. Select location (e.g., Falkenstein, Nuremberg, Helsinki, Ashburn, Hillsboro, Singapore)
6. Enter a name
7. Click **Create**
8. To set rDNS: click the created IP, go to **Networking** tab, click the rDNS edit icon, enter the hostname

**Pros:**
- Zero tooling required
- Visual location selection via dropdown
- Immediate feedback on the allocated address
- rDNS can be set after creation via the same UI

**Cons:**
- No audit trail beyond Hetzner's internal logs
- Cannot reproduce across environments or projects
- No version control for IP allocation decisions
- rDNS is a separate manual step from IP creation
- No way to enforce organizational standards (naming, labeling, protection)
- Labels must be added one-by-one through the UI

**Verdict:** Acceptable for learning and experiments. Not suitable for any environment where IP allocations must be reproducible or auditable.

### Level 1: CLI (`hcloud`)

The Hetzner Cloud CLI provides Primary IP management commands:

```bash
# Create an IPv4 Primary IP in Falkenstein
hcloud primary-ip create \
  --name mail-ip \
  --type ipv4 \
  --location fsn1 \
  --assignee-type server

# Create an IPv6 Primary IP
hcloud primary-ip create \
  --name web-ipv6 \
  --type ipv6 \
  --location nbg1 \
  --assignee-type server

# Set reverse DNS
hcloud primary-ip set-rdns \
  --ip 203.0.113.42 \
  --hostname mail.example.com \
  mail-ip

# Enable delete protection
hcloud primary-ip enable-protection mail-ip delete

# Inspect
hcloud primary-ip describe mail-ip

# List all Primary IPs
hcloud primary-ip list

# Add labels
hcloud primary-ip add-label mail-ip env=production

# Remove delete protection (required before deletion)
hcloud primary-ip disable-protection mail-ip delete

# Delete
hcloud primary-ip delete mail-ip
```

**Pros:**
- Scriptable
- Separate commands for creation and rDNS (can set rDNS later)
- Human-readable output from `describe`
- Label management via `add-label` / `remove-label`

**Cons:**
- No state tracking — cannot detect drift
- Multi-step workflow: create IP, then set rDNS, then set protection
- No atomic operation for IP + rDNS
- Shell scripts are fragile across environments
- No structured output for downstream resource referencing (server creation needs the IP's numeric ID)

**Verdict:** Good for ad-hoc operations and debugging. Not a management solution for production IP allocations.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides `hcloud_primary_ip` and `hcloud_rdns`:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.60"
    }
  }
}

resource "hcloud_primary_ip" "mail" {
  name              = "mail-ip"
  type              = "ipv4"
  location          = "fsn1"
  assignee_type     = "server"
  auto_delete       = false
  delete_protection = true

  labels = {
    environment = "production"
    purpose     = "mail"
  }
}

resource "hcloud_rdns" "mail" {
  primary_ip_id = hcloud_primary_ip.mail.id
  ip_address    = hcloud_primary_ip.mail.ip_address
  dns_ptr       = "mail.example.com"
}

output "primary_ip_id" {
  value = hcloud_primary_ip.mail.id
}

output "ip_address" {
  value = hcloud_primary_ip.mail.ip_address
}
```

**Key provider behaviors:**
- `hcloud_primary_ip`: `type` and `location` are `ForceNew` — changing either destroys and recreates the IP (and its address changes). `name`, `labels`, `auto_delete`, and `delete_protection` can be updated in-place.
- `hcloud_rdns`: `ip_address` and the parent ID field (`primary_ip_id`) are `ForceNew`. The `dns_ptr` value can be updated in-place.
- `hcloud_rdns` requires a numeric `primary_ip_id` (int) and the IP address as a string. Both are computed outputs from the Primary IP resource.
- `location`, `datacenter`, and `assignee_id` form an `ExactlyOneOf` group — exactly one must be set. `datacenter` is deprecated (removal planned after 2026-07-01).

**Pros:**
- State tracking and drift detection for both IP and rDNS
- Plan/apply workflow shows exact changes before they happen
- Explicit dependency: `hcloud_rdns` automatically depends on `hcloud_primary_ip` through attribute references
- Version-controlled IP allocation decisions
- `auto_delete` and `delete_protection` are visible and auditable

**Cons:**
- Two separate resources for one logical concept (IP with rDNS)
- Must explicitly set `assignee_type = "server"` every time (it is the only valid value but not defaulted)
- Must explicitly set `auto_delete = false` if you want independent lifecycle (the provider defaults to `false` but the intent is not self-documenting)
- No validation that the rDNS hostname resolves back to the allocated IP (forward/reverse DNS consistency is the user's responsibility)

**Verdict:** Production-grade for Terraform teams. The standard choice before OpenMCF. The two-resource model for IP + rDNS is correct but introduces wiring that could be encapsulated.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `PrimaryIp` and `Rdns`:

```go
package main

import (
    "strconv"

    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        ip, err := hcloud.NewPrimaryIp(ctx, "mail", &hcloud.PrimaryIpArgs{
            Name:             pulumi.String("mail-ip"),
            Type:             pulumi.String("ipv4"),
            Location:         pulumi.StringPtr("fsn1"),
            AssigneeType:     pulumi.String("server"),
            AutoDelete:       pulumi.Bool(false),
            DeleteProtection: pulumi.Bool(true),
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
        })
        if err != nil {
            return err
        }

        // ID type conversion: PrimaryIp.ID() returns IDOutput (string),
        // but RdnsArgs.PrimaryIpId expects IntInput.
        primaryIpIdInt := ip.ID().ApplyT(func(id pulumi.ID) (int, error) {
            return strconv.Atoi(string(id))
        }).(pulumi.IntOutput)

        _, err = hcloud.NewRdns(ctx, "mail-rdns", &hcloud.RdnsArgs{
            PrimaryIpId: primaryIpIdInt,
            IpAddress:   ip.IpAddress,
            DnsPtr:      pulumi.String("mail.example.com"),
        })
        if err != nil {
            return err
        }

        ctx.Export("primaryIpId", ip.ID())
        ctx.Export("ipAddress", ip.IpAddress)
        return nil
    })
}
```

**A notable friction point:** The `PrimaryIp` resource's `ID()` returns `IDOutput` (a string representation of the numeric ID), but `RdnsArgs.PrimaryIpId` expects `IntInput`. Every Pulumi user who wants rDNS on a Primary IP must write the `ApplyT` conversion. The OpenMCF Pulumi module handles this once in the resource file.

**Pros:**
- Full programming language (Go, TypeScript, Python)
- Type safety catches field name errors at compile time
- Built-in secret management for API tokens
- Explicit dependency tracking via output references

**Cons:**
- The ID type mismatch (`string` ID vs `int` PrimaryIpId) adds boilerplate for rDNS
- Two separate resource types for one logical concept
- More verbose than HCL for a simple IP allocation
- Smaller community for Hetzner Cloud specifically

**Verdict:** Excellent for Go/TypeScript teams. OpenMCF uses Pulumi (Go) internally for its IaC modules.

## Comparative Analysis

| Method | State Tracking | Drift Detection | Atomic IP+rDNS | Audit Trail | Lifecycle Control |
|--------|---------------|-----------------|-----------------|-------------|-------------------|
| Console | No | No | No (separate steps) | Minimal | Manual |
| CLI | No | No | No (separate commands) | No | Manual |
| Terraform | Yes | Yes | No (two resources) | Via VCS | `auto_delete` + `delete_protection` |
| Pulumi | Yes | Yes | No (two resources) | Via VCS | `auto_delete` + `delete_protection` |
| **OpenMCF** | **Yes** | **Yes** | **Yes (single manifest)** | **Via VCS** | **Hardcoded safe defaults** |

The key differentiators for OpenMCF:

1. **Single manifest**: One YAML declares the IP and its rDNS. No resource wiring.
2. **Safe defaults hardcoded**: `auto_delete=false` and `assignee_type="server"` are not exposed because there is only one correct answer in OpenMCF's component model.
3. **Assignment is the server's concern**: The Primary IP component allocates the IP. The Server component assigns it. Each component manages its own lifecycle without implicit coupling.

## The OpenMCF Approach

### Manifest Format

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: mail-ip
spec:
  type: ipv4
  location: fsn1
  dnsPtr: mail.example.com
  deleteProtection: true
```

### What OpenMCF Automates

1. **Naming:** The Primary IP name in Hetzner Cloud is derived from `metadata.name`.
2. **Labeling:** Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are computed from metadata and merged with user-specified labels.
3. **Safe lifecycle defaults:** `auto_delete` is always `false`. In OpenMCF, resources are managed independently — auto-deletion would silently destroy a separately-managed resource when a server is removed.
4. **Assignee type:** Always `"server"` — the only type Hetzner Cloud supports.
5. **Assignment exclusion:** `assignee_id` is not set. The HetznerCloudServer component handles assignment by referencing `primary_ip_id` via StringValueOrRef.
6. **Conditional rDNS:** When `dnsPtr` is non-empty, the IaC module creates an `hcloud_rdns` resource automatically, wiring the Primary IP's numeric ID and allocated address without user intervention.
7. **Provider configuration:** The Hetzner Cloud API token is resolved from provider config or `HCLOUD_TOKEN`, not hardcoded in the manifest.
8. **Dual IaC:** The same manifest drives both Pulumi and Terraform backends.

### The 80/20 Principle

The Hetzner Cloud Primary IP API has several attributes. OpenMCF's `HetznerCloudPrimaryIpSpec` exposes the attributes that matter for 80% of use cases.

**Included:**
- `type` — IPv4 or IPv6. This is the most fundamental decision and must be explicit.
- `location` — Where the IP is allocated. Must match the server's location for assignment to work.
- `dns_ptr` — Reverse DNS hostname. Optional but critical for mail servers and services that perform reverse lookups.
- `delete_protection` — Prevents accidental deletion via the API. Important for production IPs that anchor DNS records and reputation.

**Handled by the platform (hardcoded):**
- `auto_delete = false` — Resources in OpenMCF are independently managed. Auto-deletion on server removal would silently destroy a separately-managed resource, violating the component model's guarantee that deleting component A does not affect component B.
- `assignee_type = "server"` — The only value Hetzner Cloud supports. Exposing it in the spec would add a required field with exactly one valid answer.
- `name` — Derived from `metadata.name`. Consistent naming across all components.
- `labels` — Computed from metadata with user labels merged in.

**Deliberately excluded:**
- `assignee_id` — Assignment is the server's responsibility, not the IP's. The HetznerCloudServer component references `primary_ip_id` to claim the IP. This separation ensures that IP lifecycle and server lifecycle are managed independently.
- `datacenter` — Deprecated in the provider (removal after 2026-07-01). The `location` field supersedes it. OpenMCF does not expose deprecated fields.

### API Design Decisions

**rDNS as an optional field, not a sub-message:** The `dns_ptr` field is a simple string on the spec rather than a nested message. This reflects the reality that rDNS for a Primary IP is a single pointer record for a single IP address — there is no complex configuration. A nested message would add structural overhead for no benefit.

**IpType as a proto enum:** The `type` field uses an enum (`ipv4`, `ipv6`) rather than a string. This prevents invalid type strings at schema validation time. Terraform accepts any string and rejects invalid values at API call time. OpenMCF rejects them at proto validation.

**No IPv4/IPv6 dual-stack in one resource:** A single Primary IP is either IPv4 or IPv6, never both. Users who need both address families create two HetznerCloudPrimaryIp resources. This matches the Hetzner Cloud API model (one address per Primary IP resource) and keeps the component's semantics unambiguous.

**Three outputs, not one:** The component exports `primary_ip_id`, `ip_address`, and `ip_network`. The server component needs `primary_ip_id` for assignment. Users need `ip_address` for DNS record configuration. The `ip_network` output captures the IPv6 /64 CIDR (empty for IPv4), which is useful for firewall rules that need to allow the entire allocated block.

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Created When | Description |
|------------|----------|-------|-------------|-------------|
| Pulumi | `hcloud.PrimaryIp` | 1 | Always | Primary IP with type, location, labels, and protection settings |
| Pulumi | `hcloud.Rdns` | 0 or 1 | When `dns_ptr` is non-empty | Reverse DNS pointer for the allocated address |
| Terraform | `hcloud_primary_ip` | 1 | Always | Same as Pulumi |
| Terraform | `hcloud_rdns` | 0 or 1 | When `dns_ptr` is non-empty | Conditional via `count` |

### ID Type Conversion (Pulumi)

The Pulumi hcloud SDK has a type mismatch: `PrimaryIp.ID()` returns `IDOutput` (a string representation of the numeric ID), but `RdnsArgs.PrimaryIpId` expects `IntInput`. The module converts once:

```go
primaryIpIdInt := createdPrimaryIp.ID().ApplyT(func(id pulumi.ID) (int, error) {
    return strconv.Atoi(string(id))
}).(pulumi.IntOutput)
```

This conversion is only needed when `dns_ptr` is non-empty. The module checks `spec.DnsPtr != ""` before entering the rDNS creation block.

### Conditional rDNS

**Pulumi module:** A simple `if spec.DnsPtr != ""` guard. When the condition is false, no rDNS resource is created and no Pulumi resource name is consumed.

**Terraform module:** Uses `count = var.spec.dns_ptr != null && var.spec.dns_ptr != "" ? 1 : 0`. The `hcloud_rdns.this[0]` resource is created only when the condition is met.

Both approaches ensure that a minimal manifest (type + location only) creates exactly one resource, while a manifest with `dnsPtr` creates two.

### Label Management

Both IaC modules apply a standard label set to the `hcloud_primary_ip` resource (rDNS resources do not support labels in the Hetzner Cloud API):

| Label Key | Source | Example |
|-----------|--------|---------|
| `resource` | Constant | `"true"` |
| `name` | `metadata.name` | `"mail-ip"` |
| `kind` | Constant | `"HetznerCloudPrimaryIp"` |
| `org` | `metadata.org` | `"acme-corp"` |
| `env` | `metadata.env` | `"production"` |
| `id` | `metadata.id` | `"hcpip-abc123"` |

User-specified `metadata.labels` are merged in, with standard labels taking precedence on key conflicts.

### Dependency Role

`HetznerCloudPrimaryIp` is a **Phase 2 resource** with no upstream dependencies. It is referenced by:

- `HetznerCloudServer` — Assigns the Primary IP to a server at creation time

In infra charts, the pattern is:

```
HetznerCloudPrimaryIp (IP allocation)
  └── primary_ip_id output
        └── HetznerCloudServer.spec (via StringValueOrRef)
```

The Primary IP does not know which server will use it. The server declares the reference, creating a one-way dependency edge.

## Production Best Practices

### IPv4 vs IPv6

Hetzner Cloud charges for IPv4 addresses (they are a scarce resource globally) but provides IPv6 /64 blocks at no additional cost. Choose based on your requirements:

| Factor | IPv4 | IPv6 |
|--------|------|------|
| Cost | Monthly fee per address | Included |
| Compatibility | Universal | Some clients/networks lack IPv6 support |
| Address count | 1 address | /64 block (~18 quintillion addresses) |
| rDNS scope | One pointer for one address | One pointer per address in the block |
| Use case | Public-facing services, email, APIs | Modern services with IPv6-ready clients |

**Recommendation:** Use IPv4 for services that must be reachable from all clients (email, public APIs, websites without IPv6 fallback). Use IPv6 when cost matters and all clients support it, or as a secondary address family alongside IPv4.

### Location Selection

A Primary IP can only be assigned to a server in the same location. Plan location before allocation:

| Location Code | City | Region |
|--------------|------|--------|
| `fsn1` | Falkenstein | EU (Germany) |
| `nbg1` | Nuremberg | EU (Germany) |
| `hel1` | Helsinki | EU (Finland) |
| `ash` | Ashburn | US East |
| `hil` | Hillsboro | US West |
| `sin` | Singapore | Asia Pacific |

**Recommendation:** Decide the server's location first, then allocate the Primary IP in the same location. If you need to move a service to a different location, you must allocate a new Primary IP — the existing one cannot be relocated.

### Reverse DNS for Email

If the Primary IP will be used by a mail server, rDNS is not optional — it is required for deliverability:

1. **Set `dnsPtr` to the mail server's FQDN** (e.g., `mail.example.com`)
2. **Ensure forward DNS matches:** The A record for `mail.example.com` must resolve to the Primary IP's allocated address
3. **Configure SPF:** Include the Primary IP's address in the domain's SPF record
4. **Forward/reverse consistency:** The rDNS hostname must resolve back to the same IP via forward DNS. Mismatches cause mail rejection.

### Delete Protection

Enable `deleteProtection: true` for any Primary IP that:
- Has DNS records pointing to it (deleting the IP orphans the DNS records)
- Is assigned to a production server (deleting unassigns the IP, leaving the server without a public address)
- Has built email sending reputation (new IPs start with neutral reputation; established IPs are valuable)
- Is referenced by external firewall rules or allow-lists

Delete protection must be explicitly disabled before the IP can be removed, providing a deliberate two-step teardown that prevents accidental destruction.

### Immutability Considerations

The `type` and `location` fields are immutable in the Hetzner Cloud API. Changing either in the manifest triggers resource replacement: the old Primary IP is destroyed and a new one is created with a different address. This has cascading effects:

- DNS records pointing to the old address must be updated
- Servers assigned to the old IP lose their public address during the transition
- Email reputation associated with the old address is lost
- External firewall rules referencing the old address break

**Recommendation:** Treat `type` and `location` as permanent decisions. If you need a different type or location, create a new Primary IP alongside the existing one, migrate services, update DNS, and then decommission the old IP.

## References

- [Hetzner Cloud Primary IPs Documentation](https://docs.hetzner.cloud/#primary-ips)
- [Hetzner Cloud API — Primary IPs](https://docs.hetzner.cloud/#primary-ips-get-all-primary-ips)
- [Hetzner Cloud API — Reverse DNS](https://docs.hetzner.cloud/#primary-ips-change-reverse-dns-entry-for-a-primary-ip)
- [Terraform hcloud_primary_ip Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/primary_ip)
- [Terraform hcloud_rdns Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/rdns)
- [Pulumi hcloud.PrimaryIp Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/primaryip/)
- [Pulumi hcloud.Rdns Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/rdns/)
