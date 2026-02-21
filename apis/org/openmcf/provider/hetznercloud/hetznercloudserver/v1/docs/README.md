# HetznerCloud Server — Research Documentation

## Introduction

A Hetzner Cloud Server is a virtual machine running a chosen operating system on a specific hardware profile (vCPU, RAM, disk) in one of Hetzner's six datacenter locations. It is the core compute primitive in the Hetzner Cloud platform and the most-referenced resource in the OpenMCF Hetzner Cloud catalog.

The `HetznerCloudServer` component provisions a server and an optional reverse DNS record. What makes this component distinctive is its role as the **hub resource** — the central node that ties together almost every other Hetzner Cloud component. SSH keys are injected at creation for access. Firewalls are applied at creation for security. Placement groups enforce anti-affinity for reliability. Private networks provide internal communication. Primary IPs provide stable public addressing. And after creation, Volumes, Snapshots, Floating IPs, and Load Balancers all reference the server's ID.

This hub role means the server component's spec is the most connected in the catalog: five foreign key reference fields point to upstream components (`sshKeys`, `placementGroupId`, `firewallIds`, `networks[].networkId`, `publicNet.ipv4/ipv6`), and four downstream components reference its `server_id` output. Every design decision in this component has implications across the resource graph.

OpenMCF bundles only the server and its optional rDNS record into this component. Resources like Volume, Floating IP, and Snapshot have independent lifecycles (a volume should survive server replacement; a floating IP can be reassigned) and are modeled as separate components that reference the server. This keeps the server component focused on compute and allows each dependent resource to be managed independently.

## Historical Context

### The Server as Cloud Building Block

Cloud virtual machines are the oldest and most fundamental cloud computing primitive, dating back to Amazon EC2's launch in 2006. The core abstraction has remained remarkably stable: select a hardware profile, choose an OS image, pick a region, and get a running machine with a public IP address. What has grown complex over the years is not the server itself but the web of resources that surround it.

Modern server provisioning involves:
- **Identity and access**: SSH keys, IAM roles, service accounts
- **Networking**: VPCs, subnets, network interfaces, private IPs, public IPs, floating IPs
- **Security**: Security groups, firewalls, network ACLs
- **Storage**: Block volumes, snapshots, backups
- **Scheduling**: Placement groups, availability zones, anti-affinity rules
- **Initialization**: Cloud-init, user data, startup scripts

This web of dependencies is what makes server provisioning complex — not the server creation itself.

### Hetzner Cloud's Position

Hetzner Cloud launched in 2018 as a European alternative to the hyperscalers, targeting developers and small-to-medium teams with a straightforward product lineup and aggressive pricing. Key characteristics:

- **Focused product set**: ~28 provider resources total (vs. thousands in AWS/GCP/Azure)
- **European data sovereignty**: Datacenters in Germany (Falkenstein, Nuremberg), Finland (Helsinki), and expanding to US (Ashburn, Hillsboro) and Asia (Singapore)
- **Price-competitive**: Shared vCPU servers start at €3.29/month for 2 vCPU / 2 GB RAM
- **ARM64 support**: Ampere-based `cax` server types at lower cost per vCPU
- **No enterprise complexity**: No IAM system, no VPC peering, no transit gateways. Networks are flat, API tokens are per-project, firewalls are simple allow/deny rules.

This simplicity is both Hetzner Cloud's strength and its design constraint. There are no resource policies, no tagging enforcement, no centralized security controls. What you configure explicitly is what you get.

### The Server-as-Hub Pattern

In Hetzner Cloud's resource model, the server sits at the center of a dependency graph:

```
                HetznerCloudSshKey (ssh_key_ids)
                         │
                         ▼
HetznerCloudFirewall ──► SERVER ◄── HetznerCloudPlacementGroup
                         │  ▲
                         │  │
            ┌────────────┘  └───────────────┐
            ▼                               │
    HetznerCloudNetwork            HetznerCloudPrimaryIp
    (private networking)           (stable public IPs)
            │
            ▼
    ┌───────┴───────┐
    │   Downstream  │
    │   resources   │
    ├───────────────┤
    │ Volume        │
    │ Snapshot      │
    │ FloatingIp    │
    │ LoadBalancer  │
    └───────────────┘
```

Every arrow in this graph is a foreign key reference. In raw Terraform or Pulumi, each reference is a numeric ID that must be obtained from another resource's output and often converted between string and integer types. OpenMCF's `StringValueOrRef` abstraction eliminates this wiring for users.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Servers** in the left sidebar
3. Click **Add Server**
4. Select a location (Falkenstein, Nuremberg, Helsinki, Ashburn, Hillsboro, Singapore)
5. Choose an OS image from the catalog (Ubuntu, Debian, Fedora, Rocky, etc.)
6. Select a server type (Shared, Dedicated, ARM)
7. Choose networking options:
   - Attach to an existing network (private networking)
   - Enable/disable public IPv4 and IPv6
   - Select existing Primary IPs
