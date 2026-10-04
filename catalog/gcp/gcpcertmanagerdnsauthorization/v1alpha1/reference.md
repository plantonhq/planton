# GcpCertManagerDnsAuthorization

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCertManagerDnsAuthorizationSpec creates one Certificate Manager DNS
authorization: the proof-of-domain-control a Google-managed certificate
needs before it can be issued for a domain the load balancer is not yet
serving.

One authorization covers a single domain AND its wildcard — an
authorization for "example.com" issues certificates for both
"example.com" and "*.example.com".

The authorization exports the DNS validation record (a CNAME) that must
exist in the domain's zone. Compose it with a GcpDnsRecord: point the
record's name/type/values at this kind's dns_record_name /
dns_record_type / dns_record_data outputs and issuance validates
automatically — including BEFORE the certificate is created, which is
what makes zero-downtime certificate migration possible.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerDnsAuthorization
metadata:
  name: test-dns-authorization
spec:
  projectId:
    value: test-gcp-project
  domain: example.com
  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.authorizationName` | `string` |  |  |  |
| `spec.domain` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to create the authorization in.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.authorizationName

`string`

Name of the authorization in GCP. Must be 1-64 characters: start with
a letter, then letters, digits, hyphens, or underscores.
If not specified, defaults to metadata.name.
Immutable: renaming destroys and recreates the authorization.

- rule: authorization_name must be 1-64 characters, start with a letter, and contain only letters, digits, hyphens, or underscores

### spec.domain

`string` · required

The domain being authorized. Covers the domain itself and its
wildcard: authorizing "example.com" issues certificates for
"example.com" and "*.example.com". Immutable.
Bare domain only — no "*." prefix and no trailing dot.

- rule: domain must be a bare domain name (no wildcard prefix, no trailing dot), e.g. example.com
- rule: {"required":true}

### spec.description

`string`

Human-readable description of the authorization.

### spec.location

`string`

The Certificate Manager location. Defaults to "global" — the correct
choice for classic external HTTPS load balancers. Regional
authorizations pair with regional certificates only.
Immutable: changing the location destroys and recreates the
authorization.

### spec.type

`string`

How the validation record is scoped:
  FIXED_RECORD (default for global): the classic DNS-01 style record,
    one per (domain, authorization).
  PER_PROJECT_RECORD: one record per (domain, project) — lets multiple
    Certificate Manager resources across projects share the same
    validation record, and is the default for non-global locations.
If omitted, GCP picks the location-appropriate default. Immutable.

- rule: type must be FIXED_RECORD or PER_PROJECT_RECORD

### spec.labels

`map<string, string>`

User labels merged onto the authorization beneath the platform's
attribution labels (platform keys win on conflicts).
Keys/values: lowercase letters, digits, underscores, hyphens.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the authorization is deleted (certificates already
               issued against it keep serving; renewal for domains
               not yet serving through the LB stops validating)
  "PREVENT" -- destroy FAILS; protects the validation chain a
               certificate migration depends on
  "ABANDON" -- the authorization is removed from management but
               left in GCP, still validating its domain

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCertManagerDnsAuthorization, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.authorization_id` | `string` | Fully-qualified resource ID of the authorization (projects/{project}/locations/{location}/dnsAuthorizations/{name}). This is the exact value a GcpCertManagerCert's dns_authorizations list consumes — the composition key of the certificate family. |
| `status.outputs.authorization_name` | `string` | Name of the authorization as it exists in GCP. |
| `status.outputs.domain` | `string` | The authorized domain (covers the domain and its wildcard). |
| `status.outputs.dns_record_name` | `string` | The DNS validation record Cloud DNS must serve — compose these three fields into a GcpDnsRecord to complete domain validation: the record's fully-qualified name (e.g. "_acme-challenge.example.com."). |
| `status.outputs.dns_record_type` | `string` | The validation record's type (always "CNAME" today). |
| `status.outputs.dns_record_data` | `string` | The validation record's data — the value the CNAME must point at. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCertManagerCert | `spec.managed.dnsAuthorizations` | `status.outputs.authorization_id` |

## See Also

- [Overview](../README.md)
