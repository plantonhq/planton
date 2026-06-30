# GCP Compute Instance: Technical Research and Implementation Guide

## Introduction

Google Compute Engine is Google Cloud's Infrastructure-as-a-Service (IaaS) offering that provides virtual machines running on Google's global infrastructure. Compute Engine VMs are the foundational building blocks for running workloads on GCP, offering a wide range of machine types, storage options, and networking configurations.

This document provides comprehensive research on GCP Compute Engine instances, exploring the deployment landscape, best practices, and the rationale behind Planton's implementation choices.

## The Evolution of Compute Engine

### Historical Context

Google Compute Engine was launched in 2012 as Google's answer to Amazon EC2. Since then, it has evolved significantly:

- **2012**: Initial launch with basic VM functionality
- **2014**: Live migration and custom machine types
- **2016**: Preemptible VMs for cost optimization
- **2018**: Sole-tenant nodes for compliance
- **2020**: Confidential VMs with memory encryption
- **2022**: Spot VMs replacing preemptible VMs
- **2023**: C3 and C3D machine types with latest Intel/AMD processors

### Current Capabilities

Today, Compute Engine offers:

- 100+ predefined machine types across multiple families
- Custom machine types with user-defined CPU/memory ratios
- GPU and TPU accelerators for ML/AI workloads
- Local SSDs for high-performance temporary storage
- Persistent disks with regional and zonal options
- Advanced networking with VPC, load balancing, and CDN integration

## Deployment Methods

### 1. Google Cloud Console

The web-based console provides a guided experience:

**Advantages:**
- Visual interface with real-time validation
- Helpful defaults and recommendations
- Easy exploration of available options

**Disadvantages:**
- Not reproducible or version-controlled
- Manual process prone to human error
- Not suitable for automation

### 2. gcloud CLI

The command-line interface for direct API access:

```bash
gcloud compute instances create my-vm \
  --project=my-project \
  --zone=us-central1-a \
  --machine-type=e2-medium \
  --image-family=debian-11 \
  --image-project=debian-cloud \
  --boot-disk-size=20GB \
  --boot-disk-type=pd-ssd
```

**Advantages:**
- Scriptable and automatable
- Can be version-controlled
- Suitable for CI/CD pipelines

**Disadvantages:**
- Imperative rather than declarative
- State management is manual
- Complex for multi-resource deployments

### 3. Terraform/OpenTofu

Infrastructure as Code with declarative configuration:

```hcl
resource "google_compute_instance" "vm" {
  name         = "my-vm"
  machine_type = "e2-medium"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }

  network_interface {
    network = "default"
    access_config {
      // Ephemeral public IP
    }
  }
}
```

**Advantages:**
- Declarative and idempotent
- State tracking and drift detection
- Plan/apply workflow for safety
- Large ecosystem and community

**Disadvantages:**
- Requires Terraform knowledge
- State file management complexity
- Learning curve for HCL syntax

### 4. Pulumi

Infrastructure as Code using general-purpose languages:

```go
instance, err := compute.NewInstance(ctx, "my-vm", &compute.InstanceArgs{
    MachineType: pulumi.String("e2-medium"),
    Zone:        pulumi.String("us-central1-a"),
    BootDisk: &compute.InstanceBootDiskArgs{
        InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
            Image: pulumi.String("debian-cloud/debian-11"),
        },
    },
    NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
        &compute.InstanceNetworkInterfaceArgs{
            Network: pulumi.String("default"),
        },
    },
})
```

**Advantages:**
- Use familiar programming languages
- Full IDE support and type checking
- Better abstraction capabilities
- Easier testing with standard tools

**Disadvantages:**
- Requires programming knowledge
- Smaller community than Terraform
- More complex setup

### 5. Cloud Deployment Manager

Google's native IaC tool:

```yaml
resources:
- name: my-vm
  type: compute.v1.instance
  properties:
    zone: us-central1-a
    machineType: zones/us-central1-a/machineTypes/e2-medium
    disks:
    - boot: true
      autoDelete: true
      initializeParams:
        sourceImage: projects/debian-cloud/global/images/family/debian-11
    networkInterfaces:
    - network: global/networks/default
```

**Advantages:**
- Native GCP integration
- No external tools required
- Integrated with Cloud Console

**Disadvantages:**
- GCP-specific, not multi-cloud
- Limited community and examples
- Less flexible than Terraform/Pulumi

### 6. Config Connector (Kubernetes)

Manage GCP resources through Kubernetes:

```yaml
apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeInstance
metadata:
  name: my-vm
spec:
  zone: us-central1-a
  machineType: e2-medium
  bootDisk:
    initializeParams:
      sourceImage: projects/debian-cloud/global/images/family/debian-11
  networkInterface:
  - network: default
```

**Advantages:**
- Kubernetes-native experience
- GitOps-friendly
- Unified control plane for GCP and K8s

**Disadvantages:**
- Requires Kubernetes cluster
- Additional complexity
- Slower than native API

## Comparative Analysis

| Method | Reproducibility | Automation | Learning Curve | Multi-Cloud |
|--------|----------------|------------|----------------|-------------|
| Console | Low | None | Low | No |
| gcloud | Medium | Good | Medium | No |
| Terraform | High | Excellent | Medium | Yes |
| Pulumi | High | Excellent | Medium-High | Yes |
| Deployment Manager | High | Good | Medium | No |
| Config Connector | High | Excellent | High | No |

