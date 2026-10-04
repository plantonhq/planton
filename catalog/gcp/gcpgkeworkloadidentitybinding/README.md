# GCP GKE Workload Identity Binding

Grants a Kubernetes ServiceAccount (KSA) the right to impersonate a Google
Service Account (GSA) — the GCP half of GKE Workload Identity, and the
keyless alternative to exported service account keys.

## Purpose

GKE Workload Identity is the recommended way for applications running in GKE
to authenticate to Google Cloud services: pods mint short-lived tokens as a
GSA instead of mounting a distributable key file. The GCP side of the
handshake is one precise IAM grant — `roles/iam.workloadIdentityUser` on the
GSA, to the principal
`serviceAccount:{project}.svc.id.goog[{namespace}/{ksa}]`.

This kind owns exactly that grant, and constructs the brittle principal
string from simple validated inputs — a typo'd principal is impossible by
construction, and namespace/name are validated against Kubernetes naming
rules before anything deploys.

## What This Kind Does — and Does Not — Manage

- **Managed here (GCP side):** the additive IAM grant on the GSA. Additive
  means it merges into the GSA's IAM policy without touching any other
  principal's bindings, and removal subtracts only this exact grant.
- **Not managed here (Kubernetes side):** the
  `iam.gke.io/gcp-service-account: <gsa-email>` annotation on the KSA. The
  annotation lives on the Kubernetes object and belongs to the workload's
  own deployment (its chart or manifest) — the same place the KSA itself is
  created. The `service_account_email` output of this component is exactly
  the value that annotation needs.

## Features

- **Zero-credential architecture**: no service account keys to rotate or
  leak
- **Pod-level IAM identity**: each workload can have distinct GCP
  permissions
- **Constructed principal**: the `{project}.svc.id.goog[{ns}/{ksa}]` string
  is derived from validated parts, never hand-typed
- **Optional IAM condition**: time-box or scope the grant with a CEL
  condition
- **Clear audit trail**: Cloud Audit Logs show exactly which KSA
  impersonated which GSA

## Prerequisites

- GKE cluster with Workload Identity enabled (the cluster project's
  implicit pool is `<project>.svc.id.goog`)
- Node pools configured with `workloadMetadataConfig.mode = GKE_METADATA`
- A Google Service Account (GSA) with the IAM roles your workload needs
- A Kubernetes ServiceAccount (KSA) that your pods use, annotated with
  `iam.gke.io/gcp-service-account: <gsa-email>` by the workload's own
  deployment

## Basic Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: cert-manager-dns-binding
spec:
  projectId:
    value: prod-project-123
  serviceAccountEmail:
    valueFrom:
      kind: GcpServiceAccount
      name: dns01-solver
      fieldPath: status.outputs.email
  ksaNamespace: cert-manager
  ksaName: cert-manager
```

This grant allows pods using the `cert-manager` ServiceAccount in the
`cert-manager` namespace to impersonate the referenced GSA.

## Common Use Cases

- **cert-manager with Cloud DNS**: bind cert-manager to a GSA holding
  `roles/dns.admin` to solve DNS-01 ACME challenges.
- **external-dns**: bind external-dns to a GSA that manages zone records.
- **Application workloads**: bind an app's KSA to a GSA holding exactly the
  roles the app needs (Cloud SQL client, GCS object access, Pub/Sub
  publisher) — one binding per KSA-GSA pair.

## Cross-Project Bindings

The GSA may live in a different project than the GKE cluster. `projectId`
is always the CLUSTER's project (it names the workload-identity pool); the
GSA's own project is inferred from its email.

## Outputs

| Output | Description |
|--------|-------------|
| `member` | The constructed workload-identity principal added to the GSA's policy |
| `service_account_email` | The bound GSA email — the value the KSA annotation needs |

## Related Kinds

- **GcpServiceAccount** — creates the GSA this binding targets
- **GcpGkeCluster** — the cluster whose project hosts the identity pool
- **GcpProjectIamMember** — grants the GSA its project-level roles

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
