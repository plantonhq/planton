# HetznerCloud Placement Group — Research Documentation

## Introduction

A placement group controls the physical distribution of servers across Hetzner Cloud's infrastructure. When servers are assigned to a placement group with a `spread` strategy, the platform guarantees they run on different physical hosts. If the datacenter cannot satisfy the anti-affinity constraint, server creation fails rather than silently co-locating — a deliberate design that prevents hidden single-points-of-failure.

The `HetznerCloudPlacementGroup` component creates a single `hcloud_placement_group` resource. It is a **foundation resource**: no dependencies, but referenced by `HetznerCloudServer` (and future infra charts) via `placement_group_id`. Every high-availability server deployment on Hetzner Cloud starts with a placement group.

Planton exposes exactly one optional spec field — `type` — because the placement group resource has exactly one user-controllable behavior attribute. The name comes from `metadata.name`, labels are computed from metadata, and the numeric ID is a computed output. The `type` field defaults to `spread`, which is currently the only strategy Hetzner Cloud supports. This makes the component almost zero-configuration: a manifest with an empty `spec` block is valid and production-ready.

## Historical Context

Anti-affinity in cloud computing emerged from a simple observation: putting all your servers on the same physical host means a single hardware failure takes everything down.

**Pre-cloud era:** Operators negotiated rack placement with datacenter staff. "Don't put the database and the app server in the same chassis" was communicated verbally or tracked in spreadsheets. There was no API, no automation, and no enforcement — just trust and tribal knowledge.

**Early cloud era:** Public cloud providers (AWS, GCP) introduced placement groups as first-class API objects. AWS launched Spread Placement Groups in 2017, allowing users to request that instances land on distinct hardware. Hetzner Cloud followed with its own placement group API, offering a simpler model: one strategy (`spread`), one constraint (max 10 servers per group), no partition or cluster strategies.

**IaC era:** Terraform and Pulumi made placement groups declarative. Instead of imperative API calls, teams define the group in code and reference it from server resources. This brought version control, drift detection, and reproducibility — but each team still wrote their own module with their own conventions.

**Planton approach:** A standardized manifest format that works across both Pulumi and Terraform backends. The placement group is declared once, its `placement_group_id` output is referenced by server components through `StringValueOrRef`, and the entire lifecycle is handled through `planton apply`. The near-empty spec reflects that placement groups have almost no configuration surface — the value is in the reference chain, not the resource itself.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to the project
3. Click **Placement Groups** in the left sidebar
4. Click **Create Placement Group**
5. Enter a name
6. Select type: **Spread** (only option)
7. Add labels (optional)
8. Click **Create Placement Group**

When creating a server, select the placement group from a dropdown.

**Pros:**
- Zero tooling required
- Immediate visual confirmation of group membership

**Cons:**
- No audit trail beyond Hetzner's internal logs
- No version control — cannot review or reproduce changes
- Naming and labeling conventions vary by operator
- Server-to-group assignments are easy to forget during manual server creation

**Verdict:** Acceptable for experimentation. Not suitable for any workload where HA matters — the exact workloads that need placement groups.

### Level 1: CLI (`hcloud`)

```bash
# Create
hcloud placement-group create --name ha-db-group --type spread

# List
hcloud placement-group list

# Describe (shows member servers)
hcloud placement-group describe ha-db-group

# Add labels
hcloud placement-group add-label ha-db-group env=production

# Delete
hcloud placement-group delete ha-db-group

# Use when creating a server
hcloud server create --name db-01 \
  --type cx22 \
  --image ubuntu-24.04 \
  --placement-group ha-db-group
```

**Pros:**
- Scriptable
- Full access to all attributes
- Fast for ad-hoc operations

**Cons:**
- No state tracking — cannot detect drift
- No dependency awareness (nothing prevents deleting a group with active servers referencing it)
- Shell scripts are fragile across environments

**Verdict:** Good for quick operations and debugging. Not a management solution for production HA infrastructure.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides the `hcloud_placement_group` resource:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.60"
    }
  }
}

resource "hcloud_placement_group" "ha_db" {
  name = "ha-db-group"
  type = "spread"
  labels = {
    environment = "production"
    role        = "database"
  }
}

