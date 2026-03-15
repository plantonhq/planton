# Cloudflare Worker Examples

Complete, working examples for common Cloudflare Worker patterns using OpenMCF.

## Table of Contents

- [Example 1: Minimal Worker (No Route)](#example-1-minimal-worker-no-route)
- [Example 2: Worker with Custom Domain](#example-2-worker-with-custom-domain)
- [Example 3: API Gateway with KV Storage](#example-3-api-gateway-with-kv-storage)
- [Example 4: Webhook Handler](#example-4-webhook-handler)
- [Example 5: Authentication Middleware](#example-5-authentication-middleware)
- [Example 6: Multi-Environment Deployment](#example-6-multi-environment-deployment)
- [Example 7: Worker with Environment Variables](#example-7-worker-with-environment-variables)
- [Example 8: A/B Testing Worker](#example-8-ab-testing-worker)

---

## Example 1: Minimal Worker (No Route)

**Use Case**: Deploy a Worker without attaching it to a custom domain. Accessible via `*.workers.dev` subdomain.

**Features**:
- Simplest possible configuration
- No DNS or route configuration
- Good for testing and development

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: hello-worker
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"  # Replace with your account ID
  workerName: hello-worker
  
  scriptBundle:
    bucket: my-workers-bucket
    path: builds/hello-worker-v1.0.0.js
  
  # No DNS configuration - Worker accessible at hello-worker.<account>.workers.dev
  compatibilityDate: "2025-01-15"
```

**Worker Code** (`hello-worker-v1.0.0.js`):
```javascript
export default {
  async fetch(request, env, ctx) {
    return new Response("Hello from Cloudflare Workers!", {
      headers: { "Content-Type": "text/plain" }
    });
  }
};
```

**Access**:
```bash
curl https://hello-worker.<your-subdomain>.workers.dev
# Returns: Hello from Cloudflare Workers!
```

---

## Example 2: Worker with Custom Domain

**Use Case**: Production Worker accessible via your own domain.

**Features**:
- Custom domain (api.example.com)
- Automatic DNS record creation
- Route pattern configuration

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-worker
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  workerName: api-production
  
  scriptBundle:
    bucket: workers-prod
    path: api/v1.0.0.js
  
  dns:
    enabled: true
    zoneId: { value: "zone123abc456def" }
    # Or reference a CloudflareDnsZone resource:
    # zoneId: { valueFrom: { kind: CloudflareDnsZone, name: "my-zone", fieldPath: "status.outputs.zone_id" } }
    hostname: api.example.com
    routePattern: api.example.com/*  # Match all paths
  
  compatibilityDate: "2025-01-15"
```

**Expected Behavior**:
1. DNS record created: `api.example.com` (AAAA record with proxy enabled)
2. Worker route created: `api.example.com/*`
3. All requests to `https://api.example.com/*` handled by Worker

**Worker Code**:
```javascript
export default {
  async fetch(request) {
    const url = new URL(request.url);
    return new Response(`You requested: ${url.pathname}`);
  }
};
```

**Test**:
```bash
curl https://api.example.com/users
# Returns: You requested: /users
```

---

## Example 3: API Gateway with KV Storage

**Use Case**: Production API gateway with rate limiting and caching using KV.

**Features**:
- KV namespace bindings
- Environment variables
- Secrets management
- Custom domain

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-gateway
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  workerName: api-gateway-prod
  
  scriptBundle:
    bucket: workers-prod
    path: gateway/v2.1.0.js
  
  # Bind KV namespaces
  kvBindings:
    - name: RATE_LIMIT_KV
      field_path: "rate-limit-namespace-id"
    - name: CACHE_KV
      field_path: "cache-namespace-id"
  
  dns:
    enabled: true
    zoneId: { value: "zone123" }
    hostname: api.example.com
    routePattern: api.example.com/v1/*
  
  env:
    variables:
      LOG_LEVEL: "info"
      RATE_LIMIT: "1000"
      CACHE_TTL: "3600"
    secrets:
      BACKEND_API_KEY: "$secrets-group/backend/api-key"
      JWT_SECRET: "$secrets-group/auth/jwt-secret"
  
  compatibilityDate: "2025-01-15"
  usageModel: 0  # BUNDLED
```

**Worker Code**:
```javascript
export default {
  async fetch(request, env) {
    const clientIp = request.headers.get('CF-Connecting-IP');
    
    // Rate limiting with KV
    const requestCount = await env.RATE_LIMIT_KV.get(clientIp) || 0;
    if (requestCount > parseInt(env.RATE_LIMIT)) {
      return new Response('Rate limit exceeded', { status: 429 });
    }
    await env.RATE_LIMIT_KV.put(clientIp, requestCount + 1, { expirationTtl: 60 });
    
    // Check cache
    const cacheKey = request.url;
    const cached = await env.CACHE_KV.get(cacheKey);
    if (cached) {
      return new Response(cached, { 
        headers: { 'X-Cache': 'HIT' }
      });
    }
    
    // Call backend with secret
    const response = await fetch('https://backend.example.com/api', {
      headers: {
        'Authorization': `Bearer ${env.BACKEND_API_KEY}`
      }
    });
    
    const data = await response.text();
    
    // Cache response
    await env.CACHE_KV.put(cacheKey, data, { 
      expirationTtl: parseInt(env.CACHE_TTL)
    });
    
    return new Response(data);
  }
};
```

---

## Example 4: Webhook Handler

**Use Case**: Handle GitHub webhooks at the edge.

**Features**:
- Route pattern matching specific path
- Webhook signature verification
- Secret management

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: github-webhook
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  workerName: github-webhook-handler
  
  scriptBundle:
    bucket: workers-prod
    path: webhooks/github-v1.0.0.js
  
  dns:
    enabled: true
    zoneId: { value: "zone123" }
    hostname: webhooks.example.com
    routePattern: webhooks.example.com/github/*
  
  env:
    secrets:
      GITHUB_WEBHOOK_SECRET: "$secrets-group/github/webhook-secret"
  
  compatibilityDate: "2025-01-15"
```

**Worker Code**:
```javascript
import crypto from 'crypto';

export default {
  async fetch(request, env) {
    if (request.method !== 'POST') {
      return new Response('Method not allowed', { status: 405 });
    }
    
    // Verify GitHub signature
    const signature = request.headers.get('X-Hub-Signature-256');
    const body = await request.text();
    
    const hmac = crypto.createHmac('sha256', env.GITHUB_WEBHOOK_SECRET);
    const digest = 'sha256=' + hmac.update(body).digest('hex');
    
    if (signature !== digest) {
      return new Response('Invalid signature', { status: 401 });
    }
    
    // Process webhook
    const payload = JSON.parse(body);
    console.log(`Received ${payload.action} event`);
    
    // TODO: Process event (e.g., trigger CI/CD, update KV, call backend)
    
    return new Response('Webhook processed', { status: 200 });
  }
};
```

---

## Example 5: Authentication Middleware

**Use Case**: Validate JWTs at the edge before proxying to backend.

**Features**:
- JWT verification
- Fast authentication (edge-native)
- Backend protection

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: auth-middleware
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  workerName: auth-middleware
  
  scriptBundle:
    bucket: workers-prod
    path: auth/middleware-v1.0.0.js
  
  dns:
    enabled: true
    zoneId: { value: "zone123" }
    hostname: app.example.com
    routePattern: app.example.com/api/*
  
  env:
    secrets:
      JWT_PUBLIC_KEY: "$secrets-group/auth/jwt-public-key"
    variables:
      BACKEND_URL: "https://backend.example.com"
  
  compatibilityDate: "2025-01-15"
```

**Worker Code**:
```javascript
import { jwtVerify } from 'jose';

export default {
  async fetch(request, env) {
    // Extract JWT from Authorization header
    const authHeader = request.headers.get('Authorization');
    if (!authHeader?.startsWith('Bearer ')) {
      return new Response('Unauthorized', { status: 401 });
    }
    
    const token = authHeader.substring(7);
    
    // Verify JWT
    try {
      const { payload } = await jwtVerify(token, env.JWT_PUBLIC_KEY);
      
      // Add user info to headers for backend
      const newRequest = new Request(request);
      newRequest.headers.set('X-User-ID', payload.sub);
      newRequest.headers.set('X-User-Role', payload.role);
      
      // Proxy to backend
      const url = new URL(request.url);
      url.hostname = new URL(env.BACKEND_URL).hostname;
      
      return fetch(url.toString(), newRequest);
    } catch (err) {
      return new Response('Invalid token', { status: 401 });
    }
  }
};
```

---

## Example 6: Multi-Environment Deployment

**Use Case**: Separate Workers for staging and production with different configurations.

```yaml
---
# Staging Worker
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-staging
spec:
  accountId: "staging-account-32-hex-chars"
  workerName: api-staging
  
  scriptBundle:
    bucket: workers-staging
    path: api/v2.1.0.js
  
  dns:
    enabled: true
    zoneId: { value: "staging-zone-id" }
    hostname: api-staging.example.com
  
  env:
    variables:
      ENVIRONMENT: "staging"
      DEBUG: "true"
      BACKEND_URL: "https://backend-staging.example.com"
  
  compatibilityDate: "2025-01-15"

---
# Production Worker
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-prod
spec:
  accountId: "prod-account-32-hex-chars"
  workerName: api-prod
  
  scriptBundle:
    bucket: workers-prod
    path: api/v2.1.0.js  # Same bundle, different config
  
  dns:
    enabled: true
    zoneId: { value: "prod-zone-id" }
    hostname: api.example.com
  
  env:
    variables:
      ENVIRONMENT: "production"
      DEBUG: "false"
      BACKEND_URL: "https://backend.example.com"
    secrets:
      API_KEY: "$secrets-group/prod/api-key"
  
  compatibilityDate: "2025-01-15"
  usageModel: 1  # UNBOUND for high traffic
```

---

## Example 7: Worker with Environment Variables

**Use Case**: Configure Worker behavior via environment variables.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: config-demo
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  workerName: config-demo-worker
  
  scriptBundle:
    bucket: workers-prod
    path: demos/config-v1.0.0.js
  
  env:
    variables:
      # Feature flags
      FEATURE_NEW_UI: "true"
      FEATURE_BETA_API: "false"
      
      # Service URLs
      AUTH_SERVICE_URL: "https://auth.example.com"
      USER_SERVICE_URL: "https://users.example.com"
      
      # Configuration
      MAX_RETRIES: "3"
      TIMEOUT_MS: "5000"
      LOG_LEVEL: "info"
    
    secrets:
      # Sensitive credentials
      DATABASE_URL: "$secrets-group/db/connection-string"
      OAUTH_CLIENT_SECRET: "$secrets-group/oauth/client-secret"
      ENCRYPTION_KEY: "$secrets-group/crypto/encryption-key"
  
  compatibilityDate: "2025-01-15"
```

**Worker Code**:
```javascript
export default {
  async fetch(request, env) {
    // Access environment variables
    const logLevel = env.LOG_LEVEL || 'info';
    const maxRetries = parseInt(env.MAX_RETRIES);
    
    // Feature flag check
    if (env.FEATURE_NEW_UI === 'true') {
      // New UI logic
    }
    
    // Use secrets (never logged)
    const dbConnection = env.DATABASE_URL;
    
    return new Response('Config loaded');
  }
};
```

---

## Example 8: A/B Testing Worker

**Use Case**: Route percentage of traffic to different backends for A/B testing.

**Features**:
- Cookie-based assignment
- Percentage-based traffic splitting
- KV for persistent assignments

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: ab-test-worker
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  workerName: ab-test-router
  
  scriptBundle:
    bucket: workers-prod
    path: experiments/ab-test-v1.0.0.js
  
  kvBindings:
    - name: AB_TEST_KV
      field_path: "ab-test-namespace-id"
  
  dns:
    enabled: true
    zoneId: { value: "zone123" }
    hostname: app.example.com
  
  env:
    variables:
      VARIANT_A_URL: "https://app-v1.example.com"
      VARIANT_B_URL: "https://app-v2.example.com"
      VARIANT_B_PERCENTAGE: "10"  # 10% to variant B
  
  compatibilityDate: "2025-01-15"
```

**Worker Code**:
```javascript
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    
    // Check for existing assignment cookie
    const cookie = request.headers.get('Cookie');
    let variant = cookie?.match(/variant=([AB])/)?.[1];
    
    if (!variant) {
      // New user - assign variant
      const random = Math.random() * 100;
      variant = random < parseInt(env.VARIANT_B_PERCENTAGE) ? 'B' : 'A';
      
      // Persist assignment
      await env.AB_TEST_KV.put(
        url.pathname,
        JSON.stringify({ variant, timestamp: Date.now() }),
        { expirationTtl: 86400 }
      );
    }
    
    // Route to appropriate backend
    const backendUrl = variant === 'B' ? env.VARIANT_B_URL : env.VARIANT_A_URL;
    url.hostname = new URL(backendUrl).hostname;
    
    const response = await fetch(url.toString(), request);
    
    // Set variant cookie
    const newResponse = new Response(response.body, response);
    newResponse.headers.set('Set-Cookie', `variant=${variant}; Max-Age=86400; Path=/`);
    newResponse.headers.set('X-Variant', variant);
    
    return newResponse;
  }
};
```

---

## Bundle Creation and Upload

Before deploying any Worker, you must create and upload the bundle to R2:

### Step 1: Build Worker

```bash
# Using Wrangler
npx wrangler build

# Output: dist/worker.js
```

### Step 2: Upload to R2

```bash
# Upload to R2 using AWS CLI
aws s3 cp dist/worker.js \
  s3://my-workers-bucket/builds/worker-v1.0.0.js \
  --endpoint-url https://<account-id>.r2.cloudflarestorage.com
```

### Step 3: Deploy Worker

Use the R2 path in your CloudflareWorker manifest:

```yaml
scriptBundle:
  bucket: my-workers-bucket
  path: builds/worker-v1.0.0.js
```

---

## Testing Your Worker

### Test Without Custom Domain

Access via `*.workers.dev`:

```bash
curl https://your-worker.<subdomain>.workers.dev
```

### Test With Custom Domain

```bash
curl https://api.example.com/test
```

### Test with Headers

```bash
curl -H "Authorization: Bearer token123" https://api.example.com/
```

### View Logs

Check Cloudflare Dashboard → Workers & Pages → your-worker → Logs

---

## Common Issues and Solutions

### Issue: "Script not found"

**Cause**: Bundle doesn't exist in R2 or wrong path

**Solution**:
1. Verify bundle uploaded to R2:
   ```bash
   aws s3 ls s3://my-bucket/builds/ \
     --endpoint-url https://<account-id>.r2.cloudflarestorage.com
   ```
2. Check path matches exactly (case-sensitive)

---

### Issue: "KV namespace binding failed"

**Cause**: KV namespace ID doesn't exist or is incorrect

**Solution**:
1. Verify KV namespace exists
2. Check `field_path` contains correct namespace ID
3. Ensure namespace is in the same account

---

### Issue: "Route already exists"

**Cause**: Another Worker is using the same route pattern

**Solution**:
1. Use more specific route pattern
2. Remove conflicting route
3. Use different hostname

---

### Issue: Worker returns 502

**Cause**: Script threw unhandled exception

**Solution**:
1. Check Worker logs in Cloudflare Dashboard
2. Add error handling:
   ```javascript
   try {
     // your code
   } catch (err) {
     console.error(err);
     return new Response('Internal error', { status: 500 });
   }
   ```

---

## Next Steps

- Read the [README.md](./README.md) for detailed field documentation
- Review [research documentation](./docs/README.md) for V8 isolates deep dive
- Deploy using [Pulumi](./iac/pulumi/README.md) or [Terraform](./iac/tf/README.md)
- Check [Cloudflare Workers documentation](https://developers.cloudflare.com/workers/)

---

**Questions or Issues?** Refer to the [README.md](./README.md) or Cloudflare's official Workers documentation.

