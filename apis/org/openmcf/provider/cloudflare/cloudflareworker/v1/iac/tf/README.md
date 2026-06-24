# Terraform Module: Cloudflare Worker

This directory contains the Terraform module for deploying Cloudflare Workers.

## Overview

The Terraform module provisions Cloudflare Workers with:
1. R2 bundle fetching (via AWS S3 provider)
2. Worker script creation with configuration
3. KV namespace bindings
4. Optional DNS records and routes for custom domains
5. Environment variables and secrets

## Module Structure

```
iac/tf/
├── README.md        # This file - deployment guide
├── variables.tf     # Input variables
├── provider.tf      # Cloudflare and AWS (R2) provider configuration
├── locals.tf        # Local variables and computed values
├── main.tf          # Resource definitions
└── outputs.tf       # Output values
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

Create in: Dashboard → R2 → Manage R2 API Tokens

### 3. Environment Variables

```bash
# Cloudflare API token
export CLOUDFLARE_API_TOKEN="your-cloudflare-api-token"

# R2 credentials for bundle fetching
export AWS_ACCESS_KEY_ID="your-r2-access-key-id"
export AWS_SECRET_ACCESS_KEY="your-r2-secret-access-key"
```

### 4. Terraform CLI

```bash
# macOS
brew install terraform

# Linux
wget https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip
unzip terraform_1.6.0_linux_amd64.zip
sudo mv terraform /usr/local/bin/

# Verify
terraform version
```

## Usage

### Step 1: Build and Upload Bundle

```bash
# Build worker
npx wrangler build

# Upload to R2
aws s3 cp dist/worker.js \
  s3://my-workers-bucket/builds/worker-v1.0.0.js \
  --endpoint-url https://<account-id>.r2.cloudflarestorage.com
```

### Step 2: Create Terraform Configuration

```hcl
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

module "cloudflare_worker" {
  source = "./iac/tf"

  metadata = {
    name = "api-gateway"
  }

  spec = {
    account_id = "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
    
    script = {
      name = "api-gateway-prod"
      bundle = {
        bucket = "workers-prod"
        path   = "builds/api-v1.0.0.js"
      }
    }
    
    dns = {
      enabled       = true
      zone_id       = "zone123abc"
      hostname      = "api.example.com"
      route_pattern = "api.example.com/*"
    }
    
    compatibility_date = "2025-01-15"
    
    env = {
      variables = {
        LOG_LEVEL = "info"
      }
      secrets = {
        API_KEY = "secret-value"
      }
    }
  }
}

output "script_id" {
  value = module.cloudflare_worker.script_id
}

output "route_urls" {
  value = module.cloudflare_worker.route_urls
}
```

### Step 3: Initialize Terraform

```bash
terraform init
```

### Step 4: Plan and Apply

```bash
# Preview changes
terraform plan

# Apply
terraform apply
```

**Expected Output**:

```
Plan: 3 to add, 0 to change, 0 to destroy.

  + cloudflare_workers_script.main
  + cloudflare_dns_record.worker_dns[0]
  + cloudflare_workers_route.main[0]

Apply complete! Resources: 3 added, 0 changed, 0 destroyed.

Outputs:
script_id = "worker-script-id"
route_urls = ["https://api.example.com"]
```

### Step 5: Test Worker

```bash
curl https://api.example.com
```

## Input Variables

### Required

#### `metadata` (object)
```hcl
metadata = {
  name = "api-gateway"
}
```

#### `spec` (object)

**Required fields**:

- `account_id` (string): Cloudflare account ID (32 hex chars)
- `script` (object):
  - `name` (string): Worker script name
  - `bundle` (object):
    - `bucket` (string): R2 bucket name
    - `path` (string): Path to bundle in R2

**Optional fields**:

- `kv_bindings` (list): KV namespace bindings
- `dns` (object): Custom domain configuration
- `compatibility_date` (string): Runtime version (YYYY-MM-DD)
- `usage_model` (number): 0=BUNDLED, 1=UNBOUND
- `env` (object): Variables and secrets

## Outputs

| Output | Description |
|--------|-------------|
| `script_id` | Worker script ID |
| `script_name` | Worker script name |
| `route_urls` | List of accessible URLs |
| `route_pattern` | Route pattern |

## Examples

### Minimal Worker

```hcl
module "hello_worker" {
  source = "./iac/tf"

  metadata = { name = "hello" }
  
