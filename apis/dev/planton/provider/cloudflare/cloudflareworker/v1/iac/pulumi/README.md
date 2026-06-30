# Pulumi Module: Cloudflare Worker

This directory contains the Pulumi module for deploying Cloudflare Workers using Planton.

## Overview

The Pulumi module provisions Cloudflare Workers with:
1. **R2 Bundle Fetching**: Retrieves Worker script from R2 bucket
2. **Worker Script Creation**: Deploys script with configuration
3. **KV Bindings**: Connects Worker to KV namespaces
4. **DNS Records**: Creates custom domain configuration
5. **Route Attachment**: Maps Worker to URL patterns
6. **Environment Configuration**: Sets variables and secrets

## Directory Structure

```
iac/pulumi/
├── README.md              # This file - deployment guide
├── overview.md            # Architecture and design decisions
├── main.go                # Pulumi entrypoint
├── Pulumi.yaml            # Pulumi project configuration
├── Makefile               # Build and deployment helpers
├── debug.sh               # Debug script for local testing
└── module/
    ├── main.go            # Module entrypoint (Resources function)
    ├── locals.go          # Local variables and helpers
    ├── worker_script.go   # Worker script provisioning
    ├── route.go           # Route and DNS configuration
    ├── secrets.go         # Secrets management
    └── outputs.go         # Output constant definitions
```

## Prerequisites

### 1. Cloudflare Account

- Active Cloudflare account
- Workers enabled
- API token with permissions:
  - Workers Scripts: Edit
  - Workers Routes: Edit
  - DNS: Edit
  - Account Settings: Read

### 2. R2 Credentials

Worker bundles are stored in R2. You need:
- R2 Access Key ID
- R2 Secret Access Key

Create R2 API token in: Dashboard → R2 → Manage R2 API Tokens

### 3. Environment Variables

```bash
# Cloudflare API token
export CLOUDFLARE_API_TOKEN="your-cloudflare-api-token"

# R2 credentials for bundle fetching
export AWS_ACCESS_KEY_ID="your-r2-access-key-id"
export AWS_SECRET_ACCESS_KEY="your-r2-secret-access-key"
```

### 4. Pulumi CLI

```bash
# macOS
brew install pulumi

# Linux
curl -fsSL https://get.pulumi.com | sh

# Verify
pulumi version
```

### 5. Go SDK

- Go 1.21 or later
- Pulumi Cloudflare and AWS provider SDKs (auto-installed)

## Deployment

### Step 1: Build and Upload Worker Bundle

```bash
# Build worker
npx wrangler build

# Upload to R2
aws s3 cp dist/worker.js \
  s3://my-workers-bucket/builds/worker-v1.0.0.js \
  --endpoint-url https://<account-id>.r2.cloudflarestorage.com
```

### Step 2: Create Stack Input

```yaml
# cloudflare-worker-stack-input.yaml
target:
  metadata:
    name: api-gateway
  spec:
    account_id: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
    script:
      name: api-gateway-prod
      bundle:
        bucket: my-workers-bucket
        path: builds/worker-v1.0.0.js
    dns:
      enabled: true
      zone_id: "zone123abc"
      hostname: api.example.com
      route_pattern: api.example.com/*
    compatibility_date: "2025-01-15"
    env:
      variables:
        LOG_LEVEL: "info"
      secrets:
        API_KEY: "secret-key-value"

provider_config:
  r2:
    access_key_id: "${AWS_ACCESS_KEY_ID}"
    secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
```

### Step 3: Initialize Pulumi Stack

```bash
cd iac/pulumi

# Initialize stack
pulumi stack init dev

# Optional: Configure Pulumi secrets
pulumi config set cloudflare:apiToken --secret "${CLOUDFLARE_API_TOKEN}"
```

### Step 4: Deploy

```bash
# Preview changes
pulumi preview --stack dev

# Apply changes
pulumi up --stack dev
```

**Expected Output**:

```
Updating (dev)

     Type                              Name                Status
 +   pulumi:pulumi:Stack               worker-dev          created
 +   ├─ cloudflare:index:WorkersScript      workers-script     created
 +   ├─ cloudflare:index:Record             worker-dns         created
 +   └─ cloudflare:index:WorkerRoute        worker-route       created

Outputs:
    script_id: "worker-script-id"
    route_urls: ["https://api.example.com"]

Resources:
    + 3 created

Duration: 8s
```

### Step 5: Test Worker

```bash
curl https://api.example.com
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `script_id` | Cloudflare Worker script ID |
| `route_urls` | List of URLs where Worker is accessible |

Access outputs:

```bash
pulumi stack output
pulumi stack output script_id
```

## Updating the Worker

### Update Bundle Version

1. Build new version
2. Upload to R2 with new path
3. Update stack input:
   ```yaml
   bundle:
     path: builds/worker-v1.1.0.js  # New version
   ```
4. Deploy:
   ```bash
   pulumi up --stack dev
   ```

### Add KV Binding

```yaml
kv_bindings:
  - name: MY_KV
    field_path: "namespace-id-here"
```

### Add Environment Variable

```yaml
env:
  variables:
    NEW_FEATURE: "enabled"
```

## Destroying the Worker

```bash
pulumi destroy --stack dev
```

**Warning**: This deletes the Worker script, routes, and DNS records.

## Debugging

### Enable Debug Mode

```bash
./debug.sh
```

or

```bash
pulumi up --stack dev --logtostderr -v=9
```

### View Worker Logs

Check Cloudflare Dashboard → Workers & Pages → your-worker → Logs

### Common Issues

**Issue**: `Error: failed to fetch bundle from R2`

**Solution**:
1. Verify R2 credentials are correct
2. Check bundle exists:
   ```bash
   aws s3 ls s3://bucket/path \
     --endpoint-url https://<account-id>.r2.cloudflarestorage.com
   ```

---

**Issue**: `Error: authentication error`

**Solution**: Verify `CLOUDFLARE_API_TOKEN` has Workers edit permissions

---

**Issue**: `Error: route already exists`

**Solution**: Another Worker is using the same route. Use different hostname or more specific pattern.

## Multi-Environment

Use separate stacks:

```bash
# Dev
pulumi stack init dev
pulumi up --stack dev

# Staging
pulumi stack init staging
pulumi up --stack staging

# Production
pulumi stack init prod
pulumi up --stack prod
```

## State Management

Default: Pulumi Cloud (free for individuals)

Self-hosted options:

```bash
# S3
pulumi login s3://my-state-bucket

# Local (not recommended for teams)
pulumi login file://~/.pulumi

# Azure Blob
pulumi login azblob://container

# GCS
pulumi login gs://bucket
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy Cloudflare Worker

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Upload bundle to R2
        run: |
          aws s3 cp dist/worker.js \
            s3://workers-prod/builds/worker-${{ github.sha }}.js \
            --endpoint-url https://${{ secrets.CLOUDFLARE_ACCOUNT_ID }}.r2.cloudflarestorage.com
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.R2_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.R2_SECRET_ACCESS_KEY }}
      
      - uses: pulumi/actions@v4
        with:
          command: up
          stack-name: prod
          work-dir: iac/pulumi
        env:
          PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          AWS_ACCESS_KEY_ID: ${{ secrets.R2_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.R2_SECRET_ACCESS_KEY }}
```

## Best Practices

1. **Version bundles semantically**: Use `v1.0.0`, `v1.0.1` in R2 paths
2. **Pin compatibility dates**: Avoid automatic runtime updates
3. **Test locally first**: Use `wrangler dev` before deploying
4. **Use environment-specific stacks**: Separate dev/staging/prod
5. **Store secrets securely**: Use Pulumi secrets or external vaults
6. **Monitor Worker logs**: Enable observability in Cloudflare Dashboard

## Additional Resources

- [Pulumi Cloudflare Provider](https://www.pulumi.com/registry/packages/cloudflare/)
- [Cloudflare Workers API](https://developers.cloudflare.com/api/operations/worker-script-upload-worker-module)
- [overview.md](./overview.md) - Architecture and design decisions
- [Component README](../../README.md) - User-facing documentation

---

**Ready to deploy?** Run `pulumi up --stack dev` to get started!

