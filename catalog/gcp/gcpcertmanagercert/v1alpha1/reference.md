# GcpCertManagerCert

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCertManagerCertSpec creates one Certificate Manager certificate — the
modern certificate resource external Application Load Balancers consume
via a target HTTPS proxy's certificate_manager_certificates list or a
certificate map.

Exactly one of two arms must be configured:

  - **managed**: Google provisions and RENEWS the certificate
    automatically for the listed domains. Domain control is proven
    either through DNS authorizations (first-class
    GcpCertManagerDnsAuthorization resources — required for wildcard
    domains and for issuing before traffic is serving), through a
    private-PKI issuance config, or — when neither is set — through
    load-balancer authorization (GCP validates via the serving load
    balancer itself once traffic reaches it).

  - **self_managed**: you upload a PEM certificate chain and its private
    key. Renewal before expiry is your responsibility; rotation is a
    spec update with the new material.

(The classic compute certificates are separate kinds:
GcpManagedSslCertificate for Google-managed classic certificates and
GcpSslCertificate for self-managed classic certificates.)

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerCert
metadata:
  name: test-certificate
spec:
  projectId:
    value: test-gcp-project
  managed:
    domains:
      - example.com
    dnsAuthorizations:
      - value: projects/test-gcp-project/locations/global/dnsAuthorizations/example-com-auth
  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.certName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.scope` | `string` |  |  |  |
| `spec.managed` | `GcpCertManagerCertManaged` |  |  |  |
| `spec.managed.domains` | `[]string` | yes |  |  |
| `spec.managed.dnsAuthorizations` | `[]string \| valueFrom` |  |  | GcpCertManagerDnsAuthorization (`status.outputs.authorization_id`) |
| `spec.managed.issuanceConfig` | `string \| valueFrom` |  |  | GcpCertManagerIssuanceConfig (`status.outputs.issuance_config_id`) |
| `spec.selfManaged` | `GcpCertManagerCertSelfManaged` |  |  |  |
| `spec.selfManaged.pemCertificate` | `string` | yes |  |  |
| `spec.selfManaged.pemPrivateKey` | `string` (sensitive) | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to create the certificate in.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing the project destroys and recreates the certificate.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.certName

`string`

Name of the certificate in GCP. Must be 1-64 characters: start with a
letter, then letters, digits, hyphens, or underscores. Certificate
names must be unique per location. Immutable.
If not specified, defaults to metadata.name.

- rule: cert_name must be 1-64 characters, start with a letter, and contain only letters, digits, hyphens, or underscores

### spec.description

`string`

Human-readable description of the certificate.

### spec.location

`string`

The Certificate Manager location. Defaults to "global" — the correct
choice for classic external HTTPS load balancers. Regional
certificates serve regional load balancers only. Immutable.

### spec.scope

`string`

Where the certificate is served from. Immutable.
  DEFAULT: core Google data centers (choose this if unsure).
  EDGE_CACHE: Edge Points of Presence (Media CDN).
  ALL_REGIONS: every GCP region (global certificates only —
    cross-region internal Application Load Balancers).
  CLIENT_AUTH: presented BY the load balancer to the backend when
    backend mTLS is configured.

- rule: scope must be DEFAULT, EDGE_CACHE, ALL_REGIONS, or CLIENT_AUTH

### spec.managed

`GcpCertManagerCertManaged`

Google-managed arm: provisioned and renewed automatically.

- rule: dns_authorizations and issuance_config are mutually exclusive (omit both for load-balancer authorization)
- rule: wildcard domains require dns_authorizations (load-balancer authorization cannot validate wildcards)

### spec.managed.domains

`[]string` · required

The domains this certificate covers (e.g. "example.com",
"*.example.com"). Wildcards are supported only with DNS authorizations
or an issuance config. Immutable.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"domain.valid","message":"each domain must be a bare or wildcard domain name (no trailing dot), e.g. example.com or *.example.com","expression":"this.matches('^([*][.])?(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?[.])+[a-z]{2,}$')"}]}}}

### spec.managed.dnsAuthorizations

`[]string | valueFrom`

DNS authorizations proving control of the domains — one per distinct
domain (an authorization covers its domain and that domain's
wildcard). Each entry is the authorization's fully-qualified resource
ID; reference GcpCertManagerDnsAuthorization resources.
Omit (with no issuance_config) for load-balancer authorization.

- references: GcpCertManagerDnsAuthorization (`status.outputs.authorization_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCertManagerDnsAuthorization, name: <that resource's name>, fieldPath: status.outputs.authorization_id}} -- a bare string does not parse

