---
title: "Private OKE Cluster"
description: "This preset creates a fully private Enhanced OKE cluster with no public API endpoint, customer-managed KMS encryption for Kubernetes secrets at rest, and container image signature verification. The..."
type: "preset"
rank: "02"
presetSlug: "02-private-cluster"
componentSlug: "container-engine-cluster"
componentTitle: "Container Engine Cluster"
provider: "oci"
icon: "package"
order: 2
---

# Private OKE Cluster

This preset creates a fully private Enhanced OKE cluster with no public API endpoint, customer-managed KMS encryption for Kubernetes secrets at rest, and container image signature verification. The API server is accessible only from within the VCN (or via VPN/FastConnect/Bastion), making this the standard pattern for regulated industries, financial services, healthcare, and any environment where zero public exposure is a hard requirement.

## When to Use

- Enterprise environments where security policy prohibits public Kubernetes API endpoints
- Regulated industries (finance, healthcare, government) with compliance requirements for private control planes
- Clusters handling sensitive data that require customer-managed encryption keys for Kubernetes secrets
- Environments that enforce container image signing to prevent deployment of untrusted images
- Multi-tenant OCI tenancies where network isolation between teams is critical

## Key Configuration Choices

- **Private API endpoint** (`endpointConfig.isPublicIpEnabled: false`) -- The Kubernetes API server has no public IP and is reachable only from within the VCN. Access from outside the VCN requires an OCI Bastion session, site-to-site VPN, or FastConnect. This eliminates the entire class of attacks that target publicly exposed API servers.
- **NSG on the API endpoint** (`endpointConfig.nsgIds`) -- Even within the VCN, the API endpoint is protected by a Network Security Group. Configure ingress rules to allow only specific subnets (bastion subnet, CI/CD runner subnet, operator workstation subnet) on port 6443.
- **Private API endpoint subnet** -- The endpoint subnet placeholder is named `<private-api-endpoint-subnet-ocid>` to emphasize that this must be a private subnet (no route to an Internet Gateway). Using a public subnet with `isPublicIpEnabled: false` works but is misleading; use a dedicated private subnet for clarity.
- **KMS encryption for Kubernetes secrets** (`kmsKeyId`) -- Encrypts all Kubernetes Secret objects at rest using a customer-managed key in OCI Vault. Without this, secrets are encrypted with an Oracle-managed key. Customer-managed keys give you full control over key rotation, access policies, and audit logging.
- **Image signature verification** (`imagePolicyConfig.isPolicyEnabled: true`) -- Requires all container images deployed to the cluster to be signed with one of the specified KMS keys. Unsigned or incorrectly signed images are rejected by the API server. The signing KMS key can be the same as or different from the secrets encryption key.
- **Enhanced cluster type and VCN-native CNI** -- Same as 01-standard-production. Enhanced type is required for workload identity (pods authenticating to OCI without static credentials), which is especially important in private clusters where you want to avoid distributing API keys.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the cluster will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN hosting the cluster | OCI Console > Networking > VCNs, or `OciVcn` status outputs |
| `<kubernetes-version>` | Kubernetes version for the control plane (e.g., `v1.30.1`) | `oci ce cluster-options list --cluster-option-id all` or OCI Console > Developer Services > Kubernetes Clusters > Create |
| `<private-api-endpoint-subnet-ocid>` | OCID of a private subnet hosting the API server endpoint | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<api-endpoint-nsg-ocid>` | OCID of the NSG controlling access to the API server endpoint | OCI Console > Networking > VCNs > Network Security Groups, or `OciSecurityGroup` status outputs |
| `<service-lb-subnet-ocid>` | OCID of the subnet where Kubernetes Service load balancers will be placed | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<kms-key-ocid>` | OCID of the KMS key for encrypting Kubernetes secrets at rest | OCI Console > Identity & Security > Vault > Keys, or `OciKmsKey` status outputs |
| `<image-verification-kms-key-ocid>` | OCID of the KMS key used for container image signature verification | OCI Console > Identity & Security > Vault > Keys, or `OciKmsKey` status outputs |

## Related Presets

- **01-standard-production** -- Use instead when a public API endpoint is acceptable and KMS encryption is not required
- **03-development** -- Use instead for dev/test clusters where security hardening is unnecessary
