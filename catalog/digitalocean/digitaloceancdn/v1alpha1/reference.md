# DigitalOceanCdn

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

DigitalOceanCdnSpec models the full digitalocean_cdn resource surface: a
CDN endpoint that serves a Spaces bucket's content from DigitalOcean's
global edge network, optionally under a custom subdomain with a managed
TLS certificate.

Certificates are referenced BY NAME, never by UUID: a Let's Encrypt
certificate's UUID rotates on every auto-renewal while its name stays
stable, so DigitalOcean addresses CDN certificates by name (the provider's
numeric certificate_id argument is deprecated for exactly this reason and
is deliberately not modeled here -- its update path can silently detach
the certificate when the custom domain changes).

## Example

```yaml
# Reference manifest for DigitalOceanCdn -- protovalidate-valid, embedded
# as the reference page's Example block, and the document the offline tofu
# plan renders. One document: origin plus an explicit cache TTL (the
# certificate/custom-domain pair needs a delegated domain, shown in the
# presets instead).
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanCdn
metadata:
  name: app-assets-cdn
spec:
  # Literal Space FQDN; use valueFrom to reference a DigitalOceanBucket
  # resource instead.
  origin:
    value: app-assets.nyc3.digitaloceanspaces.com
  ttl: 3600
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.origin` | `string \| valueFrom` | yes |  | DigitalOceanBucket (`status.outputs.bucket_domain_name`) |
| `spec.ttl` | `int32` |  |  |  |
| `spec.certificate` | `string \| valueFrom` |  |  | DigitalOceanCertificate (`status.outputs.certificate_id`) |
| `spec.customDomain` | `string` |  |  |  |

## Field Details

### spec.origin

`string | valueFrom` · required

The origin server -- the fully-qualified domain name of the Space whose
content the CDN serves (for example
"my-bucket.nyc3.digitaloceanspaces.com"). Use a literal FQDN or a
reference to a DigitalOceanBucket resource. Changing it replaces the
CDN endpoint.

- references: DigitalOceanBucket (`status.outputs.bucket_domain_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_domain_name}} -- a bare string does not parse

### spec.ttl

`int32` · optional (explicit presence)

(Optional) How long, in seconds, the edge caches content before
revalidating against the origin. When unset, DigitalOcean applies its
default of 3600 (one hour) and reports it back. The floor is 1 because
an explicit zero can never reach the API -- the provider drops zero
values on the way out, making "ttl = 0" indistinguishable from unset.

- rule: {"int32":{"gte":1}}

### spec.certificate

`string | valueFrom`

(Optional) The DigitalOcean-managed TLS certificate for the custom
domain, referenced by its stable NAME (which is also what
DigitalOceanCertificate exports as certificate_id). The literal value
"needs-cloudflare-cert" is a DigitalOcean sentinel that requests an
edge-provisioned certificate instead of an account-managed one.

- references: DigitalOceanCertificate (`status.outputs.certificate_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanCertificate, name: <that resource's name>, fieldPath: status.outputs.certificate_id}} -- a bare string does not parse

### spec.customDomain

`string`

(Optional) A fully-qualified custom subdomain to serve the CDN content
from (for example "assets.example.com"). Requires certificate: serving
a custom domain without TLS is not supported by DigitalOcean, and the
provider only surfaces the failure at apply time -- it is rejected at
validation time here instead.

- rule: custom_domain must be a fully-qualified domain name

## Validation Rules

- `spec.custom_domain_requires_certificate`: custom_domain requires certificate (the TLS certificate that will serve it)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanCdn, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cdn_id` | `string` | UUID of the CDN endpoint (the resource's API identity and its import id). |
| `status.outputs.endpoint` | `string` | The fully-qualified domain name the CDN serves content from (for example "my-bucket.nyc3.cdn.digitaloceanspaces.com"). Point CNAME records at this host when fronting it with a custom domain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.origin` | DigitalOceanBucket | `status.outputs.bucket_domain_name` |
| `spec.certificate` | DigitalOceanCertificate | `status.outputs.certificate_id` |

## See Also

- [Overview](../README.md)
