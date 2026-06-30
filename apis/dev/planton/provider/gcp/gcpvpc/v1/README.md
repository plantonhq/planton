# GCP VPC

## Overview

The `GcpVpc` component provides a streamlined, best-practice approach to creating Google Cloud Virtual Private Cloud (VPC) networks. It is designed with the principle of **making the right thing easy and the wrong thing hard**, steering users toward custom-mode VPCs with explicit configuration while preventing common networking pitfalls.

This component is part of Planton's infrastructure-as-code framework, offering a simple protobuf-based API that generates production-ready Terraform and Pulumi code for GCP VPC deployment.

## Key Features

- **Custom Mode by Default**: Defaults to `auto_create_subnetworks: false`, avoiding the IP overlap issues that plague auto-mode VPCs
- **Explicit Routing Configuration**: Choose between REGIONAL (default) or GLOBAL routing modes based on your hybrid connectivity needs
- **Best-Practice Defaults**: Sensible defaults that work for 80% of use cases while remaining configurable for advanced scenarios
- **Multi-Cloud Consistency**: Part of Planton's unified infrastructure API across GCP, AWS, and Azure
- **Infrastructure-as-Code Native**: Generates Terraform or Pulumi code with proper state management and drift detection
- **Foreign Key Support**: Reference other Planton resources (like `GcpProject`) directly for clean dependency management

## Use Cases

- **Development and Testing Environments**: Create isolated custom-mode VPCs with controlled IP ranges
- **Production Networks**: Deploy multi-region VPCs with global routing for hybrid connectivity
- **Shared VPC Host Networks**: Establish host project networks that multiple service projects can attach to
- **GKE-Ready Networks**: Create VPC foundations for Google Kubernetes Engine clusters with secondary IP ranges (subnets managed separately)

## API Structure

The `GcpVpc` resource follows Planton's standard Kubernetes-style API structure:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: <vpc-name>
spec:
  project_id: <gcp-project-id>
  auto_create_subnetworks: <true|false>
  routing_mode: <REGIONAL|GLOBAL>
```

## Specification Fields

### Metadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Name of the VPC network (must be unique within the project) |
| `id` | string | No | Unique identifier for tracking this resource |
| `org` | string | No | Organization identifier for labeling |
| `env` | string | No | Environment identifier (e.g., dev, staging, prod) |

### Spec

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `project_id` | StringValueOrRef | Yes | - | GCP project ID where the VPC will be created. Can reference a `GcpProject` resource |
| `auto_create_subnetworks` | bool | No | `false` | Whether to automatically create subnets in all regions (not recommended for production) |
| `routing_mode` | GcpVpcRoutingMode | No | `REGIONAL` | Dynamic routing mode for Cloud Routers. Use `GLOBAL` for multi-region routing |

### GcpVpcRoutingMode Enum

- `REGIONAL` (0): Cloud Routers advertise routes only within their region (recommended for most use cases)
- `GLOBAL` (1): Cloud Routers advertise routes across all regions (needed for multi-region hybrid connectivity)

## Quick Start

### Basic Custom-Mode VPC

The simplest production-ready VPC:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-network
spec:
  project_id:
    value: my-prod-project-123
```

This creates:
- A custom-mode VPC (no automatic subnets)
- Regional routing mode (default)
- Ready for explicit subnet creation

### Multi-Region VPC with Global Routing

For hybrid connectivity (VPN/Interconnect):

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: global-network
  org: platform-team
  env: prod
spec:
  project_id:
    value: network-host-project
  routing_mode: GLOBAL
```

### Referencing a GcpProject Resource

Using foreign key references:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: dev-vpc
spec:
  project_id:
    ref:
      kind: GcpProject
      name: my-dev-project
```

## Outputs

After successful deployment, the following outputs are available:

| Output | Type | Description |
|--------|------|-------------|
| `network_self_link` | string | Full self-link URL of the VPC network (format: `projects/<project>/global/networks/<name>`) |

These outputs can be referenced by other resources (e.g., `GcpSubnetwork`, `GcpGkeCluster`) to establish proper dependencies.

## Best Practices

### 1. Always Use Custom Mode

**❌ Don't do this** (auto-mode VPCs):
```yaml
spec:
  auto_create_subnetworks: true  # Creates subnets in all regions with fixed IP ranges
```

**✅ Do this instead**:
```yaml
spec:
  auto_create_subnetworks: false  # Explicitly create subnets as needed
```

