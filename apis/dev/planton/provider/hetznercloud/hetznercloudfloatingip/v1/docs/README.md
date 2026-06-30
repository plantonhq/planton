# HetznerCloud Floating IP — Research Documentation

## Introduction

A Hetzner Cloud Floating IP is a reassignable public IP address — either a single IPv4 address or an IPv6 /64 network block — that can be moved between servers in the same location without being destroyed and recreated. The Floating IP persists independently of any server: it can exist unassigned, be attached to a server, detached, and reattached to a different server. This reassignability is the core value proposition — it enables failover patterns where a stable public endpoint survives server replacement with zero DNS propagation delay.

The `HetznerCloudFloatingIp` component provisions a Floating IP at a specific Hetzner Cloud location with an optional server assignment and an optional reverse DNS (rDNS) record. It is a **Phase 2 resource** with no mandatory upstream dependencies. The `server_id` field optionally references a `HetznerCloudServer` via `StringValueOrRef`, but the Floating IP can also be created unassigned and bound to a server later. The component's `floating_ip_id` output can be used for monitoring, DNS configuration, or any downstream automation that needs to reference the IP.

Planton bundles the Floating IP and its optional rDNS record into a single component because rDNS is tightly coupled to the allocated IP address: you cannot set a reverse DNS pointer without knowing the address, and there is no realistic use case for managing rDNS on a Floating IP as a separate resource. The optional server assignment is included on the Floating IP side (rather than the server side) because Hetzner Cloud's API models assignment as a property of the Floating IP — the `server_id` attribute on `hcloud_floating_ip` controls which server holds the IP. This keeps the Floating IP component self-contained: it can allocate, assign, and configure rDNS in a single manifest.

### Floating IP vs Primary IP

Hetzner Cloud offers two types of persistent public IPs, and the distinction matters:

| Aspect | Primary IP | Floating IP |
|--------|-----------|-------------|
| Assignment slot | Replaces the server's auto-assigned primary IP | Additional IP that must be configured in the server's OS |
| Reassignment | Can be moved between servers in the same location | Can be moved between servers in the same location |
| OS configuration | Automatic — the server uses it as its primary interface | Manual — requires IP alias or network configuration inside the server |
| Use case | Stable server identity (the server's "main" IP) | Failover, high availability, service migration |
| Lifecycle | Created independently, assigned at server creation | Created independently, assigned/reassigned at any time |
| Auto-delete | Configurable (`auto_delete` attribute) | Not applicable — Floating IPs never auto-delete |

The key operational difference: when a Primary IP is assigned to a server, the server uses it automatically as its primary public interface. A Floating IP, once assigned, requires OS-level configuration (an IP alias on the network interface) before the server will respond on that address. Hetzner Cloud routes traffic to the assigned server at the infrastructure level, but the server's operating system must be configured to accept packets on the Floating IP address.

## Historical Context

Reassignable public IPs are one of the oldest cloud networking primitives, predating even the term "cloud computing" in its modern sense.

**The problem:** Public IP addresses are external identifiers. DNS records, firewall allow-lists, client configurations, email reputation, SSL certificates (for IP-based TLS), and monitoring systems all reference specific IP addresses. When the server behind that IP fails or needs replacement, changing the IP address triggers a cascade of updates across every system that references it. DNS propagation alone can take hours due to TTL caching. For services that require instant failover — databases, mail servers, payment gateways — waiting for DNS propagation is not an option.

**Elastic IP (AWS, 2008):** AWS introduced Elastic IPs as the first cloud-native answer: a static IPv4 address allocated to an account, attachable to any instance in the same region. The insight was that IP lifecycle and compute lifecycle are separate concerns. GCP followed with "Static External IP Addresses," Azure with "Public IP Addresses" (static allocation). The pattern became universal: allocate an IP independently, attach it to compute, detach and reattach as needed.

**Hetzner Cloud's dual model:** Hetzner Cloud introduced both Primary IPs and Floating IPs, each serving a different purpose. Primary IPs occupy the server's main public IP slot — the address the server is natively reachable on. Floating IPs are secondary addresses that require OS-level configuration but offer the same reassignment capability. This dual model reflects Hetzner's philosophy of exposing the underlying networking mechanics rather than hiding them behind abstraction layers. It gives operators more control at the cost of more configuration.

**The rDNS dimension:** Reverse DNS (mapping IPs back to hostnames) is a per-IP concern, not a per-server concern. In Hetzner Cloud, rDNS is set on the IP resource itself — whether Primary IP or Floating IP. For email deliverability, this is critical: receiving mail servers check that the sending IP's rDNS record matches the sender's domain. Losing an IP means losing its rDNS configuration and, for email, its sending reputation.

**The assignment model:** Hetzner Cloud models Floating IP assignment as a property of the Floating IP, not the server. The `hcloud_floating_ip` resource has a `server_id` attribute; the server resource does not have a `floating_ip_ids` attribute. There is also a separate `hcloud_floating_ip_assignment` resource for managing assignment independently. Planton uses the inline `server_id` attribute on the Floating IP because it covers the common case (assign at creation, update assignment by changing the spec) without introducing a third resource type.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Floating IPs** in the left sidebar (under Networking)
3. Click **Create Floating IP**
4. Select address type: IPv4 or IPv6
5. Select location (e.g., Falkenstein, Nuremberg, Helsinki, Ashburn, Hillsboro, Singapore)
6. Optionally select a server to assign the IP to
7. Enter a description
8. Click **Create**
9. To set rDNS: click the created IP, click the rDNS edit icon, enter the hostname
10. To enable delete protection: click the created IP, go to settings, enable protection

**Pros:**
- Zero tooling required
- Visual location and server selection via dropdowns
- Immediate feedback on the allocated address
- rDNS and protection can be set post-creation through the same UI

**Cons:**
- No audit trail beyond Hetzner's internal logs
- Cannot reproduce across environments or projects
- No version control for IP allocation decisions
- rDNS and protection are separate manual steps from IP creation
- No way to enforce organizational standards (naming, labeling)
- Server assignment is selected from a dropdown — easy to pick the wrong server
- Labels must be added one-by-one through the UI

**Verdict:** Acceptable for learning and one-off experiments. Not suitable for any environment where IP allocations must be reproducible or auditable.

### Level 1: CLI (`hcloud`)

The Hetzner Cloud CLI provides Floating IP management commands:

```bash
# Create an IPv4 Floating IP in Falkenstein
hcloud floating-ip create \
  --name failover-ip \
  --type ipv4 \
  --home-location fsn1 \
  --description "Production failover IP"

# Create and immediately assign to a server
hcloud floating-ip create \
  --name failover-ip \
  --type ipv4 \
  --home-location fsn1 \
  --server my-server-01

# Create an IPv6 Floating IP
hcloud floating-ip create \
  --name web-ipv6 \
  --type ipv6 \
  --home-location nbg1

# Set reverse DNS
hcloud floating-ip set-rdns \
  --ip 203.0.113.42 \
  --hostname mail.example.com \
  failover-ip

# Enable delete protection
hcloud floating-ip enable-protection failover-ip delete

# Assign to a different server (reassignment)
hcloud floating-ip assign failover-ip my-server-02

# Unassign from current server
hcloud floating-ip unassign failover-ip

# Inspect
hcloud floating-ip describe failover-ip

# List all Floating IPs
hcloud floating-ip list

# Add labels
hcloud floating-ip add-label failover-ip env=production

# Disable protection (required before deletion)
hcloud floating-ip disable-protection failover-ip delete

# Delete
hcloud floating-ip delete failover-ip
```

**Pros:**
- Scriptable and reproducible in shell scripts
- Separate commands for creation, assignment, rDNS, and protection
- Assignment and unassignment are explicit operations
- Human-readable output from `describe`
- Label management via `add-label` / `remove-label`

**Cons:**
- No state tracking — cannot detect drift
- Multi-step workflow: create IP, then assign, then set rDNS, then set protection
- No atomic operation for IP + assignment + rDNS
- Shell scripts are fragile across environments
- Assignment changes require manual coordination between the Floating IP and the server's OS configuration

**Verdict:** Good for ad-hoc operations, debugging, and failover scripts that reassign IPs during incidents. Not a management solution for declarative infrastructure.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides three resources for Floating IP management:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.60"
    }
  }
}