8. Add SSH keys from the project's key list
9. Select firewalls
10. Select a placement group
11. Optionally add a cloud-config script in the user data field
12. Enable or disable backups
13. Name the server and add labels
14. Click **Create & Buy now**

After creation, to set reverse DNS: navigate to the server's Networking tab, click the rDNS edit icon next to the IPv4 address, and enter the hostname.

**Pros:**
- Zero tooling required
- Visual server type selection with pricing information
- Immediate feedback with a live console
- Good for initial exploration and learning

**Cons:**
- No audit trail for configuration decisions
- Cannot reproduce across environments (no version control)
- Every dependency (SSH key, firewall, network) must exist before clicking "Create" — no way to express relationships
- Post-creation steps (rDNS, protections) are separate manual actions
- No enforcement of organizational standards (naming, labeling)
- Server type selection is visual but imprecise for automation

**Verdict:** Useful for learning and one-off experiments. Not suitable for any environment where server configurations must be reproducible or auditable.

### Level 1: CLI (`hcloud`)

The Hetzner Cloud CLI provides comprehensive server management:

```bash
# Create a minimal server
hcloud server create \
  --name web-01 \
  --type cx22 \
  --image ubuntu-24.04 \
  --location fsn1

# Create a server with all options
hcloud server create \
  --name web-01 \
  --type cx22 \
  --image ubuntu-24.04 \
  --location fsn1 \
  --ssh-key deploy-key \
  --ssh-key backup-key \
  --firewall web-fw \
  --placement-group ha-spread \
  --network main-vpc \
  --user-data-from-file cloud-init.yaml \
  --backups \
  --label env=production \
  --label team=platform

# Set reverse DNS
hcloud server set-rdns \
  --ip 203.0.113.42 \
  --hostname web-01.example.com \
  web-01

# Enable protections
hcloud server enable-protection web-01 delete rebuild

# Attach to a network with specific IP
hcloud server attach-to-network \
  --network main-vpc \
  --ip 10.0.1.10 \
  web-01

# Resize (server must be stopped)
hcloud server change-type --keep-disk web-01 cpx31

# Power management
hcloud server shutdown web-01     # graceful ACPI shutdown
hcloud server poweroff web-01     # hard power off
hcloud server poweron web-01

# Inspect
hcloud server describe web-01
hcloud server list

# Delete (must disable protection first)
hcloud server disable-protection web-01 delete rebuild
hcloud server delete web-01
```

**Pros:**
- Scriptable and reproducible in shell scripts
- Single command for creation with all options as flags
- Separate commands for post-creation actions (rDNS, protections, network attachment)
- Human-readable output from `describe`

**Cons:**
- No state tracking — cannot detect drift between intended and actual configuration
- Dependencies (SSH key, firewall, network, placement group) must exist and be referenced by name — no ID resolution
- Multi-step workflow: create server, then set rDNS, then set protections
- No atomic operation for server + rDNS + protections
- Shell scripts are fragile across environments and operating systems
- No plan/preview — changes execute immediately

**Verdict:** Good for ad-hoc operations, scripted deployments in CI, and operational tasks (resize, restart). Not a declarative management solution.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides the `hcloud_server` resource with dynamic blocks for public networking and private network attachments:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.60"
    }
  }
}

resource "hcloud_ssh_key" "deploy" {
  name       = "deploy-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

resource "hcloud_firewall" "web" {
  name = "web-fw"
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "443"
    source_ips = ["0.0.0.0/0", "::/0"]
  }
}

