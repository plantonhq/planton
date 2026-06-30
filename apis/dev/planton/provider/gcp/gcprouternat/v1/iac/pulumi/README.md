# GCP Cloud Router NAT - Pulumi Module

This directory contains the Pulumi implementation for deploying Google Cloud Router with Cloud NAT based on the `GcpRouterNat` API resource specification.

## Overview

The Pulumi module provisions:
- **Google Cloud Router**: Regional router resource attached to the specified VPC network.
- **Google Compute Address (optional)**: Static external IP addresses if manual NAT IP allocation is specified.
- **Google Router NAT**: NAT gateway configuration with automatic or manual IP allocation, subnet coverage, and logging.

## Module Structure

```
iac/pulumi/
├── main.go                 # Pulumi program entrypoint
├── Pulumi.yaml             # Pulumi project configuration
├── Makefile                # Build automation
├── debug.sh                # Local debugging helper
├── README.md               # This file
└── module/
    ├── main.go             # Module entrypoint (calls resources)
    ├── locals.go           # Local value transformations (labels, config)
    ├── router_nat.go       # Cloud Router and NAT resource provisioning
    └── outputs.go          # Output definitions (matching stack_outputs.proto)
```

## Inputs

The module accepts `GcpRouterNatStackInput` as input, which contains:

### `target` (GcpRouterNat)
- **`metadata`**: Resource metadata (name, org, env, labels)
- **`spec`**: Resource specification
  - **`vpc_self_link`**: VPC network self-link or reference
  - **`region`**: GCP region for Cloud Router and NAT
  - **`subnetwork_self_links`** (optional): Specific subnets to cover (default: all subnets)
  - **`nat_ip_names`** (optional): Static IP names for manual allocation (default: auto-allocate)
  - **`log_filter`** (optional): Logging level (`DISABLED`, `ERRORS_ONLY`, `ALL`; default: `ERRORS_ONLY`)

### `provider_config` (GcpProviderConfig)
- **`project_id`**: GCP project ID
- **`credentials`**: GCP service account credentials

## Outputs

The module exports outputs matching `GcpRouterNatStackOutputs`:

- **`name`**: Name of the Cloud NAT gateway
- **`router_self_link`**: Self-link URL of the Cloud Router
- **`nat_ip_addresses`**: List of external IP addresses used by NAT (auto-allocated or static)

## Resource Creation Logic

### 1. Cloud Router

Creates a regional Cloud Router in the specified VPC network:

```go
google_compute_router.router {
  name    = metadata.name
  region  = spec.region
  network = spec.vpc_self_link
}
```

### 2. Static External IPs (Conditional)

If `spec.nat_ip_names` is provided, creates static external IP addresses:

```go
google_compute_address.nat_ips[*] {
  name         = spec.nat_ip_names[i]
  region       = spec.region
  address_type = "EXTERNAL"
  labels       = gcp_labels
}
```

### 3. Cloud NAT

Creates the NAT gateway with appropriate configuration:

```go
google_compute_router_nat.nat {
  name   = metadata.name
  router = google_compute_router.router.name
  region = spec.region
  
  # IP allocation strategy
  nat_ip_allocate_option = len(nat_ip_names) > 0 ? "MANUAL_ONLY" : "AUTO_ONLY"
  nat_ips                = len(nat_ip_names) > 0 ? nat_ip_addresses : []
  
  # Subnet coverage
  source_subnetwork_ip_ranges_to_nat = len(subnetwork_self_links) > 0 ? "LIST_OF_SUBNETWORKS" : "ALL_SUBNETWORKS_ALL_IP_RANGES"
  subnetworks = [...] # If LIST_OF_SUBNETWORKS
  
  # Logging configuration
  log_config {
    enable = log_filter != "DISABLED"
    filter = log_filter  # ERRORS_ONLY or ALL
  }
}
```

## Local Values

The `locals.go` file computes derived values:

- **GCP Labels**: Standardized labels for resource tagging
  - `resource: true`
  - `resource-name: <name>`
  - `resource-kind: gcprouternat`
  - `organization: <org>` (if provided)
  - `environment: <env>` (if provided)
  - `resource-id: <id>` (if provided)

## Usage

### Local Development

```bash
# Navigate to Pulumi directory
cd apis/dev/planton/provider/gcp/gcprouternat/v1/iac/pulumi

# Set stack input (example)
export STACK_INPUT=$(cat <<EOF
{
  "target": {
    "metadata": {
      "name": "dev-nat-uscentral1"
    },
    "spec": {
      "vpc_self_link": {
        "value": "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-vpc"
      },
      "region": "us-central1"
    }
  },
  "provider_config": {
    "project_id": "my-project",
    "credentials": "..."
  }
}
EOF
)

# Preview changes
pulumi preview

# Deploy
pulumi up

# Destroy
pulumi destroy
```