### spec.managed.issuanceConfig

`string | valueFrom`

Private-PKI issuance: the certificate issuance config that has your own
Certificate Authority Service pool sign this certificate instead of a
public CA, by full name
(projects/{project}/locations/{location}/certificateIssuanceConfigs/{name}).
Reference a GcpCertManagerIssuanceConfig -- its `issuance_config_id`
output is exactly this value, and the reference orders the certificate
after the config. Mutually exclusive with dns_authorizations. Immutable.

- references: GcpCertManagerIssuanceConfig (`status.outputs.issuance_config_id`)
- rule: a literal issuance_config must be projects/{project}/locations/{location}/certificateIssuanceConfigs/{name}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCertManagerIssuanceConfig, name: <that resource's name>, fieldPath: status.outputs.issuance_config_id}} -- a bare string does not parse

### spec.selfManaged

`GcpCertManagerCertSelfManaged`

Self-managed arm: bring-your-own PEM certificate and key.

### spec.selfManaged.pemCertificate

`string` · required

The certificate chain in PEM form: leaf certificate first, followed by
any intermediates. Certificate material is public — only the key is
secret. Updating the pair in place rotates the certificate.

- rule: pem_certificate must be a PEM CERTIFICATE block (did you swap the certificate and private key?)
- rule: {"required":true}

### spec.selfManaged.pemPrivateKey

`string` · required · sensitive

The leaf certificate's private key in PEM form — a PRIVATE KEY block
(take care not to swap the certificate and the key). Secret material —
never logged, masked in outputs. The PEM framing is taught here rather
than enforced by a validation rule, because sensitive fields hold a
managed-secret reference on consuming platforms and a content-shape
rule would reject every reference.

- rule: {"required":true}

### spec.labels

`map<string, string>`

User labels merged onto the certificate beneath the platform's
attribution labels (platform keys win on conflicts).
Keys/values: lowercase letters, digits, underscores, hyphens.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the certificate is deleted (GCP refuses while a proxy
               or certificate map still references it)
  "PREVENT" -- destroy FAILS; a guard rail for a certificate whose
               replacement is not yet serving
  "ABANDON" -- the certificate is removed from management but left
               in GCP (useful mid-rotation handoff)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.exactly_one_arm`: exactly one of managed or self_managed must be set
- `spec.all_regions_is_global`: scope ALL_REGIONS is only valid for global certificates (leave location empty or 'global')

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCertManagerCert, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.certificate_id` | `string` | Fully-qualified resource ID of the certificate (projects/{project}/locations/{location}/certificates/{name}). |
| `status.outputs.certificate_name` | `string` | Name of the certificate as it exists in GCP. This is the value a GcpTargetHttpsProxy's certificate_manager_certificates list consumes. |
| `status.outputs.san_dnsnames` | `[]string` | The Subject Alternative Names (dnsName type) present in the issued certificate. |
| `status.outputs.location` | `string` | The Certificate Manager location the certificate lives in ("global" unless a regional location was configured). |
| `status.outputs.managed_state` | `string` | State of a managed certificate (PROVISIONING, FAILED, ACTIVE); empty for self-managed certificates. A managed certificate stays PROVISIONING until domain validation completes. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.managed.dnsAuthorizations` | GcpCertManagerDnsAuthorization | `status.outputs.authorization_id` |
| `spec.managed.issuanceConfig` | GcpCertManagerIssuanceConfig | `status.outputs.issuance_config_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCertificateMap | `spec.entries[].certificates` | `status.outputs.certificate_id` |
| GcpTargetHttpsProxy | `spec.certificateManagerCertificates` | `status.outputs.certificate_name` |

## See Also

- [Overview](../README.md)
