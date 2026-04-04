# Cloudflare KV Namespace Examples

This document provides copy-paste ready examples for creating and managing Cloudflare Workers KV namespaces using OpenMCF.

## Prerequisites

- OpenMCF CLI installed
- Cloudflare account with Workers enabled
- Cloudflare credentials configured (see [Credentials Setup](#credentials-setup))

## Credentials Setup

Create a Cloudflare credential resource with the required API token:

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: CloudflareCredential
metadata:
  name: my-cloudflare-credential
spec:
  api_token: ${CLOUDFLARE_API_TOKEN}  # Set via environment variable
```

**API Token Requirements:**
Your Cloudflare API token needs these permissions:
- **Workers KV Storage: Edit** (Account scope)
- **Workers Scripts: Edit** (Account scope) - if deploying Workers
- **Workers Routes: Edit** (Zone scope) - if using routes

Create the token at: https://dash.cloudflare.com/profile/api-tokens

## Example 1: Minimal Namespace

The simplest possible KV namespace with just a name:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-config
spec:
  namespaceName: myapp-config-prod
```

**Deploy:**

```bash
# Validate the manifest
openmcf validate --manifest kv-minimal.yaml

# Deploy with Pulumi
openmcf pulumi up \
  --manifest kv-minimal.yaml \
  --stack-name dev \
  --credential my-cloudflare-credential

# Deploy with OpenTofu/Terraform
openmcf tofu apply \
  --manifest kv-minimal.yaml \
  --auto-approve \
  --credential my-cloudflare-credential
```

**Output:**

```
Outputs:
  namespace_id: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
```

Use this `namespace_id` to bind the namespace to your Worker.

---

## Example 2: Development Environment with TTL

A development namespace with automatic key expiration for testing:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-dev-cache
spec:
  namespaceName: myapp-dev-cache
  ttlSeconds: 300  # Keys expire after 5 minutes
  description: "Development cache with short TTL for rapid testing"
```

**Use case:** Development environment where you want cache entries to expire quickly during testing.

**Deploy:**

```bash
openmcf pulumi up \
  --manifest kv-dev.yaml \
  --stack-name dev \
  --credential my-cloudflare-credential
```

---

## Example 3: Production Configuration Store

A production namespace for persistent configuration data:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-config-prod
spec:
  namespaceName: myapp-config-prod
  ttlSeconds: 0  # No automatic expiration (persistent data)
  description: "Feature flags and application configuration for production"
```

**Use case:** Configuration data and feature flags that should persist indefinitely.

**Deploy:**

```bash
openmcf pulumi up \
  --manifest kv-prod-config.yaml \
  --stack-name prod \
  --credential my-cloudflare-credential
```

---

## Example 4: Multi-Namespace Strategy

Deploy multiple namespaces for different purposes in production:

**Config Namespace:**

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-config-prod
spec:
  namespaceName: myapp-config-prod
  ttlSeconds: 0
  description: "Persistent configuration and feature flags"
```

**Cache Namespace:**

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-cache-prod
spec:
  namespaceName: myapp-cache-prod
  ttlSeconds: 3600  # Cache entries expire after 1 hour
  description: "Cached API responses and transient data"
```

**Session Namespace:**

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-session-prod
spec:
  namespaceName: myapp-session-prod
  ttlSeconds: 1800  # Sessions expire after 30 minutes
  description: "User session tokens and temporary auth data"
```

**Deploy all at once with Kustomize:**

```bash
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - config-namespace.yaml
  - cache-namespace.yaml
  - session-namespace.yaml
```

```bash
openmcf pulumi up \
  --manifest kustomization.yaml \
  --stack-name prod \
  --credential my-cloudflare-credential
```

---

## Example 5: Staging Environment

A staging namespace that mirrors production settings:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-config-staging
spec:
  namespaceName: myapp-config-staging
  ttlSeconds: 0
  description: "Staging environment feature flags (mirrors prod)"
```

**Use case:** Testing production-like behavior before deploying to prod.

---

## Example 6: Per-Developer Namespaces

Individual namespaces for each developer:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-dev-alice
spec:
  namespaceName: myapp-dev-alice
  ttlSeconds: 300
  description: "Alice's personal dev environment"
---
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-dev-bob
spec:
  namespaceName: myapp-dev-bob
  ttlSeconds: 300
  description: "Bob's personal dev environment"
```

**Naming convention:** `{app}-dev-{developer}`

**Use case:** Isolate development work to avoid conflicts.

---

## Post-Deployment: Seeding Data

After creating a namespace, you can seed it with initial data using Wrangler:

### Seed Individual Keys

```bash
# Get namespace ID from stack outputs
NAMESPACE_ID=$(openmcf pulumi output namespace_id --stack-name prod)

# Write a single key
wrangler kv:key put \
  --namespace-id=$NAMESPACE_ID \
  "feature_flags" \
  '{"dark_mode": true, "beta_features": false}'
```

### Bulk Import from JSON

Create a JSON file with key-value pairs:

```json
[
  {"key": "config/app_name", "value": "MyApp"},
  {"key": "config/api_url", "value": "https://api.example.com"},
  {"key": "flags/dark_mode", "value": "true"},
  {"key": "flags/beta_features", "value": "false"}
]
```

Import in bulk:

```bash
wrangler kv:bulk put \
  --namespace-id=$NAMESPACE_ID \
  seed-data.json
```

---

## Integration with Workers

After creating the namespace, bind it to your Worker in `wrangler.toml`:

```toml
name = "my-worker"
main = "src/index.js"

[[kv_namespaces]]
binding = "CONFIG"
id = "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"  # From stack outputs

[[kv_namespaces]]
binding = "CACHE"
id = "b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7"  # Another namespace

[env.dev]
[[env.dev.kv_namespaces]]
binding = "CONFIG"
id = "c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8"  # Dev namespace
```

Use in Worker code:

```javascript
export default {
  async fetch(request, env) {
    // Read from CONFIG namespace
    const appName = await env.CONFIG.get("config/app_name");
    
    // Read from CACHE namespace with JSON parsing
    const cachedData = await env.CACHE.get("api_response", { type: "json" });
    
    // Write with TTL
    await env.CONFIG.put("last_access", Date.now().toString(), {
      expirationTtl: 3600  // Expire after 1 hour
    });
    
    return new Response(`Hello from ${appName}!`);
  }
};
```

---

## Cleanup

To delete a namespace and all its data:

```bash
openmcf pulumi destroy \
  --manifest kv-namespace.yaml \
  --stack-name dev \
  --credential my-cloudflare-credential
```

**Warning:** This deletes the namespace and **all keys and values** within it. This operation cannot be undone.

---

## Best Practices

1. **Separate by environment:** Use distinct namespaces for dev, staging, and production
2. **Separate by purpose:** Don't mix configuration, cache, and session data in one namespace
3. **Use descriptive names:** Follow a consistent naming convention like `{app}-{purpose}-{env}`
4. **Document with descriptions:** Always fill in the `description` field for clarity
5. **Set appropriate TTLs:** Use TTLs for ephemeral data (sessions, cache), leave persistent config without TTL
6. **Monitor usage:** Track read/write operations to avoid exceeding quotas

---

## Troubleshooting

### Namespace creation fails

**Error:** `The namespace name already exists`

**Solution:** Namespace names must be unique within your Cloudflare account. Choose a different name or delete the existing namespace.

### Cannot access KV from Worker

**Error:** `CONFIG is undefined`

**Solution:** Ensure the namespace is properly bound in `wrangler.toml` with the correct `binding` name and `id`.

### Keys not propagating

**Issue:** Write a key but can't read it immediately from another region

**Solution:** This is expected behavior. KV is eventually consistent with propagation delays up to 60+ seconds. Design your application to tolerate this delay.

### Hitting quota limits

**Error:** `Daily read limit exceeded`

**Solution:** Either upgrade to a paid plan or optimize your read patterns (cache results, reduce redundant reads).

---

## Further Reading

- [Comprehensive Deployment Guide](./docs/README.md)
- [Cloudflare Workers KV Documentation](https://developers.cloudflare.com/kv/)
- [Wrangler CLI Reference](https://developers.cloudflare.com/workers/wrangler/commands/#kv)