### Debugging

Use the `debug.sh` helper script:

```bash
./debug.sh
```

This script loads stack input from `hack/manifest.yaml` and runs `pulumi preview`.

## Integration with Planton

This Pulumi module is invoked by the Planton CLI when processing `GcpRouterNat` resources:

```bash
# Deploy NAT gateway
planton apply -f nat-config.yaml

# Status
planton status gcprouternat dev-nat-uscentral1

# Destroy
planton destroy gcprouternat dev-nat-uscentral1
```

## Implementation Details

### NAT IP Allocation Strategy

**Auto-Allocation (Default):**
- `nat_ip_names` is empty
- `nat_ip_allocate_option = "AUTO_ONLY"`
- Google automatically assigns and scales IPs
- No static IP resources created

**Manual Allocation:**
- `nat_ip_names` contains IP names
- `nat_ip_allocate_option = "MANUAL_ONLY"`
- Static IP resources created for each name
- NAT uses specified static IPs

### Subnet Coverage Strategy

**All Subnets (Default):**
- `subnetwork_self_links` is empty
- `source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"`
- All subnets in the region automatically covered

**Specific Subnets:**
- `subnetwork_self_links` contains subnet self-links
- `source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"`
- Only specified subnets covered
- Each subnet configured with `source_ip_ranges_to_nat = ["ALL_IP_RANGES"]`

### Logging Configuration

**ERRORS_ONLY (Default):**
```go
log_config {
  enable = true
  filter = "ERRORS_ONLY"
}
```

**ALL (Full Logging):**
```go
log_config {
  enable = true
  filter = "ALL"
}
```

**DISABLED:**
```go
log_config {
  enable = false
}
```

## Dependencies

The module uses these Pulumi SDKs:

- `github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp` - GCP provider
- `github.com/pulumi/pulumi/sdk/v3/go/pulumi` - Pulumi core

## Error Handling

The module wraps errors with context:

- `"failed to create router"` - Cloud Router creation failed
- `"failed to create static nat ip"` - Static IP creation failed
- `"failed to create router nat"` - Cloud NAT creation failed

## Outputs Export

Outputs are exported to match `GcpRouterNatStackOutputs`:

```go
ctx.Export("name", createdRouterNat.Name)
ctx.Export("router_self_link", createdRouter.SelfLink)
ctx.Export("nat_ip_addresses", createdRouterNat.NatIps)
```

These outputs are captured by Planton and stored in `status.outputs`.

## Best Practices

1. **Use Auto-Allocation**: Default behavior is recommended unless external partners require allowlisting.
2. **Cover All Subnets**: Future-proof—new subnets automatically get NAT.
3. **Enable Logging**: `ERRORS_ONLY` is production-ready default.
4. **Regional Resources**: Cloud Router and NAT are regional—deploy one per region.

## Troubleshooting

### Error: "Network not found"
**Cause:** VPC self-link is incorrect or VPC doesn't exist.
**Solution:** Verify `spec.vpc_self_link` points to an existing VPC.

### Error: "IP address already exists"
**Cause:** Static IP name already exists in the region.
**Solution:** Use unique IP names or delete existing IPs.

### Error: "Region mismatch"
**Cause:** Static IPs created in different region than NAT.
**Solution:** Ensure all resources use the same region.

## Related Documentation

- **API Specification**: [spec.proto](../../spec.proto)
- **Examples**: [examples.md](../../examples.md)
- **Overview**: [README.md](../../README.md)
- **Architecture**: [overview.md](overview.md)
- **Terraform Module**: [../tf/README.md](../tf/README.md)

## Testing

### Unit Testing

Currently, unit tests focus on protobuf validation:

```bash
cd ../../
go test -v
```

### Integration Testing

Use the Planton CLI for end-to-end testing:

```bash
# Deploy test NAT
planton apply -f hack/manifest.yaml

# Verify outputs
planton status gcprouternat test-nat

# Cleanup
planton destroy gcprouternat test-nat
```

## Maintenance

### Updating Dependencies

```bash
# Update Pulumi GCP SDK
go get -u github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp

# Update Pulumi SDK
go get -u github.com/pulumi/pulumi/sdk/v3/go/pulumi

# Sync dependencies
go mod tidy
```

### Regenerating After Proto Changes

After modifying `spec.proto`:

```bash
# Regenerate Go stubs
cd ../../../../../..
make protos

# Update Pulumi module if needed
cd apis/dev/planton/provider/gcp/gcprouternat/v1/iac/pulumi/module
# Update code to use new fields
```

---

For architectural details, see [overview.md](overview.md).