# The Floating IP itself
resource "hcloud_floating_ip" "failover" {
  name              = "failover-ip"
  type              = "ipv4"
  home_location     = "fsn1"
  description       = "Production failover IP"
  server_id         = hcloud_server.web.id  # optional inline assignment
  labels = {
    environment = "production"
    purpose     = "failover"
  }
  delete_protection = true
}

# Reverse DNS (optional)
resource "hcloud_rdns" "failover" {
  floating_ip_id = hcloud_floating_ip.failover.id
  ip_address     = hcloud_floating_ip.failover.ip_address
  dns_ptr        = "failover.example.com"
}

output "floating_ip_id" {
  value = hcloud_floating_ip.failover.id
}

output "ip_address" {
  value = hcloud_floating_ip.failover.ip_address
}
```

**Alternative: separate assignment resource:**

```hcl
# Create Floating IP without assignment
resource "hcloud_floating_ip" "failover" {
  name          = "failover-ip"
  type          = "ipv4"
  home_location = "fsn1"
}

# Manage assignment separately
resource "hcloud_floating_ip_assignment" "failover" {
  floating_ip_id = hcloud_floating_ip.failover.id
  server_id      = hcloud_server.web.id
}
```

**Key provider behaviors:**
- `hcloud_floating_ip`: `type` and `home_location` are `ForceNew` — changing either destroys and recreates the IP (allocating a new address). `name`, `description`, `labels`, `server_id`, and `delete_protection` can be updated in-place.
- `hcloud_floating_ip.server_id` vs `hcloud_floating_ip_assignment`: Both control assignment. Using both creates a conflict. The inline `server_id` attribute is simpler for static assignment; the separate `hcloud_floating_ip_assignment` resource is useful when assignment changes independently of the Floating IP's lifecycle.
- `hcloud_rdns`: `ip_address` and the parent ID field (`floating_ip_id`) are `ForceNew`. The `dns_ptr` value can be updated in-place.
- `hcloud_rdns` requires a numeric `floating_ip_id` (int) and the IP address as a string. Both are computed outputs from the Floating IP resource.

**Pros:**
- State tracking and drift detection for IP, assignment, and rDNS
- Plan/apply workflow shows exact changes before they happen
- Explicit dependency: rDNS automatically depends on the Floating IP through attribute references
- Version-controlled IP allocation and assignment decisions
- `delete_protection` is visible and auditable
- Choice between inline assignment and separate assignment resource

**Cons:**
- Two or three separate resources for one logical concept (IP + optional assignment + optional rDNS)
- Must choose between inline `server_id` and `hcloud_floating_ip_assignment` — using both is an error
- No validation that rDNS hostname resolves back to the allocated IP (forward/reverse consistency is the user's responsibility)
- Assignment changes in Terraform trigger an update, but the server's OS still needs to be configured to accept the Floating IP — Terraform cannot do that

**Verdict:** Production-grade for Terraform teams. The standard choice before Planton. The multi-resource model is correct but requires understanding the inline vs separate assignment trade-off.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `FloatingIp`, `FloatingIpAssignment`, and `Rdns`:

```go
package main

