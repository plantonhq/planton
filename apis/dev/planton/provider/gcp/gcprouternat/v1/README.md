# Overview

The GCP Cloud Router NAT API resource provides a streamlined, declarative interface for deploying Google Cloud Router with Cloud NAT (Network Address Translation) within Google Cloud Platform. By abstracting the complexity of NAT gateway configuration, this resource enables you to provision secure, scalable egress connectivity for private GCP resources without managing individual public IPs or NAT appliances.

## Why We Created This API Resource

Managing internet egress for private cloud infrastructure can be challenging. Traditional approaches—assigning public IPs to every VM, maintaining NAT appliances, or dealing with port exhaustion—introduce security risks, operational overhead, and scalability concerns. To address these challenges and promote a standardized approach, we developed this API resource. It enables you to:

- **Simplify NAT Deployment**: Provision Cloud Router and NAT gateway with minimal configuration, abstracting low-level GCP networking details.
- **Enhance Security**: Consolidate egress through controlled NAT IPs instead of exposing individual VMs with public addresses.
- **Ensure Consistency**: Maintain uniform NAT configurations across environments and regions.
- **Enable Private GKE Access**: Allow private GKE clusters and other private resources to reach the internet for package downloads, API calls, and external integrations.
- **Support IP Allowlisting**: Optionally provision static NAT IPs for partner integrations requiring allowlisted egress addresses.
- **Improve Operational Efficiency**: Eliminate the need to manage NAT VMs, routing tables, or port allocation manually—Cloud NAT handles this at Google's infrastructure layer.

## Key Features

### Environment Integration

- **Environment Info**: Seamlessly integrates with Planton's environment management system to deploy NAT gateways within specific environments.
- **Stack Job Settings**: Supports custom stack-update settings for infrastructure-as-code deployments using Pulumi or Terraform.

### GCP Credential Management

- **GCP Credential ID**: Uses specified GCP credentials to ensure secure and authorized operations.
- **Project ID**: Required field specifying the GCP project ID where the Cloud Router and NAT will be created. Can be specified as a literal value or reference to a `GcpProject` resource.

### Simplified NAT Configuration

- **VPC Network Reference**: Specify the target VPC network using self-link or reference to an existing `GcpVpc` resource.
- **Regional Deployment**: Cloud Router and NAT are regional resources—specify the region where your private resources reside.
- **Automatic IP Allocation**: By default, Cloud NAT automatically allocates and scales external IPs as needed (auto-allocation mode).
- **Manual IP Allocation**: Optionally specify static external IP names for deterministic egress IPs (manual allocation mode)—useful for external partner allowlisting.
- **Flexible Subnet Coverage**:
  - **All Subnets (Default)**: NAT automatically covers all subnets in the region, including future subnets.
  - **Specific Subnets**: Optionally specify which subnets should use NAT for fine-grained control.
- **Production-Ready Logging**: Configurable logging with sensible defaults (`ERRORS_ONLY` for production) to detect port exhaustion and connection failures without excessive log volume.

### Proxy-less Architecture

Unlike traditional NAT appliances, Cloud NAT operates at Google's software-defined networking layer (Andromeda), distributing translation across infrastructure:

- **No Bottlenecks**: Traffic doesn't funnel through a single appliance—each VM's traffic is translated at the host level.
- **High Availability**: Inherently highly available within a region with no single point of failure.
- **Full Bandwidth**: Maintains full VM bandwidth and minimal latency—no NAT throughput ceiling.
- **Zero Management**: No VMs to patch, scale, or monitor—Google handles all operations.

### Validation and Compliance

- **Input Validation**: Enforces validation rules to ensure correct configuration.
  - **Required Fields**: Project ID, VPC self-link, region, router name, and NAT name are mandatory.
  - **Optional Fields**: Subnet coverage and NAT IPs are validated when provided.
  - **Naming Validation**: Router and NAT names must be 1-63 lowercase characters, starting with a letter.

## Benefits

