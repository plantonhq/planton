# Hardened Encrypted OKE Node Pool

This preset creates a security-hardened OKE node pool with customer-managed KMS encryption for boot volumes, in-transit encryption for paravirtualized volume attachments, and explicit fault domain constraints for maximum failure isolation. SSH access is intentionally omitted -- use OCI Bastion for node access in regulated environments. This preset pairs with the `02-private-cluster` OKE cluster preset for end-to-end private, encrypted Kubernetes infrastructure.

## When to Use

- Regulated industries (finance, healthcare, government) with compliance requirements for encryption at rest and in transit
- Enterprise environments where security policy mandates customer-managed encryption keys for all compute resources
- Clusters handling PII, PHI, or other sensitive data where node-level encryption is a hard requirement
- Multi-tenant environments where fault domain isolation is needed to limit blast radius of hardware failures
- Any deployment where SSH access to worker nodes is prohibited by policy (Bastion-only access model)

## Key Configuration Choices

- **4 OCPUs / 64 GB per node** (`nodeShapeConfig`) -- Larger nodes for higher pod density, reducing the total number of nodes and the associated management overhead. In regulated environments, fewer larger nodes also mean fewer attack surface points. The 16 GB/OCPU ratio remains the general-purpose standard.
- **KMS boot volume encryption** (`nodeConfigDetails.kmsKeyId`) -- Encrypts the boot volume of every node in this pool with a customer-managed key from OCI Vault. Without this, boot volumes use Oracle-managed encryption. Customer-managed keys provide control over key rotation schedules, access policies, and centralized audit logging through OCI Audit.
- **In-transit encryption** (`nodeConfigDetails.isPvEncryptionInTransitEnabled: true`) -- Encrypts data moving between the compute instance and its paravirtualized boot/block volume attachments. This protects against network-level interception within OCI's data center fabric. Required for many compliance frameworks (PCI DSS, HIPAA).
- **Fault domain constraints** (`placementConfigs[].faultDomains`) -- Each placement config restricts nodes to specific fault domains within the AD. Fault domains represent isolated power, cooling, and networking within an AD. Constraining to FD-1 and FD-2 (out of 3) ensures nodes spread across physical infrastructure while leaving FD-3 available for other workloads or future scaling.
- **100 GB boot volume** (`nodeSourceDetails.bootVolumeSizeInGbs: 100`) -- Larger than the 50 GB default to accommodate audit agent logs, container image cache, and ephemeral storage for compliance-sensitive workloads that generate significant local data before offloading to persistent storage.
- **Explicit image OCID** (`nodeSourceDetails.imageId`) -- Unlike preset 01 which relies on OKE's default image, this preset pins a specific image OCID. In regulated environments, OS images must be pre-approved, scanned for vulnerabilities, and tracked. Use a CIS-hardened Oracle Linux image from your approved image catalog.
- **No SSH key** -- Intentionally omitted. In hardened environments, SSH access to worker nodes should go through OCI Bastion sessions that provide auditable, time-limited access. Embedding an SSH key in the node pool creates a persistent, unauditable access path.
- **30-minute eviction grace** (`nodeEvictionSettings.evictionGraceDuration: PT30M`) -- Shorter than the standard 60-minute window. Regulated environments often have SLAs requiring that security patches and version upgrades complete within a defined maintenance window. The 30-minute grace balances workload safety with upgrade timeliness.
- **Node label for scheduling** (`initialNodeLabels: pool=hardened`) -- Enables workloads with compliance requirements to be scheduled exclusively on hardened nodes using `nodeSelector` or affinity rules.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the node pool will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<oke-cluster-ocid>` | OCID of the OKE cluster to attach this node pool to | OCI Console > Developer Services > Kubernetes Clusters, or `OciContainerEngineCluster` status outputs |
| `<kubernetes-version>` | Kubernetes version for the worker nodes (e.g., `v1.30.1`) | Should match or be compatible with the cluster's version. `oci ce node-pool-options get --node-pool-option-id all` |
| `<node-image-ocid>` | OCID of an approved, CIS-hardened Oracle Linux image for nodes | Your organization's approved image catalog, or `oci ce node-pool-options get --node-pool-option-id all` |
| `<availability-domain-1>` | First availability domain name (e.g., `Uocm:PHX-AD-1`) | `oci iam availability-domain list` or OCI Console |
| `<availability-domain-2>` | Second availability domain name (e.g., `Uocm:PHX-AD-2`) | Same as above |
| `<availability-domain-3>` | Third availability domain name (e.g., `Uocm:PHX-AD-3`) | Same as above |
| `<private-worker-subnet-ocid>` | OCID of a private subnet for worker node VNICs (no Internet Gateway route) | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<worker-nsg-ocid>` | OCID of the NSG applied to worker node VNICs | OCI Console > Networking > Network Security Groups, or `OciSecurityGroup` status outputs |
| `<kms-key-ocid>` | OCID of the KMS key for boot volume encryption at rest | OCI Console > Identity & Security > Vault > Keys, or `OciKmsKey` status outputs |
| `<pod-subnet-ocid>` | OCID of the subnet for pod IP allocation (VCN-native CNI) | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<pod-nsg-ocid>` | OCID of the NSG applied to pod VNICs | OCI Console > Networking > Network Security Groups, or `OciSecurityGroup` status outputs |

## Related Presets

- **01-standard-production** -- Use instead when customer-managed encryption and strict hardening are not required
- **03-preemptible-dev** -- Use instead for development clusters where security hardening is unnecessary
