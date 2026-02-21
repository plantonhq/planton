# HetznerCloud Firewall — Research Documentation

## Introduction

A Hetzner Cloud firewall is an account-level set of rules that controls inbound and outbound network traffic for servers. When applied to a server, the firewall enforces a **deny-by-default inbound** policy: every incoming packet that does not match at least one allow rule is silently dropped. Outbound traffic is allowed by default unless explicitly restricted by outbound rules. This model is stateful — return traffic for an established connection is automatically permitted regardless of rule direction.

The `HetznerCloudFirewall` component creates a single `hcloud_firewall` resource with inline rules. It is a **foundation resource**: it has no dependencies, but it is referenced by `HetznerCloudServer` via `firewall_ids`. Every production server deployment on Hetzner Cloud should have at least one firewall restricting access to required ports.

OpenMCF exposes a single spec field — `rules` — because the firewall resource has exactly one user-controlled behavioral attribute beyond naming and labeling. Each rule is a structured message with direction, protocol, optional port, source or destination CIDRs, and an optional description. The spec uses proto enums for direction and protocol (not free-form strings) and CEL validation rules to enforce constraints that the Hetzner Cloud API would otherwise reject at apply time — catching errors at validation rather than during deployment.

## Historical Context

Network access control for cloud servers has evolved through several distinct phases, each driven by the increasing scale and complexity of infrastructure.

**iptables era:** Linux administrators wrote iptables rules directly on each server. Rules were either applied manually via SSH or baked into provisioning scripts. There was no centralized management — each server had its own rule set, and auditing meant SSH-ing into every host. A mistake in a rule could lock out the administrator entirely. Configuration drift was the norm, not the exception.

**Security group era:** AWS popularized the concept of security groups in 2009 — account-level firewall definitions applied to instances at launch. This shifted firewall management from the OS level to the cloud API. Google Cloud followed with VPC firewall rules, Azure with Network Security Groups. The key insight was that firewalls are infrastructure metadata, not server configuration. They belong in the control plane, not on the data plane.

**Hetzner Cloud's approach:** Hetzner Cloud introduced firewalls as a relatively late addition to their API. The model is simpler than AWS security groups: there are no separate ingress/egress rule resources, no security group chaining, and no VPC-level vs instance-level distinction. A Hetzner firewall is a flat list of up to 50 rules, each specifying a direction, protocol, optional port, and CIDR blocks. Firewalls are applied to servers either directly (via `firewall_ids` at server creation) or through the `apply_to` mechanism (label selectors or server IDs on the firewall itself). This simplicity is a feature — fewer moving parts, fewer edge cases, fewer ways to misconfigure.

**IaC era:** Terraform and Pulumi brought version control and drift detection to firewall management. Rules are declared in code, reviewed in pull requests, and applied through CI pipelines. But each team writes their own module with their own conventions for naming, labeling, and structuring rules.

**OpenMCF approach:** A standardized manifest format where rules are declared inline, the firewall name comes from `metadata.name`, labels are computed from metadata, and the `firewall_id` output feeds into `HetznerCloudServer.spec.firewallIds` via `StringValueOrRef`. The firewall does not know which servers use it — servers reference firewalls, keeping the dependency graph unidirectional and composable.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Firewalls** in the left sidebar
3. Click **Create Firewall**
4. Enter a name
5. Add rules one by one:
   - Select direction (Inbound / Outbound)
   - Select protocol (TCP, UDP, ICMP, ESP, GRE)
   - Enter port or port range (for TCP/UDP)
   - Enter source or destination IPs (CIDR notation)
   - Optionally add a description
6. Optionally apply to servers or label selectors
7. Click **Create Firewall**

**Pros:**
- Zero tooling required
- Visual rule builder prevents some syntax errors
- Immediate visual confirmation of applied rules

**Cons:**
- No audit trail beyond Hetzner's internal logs
- No version control — impossible to review rule changes
- Rules cannot be reproduced across environments
- Easy to forget rules when recreating in a new project
- No way to enforce organizational standards for firewall configurations
- Adding rules one-by-one is tedious for complex policies

**Verdict:** Acceptable for personal projects and quick experiments. Not suitable for any environment where security policy must be auditable or reproducible.

### Level 1: CLI (`hcloud`)

The Hetzner Cloud CLI provides firewall management commands:

```bash
# Create a firewall
hcloud firewall create --name web-firewall

# Add an inbound rule (SSH)
hcloud firewall add-rule web-firewall \
  --direction in \
  --protocol tcp \
  --port 22 \
  --source-ips 0.0.0.0/0 \
  --source-ips ::/0 \
  --description "allow SSH"

# Add an inbound rule (HTTP)
hcloud firewall add-rule web-firewall \
  --direction in \
  --protocol tcp \
  --port 80 \
  --source-ips 0.0.0.0/0 \
  --source-ips ::/0 \
  --description "allow HTTP"

# Add an outbound rule
hcloud firewall add-rule web-firewall \
  --direction out \
  --protocol tcp \
  --port any \
  --destination-ips 0.0.0.0/0 \
  --destination-ips ::/0 \
  --description "allow all outbound TCP"

# List firewalls
hcloud firewall list

# Describe (shows rules and applied-to resources)
hcloud firewall describe web-firewall

# Apply to a server
hcloud firewall apply-to-resource web-firewall \
  --type server --server my-server

# Delete a rule (by index)
hcloud firewall delete-rule web-firewall --direction in --protocol tcp --port 80 \
  --source-ips 0.0.0.0/0 --source-ips ::/0

# Delete the firewall
hcloud firewall delete web-firewall
```

**Pros:**
- Scriptable
- Full access to all attributes including `apply-to`
- Rules can be added incrementally
- Fast for ad-hoc operations and debugging

**Cons:**
- No state tracking — cannot detect drift
- Rules are added one at a time (no atomic multi-rule creation)
- Deleting a rule requires specifying all its attributes exactly
- Shell scripts are fragile across environments
- No structured output referencing for downstream resources

**Verdict:** Good for quick operations, debugging, and verifying firewall state. Not a management solution for production security policies.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides the `hcloud_firewall` resource with inline `rule` blocks:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.60"
    }
  }
}

resource "hcloud_firewall" "web" {
  name = "web-firewall"

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = ["0.0.0.0/0", "::/0"]
    description = "allow SSH"
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "80"
    source_ips = ["0.0.0.0/0", "::/0"]
    description = "allow HTTP"
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "443"
    source_ips = ["0.0.0.0/0", "::/0"]
    description = "allow HTTPS"
  }

  rule {
    direction  = "in"
    protocol   = "icmp"
    source_ips = ["0.0.0.0/0", "::/0"]
    description = "allow ping"
  }

  labels = {
    environment = "production"
    role        = "web"
  }
}

resource "hcloud_server" "web" {
  name        = "web-01"
  server_type = "cx22"
  image       = "ubuntu-24.04"
  firewall_ids = [hcloud_firewall.web.id]
}

output "firewall_id" {
  value = hcloud_firewall.web.id
}
```

**Attributes:**
- `name` (required) — Display name in Hetzner Cloud
- `rule` (optional, repeatable block) — Inline firewall rules
- `labels` (optional) — Key-value metadata map
- `apply_to` (optional, repeatable block) — Server IDs or label selectors

**Rule block attributes:**
- `direction` (required) — `"in"` or `"out"`
- `protocol` (required) — `"tcp"`, `"udp"`, `"icmp"`, `"esp"`, `"gre"`
- `port` (optional) — Port, range, or `"any"` (required for TCP/UDP)
- `source_ips` (optional) — CIDR blocks for inbound rules
- `destination_ips` (optional) — CIDR blocks for outbound rules
- `description` (optional) — Human-readable description

**Computed:**
- `id` — Hetzner Cloud numeric ID

**Behavior:**
- Adding, removing, or modifying rules triggers an in-place update (not replacement)
- Changing `name` or `labels` triggers an in-place update
- Rules are a `TypeSet`, so ordering in HCL does not matter — Terraform compares rules by content

**The `hcloud_firewall_attachment` resource** is an alternative to `apply_to` for managing server-to-firewall bindings outside the firewall resource. Only one attachment resource per firewall is allowed.

**Pros:**
- State tracking and drift detection
- Plan/apply workflow shows exact rule changes before applying
- Atomic: all rules are applied as a unit
- Direct reference from server resources via `firewall_ids`
- Version-controlled security policy

**Cons:**
- Requires HCL knowledge
- State management overhead
- `rule` blocks use `TypeSet` — diffs can be hard to read when multiple rules change simultaneously
- No built-in organizational conventions for rule structuring

**Verdict:** Production-grade for Terraform teams. The standard choice before OpenMCF.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `hcloud.Firewall` with `FirewallRuleArray`:

```go
package main