- **Simplified Deployment**: Abstracts the complexities of Cloud Router and NAT configuration into an easy-to-use declarative API.
- **Consistency**: Ensures all NAT deployments adhere to organizational standards and best practices.
- **Scalability**: Cloud NAT scales automatically to handle traffic from thousands of VMs without manual intervention.
- **Security**: Private resources reach the internet without inbound exposure. Controlled egress IPs enable partner integration and allowlisting.
- **Cost Efficiency**: Auto-allocation mode avoids paying for unused static IPs. Private Google Access keeps Google API traffic internal (no egress costs).
- **Operational Simplicity**: No NAT VMs to manage, no routing complexity, no port allocation tuning—just declare your desired state.
- **Compliance**: Helps maintain compliance with security policies by reducing attack surface and enabling egress logging.

## Use Cases

### Private GKE Clusters

Cloud NAT is essential for private GKE clusters where nodes lack external IPs:

- **Container Image Pulls**: Nodes can pull images from Docker Hub, GCR, or other registries.
- **Package Management**: Nodes can reach package repositories for system updates.
- **External API Access**: Applications can call external APIs and services.
- **Integration**: Works seamlessly with Private Google Access for efficient Google API access.

### Cloud SQL Private IP

When Cloud SQL instances use private IP only (no public endpoint), they need Cloud NAT to:

- **Import Data**: Import data from external URLs.
- **Replicate**: Replicate from external MySQL or PostgreSQL servers.
- **External Integrations**: Reach external services for backups or monitoring.

### Multi-Region High Availability

Cloud NAT is regional, so multi-region architectures deploy one NAT gateway per region:

- **Active-Active Deployments**: Each region has its own NAT gateway for local egress.
- **Active-Passive Failover**: Standby regions have NAT configured for seamless failover.
- **Partner Allowlisting**: Provide NAT IPs from all active regions for partner firewall rules.

### Hybrid Cloud and Migration

During cloud migrations or hybrid deployments:

- **On-Premises Access**: Private GCP resources can reach on-premises services via VPN or Interconnect.
- **Legacy System Integration**: Call legacy APIs and services without exposing GCP workloads.
- **Gradual Migration**: Migrate workloads incrementally while maintaining connectivity.

## Configuration Examples

For comprehensive configuration examples, see [examples.md](examples.md).

## Architecture

Cloud Router NAT operates at Google's infrastructure layer, providing distributed, proxy-less NAT translation:

```
┌─────────────────────────────────────────────────────────────────┐
│                         GCP Region                              │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                      VPC Network                        │  │
│  │                                                         │  │
│  │  ┌──────────────┐        ┌──────────────┐            │  │
│  │  │  Subnet A    │        │  Subnet B    │            │  │
│  │  │              │        │              │            │  │
│  │  │ ┌──────┐    │        │ ┌──────┐    │            │  │
│  │  │ │ VM 1 │────┼────────┼─│ VM 2 │────┼────┐       │  │
│  │  │ └──────┘    │        │ └──────┘    │    │       │  │
│  │  │  (no pub IP)│        │  (no pub IP)│    │       │  │
│  │  └──────────────┘        └──────────────┘    │       │  │
│  │                                               │       │  │
│  │       ┌───────────────────────────────────────┘       │  │
│  │       │                                               │  │
│  │       ▼                                               │  │
│  │  ┌─────────────────────────────────────────────┐     │  │
│  │  │          Cloud Router + NAT                 │     │  │
│  │  │  • Distributed translation (no bottleneck)  │     │  │
│  │  │  • Auto or Manual IP allocation             │     │  │
│  │  │  • Logging (ERRORS_ONLY default)            │     │  │
│  │  └─────────────────────────────────────────────┘     │  │
│  │                      │                                │  │
│  └──────────────────────┼────────────────────────────────┘  │
│                         │                                   │
└─────────────────────────┼───────────────────────────────────┘
                          │
                          ▼
                    Internet
             (via NAT IP: 203.0.113.10)
```

**Key Architectural Points:**

- **Regional Scope**: Cloud Router and NAT are regional—deploy one per region where private resources need internet access.
- **Distributed Translation**: No central NAT appliance—translation happens at each VM's host for full bandwidth and HA.
- **Auto Scaling**: Cloud NAT automatically scales IP and port allocation based on VM count and traffic patterns.
- **No Configuration on VMs**: Private VMs require zero configuration—NAT is transparent to workloads.