**Why**: Auto-mode VPCs use predefined IP ranges (`10.128.0.0/9`) that cannot be changed and will conflict when peering VPCs or connecting to on-premises networks.

### 2. Choose Routing Mode Based on Architecture

- **Use REGIONAL** (default) for:
  - Single-region workloads
  - Independent per-region routing
  - Most modern cloud-native applications

- **Use GLOBAL** only when:
  - You have multi-region Cloud VPN or Interconnect
  - You need Cloud Routers in any region to advertise routes globally
  - You're building a hybrid architecture with on-premises connectivity

### 3. Plan Your Subnets Separately

VPCs and subnets are separate resources for good reason:
- **VPC**: Defines the network container and routing behavior
- **Subnets**: Define IP ranges per region, secondary ranges for GKE, Private Google Access, etc.

Manage subnets via separate `GcpSubnetwork` resources that reference this VPC's `network_self_link`.

### 4. Use Labels for Organization

```yaml
metadata:
  name: prod-network
  org: platform-team
  env: production
  id: gcpvpc-01234567890abcdef
```

These metadata fields are automatically converted to GCP labels for cost tracking and resource organization.

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (TypeScript, Python, Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Examples

For more comprehensive examples, see [`examples.md`](examples.md), including:
- Basic development VPC
- Production shared VPC setup
- Multi-region VPC with global routing
- GKE-ready VPC foundation

## Architecture

For detailed rationale behind design decisions, network provisioning evolution, and comparison of IaC tools, see the [research documentation](docs/README.md).

Key architectural principles:
- **Guard Rails, Not Handcuffs**: Sensible defaults that can be overridden when needed
- **80/20 Scoping**: Focus on the 20% of configuration that covers 80% of use cases
- **Custom Mode First**: Prevent IP conflicts before they happen
- **Explicit Over Implicit**: Routing mode and subnet creation are conscious choices

## Limitations

### What This Component Does NOT Do

- **Subnet Management**: Create subnets via separate `GcpSubnetwork` resources
- **Firewall Rules**: Manage firewall rules via `GcpFirewallRule` resources
- **VPC Peering**: Set up peering via `GcpVpcPeering` resources
- **Cloud Router/NAT**: Configure routing via separate resources

This separation provides better modularity and reusability.

### Advanced Configuration Not Exposed

The following GCP VPC features are not exposed in the base API (can be added via advanced configuration if needed):
- MTU settings (defaults to 1460 bytes)
- IPv6 and ULA internal ranges
- Custom BGP advertisement configuration
- Network firewall policy enforcement

If you need these features, you can:
1. Extend the protobuf spec with additional fields
2. Use Terraform/Pulumi directly for that specific VPC
3. Layer advanced configuration on top of the base VPC

## Troubleshooting

### "API compute.googleapis.com has not been used in project"

**Cause**: The Compute Engine API is not enabled in your project.

**Solution**: The component automatically enables the API, but if you encounter this error:
```bash
gcloud services enable compute.googleapis.com --project=<your-project-id>
```

### "The resource already exists"

**Cause**: A VPC with the same name already exists in the project.

**Solution**: 
- Choose a different `metadata.name`
- Or delete the existing VPC: `gcloud compute networks delete <vpc-name> --project=<project-id>`

### IP Range Conflicts When Peering

**Cause**: Using auto-mode VPCs or overlapping custom subnets.

**Solution**: 
- Use custom-mode VPCs (`auto_create_subnetworks: false`)
- Plan IP ranges across environments (e.g., dev: `10.10.0.0/16`, staging: `10.20.0.0/16`, prod: `10.30.0.0/16`)

## Related Resources

- [`GcpSubnetwork`](../gcpsubnetwork/v1/README.md) - Create subnets within this VPC
- [`GcpGkeCluster`](../gcpgkecluster/v1/README.md) - Deploy GKE clusters using this VPC
- [`GcpProject`](../gcpproject/v1/README.md) - Manage GCP projects

## Additional Resources

- [GCP VPC Documentation](https://cloud.google.com/vpc/docs)
- [VPC Network Overview](https://cloud.google.com/vpc/docs/vpc)
- [Best Practices for VPC Design](https://cloud.google.com/architecture/best-practices-vpc-design)
- [IP Address Planning for GKE](https://cloud.google.com/kubernetes-engine/docs/concepts/alias-ips)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

