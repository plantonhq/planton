# Auth0Connection Pulumi Module

This directory contains the Pulumi implementation for the Auth0Connection kind.

## Overview

The Auth0Connection Pulumi module creates and manages Auth0 identity connections, including:
- Database connections (Auth0 hosted)
- Social identity providers (Google, Facebook, GitHub, etc.)
- Enterprise SSO (SAML, OIDC, Azure AD)

## Prerequisites

1. **Pulumi CLI**: Install from https://www.pulumi.com/docs/install/
2. **Go 1.21+**: Required for building the module
3. **Auth0 Account**: With a Machine-to-Machine application configured

## Environment Variables

The module reads its `Auth0ConnectionIacInput` (the manifest under `target`, plus an optional Auth0 `provider_config`) from the Pulumi config key `planton:iac-input`, the `IAC_INPUT_YAML` environment variable (YAML content), or `IAC_INPUT_YAML_FILE` (a path to that YAML):

```bash
export IAC_INPUT_YAML_FILE=/path/to/iac-input.yaml
```

```yaml
# iac-input.yaml
target:
  apiVersion: auth0.planton.dev/v1alpha1
  kind: Auth0Connection
  metadata:
    name: ...
  spec:
    ...
```

`planton pulumi` builds this file from a manifest for you (see Usage).

Alternatively, Auth0 credentials can be provided via environment variables:
- `AUTH0_DOMAIN`: Your Auth0 tenant domain
- `AUTH0_CLIENT_ID`: M2M application client ID
- `AUTH0_CLIENT_SECRET`: M2M application client secret

## Usage

Run the module from this directory with the planton CLI (the directory holds `Pulumi.yaml`, so the CLI runs this module). Pass Auth0 credentials with `-p <provider-config.yaml>`, or leave it off to use the environment variables above:

```bash
# Initialize stack
planton pulumi init --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>

# Preview changes
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>

# Apply changes
planton pulumi update --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>

# Destroy resources
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
```

## Module Structure

```
pulumi/
├── main.go           # Entry point, loads IaC input and calls module
├── Pulumi.yaml       # Pulumi project configuration
├── BUILD.bazel       # Bazel build target
├── README.md         # This file
└── module/
    ├── main.go       # Resources orchestration
    ├── locals.go     # Local value initialization
    ├── outputs.go    # Output exports
    └── connection.go # Auth0 connection resource creation
```

## Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `id` | Auth0 connection ID |
| `name` | Connection name |
| `strategy` | Connection strategy type |
| `is_enabled` | Whether the connection has enabled clients |
| `enabled_client_ids` | List of enabled client IDs |
| `realms` | Connection realms |

## Troubleshooting

### "failed to create Auth0 provider"

Ensure Auth0 credentials are correctly configured either via:
- Provider config in IaC input
- Environment variables

### "connection already exists"

Auth0 connection names must be unique within a tenant. Either:
- Delete the existing connection
- Use a different name in the manifest

### Plugin Not Found

Run `pulumi plugin install resource auth0` to install the Auth0 provider plugin.

## Related Documentation

- [Auth0Connection spec.proto](../../v1alpha1/spec.proto)