resource "hcloud_network" "main" {
  name     = "main-vpc"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "servers" {
  network_id   = hcloud_network.main.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

resource "hcloud_placement_group" "ha" {
  name = "ha-spread"
  type = "spread"
}

resource "hcloud_server" "web" {
  name                   = "web-01"
  server_type            = "cx22"
  image                  = "ubuntu-24.04"
  location               = "fsn1"
  ssh_keys               = [hcloud_ssh_key.deploy.id]
  placement_group_id     = hcloud_placement_group.ha.id
  firewall_ids           = [hcloud_firewall.web.id]
  backups                = true
  keep_disk              = true
  delete_protection      = true
  rebuild_protection     = true
  shutdown_before_deletion = true
  labels = {
    environment = "production"
  }

  user_data = file("cloud-init.yaml")

  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }

  dynamic "network" {
    for_each = [hcloud_network.main.id]
    content {
      network_id = network.value
      ip         = "10.0.1.10"
      alias_ips  = []
    }
  }

  depends_on = [hcloud_network_subnet.servers]
}

# Conditional rDNS
resource "hcloud_rdns" "web" {
  server_id  = hcloud_server.web.id
  ip_address = hcloud_server.web.ipv4_address
  dns_ptr    = "web-01.example.com"
}

output "server_id" {
  value = hcloud_server.web.id
}

output "ipv4_address" {
  value = hcloud_server.web.ipv4_address
}
```

**Key provider behaviors:**
- `hcloud_server`: `image` and `location` are `ForceNew` — changing either destroys and recreates the server. `server_type` changes trigger an in-place resize (server is stopped temporarily). `ssh_keys` is `ForceNew`.
- `public_net`: When omitted entirely, the server gets auto-assigned IPv4 and IPv6. When present, each field (`ipv4_enabled`, `ipv6_enabled`, `ipv4`, `ipv6`) must be explicitly set.
- `network`: Uses dynamic blocks. The `network_id` attribute requires an integer. `alias_ips` must be explicitly set (even to an empty list) to avoid drift detection issues.
- `hcloud_rdns`: The `server_id` field is an integer. `ip_address` and `server_id` are `ForceNew`.
- `depends_on` is needed for network attachments: the subnet must exist before a server can attach to the network, but there is no direct attribute reference from server to subnet.

**Pros:**
- Full state tracking and drift detection
- Plan/apply workflow previews changes before execution
- Automatic dependency resolution via attribute references (server depends on SSH key, firewall, etc.)
- Version-controlled configuration
- `dynamic` blocks handle conditional and repeated structures

**Cons:**
- Five or more separate resources for a typical server deployment (SSH key + firewall + network + subnet + placement group + server + rDNS)
- Must manage `depends_on` manually for subnet-before-network-attachment ordering
- `alias_ips` must always be passed (even empty) to avoid Terraform detecting drift on every apply — a known provider quirk
- `public_net` has two different behaviors depending on whether the block is present or absent — omitting it entirely is different from including an empty block
- rDNS is a separate resource that must be manually wired with `server_id` and `ip_address`
- All ID references are integers, requiring `tonumber()` conversions when IDs arrive as strings from other state sources

**Verdict:** Production-grade for Terraform teams. The standard choice before OpenMCF. The complexity is proportional to the server's hub role — wiring five dependencies and managing conditional resources requires understanding the provider's behavior.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `Server`, `Rdns`, and all dependency resources:

```go
package main

import (
    "strconv"

    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        sshKey, err := hcloud.NewSshKey(ctx, "deploy", &hcloud.SshKeyArgs{
            Name:      pulumi.String("deploy-key"),
            PublicKey: pulumi.String("ssh-ed25519 AAAA..."),
        })
        if err != nil {
            return err
        }

        fw, err := hcloud.NewFirewall(ctx, "web", &hcloud.FirewallArgs{
            Name: pulumi.String("web-fw"),
            Rules: hcloud.FirewallRuleArray{
                &hcloud.FirewallRuleArgs{
                    Direction: pulumi.String("in"),
                    Protocol:  pulumi.String("tcp"),
                    Port:      pulumi.StringPtr("443"),
                    SourceIps: pulumi.StringArray{
                        pulumi.String("0.0.0.0/0"),
                        pulumi.String("::/0"),
                    },
                },
            },
        })
        if err != nil {
            return err
        }

        pg, err := hcloud.NewPlacementGroup(ctx, "ha", &hcloud.PlacementGroupArgs{
            Name: pulumi.String("ha-spread"),
            Type: pulumi.String("spread"),
        })
        if err != nil {
            return err
        }

        net, err := hcloud.NewNetwork(ctx, "main", &hcloud.NetworkArgs{
            Name:    pulumi.String("main-vpc"),
            IpRange: pulumi.String("10.0.0.0/16"),
        })
        if err != nil {
            return err
        }

        _, err = hcloud.NewNetworkSubnet(ctx, "servers", &hcloud.NetworkSubnetArgs{
            NetworkId:   net.ID().ApplyT(func(id pulumi.ID) (int, error) {
                return strconv.Atoi(string(id))
            }).(pulumi.IntOutput),
            Type:        pulumi.String("cloud"),
            NetworkZone: pulumi.String("eu-central"),
            IpRange:     pulumi.String("10.0.1.0/24"),
        })
        if err != nil {
            return err
        }

        // Firewall IDs: server expects IntArray, but firewall ID is IDOutput (string)
        fwIdInt := fw.ID().ApplyT(func(id pulumi.ID) (int, error) {
            return strconv.Atoi(string(id))
        }).(pulumi.IntOutput)

        pgIdInt := pg.ID().ApplyT(func(id pulumi.ID) (int, error) {
            return strconv.Atoi(string(id))
        }).(pulumi.IntOutput)

        netIdInt := net.ID().ApplyT(func(id pulumi.ID) (int, error) {
            return strconv.Atoi(string(id))
        }).(pulumi.IntOutput)

        srv, err := hcloud.NewServer(ctx, "web", &hcloud.ServerArgs{
            Name:                   pulumi.String("web-01"),
            ServerType:             pulumi.String("cx22"),
            Image:                  pulumi.StringPtr("ubuntu-24.04"),
            Location:               pulumi.StringPtr("fsn1"),
            SshKeys:                pulumi.StringArray{sshKey.ID()},
            PlacementGroupId:       pgIdInt,
            FirewallIds:            pulumi.IntArray{fwIdInt},
            Backups:                pulumi.BoolPtr(true),
            KeepDisk:               pulumi.BoolPtr(true),
            DeleteProtection:       pulumi.BoolPtr(true),
            RebuildProtection:      pulumi.BoolPtr(true),
            ShutdownBeforeDeletion: pulumi.BoolPtr(true),
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
            PublicNets: hcloud.ServerPublicNetArray{
                &hcloud.ServerPublicNetArgs{
                    Ipv4Enabled: pulumi.BoolPtr(true),
                    Ipv6Enabled: pulumi.BoolPtr(true),
                },
            },
            Networks: hcloud.ServerNetworkTypeArray{
                hcloud.ServerNetworkTypeArgs{
                    NetworkId: netIdInt,
                    Ip:        pulumi.StringPtr("10.0.1.10"),
                    AliasIps:  pulumi.StringArray{}, // must be set to avoid drift
                },
            },
        })
        if err != nil {
            return err
        }

        // rDNS requires server ID as int
        srvIdInt := srv.ID().ApplyT(func(id pulumi.ID) (int, error) {
            return strconv.Atoi(string(id))
        }).(pulumi.IntOutput)

        _, err = hcloud.NewRdns(ctx, "web-rdns", &hcloud.RdnsArgs{
            ServerId:  srvIdInt,
            IpAddress: srv.Ipv4Address,
            DnsPtr:    pulumi.String("web-01.example.com"),
        })
        if err != nil {
            return err
        }

        ctx.Export("serverId", srv.ID())
        ctx.Export("ipv4Address", srv.Ipv4Address)
        return nil
    })
}
```

**The ID type mismatch pattern is pervasive.** In the example above, there are **five** `ApplyT(strconv.Atoi)` conversions — one each for the firewall, placement group, network, subnet's network ID, and the server's own ID for rDNS. Every Pulumi user writing Hetzner Cloud server code must write these conversions. The SDK returns `IDOutput` (string) but resource arguments expect `IntInput` for ID fields.

**The AliasIps workaround:** `Networks[].AliasIps` must always be set (even to an empty array) because the Terraform bridge has a bug (#650) where an unset `alias_ips` causes Pulumi to detect drift and attempt a network detach/reattach on every `pulumi up`. This is a known issue that the OpenMCF module handles by always passing `AliasIps`.

**The PublicNets nil-vs-present distinction:** When `PublicNets` is not set on `ServerArgs`, the provider uses its default behavior (auto-assign IPv4 + IPv6). When it is set, each field must be explicitly configured. An empty `ServerPublicNetArgs{}` disables both IPv4 and IPv6 because the fields default to `false`. This is a footgun that catches users who set the block but forget to explicitly enable the protocols.

**Pros:**
- Full programming language (Go, TypeScript, Python) with compile-time type checking
- Automatic dependency tracking via output references
- Built-in secret management for API tokens
- Conditional logic is native language constructs (not HCL `dynamic` blocks)
- Can implement complex orchestration (wait for server health check before creating rDNS)

**Cons:**
- Five ID type conversions add significant boilerplate
- The AliasIps bug requires a workaround that is not documented in the SDK
- More verbose than HCL for a simple server — the Go example is ~100 lines vs ~60 lines in HCL
- Smaller Hetzner Cloud community means fewer examples and Stack Overflow answers
- The PublicNets nil-vs-present distinction is the same footgun as Terraform

**Verdict:** Excellent for Go/TypeScript teams. OpenMCF uses Pulumi (Go) internally for its IaC modules.

## Comparative Analysis

| Aspect | Console | CLI | Terraform | Pulumi | OpenMCF |
|--------|---------|-----|-----------|--------|---------|
| State tracking | No | No | Yes | Yes | Yes |
| Drift detection | No | No | Yes | Yes | Yes |
| Dependency resolution | Manual (prereqs must exist) | Manual | Automatic via refs | Automatic via outputs | Automatic via `valueFrom` |
| ID type handling | N/A | By name | Integer attributes | `ApplyT(strconv.Atoi)` x5 | Automatic |
| Conditional rDNS | Separate manual step | Separate command | Separate resource + `count` | Separate resource + `if` | Single `dnsPtr` field |
| Public net configuration | Checkboxes | Flags | Dynamic block | PublicNetArgs | `publicNet` sub-message |
| Network attachment | Dropdown | `attach-to-network` | Dynamic block | NetworkTypeArray | `networks[]` array |
| AliasIps workaround | N/A | N/A | Must pass empty list | Must pass empty array | Handled automatically |
| Audit trail | Minimal | Via shell history | Via VCS | Via VCS | Via VCS |

OpenMCF's key differentiators for the server resource:

1. **Single manifest, five dependency types**: One YAML declares the server with all its references to SSH keys, firewalls, placement groups, networks, and Primary IPs. No resource wiring.
2. **StringValueOrRef for all references**: Every ID field accepts a literal value or a `valueFrom` reference. Users never convert between string and integer types.
3. **Conditional rDNS as a field, not a resource**: The `dnsPtr` field on the spec conditionally creates the rDNS resource. No separate resource declaration, no ID wiring.
4. **PublicNet defaults handled correctly**: The IaC module implements `optional bool` with default `true` for `ipv4Enabled`/`ipv6Enabled`. Omitting the `publicNet` block entirely preserves the provider's auto-assign behavior.
5. **AliasIps bug handled once**: The Pulumi module always passes `AliasIps` (even when empty), so users never encounter the bridge bug.

## The OpenMCF Approach

### Manifest Format

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: web-01
  org: acme-corp
  env: production
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  sshKeys:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: prod-key
        fieldPath: status.outputs.ssh_key_id
  firewallIds:
    - valueFrom:
        kind: HetznerCloudFirewall
        name: web-fw
        fieldPath: status.outputs.firewall_id
  placementGroupId:
    valueFrom:
      kind: HetznerCloudPlacementGroup
      name: ha-spread
      fieldPath: status.outputs.placement_group_id
  networks:
    - networkId:
        valueFrom:
          kind: HetznerCloudNetwork
          name: main-vpc
          fieldPath: status.outputs.network_id
      ip: "10.0.1.10"
  publicNet:
    ipv4Enabled: true
    ipv6Enabled: true
  backups: true
  keepDisk: true
  deleteProtection: true
  rebuildProtection: true
  shutdownBeforeDeletion: true
  dnsPtr: web-01.example.com
```

