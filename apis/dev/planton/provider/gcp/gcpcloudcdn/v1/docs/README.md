# GCP Cloud CDN Deployment Architecture

## Introduction: The Invisible CDN

When you search for "how to create a Cloud CDN resource" in GCP documentation, you encounter a curious puzzle: **there is no standalone Cloud CDN resource**. Unlike other GCP services where you create a discrete resource (a VM, a bucket, a database), Cloud CDN exists only as a feature flag—a boolean switch and configuration block nested within another resource.

This is not a documentation oversight. It's the architectural reality of how Google designed their CDN: Cloud CDN is an integrated caching layer within the **Global Application Load Balancer** stack, not a separate product you provision independently.

This architectural choice has profound implications:
- You cannot "create a CDN" without first creating (or referencing) a load balancer backend
- The CDN configuration lives as a `cdn_policy` block on either a `BackendService` or `BackendBucket` resource
- All the power of Cloud CDN—caching, SSL/TLS termination, DDoS protection—flows from its tight integration with Google's global edge infrastructure (the Google Front End, or GFE)

This document explains how to think about deploying Cloud CDN, what deployment methods exist across the maturity spectrum, and how Planton abstracts this complexity into a clean, declarative API.

## The Architecture: CDN as a Load Balancer Feature

### The Core Dependency

Cloud CDN operates exclusively as part of Google's **Global Application Load Balancer**. Understanding the request flow illuminates why this integration is mandatory:

1. **User Request** → A user requests `https://example.com/image.png`
2. **Anycast Routing** → DNS resolves to a global anycast IP, routing the user to the nearest Google edge Point of Presence (PoP)
3. **GFE Cache Check** → The Google Front End checks its local Cloud CDN cache
   - **Cache Hit**: Content served instantly from edge, never touching the origin
   - **Cache Miss**: Request forwarded to the origin server (GCS bucket, Compute Engine, Cloud Run, etc.)
4. **Response & Cache Fill** → Origin response flows back to user and populates the edge cache for future requests

This architecture provides several automatic benefits:
- **Integrated SSL/TLS**: HTTPS termination happens at the load balancer, with free Google-managed certificates
- **Integrated Security**: Cloud Armor (WAF/DDoS protection) sits at the same layer, protecting both the CDN and origin
- **Premium Network**: Traffic travels over Google's private global backbone, not the public internet (note: Cloud CDN **requires** Premium Network Tier—you cannot use the cheaper Standard Tier)

### The Two Backend Types

Cloud CDN is enabled on one of two underlying GCP resources:

1. **`BackendBucket`** - For Google Cloud Storage origins
   - Use case: Static websites, file downloads, media hosting
   - Simplest configuration: point a backend bucket at a GCS bucket, flip `enable_cdn = true`
   - Represents ~80% of Cloud CDN deployments

2. **`BackendService`** - For compute origins
   - Use case: Dynamic web apps, APIs, containerized services
   - Supports multiple backend types:
     - Compute Engine (Managed Instance Groups)
     - GKE (via Ingress and Serverless NEGs)
     - Cloud Run (via Serverless NEGs)
     - External/hybrid origins (via Internet NEGs, e.g., AWS S3, on-prem servers)

The choice of backend dictates not just where your content comes from, but also which configuration options are available.

### Cloud CDN vs. Media CDN

Google offers two distinct CDN products, optimized for different workloads:

| Product | Optimized For | Primary API | Analogous To |
|---------|---------------|-------------|--------------|
| **Cloud CDN** | Web and API acceleration (small, frequent requests) | `compute.v1.BackendService/BackendBucket` with `cdn_policy` | Cloud SQL (transactional, low-latency) |
| **Media CDN** | Video streaming and large file downloads (high throughput) | `networkservices.v1.EdgeCacheService` | BigQuery (analytical, high-throughput) |

This document focuses on **Cloud CDN** (the web/API acceleration product). Media CDN uses a completely different API surface and resource model.

## The Maturity Spectrum: Deployment Methods

