# GCP Cloud CDN

Deploys a Google Cloud CDN endpoint backed by a Global Application Load Balancer, with support for GCS bucket, Compute Engine, Cloud Run, and external origin backends. The component provisions the backend resource, URL map, global IP address, forwarding rules, and optionally Google-managed SSL certificates, HTTP-to-HTTPS redirect, and Cloud Armor integration.

## What Gets Created

When you deploy a GcpCloudCdn resource, Planton provisions:

- **Global IP Address** — a static anycast IP for the load balancer frontend
- **Backend Bucket or Backend Service** — one of the following, depending on the configured backend type:
  - Backend Bucket with CDN enabled (for GCS bucket origins)
  - Backend Service with CDN enabled and a Serverless NEG (for Cloud Run origins)
  - Backend Service with CDN enabled and an Internet NEG (for external origins)
  - Backend Service with CDN enabled (for Compute Engine instance group origins)
- **Health Check** — HTTP health check for Compute Engine backends (when `healthCheck` is configured)
- **URL Map** — routing configuration that directs all traffic to the backend
- **HTTPS Proxy + Forwarding Rule** — TLS termination and port-443 forwarding (when `frontendConfig` is specified)
- **Google-managed SSL Certificate** — automatic certificate provisioning and renewal (when `frontendConfig.sslCertificate.googleManaged` is specified)
- **HTTP Proxy + Forwarding Rule** — port-80 forwarding (when no `frontendConfig` is specified, or for HTTP-to-HTTPS redirect)
- **HTTP-to-HTTPS Redirect** — 301 redirect URL map, HTTP proxy, and forwarding rule (enabled by default when `frontendConfig` is specified)

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `gcpProjectId`
- **IAM permissions** to create Compute Engine load balancer resources in the target project
- **An existing origin** — depending on backend type:
  - A GCS bucket (for `gcsBucket` backend)
  - A Managed Instance Group (for `computeService` backend)
  - A deployed Cloud Run service (for `cloudRunService` backend)
  - A reachable external hostname (for `externalOrigin` backend)
- **DNS configuration** (optional) — if using custom domains, the ability to create DNS records pointing to the global IP

## Quick Start

Create a file `cdn.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudCdn
metadata:
  name: my-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpCloudCdn.my-cdn
spec:
  gcpProjectId:
    value: my-gcp-project-123
  backend:
    gcsBucket:
      bucketName: my-static-site-bucket
```

Deploy:

```shell
planton apply -f cdn.yaml
```

This creates a Cloud CDN endpoint backed by a GCS bucket, using the default `CACHE_ALL_STATIC` cache mode with a 1-hour default TTL and a 1-day max TTL.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `gcpProjectId` | `StringValueOrRef` | GCP project ID. Can reference a GcpProject resource via `valueFrom`. | Required |
| `backend` | `object` | Backend origin configuration. Exactly one backend type must be specified. | Required |

**Backend Types (one of):**

