# GcpCertManagerDnsAuthorization Pulumi Module

This Pulumi module creates one Certificate Manager DNS authorization and
enables the Certificate Manager API on the target project.

## Usage

This module is typically invoked by the Planton CLI, but can also be used directly.

### With Planton CLI

```bash
planton pulumi up --manifest dns-authorization.yaml
```

### Standalone Usage

1. Set the IaC input as an environment variable. The module reads a `GcpCertManagerDnsAuthorizationIacInput` from `IAC_INPUT_YAML` (YAML content), `IAC_INPUT_YAML_FILE` (a path to that YAML) or the Pulumi config key `planton:iac-input`; the manifest goes under `target`:

```bash
export IAC_INPUT_YAML=$(cat <<EOF
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpCertManagerDnsAuthorization
  metadata:
    name: example-com-auth
  spec:
    projectId:
      value: my-gcp-project
    domain: example.com
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

The module reads its configuration from the `GcpCertManagerDnsAuthorizationIacInput` proto message:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| target | GcpCertManagerDnsAuthorization | Yes | The authorization resource manifest |
| provider_config | GcpProviderConfig | Yes | GCP provider configuration |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| authorization_id | string | Fully-qualified resource ID — consumed by certificates |
| authorization_name | string | Authorization name in GCP |
| domain | string | The authorized domain |
| dns_record_name | string | Fully-qualified name of the validation record |
| dns_record_type | string | Validation record type (CNAME) |
| dns_record_data | string | Validation record data — the CNAME target |

The validation-record outputs are exported via len/nil-guarded ApplyT
callbacks degrading to `""` (the optional-output export contract),
mirroring the Terraform module's index access on the computed record list.

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs.