Cloud CDN can be deployed through a wide range of methods, from manual console clicks to fully declarative infrastructure-as-code. Understanding this spectrum helps you choose the right tool for your maturity level and use case.

### Level 0: The Anti-Pattern (Manual Console Configuration)

**How it works**: Navigate to "Network Services > Load balancing" in the GCP Console, click through a multi-step wizard, check the "Enable Cloud CDN" box.

**Prerequisites**: Before you can enable CDN, you must manually create:
- A global static IP address
- An SSL certificate (for HTTPS)
- A target HTTPS proxy
- A URL map
- A backend service or backend bucket

**Verdict**: This workflow is suitable only for initial experimentation or one-off demos. Any production deployment requires repeatability and version control. The console workflow is error-prone (one missing prerequisite breaks everything) and leaves no audit trail.

### Level 1: Scripted Automation (gcloud CLI)

**How it works**: Use imperative shell commands to create and configure resources:

```bash
# Create a backend bucket with CDN enabled
gcloud compute backend-buckets create static-site-backend \
  --gcs-bucket-name=my-static-site-bucket \
  --enable-cdn

# Update CDN policy
gcloud compute backend-services update webapp-backend \
  --enable-cdn \
  --cache-mode=CACHE_ALL_STATIC \
  --default-ttl=3600

# Invalidate cache (imperative operation)
gcloud compute url-maps invalidate-cdn-cache my-url-map \
  --path="/images/*"
```

**Verdict**: Useful for CI/CD scripts (especially cache invalidation tasks) and one-off operational tasks. Not suitable as the primary deployment method because state is not tracked—repeated execution may create duplicate resources or fail unpredictably.

### Level 2: Configuration Management (Ansible)

**How it works**: Use Ansible's `google.cloud` collection to manage GCP resources declaratively:

```yaml
- name: Create CDN-enabled backend bucket
  google.cloud.gcp_compute_backend_bucket:
    name: static-site-backend
    bucketName: my-static-site-bucket
    enable_cdn: true
    cdn_policy:
      cache_mode: CACHE_ALL_STATIC
      default_ttl: 3600
    state: present
```

**Verdict**: A solid choice for teams already invested in Ansible. Provides idempotency and integrates well with existing playbook-based infrastructure. However, Ansible lacks native state management (like Terraform state files), making it harder to detect and reconcile drift.

### Level 3: The Production Foundation (Terraform, Pulumi, OpenTofu)

**How it works**: Define infrastructure as code using declarative DSLs (Terraform HCL) or general-purpose languages (Pulumi TypeScript/Python).

**Terraform Example**:

```hcl
resource "google_compute_backend_bucket" "static_site" {
  name        = "static-site-backend"
  bucket_name = google_storage_bucket.static_site.name
  enable_cdn  = true

  cdn_policy {
    cache_mode  = "CACHE_ALL_STATIC"
    default_ttl = 3600
    max_ttl     = 86400

    cache_key_policy {
      include_query_string   = true
      query_string_whitelist = ["version", "lang"]
    }
  }
}
```

**Pulumi Example**:

```typescript
const staticSiteBackend = new gcp.compute.BackendBucket("static-site", {
  bucketName: staticSiteBucket.name,
  enableCdn: true,
  cdnPolicy: {
    cacheMode: "CACHE_ALL_STATIC",
    defaultTtl: 3600,
    maxTtl: 86400,
    cacheKeyPolicy: {
      includeQueryString: true,
      queryStringWhitelist: ["version", "lang"],
    },
  },
});
```

**Why this is the production standard**:
- **State Management**: Tracks the current state of infrastructure, enabling plan/apply workflows
- **Drift Detection**: Can detect when manual changes diverge from declared configuration
- **Reusability**: Modules and components enable DRY principles
- **Multi-Cloud**: Terraform and Pulumi support multiple cloud providers with a consistent workflow

**Verdict**: This is the industry standard for production infrastructure. Choose Terraform/OpenTofu for HCL-based declarative config, or Pulumi if you prefer expressing infrastructure in TypeScript, Python, or Go.