import (
    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        fw, err := hcloud.NewFirewall(ctx, "web-firewall", &hcloud.FirewallArgs{
            Name: pulumi.String("web-firewall"),
            Rules: hcloud.FirewallRuleArray{
                &hcloud.FirewallRuleArgs{
                    Direction:   pulumi.String("in"),
                    Protocol:    pulumi.String("tcp"),
                    Port:        pulumi.String("22"),
                    SourceIps:   pulumi.StringArray{pulumi.String("0.0.0.0/0"), pulumi.String("::/0")},
                    Description: pulumi.String("allow SSH"),
                },
                &hcloud.FirewallRuleArgs{
                    Direction:   pulumi.String("in"),
                    Protocol:    pulumi.String("tcp"),
                    Port:        pulumi.String("80"),
                    SourceIps:   pulumi.StringArray{pulumi.String("0.0.0.0/0"), pulumi.String("::/0")},
                    Description: pulumi.String("allow HTTP"),
                },
                &hcloud.FirewallRuleArgs{
                    Direction:   pulumi.String("in"),
                    Protocol:    pulumi.String("icmp"),
                    SourceIps:   pulumi.StringArray{pulumi.String("0.0.0.0/0"), pulumi.String("::/0")},
                    Description: pulumi.String("allow ping"),
                },
            },
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
        })
        if err != nil {
            return err
        }

        ctx.Export("firewallId", fw.ID())
        return nil
    })
}
```

**Pros:**
- Full programming language (Go, TypeScript, Python)
- Type safety: `FirewallRuleArgs` catches field name typos at compile time
- Programmatic rule generation (e.g., loop over a list of ports)
- Built-in secret management for tokens

**Cons:**
- More verbose than HCL for static rule lists
- Requires programming skills
- Smaller community for Hetzner Cloud specifically

**Verdict:** Excellent for Go/TypeScript teams and dynamic rule generation. OpenMCF uses Pulumi (Go) internally for its IaC modules.

## Comparative Analysis

| Method | State Tracking | Drift Detection | Atomic Apply | Audit Trail | Rule Validation |
|--------|---------------|-----------------|-------------|-------------|-----------------|
| Console | No | No | No | Minimal | UI only |
| CLI | No | No | No | No | API-level |
| Terraform | Yes | Yes | Yes | Via VCS | API-level |
| Pulumi | Yes | Yes | Yes | Via VCS | API-level |
| **OpenMCF** | **Yes** | **Yes** | **Yes** | **Via VCS** | **Proto + CEL** |

The key differentiator for OpenMCF is **proto-level validation with CEL rules**. Terraform and Pulumi catch errors at the Hetzner Cloud API level (during `apply`/`up`). OpenMCF catches three classes of errors at manifest validation time, before any API call:

1. `port` is required when protocol is `tcp` or `udp`
2. `source_ips` is required when direction is `in`
3. `destination_ips` is required when direction is `out`

This shifts error feedback from minutes (API round-trip) to milliseconds (local validation).

## The OpenMCF Approach

### Manifest Format

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: web-firewall
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow SSH"
    - direction: in
      protocol: tcp
      port: "443"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow HTTPS"
```

### What OpenMCF Automates

1. **Naming:** The firewall name in Hetzner Cloud is derived from `metadata.name` — no separate `name` field in the spec.
2. **Labeling:** Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are computed from metadata and merged with user-specified labels.
3. **Provider configuration:** Hetzner Cloud API token is resolved from provider config or environment variables, not hardcoded in the manifest.
4. **Dual IaC:** The same manifest drives both Pulumi and Terraform backends.
5. **Output referencing:** The `firewall_id` output feeds into `HetznerCloudServer.spec.firewallIds` via `StringValueOrRef`, enabling declarative composition without hardcoded IDs.
6. **Validation:** CEL rules enforce port/protocol and direction/CIDR constraints at manifest validation time.

### The 80/20 Principle

The Hetzner Cloud firewall API has 4 user-controllable attributes: `name`, `rules`, `labels`, and `apply_to`. OpenMCF's `HetznerCloudFirewallSpec` exposes 1 field: `rules`.

**Included:**
- `rules` — The firewall rule list. This is the only attribute that varies meaningfully per firewall. Each rule is a structured message with direction, protocol, port, CIDRs, and description.

