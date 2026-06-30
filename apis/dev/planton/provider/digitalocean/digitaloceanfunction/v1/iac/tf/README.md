# DigitalOcean Function - Terraform Module Status

## Why Pulumi-Only?

**This component intentionally supports only Pulumi deployment, not Terraform.** Here's why:

### The App Platform Approach

DigitalOcean Functions are deployed via **DigitalOcean App Platform** for production readiness:
- ✅ VPC integration for secure database access
- ✅ DigitalOcean Insights monitoring
- ✅ Encrypted secret management
- ✅ One-click rollbacks and zero-downtime deployments

### Terraform Limitation

While Terraform's `digitalocean_app` resource technically supports functions, the implementation is **extremely verbose and error-prone** due to how App Platform specs must be defined in HCL.

**Example complexity (Terraform):**
```hcl
resource "digitalocean_app" "function" {
  spec {
    name   = "my-function"
    region = "nyc1"
    
    function {
      name = "api-handler"
      
      github {
        repo           = "myorg/my-functions"
        branch         = "main"
        deploy_on_push = true
      }
      
      source_dir = "/functions/api"
      
      env {
        key   = "DB_URL"
        value = "postgresql://..."
        type  = "SECRET"
      }
      
      # Limits must be separately defined
      routes {
        path = "/"
      }
    }
  }
}
```

**Issues with Terraform approach:**
1. **No App-Spec Abstraction**: Must manually construct nested HCL blocks
2. **Type Mismatches**: Env vars, routes, and limits have complex type structures
3. **Error Messages**: Terraform provider errors for App Platform are cryptic
4. **State Drift**: App Platform auto-updates cause state inconsistencies

### Pulumi Advantage

Pulumi's type-safe Go SDK provides **native App Platform support** with:
- ✅ Compile-time type checking
- ✅ Programmatic app-spec generation
- ✅ Better error messages
- ✅ Cleaner state management

**Example simplicity (Pulumi):**
```go
app, _ := digitalocean.NewApp(ctx, "function", &digitalocean.AppArgs{
    Spec: digitalocean.AppSpecArgs{
        Name:   pulumi.String("my-function"),
        Region: pulumi.String("nyc1"),
        Functions: digitalocean.AppSpecFunctionArray{
            &digitalocean.AppSpecFunctionArgs{
                Name: pulumi.String("api-handler"),
                Github: &digitalocean.AppSpecFunctionGithubArgs{
                    Repo:         pulumi.String("myorg/my-functions"),
                    Branch:       pulumi.String("main"),
                    DeployOnPush: pulumi.Bool(true),
                },
                SourceDir: pulumi.String("/functions/api"),
                Envs: digitalocean.AppSpecFunctionEnvArray{
                    &digitalocean.AppSpecFunctionEnvArgs{
                        Key:   pulumi.String("DB_URL"),
                        Value: pulumi.String("postgresql://..."),
                        Type:  pulumi.String("SECRET"),
                    },
                },
            },
        },
    },
})
```

## Alternative: Use Pulumi

If you need Infrastructure-as-Code for DigitalOcean Functions, **use the Pulumi module**:

```bash
cd ../pulumi
pulumi stack init production
pulumi up
```

See [../pulumi/README.md](../pulumi/README.md) for usage instructions.

## Alternative: Standalone Functions (Development Only)

If you're doing local development or prototyping, you can use **Standalone Functions** via the `doctl` CLI:

```bash
# NOT recommended for production (no VPC, no monitoring, no IaC)
doctl serverless init --language nodejs my-function
cd my-function
doctl serverless deploy .
```

**Warning**: Standalone Functions lack:
- ❌ VPC networking (cannot securely connect to databases)
- ❌ Monitoring (no DigitalOcean Insights integration)
- ❌ IaC support (no Terraform or Pulumi resources exist)

See [../../docs/README.md](../../docs/README.md) for the comprehensive comparison of deployment methods.

## Summary

- **Production**: Use Pulumi module (App Platform deployment)
- **Development**: Use `doctl serverless` for local testing only
- **Terraform**: Not supported due to App Platform complexity in HCL

For questions or contributions, see the main [README.md](../../README.md).

