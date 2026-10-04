# KubernetesRedis: Migration to Bitnami Legacy Redis Image

**Date**: December 4, 2025
**Type**: Bug Fix
**Components**: Kubernetes Provider, Pulumi CLI Integration, Terraform Integration

## Summary

Fixed critical deployment failures in KubernetesRedis by migrating from the deprecated `docker.io/bitnami/redis` image to `docker.io/bitnamilegacy/redis:8.2.1-debian-12-r0`. The change updates both Pulumi and Terraform implementations with centralized image configuration, ensuring consistent deployments across all Redis instances.

## Problem Statement / Motivation

The official Bitnami Redis repository (`docker.io/bitnami/redis`) was deprecated and removed, causing all new KubernetesRedis deployments to fail with `ImagePullBackOff` errors. Users attempting to deploy Redis instances encountered:

```
[Pod redis-app-prod-tekton-logs/tekton-logs-master-0]: containers with unready status: [redis][ErrImagePull] 
rpc error: code = NotFound desc = failed to pull and unpack image "docker.io/bitnami/redis:7.0.11-debian-11-r0": 
failed to resolve reference "docker.io/bitnami/redis:7.0.11-debian-11-r0": 
docker.io/bitnami/redis:7.0.11-debian-11-r0: not found
```

### Pain Points

- **All new Redis deployments failing** with image pull errors
- **Production Redis instances** (central-cache, tekton-logs) unable to deploy
- **No clear migration path** from deprecated repository
- **Inconsistent configuration** between Pulumi and Terraform implementations
- **Hardcoded image values** making updates difficult

## Solution / What's New

Migrated to the Bitnami Legacy repository (`docker.io/bitnamilegacy/redis`) which provides continued access to stable Redis images. The solution includes:

1. **Centralized image configuration** in variables/locals
2. **Explicit image overrides** in Helm chart values
3. **Consistent implementation** across both Pulumi and Terraform
4. **Updated to Redis 8.2.1** (latest stable version)

### Image Configuration

**New Image**: `docker.io/bitnamilegacy/redis:8.2.1-debian-12-r0`

**Pulumi Configuration** (`variables.go`):
```go
RedisImageRegistry:      "docker.io",
RedisImageRepository:    "bitnamilegacy/redis",
RedisImageTag:           "8.2.1-debian-12-r0",
```

**Terraform Configuration** (`locals.tf`):
```hcl
redis_image_registry    = "docker.io"
redis_image_repository  = "bitnamilegacy/redis"
redis_image_tag         = "8.2.1-debian-12-r0"
```

## Implementation Details

### Pulumi Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetesredis/v1/iac/pulumi/module/variables.go`

Added three new configuration variables:
- `RedisImageRegistry`: Docker registry hosting the image
- `RedisImageRepository`: Repository path for Redis image
- `RedisImageTag`: Specific image version tag

**File**: `apis/dev/planton/provider/kubernetes/kubernetesredis/v1/iac/pulumi/module/helm_chart.go`

Updated Helm chart values to explicitly override the image:
```go
Values: pulumi.Map{
    "fullnameOverride": pulumi.String(locals.KubernetesRedis.Metadata.Name),
    "architecture":     pulumi.String("standalone"),
    "image": pulumi.Map{
        "registry":   pulumi.String(vars.RedisImageRegistry),
        "repository": pulumi.String(vars.RedisImageRepository),
        "tag":        pulumi.String(vars.RedisImageTag),
    },
    // ... rest of configuration
}
```

### Terraform Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetesredis/v1/iac/tf/locals.tf`

Added Redis image locals matching Pulumi configuration:
```hcl
# Redis image configuration (using legacy Bitnami repository)
redis_image_registry    = "docker.io"
redis_image_repository  = "bitnamilegacy/redis"
redis_image_tag         = "8.2.1-debian-12-r0"
```

**File**: `apis/dev/planton/provider/kubernetes/kubernetesredis/v1/iac/tf/helm_chart.tf`

Added image override in Helm values:
```hcl
values = [
  yamlencode({
    fullnameOverride = var.metadata.name
    architecture     = "standalone"

    image = {
      registry   = local.redis_image_registry
      repository = local.redis_image_repository
      tag        = local.redis_image_tag
    }
    // ... rest of configuration
  })
]
```

### Helm Chart Compatibility

The image override aligns with the official Bitnami Redis Helm chart schema:

```yaml
image:
  registry: docker.io
  repository: bitnami/redis
  tag: 8.2.1-debian-12-r0
```

Our configuration properly overrides these default values to use `bitnamilegacy/redis`.

## Benefits

### Immediate Benefits

- ✅ **All Redis deployments now succeed** - No more ImagePullBackOff errors
- ✅ **Production services restored** - Critical infrastructure components operational
- ✅ **Latest Redis version** - Upgraded from 7.0.11 to 8.2.1
- ✅ **Consistent implementation** - Same image across Pulumi and Terraform

### Long-term Benefits

- 🔧 **Centralized configuration** - Single source of truth for image settings
- 🔧 **Easy version updates** - Change once in variables/locals, applies everywhere
- 🔧 **Better maintainability** - No hardcoded image references scattered in code
- 🔧 **Clear migration path** - Simple to update to newer images in future

## Impact

### Affected Components

**Direct Impact**:
- All `KubernetesRedis` resource deployments (Pulumi and Terraform)
- Existing Redis instances: `central-cache`, `tekton-logs`
- Helm chart integration for Redis

**User Impact**:
- Users can now successfully deploy Redis instances
- Existing deployments are unaffected (running containers continue)
- New deployments will automatically use the legacy image

### Deployment Changes

**Before**:
```bash
$ planton pulumi up -f redis.yaml
# ❌ Failed: ImagePullBackOff - docker.io/bitnami/redis:7.0.11 not found
```

**After**:
```bash
$ planton pulumi up -f redis.yaml
# ✅ Success: Pods running with docker.io/bitnamilegacy/redis:8.2.1
```

## Testing

### Verification Steps

1. Deploy Redis with Pulumi:
   ```bash
   planton pulumi up -f redis-central-cache.yaml --module-dir ${REDIS_MODULE}
   ```

2. Verify image in running pods:
   ```bash
   kubectl get pods -n redis-app-prod-central-cache -o jsonpath='{.items[*].spec.containers[*].image}'
   # Output: docker.io/bitnamilegacy/redis:8.2.1-debian-12-r0
   ```

3. Confirm Redis connectivity:
   ```bash
   kubectl exec -n redis-app-prod-central-cache redis-master-0 -- redis-cli ping
   # Output: PONG
   ```

## Related Work

### Previous Issues

This fix resolves deployment failures that affected:
- Production Redis for central application cache
- Redis for Tekton pipeline log storage
- All development and staging Redis instances

### Future Considerations

- **Monitor Bitnami Legacy lifecycle** - The legacy repository is maintained but may eventually be deprecated
- **Consider Redis Operator** - For production workloads, Redis Operator provides better high availability
- **Version update strategy** - Define process for upgrading Redis versions across environments

## Known Limitations

- **Legacy repository dependency** - Using `bitnamilegacy` which may have slower updates
- **Manual version updates required** - Image version is pinned and must be manually updated
- **No automated security updates** - Image tag is fixed, won't auto-update for security patches

---

**Status**: ✅ Production Ready  
**Files Changed**: 6 (Pulumi: 2, Terraform: 2, Build: 2)  
**Commit**: `c6a449e4` - `fix(kubernetesredis): use bitnamilegacy redis image to fix deprecated registry`

