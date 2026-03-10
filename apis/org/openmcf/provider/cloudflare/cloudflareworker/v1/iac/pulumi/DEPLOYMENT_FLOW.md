# CloudflareWorker Deployment Flow

## Overview

The CloudflareWorker Pulumi module now automatically creates DNS records and routes for your Workers, making deployment seamless and fully automated.

## What Gets Created

When you deploy a CloudflareWorker with DNS configuration enabled, the module creates:

1. **WorkersScript** - The Worker code deployed to Cloudflare's edge
2. **DNS Record** (A record) - Automatically created with proxy enabled
3. **WorkersRoute** - Attaches the Worker to the URL pattern

## Architecture

### Deployment Order

```
1. Lookup Zone ID from domain name
   ↓
2. Create DNS A Record (proxied)
   ↓
3. Create Worker Route (depends on DNS record)
   ↓
4. Export outputs (script_id, route_urls)
```

### Key Features

- **Automatic Zone Lookup** - Provide domain name, zone ID is looked up automatically
- **DNS Record Creation** - No manual DNS record creation needed
- **Smart Defaults** - Route pattern defaults to `hostname/*` if not specified
- **Toggle Control** - Easy enable/disable via `dns.enabled` flag

## Example Configuration

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: my-worker
spec:
  account_id: "074755a78d8e8f77c119a90a125e8a06"
  
  script:
    name: my-worker
    bundle:
      bucket: my-r2-bucket
      path: workers/my-worker.js
  
  dns:
    enabled: true
    domain: example.com              # Zone name - ID looked up automatically
    hostname: api.example.com        # DNS record created automatically
    route_pattern: api.example.com/* # Optional - defaults to "hostname/*"
  
  compatibility_date: "2024-09-23"
  
  env:
    variables:
      LOG_LEVEL: info
      ENV: production
```

## DNS Record Details

When `dns.enabled: true`, the module automatically creates:

```
Type: A
Name: <hostname from dns.hostname>
Content: 100.0.0.1 (dummy IP - never reached)
Proxied: true (orange cloud enabled)
TTL: 1 (automatic)
Comment: "Managed by Planton - Routes to Cloudflare Worker"
```

### Why 100.0.0.1?

The IP address is a placeholder and is never used because:
1. DNS record is **proxied** (orange cloud)
2. Requests hit Cloudflare's edge network
3. Worker Route intercepts requests
4. Your Worker script executes
5. Origin IP is never contacted

## Deployment Scenarios

### Scenario 1: Worker with DNS Route (Most Common)

```yaml
dns:
  enabled: true
  domain: planton.live
  hostname: git-webhooks.planton.live
```

**Creates:**
- ✅ WorkersScript
- ✅ DNS A record for `git-webhooks.planton.live`
- ✅ WorkersRoute for `git-webhooks.planton.live/*`

**Result:** Worker is accessible at `https://git-webhooks.planton.live/`

### Scenario 2: Worker without Route (Testing/Development)

```yaml
dns:
  enabled: false
  # DNS fields can be kept for future use
  domain: planton.live
  hostname: git-webhooks.planton.live
```

**Creates:**
- ✅ WorkersScript only

**Result:** Worker exists but is not accessible via URL

### Scenario 3: Worker with Custom Route Pattern

```yaml
dns:
  enabled: true
  domain: planton.live
  hostname: api.planton.live
  route_pattern: api.planton.live/webhooks/*
```

**Creates:**
- ✅ WorkersScript
- ✅ DNS A record for `api.planton.live`
- ✅ WorkersRoute for `api.planton.live/webhooks/*`

**Result:** Worker only handles requests to `/webhooks/*` paths

## Prerequisites

### 1. R2 Bucket

Worker script must be uploaded to R2 before deployment:

```bash
cd backend/services/your-worker
make publish
```

This uploads the bundle to R2 and outputs the path to use in the manifest.

### 2. Environment Variables

Required for Pulumi to access R2:

```bash
export AWS_ACCESS_KEY_ID=<r2-access-key>
export AWS_SECRET_ACCESS_KEY=<r2-secret-key>
```

### 3. Cloudflare API Token

The API token needs these permissions:
- **Workers Scripts: Edit** (Account)
- **Workers Routes: Edit** (Zone)
- **DNS: Edit** (Zone)
- **Zone: Read** (Zone)

## Deployment Commands

```bash
# 1. Build the Worker bundle
cd backend/services/git-webhooks-receiver
make publish

# 2. Update manifest with the bundle path from publish output

# 3. Export R2 credentials
export AWS_ACCESS_KEY_ID=<r2-access-key-id>
export AWS_SECRET_ACCESS_KEY=<r2-secret-access-key>

# 4. Deploy with Pulumi
cd ops/organizations/planton/infra-hub/cloud-resources/app-prod/cloudflare
export CLOUDFLARE_WORKER_MODULE=~/scm/github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareworker/v1/iac/pulumi

openmcf pulumi up --manifest worker.your-worker.yaml --module-dir ${CLOUDFLARE_WORKER_MODULE}
```

## Stack Outputs

After successful deployment:

```
Outputs:
  script_id:  "abcd1234..."
  route_urls: ["git-webhooks.planton.live/*"]
```

## Troubleshooting

### DNS Record Already Exists

If a DNS record already exists for the hostname, you may get:

```
error: record already exists
```

**Solution:**
1. Remove the existing DNS record manually
2. Or set `dns.enabled: false` to skip DNS/route creation
3. Re-run deployment

### Authentication Error (10000)

```
error: Authentication error (10000)
```

**Cause:** API token missing required permissions.

**Solution:** Update API token to include:
- Workers Routes: Edit (Zone permission)
- DNS: Edit (Zone permission)

### Zone Not Found

```
error: zone not found
```

**Cause:** Domain name doesn't match any zone in your Cloudflare account.

**Solution:** Verify the domain in `dns.domain` matches an existing Cloudflare zone.

## Migration from Old Format

### Old Format (Deprecated)

```yaml
route_pattern: git-webhooks.planton.live/*
zone_id: "77c6a34cf87dd1e8b497dc895bf5ea1b"
```

### New Format

```yaml
dns:
  enabled: true
  domain: planton.live
  hostname: git-webhooks.planton.live
  route_pattern: git-webhooks.planton.live/*
```

## Benefits

1. **No Manual DNS Management** - DNS records created automatically
2. **User-Friendly** - Use domain names instead of obscure zone IDs
3. **Safe Defaults** - Sensible defaults for route patterns
4. **Flexible** - Easy to toggle DNS/routing on and off
5. **Complete Automation** - One command deployment from code to live URL

