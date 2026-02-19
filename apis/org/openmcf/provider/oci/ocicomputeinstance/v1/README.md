# Overview

The **OCI Compute Instance API Resource** provides a consistent and standardized interface for deploying and managing virtual machines and bare metal hosts on Oracle Cloud Infrastructure. A compute instance is OCI's fundamental workload primitive — every application server, CI runner, database node, and container host ultimately runs on a compute instance. This component wraps the full `oci_core_instance` API surface with the standard OpenMCF KRM pattern.

## Purpose

This API resource streamlines the deployment of OCI compute instances by offering a unified interface that covers the full range of instance configurations — from a minimal 1-OCPU flex VM to a security-hardened bare metal host with NUMA tuning. It enables users to:

- **Right-Size with Flex Shapes**: Allocate exact OCPUs and memory via `shapeConfig` instead of choosing from fixed instance types. Flex shapes (E4, A1, E5) are OCI's modern default and eliminate the "closest fit" problem common in other clouds.
- **Configure Networking Inline**: The primary VNIC — subnet placement, NSG associations, public IP assignment, and DNS hostname — is declared as part of the instance spec, matching OCI's API where the primary VNIC is inseparable from the instance lifecycle.
- **Bootstrap with Cloud-Init**: Pass SSH keys and cloud-init scripts through the `metadata` map using the standard OCI keys (`ssh_authorized_keys`, `user_data`) without needing separate provisioner resources.
- **Reduce Cost with Preemptible Instances**: Configure spot-like pricing via `preemptibleInstanceConfig` for fault-tolerant workloads, with control over whether the boot volume is preserved on preemption.
- **Harden Security at the Platform Level**: Enable Secure Boot, Measured Boot, TPM, and memory encryption through `platformConfig` without navigating the complex matrix of platform types and shape families.
- **Compose with Other OCI Resources**: Reference OciCompartment, OciSubnet, OciNetworkSecurityGroup, and OciKmsKey outputs via `StringValueOrRef` for declarative, cross-resource dependency chains.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Full Flex Shape Support**: First-class `shapeConfig` for specifying OCPUs, memory, baseline utilization (burstable), and NVMe drives. Fixed shapes work without `shapeConfig`.
- **Primary VNIC Configuration**: Subnet, NSG association (up to 5), public IP control, hostname label, private IP pinning, and source/destination check — all declared inline.
- **Cloud-Init and SSH Keys**: The `metadata` map directly supports the OCI-standard `ssh_authorized_keys` and `user_data` keys. No intermediate resources or external provisioners needed.
- **Preemptible Instances**: Built-in support for preemptible (spot-like) pricing with configurable boot volume preservation on preemption.
- **Platform Security**: Secure Boot, Measured Boot, TPM, and memory encryption for VM shapes. Additional bare metal controls: NUMA nodes per socket, SMT, nested virtualization, IOMMU, core percentage.
- **Infrastructure Placement**: Availability domain, fault domain, capacity reservation, and dedicated VM host fields provide full control over physical placement for HA and compliance.
- **Oracle Cloud Agent Control**: Enable, disable, or configure individual Cloud Agent plugins (monitoring, management, vulnerability scanning) per instance.
- **Automatic Tagging**: Standard OpenMCF freeform tags applied to every instance (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Infra-Chart Composability**: Exports 5 stack outputs (`instanceId`, `privateIp`, `publicIp`, `bootVolumeId`, `availabilityDomain`) for downstream `StringValueOrRef` references.

## How OCI Compute Differs from Other Providers

Understanding these differences is essential when coming from AWS, GCP, or Azure:

- **Flex Shapes vs Fixed Instance Types**: AWS, GCP, and Azure offer fixed instance types (e.g., `m5.xlarge`, `n2-standard-4`, `Standard_D4s_v3`) where CPU and memory are predetermined. OCI's flex shapes let you specify exact OCPUs and memory within a ratio range, eliminating the need to choose between dozens of instance types. The shape (e.g., `VM.Standard.E4.Flex`) selects the processor family; `shapeConfig` selects the resource allocation.
- **OCPUs vs vCPUs**: An OCI OCPU is a physical core with SMT, equivalent to 2 vCPUs in AWS/GCP/Azure. A 1-OCPU instance has comparable compute power to a 2-vCPU instance on other providers.
- **Availability Domains vs Availability Zones**: OCI availability domains are equivalent to AWS availability zones — isolated data centers within a region. The key difference: OCI's `availabilityDomain` is a required field on every instance (not inherited from the subnet), and the format includes a tenancy-specific prefix (e.g., `Ixxj:US-ASHBURN-AD-1`).
- **Fault Domains**: Within each availability domain, OCI provides three fault domains — logical groupings of hardware with independent power and networking. This is an additional HA layer that AWS and GCP don't expose directly (AWS placement groups serve a similar but different purpose).
- **Compartment Scoping**: Every OCI instance lives in a compartment — OCI's hierarchical resource isolation model. This is the first field in the spec because it determines IAM policy scope, cost tracking, and resource visibility. AWS uses accounts, GCP uses projects, Azure uses resource groups.
- **VNIC Model**: OCI's primary VNIC is created with the instance and cannot be detached. Secondary VNICs can be added and removed independently. This differs from AWS ENIs, which are all independently attachable/detachable including the primary.
- **Oracle Cloud Agent**: OCI instances include an agent for monitoring, OS management, and vulnerability scanning — built into the platform rather than requiring a separate agent installation like AWS SSM or GCP OS Config.
- **Preemptible vs Spot**: OCI preemptible instances are similar to AWS Spot Instances and GCP Preemptible VMs. The pricing model and reclamation behavior differ, but the concept — cheaper instances that can be terminated when capacity is needed — is the same.

## Critical Constraints

- **Flex Shapes Require shapeConfig**: If the shape name contains "Flex" (e.g., `VM.Standard.E4.Flex`, `VM.Standard.A1.Flex`), you must provide `shapeConfig` with at least `ocpus`. Without it, OCI may assign minimal defaults or reject the request.
- **Availability Domain Format**: The availability domain string includes a tenancy-specific prefix (e.g., `Ixxj:US-ASHBURN-AD-1`). This prefix varies by tenancy. Use `oci iam availability-domain list` to get the correct values.
- **Boot Volume Size Minimum**: When launching from an image, `bootVolumeSizeInGbs` must be >= the image's minimum boot volume size. When omitted, defaults to the image minimum.
- **VNIC NSG Limit**: A single VNIC supports a maximum of 5 network security groups (`nsgIds` max 5 items). This is an OCI platform limit.
- **assignPublicIp Tri-State**: The `assignPublicIp` field has three states: unset (use subnet default), `true` (assign), `false` (don't assign). The default depends on the subnet type — public subnets assign by default, private subnets do not.
- **Availability Domain Is Immutable**: Changing `availabilityDomain` forces instance recreation. Instances cannot be live-migrated across availability domains.
- **Preemptible Limitations**: Preemptible instances cannot be stopped and started — only terminated. If reclaimed, the instance is terminated and optionally the boot volume is preserved based on `preserveBootVolume`.
- **PlatformConfig Type Must Match Shape**: The `platformConfig.type` field must correspond to the instance's shape family. Using `amd_vm` with an Intel shape will fail. OCI validates this server-side.

## Use Cases

- **Web and Application Servers**: Flex shapes sized to the workload with public or private subnets, NSG-based security, and cloud-init for automated bootstrapping.
- **Development and Testing**: Small flex VMs (1 OCPU, 8-16 GB) in development compartments. Burstable shapes (`baselineOcpuUtilization: BASELINE_1_8`) reduce cost for intermittent workloads.
- **CI/CD Runners**: Preemptible instances for cost-sensitive build workers. Boot volume preservation allows quick re-launch after preemption without re-downloading build caches.
- **Batch Processing**: High-OCPU flex shapes for compute-intensive jobs. Preemptible pricing for workloads that can tolerate interruption.
- **Security-Hardened VMs**: Platform security (Secure Boot, Measured Boot, TPM, memory encryption), IMDSv2-only access, in-transit encryption, and private subnets for compliance-sensitive workloads.
- **Dedicated Hosts for Licensing**: Place instances on dedicated VM hosts via `dedicatedVmHostId` for BYOL (Bring Your Own License) scenarios where license terms require physical isolation.
- **Bare Metal for Performance**: Use bare metal shapes with NUMA tuning, SMT control, core percentage, and nested virtualization for workloads requiring direct hardware access — databases, HPC, or hypervisors.

## Production Features

This resource provides complete support for production-grade compute instance deployments, including:

- **Platform Security**: Secure Boot, Measured Boot, TPM, and memory encryption protect the instance from boot-level and runtime attacks.
- **In-Transit Encryption**: Paravirtualized boot and data volume attachments encrypted in transit via `isPvEncryptionInTransitEnabled`.
- **Capacity Reservations**: Reserve compute capacity in advance via `capacityReservationId` to guarantee availability during demand spikes.
- **Dedicated VM Hosts**: Physical isolation via `dedicatedVmHostId` for compliance requirements or software licensing constraints.
- **Oracle Cloud Agent**: Fine-grained control over monitoring, management, and vulnerability scanning plugins per instance.
- **Availability and Recovery**: Configure live migration preference and recovery action (restore vs stop) for infrastructure maintenance events.
- **Freeform Tagging**: Standard OpenMCF labels applied as OCI freeform tags for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Infra-Chart Composability**: Designed to compose with OciCompartment, OciSubnet, OciNetworkSecurityGroup, and future OciBlockVolume and OciKmsKey components via `StringValueOrRef`.
