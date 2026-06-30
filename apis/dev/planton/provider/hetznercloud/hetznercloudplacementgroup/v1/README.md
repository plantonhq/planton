# HetznerCloudPlacementGroup

The **HetznerCloudPlacementGroup** resource creates a placement group in a Hetzner Cloud account. Servers assigned to a spread placement group are guaranteed to run on different physical hosts, providing fault tolerance for high-availability workloads.

## What It Represents

A [Hetzner Cloud Placement Group](https://docs.hetzner.cloud/#placement-groups) is a named anti-affinity constraint applied to servers at creation time. Hetzner Cloud currently supports a single strategy — `spread` — which distributes servers across distinct physical hosts. If the constraint cannot be satisfied, server creation fails rather than silently co-locating.

## Bundled Resources

| Terraform Resource | Created When | Purpose |
|---|---|---|
| `hcloud_placement_group` | Always | Creates the placement group with the specified strategy |

This is a single-resource component — no optional or conditional sub-resources.

## Key Features

### Near-Zero Configuration

The only user-specified field is `type`, which defaults to `spread`. A manifest with an empty `spec` block is valid and production-ready. The placement group name is derived from `metadata.name`, and labels are computed from metadata.

### Spread Strategy

The `spread` strategy guarantees that servers in the group run on different physical hosts. Hetzner Cloud enforces a maximum of 10 servers per placement group — a hard limit imposed by the physical infrastructure.

### Enum-Based Type Field

The `type` field is defined as a proto enum (`type_unspecified`, `spread`) rather than a free-form string. This provides compile-time validation and positions the schema to accept future strategies (if Hetzner Cloud introduces them) without a breaking change.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the Hetzner Cloud placement group from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence.

## Upstream Dependencies (What This Resource Needs)

None. `HetznerCloudPlacementGroup` is a root resource with no foreign key dependencies.

## Downstream Dependents (What References This Resource)

| Dependent | Field | Purpose |
|---|---|---|
| `HetznerCloudServer` | `spec.placementGroupId` | Assign server to this placement group at creation |

## Stack Outputs

| Output | Description |
|---|---|
| `placement_group_id` | Hetzner Cloud numeric ID of the created placement group (as string) |

## References

- [Hetzner Cloud Placement Groups Documentation](https://docs.hetzner.cloud/#placement-groups)
- [Terraform hcloud_placement_group Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/placement_group)
- [Pulumi hcloud.PlacementGroup Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/placementgroup/)