resource "hcloud_server" "db" {
  count              = 3
  name               = "db-${count.index + 1}"
  server_type        = "cx22"
  image              = "ubuntu-24.04"
  placement_group_id = hcloud_placement_group.ha_db.id
}

output "placement_group_id" {
  value = hcloud_placement_group.ha_db.id
}
```

**Attributes:**
- `name` (required) — Display name in Hetzner Cloud
- `type` (required) — Strategy type; only `"spread"` is supported
- `labels` (optional) — Key-value metadata map

**Computed:**
- `id` — Hetzner Cloud numeric ID
- `servers` — List of server IDs currently in the group

**Behavior:**
- Changing `name` or `labels` triggers an in-place update
- Changing `type` triggers resource replacement (destroy + create), which requires all member servers to be removed first
- The `servers` attribute is read-only and reflects current group membership

**Pros:**
- State tracking and drift detection
- Plan/apply workflow for safe changes
- Direct reference from server resources via `placement_group_id`

**Cons:**
- Requires HCL knowledge
- State management overhead for a resource with almost no configuration
- No built-in organizational conventions

**Verdict:** Production-grade for Terraform teams. The standard choice before Planton.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `hcloud.PlacementGroup`:

```go
package main

import (
    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        pg, err := hcloud.NewPlacementGroup(ctx, "ha-db-group", &hcloud.PlacementGroupArgs{
            Name: pulumi.String("ha-db-group"),
            Type: pulumi.String("spread"),
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
                "role":        pulumi.String("database"),
            },
        })
        if err != nil {
            return err
        }

        ctx.Export("placementGroupId", pg.ID())
        return nil
    })
}
```

**Pros:**
- Full programming language (Go, TypeScript, Python)
- Type safety catches errors at compile time
- Direct resource referencing for server creation

**Cons:**
- More verbose than HCL for a single resource
- Smaller community for Hetzner Cloud specifically

**Verdict:** Excellent for Go/TypeScript teams. Planton uses Pulumi (Go) internally for its IaC modules.

## Comparative Analysis

| Method | State Tracking | Drift Detection | Dependency Awareness | Audit Trail | Automation |
|--------|---------------|-----------------|---------------------|-------------|------------|
| Console | No | No | No | Minimal | No |
| CLI | No | No | No | No | Partial |
| Terraform | Yes | Yes | Yes (via `placement_group_id` reference) | Via VCS | Yes |
| Pulumi | Yes | Yes | Yes (via resource reference) | Via VCS | Yes |
| **Planton** | **Yes** | **Yes** | **Yes (via StringValueOrRef)** | **Via VCS** | **Yes** |

The differentiator for Planton is not the state management — Terraform and Pulumi handle that. It is the **standardized manifest format** and **output referencing** that makes placement groups composable with servers and infra charts without writing custom glue code.

## The Planton Approach

### Manifest Format

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-db-group
spec:
  type: spread
```

Or, since `spread` is the default:

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-db-group
spec: {}
```

### What Planton Automates

1. **Naming:** The placement group name in Hetzner Cloud is derived from `metadata.name`
2. **Labeling:** Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are computed from metadata and merged with user-specified labels
3. **Provider configuration:** Hetzner Cloud API token is resolved from provider config or environment variables
4. **Dual IaC:** The same manifest drives both Pulumi and Terraform backends
5. **Output referencing:** The `placement_group_id` output feeds into `HetznerCloudServer.spec.placementGroupId` via `StringValueOrRef`

### The 80/20 Principle

The Hetzner Cloud placement group API has 3 user-controllable attributes: `name`, `type`, and `labels`. Planton's `HetznerCloudPlacementGroupSpec` exposes 1 field: `type`.

**Included:**
- `type` — The placement strategy. Defaults to `spread`. Currently the only supported value, but exposed as an enum to accommodate future strategies without a schema change.

**Handled by the platform:**
- `name` — Derived from `metadata.name`. Consistent naming across all components.
- `labels` — Computed from metadata (org, env, kind, id) with user labels merged in.

This is the minimum viable configuration surface. A manifest with an empty `spec` creates a production-ready placement group because the only meaningful attribute (`type`) has a sensible default.

### API Design Decisions

**Optional type with proto default:** The `type` field uses `optional Type type = 1` with a `(dev.planton.shared.options.default) = "spread"` annotation. If the user omits the field, middleware applies the default. If Hetzner Cloud adds new strategies in the future (e.g., `cluster`, `partition`), users can opt in without a schema migration.

**Enum instead of string:** The `Type` enum (`type_unspecified = 0`, `spread = 1`) provides compile-time validation. Invalid type values are rejected at the proto level rather than at the cloud API level, catching errors earlier in the feedback loop.

**Single output:** The only output is `placement_group_id` — the Hetzner Cloud numeric ID. The `servers` list (available in the Terraform data source) is intentionally excluded because it represents runtime state, not provisioned infrastructure. Server membership is managed by the server component, not the placement group.

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Description |
|------------|----------|-------|-------------|
| Pulumi | `hcloud.PlacementGroup` | 1 | Placement group with spread strategy |
| Terraform | `hcloud_placement_group` | 1 | Placement group with spread strategy |

This is a single-resource component with no sub-resources, attachments, or optional companions.

### Dependency Role

`HetznerCloudPlacementGroup` is a **root resource** — it has no foreign key dependencies. It is referenced by:

- `HetznerCloudServer.spec.placementGroupId` — Servers are assigned to the group at creation time

In infra charts, the pattern is:

```
HetznerCloudPlacementGroup (foundation)
  └── placement_group_id output
        └── HetznerCloudServer.spec.placementGroupId (via StringValueOrRef)
