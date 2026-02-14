---
title: "GKE Workload Identity Binding"
description: "GKE Workload Identity Binding deployment documentation"
icon: "package"
order: 100
componentName: "gcpgkeworkloadidentitybinding"
---

# GCP GKE Workload Identity Binding

Creates an IAM policy binding that allows a Kubernetes ServiceAccount (KSA) in a GKE cluster to impersonate a Google Service Account (GSA) via Workload Identity Federation. This component grants `roles/iam.workloadIdentityUser` on the target GSA so the specified KSA can authenticate as that GSA without managing keys.

## What Gets Created

When you deploy a GcpGkeWorkloadIdentityBinding resource, OpenMCF provisions:

- **IAM Member Binding** — a `google_service_account_iam_member` resource that grants the `roles/iam.workloadIdentityUser` role on the target GSA to the Workload Identity member principal (`serviceAccount:<project>.svc.id.goog[<namespace>/<ksa-name>]`)

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** with a GKE cluster that has Workload Identity enabled
- **A Google Service Account** (GSA) that the Kubernetes workload should impersonate
- **A Kubernetes ServiceAccount** (KSA) already created (or planned) in the target namespace and cluster

## Quick Start

Create a file `workload-identity-binding.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: cert-manager-binding
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpGkeWorkloadIdentityBinding.cert-manager-binding
spec:
  projectId: my-gcp-project-123
  serviceAccountEmail: cert-manager@my-gcp-project-123.iam.gserviceaccount.com
  ksaNamespace: cert-manager
  ksaName: cert-manager
```

Deploy:

```shell
openmcf apply -f workload-identity-binding.yaml
```

This grants the `cert-manager` Kubernetes ServiceAccount in the `cert-manager` namespace permission to impersonate the `cert-manager` GSA.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` or `valueFrom` | The GCP project that hosts the GKE cluster (and its Workload Identity pool `<project>.svc.id.goog`). Can be a literal value or a reference to a GcpProject resource. | Required |
| `serviceAccountEmail` | `string` or `valueFrom` | Email of the Google Service Account to impersonate. Can be a literal value or a reference to a GcpServiceAccount resource. | Required |
| `ksaNamespace` | `string` | Kubernetes namespace of the ServiceAccount that will assume the GSA identity. | Required |
| `ksaName` | `string` | Name of the Kubernetes ServiceAccount that will assume the GSA identity. | Required |

### Optional Fields

This component has no optional fields. All spec fields are required.

## Examples

### Binding for cert-manager

Allow the cert-manager controller to use a GSA for DNS-01 ACME challenges:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: cert-manager-binding
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpGkeWorkloadIdentityBinding.cert-manager-binding
spec:
  projectId: my-gcp-project-123
  serviceAccountEmail: cert-manager@my-gcp-project-123.iam.gserviceaccount.com
  ksaNamespace: cert-manager
  ksaName: cert-manager
```

### Binding for an Application Workload

Grant a backend service running in the `payments` namespace access to a GSA with Cloud SQL and Pub/Sub permissions:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: payments-api-binding
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGkeWorkloadIdentityBinding.payments-api-binding
spec:
  projectId: my-gcp-project-123
  serviceAccountEmail: payments-api@my-gcp-project-123.iam.gserviceaccount.com
  ksaNamespace: payments
  ksaName: payments-api
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding the project ID and service account email:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: external-dns-binding
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGkeWorkloadIdentityBinding.external-dns-binding
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  serviceAccountEmail:
    valueFrom:
      kind: GcpServiceAccount
      name: external-dns
      fieldPath: status.outputs.email
  ksaNamespace: external-dns
  ksaName: external-dns
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `member` | `string` | The IAM member string added to the policy, e.g. `serviceAccount:my-project.svc.id.goog[external-dns/external-dns]` |
| `serviceAccountEmail` | `string` | The bound GSA email (echoed from spec for convenience) |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project that hosts the GKE cluster
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — creates the Google Service Account referenced by `serviceAccountEmail`
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — provisions the GKE cluster where Workload Identity is enabled
- [GcpDnsZone](/docs/catalog/gcp/dns-zone) — commonly paired with Workload Identity bindings for cert-manager or external-dns
