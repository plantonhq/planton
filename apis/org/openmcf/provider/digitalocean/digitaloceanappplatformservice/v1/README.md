# DigitalOcean App Platform Service

## Overview

The **DigitalOcean App Platform Service** API resource provides a simplified, declarative interface for deploying containerized applications and services on DigitalOcean's fully-managed Platform-as-a-Service (PaaS). Following OpenMCF's 80/20 principle, this API exposes only the essential configuration fields needed for the vast majority of deployments, abstracting away platform complexity while maintaining production-grade capabilities.

DigitalOcean App Platform eliminates infrastructure management by automatically handling container orchestration, load balancing, SSL certificates, autoscaling, and zero-downtime deployments. Unlike managed Kubernetes or bare Droplets, App Platform provides a Heroku-like developer experience with transparent, predictable pricing and deep integration with DigitalOcean's ecosystem (Managed Databases, Container Registry, Spaces object storage, VPC networking).

This API resource integrates with OpenMCF's unified infrastructure framework, using a Kubernetes-style manifest format (`apiVersion`, `kind`, `metadata`, `spec`) for consistency across cloud providers.

## Key Features

### Deployment Patterns

- **Multiple Service Types**:
  - `web_service`: HTTP-accessible services with load balancing and optional autoscaling
  - `worker`: Background processing services (queue workers, async tasks)
  - `job`: One-off or pre-deployment tasks (database migrations, cron jobs)

- **Flexible Source Configuration**:
  - **Git-based deployment**: Build from GitHub/GitLab repositories using Cloud Native Buildpacks or Dockerfiles
  - **Container image deployment**: Deploy pre-built images from DigitalOcean Container Registry (DOCR) or other registries

### Production Features

- **Autoscaling**: CPU-based horizontal autoscaling for web services (configure min/max instance counts)
- **Instance Sizing**: Flexible instance types from Basic (shared CPU) to Professional (dedicated CPU) tiers
- **Environment Variables**: Secure runtime configuration with encrypted secret storage
- **Custom Domains**: Automatic SSL certificate provisioning and DNS integration
- **Zero-downtime Deployments**: Rolling updates with health checks and automatic rollback
- **Built-in Monitoring**: Request metrics, logs, and resource utilization dashboards

### Integration Capabilities

- **Foreign Key References**: Reference other OpenMCF resources (DNS zones, databases, container registries) using declarative field paths
- **Automatic Credential Management**: Seamless integration with DigitalOcean Container Registry for private image pulls
- **Multi-environment Support**: Deploy consistent configurations across dev, staging, and production using variable substitution

## Use Cases

### Web Applications

Deploy REST APIs, web services, or microservices that handle HTTP traffic with autoscaling and load balancing.

**Ideal for**: Node.js/Python/Ruby/Go web applications, API gateways, GraphQL servers

### Background Workers

Run long-running background processes for queue processing, data pipelines, or asynchronous task execution.

**Ideal for**: Redis queue workers (Sidekiq, Celery, Bull), Kafka consumers, ETL processors

### Scheduled Jobs

Execute periodic tasks like database migrations, data cleanup, or cron-based operations.

**Ideal for**: Schema migrations, backup tasks, report generation, batch processing

### Static Sites with APIs

Combine static frontends (served via CDN) with dynamic API backends in a single app configuration.

**Ideal for**: JAMstack applications, SPAs with backend APIs

## Basic Example

Here's a minimal configuration for deploying a web service from a Git repository:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanAppPlatformService
metadata:
  name: my-api
spec:
  service_name: my-api
  region: nyc1
  service_type: web_service
  
  git_source:
    repo_url: https://github.com/myorg/my-api.git
    branch: main
  
  instance_size_slug: basic_xxs
  instance_count: 1
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | Yes | Unique name for the app (DNS-friendly, max 63 chars) |
| `region` | enum | Yes | DigitalOcean region (e.g., `nyc1`, `sfo3`, `ams3`) |
| `service_type` | enum | Yes | Service type: `web_service`, `worker`, or `job` |
| `git_source` | object | Conditional | Git repository source (either this or `image_source` required) |
| `image_source` | object | Conditional | Container image source (either this or `git_source` required) |
| `instance_size_slug` | enum | Yes | Instance size (e.g., `basic_xxs`, `professional_s`) |
| `instance_count` | int | No | Number of instances (default: 1, ignored if autoscale enabled) |
| `enable_autoscale` | bool | No | Enable autoscaling (default: false, only for `web_service`) |
| `min_instance_count` | int | Conditional | Minimum instances for autoscaling (required if autoscale enabled) |
| `max_instance_count` | int | Conditional | Maximum instances for autoscaling (required if autoscale enabled) |
| `env` | map | No | Environment variables as key-value pairs |
| `custom_domain` | string | No | Custom domain name (automatically provisions SSL) |

### Git Source Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repo_url` | string | Yes | HTTPS URL of the git repository |
| `branch` | string | Yes | Branch to deploy from |
| `build_command` | string | No | Custom build command (overrides auto-detection) |
| `run_command` | string | No | Custom run command (overrides auto-detection) |