  spec = {
    account_id = "a1b2c3d4...o5p6"
    script = {
      name = "hello-worker"
      bundle = {
        bucket = "workers-prod"
        path   = "hello-v1.0.0.js"
      }
    }
    compatibility_date = "2025-01-15"
  }
}
```

### With Custom Domain

```hcl
module "api_worker" {
  source = "./iac/tf"

  metadata = { name = "api" }
  
  spec = {
    account_id = "a1b2c3d4...o5p6"
    script = {
      name = "api-prod"
      bundle = {
        bucket = "workers-prod"
        path   = "api-v1.0.0.js"
      }
    }
    dns = {
      enabled       = true
      zone_id       = "zone123"
      hostname      = "api.example.com"
      route_pattern = "api.example.com/*"
    }
    compatibility_date = "2025-01-15"
  }
}
```

### With KV Bindings

```hcl
module "gateway_worker" {
  source = "./iac/tf"

  metadata = { name = "gateway" }
  
  spec = {
    account_id = "a1b2c3d4...o5p6"
    script = {
      name = "gateway"
      bundle = {
        bucket = "workers-prod"
        path   = "gateway-v2.0.0.js"
      }
    }
    kv_bindings = [
      {
        name       = "CACHE_KV"
        field_path = "cache-namespace-id"
      }
    ]
    env = {
      variables = {
        LOG_LEVEL = "info"
      }
    }
    compatibility_date = "2025-01-15"
  }
}
```

## Updating the Worker

### Update Bundle Version

1. Build new version
2. Upload to R2 with new path
3. Update Terraform:
   ```hcl
   bundle = {
     path = "builds/worker-v1.1.0.js"  # New version
   }
   ```
4. Apply:
   ```bash
   terraform apply
   ```

### Add Environment Variable

```hcl
env = {
  variables = {
    NEW_FEATURE = "enabled"
  }
}
```

### Enable Custom Domain

```hcl
dns = {
  enabled  = true
  zone_id  = "zone123"
  hostname = "api.example.com"
}
```

## Destroying the Worker

```bash
terraform destroy
```

**Warning**: Deletes Worker script, routes, and DNS records.

## State Management

### Remote State (Recommended)

```hcl
terraform {
  backend "s3" {
    bucket         = "terraform-state"
    key            = "cloudflare/worker/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-lock"
  }
}
```

### Terraform Cloud

```hcl
terraform {
  cloud {
    organization = "my-org"
    workspaces {
      name = "cloudflare-worker"
    }
  }
}
```

## Troubleshooting

### Issue: "Failed to fetch bundle from R2"

**Solution**:
1. Verify R2 credentials
2. Check bundle exists:
   ```bash
   aws s3 ls s3://bucket/path --endpoint-url https://<account-id>.r2.cloudflarestorage.com
   ```

### Issue: "Route already exists"

**Solution**: Another Worker uses same route. Use different hostname or pattern.

### Issue: "Authentication error"

**Solution**: Verify `CLOUDFLARE_API_TOKEN` has Workers edit permissions.

## Best Practices

1. **Version bundles semantically**: `v1.0.0`, `v1.0.1`
2. **Use remote state**: S3 or Terraform Cloud
3. **Store secrets securely**: GitHub Secrets, AWS Secrets Manager
4. **Separate environments**: Use workspaces or separate state files
5. **Pin compatibility dates**: Avoid automatic runtime changes
6. **Test locally**: Use `wrangler dev` before deploying

## Limitations

### Secrets Upload

Terraform doesn't support uploading Worker secrets via the Cloudflare API. Secrets in `env.secrets` are treated as plain text bindings.

**Workaround**: Upload secrets manually via:
- Cloudflare Dashboard
- Wrangler CLI: `wrangler secret put API_KEY`
- Cloudflare API directly

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy Worker

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: hashicorp/setup-terraform@v2
      
      - name: Upload bundle
        run: |
          aws s3 cp dist/worker.js s3://workers-prod/builds/${{ github.sha }}.js \
            --endpoint-url https://${{ secrets.ACCOUNT_ID }}.r2.cloudflarestorage.com
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.R2_ACCESS_KEY }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.R2_SECRET_KEY }}
      
      - name: Terraform Apply
        run: terraform apply -auto-approve
        working-directory: iac/tf
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          AWS_ACCESS_KEY_ID: ${{ secrets.R2_ACCESS_KEY }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.R2_SECRET_KEY }}
```

## Additional Resources

- [Terraform Cloudflare Provider](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
- [Cloudflare Workers API](https://developers.cloudflare.com/api/operations/worker-script-upload-worker-module)
- [Component README](../../README.md)
- [Examples](../../examples.md)

---

**Ready to deploy?** Run `terraform init && terraform apply` to get started!

