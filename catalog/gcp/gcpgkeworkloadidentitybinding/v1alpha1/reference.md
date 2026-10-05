# GcpGkeWorkloadIdentityBinding

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpGkeWorkloadIdentityBindingSpec defines one ADDITIVE IAM grant on a
Google Service Account (GSA): it gives a Kubernetes ServiceAccount (KSA)
the roles/iam.workloadIdentityUser role on that GSA, which is exactly what
GKE Workload Identity needs for the KSA to mint tokens as the GSA — keyless
impersonation, no exported key material.

The IAM member principal is constructed from its parts —
  serviceAccount:<pool-project>.svc.id.goog[<ksa-namespace>/<ksa-name>]
— so a typo'd principal string is impossible by construction. The pool
project is the project that hosts the GKE cluster (every project has one
implicit workload-identity pool named <project>.svc.id.goog).

Completing the workload-identity handshake also requires the KSA to carry
the iam.gke.io/gcp-service-account=<gsa-email> annotation. That annotation
lives on the Kubernetes object and belongs to the workload's own
deployment (chart or manifest) — this component owns only the GCP side.

Additive semantics: this grant merges into the GSA's IAM policy without
touching any other principal's bindings, and removal subtracts only this
exact grant. Every field is immutable, mirroring the underlying API: an
IAM grant has no update — any change replaces the grant atomically.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: test-wib
spec:
  projectId:
    value: test-gcp-project
  serviceAccountEmail:
    value: test-gsa@test-gcp-project.iam.gserviceaccount.com
  ksaNamespace: cert-manager
  ksaName: cert-manager
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.ksaNamespace` | `string` | yes |  |  |
| `spec.ksaName` | `string` | yes |  |  |
| `spec.condition` | `GcpGkeWorkloadIdentityBindingCondition` |  |  |  |
| `spec.condition.title` | `string` | yes |  |  |
| `spec.condition.expression` | `string` | yes |  |  |
| `spec.condition.description` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that hosts the GKE cluster — and therefore the implicit
workload-identity pool <project>.svc.id.goog the principal lives in.
This may differ from the GSA's own project in cross-project setups; the
GSA's project is derived from its email.
If omitted, the provider's default project is used — the common case
when the cluster lives in the credentials' project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.serviceAccountEmail

`string | valueFrom` · required

The email of the Google Service Account the KSA impersonates. The grant
is attached to THIS service account's IAM policy.
Example: "cert-manager@my-project.iam.gserviceaccount.com"

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.ksaNamespace

`string` · required

Kubernetes namespace of the ServiceAccount running in the cluster.
Validated as an RFC 1123 label — a typo here would otherwise produce a
syntactically valid but permanently broken principal.

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.ksaName

`string` · required

Name of the Kubernetes ServiceAccount. Validated as an RFC 1123 DNS
subdomain (Kubernetes object-name rules).

- rule: {"required":true,"string":{"maxLen":"253","pattern":"^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$"}}

### spec.condition

`GcpGkeWorkloadIdentityBindingCondition`

Optional IAM Condition restricting when this grant applies. The
condition is part of the grant's identity: the same grant with and
without a condition are two independent grants that do not interfere.

### spec.condition.title

`string` · required

Short human-readable title identifying the condition's intent,
e.g. "expires-2026-12-31".

- rule: {"required":true,"string":{"maxLen":"100"}}

### spec.condition.expression

`string` · required

The CEL condition expression, e.g.
request.time < timestamp("2027-01-01T00:00:00Z").

- rule: {"required":true}

### spec.condition.description

`string`

Optional longer explanation of what the condition does and why it exists.

- rule: {"string":{"maxLen":"256"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGkeWorkloadIdentityBinding, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.member` | `string` | The IAM member string added to the policy, e.g. "serviceAccount:my-project.svc.id.goog[cert-manager/cert-manager]". |
| `status.outputs.service_account_email` | `string` | The bound GSA email (echoed from spec for convenience). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)