## Best Practices

### 1. Use Auto-Allocation for Most Cases

```yaml
nat_ip_names: []  # Empty = auto-allocation (recommended default)
```

**Why:** Simplifies management, scales automatically, avoids paying for unused static IPs.

**When to use manual allocation:** External partners require allowlisting specific egress IPs.

### 2. Cover All Subnets by Default

```yaml
subnetwork_self_links: []  # Empty = all subnets (recommended)
```

**Why:** Future-proof—new subnets automatically get NAT coverage without config updates.

**When to restrict:** Security requirement to exclude certain subnets from internet access.

### 3. Enable Logging in Production

```yaml
log_filter: ERRORS_ONLY  # Default (recommended for production)
```

**Why:** Detect port exhaustion, connection failures, and NAT issues without excessive log volume.

- Use `DISABLED` for non-production to reduce costs.
- Use `ALL` only for security auditing or troubleshooting (generates significant log volume).

### 4. Combine with Private Google Access

Enable Private Google Access on your subnets so traffic to Google APIs (GCS, GCR, Cloud SQL Admin API) stays internal:

- **No Egress Costs**: Traffic to Google services doesn't traverse NAT or the internet.
- **Lower Latency**: Internal routing is faster than internet paths.
- **Simplified Allowlisting**: NAT only handles true external traffic.

### 5. Plan Static IPs for Partner Integration

If external partners need to allowlist your egress IPs:

1. Create static external IP addresses: `gcloud compute addresses create nat-ip-1 --region=us-central1`
2. Reference them in `nat_ip_names`: `["nat-ip-1", "nat-ip-2"]`
3. Provide the IP addresses to partners for firewall rules.
4. Note: Switching from auto to manual allocation briefly disrupts existing connections—plan during maintenance window.

### 6. Deploy One NAT per Region

Cloud NAT is regional—deploy a separate NAT gateway in each region where you have private resources:

```yaml
# us-central1
region: us-central1

# europe-west1 (separate NAT deployment)
region: europe-west1
```

### 7. Monitor NAT Metrics

Cloud NAT exports metrics to Cloud Monitoring:

- **Port Allocation**: Monitor port usage to detect exhaustion before it impacts applications.
- **Dropped Connections**: Alert on dropped connections due to resource limits.
- **NAT IP Count**: Track IP allocation in manual mode.

## Integration with Planton

This resource integrates seamlessly with other Planton components:

- **GcpVpc**: Reference VPC network output using `status.outputs.network_self_link`.
- **GcpGkeCluster**: Private GKE clusters automatically use NAT configured in their VPC and region.
- **GcpCloudSql**: Private Cloud SQL instances use NAT for external access if needed.
- **GcpServiceAccount**: Use workload identity for secure credential management.

## Production Readiness

This API resource is production-ready and provides:

- ✅ **Declarative Configuration**: Define desired state, let the system handle provisioning.
- ✅ **Validation**: Input validation ensures correct configuration before deployment.
- ✅ **Sensible Defaults**: Auto-allocation, all-subnets coverage, ERRORS_ONLY logging—production-ready out of the box.
- ✅ **IaC Support**: Full Pulumi and Terraform module implementations.
- ✅ **Logging**: Configurable NAT translation logging for troubleshooting and auditing.
- ✅ **High Availability**: Cloud NAT is inherently highly available within a region.
- ✅ **Zero Operational Overhead**: No VMs to manage, patch, or scale.

## Related Resources

- **GcpVpc**: Create custom VPC networks with subnets.
- **GcpGkeCluster**: Deploy private GKE clusters that use Cloud NAT.
- **GcpStaticIp**: Create static external IPs for manual NAT IP allocation.
- **GcpFirewall**: Define firewall rules for egress control.

## Additional Documentation

- **Examples**: See [examples.md](examples.md) for comprehensive configuration examples.
- **Research**: See [docs/README.md](docs/README.md) for deep-dive into Cloud NAT architecture and deployment methods.
- **Pulumi Module**: See [iac/pulumi/README.md](iac/pulumi/README.md) for Pulumi usage.
- **Terraform Module**: See [iac/tf/README.md](iac/tf/README.md) for Terraform usage.