import (
    "strconv"

    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        fip, err := hcloud.NewFloatingIp(ctx, "failover", &hcloud.FloatingIpArgs{
            Name:             pulumi.String("failover-ip"),
            Type:             pulumi.String("ipv4"),
            HomeLocation:     pulumi.StringPtr("fsn1"),
            Description:      pulumi.StringPtr("Production failover IP"),
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
            DeleteProtection: pulumi.Bool(true),
        })
        if err != nil {
            return err
        }

        // ID type conversion: FloatingIp.ID() returns IDOutput (string),
        // but RdnsArgs.FloatingIpId expects IntInput.
        floatingIpIdInt := fip.ID().ApplyT(func(id pulumi.ID) (int, error) {
            return strconv.Atoi(string(id))
        }).(pulumi.IntOutput)

        _, err = hcloud.NewRdns(ctx, "failover-rdns", &hcloud.RdnsArgs{
            FloatingIpId: floatingIpIdInt,
            IpAddress:    fip.IpAddress,
            DnsPtr:       pulumi.String("failover.example.com"),
        })
        if err != nil {
            return err
        }

        ctx.Export("floatingIpId", fip.ID())
        ctx.Export("ipAddress", fip.IpAddress)
        return nil
    })
}
```

**The same ID type friction as Primary IPs:** `FloatingIp.ID()` returns `IDOutput` (a string representation of the numeric ID), but `RdnsArgs.FloatingIpId` expects `IntInput`. Every Pulumi user who wants rDNS on a Floating IP must write the `ApplyT` conversion. The Planton Pulumi module handles this once.

**Server assignment via inline `ServerId`:** The `FloatingIpArgs.ServerId` field accepts `IntPtrInput`. When assigning from a server's output (which is a string ID), the same `strconv.Atoi` conversion is needed. This is a second instance of the ID type mismatch pattern in the Hetzner Cloud Pulumi SDK.

**Pros:**
- Full programming language (Go, TypeScript, Python)
- Type safety catches field name errors at compile time
- Built-in secret management for API tokens
- Explicit dependency tracking via output references
- Can implement conditional logic (assign only if server exists) natively

**Cons:**
- ID type mismatch (string ID vs int inputs) adds boilerplate for both rDNS and server assignment
- Two or three separate resource types for one logical concept
- More verbose than HCL for a simple IP allocation
- Smaller community for Hetzner Cloud specifically

**Verdict:** Excellent for Go/TypeScript teams. Planton uses Pulumi (Go) internally for its IaC modules.

## Comparative Analysis

| Method | State Tracking | Drift Detection | Atomic IP+rDNS | Assignment Model | Audit Trail |
|--------|---------------|-----------------|-----------------|------------------|-------------|
| Console | No | No | No (separate steps) | Dropdown selection | Minimal |
| CLI | No | No | No (separate commands) | `assign`/`unassign` commands | No |
| Terraform | Yes | Yes | No (two resources) | Inline `server_id` or separate resource | Via VCS |
| Pulumi | Yes | Yes | No (two resources) | `ServerId` field or separate resource | Via VCS |
| **Planton** | **Yes** | **Yes** | **Yes (single manifest)** | **Inline `serverId` with `StringValueOrRef`** | **Via VCS** |

Planton's key differentiators:

1. **Single manifest**: One YAML declares the IP, its assignment, and its rDNS. No resource wiring.
2. **StringValueOrRef for assignment**: The `serverId` field accepts a literal ID or a `valueFrom` reference to a HetznerCloudServer resource's output. This eliminates hardcoded IDs in infra-chart compositions.
3. **No assignment model choice**: Planton uses the inline `server_id` attribute — there is no separate `hcloud_floating_ip_assignment` resource to choose between. This removes a decision point that has no benefit in Planton's component model.
4. **Safe defaults**: `name` and `labels` are derived from metadata. Delete protection is an explicit opt-in field on the spec.

## The Planton Approach

### Manifest Format

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudFloatingIp
metadata:
  name: failover-ip
  org: acme-corp
  env: production
spec:
  type: ipv4
  homeLocation: fsn1
  description: Production web frontend failover IP
  serverId:
    value: "12345678"
  dnsPtr: failover.example.com
  deleteProtection: true
```

### What Planton Automates

1. **Naming:** The Floating IP name in Hetzner Cloud is derived from `metadata.name`.
2. **Labeling:** Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are computed from metadata and merged with user-specified labels.
3. **Assignment wiring:** The `serverId` field accepts a literal string or a `valueFrom` reference. The IaC module converts the string ID to the integer that the provider expects. Users never deal with type conversions.
4. **Conditional rDNS:** When `dnsPtr` is non-empty, the IaC module creates an `hcloud_rdns` resource automatically, wiring the Floating IP's numeric ID and allocated address without user intervention.
5. **Provider configuration:** The Hetzner Cloud API token is resolved from provider config or `HCLOUD_TOKEN`, not hardcoded in the manifest.
6. **Dual IaC:** The same manifest drives both Pulumi and Terraform backends.

### The 80/20 Principle

The Hetzner Cloud Floating IP API has several attributes. Planton's `HetznerCloudFloatingIpSpec` exposes the attributes that matter for 80% of use cases.

**Included:**
- `type` — IPv4 or IPv6. The most fundamental decision; must be explicit. Changing it forces replacement.
- `homeLocation` — Where the IP is allocated. Must match the target server's location for assignment to work. Changing it forces replacement.
- `description` — Human-readable text visible in the console and API. Useful for documenting purpose without relying on labels.
- `serverId` — Optional server assignment via `StringValueOrRef`. Covers the common case of assigning at creation time and the infra-chart case of referencing a server component's output.
- `dnsPtr` — Reverse DNS hostname. Optional but critical for mail servers and services requiring identity verification.
- `deleteProtection` — Prevents accidental deletion via the API. Important for production IPs that anchor DNS records and email reputation.

