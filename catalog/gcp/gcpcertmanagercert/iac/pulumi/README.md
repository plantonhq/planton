# GcpCertManagerCert Pulumi Module

This Pulumi module creates one Certificate Manager certificate —
Google-managed (auto-renewed) or self-managed (uploaded PEM) — and enables
the Certificate Manager API on the target project.

## Usage

This module is typically invoked by the Planton CLI, but can also be used directly.

### With Planton CLI

```bash
planton pulumi up --manifest certificate.yaml
```

### Standalone Usage

1. Set the IaC input as an environment variable. The module reads a `GcpCertManagerCertIacInput` from `IAC_INPUT_YAML` (YAML content), `IAC_INPUT_YAML_FILE` (a path to that YAML) or the Pulumi config key `planton:iac-input`; the manifest goes under `target`:

```bash
export IAC_INPUT_YAML=$(cat <<EOF
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpCertManagerCert
  metadata:
    name: web-cert
  spec:
    projectId:
      value: my-gcp-project
    managed:
      domains:
        - app.example.com
      dnsAuthorizations:
        - value: projects/my-gcp-project/locations/global/dnsAuthorizations/app-auth
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

The module reads its configuration from the `GcpCertManagerCertIacInput` proto message:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| target | GcpCertManagerCert | Yes | The certificate resource manifest |
| provider_config | GcpProviderConfig | Yes | GCP provider configuration |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| certificate_id | string | Fully-qualified resource ID |
| certificate_name | string | Certificate name — consumed by target HTTPS proxies |
| san_dnsnames | string[] | SANs in the issued certificate |
| location | string | The Certificate Manager location |
| managed_state | string | PROVISIONING/FAILED/ACTIVE for managed; empty for self-managed (exported via a nil-guarded ApplyT, mirroring the Terraform module's try()) |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs.

## Implementation Notes

- DNS authorizations are separate `GcpCertManagerDnsAuthorization`
  resources; this module never creates them or their DNS records.
- The self-managed arm's `pem_private_key` is secret material — masked in
  state and outputs.
- A managed certificate stays PROVISIONING until its validation records
  resolve publicly — creation succeeds regardless.
