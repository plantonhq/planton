# GCP GKE Workload Identity Binding

Creates an IAM policy binding that allows a Kubernetes ServiceAccount (KSA) in a GKE cluster to impersonate a Google ServiceAccount (GSA) via Workload Identity Federation. This eliminates the need for exported JSON service account keys, enabling GKE pods to authenticate to GCP APIs using the cluster's Workload Identity pool. The IAM member principal is constructed from its parts — `serviceAccount:{project}.svc.id.goog[{namespace}/{name}]` — so a typo'd principal string is impossible by construction.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **IAM Member Binding** -- a `serviceaccount.IAMMember` that grants `roles/iam.workloadIdentityUser` on the specified Google ServiceAccount to the Kubernetes ServiceAccount, using the member format `serviceAccount:{projectId}.svc.id.goog[{namespace}/{name}]`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GKE cluster with Workload Identity enabled** in the project specified by `projectId`. The cluster must have a Workload Identity pool (`{projectId}.svc.id.goog`) configured.
- **A Google ServiceAccount** that the KSA will impersonate. The GSA must exist in the project and have the IAM roles needed by the workload (e.g., `roles/cloudsql.client`, `roles/storage.objectViewer`).
- **A Kubernetes ServiceAccount** in the specified namespace. The KSA must be annotated with `iam.gke.io/gcp-service-account={gsa-email}` for the binding to take effect.
- **IAM Service Account Credentials API** (`iamcredentials.googleapis.com`) enabled in the target project.

## Deploy

### Console

Open the deployment store, find **GCP GKE Workload Identity Binding**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Workload Identity Binding** preset in the [Presets](#presets) tab to pre-populate a typical KSA-to-GSA binding.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: cert-manager-binding
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  serviceAccountEmail:
    value: "cert-manager@acme-prod-12345.iam.gserviceaccount.com"
  ksaNamespace: cert-manager
  ksaName: cert-manager
```

```shell
planton apply -f workload-identity-binding.yaml
```

This grants the `cert-manager` Kubernetes ServiceAccount permission to impersonate the specified Google ServiceAccount via Workload Identity. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the binding to a GCP project and service account deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  serviceAccountEmail:
    valueFrom:
      kind: GcpServiceAccount
      name: cert-manager-sa
      fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, deploys the project and service account first, then creates the Workload Identity binding with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Workload Identity binding. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**KSA namespace and name** -- The `ksaNamespace` and `ksaName` fields must exactly match the Kubernetes ServiceAccount that pods use. The binding is namespace-scoped -- a KSA in namespace `default` cannot use a binding created for namespace `cert-manager`.

**Google ServiceAccount selection** -- The GSA referenced by `serviceAccountEmail` determines what GCP permissions the KSA inherits. Follow least-privilege principles: create purpose-specific GSAs with only the IAM roles the workload needs, rather than reusing a broadly-permissioned account.

**One binding per pair** -- Each unique KSA-to-GSA pair requires its own binding. If multiple KSAs need to impersonate the same GSA, create separate bindings for each.

**IAM Condition** -- an optional CEL expression scoping WHEN the binding applies (a time-boxed migration window is the everyday case). The condition is part of the grant's identity: the same binding with and without a condition are two independent grants.

**Everything replaces atomically** -- an IAM grant has no update. Changing the project, GSA, KSA coordinates, or condition replaces the grant (a brief moment where the KSA cannot mint tokens); the replacement destroys nothing and is GCP's own designed change workflow.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** | `serviceAccountEmail` | `status.outputs.email` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `member` | IAM member string (e.g., `serviceAccount:project.svc.id.goog[ns/sa]`) | Audit log queries, IAM policy inspection |
| `service_account_email` | Bound GSA email (echoed from spec) | KSA annotation value, documentation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard binding** -- A single KSA-to-GSA binding for common use cases like cert-manager DNS validation, external-dns record management, or application access to Cloud SQL, Cloud Storage, or Pub/Sub. Provide the project, GSA email, and KSA coordinates. Start from the **Standard Workload Identity Binding** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project that hosts the GKE Workload Identity pool
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- provides the Google ServiceAccount that the KSA impersonates