**Handled by the platform (hardcoded or derived):**
- `name` — Derived from `metadata.name`. Consistent naming across all components.
- `labels` — Computed from metadata with user labels merged in. Standard labels take precedence.

**Deliberately excluded:**
- `hcloud_floating_ip_assignment` as a separate resource — The inline `server_id` attribute on the Floating IP covers assignment. A separate assignment resource adds complexity (three resources for one concept) without benefit in Planton's model where the Floating IP component owns its own assignment.

### API Design Decisions

**Assignment on the Floating IP, not the Server:** In the Hetzner Cloud API, assignment is modeled as a property of the Floating IP (`server_id` on `hcloud_floating_ip`), not the server. Planton follows this convention. The Floating IP component declares which server it is assigned to. The server component does not declare which Floating IPs are attached. This means:
- Creating a Floating IP with `serverId` assigns it immediately
- Removing `serverId` from the spec unassigns it
- Changing `serverId` reassigns it to a different server (an in-place update, not a replacement)

This is the opposite of how Primary IPs work in Planton: there, the server component references the Primary IP's ID, and assignment happens at server creation. The difference reflects the underlying API: Primary IPs occupy the server's primary IP slot (a server property), while Floating IPs are additional IPs managed independently (a Floating IP property).

**`serverId` as `StringValueOrRef`:** The `server_id` field uses the `StringValueOrRef` type, which accepts either a literal string value or a `valueFrom` reference. This enables two usage patterns:
1. **Standalone deployment:** Set `serverId.value` to a known server ID (as a string).
2. **Infra-chart composition:** Use `serverId.valueFrom` to reference a `HetznerCloudServer` resource's `status.outputs.server_id`, establishing a dependency edge in the deployment DAG.

**rDNS as a simple string, not a sub-message:** The `dnsPtr` field is a flat string on the spec rather than a nested message. For a Floating IP, rDNS is always a single pointer record for the allocated address — there is no complex configuration. A nested message would add structural overhead for no benefit.

**IpType as a proto enum:** The `type` field uses an enum (`ipv4`, `ipv6`) rather than a string. This catches invalid type values at schema validation time rather than at cloud API call time.

**Three outputs:** The component exports `floating_ip_id`, `ip_address`, and `ip_network`. The first is useful for monitoring and automation. The second is needed for DNS record configuration (creating an A record pointing to the Floating IP). The third captures the IPv6 /64 CIDR (empty for IPv4), which is useful for firewall rules that allow traffic to the entire allocated block.

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Created When | Description |
|------------|----------|-------|-------------|-------------|
| Pulumi | `hcloud.FloatingIp` | 1 | Always | Floating IP with type, location, optional server assignment, labels, and protection |
| Pulumi | `hcloud.Rdns` | 0 or 1 | When `dns_ptr` is non-empty | Reverse DNS pointer for the allocated address |
| Terraform | `hcloud_floating_ip` | 1 | Always | Same as Pulumi |
| Terraform | `hcloud_rdns` | 0 or 1 | When `dns_ptr` is non-empty | Conditional via `count` |

Neither engine creates an `hcloud_floating_ip_assignment` resource. The `server_id` attribute on the Floating IP resource handles assignment directly.

### Server ID Type Conversion (Pulumi)

The Pulumi hcloud SDK has a type mismatch: `FloatingIpArgs.ServerId` expects `IntPtrInput`, but the server ID from the spec is a string. The module converts at creation time:

```go
if spec.ServerId != nil && spec.ServerId.GetValue() != "" {
    serverIdInt, err := strconv.Atoi(spec.ServerId.GetValue())
    if err != nil {
        return errors.Wrapf(err, "failed to parse server_id %q as integer", spec.ServerId.GetValue())
    }
    floatingIpArgs.ServerId = pulumi.IntPtr(serverIdInt)
}
```

This conversion happens before resource creation, using the resolved string value from `StringValueOrRef`. If the value cannot be parsed as an integer, the module fails with a clear error message.

