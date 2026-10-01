# GcpCertManagerIssuanceConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCertManagerIssuanceConfigSpec creates one Certificate Manager certificate
issuance config: the recipe Google follows to issue Google-managed
certificates from YOUR private CA (a Certificate Authority Service pool)
instead of a public CA.

A GcpCertManagerCert opts in by naming this config in
managed.issuance_config; many certificates share one config. Google then
requests every certificate from the pool, with the key algorithm and
lifetime set here, and renews it automatically when the rotation window
is reached -- private PKI with public-CA convenience, for internal load
balancers and service-to-service TLS that must chain to your own root.

Before it works: the pool needs an enabled certificate authority, and the
Certificate Manager service agent
(service-<project_number>@gcp-sa-certificatemanager.iam.gserviceaccount.com)
needs roles/privateca.certificateRequester on the pool. Google's docs
describe both; neither is created by this kind.

Every field except labels and deletion_policy is immutable: changing the
pool, key algorithm, lifetime, or rotation window replaces the config.
Certificates already issued keep serving until their own renewal.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerIssuanceConfig
metadata:
  name: internal-tls-issuance
spec:
  projectId:
    value: my-gcp-project
  description: Internal service certificates from the private pool
  # Empty location means global, which serves global certificates.
  # location: us-central1
  caPool:
    value: projects/my-gcp-project/locations/us-central1/caPools/internal-pool
  keyAlgorithm: ECDSA_P256
  lifetime: 2592000s
  rotationWindowPercentage: 66
  labels:
    team: platform
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.issuanceConfigName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.caPool` | `string \| valueFrom` | yes |  | GcpPrivateCaPool (`status.outputs.name`) |
| `spec.keyAlgorithm` | `string` | yes |  |  |
| `spec.lifetime` | `string` | yes |  |  |
| `spec.rotationWindowPercentage` | `int32` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to create the issuance config in: a literal project ID
or a GcpProject reference. Empty uses the provider's default project.
Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.issuanceConfigName

`string`

Name of the issuance config in GCP: 1-64 characters, starting with a
letter, then letters, digits, hyphens, or underscores. Defaults to
metadata.name. Immutable.

- rule: issuance_config_name must be 1-64 characters, start with a letter, and contain only letters, digits, hyphens, or underscores

### spec.description

`string`

Human-readable description. Immutable.

### spec.location

`string`

The Certificate Manager location. Empty means "global" (the provider's
default), which serves global certificates; a regional certificate
needs an issuance config in its own region. Immutable.

### spec.caPool

`string | valueFrom` · required

The Certificate Authority Service pool that issues the certificates, by
full name: projects/{project}/locations/{location}/caPools/{pool}.
Reference a GcpPrivateCaPool -- its `name` output is exactly this value.
The pool needs an enabled authority, and the Certificate Manager service
agent needs roles/privateca.certificateRequester on it. Immutable.

- references: GcpPrivateCaPool (`status.outputs.name`)
- rule: a literal ca_pool must be projects/{project}/locations/{location}/caPools/{pool}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPrivateCaPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.keyAlgorithm

`string` · required

Key algorithm for the private key Google generates for each
certificate: RSA_2048 (widest client compatibility) or ECDSA_P256
(smaller, faster handshakes; every modern client). Immutable.

- rule: {"required":true,"string":{"in":["RSA_2048","ECDSA_P256"]}}

### spec.lifetime

`string` · required

Lifetime of each issued certificate, as a duration in seconds with up
to nine fractional digits and an "s" suffix: from 21 days ("1814400s")
to 30 days ("2592000s"). Shorter lifetimes limit the damage of a leaked
key; renewal is automatic either way. Immutable.

- rule: lifetime must be a duration in seconds from 1814400s (21 days) to 2592000s (30 days), e.g. "2592000s"
- rule: {"required":true}

### spec.rotationWindowPercentage

`int32` · required

How far into a certificate's lifetime Google renews it, as a
percentage from 1 to 99. Google requires renewal at least 7 days after
issuance and at least 7 days before expiry, so the usable range depends
on the lifetime: 34-66 for 21 days, 24-76 for 30 days. Immutable.

- rule: {"required":true,"int32":{"lte":99,"gte":1}}

### spec.labels

`map<string, string>`

User labels merged onto the issuance config beneath the platform's
attribution labels (platform keys win on conflicts).

### spec.deletionPolicy

`string`

What happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the config is deleted (Google refuses while a
               certificate still references it)
  "PREVENT" -- destroy FAILS; a guard rail while certificates depend
               on it
  "ABANDON" -- removed from management but left in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.rotation_leaves_seven_days`: rotation_window_percentage must place renewal at least 7 days after issuance and at least 7 days before expiry (lifetime x percentage / 100 >= 7 days and lifetime x (100 - percentage) / 100 >= 7 days)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCertManagerIssuanceConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.issuance_config_id` | `string` | Full resource name of the issuance config (projects/{project}/locations/{location}/certificateIssuanceConfigs/{name}) -- the value a GcpCertManagerCert's managed.issuance_config takes. |
| `status.outputs.issuance_config_name` | `string` | Name of the issuance config as it exists in GCP. |
| `status.outputs.location` | `string` | The Certificate Manager location the issuance config lives in ("global" unless a regional location was configured). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.caPool` | GcpPrivateCaPool | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCertManagerCert | `spec.managed.issuanceConfig` | `status.outputs.issuance_config_id` |

## See Also

- [Overview](../README.md)
