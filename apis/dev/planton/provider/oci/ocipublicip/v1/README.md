# OCI Public IP

An Planton component that provisions an Oracle Cloud Infrastructure public IPv4 address.

## Purpose

This component manages a single OCI public IP resource through a declarative YAML manifest. It handles both reserved and ephemeral lifetime modes, optional assignment to a private IP, BYOIP pool allocation, and automatic freeform tagging based on Planton metadata.

## Key Features

- **Two lifetime modes** — `RESERVED` creates a persistent, region-scoped IP that survives instance termination and can be reassigned. `EPHEMERAL` creates an IP tied to the assigned private IP's entity lifecycle.
- **Optional private IP assignment** — reserved IPs can be created unassigned for pre-allocation (DNS records, firewall allowlists) or assigned to a specific private IP at creation time.
- **BYOIP support** — allocate the public IP from a customer-owned IP pool via `publicIpPoolId` instead of Oracle's default pool.
- **Foreign key references** — `compartmentId`, `privateIpId`, and `publicIpPoolId` all accept `StringValueOrRef`, allowing either literal OCIDs or cross-resource references via `valueFrom`.
- **Automatic tagging** — freeform tags are applied with `resource_kind`, `resource_id`, `organization`, `environment`, and any custom labels from `metadata.labels`.

## Critical Constraints

- `lifetime` cannot be changed after creation. Switching from `RESERVED` to `EPHEMERAL` (or vice versa) requires destroying and recreating the resource.
- Ephemeral IPs require `privateIpId` — the validation rule `ephemeral_requires_private_ip` enforces this at the proto level.
- For ephemeral IPs, `compartmentId` must match the compartment of the referenced private IP.
- `publicIpPoolId` cannot be changed after creation.
- The component manages exactly one public IP per manifest. Multiple IPs require separate OciPublicIp resources.

## Use Cases

| Scenario | Lifetime | privateIpId | publicIpPoolId |
|----------|----------|-------------|----------------|
| Pre-allocate a stable IP for later assignment | `RESERVED` | omitted | omitted |
| Assign a stable IP to a compute instance | `RESERVED` | set | omitted |
| Use a corporate-owned IP range | `RESERVED` | optional | set |
| Temporary IP for a dev/test instance | `EPHEMERAL` | required | omitted |

## Production Features

- **Freeform tags** propagate organization, environment, resource kind, and resource ID to the OCI resource for cost tracking and governance.
- **Display name fallback** — `displayName` defaults to `metadata.name` when not explicitly set, keeping the OCI Console readable without extra configuration.
- **Stack outputs** — the public IP OCID and allocated IPv4 address are exported for consumption by DNS records, load balancer configurations, and downstream Planton resources.

## Files

| Path | Description |
|------|-------------|
| `api.proto` | Top-level OciPublicIp message with apiVersion, kind, metadata, spec, status |
| `spec.proto` | OciPublicIpSpec — compartmentId, lifetime, displayName, privateIpId, publicIpPoolId |
| `stack_outputs.proto` | OciPublicIpStackOutputs — publicIpId, ipAddress |
| `iac/pulumi/module/main.go` | Pulumi module entry point |
| `iac/pulumi/module/public_ip.go` | Creates the `core.PublicIp` resource and exports outputs |
| `iac/pulumi/module/locals.go` | Initializes display name fallback and freeform tags |
| `iac/pulumi/module/outputs.go` | Output key constants |
