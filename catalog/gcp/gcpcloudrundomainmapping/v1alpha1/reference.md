# GcpCloudRunDomainMapping

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudRunDomainMappingSpec maps a custom domain directly onto a Cloud
Run service (`google_cloud_run_domain_mapping`) — the scale-appropriate
alternative to fronting the service with a global HTTPS load balancer.
Cloud Run serves the domain itself and (by default) provisions and
renews the TLS certificate; all you add is the DNS records this
resource emits as outputs (wire them into GcpDnsRecord, or publish them
at an external DNS host).

Two truths shape every field here:

 1. THE DOMAIN MUST BE VERIFIED FIRST. GCP refuses to map a domain the
    provisioning identity has not proven ownership of (Google Search
    Console / `gcloud domains verify`). Verification is a one-time,
    out-of-band step per domain — no Terraform or Pulumi resource
    performs it.
 2. THE MAPPING IS IMMUTABLE. Every argument on the underlying resource
    is create-only: changing ANY spec field (the domain, the target
    service, the certificate mode, even a label) replaces the mapping.
    Replacement is cheap — the object is free and re-creates in
    seconds — but expect a brief serving gap while the new mapping
    re-issues its certificate.

The production-grade path for high-traffic or multi-service domains
remains the load-balancer composition (serverless NEG → backend
service → URL map → HTTPS proxy → forwarding rule); this kind covers
the "one service, one domain, no LB" story.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunDomainMapping
metadata:
  name: my-app-domain
spec:
  # GCP project that owns the mapping. Omit to use the provider's
  # default project.
  projectId:
    value: my-gcp-project-123

  # Region of the Cloud Run service being mapped. Mappings are regional
  # and must live in the SAME region as their target service. Immutable —
  # like every field here except deletionPolicy: the underlying resource
  # is create-only end to end, so any change replaces the mapping (cheap:
  # free object, re-creates in seconds, brief serving gap while the
  # certificate re-issues).
  region: us-central1

  # The custom domain to map — this IS the mapping's name in GCP. MUST
  # already be verified by the deploying identity (Search Console /
  # `gcloud domains verify`); GCP rejects the create otherwise.
  # Subdomains of a verified domain need no separate verification.
  domain: app.example.com

  # The Cloud Run service this domain routes to. Reference a GcpCloudRun
  # resource's service_name output (shown here against the published E2E
  # fixture) or provide the service name literally. The service must
  # exist, in this same region and project, before the mapping.
  route:
    valueFrom:
      kind: GcpCloudRun
      name: planton-oss-e2e-cldrun-minimal
      fieldPath: status.outputs.service_name

  # AUTOMATIC (default): Cloud Run provisions and renews the TLS
  # certificate once the emitted DNS records are published. NONE skips
  # the managed certificate (migration shape — see the presets).
  certificateMode: AUTOMATIC

  # Leave false for the safe behavior: GCP fails the create with a
  # conflict error when the domain is already mapped elsewhere, instead
  # of silently stealing it.
  forceOverride: false

  # Labels and annotations stored on the mapping object (non-authoritative
  # Knative metadata; the Cloud Run API adds server-side annotations of
  # its own — those never show as drift).
  labels:
    team: payments
  annotations:
    note: primary customer-facing domain

  # DELETE (default): destroying the resource deletes the mapping — the
  # domain stops routing; the service is untouched. PREVENT refuses the
  # destroy; ABANDON keeps the mapping serving but drops it from
  # management.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.domain` | `string` | yes |  |  |
| `spec.route` | `string \| valueFrom` | yes |  | GcpCloudRun (`status.outputs.service_name`) |
| `spec.certificateMode` | `string` |  | `AUTOMATIC` |  |
| `spec.forceOverride` | `bool` |  |  |  |
| `spec.namespace` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the domain mapping. Can be a literal
project ID or a reference to a GcpProject resource. If omitted, the
provider's default project is used. The domain must be verified by
an identity with access to this project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

GCP region of the Cloud Run service being mapped (e.g. us-central1).
Domain mappings are regional and must be created in the SAME region
as their target service. Immutable: changing it replaces the mapping.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.domain