### Level 4: Kubernetes-Native IaC (Config Connector, Crossplane)

**How it works**: Manage GCP resources as Kubernetes Custom Resources, using `kubectl apply` instead of cloud provider CLIs.

**Config Connector (GCP-specific)**:

```yaml
apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeBackendBucket
metadata:
  name: static-site-backend
spec:
  bucketRef:
    name: static-site-bucket
  enableCdn: true
  cdnPolicy:
    cacheMode: CACHE_ALL_STATIC
    defaultTtl: 3600
```

**Crossplane (multi-cloud)**:

Similar YAML structure but uses Crossplane's composition model to abstract across clouds.

**Verdict**: Ideal for teams practicing GitOps with Kubernetes. Config Connector is simpler if you're GCP-only; Crossplane is better for multi-cloud environments. Both provide the "infrastructure as Kubernetes resources" experience that platform teams value.

### Level 5: Highest Abstraction (GKE Ingress + BackendConfig)

**How it works**: For applications running on GKE, you don't manage the `BackendService` at all. The GKE Ingress controller creates it automatically. To enable CDN, you create a `BackendConfig` CRD and annotate your Kubernetes Service:

```yaml
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: my-app-backend-config
spec:
  cdn:
    enabled: true
    cachePolicy:
      includeHost: true
      includeProtocol: true
      includeQueryString: false
---
apiVersion: v1
kind: Service
metadata:
  name: my-app-service
  annotations:
    cloud.google.com/backend-config: '{"default": "my-app-backend-config"}'
spec:
  type: LoadBalancer
  # ...
```

**Verdict**: This is the **cleanest abstraction** for GKE users. The entire GCP networking stack (load balancer, backend service, CDN) is managed by Kubernetes annotations and CRDs. Application developers never touch GCP APIs directly. However, this abstraction is GKE-specific and not portable to other clouds or GCP compute platforms.

## Production Essentials: Configuration Deep Dive

Deploying Cloud CDN is straightforward. **Optimizing** it for performance, cost, and security requires understanding its configuration knobs.

### Cache Modes: The Most Critical Decision

The `cache_mode` field dictates what gets cached and when to trust origin headers.

| Mode | Behavior | Recommended Use Case | Anti-Pattern |
|------|----------|---------------------|--------------|
| **`CACHE_ALL_STATIC`** (default) | Automatically caches common static file types (CSS, JS, images). Also caches any response with valid `Cache-Control` headers. | **90% of deployments**. Safe and flexible for backends serving a mix of static and dynamic content. | None. This is the recommended default. |
| **`USE_ORIGIN_HEADERS`** | Strict mode. Caches **only** if the origin sends valid `Cache-Control` or `Expires` headers. All other content bypasses the cache. | Backends where the application provides perfect, granular cache headers. | Using this on a GCS bucket or static server that doesn't send `Cache-Control` headers → 0% cache hit ratio. |
| **`FORCE_CACHE_ALL`** | Aggressive mode. Caches **all** 200 OK responses, ignoring `Cache-Control: private` or `no-store`. | Public GCS buckets containing 100% public, non-sensitive static assets. | **Data leak vulnerability**: Using this on any backend serving dynamic or user-specific content. |

### Cache Keys: Solving the "Cache Shattering" Problem

By default, the cache key is a composite of:

```
Protocol + Host + Path + Query String
```

This leads to a classic anti-pattern called **cache shattering**:

```
# Two requests for the same content, but different query strings
Request 1: /api/items?user_id=123&session=abc
Request 2: /api/items?user_id=456&session=xyz

# Result: Two separate cache entries, each used once → 0% cache hit ratio
```

**The Solution**: Use the `cache_key_policy` to whitelist only the query parameters that define unique content:

```hcl
cache_key_policy {
  include_query_string   = true
  query_string_whitelist = ["version", "lang", "page"]
  # Now "user_id" and "session" are ignored in the cache key
}
```

