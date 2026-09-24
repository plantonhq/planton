# GcpPrivateCaCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpPrivateCaCertificateSpec defines a certificate issued from a CA pool
(`google_privateca_certificate`) -- one signed X.509 certificate for a
key someone else holds, described by a CSR or by structured config.
Everything but labels is immutable: a change issues a new certificate.
Destroy REVOKES the certificate (Google keeps the revoked record); its
ID cannot be reused in the pool.

The pool must be ENTERPRISE: Google cannot describe or revoke
certificates in a DEVOPS pool, so neither engine could track or destroy
one. Workloads that mint many short-lived certificates call the API
directly instead.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificate
metadata:
  name: api-server
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  pool:
    value: projects/my-gcp-project/locations/us-central1/caPools/internal-servers
  certificateId: api-server-2026-09
  certificateAuthority:
    value: issuing-ca
  certificateTemplate:
    value: projects/my-gcp-project/locations/us-central1/certificateTemplates/tls-server
  lifetime: 2592000s
  config:
    subjectConfig:
      subject:
        commonName: api.internal.example.com
      subjectAltName:
        dnsNames:
          - api.internal.example.com
    x509Config:
      caOptions:
        isCa: false
      keyUsage:
        baseKeyUsage:
          digitalSignature: true
          keyEncipherment: true
        extendedKeyUsage:
          serverAuth: true
    publicKey:
      key: LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFcy9jMjdMT3BJdmJWU0RzalMvTWp6VDM3bk5IeQp1dUtmNVZ0eVE2SmNUTUltOEQyRzl4MmNRVXZteFdjZzFFVzQrbXpwZ0RKOXdWR1dDeitQclUrejNBPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==
  labels:
    app: api
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.pool` | `string \| valueFrom` | yes |  | GcpPrivateCaPool (`status.outputs.name`) |
| `spec.certificateId` | `string` |  |  |  |
| `spec.certificateAuthority` | `string \| valueFrom` |  |  | GcpPrivateCaCertificateAuthority (`status.outputs.name`) |
| `spec.certificateTemplate` | `string \| valueFrom` |  |  | GcpPrivateCaCertificateTemplate (`status.outputs.name`) |
| `spec.lifetime` | `string` |  |  |  |
| `spec.pemCsr` | `string` |  |  |  |
| `spec.config` | `GcpPrivateCaCertificateConfig` |  |  |  |
| `spec.config.subjectConfig` | `GcpPrivateCaCertificateSubjectConfig` | yes |  |  |
| `spec.config.subjectConfig.subject` | `GcpPrivateCaCertificateSubject` | yes |  |  |
| `spec.config.subjectConfig.subject.commonName` | `string` | yes |  |  |
| `spec.config.subjectConfig.subject.countryCode` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.organization` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.organizationalUnit` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.locality` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.province` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.streetAddress` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.postalCode` | `string` |  |  |  |
| `spec.config.subjectConfig.subjectAltName` | `GcpPrivateCaCertificateSubjectAltName` |  |  |  |
| `spec.config.subjectConfig.subjectAltName.dnsNames` | `[]string` |  |  |  |
| `spec.config.subjectConfig.subjectAltName.uris` | `[]string` |  |  |  |
| `spec.config.subjectConfig.subjectAltName.emailAddresses` | `[]string` |  |  |  |
| `spec.config.subjectConfig.subjectAltName.ipAddresses` | `[]string` |  |  |  |
| `spec.config.subjectKeyId` | `string` |  |  |  |
| `spec.config.x509Config` | `GcpPrivateCaCertificateX509Parameters` | yes |  |  |
| `spec.config.x509Config.keyUsage` | `GcpPrivateCaCertificateKeyUsage` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage` | `GcpPrivateCaCertificateBaseKeyUsage` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.digitalSignature` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.contentCommitment` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.keyEncipherment` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.dataEncipherment` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.keyAgreement` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.certSign` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.crlSign` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.encipherOnly` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.decipherOnly` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage` | `GcpPrivateCaCertificateExtendedKeyUsage` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.serverAuth` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.clientAuth` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.codeSigning` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.emailProtection` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.timeStamping` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.ocspSigning` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.unknownExtendedKeyUsages` | `[]GcpPrivateCaCertificateObjectId` |  |  |  |
| `spec.config.x509Config.keyUsage.unknownExtendedKeyUsages[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.config.x509Config.caOptions` | `GcpPrivateCaCertificateCaOptions` |  |  |  |
| `spec.config.x509Config.caOptions.isCa` | `bool` |  |  |  |
| `spec.config.x509Config.caOptions.maxIssuerPathLength` | `int32` |  |  |  |
| `spec.config.x509Config.policyIds` | `[]GcpPrivateCaCertificateObjectId` |  |  |  |
| `spec.config.x509Config.policyIds[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.config.x509Config.aiaOcspServers` | `[]string` |  |  |  |
| `spec.config.x509Config.additionalExtensions` | `[]GcpPrivateCaCertificateX509Extension` |  |  |  |
| `spec.config.x509Config.additionalExtensions[].objectId` | `GcpPrivateCaCertificateObjectId` | yes |  |  |
| `spec.config.x509Config.additionalExtensions[].objectId.objectIdPath` | `[]int32` | yes |  |  |
| `spec.config.x509Config.additionalExtensions[].critical` | `bool` |  |  |  |
| `spec.config.x509Config.additionalExtensions[].value` | `string` | yes |  |  |
| `spec.config.x509Config.nameConstraints` | `GcpPrivateCaCertificateNameConstraints` |  |  |  |
| `spec.config.x509Config.nameConstraints.critical` | `bool` |  |  |  |
| `spec.config.x509Config.nameConstraints.permittedDnsNames` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.excludedDnsNames` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.permittedIpRanges` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.excludedIpRanges` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.permittedEmailAddresses` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.excludedEmailAddresses` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.permittedUris` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.excludedUris` | `[]string` |  |  |  |
| `spec.config.publicKey` | `GcpPrivateCaCertificatePublicKey` | yes |  |  |
| `spec.config.publicKey.key` | `string` | yes |  |  |
| `spec.config.publicKey.format` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the certificate is issued in (its pool's project): a literal project ID or a GcpProject reference.
If omitted, the provider's default project is used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The CA Service region, e.g. "us-central1" -- its pool's region. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.pool

`string | valueFrom` · required

The CA pool that issues the certificate. A GcpPrivateCaPool reference resolves to its full
resource path (name output); a literal takes the full path
projects/{project}/locations/{location}/caPools/{id} or the bare pool
ID. The modules derive the bare ID Google's resource expects; the pool
must be in this resource's project and location. Immutable.

- references: GcpPrivateCaPool (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPrivateCaPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.certificateId

`string`

The certificate's ID, unique in its parent: letters, digits, "-" and "_",
at most 63 characters. Defaults to metadata.name. Immutable.

- rule: certificate_id must be 1-63 letters, digits, '-' or '_'

### spec.certificateAuthority

`string | valueFrom`

The authority in the pool that signs the certificate -- a
GcpPrivateCaCertificateAuthority reference (its full name) or a
literal full name or bare ID; the modules derive the ID. Empty lets
the pool choose an enabled authority. Immutable.

- references: GcpPrivateCaCertificateAuthority (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPrivateCaCertificateAuthority, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.certificateTemplate

`string | valueFrom`

The template to issue with -- a GcpPrivateCaCertificateTemplate
reference (its full name) or a literal
projects/*/locations/*/certificateTemplates/*, in the certificate's
location. The caller needs privateca.templateUser on it. Immutable.

- references: GcpPrivateCaCertificateTemplate (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPrivateCaCertificateTemplate, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.lifetime

`string`

How long the certificate is valid. Defaults to 315360000s (10 years);
cut short by the pool's or template's maximum and by the issuing
authority's own expiry. Immutable.

- rule: lifetime must be a duration in seconds with an 's' suffix, e.g. 86400s

### spec.pemCsr

`string`

A PEM certificate signing request. Exactly one of pem_csr or config.
Immutable.

### spec.config

`GcpPrivateCaCertificateConfig`

The certificate as structured fields. Exactly one of pem_csr or
config. Immutable.

### spec.config.subjectConfig

`GcpPrivateCaCertificateSubjectConfig` · required

The certificate's subject.

- rule: {"required":true}

### spec.config.subjectConfig.subject

`GcpPrivateCaCertificateSubject` · required

The distinguished name. common_name is required.

- rule: {"required":true}

### spec.config.subjectConfig.subject.commonName

`string` · required

The common name (CN), e.g. "Example Root CA" or a host name.

- rule: {"string":{"minLen":"1"}}

### spec.config.subjectConfig.subject.countryCode

`string`

Two-letter country code (C), e.g. "US".

### spec.config.subjectConfig.subject.organization

`string`

Organization (O).

### spec.config.subjectConfig.subject.organizationalUnit

`string`

Organizational unit (OU).

### spec.config.subjectConfig.subject.locality

`string`

Locality or city (L).

### spec.config.subjectConfig.subject.province

`string`

Province, territory, or state (ST).

### spec.config.subjectConfig.subject.streetAddress

`string`

Street address.

### spec.config.subjectConfig.subject.postalCode

`string`

Postal code.

### spec.config.subjectConfig.subjectAltName

`GcpPrivateCaCertificateSubjectAltName`

Subject alternative names -- what TLS clients actually check against
the host they dialed. At least one entry when set.

- rule: subject_alt_name needs at least one of dns_names, uris, email_addresses, or ip_addresses

### spec.config.subjectConfig.subjectAltName.dnsNames

`[]string`

DNS names, e.g. "api.internal.example.com".

### spec.config.subjectConfig.subjectAltName.uris

`[]string`

URIs, e.g. a SPIFFE ID "spiffe://example.org/ns/prod/sa/api".

### spec.config.subjectConfig.subjectAltName.emailAddresses

`[]string`

Email addresses.

### spec.config.subjectConfig.subjectAltName.ipAddresses

`[]string`

IPv4 or IPv6 addresses.

### spec.config.subjectKeyId

`string`

A custom Subject Key Identifier, lowercase hex. Omit to let Google
derive it.

### spec.config.x509Config

`GcpPrivateCaCertificateX509Parameters` · required

The certificate's X.509 fields. A TLS server leaf sets
key_usage.base_key_usage.digital_signature and key_encipherment and
key_usage.extended_key_usage.server_auth.

- rule: {"required":true}

### spec.config.x509Config.keyUsage

`GcpPrivateCaCertificateKeyUsage`

What the certificate's key may be used for (the key usage and extended
key usage extensions). Omit for no key-usage statement; both usage
groups are optional and an empty group states nothing.

### spec.config.x509Config.keyUsage.baseKeyUsage

`GcpPrivateCaCertificateBaseKeyUsage`

The key usage extension's bits.

### spec.config.x509Config.keyUsage.baseKeyUsage.digitalSignature

`bool`

The key may make digital signatures other than certificate and CRL
signatures (TLS handshakes, signed tokens).

### spec.config.x509Config.keyUsage.baseKeyUsage.contentCommitment

`bool`

The key may protect content against the signer later denying it
(non-repudiation).

### spec.config.x509Config.keyUsage.baseKeyUsage.keyEncipherment

`bool`

The key may encipher other keys (RSA key transport in TLS).

### spec.config.x509Config.keyUsage.baseKeyUsage.dataEncipherment

`bool`

The key may encipher raw data directly.

### spec.config.x509Config.keyUsage.baseKeyUsage.keyAgreement

`bool`

The key may be used in key agreement (ECDH).

### spec.config.x509Config.keyUsage.baseKeyUsage.certSign

`bool`

The key may sign certificates -- set on every CA certificate.

### spec.config.x509Config.keyUsage.baseKeyUsage.crlSign

`bool`

The key may sign certificate revocation lists -- set on every CA
certificate that publishes CRLs.

### spec.config.x509Config.keyUsage.baseKeyUsage.encipherOnly

`bool`

With key_agreement, the key may only encipher.

### spec.config.x509Config.keyUsage.baseKeyUsage.decipherOnly

`bool`

With key_agreement, the key may only decipher.

### spec.config.x509Config.keyUsage.extendedKeyUsage

`GcpPrivateCaCertificateExtendedKeyUsage`

The extended key usage extension's well-known purposes.

### spec.config.x509Config.keyUsage.extendedKeyUsage.serverAuth

`bool`

TLS server authentication -- what an HTTPS or gRPC server certificate
needs.

### spec.config.x509Config.keyUsage.extendedKeyUsage.clientAuth

`bool`

TLS client authentication -- what an mTLS client certificate needs.

### spec.config.x509Config.keyUsage.extendedKeyUsage.codeSigning

`bool`

Code signing.

### spec.config.x509Config.keyUsage.extendedKeyUsage.emailProtection

`bool`

S/MIME email protection.

### spec.config.x509Config.keyUsage.extendedKeyUsage.timeStamping

`bool`

Trusted timestamping.

### spec.config.x509Config.keyUsage.extendedKeyUsage.ocspSigning

`bool`

Signing OCSP responses.

### spec.config.x509Config.keyUsage.unknownExtendedKeyUsages

`[]GcpPrivateCaCertificateObjectId`

Extended key usages with no named field here, each an OID path.

### spec.config.x509Config.keyUsage.unknownExtendedKeyUsages[].objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.config.x509Config.caOptions

`GcpPrivateCaCertificateCaOptions`

The basic constraints extension: whether the certificate is a CA and
how many CA levels may sit below it. Omit to leave the extension to
Google's default (for a leaf certificate, is_ca false).

### spec.config.x509Config.caOptions.isCa

`bool` · optional (explicit presence)

true makes a CA certificate (CA:TRUE), false states CA:FALSE; unset
leaves the CA flag out of the extension.

### spec.config.x509Config.caOptions.maxIssuerPathLength

`int32` · optional (explicit presence)

The path length constraint: how many CA levels may sit below this
certificate. 0 means it may sign only leaf certificates -- a
subordinate that issues workload certificates; unset means no limit
is stated.

- rule: {"int32":{"gte":0}}

### spec.config.x509Config.policyIds

`[]GcpPrivateCaCertificateObjectId`

Certificate policy object identifiers (RFC 5280 section 4.2.1.4), each
an OID path such as [2, 23, 140, 1, 2, 1] (CA/Browser Forum
domain-validated).

### spec.config.x509Config.policyIds[].objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.config.x509Config.aiaOcspServers

`[]string`

OCSP responder URLs placed in the Authority Information Access
extension, e.g. "http://ocsp.example.com". Google runs no OCSP
responder; list your own.

### spec.config.x509Config.additionalExtensions

`[]GcpPrivateCaCertificateX509Extension`

Custom X.509 extensions, each an OID, a base64 DER value, and whether
a relying party that does not understand it must reject the
certificate (critical).

### spec.config.x509Config.additionalExtensions[].objectId

`GcpPrivateCaCertificateObjectId` · required

The extension's OID.

- rule: {"required":true}

### spec.config.x509Config.additionalExtensions[].objectId.objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.config.x509Config.additionalExtensions[].critical

`bool`

A relying party that does not understand a critical extension must
reject the certificate. Mark an extension critical only when every
consumer knows it.

### spec.config.x509Config.additionalExtensions[].value

`string` · required

The extension's DER value, base64-encoded.

- rule: {"string":{"minLen":"1"}}

### spec.config.x509Config.nameConstraints

`GcpPrivateCaCertificateNameConstraints`

The name constraints extension (RFC 5280 section 4.2.1.10): the names
certificates below a CA may and may not carry. Meaningful on CA
certificates.

### spec.config.x509Config.nameConstraints.critical

`bool`

Whether the constraints are marked critical. RFC 5280 requires true on
a CA certificate that carries them.

### spec.config.x509Config.nameConstraints.permittedDnsNames

`[]string`

DNS names certificates below this CA may carry.

### spec.config.x509Config.nameConstraints.excludedDnsNames

`[]string`

DNS names certificates below this CA must not carry.

### spec.config.x509Config.nameConstraints.permittedIpRanges

`[]string`

IP ranges (CIDR) certificates below this CA may carry.

### spec.config.x509Config.nameConstraints.excludedIpRanges

`[]string`

IP ranges (CIDR) certificates below this CA must not carry.

### spec.config.x509Config.nameConstraints.permittedEmailAddresses

`[]string`

Email addresses, hosts, or ".domain" suffixes that may appear.

### spec.config.x509Config.nameConstraints.excludedEmailAddresses

`[]string`

Email addresses, hosts, or ".domain" suffixes that must not appear.

### spec.config.x509Config.nameConstraints.permittedUris

`[]string`

URI hosts or ".domain" suffixes that may appear.

### spec.config.x509Config.nameConstraints.excludedUris

`[]string`

URI hosts or ".domain" suffixes that must not appear.

### spec.config.publicKey

`GcpPrivateCaCertificatePublicKey` · required

The public key to certify.

- rule: {"required":true}

### spec.config.publicKey.key

`string` · required

The PEM public key (a SubjectPublicKeyInfo block), base64-encoded as a
whole -- Terraform's filebase64("key.pub.pem").

- rule: {"string":{"minLen":"1"}}

### spec.config.publicKey.format

`string`

The key's format. PEM is the only value Google accepts; empty sends
PEM.

- rule: format must be PEM

### spec.labels

`map<string, string>`

Labels on the certificate. The platform attribution labels are added on
top and win on a key conflict.

### spec.deletionPolicy

`string`

What happens to the certificate when this resource is destroyed:
  "" / "DELETE" -- deleted (the provider's default) -- which revokes it
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.exactly_one_request`: set exactly one of pem_csr or config

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPrivateCaCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/caPools/{pool}/certificates/{certificate_id}. |
| `status.outputs.certificate_id` | `string` | The certificate's ID (the last segment of name). |
| `status.outputs.pem_certificate` | `string` | The signed certificate, PEM. |
| `status.outputs.pem_certificate_chain` | `[]string` | The chain that verifies it, PEM, issuer first and root last. |
| `status.outputs.issuer_certificate_authority` | `string` | The full name of the authority that signed it. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.pool` | GcpPrivateCaPool | `status.outputs.name` |
| `spec.certificateAuthority` | GcpPrivateCaCertificateAuthority | `status.outputs.name` |
| `spec.certificateTemplate` | GcpPrivateCaCertificateTemplate | `status.outputs.name` |

## See Also

- [Overview](../README.md)