`string` · required

The custom domain to map — this IS the mapping's name in GCP (e.g.
"app.example.com"). MUST already be verified by the provisioning
identity (Search Console / `gcloud domains verify`); GCP rejects the
create otherwise. Subdomains of a verified domain need no separate
verification. Immutable: changing it replaces the mapping.

- rule: {"required":true,"string":{"maxLen":"253","pattern":"^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$"}}

### spec.route

`string | valueFrom` · required

The Cloud Run service this domain routes to. Reference a GcpCloudRun
resource or provide the service name literally. The service must
exist, in this same region and project, before the mapping is
created. Immutable: repointing the domain at a different service
replaces the mapping.

- references: GcpCloudRun (`status.outputs.service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudRun, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.certificateMode

`string`

How the domain's TLS certificate is provided:
  "AUTOMATIC" -- Cloud Run provisions and renews a managed
                 certificate (the default, and what almost every
                 mapping wants)
  "NONE"      -- no managed certificate; the domain serves without
                 TLS until one exists (used for migrations where the
                 DNS records must be published before certificate
                 issuance can succeed)
Immutable: changing it replaces the mapping.

- default: `AUTOMATIC`
- rule: certificate_mode must be AUTOMATIC (managed certificate) or NONE (no managed certificate)

### spec.forceOverride

`bool`

When true, this mapping overrides any existing mapping of the same
domain without warning. Leave unset for the safe behavior: GCP then
fails the create with a conflict error instead of silently stealing
a domain another mapping already serves. Set it only after such a
conflict error confirmed the override is intended. Immutable.

### spec.namespace

`string`

The Cloud Run namespace for the mapping — GCP requires it to equal
the project ID or the project NUMBER. Leave empty for the sensible
default (the module uses the project ID); set it only when a
numbered-namespace convention requires the project number instead.
Immutable.

### spec.labels

`map<string, string>`

Labels stored on the mapping object (Knative-style metadata labels,
grouped with the platform's own resource labels). Non-authoritative:
the module manages only the labels declared here plus the platform
set. Immutable: changing them replaces the mapping.

### spec.annotations

`map<string, string>`

Annotations stored on the mapping object. Non-authoritative, and the
Cloud Run API adds server-side annotations of its own (those never
show up as drift — the module manages only the entries declared
here). Immutable: changing them replaces the mapping.

### spec.deletionPolicy

`string`

Deletion policy for the domain mapping:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- destroying the resource deletes the mapping; the
               domain simply stops routing (the service is untouched)
  "PREVENT" -- destroy FAILS; protects a mapping production traffic
               depends on
  "ABANDON" -- the mapping is removed from management but keeps
               serving in GCP
This is the one field that updates in place (it never touches the
API object).

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudRunDomainMapping, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.domain` | `string` | The mapped domain (the mapping's name in GCP) — the join key consumers and verifiers address the mapping by. |
| `status.outputs.region` | `string` | GCP region the mapping lives in (plain region name). |
| `status.outputs.resource_records` | `[]GcpCloudRunDomainMappingResourceRecord` | The DNS records GCP requires the domain's zone to publish before the domain resolves to the service: a root domain receives A/AAAA records, a subdomain one CNAME (ghs.googlehosted.com.). Wire these into GcpDnsRecord (or your external DNS host) — until they are published, the mapping exists but the domain does not serve and a managed certificate cannot be issued. |
| `status.outputs.resource_records[].record_type` | `string` | DNS record type: "A", "AAAA", or "CNAME". |
| `status.outputs.resource_records[].record_name` | `string` | Relative name of the record within the zone. Only set for CNAME records (e.g. "www"); empty for root-domain A/AAAA records. |
| `status.outputs.resource_records[].rrdata` | `string` | The record's value (RFC 1035 rrdata) — an IP address for A/AAAA, "ghs.googlehosted.com." for CNAME. |
| `status.outputs.mapped_route_name` | `string` | The Cloud Run route (service) the mapping currently points to, as reported by GCP. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.route` | GcpCloudRun | `status.outputs.service_name` |

## See Also

- [Overview](../README.md)