**Handled by the platform:**
- `name` — Derived from `metadata.name`. Consistent naming across all components.
- `labels` — Computed from metadata (org, env, kind, id) with user labels merged in. Consistent labeling across all Hetzner Cloud resources.

**Deliberately excluded:**
- `apply_to` — The Hetzner Cloud API supports applying firewalls to servers via the firewall's `apply_to` field (server IDs or label selectors). OpenMCF inverts this: servers reference firewalls via `firewall_ids` in `HetznerCloudServer.spec`. This keeps the dependency graph unidirectional — firewalls are foundation resources with no forward references to compute resources. In infra charts, the DAG flows naturally: firewall → server, never server ← firewall.

### API Design Decisions

**Enum types for direction and protocol:** The spec uses proto enums (`Direction`, `Protocol`) instead of free-form strings. This provides compile-time validation — an invalid direction like `"inbound"` is caught by the proto schema, not at the Hetzner Cloud API level. The enum values (`in`/`out`, `icmp`/`tcp`/`udp`/`esp`/`gre`) match the Hetzner Cloud API's string values exactly.

**CEL validation rules:** Three cross-field constraints are enforced at the proto level:
- `port_required_for_tcp_udp`: Port must be non-empty when protocol is `tcp` or `udp`.
- `source_ips_required_for_inbound`: `source_ips` must contain at least one entry when direction is `in`.
- `destination_ips_required_for_outbound`: `destination_ips` must contain at least one entry when direction is `out`.

These catch the three most common firewall misconfiguration errors before any cloud API call.

**Empty rules list:** An empty `rules` list is valid. It creates a firewall that, when applied to a server, blocks all inbound traffic and allows all outbound traffic. This is useful as a "lockdown" firewall for servers that should only initiate outbound connections (e.g., pull-based deployment agents).

**Port as string:** The `port` field is a string, not an integer, because it supports three formats: single port (`"80"`), range (`"80-443"`), and the keyword `"any"`. Encoding this as a proto string matches the Hetzner Cloud API's representation directly.

**No `apply_to` field:** As discussed above, server-to-firewall binding is handled by the server component, not the firewall component. This is a deliberate architectural choice — firewalls do not need to know which servers use them.

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Description |
|------------|----------|-------|-------------|
| Pulumi | `hcloud.Firewall` | 1 | Firewall with inline rules |
| Terraform | `hcloud_firewall` | 1 | Firewall with dynamic rule blocks |

This is a single-resource component. The `hcloud_firewall_attachment` resource is not used because server-to-firewall binding is managed by the server component.

### Dependency Role

`HetznerCloudFirewall` is a **foundation resource** — it has no foreign key dependencies. It is referenced by:

- `HetznerCloudServer.spec.firewallIds` — Servers apply this firewall at creation time

In infra charts, the pattern is:

```
HetznerCloudFirewall (foundation)
  └── firewall_id output
        └── HetznerCloudServer.spec.firewallIds (via StringValueOrRef)
```

### Label Management

Both IaC modules apply a standard label set to the Hetzner Cloud firewall resource:

| Label Key | Source | Example |
|-----------|--------|---------|
| `planton-ai_resource` | Constant | `"true"` |
| `planton-ai_name` | `metadata.name` | `"web-firewall"` |
| `planton-ai_kind` | Constant | `"HetznerCloudFirewall"` |
| `planton-ai_org` | `metadata.org` | `"my-org"` |
| `planton-ai_env` | `metadata.env` | `"production"` |
| `planton-ai_id` | `metadata.id` | `"hcfw-abc123"` |

User-specified `metadata.labels` are merged in, with standard labels taking precedence in case of key conflicts.

### Rule Mapping

The Pulumi module maps proto rules to `hcloud.FirewallRuleArgs` in a loop:

1. `direction` and `protocol` enum values are converted to strings via `.String()`.
2. `port` is set only when non-empty (ICMP, ESP, GRE rules have no port).
3. `source_ips` and `destination_ips` are set only when the respective slice is non-empty.
4. `description` is set only when non-empty.

The Terraform module uses a `dynamic "rule"` block that iterates over `var.spec.rules`, passing all attributes directly. Optional attributes (`port`, `source_ips`, `destination_ips`, `description`) accept `null` gracefully.

## Production Best Practices

### Stateful Firewall Behavior

Hetzner Cloud firewalls are **stateful**. When an inbound rule allows traffic on port 443, the return traffic (server responses) is automatically allowed without requiring an outbound rule. Similarly, outbound traffic initiated by the server automatically permits the return inbound traffic.

