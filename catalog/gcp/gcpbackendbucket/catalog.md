# GCP Backend Bucket

Deploys a Compute Engine backend bucket — the piece that serves a Cloud Storage bucket's objects through an external HTTP(S) load balancer, optionally cached at Google's edge by Cloud CDN. It is the static-content counterpart of a backend service: URL maps route paths like /assets/* to a backend bucket while dynamic paths go to backend services. The backend bucket is deliberately a separate node from the bucket itself — one GCS bucket can sit behind several backend buckets with different CDN policies, and swapping the origin bucket is an in-place update that leaves the URL map untouched.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine API enablement** (`compute.googleapis.com`) on the target project (never disabled on destroy)
- **Compute Engine Backend Bucket** -- fronting the configured GCS origin for external load balancers (or, rarely, the cross-region internal ALB)
- **Cloud CDN policy** -- when enabled: cache mode, TTLs, negative caching, the cache key, and up to 3 signed-URL signing keys

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the backend bucket will be created — it may differ from the project owning the GCS bucket (cross-project origins are valid).
- **The origin bucket** (GcpGcsBucket) with publicly readable objects — or signed-URL serving configured. The load balancer does not authenticate to the bucket.

## Deploy

### Console

Open the deployment store, find **GCP Backend Bucket**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **CDN-Cached Static Assets** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBackendBucket
metadata:
  name: assets-backend
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  bucketName:
    value: "acme-static-assets"
  enableCdn: true
  cdnPolicy:
    cacheMode: CACHE_ALL_STATIC
    defaultTtl: 86400
```

```shell
planton apply -f backend-bucket.yaml
```

This creates the standard static-assets shape: a GCS origin cached at the edge for a day. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the origin:

```yaml
spec:
  bucketName:
    valueFrom:
      kind: GcpGcsBucket
      name: static-site
      fieldPath: status.outputs.bucket_id
```

The InfraPipeline resolves the dependency graph — the bucket first, then the backend bucket — and a downstream GcpUrlMap references this backend's `self_link` for its static routes.

## Key Configuration

These are the most important decisions when configuring a backend bucket. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Origin bucket** -- The one REQUIRED field, and deliberately MUTABLE: pointing at a different bucket is an in-place update, which makes blue/green static releases a one-field edit (upload the new release to a fresh bucket, swap the origin, roll back by swapping again).

**Cloud CDN** -- `enableCdn` plus `cdnPolicy`: the cache mode (CACHE_ALL_STATIC is GCP's default; USE_ORIGIN_HEADERS defers to the bucket's Cache-Control; FORCE_CACHE_ALL never for private content), TTLs, negative caching (cache the 404 storm away from the origin), request coalescing, and this kind's cache key (query whitelist + key headers only). Incompatible with the INTERNAL_MANAGED scheme — CDN only fronts external LBs.

**Signed-URL keys** -- Up to 3 named keys (each value secret-handled) for serving private content from the cache with expiring, tamper-proof links; rotate add-then-remove.

**Load balancer family** -- Leave `loadBalancingScheme` unset for global external HTTP(S) load balancers — the overwhelmingly common case. INTERNAL_MANAGED serves cross-region internal ALBs instead. Immutable.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpGcsBucket** | `bucketName` | `status.outputs.bucket_id` |
| **GcpCloudArmorPolicy** | `edgeSecurityPolicy` | `status.outputs.policy_self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the backend bucket | GcpUrlMap static-path routes and asset-domain default services |
| `backend_bucket_name` | Name as it exists in GCP | Audit, fleet inventory |
| `bucket_name` | The live origin bucket | Verifying which release is being served after a swap |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CDN static assets** -- The standard shape: long TTLs, narrow cache key, fingerprinted assets. Start from the **CDN-Cached Static Assets** preset.

**Plain origin** -- No CDN — every request proxies to the bucket; the starting point before caching decisions. Start from the **Plain Origin (No CDN)** preset.

**Signed-URL private CDN** -- Paid/gated downloads served from the edge with expiring links. Start from the **Private Content via CDN Signed URLs** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the backend bucket is created
- [**GCP GCS Bucket**](/infra-catalog/gcp-gcs-bucket) -- the origin whose objects are served
- [**GCP URL Map**](/infra-catalog/gcp-url-map) -- consumes this backend's `self_link` for static routes
- [**GCP Backend Service**](/infra-catalog/gcp-backend-service) -- the dynamic sibling on the same URL map
