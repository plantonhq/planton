# GcpProject Pulumi Module

This Pulumi module creates one Google Cloud project under an organization
or folder — billing linked, labels/tags applied, the default network
suppressed by default, requested APIs pre-enabled, and the configured
deletion policy applied.

## Usage

This module is typically invoked by the Planton CLI, but can also be used directly.

### With Planton CLI

```bash
planton pulumi up --manifest project.yaml
```

### Standalone Usage

1. Set the stack input as an environment variable:

```bash
export PLANTON_CLOUD_RESOURCE_MANIFEST=$(cat <<EOF
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpProject
metadata:
  name: prod-workloads
spec:
  projectId: acme-prod-workloads
  parentType: folder
  parentId: "123456789012"   # or folderId: {valueFrom: {kind: GcpFolder, name: ...}}
  billingAccountId: 0123AB-4567CD-89EFGH
  deletionPolicy: PREVENT
  enabledApis:
    - compute.googleapis.com
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

The module reads its configuration from the `GcpProjectStackInput` proto message:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| target | GcpProject | Yes | The project resource manifest |
| provider_config | GcpProviderConfig | Yes | GCP provider configuration |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| project_id | string | The unique project ID — resolved by every other kind's project reference |
| project_number | string | The numeric identifier assigned by Google |
| name | string | The display name |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs. The scopes are split:
project creation authorizes on the parent organization or folder, and the
billing link authorizes on the billing account itself.

## Implementation Notes

- IAM grants are separate `GcpProjectIamMember` resources; this module
  never touches the project's IAM policy.
- `deletion_policy` defaults to DELETE explicitly so destroy semantics are
  identical on both engines.
- API enablement uses `disableOnDestroy: false`: removing an entry (or the
  project resource) never disables a service other tooling depends on.
