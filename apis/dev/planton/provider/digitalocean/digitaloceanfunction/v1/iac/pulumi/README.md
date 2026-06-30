# DigitalOcean Function - Pulumi Module

This Pulumi module provisions serverless functions on DigitalOcean using **App Platform** for production-ready VPC integration, monitoring, and IaC support.

## Overview

The module implements Planton's `DigitalOceanFunction` protobuf spec, deploying functions via DigitalOcean App Platform to ensure:

- **VPC Integration**: Secure database access via private network addresses
- **Production Monitoring**: Full DigitalOcean Insights integration
- **Secret Management**: Encrypted environment variables
- **Zero-Downtime Deployments**: Automatic rollbacks and health checks

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml          # Project configuration
├── Makefile             # Build helpers
├── debug.sh             # Local debugging script
├── module/
│   ├── main.go          # Module orchestration
│   ├── locals.go        # Context initialization
│   ├── function.go      # App Platform provisioning
│   └── outputs.go       # Output constants
└── README.md            # This file
```

## Prerequisites

- **Pulumi CLI**: Version 3.0+ ([install](https://www.pulumi.com/docs/install/))
- **Go**: Version 1.21+ 
- **DigitalOcean API Token**: Set via environment variable or Pulumi config

## Quick Start

### 1. Set Up Environment

```bash
cd iac/pulumi

# Set your DigitalOcean API token
export DIGITALOCEAN_TOKEN="your-token-here"
```

### 2. Configure Stack Input

Create `stack-input.yaml`:

**Simple HTTP Function:**
```yaml
provider_config:
  credential_id: "digitalocean-prod-credential"

target:
  spec:
    function_name: api-handler
    region: nyc1
    runtime: nodejs_20
    github_source:
      repo: myorg/my-functions
      branch: main
      deploy_on_push: true
    source_directory: /functions/api-handler
    memory_mb: 256
    timeout_ms: 3000
    is_web: true
    environment_variables:
      NODE_ENV: production
    secret_environment_variables:
      DATABASE_URL: postgresql://private-db:5432/users
```

**Scheduled Background Job:**
```yaml
provider_config:
  credential_id: "digitalocean-prod-credential"

target:
  spec:
    function_name: nightly-cleanup
    region: nyc1
    runtime: python_311
    github_source:
      repo: myorg/background-jobs
      branch: main
    source_directory: /jobs/cleanup
    memory_mb: 1024
    timeout_ms: 300000  # 5 minutes
    is_web: false
    cron_schedule: "0 0 * * *"  # Midnight daily
    secret_environment_variables:
      DATABASE_URL: postgresql://cleanup-db:5432/analytics
```

### 3. Deploy

```bash
# Initialize stack
pulumi stack init production

# Preview changes
pulumi preview

# Deploy
pulumi up
```

### 4. Access Outputs

```bash
# Get function ID
pulumi stack output function_id

# Get HTTPS endpoint
pulumi stack output https_endpoint
```

## Stack Input Schema

### `provider_config` (Required)

```yaml
provider_config:
  credential_id: string  # DigitalOcean credential ID
```

### `target.spec` (Required)

```yaml
target:
  spec:
    # Essential (Required)
    function_name: string       # Unique function name (1-64 chars)
    region: string              # DigitalOcean region (e.g., nyc1, sfo3)
    runtime: string             # Runtime (nodejs_18, nodejs_20, python_39, python_310, python_311, go_120, go_121, php_82)
    
    # GitHub Source (Required)
    github_source:
      repo: string              # GitHub repo (owner/repo format)
      branch: string            # Git branch (e.g., main, production)
      deploy_on_push: bool      # Auto-deploy on push (default: true)
    
    # Source Code (Required)
    source_directory: string    # Path to function code (e.g., /functions/api)
    
    # Resources (Optional, have defaults)
    memory_mb: int              # Memory in MB (128, 256, 512, 1024, 2048; default: 256)
    timeout_ms: int             # Timeout in milliseconds (max: 300000; default: 3000)
    
    # Environment (Optional)
    environment_variables:      # Non-secret env vars
      KEY: value
    secret_environment_variables:  # Encrypted secrets
      DB_URL: postgresql://...
      API_KEY: secret-key
    
    # Function Type (Optional)
    is_web: bool                # Expose as HTTP endpoint (default: true)
    cron_schedule: string       # Cron expression for scheduled execution
    entrypoint: string          # Function entrypoint name