### Floating IP ID Type Conversion for rDNS (Pulumi)

A second type mismatch: `RdnsArgs.FloatingIpId` expects `IntInput`, but `FloatingIp.ID()` returns `IDOutput` (string). The module converts after creation:

```go
floatingIpIdInt := createdFloatingIp.ID().ApplyT(func(id pulumi.ID) (int, error) {
    return strconv.Atoi(string(id))
}).(pulumi.IntOutput)
```

This conversion is only needed when `dns_ptr` is non-empty. The `ApplyT` callback runs during Pulumi's deployment phase when the Floating IP's actual ID is known.

### Conditional rDNS

**Pulumi module:** A simple `if spec.DnsPtr != ""` guard. When the condition is false, no rDNS resource is created and no Pulumi resource name is consumed.

**Terraform module:** Uses `count = var.spec.dns_ptr != null && var.spec.dns_ptr != "" ? 1 : 0`. The `hcloud_rdns.this[0]` resource is created only when the condition is met.

Both approaches ensure that a minimal manifest (type + home_location only) creates exactly one resource, while a manifest with `dnsPtr` creates two.

### Server Assignment

**Pulumi module:** The `server_id` field is checked for `nil` and empty string. When present, the string value is converted to an integer and set on `FloatingIpArgs.ServerId`. This is a creation-time setting — on subsequent updates, changing `server_id` triggers an in-place update (reassignment), not a replacement.

**Terraform module:** Uses `server_id = var.spec.server_id != null ? tonumber(var.spec.server_id) : null`. When `null`, the Floating IP is created unassigned.

### Label Management

Both IaC modules apply a standard label set to the `hcloud_floating_ip` resource (rDNS resources do not support labels in the Hetzner Cloud API):

| Label Key | Source | Example |
|-----------|--------|---------|
| `resource` | Constant | `"true"` |
| `name` | `metadata.name` | `"failover-ip"` |
| `kind` | Constant | `"HetznerCloudFloatingIp"` |
| `org` | `metadata.org` | `"acme-corp"` |
| `env` | `metadata.env` | `"production"` |
| `id` | `metadata.id` | `"hcfip-abc123"` |

User-specified `metadata.labels` are merged in, with standard labels taking precedence on key conflicts.

### Dependency Role

`HetznerCloudFloatingIp` is a **Phase 2 resource**. It has one optional upstream dependency:

- `HetznerCloudServer` — When `serverId` references a server via `valueFrom`, the Floating IP depends on the server being created first.

It has no mandatory downstream dependents in the current component catalog. In infra charts, the pattern is:

```
HetznerCloudServer (server creation)
  └── server_id output
        └── HetznerCloudFloatingIp.spec.serverId (via StringValueOrRef)
```

The Floating IP declares the reference to the server, creating a one-way dependency edge. The server does not know which Floating IPs are assigned to it.

## Production Best Practices

### IPv4 vs IPv6

Hetzner Cloud charges for IPv4 Floating IPs (they are a scarce resource globally) but provides IPv6 /64 blocks at no additional cost. Choose based on your requirements:

| Factor | IPv4 | IPv6 |
|--------|------|------|
| Cost | Monthly fee per address | Included |
| Compatibility | Universal | Some clients/networks lack IPv6 support |
| Address count | 1 address | /64 block (~18 quintillion addresses) |
| rDNS scope | One pointer for one address | One pointer per address in the block |
| OS configuration | Single IP alias | /64 block needs routing configuration |
| Use case | Failover for public-facing services, email | Cost-effective failover for IPv6-ready clients |

**Recommendation:** Use IPv4 for services that must be universally reachable (email, public APIs, websites without IPv6 fallback). Use IPv6 when cost matters and all clients support it, or as a secondary address family alongside IPv4.

### Location Selection

A Floating IP can only be assigned to a server in the same location. Plan location before allocation:

| Location Code | City | Region |
|--------------|------|--------|
| `fsn1` | Falkenstein | EU (Germany) |
| `nbg1` | Nuremberg | EU (Germany) |
| `hel1` | Helsinki | EU (Finland) |
| `ash` | Ashburn | US East |
| `hil` | Hillsboro | US West |
| `sin` | Singapore | Asia Pacific |

