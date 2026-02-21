# DigitalOcean Container Registry

Manage private Docker container registries on DigitalOcean using a type-safe, protobuf-defined API with OpenMCF.

## Overview

**DigitalOceanContainerRegistry** enables you to provision and manage private container registries on DigitalOcean with a focus on the essential 80/20 configuration: registry name, subscription tier, region, and garbage collection.

DigitalOcean Container Registry (DOCR) is a fully managed private container registry designed to integrate seamlessly with DigitalOcean Kubernetes (DOKS) and App Platform. It provides OCI-compliant storage for Docker images and Helm charts with transparent pricing and zero operational overhead.

## Why Use This Component?

- **Type-Safe Configuration**: Protobuf-based API with compile-time validation prevents invalid registry configurations
- **80/20 Focused**: Exposes only the essential fields needed for most use cases (name, tier, region, garbage collection)
- **Production-Ready Defaults**: Follows DigitalOcean best practices for cost management and operational safety
- **Transparent Pricing**: Subscription tiers make cost predictable (Starter: free, Basic/Professional: fixed monthly rates)

## Quick Start

### Minimal Configuration (Starter Tier)

For development and testing:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: dev-registry
spec:
  name: dev-registry
  subscription_tier: STARTER
  region: nyc3
  garbage_collection_enabled: false
```

### Production Configuration (Professional Tier)

For production workloads with automated cleanup:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-registry
spec:
  name: prod-registry
  subscription_tier: PROFESSIONAL
  region: nyc3
  garbage_collection_enabled: true
```

## Key Features

### Subscription Tiers
- **STARTER**: Free tier with 500MB storage, 500MB bandwidth/month - perfect for dev/testing
- **BASIC**: $20/month with 5GB storage, 5GB bandwidth/month - suitable for small teams
- **PROFESSIONAL**: $50/month with 100GB storage, 100GB bandwidth/month - production-ready

### Garbage Collection
- **Automated Cleanup**: When enabled, OpenMCF schedules garbage collection to remove untagged images
- **Cost Control**: Prevents storage bloat from orphaned layers and old image versions
- **Read-Only Awareness**: GC operations are scheduled during maintenance windows to avoid disrupting CI/CD

### DOKS Integration
- **Automatic Integration**: OpenMCF handles the "1-click" DOKS integration that Terraform/Pulumi can't do
- **ImagePullSecrets**: Automatically configures service accounts across all namespaces
- **Seamless Pulls**: Kubernetes pods can pull images without manual credential management

## Configuration Reference

### Core Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Registry name (1-63 chars, lowercase, alphanumeric + hyphens) |
| `subscription_tier` | enum | Yes | STARTER, BASIC, or PROFESSIONAL |
| `region` | enum | Yes | DigitalOcean region (e.g., nyc3, sfo3, lon1) |
| `garbage_collection_enabled` | bool | No | Enable automated cleanup of untagged images (default: false) |

### Subscription Tier Selection

**Choose STARTER when:**
- Running dev/test environments
- Learning Docker and container workflows
- Building personal projects
- Cost is the primary concern

**Choose BASIC when:**
- Running small production workloads
- Team of 2-5 developers
- Storage needs are < 5GB
- Bandwidth usage is predictable

**Choose PROFESSIONAL when:**
- Running production at scale
- Multiple teams/projects sharing the registry
- Storage needs are > 5GB
- Need headroom for growth

### Region Selection

**Best Practices:**
- **Co-locate with DOKS**: Choose the same region as your Kubernetes cluster for fastest image pulls
- **Data Residency**: If you have compliance requirements, choose a region in the appropriate jurisdiction
- **Latency**: For App Platform, choose the region closest to your users

**Common Regions:**
- `nyc3` - New York (US East)
- `sfo3` - San Francisco (US West)
- `lon1` - London (Europe)
- `fra1` - Frankfurt (Europe)
- `sgp1` - Singapore (Asia)

## Outputs

After successful provisioning, the following outputs are available:

| Output | Description |
|--------|-------------|
| `endpoint` | Registry hostname (e.g., `registry.digitalocean.com/my-registry`) |
| `server_url` | Full registry URL for Docker login |
| `name` | Registry name |

## Common Use Cases

### 1. Development Registry (Free Tier)

```yaml
spec:
  name: dev-team-registry
  subscription_tier: STARTER
  region: nyc3
  garbage_collection_enabled: true  # Keep it clean even on free tier
```

### 2. Production Registry (Professional Tier)

```yaml
spec:
  name: prod-app-registry
  subscription_tier: PROFESSIONAL
  region: nyc3  # Same as DOKS cluster
  garbage_collection_enabled: true
```

### 3. Multi-Region Setup

Deploy separate registries in different regions for global apps:

```yaml
# US Registry
spec:
  name: us-registry
  subscription_tier: PROFESSIONAL
  region: nyc3
  garbage_collection_enabled: true
---
# EU Registry
spec:
  name: eu-registry
  subscription_tier: PROFESSIONAL
  region: fra1
  garbage_collection_enabled: true
```

## Best Practices

### Cost Management
1. **Enable Garbage Collection**: Prevents storage bloat from untagged images (can save 50%+ on storage costs)
2. **Start with STARTER**: Use the free tier for dev/test, upgrade to paid tiers only for production
3. **Monitor Usage**: Track storage consumption in the DigitalOcean dashboard
4. **Tag Strategy**: Use semantic versioning tags instead of `latest` to make GC more predictable

### Security
1. **Limit Access**: Use DigitalOcean's token-based authentication with least-privilege scopes
2. **Scan Images**: Integrate with Snyk (via DigitalOcean UI) for vulnerability scanning
3. **Private by Default**: All DOCR registries are private; images are never publicly accessible
4. **Credential Rotation**: Docker credentials expire; use DOKS integration for automatic rotation

### Operational Safety
1. **Schedule GC During Off-Hours**: Garbage collection puts the registry in read-only mode
2. **Test in Staging First**: Validate registry configs in a non-production environment
3. **Co-locate with Workloads**: Same-region placement = faster pulls and lower latency
4. **Plan for Growth**: Professional tier offers 100GB; consider this when estimating image storage

## Integration

### With DigitalOcean Kubernetes (DOKS)

OpenMCF automatically configures DOKS integration:
- Creates imagePullSecrets in all namespaces
- Patches service accounts to use these secrets
- Handles credential rotation

Your Kubernetes deployments can reference images directly:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: registry.digitalocean.com/prod-registry/my-app:v1.0.0
```

### With App Platform

App Platform auto-deploys when you push to DOCR (exclusive feature):

1. Deploy registry via OpenMCF
2. Configure App Platform component to watch DOCR
3. `docker push` triggers automatic deployment

## Troubleshooting

### Registry Name Already Exists

**Cause**: Registry names are globally unique across all DigitalOcean accounts.

**Solution**: Choose a different, more specific name (e.g., `company-app-registry` instead of `app-registry`).

### Storage Limit Exceeded

**Cause**: Too many images/layers, or garbage collection disabled.

**Solution**:
1. Enable `garbage_collection_enabled: true`
2. Wait for scheduled GC to run
3. If still over limit, upgrade to next tier (Starter → Basic → Professional)

### Image Pull Failures in DOKS

**Cause**: DOKS integration not configured or credentials expired.

**Solution**: OpenMCF handles this automatically. If using manual Terraform/Pulumi, you'll need to manually create imagePullSecrets.

### Garbage Collection Breaks CI/CD

**Cause**: GC runs during business hours, putting registry in read-only mode.

**Solution**: OpenMCF schedules GC during configured maintenance windows. Verify your maintenance window is set correctly.

## Validation Rules

The protobuf spec enforces these constraints at compile-time:

- `name`: 1-63 characters, lowercase alphanumeric + hyphens, must start/end with alphanumeric
- `subscription_tier`: Must be STARTER, BASIC, or PROFESSIONAL
- `region`: Must be a valid DigitalOcean region enum value
- `garbage_collection_enabled`: Boolean (default: false)

## Limitations and Workarounds

### What OpenMCF Handles

✅ Registry provisioning (name, tier, region)  
✅ Garbage collection scheduling (via custom controller)  
✅ DOKS integration (imagePullSecrets automation)  
✅ Cost optimization (automated cleanup)

### What Requires Manual Configuration

❌ **Image Signing**: Not supported by DOCR; use external tools like Cosign  
❌ **Vulnerability Scanning**: Configure via DigitalOcean UI (Snyk integration)  
❌ **Access Control**: Manage via DigitalOcean API tokens (not registry-level permissions)

### The IaC Gap

Terraform and Pulumi can create the registry but **cannot**:
- Schedule or automate garbage collection
- Integrate with DOKS ("1-click" feature)
- Configure vulnerability scanning

OpenMCF bridges these gaps with custom controllers.

## Further Reading

- **Comprehensive Guide**: See [docs/README.md](./docs/README.md) for deep-dive coverage of deployment methods, anti-patterns, and production essentials
- **Examples**: See [examples.md](./examples.md) for copy-paste ready manifests
- **Pulumi Module**: See [iac/pulumi/README.md](./iac/pulumi/README.md) for standalone Pulumi usage
- **Terraform Module**: See [iac/tf/README.md](./iac/tf/README.md) for standalone Terraform usage

## Support

For issues, questions, or contributions, refer to the [OpenMCF documentation](https://openmcf.org) or file an issue in the repository.

---

**TL;DR**: Use STARTER tier for dev/test, PROFESSIONAL for production. Enable garbage collection to avoid surprise bills. Co-locate with DOKS for fastest image pulls.
