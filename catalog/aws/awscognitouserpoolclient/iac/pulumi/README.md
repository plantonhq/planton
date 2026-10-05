# AwsCognitoUserPoolClient — Pulumi Module

## Overview

This Pulumi (Go) module provisions a Cognito User Pool app client. It is behaviorally identical to the Terraform module — same resource, same name, same outputs.

## Module Structure

```
main.go           — entrypoint (loads IaC input, runs module.Resources)
module/main.go    — orchestration: provider → client
module/client.go  — cognito.UserPoolClient with the full spec surface
module/locals.go  — naming basis (the client resource is not taggable)
module/outputs.go — output name constants
```

## Deploy

```bash
# from iac/pulumi/ with an iac-input.yaml present
make preview
make up
make destroy
```

## Notes

- The client's cloud name is `metadata.name` — the cross-engine naming basis.
- `client_secret` is only minted when `generate_secret` is true; treat the output as a credential.
- Supported identity providers resolve from references (an `AwsCognitoIdentityProvider`'s `provider_name` output) or literals ("COGNITO", "Google") — references also give the deployment graph the right ordering.