### Image Source Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `registry` | string/ref | Yes | Container registry URL or reference to DigitalOceanContainerRegistry resource |
| `repository` | string | Yes | Image repository name (e.g., `myapp/backend`) |
| `tag` | string | Yes | Image tag to deploy (e.g., `latest`, `v1.0.0`) |

## Available Instance Sizes

### Basic Tier (Shared CPU)
- `basic_xxs`: $5/month - 512MB RAM, suitable for dev/test
- `basic_xs`: $12/month - 1GB RAM, small production apps
- `basic_s`: $24/month - 2GB RAM, moderate traffic
- `basic_m`: $48/month - 4GB RAM, higher throughput
- `basic_l`: $96/month - 8GB RAM, large workloads

### Professional Tier (Dedicated CPU)
- `professional_xs`: $12/month - 1GB RAM, 1 vCPU
- `professional_s`: $24/month - 2GB RAM, 1 vCPU
- `professional_m`: $48/month - 4GB RAM, 2 vCPU
- `professional_l`: $96/month - 8GB RAM, 4 vCPU
- `professional_xl`: $192/month - 16GB RAM, 8 vCPU

**Recommendation**: Use Basic tier for development and low-traffic services. Use Professional tier for production web services requiring consistent performance and autoscaling.

## Available Regions

Common DigitalOcean regions:
- `nyc1`, `nyc3` - New York (USA)
- `sfo3` - San Francisco (USA)
- `tor1` - Toronto (Canada)
- `lon1` - London (UK)
- `ams3` - Amsterdam (Netherlands)
- `fra1` - Frankfurt (Germany)
- `sgp1` - Singapore
- `blr1` - Bangalore (India)

See the [DigitalOcean documentation](https://docs.digitalocean.com/products/platform/availability-matrix/) for the complete list of supported regions.

## Stack Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `app_id` | The unique identifier of the app (DigitalOcean App Platform application ID) |
| `default_hostname` | The default hostname assigned to the app (usually ending in "ondigitalocean.app") |
| `live_url` | The publicly accessible URL (including protocol) of the deployed service |

## Examples

For comprehensive examples including:
- Web services with autoscaling
- Worker services for background processing
- Container image deployment from DOCR
- Multi-environment configurations
- Custom domain setup
- Database integration

See [examples.md](./examples.md)

## Infrastructure as Code

This API resource can be deployed using:

### Pulumi
See [iac/pulumi/README.md](./iac/pulumi/README.md) for Pulumi-based deployment instructions.

### Terraform
See [iac/tf/README.md](./iac/tf/README.md) for Terraform-based deployment instructions.

## Best Practices

### Source Configuration
- **Use Git sources** for rapid iteration in dev/staging environments
- **Use container images** for production deployments requiring immutable, tested artifacts
- **Pin image tags** to specific versions (avoid `latest` in production)

### Scaling
- **Enable autoscaling** for web services with variable traffic patterns
- **Set min_instance_count ≥ 2** for production high-availability
- **Use Professional tier instances** for services requiring autoscaling

### Security
- **Store secrets in environment variables** (encrypted at rest by DigitalOcean)
- **Rotate credentials regularly** and redeploy apps to update env vars
- **Use custom domains** for production services (avoid `.ondigitalocean.app` URLs)

### Performance
- **Choose the nearest region** to your users for lowest latency
- **Monitor CPU and memory metrics** to right-size instance types
- **Enable HTTP/2 and compression** automatically provided by App Platform

## Limitations

- Autoscaling is only supported for `web_service` type
- Custom domains require DNS verification (managed automatically if using DigitalOceanDnsZone resource)
- Git sources must be publicly accessible or use GitHub/GitLab OAuth integration
- Workers and jobs cannot use autoscaling (use fixed `instance_count`)

## Troubleshooting

### Deployment Fails with "Build Failed"
- Check build logs in DigitalOcean App Platform console
- Verify build command is correct (or let buildpack auto-detect)
- Ensure all required dependencies are declared in package manager files

### App Crashes Immediately After Deploy
- Check runtime logs for startup errors
- Verify run command starts a long-running process (not a one-off script)
- Ensure environment variables are configured correctly

### Custom Domain Shows "Not Configured"
- Verify DNS records are created and pointing to App Platform
- Allow up to 48 hours for DNS propagation
- Check SSL certificate provisioning status in DigitalOcean console

## Support

For issues and questions:
- **OpenMCF Documentation**: [docs.planton.ai](https://docs.planton.ai)
- **DigitalOcean App Platform Docs**: [docs.digitalocean.com/products/app-platform](https://docs.digitalocean.com/products/app-platform/)
- **OpenMCF GitHub**: [github.com/plantonhq/openmcf](https://github.com/plantonhq/openmcf)

## References

- [DigitalOcean App Platform Overview](https://docs.digitalocean.com/products/app-platform/)
- [App Spec Reference](https://docs.digitalocean.com/products/app-platform/reference/app-spec/)
- [Buildpack Documentation](https://docs.digitalocean.com/products/app-platform/languages-frameworks/)
- [OpenMCF API Docs](https://buf.build/openmcf/apis/docs)
