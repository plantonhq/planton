# Pulumi Module — AwsServerlessElasticache

This directory contains the Pulumi IaC module for provisioning AWS ElastiCache
Serverless caches.

## Structure

- `main.go` — Pulumi program entrypoint. Loads stack input and calls the module.
- `module/` — Reusable module containing the resource logic.
  - `main.go` — Orchestrates provider creation and resource creation.
  - `locals.go` — Pre-computes tags and references from stack input.
  - `outputs.go` — Defines output key constants.
  - `serverless_cache.go` — Creates the ElastiCache Serverless cache resource.

## Local Development

```bash
# Build
cd module && go build ./...

# Run with Pulumi
pulumi up --stack dev
```

## Debug

Pulumi encrypts every secret value in the stack under this passphrase, so use a real one and keep it, even for a scratch stack:

```bash
export PULUMI_CONFIG_PASSPHRASE="$(openssl rand -base64 32)"   # keep it: the stack's secrets open with nothing else
pulumi login --local
pulumi stack init dev
pulumi config set-all --path < ../../e2e/manifest.yaml
pulumi up
```