```

### Label Management

Both IaC modules apply a standard label set to the Hetzner Cloud placement group:

| Label Key | Source | Example |
|-----------|--------|---------|
| `planton-ai_resource` | Constant | `"true"` |
| `planton-ai_name` | `metadata.name` | `"ha-db-group"` |
| `planton-ai_kind` | Constant | `"HetznerCloudPlacementGroup"` |
| `planton-ai_organization` | `metadata.org` | `"my-org"` |
| `planton-ai_environment` | `metadata.env` | `"production"` |
| `planton-ai_id` | `metadata.id` | `"hcpg-abc123"` |

User-specified `metadata.labels` are merged in, with standard labels taking precedence on key conflicts.

## Production Best Practices

### Server Limit

Hetzner Cloud enforces a **maximum of 10 servers per placement group**. This is a hard limit imposed by the physical infrastructure — spreading more than 10 instances across distinct hosts in a single datacenter location is not guaranteed.

**Practical implication:** For workloads requiring more than 10 instances with anti-affinity, create multiple placement groups segmented by role:

```yaml
# Group for database replicas (max 3-5 servers)
metadata:
  name: ha-db-group

# Group for application servers (max 10 servers)
metadata:
  name: ha-app-group
```

### When to Use Placement Groups

**Use placement groups when:**
- Running database replicas (PostgreSQL, MySQL, Redis Sentinel) that must survive single-host failure
- Deploying stateful workloads where simultaneous failure of multiple instances causes data loss
- Building HA clusters (etcd, Consul, ZooKeeper) where quorum depends on independent failure domains

**Do NOT use placement groups when:**
- Running stateless application servers behind a load balancer — the load balancer handles failover regardless of physical host placement
- Deploying a single server — anti-affinity requires at least 2 instances to be meaningful
- Latency between instances matters more than fault isolation — spread placement increases inter-host network latency compared to co-located servers

### Lifecycle Considerations

- **Placement groups are assigned at server creation time.** Moving a running server to a different placement group requires server recreation.
- **Deleting a placement group** does not affect running servers — they continue to run but lose the anti-affinity guarantee for future operations.
- **Renaming or relabeling** a placement group is an in-place update with no impact on member servers.

### Naming Conventions

Use names that indicate the failure domain role, not the specific service:

```
ha-db-group        (good — describes the HA role)
postgres-replicas  (ok — specific but clear)
pg-001             (bad — opaque, doesn't indicate purpose)
```

In Planton, the name comes from `metadata.name` and is consistent across all components.

## References

- [Hetzner Cloud Placement Groups Documentation](https://docs.hetzner.cloud/#placement-groups)
- [Hetzner Cloud API — Placement Groups](https://docs.hetzner.cloud/#placement-groups-get-all-placementgroups)
- [Terraform hcloud_placement_group Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/placement_group)
- [Pulumi hcloud.PlacementGroup Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/placementgroup/)