**Recommendation:** Decide the server location first, then allocate the Floating IP in the same location. If you need to move a service to a different location, you must allocate a new Floating IP — the existing one cannot be relocated.

### Failover Patterns

Floating IPs enable several failover patterns:

**Active-passive failover:** Two servers in the same location. The Floating IP is assigned to the active server. A health check (external script, keepalived, or orchestration tool) detects failure and reassigns the Floating IP to the passive server. The passive server already has the IP alias configured, so it begins responding immediately.

```
Normal:    FloatingIP → Server A (active)    Server B (passive, IP alias configured)
Failover:  FloatingIP → Server B (now active) Server A (failed)
```

**Rolling deployment:** During a deployment, a new server is created with the updated application. The Floating IP is reassigned from the old server to the new server. Once traffic is confirmed healthy on the new server, the old server is destroyed.

**Multi-server services:** Multiple Floating IPs can be assigned to the same server. A server can hold one Floating IP for its web service, another for its mail service, each with its own rDNS record.

### OS Configuration Requirement

Unlike Primary IPs, Floating IPs require OS-level configuration. After assignment, the server must be configured to accept traffic on the Floating IP address. Without this, the server silently drops packets destined for the Floating IP.

**For IPv4 on most Linux distributions:**

```bash
# Add the Floating IP as an alias on the loopback interface
ip addr add 203.0.113.42/32 dev lo
```

**For persistence across reboots**, add the configuration to the network manager (netplan, NetworkManager, or `/etc/network/interfaces` depending on the distribution).

This OS-level configuration is outside Planton's scope — the component manages the IP allocation and infrastructure-level routing (Hetzner Cloud routes packets to the assigned server), but configuring the server's network stack is the user's responsibility.

### Reverse DNS for Email

If the Floating IP will be used by a mail server, rDNS is required for deliverability:

1. **Set `dnsPtr` to the mail server's FQDN** (e.g., `mail.example.com`)
2. **Ensure forward DNS matches:** The A record for `mail.example.com` must resolve to the Floating IP's allocated address
3. **Configure SPF:** Include the Floating IP's address in the domain's SPF record
4. **Forward/reverse consistency:** The rDNS hostname must resolve back to the same IP via forward DNS. Mismatches cause mail rejection.

### Delete Protection

Enable `deleteProtection: true` for any Floating IP that:
- Has DNS records pointing to it (deleting the IP orphans the DNS records)
- Is assigned to a production server (deleting unassigns and removes the IP)
- Has built email sending reputation (new IPs start with neutral reputation)
- Is referenced by external firewall rules or allow-lists
- Is part of a failover configuration (both active and passive servers are configured for it)

Delete protection must be explicitly disabled before the IP can be removed, providing a deliberate two-step teardown that prevents accidental destruction.

### Immutability Considerations

The `type` and `homeLocation` fields are immutable in the Hetzner Cloud API. Changing either in the manifest triggers resource replacement: the old Floating IP is destroyed and a new one is created with a different address. This has cascading effects:

- DNS records pointing to the old address must be updated
- Servers configured with the old IP alias must be reconfigured
- Email reputation associated with the old address is lost
- External firewall rules referencing the old address break
- Any failover configuration referencing the old IP must be updated on all servers

**Recommendation:** Treat `type` and `homeLocation` as permanent decisions. If you need a different type or location, create a new Floating IP alongside the existing one, migrate all configurations (DNS, firewall, OS aliases, failover scripts), and then decommission the old IP.

The `description`, `serverId`, `dnsPtr`, and `deleteProtection` fields can all be updated in-place without triggering replacement.

## References

- [Hetzner Cloud Floating IPs Documentation](https://docs.hetzner.cloud/#floating-ips)
- [Hetzner Cloud API — Floating IPs](https://docs.hetzner.cloud/#floating-ips-get-all-floating-ips)
- [Hetzner Cloud API — Floating IP Actions](https://docs.hetzner.cloud/#floating-ip-actions)
- [Terraform hcloud_floating_ip Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/floating_ip)
- [Terraform hcloud_floating_ip_assignment Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/floating_ip_assignment)
- [Terraform hcloud_rdns Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/rdns)
- [Pulumi hcloud.FloatingIp Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/floatingip/)
- [Pulumi hcloud.FloatingIpAssignment Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/floatingipassignment/)
- [Pulumi hcloud.Rdns Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/rdns/)