### What OpenMCF Automates

1. **Naming**: The server name in Hetzner Cloud is derived from `metadata.name`.
2. **Labeling**: Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are computed from metadata and merged with user-specified labels (standard labels take precedence).
3. **ID type conversions**: Five string-to-integer conversions happen inside the IaC module: `placementGroupId`, each `firewallId`, each `networkId`, `publicNet.ipv4`, `publicNet.ipv6`, and the server's own ID for the rDNS resource. Users pass strings; the module handles the rest.
4. **Conditional rDNS**: When `dnsPtr` is non-empty, the module creates an `hcloud_rdns` resource wired to the server's IPv4 address and numeric ID. No user intervention.
5. **PublicNet defaults**: The `optional bool` fields `ipv4Enabled` and `ipv6Enabled` default to `true` when the `publicNet` block is present but the fields are unset. Omitting `publicNet` entirely preserves the provider's auto-assign behavior.
6. **AliasIps workaround**: The Pulumi module always passes `AliasIps` to network attachments (even when empty) to avoid the Terraform bridge drift bug (#650).
7. **Provider configuration**: The Hetzner Cloud API token is resolved from provider config or `HCLOUD_TOKEN`, not hardcoded in the manifest.
8. **Dual IaC**: The same manifest drives both Pulumi and Terraform backends.

### The 80/20 Principle

The Hetzner Cloud server API has many attributes. OpenMCF exposes those that matter for 80% of use cases.

**Included — Required fields:**

| Field | Rationale |
|-------|-----------|
| `serverType` | The most fundamental choice: determines vCPU, RAM, disk, and cost. Must be explicit. |
| `image` | Determines the operating system. Must be explicit. |
| `location` | Determines the datacenter. Must be explicit because Primary IPs and Floating IPs must be in the same location. |

**Included — Optional fields:**

| Field | Rationale |
|-------|-----------|
| `sshKeys` | The standard secure access method. SSH key injection at creation is a one-time operation. |
| `userData` | Cloud-init is the standard first-boot configuration mechanism. Essential for automated server setup. |
| `placementGroupId` | Anti-affinity scheduling is critical for HA deployments. |
| `firewallIds` | Infrastructure-level security applied at creation. |
| `publicNet` | Fine-grained control over public networking, including Primary IP attachment. |
| `networks` | Private networking is essential for multi-server architectures. |
| `backups` | One-toggle backup enablement. |
| `keepDisk` | Prevents irreversible disk upgrades during server type changes. |
| `deleteProtection` | Prevents accidental deletion. |
| `rebuildProtection` | Prevents accidental re-imaging. |
| `shutdownBeforeDeletion` | Graceful shutdown for clean teardown. |
| `dnsPtr` | Reverse DNS for mail servers and identity verification. |

**Handled by the platform (hardcoded or derived):**
- `name` — Derived from `metadata.name`.
- `labels` — Computed from metadata per CG01 pattern.

**Deliberately excluded:**

| Field | Rationale |
|-------|-----------|
| `iso` | Mounting ISOs is a runtime operational action (emergency boot, custom OS installation). Not declarative IaC. |
| `rescue` | Rescue mode is an emergency recovery action. Declaring it in a spec would put the server in rescue on every apply. |
| `allow_deprecated_images` | Safety valve. Better to surface deprecated image errors explicitly than to silently allow them. |
| `ignore_remote_firewall_ids` | Terraform-specific drift suppression flag. OpenMCF's IaC module manages firewall IDs declaratively — there is no drift to suppress. |
| `datacenter` | Deprecated by Hetzner Cloud (removal after 2026-07-01). Use `location` instead. |

### API Design Decisions

**PublicNet as a sub-message with optional bool defaults:**

The `publicNet` field is an optional sub-message. When omitted entirely, the IaC module does not set `PublicNets` at all, preserving the provider's default (auto-assigned IPv4 + IPv6). When set, the `ipv4Enabled` and `ipv6Enabled` fields use `optional bool` with a documented default of `true`.

Why `optional bool`? Proto3 `bool` defaults to `false`. If `PublicNet` were set with plain `bool` fields, an empty `publicNet: {}` message would disable both IPv4 and IPv6 — almost certainly not the user's intent. The `optional bool` wrapper allows the IaC module to distinguish between "explicitly set to false" and "not set" (defaulting to `true`).

**NetworkAttachment as a repeated sub-message:**

Each `NetworkAttachment` bundles `networkId` (required), `ip` (optional), and `aliasIps` (optional). This mirrors the Terraform `network` dynamic block structure. The `networkId` is a `StringValueOrRef` with `default_kind = HetznerCloudNetwork`, enabling `valueFrom` references.

**SSH keys as StringValueOrRef[] with name support:**

Unlike other foreign key fields that reference numeric IDs, `sshKeys` accepts SSH key names or IDs as strings. The Hetzner Cloud provider accepts both formats for `ssh_keys`. This means users can reference keys by human-readable name (`value: "deploy-key"`) or by output ID (`valueFrom` referencing `ssh_key_id`). The `default_kind = HetznerCloudSshKey` annotation enables `valueFrom` shorthand.

**Four outputs:**

The component exports `server_id` (for downstream references), `ipv4_address` (for DNS configuration), `ipv6_address` (for DNS configuration), and `status` (for monitoring and health checks). The `server_id` output is the most-referenced output in the catalog — used by Volume, Snapshot, FloatingIp, and LoadBalancer components.

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Created When | Description |
|------------|----------|-------|--------------|-------------|
| Pulumi | `hcloud.Server` | 1 | Always | Server with all spec fields mapped |
| Pulumi | `hcloud.Rdns` | 0 or 1 | When `dnsPtr` is non-empty | Reverse DNS for auto-assigned IPv4 |
| Terraform | `hcloud_server` | 1 | Always | Same as Pulumi |
| Terraform | `hcloud_rdns` | 0 or 1 | When `dns_ptr` is non-empty | Conditional via `count` |

### ID Type Conversions (Pulumi Module)

The Pulumi hcloud SDK requires integer inputs where the spec and other component outputs provide strings. The module performs five categories of conversion:

| Conversion | Input Source | Target | Method |
|------------|-------------|--------|--------|
| Placement group ID | `spec.PlacementGroupId.GetValue()` | `ServerArgs.PlacementGroupId` (IntPtr) | `strconv.Atoi` at creation time |
| Firewall IDs | `spec.FirewallIds[].GetValue()` | `ServerArgs.FirewallIds` (IntArray) | Loop + `strconv.Atoi` at creation time |
| Network IDs | `net.NetworkId.GetValue()` | `ServerNetworkTypeArgs.NetworkId` (Int) | `strconv.Atoi` at creation time |
| Primary IP IDs | `pn.Ipv4.GetValue()`, `pn.Ipv6.GetValue()` | `ServerPublicNetArgs.Ipv4/Ipv6` (IntPtr) | `strconv.Atoi` at creation time |
| Server ID for rDNS | `createdServer.ID()` | `RdnsArgs.ServerId` (IntOutput) | `ApplyT(strconv.Atoi)` at deployment time |

The first four conversions use plain `strconv.Atoi` because the values are known before resource creation (resolved from `StringValueOrRef` during stack input loading). The fifth uses Pulumi's `ApplyT` because it depends on the server's actual ID, which is only available after the server is created.

### The AliasIps Bridge Bug Workaround

The Pulumi hcloud SDK (bridged from Terraform) has a bug (#650) where omitting `AliasIps` from `ServerNetworkTypeArgs` causes Pulumi to detect phantom drift on every `pulumi up` — it sees the network attachment as changed and attempts a detach/reattach cycle. The module works around this by always setting `AliasIps`:

```go
args := hcloud.ServerNetworkTypeArgs{
    NetworkId: pulumi.Int(networkId),
    AliasIps:  pulumi.ToStringArray(net.AliasIps), // always pass, even when empty
}
```

When `net.AliasIps` is nil or empty, this passes an empty array — which is different from not setting the field at all. This workaround ensures stable plans with no phantom changes.

### PublicNet Nil-vs-Present Semantics

The module distinguishes between "publicNet is nil" and "publicNet is set":

- **`spec.PublicNet == nil`**: The module does not set `ServerArgs.PublicNets` at all. The provider uses its default behavior: auto-assign public IPv4 and IPv6.
- **`spec.PublicNet != nil`**: The module calls `buildPublicNet()`, which explicitly sets `Ipv4Enabled` and `Ipv6Enabled`. If the `optional bool` fields are nil, they default to `true`. If explicitly set to `false`, IPv4/IPv6 are disabled.

This two-level check prevents the footgun where an empty `publicNet: {}` in the manifest would accidentally disable all public networking.

### Terraform Dynamic Blocks

The Terraform module uses two `dynamic` blocks:

1. **`public_net`**: `for_each = var.spec.public_net != null ? [var.spec.public_net] : []` — creates the block only when `public_net` is set in the input. Inside, each field is null-checked with a default.

2. **`network`**: `for_each = var.spec.networks != null ? { for n in var.spec.networks : n.network_id => n } : {}` — iterates over network attachments, keyed by `network_id`. Uses `tonumber()` to convert string IDs to integers.

### Label Management

Both modules apply standard labels using the CG01 pattern:

| Label Key | Source | Example |
|-----------|--------|---------|
| `resource` | Constant | `"true"` |
| `name` | `metadata.name` | `"web-01"` |
| `kind` | Constant | `"HetznerCloudServer"` |
| `org` | `metadata.org` | `"acme-corp"` |
| `env` | `metadata.env` | `"production"` |
| `id` | `metadata.id` | `"hcsrv-abc123"` |

User-specified `metadata.labels` are merged in. Standard labels take precedence on key conflicts. Labels are applied only to the server resource — the rDNS resource does not support labels in the Hetzner Cloud API.

## Production Best Practices

### Server Type Selection

| Workload | Recommended Family | Rationale |
|----------|-------------------|-----------|
| Web servers, APIs | `cx` or `cpx` (shared x86) | Cost-effective for bursty workloads with moderate CPU needs |
| Build servers, CI runners | `cpx` (shared AMD) | Good single-core performance, cost-effective |
| ARM-native applications | `cax` (shared ARM64) | Lower cost per vCPU, good for Go/Rust/Java workloads compiled for ARM |
| Databases, search engines | `ccx` (dedicated x86) | Guaranteed CPU performance, no noisy-neighbor effects |
| Memory-intensive workloads | `ccx` (dedicated x86, larger types) | Higher RAM-to-vCPU ratio on larger types |

Always start with `keepDisk: true` if you plan to iterate on the server type. Once a disk is upgraded (larger type without `keepDisk`), the server cannot be downgraded to a type with a smaller disk.

### Image Selection

- **Prefer named images** (`ubuntu-24.04`, `debian-12`) over snapshot IDs for reproducibility. Named images are maintained by Hetzner and receive security updates.
- **Use snapshot IDs** when deploying from a custom base image (created via Packer or manual snapshot). Note that snapshots are location-independent — you can create a snapshot in `fsn1` and deploy from it in `hel1`.
- **Avoid deprecated images.** The provider raises an error when deploying a deprecated image. If you encounter this, update to the latest release of the same distribution.

### Cloud-Init Patterns

**Shell scripts** (`#!/bin/bash`) are simplest for small setups:
```bash
#!/bin/bash
apt-get update && apt-get install -y nginx
systemctl enable nginx
```

**Cloud-config YAML** (`#cloud-config`) is better for structured configuration:
```yaml
#cloud-config
package_update: true
packages:
  - nginx
  - certbot
write_files:
  - path: /etc/nginx/sites-available/default
    content: |
      server { listen 80; return 301 https://$host$request_uri; }
```

**Size limit:** 32 KB. For larger configurations, use cloud-init to download a script from object storage or a Git repository.

**Immutability:** Changing `userData` forces server replacement. For iterative configuration management, use cloud-init only for bootstrapping (install agent, join cluster) and manage ongoing configuration with Ansible, Chef, or similar tools.

### SSH Key Management

- **Inject at creation only.** SSH keys in the spec are written to `~/.ssh/authorized_keys` during first boot. Post-creation key management requires direct server access.
- **Use multiple keys** for team access. Each team member's key can be a separate `HetznerCloudSshKey` component referenced via `valueFrom`.
- **Rotate keys** by creating a new key component, updating the server spec, and accepting the server replacement. Or manage ongoing key rotation via cloud-init that pulls keys from a central source.

### Firewall Strategy

- **Apply firewalls at creation.** Firewalls attached via `firewallIds` take effect immediately. No window where the server is exposed.
- **Default-deny pattern:** Start with a firewall that allows only SSH (port 22) from your management IPs. Add application ports as needed.
- **Multiple firewalls** can be combined. Use one firewall for SSH access (applied to all servers) and a separate firewall for application-specific ports (applied per role).

### Private Networking

- **Deploy the network before the server.** The `valueFrom` reference ensures this ordering automatically.
- **Use fixed IPs** (`ip` field on NetworkAttachment) for servers that need stable internal addresses — databases, DNS servers, configuration endpoints.
- **Use auto-assigned IPs** (omit `ip` field) for stateless servers where the private IP doesn't matter — web servers behind a load balancer.
- **Alias IPs** allow a single server to listen on multiple private addresses within the same network, useful for running multiple services or IP-based virtual hosting.

### Backup Strategy

- **Enable backups** (`backups: true`) for any server with local state that would be costly to reconstruct — databases, file servers, configuration stores.
- **Don't rely solely on backups** for critical data. Hetzner Cloud backups are daily snapshots retained for 14 days. For finer granularity, use application-level backups (pg_dump, mysqldump) to object storage.
- **Cost consideration:** Backups add 20% to the server price. For servers running stateless applications (web servers, API gateways), backups are usually unnecessary.

### Protection Settings

- **Enable `deleteProtection`** for any server that:
  - Hosts production workloads
  - Has volumes attached (detach volumes before deletion)
  - Is referenced by Floating IPs or Load Balancer targets
  - Has accumulated local state that cannot be recreated

- **Enable `rebuildProtection`** for servers where re-imaging would cause data loss. This is separate from delete protection — a server can be protected from deletion but still vulnerable to an accidental rebuild.

- **Enable `shutdownBeforeDeletion`** for servers running databases or applications that buffer writes. This sends an ACPI shutdown signal and waits for the OS to shut down cleanly before deletion.

### Reverse DNS Guidelines

- **Only use `dnsPtr` when the server has auto-assigned IPv4.** If you attach a Primary IP via `publicNet.ipv4`, manage rDNS on the `HetznerCloudPrimaryIp` component instead.
- **Match forward and reverse DNS.** The `dnsPtr` hostname must have an A record pointing back to the server's IPv4 address. Mismatches cause mail delivery failures and identity verification issues.
- **Required for mail servers.** Outbound email from an IP without matching rDNS is rejected by most receiving mail servers.

### Immutability Awareness

Several fields force server replacement when changed:

| Field | Change Behavior |
|-------|----------------|
| `image` | **Replacement** — new server with new OS |
| `location` | **Replacement** — server is destroyed and recreated in the new location |
| `sshKeys` | **Replacement** — keys are injected at creation only |
| `userData` | **Replacement** — cloud-init runs at first boot only |
| `serverType` | **In-place resize** — server is stopped, resized, restarted |
| `backups` | **In-place update** |
| `deleteProtection` | **In-place update** |
| `firewallIds` | **In-place update** |
| All other fields | **In-place update** |

Plan for immutable field changes by ensuring critical data is on separate volumes (which survive server replacement) and that the server can be reprovisioned from its spec without manual intervention.

## References

- [Hetzner Cloud Servers Documentation](https://docs.hetzner.cloud/#servers)
- [Hetzner Cloud Server Types](https://docs.hetzner.cloud/#server-types)
- [Hetzner Cloud Images](https://docs.hetzner.cloud/#images)
- [Hetzner Cloud API — Servers](https://docs.hetzner.cloud/#servers-get-all-servers)
- [Hetzner Cloud API — Server Actions](https://docs.hetzner.cloud/#server-actions)
- [Terraform hcloud_server Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/server)
- [Terraform hcloud_rdns Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/rdns)
- [Pulumi hcloud.Server Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/server/)
- [Pulumi hcloud.Rdns Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/rdns/)
- [Cloud-Init Documentation](https://cloudinit.readthedocs.io/)