This means:
- You do **not** need matching inbound and outbound rules for the same connection.
- An inbound-only firewall (no outbound rules) allows the server to respond to all permitted inbound connections **and** initiate any outbound connection.
- Outbound rules are only needed to **restrict** traffic the server initiates (e.g., preventing a compromised server from reaching the internet).

### Dual-Stack CIDRs

Always include both IPv4 and IPv6 CIDR blocks in rules. Hetzner Cloud assigns both IPv4 and IPv6 addresses to servers by default. A rule that only specifies `"0.0.0.0/0"` without `"::/0"` leaves IPv6 traffic unaffected by that rule.

```yaml
# Allow SSH from anywhere (both IPv4 and IPv6)
- direction: in
  protocol: tcp
  port: "22"
  sourceIps:
    - "0.0.0.0/0"
    - "::/0"
```

For restricted access, use specific CIDRs for both address families:

```yaml
# Allow PostgreSQL from private subnets only
- direction: in
  protocol: tcp
  port: "5432"
  sourceIps:
    - "10.0.0.0/16"
    - "fd00::/64"
```

### The 50-Rule Limit

Hetzner Cloud enforces a maximum of 50 rules per firewall. For most applications, this is more than sufficient. If you approach the limit:

1. **Consolidate port ranges:** Use `"80-443"` instead of separate rules for ports 80, 443, and everything in between.
2. **Use CIDR aggregation:** Combine adjacent CIDRs (e.g., `10.0.0.0/8` instead of multiple `/24` subnets).
3. **Separate concerns:** Create multiple firewalls — one for common rules (SSH, ICMP), another for application-specific rules. Servers can reference multiple `firewallIds`.
4. **Use `"any"` for port:** When all ports of a protocol need to be open, use `port: "any"` instead of specifying ranges.

### Principle of Least Privilege

Start with the minimum required rules and add more only as needed:

1. **SSH access:** Restrict to known IP ranges rather than `0.0.0.0/0` when possible.
2. **Application ports:** Open only the ports your application actually listens on.
3. **ICMP:** Allow ICMP for diagnostics (ping, traceroute, PMTU discovery). Blocking ICMP entirely causes subtle connectivity issues.
4. **Outbound restrictions:** Consider adding outbound rules for high-security environments to prevent data exfiltration from compromised servers.

### Protocol-Specific Notes

**ESP (Encapsulating Security Payload):** Used for IPsec VPN tunnels. If your servers participate in site-to-site VPNs, you need ESP rules. No port field — ESP operates at the IP layer.

**GRE (Generic Routing Encapsulation):** Used for tunneling protocols (e.g., PPTP VPN, overlay networks). No port field. Less common in modern deployments, but required for certain legacy VPN configurations.

**ICMP:** No port field. Allowing ICMP inbound enables ping and traceroute. The Hetzner Cloud firewall does not distinguish between ICMP types (echo request, echo reply, destination unreachable) — it is all-or-nothing for the protocol.

### Firewall per Role, Not per Server

Define firewalls by role (web, database, bastion) rather than per server. Servers in the same role share the same firewall rules:

```yaml
# web-firewall.yaml — shared by all web servers
metadata:
  name: web-firewall
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "80"
      sourceIps: ["0.0.0.0/0", "::/0"]
    - direction: in
      protocol: tcp
      port: "443"
      sourceIps: ["0.0.0.0/0", "::/0"]
```

This reduces duplication and ensures consistent security policy across servers in the same tier.

### Empty Firewall as Lockdown

A firewall with no rules blocks all inbound traffic when applied to a server. Use this pattern for servers that should only initiate outbound connections:

```yaml
metadata:
  name: lockdown
spec:
  rules: []
```

When applied, the server can still reach the internet (outbound is allowed by default) but cannot receive any unsolicited inbound traffic.

## References

- [Hetzner Cloud Firewalls Documentation](https://docs.hetzner.cloud/#firewalls)
- [Hetzner Cloud API — Firewalls](https://docs.hetzner.cloud/#firewalls-get-all-firewalls)
- [Terraform hcloud_firewall Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/firewall)
- [Terraform hcloud_firewall_attachment Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/firewall_attachment)
- [Pulumi hcloud.Firewall Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/firewall/)
- [Pulumi hcloud.FirewallAttachment Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/firewallattachment/)
