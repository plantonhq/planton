# Overview

The **GCP Compute Instance API Resource** provides a consistent and standardized interface for deploying and managing Google Compute Engine virtual machine instances within our infrastructure. This resource simplifies the process of creating and configuring VM instances on Google Cloud Platform (GCP), allowing users to deploy compute workloads without managing the complexities of instance configuration, networking, and disk management.

## Purpose

We developed this API resource to streamline the deployment and management of virtual machines using GCP Compute Engine. By offering a unified interface, it reduces the complexity involved in setting up and configuring production-grade VM instances, enabling users to:

- **Easily Deploy VM Instances**: Quickly create Compute Engine instances in specified GCP projects and zones.
- **Simplify Configuration**: Abstract the complexities of setting up instances, including boot disks, networking, and service accounts.
- **Integrate Seamlessly**: Utilize existing GCP credentials and integrate with VPC networks for secure connectivity.
- **Focus on Applications**: Allow developers to concentrate on running workloads rather than managing VM infrastructure.

## Key Features

- **Consistent Interface**: Aligns with our existing APIs for deploying cloud infrastructure and managed services.
- **Flexible Machine Types**: Support for all GCP machine type families (E2, N1, N2, C2, M2, etc.).
- **Simplified Deployment**: Automates the provisioning of Compute Engine instances with necessary configurations.
- **Network Security**: Support for VPC network integration, subnetworks, and external IP configuration.
- **Spot/Preemptible VMs**: Cost-effective compute options for fault-tolerant workloads.
- **Custom Boot Disks**: Configurable boot disk images, sizes, and types (SSD, balanced, standard).
- **Service Account Integration**: Attach service accounts with specific OAuth scopes for API access.
- **Startup Scripts**: Run custom initialization scripts when instances boot.

## Use Cases

- **Web Servers**: Deploy production web servers with custom configurations.
- **Development Environments**: Quickly spin up development and testing VMs.
- **Batch Processing**: Run batch jobs using Spot VMs for cost optimization.
- **CI/CD Runners**: Deploy self-hosted CI/CD runner instances.
- **Application Workloads**: Host containerized or traditional applications.
- **Data Processing**: Run data processing and analytics workloads.

## Architecture

GCP Compute Engine instances created through this API resource include:

- **Virtual Machine**: A fully managed VM instance with configurable CPU and memory.
- **Boot Disk**: Persistent disk with the selected OS image and configurable size.
- **Network Interface**: Connection to VPC networks with optional external IP.
- **Service Account**: Identity for API authentication and authorization.
- **Metadata**: Custom key-value pairs for instance configuration.
- **Labels and Tags**: For organization, billing, and firewall rules.

## Configuration Options

### Machine Types

Choose from various machine type families based on workload requirements:
- **E2**: Cost-effective for general-purpose workloads
- **N1/N2**: Balanced compute for most workloads
- **C2**: Compute-optimized for CPU-intensive workloads
- **M2**: Memory-optimized for memory-intensive workloads

### Boot Disk

- Configurable OS images (Debian, Ubuntu, CentOS, Windows, etc.)
- Size from 10GB to 65TB
- Types: `pd-standard`, `pd-ssd`, `pd-balanced`
- Auto-delete option when instance is deleted

### Networking

- **VPC Network**: Connect to custom or default VPC networks
- **Subnetwork**: Deploy in specific regional subnetworks
- **External IP**: Optional public IP with PREMIUM or STANDARD tier
- **Network Tags**: For firewall rule targeting

### Cost Optimization

- **Spot VMs**: Up to 60-91% discount, can be preempted
- **Preemptible VMs**: Legacy option, similar to Spot
- **Custom Machine Types**: Right-size CPU and memory

### Security

- **Service Accounts**: Assign specific GCP service accounts
- **OAuth Scopes**: Control API access permissions
- **SSH Keys**: Manage SSH access to instances
- **OS Login**: Integration with GCP IAM for SSH access

## Future Enhancements

As this resource continues to evolve, future updates will include:

- **GPU Attachments**: Support for attaching GPUs for ML/AI workloads.
- **Local SSDs**: High-performance local storage options.
- **Shielded VMs**: Enhanced security with Secure Boot and vTPM.
- **Sole-Tenant Nodes**: Dedicated physical servers for compliance.
- **Instance Templates**: Reusable configurations for instance groups.
- **Managed Instance Groups**: Auto-scaling and load balancing.
- **Confidential VMs**: Memory encryption for sensitive workloads.
- **Custom Images**: Support for custom OS images.