**The `Vary` Header Pitfall**: Cloud CDN only respects three `Vary` headers: `Accept`, `Accept-Encoding`, and `Origin`. If your backend sends `Vary: Cookie`, caching will break (all requests will miss). Remove or avoid unsupported `Vary` headers.

### TTLs: Time-to-Live Configuration

```hcl
cdn_policy {
  default_ttl = 3600   # 1 hour - for responses without Cache-Control
  max_ttl     = 86400  # 1 day - hard ceiling, overrides origin headers
  client_ttl  = 1800   # 30 min - overrides max-age for browser cache
  
  negative_caching = true  # Cache 404/503 errors (best practice)
}
```

**Production Pattern**: Use **versioned URLs** instead of cache invalidation:

```
# Anti-pattern: Short TTLs + frequent cache invalidation
Cache-Control: max-age=300  # 5 minutes
# On deploy: gcloud compute url-maps invalidate-cdn-cache --path="/app.css"

# Production pattern: Long TTLs + versioned filenames
Cache-Control: max-age=31536000  # 1 year
# Filename: app.a83f5b2c.css (hash changes on each build)
# On deploy: No action needed - new HTML references new filename
```

This eliminates the need for cache invalidation entirely, resulting in atomic, instant deployments.

### Security: Signed URLs for Private Content

To serve private content (e.g., paid videos, user-specific downloads) via Cloud CDN, use **Signed URLs**:

1. Create a secret signing key and attach it to the backend
2. Your application generates a time-limited URL with a cryptographic signature
3. Users present this URL to Cloud CDN
4. Cloud CDN validates the signature before serving the (potentially cached) private file

**Configuration (BackendBucket)**:

```hcl
resource "google_compute_backend_bucket_signed_url_key" "default" {
  name           = "my-cdn-key"
  key_value      = random_id.cdn_key.b64_url
  backend_bucket = google_compute_backend_bucket.default.name
}
```

**Note**: For `BackendService`, the API is inconsistent—you must use an imperative `gcloud` command or direct API call. This is a pain point that higher-level abstractions (like Planton) can smooth over.

### Observability: Measuring Cache Hit Ratio

The **Cache Hit Ratio (CHR)** is the single most important metric for both performance and cost optimization. A low CHR indicates configuration problems, not a broken CDN.

Enable logging on your backend to track cache performance:

```hcl
log_config {
  enable      = true
  sample_rate = 1.0  # 100% of requests (adjust for high-traffic sites)
}
```

Query logs in Cloud Logging to identify cache behavior:

- **Cache Hit**: `httpRequest.cacheHit = true AND httpRequest.cacheValidatedWithOriginServer != true`
- **Cache Hit (Revalidated)**: `httpRequest.cacheHit = true AND httpRequest.cacheValidatedWithOriginServer = true` (e.g., 304 Not Modified)
- **Cache Miss**: `httpRequest.cacheHit != true AND jsonPayload.statusDetails = "response_sent_by_backend"`

**Root causes of low CHR**:
1. Cache key shattering (query string not whitelisted)
2. Unsupported `Vary` headers from origin
3. Incorrect cache mode for your backend type

## Common Anti-Patterns to Avoid

| Anti-Pattern | Why It's Wrong | Correct Approach |
|--------------|----------------|------------------|
| Using `FORCE_CACHE_ALL` on dynamic backends | **Data leak**: Caches user-specific or private data | Use `CACHE_ALL_STATIC` or `USE_ORIGIN_HEADERS` |
| Not defining a `cache_key_policy` | Cache shattering → 0% hit ratio | Whitelist query params that define unique content |
| Sending `Vary: Cookie` from origin | Cloud CDN doesn't support it → all requests miss | Remove or use a different caching strategy |
| Relying on cache invalidation for deploys | Slow, rate-limited, unreliable | Use versioned/hashed filenames with long TTLs |
| Using `USE_ORIGIN_HEADERS` without sending headers | No `Cache-Control` from origin → 0% hit ratio | Ensure origin sends headers, or use `CACHE_ALL_STATIC` |

