# GCP Cloud Run Domain Mapping

Maps a custom domain directly onto a Cloud Run service — Cloud Run serves the domain itself and provisions/renews the TLS certificate, no load balancer required. The mapping emits the DNS records the domain's zone must publish as outputs, ready to wire into GcpDnsRecord or an external DNS host. It is the scale-appropriate path for "one service, one domain"; high-traffic and multi-service domains graduate to the load-balancer composition (serverless NEG → backend service → URL map → HTTPS proxy → forwarding rule).

The one prerequisite GCP enforces out-of-band: the deploying identity must have VERIFIED ownership of the domain (Search Console / `gcloud domains verify`) before the mapping can be created. Verification is one-time per domain; subdomains inherit it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Domain Mapping** -- a `google_cloud_run_domain_mapping` pointing the verified domain at the Cloud Run service, with a managed TLS certificate in the default `AUTOMATIC` mode
- **Cloud Run API enablement** -- `run.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A Cloud Run service** in the same region and project (reference a GcpCloudRun resource or name an existing service).
- **A verified domain**: the deploying identity must be a verified owner — GCP rejects the create otherwise.
- **IAM**: the deploying identity needs `roles/run.admin` or broader on the project.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Run Domain Mapping**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the region, the verified domain, and the target service reference. Start from the **Custom Domain** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunDomainMapping
metadata:
  name: app-domain
  org: acme-corp
  env: prod
spec:
  region: us-central1
  domain: app.example.com
  route:
    valueFrom:
      kind: GcpCloudRun
      name: my-service
      fieldPath: status.outputs.service_name
```

```shell
planton apply -f mapping.yaml
```

The mapping exists immediately; the domain starts serving once the `resource_records` output is published in the domain's DNS zone and the managed certificate issues (minutes after DNS propagates). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the mapping tracks the service and its DNS records flow into the zone:

```yaml
# On this GcpCloudRunDomainMapping — the route follows the deployed service:
spec:
  route:
    valueFrom:
      kind: GcpCloudRun
      name: my-web-app
      fieldPath: status.outputs.service_name

# On a GcpDnsRecord in the same chart — the CNAME value never leaves the graph:
spec:
  values:
    - valueFrom:
        kind: GcpCloudRunDomainMapping
        name: app-domain
        fieldPath: status.outputs.resource_records.[0].rrdata
```

The InfraPipeline deploys the service first, maps the domain, then publishes the DNS records.

## Key Configuration

These are the most important decisions when configuring a domain mapping. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Domain** -- the verified custom domain; it IS the mapping's name in GCP. Immutable, like every field here: the underlying resource is create-only end to end, so any change replaces the mapping (cheap — free object, seconds to re-create, brief serving gap while the certificate re-issues).

**Route** -- the Cloud Run service the domain routes to, by reference or literal name. Must exist, in this same region and project, before the mapping is created.

**Certificate mode** -- `AUTOMATIC` (the default: Cloud Run provisions and renews the certificate) or `NONE` (no managed certificate — for migrations where DNS must be published before issuance can succeed).

**Force override** -- unset, GCP fails the create with a conflict error when the domain is already mapped elsewhere; `true` steals the domain silently. Set it only after the conflict error confirmed the override is intended.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpCloudRun** | `route` | `status.outputs.service_name` |
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `domain` | The mapped domain (the mapping's name in GCP) | The join key consumers and verifiers address the mapping by |
| `resource_records` | DNS records the zone must publish (`record_type`/`record_name`/`rrdata` each) | GcpDnsRecord entries — the records never leave the graph |
| `mapped_route_name` | The service the mapping currently points to | Verifying the domain routes to the intended service |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Custom domain on a service** -- the default shape: managed certificate, DNS records wired into the zone. Start from the **Custom Domain** preset.

**Migration without a certificate gap** -- `certificateMode: NONE` first, publish DNS while the old host still serves, then flip to `AUTOMATIC` (a replacement — seconds). Start from the **Migration Without Certificate** preset.

## Works With

- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) — the service being mapped
- [**GCP DNS Record**](/infra-catalog/gcp-dns-record) — publishes the emitted records in a Cloud DNS zone
- [**GCP DNS Zone**](/infra-catalog/gcp-dns-zone) — the managed zone those records live in
- [**GCP Project**](/infra-catalog/gcp-project) — provides the project and API enablement
