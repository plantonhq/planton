# GcpPrivateCaCertificateTemplate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpPrivateCaCertificateTemplateSpec defines a certificate template
(`google_privateca_certificate_template`) -- a reusable shape for
certificates (a TLS server leaf, an mTLS client), shared across pools in
one project and location. A certificate references it to inherit its
predefined values and limits; callers need privateca.templateUser on it.
Values in the template conflict-fail against a pool's baseline values
rather than overriding them.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificateTemplate
metadata:
  name: tls-server
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  templateId: tls_server
  description: Internal TLS server certificates
  maximumLifetime: 2592000s
  predefinedValues:
    caOptions:
      isCa: false
    keyUsage:
      baseKeyUsage:
        digitalSignature: true
        keyEncipherment: true
      extendedKeyUsage:
        serverAuth: true
    aiaOcspServers:
      - http://ocsp.internal.example.com
  identityConstraints:
    allowSubjectPassthrough: true
    allowSubjectAltNamesPassthrough: true
    celExpression:
      title: Internal names
      expression: subject_alt_names.all(san, san.type == DNS && san.value.endsWith(".internal.example.com"))
  passthroughExtensions:
    knownExtensions:
      - BASE_KEY_USAGE
      - EXTENDED_KEY_USAGE
  labels:
    use: tls
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.templateId` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.maximumLifetime` | `string` |  |  |  |
| `spec.predefinedValues` | `GcpPrivateCaCertificateTemplateX509Parameters` |  |  |  |
| `spec.predefinedValues.keyUsage` | `GcpPrivateCaCertificateTemplateKeyUsage` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage` | `GcpPrivateCaCertificateTemplateBaseKeyUsage` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.digitalSignature` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.contentCommitment` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.keyEncipherment` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.dataEncipherment` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.keyAgreement` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.certSign` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.crlSign` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.encipherOnly` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.baseKeyUsage.decipherOnly` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.extendedKeyUsage` | `GcpPrivateCaCertificateTemplateExtendedKeyUsage` |  |  |  |
| `spec.predefinedValues.keyUsage.extendedKeyUsage.serverAuth` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.extendedKeyUsage.clientAuth` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.extendedKeyUsage.codeSigning` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.extendedKeyUsage.emailProtection` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.extendedKeyUsage.timeStamping` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.extendedKeyUsage.ocspSigning` | `bool` |  |  |  |
| `spec.predefinedValues.keyUsage.unknownExtendedKeyUsages` | `[]GcpPrivateCaCertificateTemplateObjectId` |  |  |  |
| `spec.predefinedValues.keyUsage.unknownExtendedKeyUsages[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.predefinedValues.caOptions` | `GcpPrivateCaCertificateTemplateCaOptions` |  |  |  |
| `spec.predefinedValues.caOptions.isCa` | `bool` |  |  |  |
| `spec.predefinedValues.caOptions.maxIssuerPathLength` | `int32` |  |  |  |
| `spec.predefinedValues.policyIds` | `[]GcpPrivateCaCertificateTemplateObjectId` |  |  |  |
| `spec.predefinedValues.policyIds[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.predefinedValues.aiaOcspServers` | `[]string` |  |  |  |
| `spec.predefinedValues.additionalExtensions` | `[]GcpPrivateCaCertificateTemplateX509Extension` |  |  |  |
| `spec.predefinedValues.additionalExtensions[].objectId` | `GcpPrivateCaCertificateTemplateObjectId` | yes |  |  |
| `spec.predefinedValues.additionalExtensions[].objectId.objectIdPath` | `[]int32` | yes |  |  |
| `spec.predefinedValues.additionalExtensions[].critical` | `bool` |  |  |  |
| `spec.predefinedValues.additionalExtensions[].value` | `string` | yes |  |  |
| `spec.predefinedValues.nameConstraints` | `GcpPrivateCaCertificateTemplateNameConstraints` |  |  |  |
| `spec.predefinedValues.nameConstraints.critical` | `bool` |  |  |  |
| `spec.predefinedValues.nameConstraints.permittedDnsNames` | `[]string` |  |  |  |
| `spec.predefinedValues.nameConstraints.excludedDnsNames` | `[]string` |  |  |  |
| `spec.predefinedValues.nameConstraints.permittedIpRanges` | `[]string` |  |  |  |
| `spec.predefinedValues.nameConstraints.excludedIpRanges` | `[]string` |  |  |  |
| `spec.predefinedValues.nameConstraints.permittedEmailAddresses` | `[]string` |  |  |  |
| `spec.predefinedValues.nameConstraints.excludedEmailAddresses` | `[]string` |  |  |  |
| `spec.predefinedValues.nameConstraints.permittedUris` | `[]string` |  |  |  |
| `spec.predefinedValues.nameConstraints.excludedUris` | `[]string` |  |  |  |
| `spec.identityConstraints` | `GcpPrivateCaCertificateTemplateIdentityConstraints` |  |  |  |
| `spec.identityConstraints.allowSubjectPassthrough` | `bool` |  |  |  |
| `spec.identityConstraints.allowSubjectAltNamesPassthrough` | `bool` |  |  |  |
| `spec.identityConstraints.celExpression` | `GcpPrivateCaCertificateTemplateCelExpression` |  |  |  |
| `spec.identityConstraints.celExpression.expression` | `string` |  |  |  |
| `spec.identityConstraints.celExpression.title` | `string` |  |  |  |
| `spec.identityConstraints.celExpression.description` | `string` |  |  |  |
| `spec.identityConstraints.celExpression.location` | `string` |  |  |  |
| `spec.passthroughExtensions` | `GcpPrivateCaCertificateTemplatePassthroughExtensions` |  |  |  |
| `spec.passthroughExtensions.knownExtensions` | `[]string` |  |  |  |
| `spec.passthroughExtensions.additionalExtensions` | `[]GcpPrivateCaCertificateTemplateObjectId` |  |  |  |
| `spec.passthroughExtensions.additionalExtensions[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the template lives in: a literal project ID or a GcpProject reference.
If omitted, the provider's default project is used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The CA Service region, e.g. "us-central1"; certificates using the template must be in the same region. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.templateId

`string`

The template's ID, unique in its parent: letters, digits, "-" and "_",
at most 63 characters. Defaults to metadata.name. Immutable.

- rule: template_id must be 1-63 letters, digits, '-' or '_'

### spec.description

`string`

What the template is for, shown to people choosing one.

### spec.maximumLifetime

`string`

The longest lifetime a certificate issued with the template may have,
e.g. 2592000s (30 days). The pool's own maximum still applies; the
shorter wins.

- rule: maximum_lifetime must be a duration in seconds with an 's' suffix, e.g. 86400s

### spec.predefinedValues

`GcpPrivateCaCertificateTemplateX509Parameters`

X.509 values stamped onto every certificate issued with the template.

### spec.predefinedValues.keyUsage

`GcpPrivateCaCertificateTemplateKeyUsage`

What the certificate's key may be used for (the key usage and extended
key usage extensions). Omit for no key-usage statement; both usage
groups are optional and an empty group states nothing.

### spec.predefinedValues.keyUsage.baseKeyUsage

`GcpPrivateCaCertificateTemplateBaseKeyUsage`

The key usage extension's bits.

### spec.predefinedValues.keyUsage.baseKeyUsage.digitalSignature

`bool`

The key may make digital signatures other than certificate and CRL
signatures (TLS handshakes, signed tokens).

### spec.predefinedValues.keyUsage.baseKeyUsage.contentCommitment

`bool`

The key may protect content against the signer later denying it
(non-repudiation).

### spec.predefinedValues.keyUsage.baseKeyUsage.keyEncipherment

`bool`

The key may encipher other keys (RSA key transport in TLS).

### spec.predefinedValues.keyUsage.baseKeyUsage.dataEncipherment

`bool`

The key may encipher raw data directly.

### spec.predefinedValues.keyUsage.baseKeyUsage.keyAgreement

`bool`

The key may be used in key agreement (ECDH).

### spec.predefinedValues.keyUsage.baseKeyUsage.certSign

`bool`

The key may sign certificates -- set on every CA certificate.

### spec.predefinedValues.keyUsage.baseKeyUsage.crlSign

`bool`

The key may sign certificate revocation lists -- set on every CA
certificate that publishes CRLs.

### spec.predefinedValues.keyUsage.baseKeyUsage.encipherOnly

`bool`

With key_agreement, the key may only encipher.

### spec.predefinedValues.keyUsage.baseKeyUsage.decipherOnly

`bool`

With key_agreement, the key may only decipher.

### spec.predefinedValues.keyUsage.extendedKeyUsage

`GcpPrivateCaCertificateTemplateExtendedKeyUsage`

The extended key usage extension's well-known purposes.

### spec.predefinedValues.keyUsage.extendedKeyUsage.serverAuth

`bool`

TLS server authentication -- what an HTTPS or gRPC server certificate
needs.

### spec.predefinedValues.keyUsage.extendedKeyUsage.clientAuth

`bool`

TLS client authentication -- what an mTLS client certificate needs.

### spec.predefinedValues.keyUsage.extendedKeyUsage.codeSigning

`bool`

Code signing.

### spec.predefinedValues.keyUsage.extendedKeyUsage.emailProtection

`bool`

S/MIME email protection.

### spec.predefinedValues.keyUsage.extendedKeyUsage.timeStamping

`bool`

Trusted timestamping.

### spec.predefinedValues.keyUsage.extendedKeyUsage.ocspSigning

`bool`

Signing OCSP responses.

### spec.predefinedValues.keyUsage.unknownExtendedKeyUsages

`[]GcpPrivateCaCertificateTemplateObjectId`

Extended key usages with no named field here, each an OID path.

### spec.predefinedValues.keyUsage.unknownExtendedKeyUsages[].objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.predefinedValues.caOptions

`GcpPrivateCaCertificateTemplateCaOptions`

The basic constraints extension: whether the certificate is a CA and
how many CA levels may sit below it. Omit to leave the extension to
Google's default (for a leaf certificate, is_ca false).

### spec.predefinedValues.caOptions.isCa

`bool` · optional (explicit presence)

true makes a CA certificate (CA:TRUE), false states CA:FALSE; unset
leaves the CA flag out of the extension.

### spec.predefinedValues.caOptions.maxIssuerPathLength

`int32` · optional (explicit presence)

The path length constraint: how many CA levels may sit below this
certificate. 0 means it may sign only leaf certificates -- a
subordinate that issues workload certificates; unset means no limit
is stated.

- rule: {"int32":{"gte":0}}

### spec.predefinedValues.policyIds

`[]GcpPrivateCaCertificateTemplateObjectId`

Certificate policy object identifiers (RFC 5280 section 4.2.1.4), each
an OID path such as [2, 23, 140, 1, 2, 1] (CA/Browser Forum
domain-validated).

### spec.predefinedValues.policyIds[].objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.predefinedValues.aiaOcspServers

`[]string`

OCSP responder URLs placed in the Authority Information Access
extension, e.g. "http://ocsp.example.com". Google runs no OCSP
responder; list your own.

### spec.predefinedValues.additionalExtensions

`[]GcpPrivateCaCertificateTemplateX509Extension`

Custom X.509 extensions, each an OID, a base64 DER value, and whether
a relying party that does not understand it must reject the
certificate (critical).

### spec.predefinedValues.additionalExtensions[].objectId

`GcpPrivateCaCertificateTemplateObjectId` · required

The extension's OID.

- rule: {"required":true}

### spec.predefinedValues.additionalExtensions[].objectId.objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.predefinedValues.additionalExtensions[].critical

`bool`

A relying party that does not understand a critical extension must
reject the certificate. Mark an extension critical only when every
consumer knows it.

### spec.predefinedValues.additionalExtensions[].value

`string` · required

The extension's DER value, base64-encoded.

- rule: {"string":{"minLen":"1"}}

### spec.predefinedValues.nameConstraints

`GcpPrivateCaCertificateTemplateNameConstraints`

The name constraints extension (RFC 5280 section 4.2.1.10): the names
certificates below a CA may and may not carry. Meaningful on CA
certificates.

### spec.predefinedValues.nameConstraints.critical

`bool`

Whether the constraints are marked critical. RFC 5280 requires true on
a CA certificate that carries them.

### spec.predefinedValues.nameConstraints.permittedDnsNames

`[]string`

DNS names certificates below this CA may carry.

### spec.predefinedValues.nameConstraints.excludedDnsNames

`[]string`

DNS names certificates below this CA must not carry.

### spec.predefinedValues.nameConstraints.permittedIpRanges

`[]string`

IP ranges (CIDR) certificates below this CA may carry.

### spec.predefinedValues.nameConstraints.excludedIpRanges

`[]string`

IP ranges (CIDR) certificates below this CA must not carry.

### spec.predefinedValues.nameConstraints.permittedEmailAddresses

`[]string`

Email addresses, hosts, or ".domain" suffixes that may appear.

### spec.predefinedValues.nameConstraints.excludedEmailAddresses

`[]string`

Email addresses, hosts, or ".domain" suffixes that must not appear.

### spec.predefinedValues.nameConstraints.permittedUris

`[]string`

URI hosts or ".domain" suffixes that may appear.

### spec.predefinedValues.nameConstraints.excludedUris

`[]string`

URI hosts or ".domain" suffixes that must not appear.

### spec.identityConstraints

`GcpPrivateCaCertificateTemplateIdentityConstraints`

Limits on the identities certificates issued with it may carry. Omit
for no limits.

### spec.identityConstraints.allowSubjectPassthrough

`bool`

Copy the Subject from the certificate request into the certificate;
false discards the requested Subject.

### spec.identityConstraints.allowSubjectAltNamesPassthrough

`bool`

Copy the subject alternative names from the request; false discards
them.

### spec.identityConstraints.celExpression

`GcpPrivateCaCertificateTemplateCelExpression`

A CEL expression over the resolved subject and subject alternative
names that must hold before a certificate is signed, e.g.
subject_alt_names.all(san, san.type == DNS && san.value.endsWith(".internal.example.com")).

### spec.identityConstraints.celExpression.expression

`string`

The expression.

### spec.identityConstraints.celExpression.title

`string`

A short title for the expression.

### spec.identityConstraints.celExpression.description

`string`

What the expression checks.

### spec.identityConstraints.celExpression.location

`string`

Where the expression came from, for error messages (a file and line).

### spec.passthroughExtensions

`GcpPrivateCaCertificateTemplatePassthroughExtensions`

Extensions a request may pass through the template.

### spec.passthroughExtensions.knownExtensions

`[]string`

Named extensions: BASE_KEY_USAGE, EXTENDED_KEY_USAGE, CA_OPTIONS,
POLICY_IDS, AIA_OCSP_SERVERS, NAME_CONSTRAINTS.

- rule: {"repeated":{"items":{"string":{"in":["BASE_KEY_USAGE","EXTENDED_KEY_USAGE","CA_OPTIONS","POLICY_IDS","AIA_OCSP_SERVERS","NAME_CONSTRAINTS"]}}}}

### spec.passthroughExtensions.additionalExtensions

`[]GcpPrivateCaCertificateTemplateObjectId`

Custom extensions, by OID.

### spec.passthroughExtensions.additionalExtensions[].objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.labels

`map<string, string>`

Labels on the template. The platform attribution labels are added on
top and win on a key conflict.

### spec.deletionPolicy

`string`

What happens to the template when this resource is destroyed:
  "" / "DELETE" -- deleted (the provider's default)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPrivateCaCertificateTemplate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/certificateTemplates/{template_id}. |
| `status.outputs.template_id` | `string` | The template's ID (the last segment of name). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpPrivateCaCertificate | `spec.certificateTemplate` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