## Integration Patterns

### Pattern 1: Static Website (GCS + BackendBucket)

**Use case**: Hosting a static website (HTML, CSS, JS, images)

**Resources**:
- GCS bucket configured for website hosting
- `BackendBucket` pointing to the GCS bucket with CDN enabled
- Global load balancer with HTTPS

**80% Configuration**:
```hcl
resource "google_compute_backend_bucket" "website" {
  name        = "static-site-backend"
  bucket_name = google_storage_bucket.website.name
  enable_cdn  = true

  cdn_policy {
    cache_mode  = "CACHE_ALL_STATIC"
    default_ttl = 3600
  }
}
```

### Pattern 2: Dynamic Web App (Compute Engine + BackendService)

**Use case**: Traditional web application on VMs (monolith or microservices)

**Resources**:
- Managed Instance Group (MIG)
- `BackendService` with CDN and custom cache key policy
- HTTPS load balancer with Cloud Armor

**Production Configuration**:
```hcl
resource "google_compute_backend_service" "webapp" {
  name     = "webapp-backend"
  protocol = "HTTP"
  
  backend {
    group = google_compute_instance_group_manager.webapp.instance_group
  }

  enable_cdn = true
  cdn_policy {
    cache_mode  = "USE_ORIGIN_HEADERS"
    default_ttl = 3600
    max_ttl     = 86400
    
    cache_key_policy {
      include_query_string   = true
      query_string_whitelist = ["page", "category", "version"]
    }
  }
}
```

### Pattern 3: Serverless (Cloud Run + Serverless NEG)

**Use case**: Modern containerized applications on Cloud Run

**Resources**:
- Cloud Run service
- Serverless Network Endpoint Group (NEG)
- `BackendService` pointing to the NEG with CDN enabled

**Planton Abstraction**: Planton would automatically create the serverless NEG and wire it to the backend service, abstracting this complexity from users.

### Pattern 4: Kubernetes (GKE + BackendConfig CRD)

**Use case**: Applications deployed on GKE

**GKE-Specific Abstraction**: Use a `BackendConfig` CRD instead of managing GCP resources directly. The GKE Ingress controller handles all the underlying infrastructure.

**Note for Planton**: This pattern requires a separate resource type (e.g., `GkeCloudCdnPolicy`) rather than the standard `GcpCloudCdn` resource, as the management model is fundamentally different.

### Pattern 5: Hybrid/Multi-Cloud (Internet NEG)

**Use case**: Using Cloud CDN in front of an external origin (AWS S3, on-prem server)

**Resources**:
- Internet Network Endpoint Group pointing to external FQDN
- `BackendService` with CDN enabled

**Strategic Note**: This pattern enables gradual migration from other clouds or on-prem to GCP while immediately gaining the performance benefits of Google's edge network.

## Cost and Performance Model

### Pricing Structure

Cloud CDN costs have three components:

1. **Cache Egress**: Data served from edge caches to users (~$0.02–$0.08/GB depending on region)
2. **Cache Fill**: Data transferred from origin to CDN cache (~$0.01–$0.04/GB)
3. **Cache Lookup Requests**: Per-request fee (~$0.0075 per 10,000 requests)

**Optimization Equation**: Maximize cache hit ratio (CHR). A high CHR converts expensive "standard egress" into cheaper "cache egress" and minimizes cache fill costs.

### The Premium Tier Mandate

Cloud CDN **only works on Google's Premium Network Tier**. This is non-negotiable. You cannot opt for the cheaper Standard Tier egress and use Cloud CDN.

- **Premium Tier**: Traffic travels over Google's private global backbone, exiting at an edge PoP near the user (high performance, higher cost)
- **Standard Tier**: Traffic exits from the origin region over the public internet (lower performance, lower cost)

**Implication**: Choosing Cloud CDN is also choosing Premium Tier networking for all traffic from that load balancer.

### Strategies for Maximizing CHR

