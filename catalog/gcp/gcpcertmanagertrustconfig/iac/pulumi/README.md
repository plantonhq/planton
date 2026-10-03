# GcpCertManagerTrustConfig Pulumi Module

This Pulumi module creates one Certificate Manager trust config and enables the Certificate Manager API on the target project.

## Usage

This module is typically invoked by the Planton CLI, but can also be used directly.

### With Planton CLI

```bash
planton pulumi up --manifest trust-config.yaml
```

### Standalone Usage

1. Set the IaC input as an environment variable:

```bash
export PLANTON_CATALOG_OBJECT_MANIFEST=$(cat <<EOT
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerTrustConfig
metadata:
  name: partner-mtls-trust
spec:
  projectId:
    value: my-gcp-project
  trustStores:
    - trustAnchors:
        - |
          -----BEGIN CERTIFICATE-----
          <partner-root-ca-pem-body>
          -----END CERTIFICATE-----
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

The module reads its configuration from the `GcpCertManagerTrustConfigIacInput` proto message:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| target | GcpCertManagerTrustConfig | Yes | The trust config resource manifest |
| provider_config | GcpProviderConfig | Yes | GCP provider configuration |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| trust_config_id | string | Full resource name — what a server TLS policy and a backend authentication config take |
| trust_config_name | string | Trust config name in GCP |
| location | string | The Certificate Manager location (`global` unless set) |

## Behavior Notes

- `location` falls back to `global` (the provider requires one), identically on both engines.
- Each PEM becomes its own one-field block in the provider; trust anchors and intermediate CAs are wrapped as Pulumi secrets, matching the provider's sensitive marking.
- The trust-store and allowlisted-certificate arguments are sent only when non-empty.
- Updates are in place; rotating a CA never replaces the config.

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs.
