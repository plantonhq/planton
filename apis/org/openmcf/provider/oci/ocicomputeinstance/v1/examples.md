# OCI Compute Instance Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure compute instances using the OpenMCF API. Each example demonstrates a different use case, progressing from a minimal flex VM to a security-hardened bare metal host with advanced hardware tuning.

## Table of Contents

- [Example 1: Minimal Flex VM](#example-1-minimal-flex-vm)
- [Example 2: Private Application Server with Cloud-Init](#example-2-private-application-server-with-cloud-init)
- [Example 3: Burstable Development Instance](#example-3-burstable-development-instance)
- [Example 4: Preemptible Batch Worker](#example-4-preemptible-batch-worker)
- [Example 5: Security-Hardened Production VM](#example-5-security-hardened-production-vm)
- [Example 6: Bare Metal with Advanced Hardware Tuning](#example-6-bare-metal-with-advanced-hardware-tuning)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Flex VM

**Use Case:** A simple VM for a web server or development workload. Uses the AMD E4 Flex shape with 1 OCPU and 16 GB memory.

**Configuration:**
- **Shape:** VM.Standard.E4.Flex (1 OCPU, 16 GB)
- **Boot Source:** Platform image
- **Networking:** Public subnet with default IP assignment

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciComputeInstance
metadata:
  name: web-server
  org: my-org
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
    memoryInGbs: 16
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
  createVnicDetails:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
  metadata:
    ssh_authorized_keys: "ssh-rsa AAAA... user@workstation"
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f web-server.yaml
```

**What happens:**
- A 1-OCPU, 16 GB VM is created in the specified availability domain.
- The instance boots from the platform image and receives a private IP from the subnet's CIDR.
- Public IP assignment follows the subnet's default (public subnets assign; private subnets do not).
- SSH access is configured via the `ssh_authorized_keys` metadata key.
- The instance ID, IP addresses, boot volume ID, and availability domain are exported as stack outputs.

---

## Example 2: Private Application Server with Cloud-Init

**Use Case:** A production application server in a private subnet, bootstrapped via cloud-init, secured by NSGs, with a larger boot volume for application data.

**Configuration:**
- **Shape:** VM.Standard.E4.Flex (2 OCPUs, 32 GB)
- **Networking:** Private subnet, no public IP, NSG via `valueFrom`
- **Boot Volume:** 100 GiB, Higher Performance (20 VPUs/GB)
- **Cloud-Init:** Base64-encoded user_data script

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciComputeInstance
metadata:
  name: app-server
  org: acme-corp
  env: prod
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 2
    memoryInGbs: 32
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
    bootVolumeSizeInGbs: 100
    bootVolumeVpusPerGb: 20
  createVnicDetails:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: private-app-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciNetworkSecurityGroup
          name: app-nsg
          fieldPath: status.outputs.networkSecurityGroupId
    assignPublicIp: false
    hostnameLabel: "appserver"
  metadata:
    ssh_authorized_keys: "ssh-rsa AAAA... admin@workstation"
    user_data: "IyEvYmluL2Jhc2gKc2V0IC1leAoKIyBJbnN0YWxsIGFwcGxpY2F0aW9uIGRlcGVuZGVuY2llcwp5dW0gaW5zdGFsbCAteSBkb2NrZXItZW5naW5lCnN5c3RlbWN0bCBlbmFibGUgLS1ub3cgZG9ja2Vy"
  faultDomain: "FAULT-DOMAIN-1"
  instanceOptions:
    areLegacyImdsEndpointsDisabled: true
```

**What happens:**
- A 2-OCPU, 32 GB VM is created in a private subnet with no public IP.
- The NSG controls inbound/outbound traffic to the instance.
- The boot volume is 100 GiB with Higher Performance (20 VPUs/GB) for faster I/O.
- Cloud-init runs the base64-decoded `user_data` script on first boot (in this example: install Docker and enable the service).
- The hostname `appserver` is registered in the subnet's DNS domain.
- Legacy IMDSv1 endpoints are disabled for security.
- Fault domain `FAULT-DOMAIN-1` is pinned explicitly (useful when you need deterministic placement for paired instances).

---

## Example 3: Burstable Development Instance

**Use Case:** A low-cost development instance using burstable baseline utilization. The instance uses 1/8 of its OCPU baseline, bursting to full capacity when needed — ideal for intermittent workloads like development environments or staging servers.

**Configuration:**
- **Shape:** VM.Standard.E4.Flex (1 OCPU, 8 GB, 1/8 baseline)
- **Networking:** Public subnet for developer access

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciComputeInstance
metadata:
  name: dev-box
  org: my-org
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
    memoryInGbs: 8
    baselineOcpuUtilization: "BASELINE_1_8"
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
  createVnicDetails:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
    assignPublicIp: true
    hostnameLabel: "devbox"
  metadata:
    ssh_authorized_keys: "ssh-rsa AAAA... developer@laptop"
```

**What happens:**
- A burstable VM is created with a guaranteed baseline of 1/8 OCPU (12.5% of a full OCPU).
- The instance can burst to the full 1 OCPU when the workload demands it, with bursting credits managed by OCI.
- Lower baseline utilization reduces the hourly cost compared to a full 1-OCPU allocation.
- A public IP is explicitly assigned for direct SSH access from developer workstations.

---

## Example 4: Preemptible Batch Worker

**Use Case:** A cost-optimized instance for fault-tolerant batch processing. OCI can reclaim the instance at any time when capacity is needed, but the boot volume is preserved so work can resume on a new instance.

**Configuration:**
- **Shape:** VM.Standard.E4.Flex (8 OCPUs, 128 GB)
- **Preemptible:** Yes, with boot volume preservation
- **Networking:** Private subnet

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciComputeInstance
metadata:
  name: batch-worker-01
  org: acme-corp
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Ixxj:US-ASHBURN-AD-2"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 8
    memoryInGbs: 128
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
    bootVolumeSizeInGbs: 200
  createVnicDetails:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
    assignPublicIp: false
  preemptibleInstanceConfig:
    preserveBootVolume: true
  metadata:
    ssh_authorized_keys: "ssh-rsa AAAA... ops@workstation"
    user_data: "IyEvYmluL2Jhc2gKIyBSZXN1bWUgYmF0Y2ggam9iIGZyb20gbGFzdCBjaGVja3BvaW50Ci9vcHQvYmF0Y2gvcmVzdW1lLnNo"
```

**What happens:**
- A high-resource preemptible VM is created at reduced cost.
- When OCI reclaims the instance, the boot volume is preserved (200 GiB with checkpoint data).
- The cloud-init script resumes the batch job from its last checkpoint on new instance launches.
- Preemptible instances cannot be stopped and restarted — they can only be terminated. Plan for automated re-launch via external orchestration.

---

## Example 5: Security-Hardened Production VM

**Use Case:** A production VM with all available platform security features enabled, private networking, KMS-encrypted boot volume, and strict IMDS configuration.

**Configuration:**
- **Shape:** VM.Standard.E4.Flex (4 OCPUs, 64 GB)
- **Platform Security:** Secure Boot, Measured Boot, TPM, memory encryption
- **Boot Volume:** 200 GiB, KMS-encrypted, Higher Performance
- **Networking:** Private subnet, NSG-secured, no public IP
- **Availability:** Live migration preferred, restore on failure

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciComputeInstance
metadata:
  name: secure-vm
  org: acme-corp
  env: prod
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: secure-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 4
    memoryInGbs: 64
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
    bootVolumeSizeInGbs: 200
    bootVolumeVpusPerGb: 20
    kmsKeyId:
      value: "ocid1.key.oc1.iad.example"
  createVnicDetails:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: private-secure-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciNetworkSecurityGroup
          name: secure-nsg
          fieldPath: status.outputs.networkSecurityGroupId
    assignPublicIp: false
    hostnameLabel: "securevm"
    assignPrivateDnsRecord: true
  isPvEncryptionInTransitEnabled: true
  platformConfig:
    type: amd_vm
    isSecureBootEnabled: true
    isMeasuredBootEnabled: true
    isTrustedPlatformModuleEnabled: true
    isMemoryEncryptionEnabled: true
  instanceOptions:
    areLegacyImdsEndpointsDisabled: true
  availabilityConfig:
    isLiveMigrationPreferred: true
    recoveryAction: restore_instance
  agentConfig:
    pluginsConfig:
      - name: "Vulnerability Scanning"
        desiredState: enabled
      - name: "OS Management Service Agent"
        desiredState: enabled
  metadata:
    ssh_authorized_keys: "ssh-rsa AAAA... admin@workstation"
```

**What happens:**
- A security-hardened VM is created with Secure Boot (verifies boot software signatures), Measured Boot (records integrity measurements), TPM (secure key storage), and memory encryption (AMD SEV).
- The boot volume is encrypted at rest with a customer-managed KMS key and in transit via paravirtualized encryption.
- Legacy IMDSv1 endpoints are disabled — only the token-based IMDSv2 endpoint is accessible.
- Live migration is preferred during infrastructure maintenance to minimize downtime. On unplanned host failure, the instance is automatically restored on a new host.
- Vulnerability Scanning and OS Management plugins are explicitly enabled for compliance monitoring.
- A private DNS record is registered for the hostname `securevm` within the subnet's DNS domain.

---

## Example 6: Bare Metal with Advanced Hardware Tuning

**Use Case:** A bare metal host for workloads requiring direct hardware access — high-performance databases, HPC, or running hypervisors. Includes NUMA tuning, SMT control, and dedicated host placement.

**Configuration:**
- **Shape:** BM.Standard3.64 (bare metal, 64 OCPUs)
- **Placement:** Dedicated VM host for physical isolation
- **Platform:** NUMA, SMT, nested virtualization
- **Boot Source:** Clone from existing boot volume

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciComputeInstance
metadata:
  name: db-host
  org: acme-corp
  env: prod
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: database-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "BM.Standard3.64"
  sourceDetails:
    sourceType: boot_volume
    sourceId: "ocid1.bootvolume.oc1.iad.example"
  createVnicDetails:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: db-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciNetworkSecurityGroup
          name: db-nsg
          fieldPath: status.outputs.networkSecurityGroupId
    assignPublicIp: false
    hostnameLabel: "dbhost"
    skipSourceDestCheck: false
  dedicatedVmHostId:
    value: "ocid1.dedicatedvmhost.oc1.iad.example"
  platformConfig:
    type: intel_icelake_bm
    isSecureBootEnabled: true
    isMeasuredBootEnabled: true
    isTrustedPlatformModuleEnabled: true
    isSymmetricMultiThreadingEnabled: true
    areVirtualInstructionsEnabled: false
    numaNodesPerSocket: "NPS1"
    percentageOfCoresEnabled: 100
  isPvEncryptionInTransitEnabled: true
  availabilityConfig:
    isLiveMigrationPreferred: false
    recoveryAction: restore_instance
  launchOptions:
    bootVolumeType: "PARAVIRTUALIZED"
    networkType: "PARAVIRTUALIZED"
    firmware: uefi_64
    isPvEncryptionInTransitEnabled: true
    isConsistentVolumeNamingEnabled: true
  metadata:
    ssh_authorized_keys: "ssh-rsa AAAA... dba@workstation"
```

**What happens:**
- A bare metal host with 64 OCPUs is created on a dedicated VM host for physical isolation (licensing compliance or regulatory requirements).
- The instance boots from a cloned boot volume (`boot_volume` source type) — useful for gold-image workflows where a pre-configured volume is replicated across hosts.
- NUMA is configured as NPS1 (one NUMA node per socket) for database workloads that benefit from memory locality.
- SMT is enabled for maximum thread count; nested virtualization is disabled (this host runs a database, not a hypervisor).
- All 64 cores are enabled (`percentageOfCoresEnabled: 100`).
- Launch options explicitly set paravirtualized boot volume and network for maximum performance, UEFI firmware, in-transit encryption, and consistent volume naming.
- Live migration is disabled — bare metal hosts do not support live migration during maintenance. Recovery action is set to restore on a new host after failure.

---

## Common Operations

### Get Instance IP After Deployment

```bash
# Pulumi
pulumi stack output private_ip
pulumi stack output public_ip

# Terraform
terraform output private_ip
terraform output public_ip
```

### SSH into the Instance

```bash
# Public IP (if assigned)
ssh -i ~/.ssh/id_rsa opc@$(pulumi stack output public_ip)

# Private IP (via bastion or VPN)
ssh -i ~/.ssh/id_rsa opc@$(pulumi stack output private_ip)
```

The default user for Oracle Linux images is `opc`. Ubuntu images use `ubuntu`. The user depends on the image.

### Use Instance Outputs in Downstream Resources

The `instance_id` and `private_ip` outputs can be referenced by other resources. For example, associating a block volume (future OciBlockVolume component):

```yaml
spec:
  instanceId:
    valueFrom:
      kind: OciComputeInstance
      name: app-server
      fieldPath: status.outputs.instanceId
```

### Check Instance Boot Volume

```bash
# Get the boot volume ID
pulumi stack output boot_volume_id

# Terraform
terraform output boot_volume_id
```

---

## Best Practices

### Choose the Right Shape

| Workload | Recommended Shape | shapeConfig |
|----------|-------------------|-------------|
| Web server, API | VM.Standard.E4.Flex | 1-2 OCPUs, 8-16 GB |
| Application server | VM.Standard.E4.Flex | 2-4 OCPUs, 16-64 GB |
| Development box | VM.Standard.E4.Flex | 1 OCPU, 8 GB, `BASELINE_1_8` |
| CI/CD runner | VM.Standard.E4.Flex | 2-4 OCPUs, 16-32 GB, preemptible |
| Arm workloads | VM.Standard.A1.Flex | 1-80 OCPUs, flexible memory |
| Database host | BM.Standard3.64 | N/A (fixed shape) |
| HPC / GPU | BM.GPU4.8 (or similar) | N/A (fixed shape) |

**Flex shapes are the default choice.** Use bare metal only when you need direct hardware access, specific NUMA/SMT control, or licensing isolation. Use Arm (A1) shapes for cost-sensitive workloads that support aarch64.

### Networking Patterns

**Public instances** (development, bastion hosts):
- Use a public subnet
- Set `assignPublicIp: true` or omit (public subnets assign by default)
- Associate NSGs that restrict inbound access to necessary ports

**Private instances** (production application and database tiers):
- Use a private subnet
- Set `assignPublicIp: false`
- Route outbound traffic through a NAT Gateway (configured on the OciVcn)
- Access via bastion host, VPN, or OCI Bastion service

### Metadata Keys Reference

| Key | Format | Purpose |
|-----|--------|---------|
| `ssh_authorized_keys` | Newline-separated public keys | SSH access to the instance |
| `user_data` | Base64-encoded string | Cloud-init script executed on first boot |

**Cloud-init encoding:**

```bash
# Encode a script for user_data
base64 -w 0 < my-script.sh
```

The cloud-init script runs as root on first boot. Use it for package installation, service configuration, and application bootstrapping.

### Preemptible Instance Caveats

- Preemptible instances can be terminated at any time — design workloads to handle interruption gracefully.
- Set `preserveBootVolume: true` if the boot volume contains checkpoint data or build caches that should survive preemption.
- Preemptible instances cannot be stopped and restarted. After termination, you must create a new instance (from the preserved boot volume if applicable).
- Not all shapes and availability domains support preemptible instances. Check OCI documentation for current availability.
- Combine with cloud-init `user_data` to automatically resume work on new instance launches.

### Security Hardening Checklist

For production instances, consider enabling:

- `platformConfig.isSecureBootEnabled: true` — verifies boot software integrity
- `platformConfig.isMeasuredBootEnabled: true` — records boot measurements in TPM
- `platformConfig.isTrustedPlatformModuleEnabled: true` — enables secure key storage
- `platformConfig.isMemoryEncryptionEnabled: true` — encrypts memory (AMD SEV / Intel TME)
- `instanceOptions.areLegacyImdsEndpointsDisabled: true` — forces IMDSv2 token-based access
- `isPvEncryptionInTransitEnabled: true` — encrypts boot/data volume traffic in transit
- `createVnicDetails.assignPublicIp: false` — no direct internet exposure
- `sourceDetails.kmsKeyId` — customer-managed encryption key for the boot volume

### Tag for Cost and Compliance

Metadata labels are applied as OCI freeform tags. Use consistent labels across all instances:

```yaml
metadata:
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
    workload: api-server
```