1. **Tune cache keys** (most important): Whitelist only content-defining query params
2. **Use correct cache mode**: `CACHE_ALL_STATIC` for most use cases
3. **Use long TTLs with versioned URLs**: Embed file hash in filename, set 1-year TTL
4. **Enable negative caching**: Cache 404s to prevent origin overload
5. **Use cache tags** (advanced): Invalidate by tag instead of URL for bulk operations

## What Planton Supports

Planton's `GcpCloudCdn` resource provides a **compositional abstraction** over GCP's complex networking stack. Rather than forcing users to manually create and wire together load balancers, backend services/buckets, SSL certificates, and CDN policies, Planton presents a unified resource that manages this entire stack as a cohesive unit.

### Design Philosophy: Composition Over References

The current `GcpCloudCdnSpec` (which contains only `gcp_project_id`) is intentionally minimal as the API evolves. The next iteration will adopt a **compositional model** where a single Planton resource:

1. Defines the **origin** (GCS bucket, Compute Engine service, Cloud Run, or external backend)
2. Specifies the **CDN policy** (cache mode, TTLs, cache keys)
3. Optionally manages the **load balancer frontend** (SSL certificates, domains, Cloud Armor policies)

This mirrors the mental model of "I want a CDN for this bucket/service" while abstracting away GCP's implementation details.

### 80/20 Configuration Principle

Following Planton's philosophy, the API will expose:

**Essential (80%)** - Always visible:
- Backend type and configuration
- Cache mode
- TTLs (`default_ttl`, `max_ttl`, `client_ttl`)
- Negative caching

**Advanced (20%)** - Nested for clarity:
- Cache key policy (query string whitelists, headers)
- Signed URL configuration
- Granular negative caching policies per status code
- Serve-while-stale settings

### Abstracting GCP's Inconsistencies

One area where Planton adds immediate value: **Signed URL key management**. GCP's API has an inconsistency:
- For `BackendBucket`: Declarative resource (`google_compute_backend_bucket_signed_url_key`)
- For `BackendService`: Imperative command only (`gcloud compute backend-services add-signed-url-key`)

Planton's API presents a unified `signed_url_policy` field. The controller handles the correct underlying implementation (declarative resource or imperative API call) based on the backend type.

### GKE: A Special Case

For applications on GKE, Planton will provide a separate resource (e.g., `GkeCloudCdnPolicy`) that models the `BackendConfig` CRD. This recognizes that the GKE management model (Kubernetes-native annotations and CRDs) is fundamentally different from managing raw GCP networking resources.

## Conclusion

Google Cloud CDN challenges conventional assumptions about how CDNs are deployed. It is not a standalone service you "create," but rather a feature you enable on a load balancer backend. This tight integration with Google's Global Application Load Balancer provides automatic SSL/TLS, DDoS protection, and access to Google's global edge network—but also imposes architectural constraints, such as the mandatory Premium Network Tier.

For teams choosing Cloud CDN, the decision is typically driven by existing GCP investments: if your origin is already on GCS, Compute Engine, GKE, or Cloud Run, Cloud CDN provides the fastest, most seamless integration.

For Planton, this architecture dictates a **compositional resource model**. Rather than asking users to separately manage backends and CDN policies, Planton's API will treat the origin and CDN as a single, cohesive unit—abstracting away GCP's complexity while preserving the full power of its configuration model.

The next evolution of the `GcpCloudCdn` API will embrace this reality, providing a clean 80/20 interface that makes the simple case trivial (static website CDN in 10 lines of YAML) while making the advanced cases possible (signed URLs, custom cache keys, multi-region failover).

---

**For deeper implementation details**, see the following guides (planned):
- [GCS Static Website CDN Setup](./gcs-static-website-guide.md)
- [Cloud Run CDN Integration](./cloud-run-cdn-guide.md)
- [Advanced Cache Key Tuning](./cache-key-optimization-guide.md)
- [Signed URLs and Private Content](./signed-urls-guide.md)