```

## Outputs

| Output | Description |
|--------|-------------|
| `function_id` | DigitalOcean App Platform app ID |
| `https_endpoint` | Public HTTPS URL for the function (if is_web=true) |

## Production Patterns

### HTTP API with Database

```yaml
spec:
  function_name: user-api
  region: nyc1
  runtime: nodejs_20
  github_source:
    repo: company/apis
    branch: main
  source_directory: /api/users
  memory_mb: 512
  timeout_ms: 15000
  is_web: true
  secret_environment_variables:
    DATABASE_URL: postgresql://10.116.0.5:5432/users  # VPC private IP
    JWT_SECRET: secret-key
```

### Scheduled Data Processing

```yaml
spec:
  function_name: data-processor
  region: sfo3
  runtime: python_311
  github_source:
    repo: company/jobs
    branch: production
  source_directory: /jobs/processor
  memory_mb: 2048
  timeout_ms: 300000
  is_web: false
  cron_schedule: "0 2 * * *"  # 2 AM daily
```

## Security Best Practices

### 1. Use Secret Environment Variables

**✅ Do:**
```yaml
secret_environment_variables:
  DATABASE_URL: postgresql://private-host:5432/db
  API_KEY: secret-key
```

**❌ Don't:**
```yaml
environment_variables:
  DATABASE_URL: postgresql://host:5432/db  # Visible in logs!
```

### 2. VPC Private Network Addresses

**✅ Do:**
```yaml
secret_environment_variables:
  DB_URL: postgresql://10.116.0.5:5432/db  # Private VPC IP
```

**❌ Don't:**
```yaml
DB_URL: postgresql://db-public.digitalocean.com:25060/db  # Public internet
```

### 3. Region Matching

Ensure function and database are in the **same region** for VPC access:

```yaml
spec:
  region: nyc1  # Must match database region
```

## Local Development

### Debug Script

Use the provided `debug.sh` for local testing:

```bash
./debug.sh
```

### Manual Build and Run

```bash
# Build
go build -o pulumi-main main.go

# Set token
export DIGITALOCEAN_TOKEN="your-token"

# Run
pulumi up
```

## Troubleshooting

### Issue: Function cannot connect to database

**Cause**: Using public database URL or wrong region.

**Solution**: Use VPC private network address and match regions:
```yaml
region: nyc1  # Same as database
secret_environment_variables:
  DB_URL: postgresql://10.116.0.5:5432/db  # Private IP
```

### Issue: Deployment fails with "source directory not found"

**Cause**: `source_directory` path doesn't exist in GitHub repo.

**Solution**: Verify path exists and contains `project.yml`:
```bash
# In your repo
ls -la /functions/api-handler/project.yml
```

### Issue: Function times out

**Cause**: Timeout too low for workload.

**Solution**: Increase timeout:
```yaml
timeout_ms: 30000  # Increase to 30 seconds
```

## Pulumi Commands Reference

```bash
# Initialize stack
pulumi stack init <stack-name>

# Preview changes
pulumi preview

# Deploy
pulumi up

# View outputs
pulumi stack output

# Destroy
pulumi destroy

# View stack state
pulumi stack

# Export state
pulumi stack export
```

## Further Reading

- **Component Overview**: See [../../README.md](../../README.md)
- **Comprehensive Guide**: See [../../docs/README.md](../../docs/README.md)
- **Examples**: See [../../examples.md](../../examples.md)
- **Pulumi DigitalOcean Provider**: [Registry Docs](https://www.pulumi.com/registry/packages/digitalocean/)

## Support

For module-specific issues:
1. Enable debug logging: `export PULUMI_DEBUG_COMMANDS=true`
2. Check logs: `pulumi up --logtostderr -v=9`
3. Validate API connectivity:
   ```bash
   curl -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
        https://api.digitalocean.com/v2/account
   ```

For component issues, see [../../README.md](../../README.md).

---

**TL;DR**: Use App Platform deployment (this module) for production. Set `secret_environment_variables` for sensitive data. Match function and database regions for VPC access. Use private network addresses for databases.

