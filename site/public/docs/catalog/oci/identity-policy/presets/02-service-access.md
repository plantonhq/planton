---
title: "Dynamic Group Service Access Policy"
description: "This preset creates an IAM policy granting a dynamic group access to specific OCI services. Dynamic groups are OCI's workload identity mechanism -- they let compute instances, OKE pods, and Functions..."
type: "preset"
rank: "02"
presetSlug: "02-service-access"
componentSlug: "identity-policy"
componentTitle: "Identity Policy"
provider: "oci"
icon: "package"
order: 2
---

# Dynamic Group Service Access Policy

This preset creates an IAM policy granting a dynamic group access to specific OCI services. Dynamic groups are OCI's workload identity mechanism -- they let compute instances, OKE pods, and Functions authenticate to OCI services without embedded credentials. This is the OCI equivalent of AWS IAM roles for EC2/EKS or GCP workload identity. After creating an `OciDynamicGroup` that matches your workload instances, this policy grants those instances the permissions they need.

## When to Use

- Compute instances that need to read secrets from OCI Vault, use KMS encryption keys, or access Object Storage
- OKE (Kubernetes) pods using workload identity to access OCI services without storing credentials in the cluster
- Functions (serverless) that need to interact with other OCI services as part of their execution
- Any automation or application running on OCI infrastructure that needs to call OCI APIs using instance principal authentication

## Key Configuration Choices

- **Least-privilege verbs** -- Each statement uses the minimum verb required for the operation. `read` for secrets (retrieve values), `use` for keys (encrypt/decrypt but not create/delete), `manage` for object storage (full CRUD on buckets and objects), `use` for networking (attach to subnets/VNICs but not create/delete them). This follows the principle of least privilege rather than granting `manage all-resources`.
- **Specific resource families** -- Statements target specific resource types (`secret-family`, `keys`, `object-family`, `virtual-network-family`) rather than `all-resources`. This limits the blast radius if a workload is compromised. Add or remove statements based on what your workload actually needs.
- **Dynamic group subject** -- Uses `dynamic-group` instead of `group`. Dynamic group membership is determined by matching rules (e.g., all instances in a compartment, all instances with a specific tag) rather than explicit user assignment. The dynamic group must be created separately using `OciDynamicGroup`.
- **Compartment-scoped** -- The policy is attached to the compartment where the target services reside. If your workload needs to access resources across multiple compartments, create a separate policy in each compartment or use a tenancy-level policy.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment this policy is attached to | OCI Console > Identity > Compartments, or `OciCompartment` status outputs (`compartmentId`) |
| `<dynamic-group-name>` | Name of the dynamic group whose members receive these permissions | OCI Console > Identity > Dynamic Groups, or the `name`/`metadata.name` of the `OciDynamicGroup` resource |
| `<compartment-name>` | Display name of the compartment containing the target resources | OCI Console > Identity > Compartments, or the `name`/`metadata.name` of the `OciCompartment` resource |

## Related Presets

- **01-compartment-admin** -- Use instead when granting a human user group full administrative access to a compartment
- **03-read-only-auditor** -- Use instead when granting inspect-level visibility for compliance or security audit purposes
