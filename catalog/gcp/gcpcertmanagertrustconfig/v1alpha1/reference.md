# GcpCertManagerTrustConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCertManagerTrustConfigSpec creates one Certificate Manager trust config:
the set of certificate authorities a Google Cloud load balancer trusts when
it validates CLIENT certificates (mutual TLS), plus individual certificates
it accepts outright.

Where it is used: a server TLS policy names the trust config in its mTLS
client-validation settings, and the target HTTPS proxy of an external or
internal Application Load Balancer attaches that policy. A backend
authentication config uses a trust config the other way round, to validate
the certificates BACKENDS present. A trust config never attaches to a
certificate.

What it holds:
  - trust_stores: the root CAs (trust anchors) and intermediate CAs a
    client certificate chain must build up to. Google currently allows
    one trust store per config.
  - allowlisted_certificates: individual certificates accepted even when
    they do not chain to a trust anchor (self-signed device certificates,
    a partner's certificate), as long as the certificate parses, the client
    proves possession of its private key, and its SAN constraints hold.

The certificates are PUBLIC material -- no private key ever goes here --
so the fields are not secrets and each checks it holds a PEM certificate.
Google's provider still masks the trust-store certificates in plans, and
the Pulumi module keeps them as secrets in state to match.

Updates are in place: rotating a CA is a spec change, never a replacement.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerTrustConfig
metadata:
  name: partner-mtls-trust
spec:
  projectId:
    value: my-gcp-project
  description: CAs that sign partner client certificates
  # Empty location means global, which serves global load balancers.
  # location: us-central1
  trustStores:
    - trustAnchors:
        - |
          -----BEGIN CERTIFICATE-----
          <partner-root-ca-pem-body>
          -----END CERTIFICATE-----
  labels:
    team: edge
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.trustConfigName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.location` | `string` |  | `global` |  |
| `spec.trustStores` | `[]GcpCertManagerTrustConfigTrustStore` |  |  |  |
| `spec.trustStores[].trustAnchors` | `[]string` |  |  |  |
| `spec.trustStores[].intermediateCas` | `[]string` |  |  |  |
| `spec.allowlistedCertificates` | `[]string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to create the trust config in: a literal project ID or a
GcpProject reference. Empty uses the provider's default project.
Immutable: changing it recreates the trust config.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.trustConfigName

`string`

Name of the trust config in GCP: 1-64 characters, starting with a letter,
then letters, digits, hyphens, or underscores. Defaults to metadata.name.
Immutable.

- rule: trust_config_name must be 1-64 characters, start with a letter, and contain only letters, digits, hyphens, or underscores

### spec.description

`string`

Human-readable description.

### spec.location

`string`

The Certificate Manager location. Defaults to "global", which serves
global external and cross-region internal Application Load Balancers; a
regional load balancer needs a trust config in its own region (for
example "us-central1"). Immutable.

- default: `global`

### spec.trustStores

`[]GcpCertManagerTrustConfigTrustStore`

The trust stores client certificates are validated against. Google
allows exactly one today ("Only one TrustStore specified is currently
allowed" -- Certificate Manager API reference).

- rule: {"repeated":{"maxItems":"1"}}

### spec.trustStores[].trustAnchors

`[]string`

Root CA certificates, each PEM-encoded ("-----BEGIN CERTIFICATE-----").
A client chain is valid when it builds up to one of them.

- rule: {"repeated":{"items":{"string":{"minLen":"1","pattern":"-----BEGIN CERTIFICATE-----"}}}}

### spec.trustStores[].intermediateCas

`[]string`

Intermediate CA certificates, each PEM-encoded, used to build a chain
from a client certificate to a trust anchor when the client does not
send its intermediates.

- rule: {"repeated":{"items":{"string":{"minLen":"1","pattern":"-----BEGIN CERTIFICATE-----"}}}}

### spec.allowlistedCertificates

`[]string`

Certificates accepted regardless of chain, each a PEM-encoded X.509
certificate ("-----BEGIN CERTIFICATE-----..."). Use for self-signed
device certificates or a single partner certificate you do not want to
trust a whole CA for. A matching certificate must still parse, prove
possession of its private key, and satisfy its SAN constraints.

- rule: {"repeated":{"items":{"string":{"minLen":"1","pattern":"-----BEGIN CERTIFICATE-----"}}}}

### spec.labels

`map<string, string>`

User labels merged onto the trust config beneath the platform's
attribution labels (platform keys win on conflicts).

### spec.deletionPolicy

`string`

What happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the trust config is deleted (Google refuses while a TLS
               policy still references it)
  "PREVENT" -- destroy FAILS; a guard rail for a trust config live
               mutual-TLS traffic depends on
  "ABANDON" -- removed from management but left in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCertManagerTrustConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.trust_config_id` | `string` | Full resource name of the trust config (projects/{project}/locations/{location}/trustConfigs/{name}) -- the value a server TLS policy's mTLS client-validation trust config and a backend authentication config take. |
| `status.outputs.trust_config_name` | `string` | Name of the trust config as it exists in GCP. |
| `status.outputs.location` | `string` | The Certificate Manager location the trust config lives in ("global" unless a regional location was configured). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
