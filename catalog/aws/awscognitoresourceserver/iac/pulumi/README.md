# AwsCognitoResourceServer — Pulumi Module

## Overview

This Pulumi (Go) module provisions a Cognito resource server. It is behaviorally identical to the Terraform module — same resource, same name, same outputs.

## Module Structure

```
main.go                    — entrypoint (loads IaC input, runs module.Resources)
module/main.go             — orchestration: provider → resource server
module/resource_server.go  — cognito.ResourceServer + computed scope identifiers
module/locals.go           — naming basis (the resource is not taggable)
module/outputs.go          — output name constants
```

## Deploy

```bash
# from iac/pulumi/ with an iac-input.yaml present
make preview
make up
make destroy
```

## Notes

- Scope identifiers are computed from the spec ("{identifier}/{scope_name}") rather than read back from the provider, so export order is deterministic on both engines.
