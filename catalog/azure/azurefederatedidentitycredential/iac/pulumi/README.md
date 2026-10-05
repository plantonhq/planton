# AzureFederatedIdentityCredential Pulumi Module

## Overview

This Pulumi module provisions a federated identity credential using the Azure
Classic provider (`pulumi-azure`). It creates a single
`armmsi.FederatedIdentityCredential`: a keyless OIDC trust rule on a
user-assigned managed identity that lets one external workload exchange its
own token for the identity's Azure credentials.

Issuer, subject, and audience update in place; the name and the parent
identity are the credential's ARM identity, so changing either replaces it.
The provider serializes credential writes per parent identity (ARM rejects
concurrent writes on one identity), so several credentials on the same
identity deploy sequentially. Federated credentials carry no ARM tags (they
are untagged child resources of the identity).

The module derives the resource group from the parent identity's ARM ID --
the SDK requires it as its own argument even though the parent ID already
carries it, and deriving it means the two can never disagree.

## Resources Created

- `armmsi.FederatedIdentityCredential` -- the trust rule on the identity

## Inputs

The module receives an `AzureFederatedIdentityCredentialIacInput` containing:

- `target.spec.name` -- the credential's name under the parent identity
- `target.spec.user_assigned_identity` -- the parent identity's ARM ID (references resolved to a literal by the platform)
- `target.spec.issuer` -- the OIDC issuer URL the token's `iss` claim must equal
- `target.spec.subject` -- the workload identifier the token's `sub` claim must equal
- `target.spec.audience` -- optional; defaults to `api://AzureADTokenExchange`
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `federated_identity_credential_id` | Full ARM ID of the credential |
| `name` | The credential's name as deployed |
| `user_assigned_identity_id` | ARM ID of the parent identity |
| `issuer` | The trusted issuer as deployed |
| `subject` | The trusted subject as deployed |
| `audience` | The required audience as deployed |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