## Planton's Approach

### Why We Created This Component

Planton provides a Kubernetes Resource Model (KRM) interface for GCP Compute Engine instances, offering:

1. **Declarative YAML Configuration**: Familiar syntax for Kubernetes users
2. **Dual IaC Implementation**: Both Pulumi (Go) and Terraform modules
3. **Cross-Resource References**: Link to GcpProject, GcpVpc, GcpSubnetwork resources
4. **Validation at Definition Time**: Proto-based schema with buf.validate rules
5. **Consistent Patterns**: Same structure across all deployment components

### 80/20 Feature Selection

We focused on the 20% of features that cover 80% of use cases:

**In Scope:**
- Machine type selection
- Boot disk configuration (image, size, type)
- Network interface configuration (VPC, subnet, external IP)
- Service account attachment
- Spot/Preemptible VMs
- Labels, tags, and metadata
- Startup scripts
- Attached data disks
- Scheduling options

**Out of Scope (for now):**
- GPU/TPU accelerators
- Local SSDs
- Shielded VM options
- Confidential VM
- Sole-tenant nodes
- Instance templates
- Managed instance groups
- Reservation affinity

### Design Decisions

#### Machine Type Flexibility
We accept any valid machine type string rather than enumerating all options. This provides:
- Forward compatibility with new machine types
- Support for custom machine types
- Simpler schema without bloated enums

#### Zone vs Region
We use zone-level deployment (not regional) because:
- Compute instances are inherently zonal resources
- Regional deployments require managed instance groups
- Simpler configuration and mental model

#### Network Interface Structure
We support multiple network interfaces because:
- Multi-NIC VMs are common for network appliances
- Each interface can have different configurations
- Matches the underlying GCP API structure

#### Boot Disk Simplification
We use `image` instead of `source_image` + `source_image_project` because:
- The combined format (`project/image-family`) is more intuitive
- GCP accepts both short and full paths
- Reduces configuration complexity

## Implementation Landscape

### Pulumi Module Architecture

```
module/
├── main.go          # Entry point, provider setup
├── locals.go        # Data transformations, labels
├── outputs.go       # Export constants
└── instance.go      # Instance resource creation
```

Key implementation details:
- Uses `compute.Instance` from Pulumi GCP provider
- Resolves StringValueOrRef fields for project, network, subnet
- Applies standard Planton labels
- Handles optional fields with nil checks

### Terraform Module Architecture

```
tf/
├── provider.tf      # Google provider configuration
├── variables.tf     # Input variables (mirrors spec.proto)
├── locals.tf        # Computed values and transformations
├── main.tf          # Instance resource definition
└── outputs.tf       # Output values
```

Key implementation details:
- Uses `google_compute_instance` resource
- Dynamic blocks for network interfaces and disks
- Conditional logic for optional configurations
- Outputs match stack_outputs.proto

## Production Best Practices

### Machine Type Selection

1. **Start Small**: Begin with E2 series for cost efficiency
2. **Right-Size**: Monitor actual usage and adjust
3. **Consider Committed Use**: 1-3 year commitments for 37-70% savings
4. **Use Spot for Fault-Tolerant**: Batch jobs, CI/CD, stateless apps

### Networking

1. **Use Custom VPCs**: Avoid the default network in production
2. **Private IPs**: Prefer private connectivity when possible
3. **Network Tags**: Use for firewall rule targeting
4. **Alias IP Ranges**: For multi-tenant container workloads

### Security

1. **Least-Privilege Service Accounts**: Never use default compute SA
2. **OS Login**: Use for SSH access management via IAM
3. **Shielded VMs**: Enable for production workloads
4. **No External IPs**: Use Cloud NAT or IAP for egress/ingress

### Reliability

1. **Live Migration**: Keep enabled for maintenance resilience
2. **Startup Scripts**: Make idempotent, add health checks
3. **Metadata**: Use for dynamic configuration
4. **Labels**: Consistent labeling for operations

### Cost Optimization

1. **Spot VMs**: Use for fault-tolerant workloads
2. **Preemptible Batch Jobs**: Schedule during off-peak hours
3. **Auto-Shutdown**: Delete dev instances after hours
4. **Right-Size Disks**: Start small, grow as needed

## Common Pitfalls

### 1. Using Default Service Account
The default compute service account has broad permissions. Always create dedicated service accounts with minimal scopes.

### 2. Public IPs Without Firewall Rules
External IPs are open to the internet. Always configure firewall rules or remove external access.

### 3. Ignoring Startup Script Failures
Startup scripts can fail silently. Implement logging and health checks to detect issues.

### 4. Not Using Labels
Labels are essential for cost allocation, automation, and operations. Establish a labeling strategy.

### 5. Hardcoding Zone Selection
Zones can have capacity issues. Consider zone-agnostic deployments or fallback zones.

## Conclusion

GCP Compute Engine instances are versatile building blocks for cloud infrastructure. Planton's GcpComputeInstance component provides a standardized, validated interface for deploying VMs with both Pulumi and Terraform, focusing on the most common configuration patterns while maintaining flexibility for advanced use cases.

The implementation balances simplicity with capability, enabling teams to quickly deploy production-grade VM instances while following GCP best practices.