| Field | Type | Description |
|-------|------|-------------|
| `backend.gcsBucket.bucketName` | `string` | Name of the GCS bucket to use as origin. | 
| `backend.computeService.instanceGroupName` | `string` | Name of the Managed Instance Group to use as backend. |
| `backend.cloudRunService.serviceName` | `string` | Name of the Cloud Run service to use as backend. |
| `backend.cloudRunService.region` | `string` | GCP region where the Cloud Run service is deployed. |
| `backend.externalOrigin.hostname` | `string` | FQDN or IP address of the external origin. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `cacheMode` | `enum` | `CACHE_ALL_STATIC` | Caching strategy: `CACHE_ALL_STATIC`, `USE_ORIGIN_HEADERS`, or `FORCE_CACHE_ALL`. |
| `defaultTtlSeconds` | `int32` | `3600` | Default TTL in seconds when origin does not specify Cache-Control. Range: 0-31536000. |
| `maxTtlSeconds` | `int32` | `86400` | Maximum TTL in seconds. Hard ceiling that overrides origin headers. Range: 0-31536000. |
| `clientTtlSeconds` | `int32` | value of `maxTtlSeconds` | Client-facing TTL in seconds. Overrides max-age for browser caches. Range: 0-31536000. |
| `enableNegativeCaching` | `bool` | `false` | Cache HTTP 4xx/5xx error responses to reduce origin load during failures. |
| `advancedConfig` | `object` | — | Advanced CDN configuration (cache keys, signed URLs, negative caching policies). |
| `advancedConfig.cacheKeyPolicy` | `object` | — | Controls which request attributes are included in the cache key. |
| `advancedConfig.cacheKeyPolicy.includeQueryString` | `bool` | `true` | Include URL query string in cache key. |
| `advancedConfig.cacheKeyPolicy.queryStringWhitelist` | `string[]` | `[]` | Whitelist of query parameters to include in cache key (all others ignored). |
| `advancedConfig.cacheKeyPolicy.includeProtocol` | `bool` | `true` | Include request protocol (HTTP vs HTTPS) in cache key. |
| `advancedConfig.cacheKeyPolicy.includeHost` | `bool` | `true` | Include Host header in cache key. |
| `advancedConfig.cacheKeyPolicy.includedHeaders` | `string[]` | `[]` | Request headers to include in cache key. Supported: Accept, Accept-Encoding, Origin. |
| `advancedConfig.signedUrlConfig.enabled` | `bool` | — | Enable signed URL validation for private content delivery. |
| `advancedConfig.signedUrlConfig.keys` | `SignedUrlKey[]` | — | Signing keys for URL validation. Each key requires `keyName` and `keyValue`. |
| `advancedConfig.negativeCachingPolicies` | `NegativeCachingPolicy[]` | `[]` | Per-status-code caching policies. Each entry requires `code` (400-599) and `ttlSeconds` (0-86400). |
| `advancedConfig.serveWhileStaleSeconds` | `int32` | `0` | Serve stale content while revalidating with origin. Range: 0-604800. |
| `advancedConfig.enableRequestCoalescing` | `bool` | `true` | Combine multiple identical requests into one origin fetch. |
| `frontendConfig` | `object` | — | Load balancer frontend configuration (SSL, domains, Cloud Armor). |
| `frontendConfig.customDomains` | `string[]` | `[]` | Custom domains for the CDN endpoint. |
| `frontendConfig.sslCertificate.googleManaged.domains` | `string[]` | — | Domains for Google-managed SSL certificate (auto-provisioned via Let's Encrypt). |
| `frontendConfig.sslCertificate.selfManaged.certificatePem` | `string` | — | PEM-encoded SSL certificate chain (bring your own certificate). |
| `frontendConfig.sslCertificate.selfManaged.privateKeyPem` | `string` | — | PEM-encoded private key for self-managed certificate. |
| `frontendConfig.cloudArmor.enabled` | `bool` | — | Enable Cloud Armor WAF/DDoS protection. |
| `frontendConfig.cloudArmor.securityPolicyName` | `string` | — | Name of existing Cloud Armor security policy to attach. |
| `frontendConfig.enableHttpsRedirect` | `bool` | `true` | Create an HTTP-to-HTTPS 301 redirect. |
| `backend.gcsBucket.enableUniformAccess` | `bool` | `true` | Use uniform bucket-level IAM access (no legacy ACLs). |
| `backend.computeService.healthCheck` | `object` | — | Health check config for Compute Engine backends. |
| `backend.computeService.protocol` | `enum` | `HTTP` | Backend protocol: `HTTP` or `HTTPS`. |
| `backend.computeService.port` | `int32` | `80`/`443` | Port where backend instances serve traffic. |
| `backend.externalOrigin.port` | `int32` | `443`/`80` | Port for the external origin. |
| `backend.externalOrigin.protocol` | `enum` | `HTTPS` | Protocol for connecting to external origin: `HTTP` or `HTTPS`. |

## Examples

### Static Website with GCS Bucket Backend

Serve a static site from a GCS bucket with default caching settings:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudCdn
metadata:
  name: static-site-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpCloudCdn.static-site-cdn
spec:
  gcpProjectId:
    value: my-prod-project
  backend:
    gcsBucket:
      bucketName: my-static-site-bucket
  cacheMode: CACHE_ALL_STATIC
  defaultTtlSeconds: 3600
  maxTtlSeconds: 86400
  enableNegativeCaching: true
```

### Cloud Run Backend with Custom Domain and HTTPS

Cache a Cloud Run service behind a custom domain with a Google-managed SSL certificate:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudCdn
metadata:
  name: api-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpCloudCdn.api-cdn
spec:
  gcpProjectId:
    value: my-prod-project
  backend:
    cloudRunService:
      serviceName: my-api-service
      region: us-central1
  cacheMode: USE_ORIGIN_HEADERS
  defaultTtlSeconds: 300
  maxTtlSeconds: 3600
  frontendConfig:
    customDomains:
      - cdn.example.com
    sslCertificate:
      googleManaged:
        domains:
          - cdn.example.com
    enableHttpsRedirect: true
```

### Full-Featured with Advanced Caching and Cloud Armor

Production deployment with tuned cache keys, negative caching policies, stale-while-revalidate, signed URLs, and Cloud Armor protection:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudCdn
metadata:
  name: media-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpCloudCdn.media-cdn
spec:
  gcpProjectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  backend:
    gcsBucket:
      bucketName: media-assets-bucket
      enableUniformAccess: true
  cacheMode: FORCE_CACHE_ALL
  defaultTtlSeconds: 86400
  maxTtlSeconds: 604800
  clientTtlSeconds: 3600
  enableNegativeCaching: true
  advancedConfig:
    cacheKeyPolicy:
      includeQueryString: true
      queryStringWhitelist:
        - version
        - lang
      includeProtocol: false
      includeHost: true
    signedUrlConfig:
      enabled: true
      keys:
        - keyName: primary-key
          keyValue: "dGhpcy1pcy1hLXNhbXBsZS1rZXk="
        - keyName: rotation-key
          keyValue: "cm90YXRpb24ta2V5LXZhbHVl"
    negativeCachingPolicies:
      - code: 404
        ttlSeconds: 600
      - code: 503
        ttlSeconds: 60
    serveWhileStaleSeconds: 86400
    enableRequestCoalescing: true
  frontendConfig:
    customDomains:
      - media.example.com
      - cdn.example.com
    sslCertificate:
      googleManaged:
        domains:
          - media.example.com
          - cdn.example.com
    cloudArmor:
      enabled: true
      securityPolicyName: media-waf-policy
    enableHttpsRedirect: true
```

### External Origin Backend

Cache content from an origin server outside GCP (multi-cloud or on-premises):

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudCdn
metadata:
  name: hybrid-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.GcpCloudCdn.hybrid-cdn
spec:
  gcpProjectId:
    value: my-staging-project
  backend:
    externalOrigin:
      hostname: origin.example.com
      port: 443
      protocol: HTTPS
  cacheMode: USE_ORIGIN_HEADERS
  defaultTtlSeconds: 1800
  maxTtlSeconds: 7200
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cdnUrl` | `string` | Public URL of the Cloud CDN endpoint (e.g., `https://<ip-address>`) |
| `globalIpAddress` | `string` | Static global anycast IP address assigned to the load balancer. Use for DNS records. |
| `backendName` | `string` | Name of the backend resource (BackendBucket or BackendService) |
| `backendId` | `string` | Full resource ID of the backend (e.g., `projects/{project}/global/backendBuckets/{name}`) |
| `cdnEnabled` | `string` | Whether Cloud CDN is enabled on the backend |
| `cacheMode` | `string` | Cache mode configured for this CDN (`CACHE_ALL_STATIC`, `USE_ORIGIN_HEADERS`, `FORCE_CACHE_ALL`) |
| `urlMapName` | `string` | URL map name for load balancer routing configuration |
| `httpsProxyName` | `string` | Target HTTPS proxy name (set when `frontendConfig` is specified) |
| `sslCertificateName` | `string` | SSL certificate name or ID (set when `frontendConfig` is specified) |
| `cloudArmorPolicyName` | `string` | Cloud Armor security policy name (empty if Cloud Armor is not configured) |
| `backendType` | `string` | Backend type: `GCS_BUCKET`, `COMPUTE_SERVICE`, `CLOUD_RUN`, or `EXTERNAL` |
| `gcsBucketName` | `string` | GCS bucket name (only set when `backendType` is `GCS_BUCKET`) |
| `instanceGroupName` | `string` | Compute Engine instance group name (only set when `backendType` is `COMPUTE_SERVICE`) |
| `cloudRunServiceName` | `string` | Cloud Run service name (only set when `backendType` is `CLOUD_RUN`) |
| `cloudRunRegion` | `string` | Cloud Run service region (only set when `backendType` is `CLOUD_RUN`) |
| `externalHostname` | `string` | External origin hostname (only set when `backendType` is `EXTERNAL`) |
| `customDomains` | `string[]` | Custom domains configured for this CDN |
| `healthCheckUrl` | `string` | Health check URL (set when health check is configured for Compute Engine backends) |
| `monitoringDashboardUrl` | `string` | Cloud Console link to view CDN cache hit ratio, bandwidth, and request metrics |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project for CDN resource creation
- [GcpGcsBucket](/docs/catalog/gcp/gcpgcsbucket) — provisions a GCS bucket that can serve as the CDN origin
- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — container orchestration that can be fronted by Cloud CDN
- [GcpDnsRecord](/docs/catalog/gcp/gcpdnsrecord) — creates DNS records pointing custom domains to the CDN global IP
- [GcpDnsZone](/docs/catalog/gcp/gcpdnszone) — manages the DNS zone for custom domain configuration
