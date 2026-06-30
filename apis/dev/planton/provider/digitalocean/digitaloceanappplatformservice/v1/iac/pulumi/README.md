# DigitalOcean App Platform Service - Pulumi Module

## Overview

This Pulumi module deploys containerized applications and services on DigitalOcean App Platform using Planton's declarative API. The module transforms the `DigitalOceanAppPlatformServiceSpec` protobuf definition into DigitalOcean App Platform resources, handling web services, workers, and jobs with support for git-based or container image deployments.

## Features

- **Multiple Service Types**: Web services, workers, and pre-deployment jobs
- **Flexible Sources**: Git repositories or container images from DOCR
- **Autoscaling**: CPU-based horizontal autoscaling for web services
- **Custom Domains**: Automatic SSL provisioning and DNS integration
- **Environment Management**: Secure environment variable configuration
- **State Management**: Pulumi tracks all resources and enables preview/update workflows

## Prerequisites

- **Pulumi CLI**: Version 3.x or higher ([installation guide](https://www.pulumi.com/docs/get-started/install/))
- **Go**: Version 1.21+ (for module development)
- **DigitalOcean Account**: With API access
- **DigitalOcean API Token**: For provider authentication

## Installation

This module is part of the Planton monorepo. To use it:

```bash
git clone https://github.com/plantonhq/planton.git
cd planton/apis/dev/planton/provider/digitalocean/digitaloceanappplatformservice/v1/iac/pulumi
```

## Usage

### Basic Example

```go
package main

import (
    digitaloceanappplatformservicev1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanappplatformservice/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanappplatformservice/v1/iac/pulumi/module"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceStackInput{
            Target: &digitaloceanappplatformservicev1.DigitalOceanAppPlatformService{
                Metadata: &shared.CloudResourceMetadata{
                    Name: "my-api",
                },
                Spec: &digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceSpec{
                    ServiceName:      "my-api",
                    Region:           digitaloceanappplatformservicev1.DigitalOceanRegion_nyc1,
                    ServiceType:      digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceType_web_service,
                    GitSource: &digitaloceanappplatformservicev1.DigitalOceanAppPlatformGitSource{
                        RepoUrl: "https://github.com/myorg/my-api.git",
                        Branch:  "main",
                    },
                    InstanceSizeSlug: digitaloceanappplatformservicev1.DigitalOceanAppPlatformInstanceSize_basic_xxs,
                    InstanceCount:    1,
                },
            },
            ProviderConfig: &digitaloceanv1.DigitalOceanProviderConfig{
                Token: os.Getenv("DIGITALOCEAN_TOKEN"),
            },
        }

        return module.Resources(ctx, stackInput)
    })
}
```

### Running Locally

1. **Set DigitalOcean credentials**:
   ```bash
   export DIGITALOCEAN_TOKEN="your-do-api-token"
   ```

2. **Initialize Pulumi stack**:
   ```bash
   pulumi stack init dev
   ```

3. **Preview changes**:
   ```bash
   pulumi preview
   ```

4. **Deploy**:
   ```bash
   pulumi up
   ```

5. **View outputs**:
   ```bash
   pulumi stack output
   ```

## Module Architecture

### Input: DigitalOceanAppPlatformServiceStackInput

The module accepts a stack input containing:

- **target**: The `DigitalOceanAppPlatformService` resource specification
- **provider_config**: DigitalOcean API credentials

### Processing Flow

1. **Initialize locals** - Extract metadata, labels, and configuration
2. **Create provider** - Set up DigitalOcean provider with credentials
3. **Build service spec** - Construct App Platform service/worker/job based on `service_type`
4. **Configure source** - Set up git repository or container image source
5. **Apply environment variables** - Configure runtime environment
6. **Enable autoscaling** (if specified) - Configure CPU-based autoscaling
7. **Add custom domain** (if specified) - Configure DNS and SSL
8. **Create app** - Deploy to DigitalOcean App Platform
9. **Export outputs** - Return app ID, URL, and metadata

### Outputs

The module exports:

- `app_id`: DigitalOcean App Platform application ID
- `live_url`: Public URL of the deployed application

## Service Type Handling

### Web Service
- Created as `digitalocean.AppSpecService`
- Supports HTTP traffic and autoscaling
- Exports live URL for public access

### Worker
- Created as `digitalocean.AppSpecWorker`
- No HTTP traffic, runs background processes
- No autoscaling (fixed instance count)

### Job
- Created as `digitalocean.AppSpecJob`
- Runs as `PRE_DEPLOY` job (executes before deployments)
- Single instance, no scaling

## Source Configuration

### Git Source
Deploys from a Git repository:
```go
GitSource: &digitaloceanappplatformservicev1.DigitalOceanAppPlatformGitSource{
    RepoUrl:      "https://github.com/myorg/myapp.git",
    Branch:       "main",
    BuildCommand: "npm run build",  // Optional
    RunCommand:   "npm start",      // Optional
}
```

### Image Source
Deploys from a container registry:
```go
ImageSource: &digitaloceanappplatformservicev1.DigitalOceanAppPlatformRegistrySource{
    Registry:   "registry.digitalocean.com/myorg",
    Repository: "myapp",
    Tag:        "v1.0.0",
}
```

## Debugging

### Enable Debug Script

Use the provided debug script for verbose output:

```bash
./debug.sh
```

This script:
- Sets `PULUMI_DEBUG_COMMANDS=true`
- Enables detailed logging
- Shows resource creation steps

### Common Issues

**Issue**: "Failed to create DigitalOcean provider"
- **Solution**: Verify `DIGITALOCEAN_TOKEN` environment variable is set

**Issue**: "Invalid service_type"
- **Solution**: Ensure `service_type` is one of: `web_service`, `worker`, or `job`

**Issue**: "Autoscaling failed"
- **Solution**: Autoscaling is only supported for `web_service` type with Professional tier instances

**Issue**: "Build failed" (git sources)
- **Solution**: Check repository is accessible, branch exists, and build command is correct

## Best Practices

### Credential Management
- **Never hardcode tokens** in source code
- Use environment variables or secret management tools
- Rotate API tokens regularly

### State Management
- Use remote state backend (Pulumi Service, S3, or Azure Blob)
- Enable state encryption
- Backup state files regularly

### Development Workflow
1. Make changes to spec
2. Run `pulumi preview` to see planned changes
3. Review resource diff carefully
4. Run `pulumi up` to apply changes
5. Verify deployment in DigitalOcean console

### Production Deployments
- Use separate Pulumi stacks for environments (dev/staging/prod)
- Pin module versions in production
- Enable stack protection (`pulumi stack tag set protect true`)
- Use CI/CD for automated deployments

## Advanced Examples

For comprehensive examples including autoscaling, custom domains, and multi-environment setups, see:
- [Pulumi examples.md](./examples.md)
- [API-level examples](../../examples.md)

## References

- [Pulumi DigitalOcean Provider](https://www.pulumi.com/registry/packages/digitalocean/)
- [DigitalOcean App Platform API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Apps)
- [Module overview](./overview.md) - Architecture and design decisions

## Support

- **Planton**: [github.com/plantonhq/planton](https://github.com/plantonhq/planton)
- **Pulumi Docs**: [pulumi.com/docs](https://www.pulumi.com/docs/)
- **DigitalOcean Support**: [digitalocean.com/support](https://www.digitalocean.com/support/)

