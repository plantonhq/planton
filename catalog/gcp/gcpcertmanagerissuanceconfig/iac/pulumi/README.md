# GcpCertManagerIssuanceConfig Pulumi Module

This Pulumi module creates one Certificate Manager certificate issuance config and enables the Certificate Manager API on the target project.

## Usage

This module is typically invoked by the Planton CLI, but can also be used directly.

### With Planton CLI

```bash
planton pulumi up --manifest issuance-config.yaml
```

### Standalone Usage

1. Set the IaC input as an environment variable:

```bash
export PLANTON_CATALOG_OBJECT_MANIFEST=$(cat <<EOT
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerIssuanceConfig
metadata:
  name: internal-tls-issuance
spec:
  projectId:
    value: my-gcp-project
  caPool:
    value: projects/my-gcp-project/locations/us-central1/caPools/internal-pool
  keyAlgorithm: ECDSA_P256
  lifetime: 2592000s
  rotationWindowPercentage: 66
EOT
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

The module reads its configuration from the `GcpCertManagerIssuanceConfigIacInput` proto message:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| target | GcpCertManagerIssuanceConfig | Yes | The issuance config resource manifest |
| provider_config | GcpProviderConfig | Yes | GCP provider configuration |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| issuance_config_id | string | Full resource name — what a certificate's `managed.issuance_config` takes |
| issuance_config_name | string | Issuance config name in GCP |
| location | string | The Certificate Manager location (`global` unless set) |

## Behavior Notes

- An empty `location` is not sent; the provider defaults it to `global`, identically on both engines, and the `location` output reports the effective value.
- `ca_pool` is wrapped into the provider's two nested single-field blocks.
- Every argument except labels and the deletion policy forces replacement.

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs.
