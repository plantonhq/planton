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

```bash
export PULUMI_CONFIG_PASSPHRASE=""
pulumi login --local
pulumi stack init dev
pulumi config set-all --path < ../../hack/manifest.yaml
pulumi up
```
