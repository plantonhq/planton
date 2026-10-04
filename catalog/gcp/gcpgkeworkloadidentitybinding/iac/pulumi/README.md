# GcpGkeWorkloadIdentityBinding Pulumi Module

This Pulumi module creates one additive IAM grant on a Google Service
Account: `roles/iam.workloadIdentityUser` to the constructed
workload-identity principal
`serviceAccount:<project>.svc.id.goog[<namespace>/<ksa>]`.

## Usage

This module is typically invoked by the Planton CLI, but can also be used directly.

### With Planton CLI

```bash
planton pulumi up --manifest workload-identity-binding.yaml
```

### Standalone Usage

1. Set the IaC input as an environment variable. The module reads a `GcpGkeWorkloadIdentityBindingIacInput` from `IAC_INPUT_YAML` (YAML content), `IAC_INPUT_YAML_FILE` (a path to that YAML) or the Pulumi config key `planton:iac-input`; the manifest goes under `target`:

```bash
export IAC_INPUT_YAML=$(cat <<EOF
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpGkeWorkloadIdentityBinding
  metadata:
    name: cert-manager-binding
  spec:
    projectId:
      value: prod-project
    serviceAccountEmail:
      value: dns01-solver@prod-project.iam.gserviceaccount.com
    ksaNamespace: cert-manager
    ksaName: cert-manager
EOF
)
```

2. Configure GCP credentials:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

3. Run Pulumi:

```bash
pulumi up
```

## Inputs

The module reads its configuration from the `GcpGkeWorkloadIdentityBindingIacInput` proto message:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| target | GcpGkeWorkloadIdentityBinding | Yes | The resource manifest |
| provider_config | GcpProviderConfig | Yes | GCP provider configuration |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| member | string | The constructed workload-identity principal |
| service_account_email | string | The bound GSA email — the value the KSA's `iam.gke.io/gcp-service-account` annotation needs |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs on the target service account.

## Implementation Notes

- The provider requires the fully-qualified service-account resource name;
  the module constructs `projects/-/serviceAccounts/<email>` — the IAM
  API's `-` wildcard infers the SA's project from the email, so
  cross-project GSAs work without extra configuration. The Terraform
  module constructs the identical value.
- The grant is additive and immutable: it merges into the GSA's policy
  without touching other bindings, removal subtracts only this grant, and
  any spec change replaces the grant atomically.
- The `iam.gke.io/gcp-service-account` annotation on the KSA is the
  Kubernetes half of the handshake and belongs to the workload's own
  deployment.

## Troubleshooting

1. **Pods still get permission errors**: verify the KSA carries the
   `iam.gke.io/gcp-service-account` annotation and the node pool runs with
   `workloadMetadataConfig.mode = GKE_METADATA`, then check
   `gcloud iam service-accounts get-iam-policy <gsa-email>` for the member.
2. **Grant already exists**: an identical binding created outside IaC is
   absorbed on the next apply — additive members are idempotent per
   (role, member, condition).
