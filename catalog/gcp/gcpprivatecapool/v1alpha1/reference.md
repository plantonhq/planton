# GcpPrivateCaPool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpPrivateCaPoolSpec defines a Certificate Authority Service CA pool
(`google_privateca_ca_pool`) -- a group of certificate authorities that
form one trust anchor, with the policy every certificate they issue
follows. Authorities (GcpPrivateCaCertificateAuthority) and certificates
(GcpPrivateCaCertificate) live inside the pool and reference it; relying
services trust the pool, so authorities rotate in and out behind it.

Choose the tier for the workload, once -- it is immutable:
  ENTERPRISE -- long-lived certificates with lifecycle management: Google
                stores every certificate, lists, describes, and revokes
                them; authorities may use your own Cloud KMS key
  DEVOPS     -- high-volume, short-lived certificates (about 3.5x the
                issuance rate per authority); certificates are not
                stored, so they cannot be listed, described, or revoked,
                and authorities use Google-managed keys only

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaPool
metadata:
  name: internal-tls
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  tier: ENTERPRISE
  publishingOptions:
    publishCaCert: true
    publishCrl: true
    encodingFormat: PEM
  issuancePolicy:
    maximumLifetime: 7776000s
    allowedKeyTypes:
      - ellipticCurveSignatureAlgorithm: ECDSA_P256
      - rsa:
          minModulusSize: 2048
          maxModulusSize: 4096
    allowedIssuanceModes:
      allowCsrBasedIssuance: true
      allowConfigBasedIssuance: true
    identityConstraints:
      allowSubjectPassthrough: true
      allowSubjectAltNamesPassthrough: true
      celExpression:
        title: Internal names
        expression: subject_alt_names.all(san, san.type == DNS && san.value.endsWith(".internal.example.com"))
    baselineValues:
      caOptions:
        isCa: false
      keyUsage:
        baseKeyUsage:
          digitalSignature: true
          keyEncipherment: true
        extendedKeyUsage:
          serverAuth: true
          clientAuth: true
  labels:
    team: platform
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.caPoolId` | `string` |  |  |  |
| `spec.tier` | `string` | yes |  |  |
| `spec.issuancePolicy` | `GcpPrivateCaPoolIssuancePolicy` |  |  |  |
| `spec.issuancePolicy.allowedKeyTypes` | `[]GcpPrivateCaPoolAllowedKeyType` |  |  |  |
| `spec.issuancePolicy.allowedKeyTypes[].rsa` | `GcpPrivateCaPoolRsaKeyType` |  |  |  |
| `spec.issuancePolicy.allowedKeyTypes[].rsa.minModulusSize` | `int64` |  |  |  |
| `spec.issuancePolicy.allowedKeyTypes[].rsa.maxModulusSize` | `int64` |  |  |  |
| `spec.issuancePolicy.allowedKeyTypes[].ellipticCurveSignatureAlgorithm` | `string` |  |  |  |
| `spec.issuancePolicy.maximumLifetime` | `string` |  |  |  |
| `spec.issuancePolicy.backdateDuration` | `string` |  |  |  |
| `spec.issuancePolicy.allowedIssuanceModes` | `GcpPrivateCaPoolIssuanceModes` |  |  |  |
| `spec.issuancePolicy.allowedIssuanceModes.allowCsrBasedIssuance` | `bool` |  |  |  |
| `spec.issuancePolicy.allowedIssuanceModes.allowConfigBasedIssuance` | `bool` |  |  |  |
| `spec.issuancePolicy.identityConstraints` | `GcpPrivateCaPoolIdentityConstraints` |  |  |  |
| `spec.issuancePolicy.identityConstraints.allowSubjectPassthrough` | `bool` |  |  |  |
| `spec.issuancePolicy.identityConstraints.allowSubjectAltNamesPassthrough` | `bool` |  |  |  |
| `spec.issuancePolicy.identityConstraints.celExpression` | `GcpPrivateCaPoolCelExpression` |  |  |  |
| `spec.issuancePolicy.identityConstraints.celExpression.expression` | `string` | yes |  |  |
| `spec.issuancePolicy.identityConstraints.celExpression.title` | `string` |  |  |  |
| `spec.issuancePolicy.identityConstraints.celExpression.description` | `string` |  |  |  |
| `spec.issuancePolicy.identityConstraints.celExpression.location` | `string` |  |  |  |
| `spec.issuancePolicy.baselineValues` | `GcpPrivateCaPoolX509Parameters` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage` | `GcpPrivateCaPoolKeyUsage` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage` | `GcpPrivateCaPoolBaseKeyUsage` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.digitalSignature` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.contentCommitment` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.keyEncipherment` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.dataEncipherment` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.keyAgreement` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.certSign` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.crlSign` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.encipherOnly` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.decipherOnly` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage` | `GcpPrivateCaPoolExtendedKeyUsage` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.serverAuth` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.clientAuth` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.codeSigning` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.emailProtection` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.timeStamping` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.ocspSigning` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.unknownExtendedKeyUsages` | `[]GcpPrivateCaPoolObjectId` |  |  |  |
| `spec.issuancePolicy.baselineValues.keyUsage.unknownExtendedKeyUsages[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.issuancePolicy.baselineValues.caOptions` | `GcpPrivateCaPoolCaOptions` |  |  |  |
| `spec.issuancePolicy.baselineValues.caOptions.isCa` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.caOptions.maxIssuerPathLength` | `int32` |  |  |  |
| `spec.issuancePolicy.baselineValues.policyIds` | `[]GcpPrivateCaPoolObjectId` |  |  |  |
| `spec.issuancePolicy.baselineValues.policyIds[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.issuancePolicy.baselineValues.aiaOcspServers` | `[]string` |  |  |  |
| `spec.issuancePolicy.baselineValues.additionalExtensions` | `[]GcpPrivateCaPoolX509Extension` |  |  |  |
| `spec.issuancePolicy.baselineValues.additionalExtensions[].objectId` | `GcpPrivateCaPoolObjectId` | yes |  |  |
| `spec.issuancePolicy.baselineValues.additionalExtensions[].objectId.objectIdPath` | `[]int32` | yes |  |  |
| `spec.issuancePolicy.baselineValues.additionalExtensions[].critical` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.additionalExtensions[].value` | `string` | yes |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints` | `GcpPrivateCaPoolNameConstraints` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.critical` | `bool` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.permittedDnsNames` | `[]string` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.excludedDnsNames` | `[]string` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.permittedIpRanges` | `[]string` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.excludedIpRanges` | `[]string` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.permittedEmailAddresses` | `[]string` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.excludedEmailAddresses` | `[]string` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.permittedUris` | `[]string` |  |  |  |
| `spec.issuancePolicy.baselineValues.nameConstraints.excludedUris` | `[]string` |  |  |  |
| `spec.publishingOptions` | `GcpPrivateCaPoolPublishingOptions` |  |  |  |
| `spec.publishingOptions.publishCaCert` | `bool` |  |  |  |
| `spec.publishingOptions.publishCrl` | `bool` |  |  |  |
| `spec.publishingOptions.encodingFormat` | `string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the pool lives in: a literal project ID or a GcpProject reference.
If omitted, the provider's default project is used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The CA Service region, e.g. "us-central1"; the pool's authorities and certificates live in the same region. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.caPoolId

`string`

The pool's ID, unique in its parent: letters, digits, "-" and "_",
at most 63 characters. Defaults to metadata.name. Immutable.

- rule: ca_pool_id must be 1-63 letters, digits, '-' or '_'

### spec.tier

`string` · required

ENTERPRISE or DEVOPS (see the message comment). Immutable.

- rule: tier must be ENTERPRISE or DEVOPS
- rule: {"required":true}

### spec.issuancePolicy

`GcpPrivateCaPoolIssuancePolicy`

The policy every certificate the pool issues follows. Omit for no
policy: any key, both request forms, any identity, no baseline values.

### spec.issuancePolicy.allowedKeyTypes

`[]GcpPrivateCaPoolAllowedKeyType`

The key types a request's public key must match. Empty admits any key.

- rule: an allowed key type is exactly one of rsa or elliptic_curve_signature_algorithm

### spec.issuancePolicy.allowedKeyTypes[].rsa

`GcpPrivateCaPoolRsaKeyType`

An RSA key; an empty message admits any RSA key the service accepts.

- rule: min_modulus_size must not exceed max_modulus_size

### spec.issuancePolicy.allowedKeyTypes[].rsa.minModulusSize

`int64`

Smallest allowed modulus (inclusive), e.g. 2048.

- rule: {"int64":{"gte":"0"}}

### spec.issuancePolicy.allowedKeyTypes[].rsa.maxModulusSize

`int64`

Largest allowed modulus (inclusive), e.g. 4096.

- rule: {"int64":{"gte":"0"}}

### spec.issuancePolicy.allowedKeyTypes[].ellipticCurveSignatureAlgorithm

`string`

An elliptic-curve key for this signature algorithm:
  ECDSA_P256 -- the common choice for TLS
  ECDSA_P384
  EDDSA_25519

- rule: elliptic_curve_signature_algorithm must be ECDSA_P256, ECDSA_P384, or EDDSA_25519

### spec.issuancePolicy.maximumLifetime

`string`

The longest lifetime an issued certificate may have, e.g. 7776000s
(90 days). A certificate is also cut short when its issuing authority
expires first.

- rule: maximum_lifetime must be a duration in seconds with an 's' suffix, e.g. 86400s

### spec.issuancePolicy.backdateDuration

`string`

Backdate every issued certificate's not-before time by this much (the
lifetime is preserved), to absorb clock skew on clients, e.g. 3600s.
Google allows at most 48 hours.

- rule: backdate_duration must be a duration in seconds with an 's' suffix, e.g. 86400s at most 172800s

### spec.issuancePolicy.allowedIssuanceModes

`GcpPrivateCaPoolIssuanceModes`

Which request forms are allowed. Omit to allow both.

### spec.issuancePolicy.allowedIssuanceModes.allowCsrBasedIssuance

`bool`

Requests that carry a PEM certificate signing request (pem_csr).

### spec.issuancePolicy.allowedIssuanceModes.allowConfigBasedIssuance

`bool`

Requests that describe the certificate as structured config (subject,
SANs, public key) instead of a CSR.

### spec.issuancePolicy.identityConstraints

`GcpPrivateCaPoolIdentityConstraints`

Limits on the identities certificates may carry. Omit for no limits.

### spec.issuancePolicy.identityConstraints.allowSubjectPassthrough

`bool`

Copy the Subject from the certificate request into the certificate;
false discards the requested Subject.

### spec.issuancePolicy.identityConstraints.allowSubjectAltNamesPassthrough

`bool`

Copy the subject alternative names from the request; false discards
them.

### spec.issuancePolicy.identityConstraints.celExpression

`GcpPrivateCaPoolCelExpression`

A CEL expression over the resolved subject and subject alternative
names that must hold before a certificate is signed, e.g.
subject_alt_names.all(san, san.type == DNS && san.value.endsWith(".internal.example.com")).

### spec.issuancePolicy.identityConstraints.celExpression.expression

`string` · required

The expression.

- rule: {"string":{"minLen":"1"}}

### spec.issuancePolicy.identityConstraints.celExpression.title

`string`

A short title for the expression.

### spec.issuancePolicy.identityConstraints.celExpression.description

`string`

What the expression checks.

### spec.issuancePolicy.identityConstraints.celExpression.location

`string`

Where the expression came from, for error messages (a file and line).

### spec.issuancePolicy.baselineValues

`GcpPrivateCaPoolX509Parameters`

X.509 values stamped onto every certificate the pool issues.

### spec.issuancePolicy.baselineValues.keyUsage

`GcpPrivateCaPoolKeyUsage`

What the certificate's key may be used for (the key usage and extended
key usage extensions). Omit for no key-usage statement; both usage
groups are optional and an empty group states nothing.

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage

`GcpPrivateCaPoolBaseKeyUsage`

The key usage extension's bits.

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.digitalSignature

`bool`

The key may make digital signatures other than certificate and CRL
signatures (TLS handshakes, signed tokens).

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.contentCommitment

`bool`

The key may protect content against the signer later denying it
(non-repudiation).

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.keyEncipherment

`bool`

The key may encipher other keys (RSA key transport in TLS).

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.dataEncipherment

`bool`

The key may encipher raw data directly.

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.keyAgreement

`bool`

The key may be used in key agreement (ECDH).

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.certSign

`bool`

The key may sign certificates -- set on every CA certificate.

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.crlSign

`bool`

The key may sign certificate revocation lists -- set on every CA
certificate that publishes CRLs.

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.encipherOnly

`bool`

With key_agreement, the key may only encipher.

### spec.issuancePolicy.baselineValues.keyUsage.baseKeyUsage.decipherOnly

`bool`

With key_agreement, the key may only decipher.

### spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage

`GcpPrivateCaPoolExtendedKeyUsage`

The extended key usage extension's well-known purposes.

### spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.serverAuth

`bool`

TLS server authentication -- what an HTTPS or gRPC server certificate
needs.

### spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.clientAuth

`bool`

TLS client authentication -- what an mTLS client certificate needs.

### spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.codeSigning

`bool`

Code signing.

### spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.emailProtection

`bool`

S/MIME email protection.

### spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.timeStamping

`bool`

Trusted timestamping.

### spec.issuancePolicy.baselineValues.keyUsage.extendedKeyUsage.ocspSigning

`bool`

Signing OCSP responses.

### spec.issuancePolicy.baselineValues.keyUsage.unknownExtendedKeyUsages

`[]GcpPrivateCaPoolObjectId`

Extended key usages with no named field here, each an OID path.

### spec.issuancePolicy.baselineValues.keyUsage.unknownExtendedKeyUsages[].objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.issuancePolicy.baselineValues.caOptions

`GcpPrivateCaPoolCaOptions`

The basic constraints extension: whether the certificate is a CA and
how many CA levels may sit below it. Omit to leave the extension to
Google's default (for a leaf certificate, is_ca false).

### spec.issuancePolicy.baselineValues.caOptions.isCa

`bool` · optional (explicit presence)

true makes a CA certificate (CA:TRUE), false states CA:FALSE; unset
leaves the CA flag out of the extension.

### spec.issuancePolicy.baselineValues.caOptions.maxIssuerPathLength

`int32` · optional (explicit presence)

The path length constraint: how many CA levels may sit below this
certificate. 0 means it may sign only leaf certificates -- a
subordinate that issues workload certificates; unset means no limit
is stated.

- rule: {"int32":{"gte":0}}

### spec.issuancePolicy.baselineValues.policyIds

`[]GcpPrivateCaPoolObjectId`

Certificate policy object identifiers (RFC 5280 section 4.2.1.4), each
an OID path such as [2, 23, 140, 1, 2, 1] (CA/Browser Forum
domain-validated).

### spec.issuancePolicy.baselineValues.policyIds[].objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.issuancePolicy.baselineValues.aiaOcspServers

`[]string`

OCSP responder URLs placed in the Authority Information Access
extension, e.g. "http://ocsp.example.com". Google runs no OCSP
responder; list your own.

### spec.issuancePolicy.baselineValues.additionalExtensions

`[]GcpPrivateCaPoolX509Extension`

Custom X.509 extensions, each an OID, a base64 DER value, and whether
a relying party that does not understand it must reject the
certificate (critical).

### spec.issuancePolicy.baselineValues.additionalExtensions[].objectId

`GcpPrivateCaPoolObjectId` · required

The extension's OID.

- rule: {"required":true}

### spec.issuancePolicy.baselineValues.additionalExtensions[].objectId.objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.issuancePolicy.baselineValues.additionalExtensions[].critical

`bool`

A relying party that does not understand a critical extension must
reject the certificate. Mark an extension critical only when every
consumer knows it.

### spec.issuancePolicy.baselineValues.additionalExtensions[].value

`string` · required

The extension's DER value, base64-encoded.

- rule: {"string":{"minLen":"1"}}

### spec.issuancePolicy.baselineValues.nameConstraints

`GcpPrivateCaPoolNameConstraints`

The name constraints extension (RFC 5280 section 4.2.1.10): the names
certificates below a CA may and may not carry. Meaningful on CA
certificates.

### spec.issuancePolicy.baselineValues.nameConstraints.critical

`bool`

Whether the constraints are marked critical. RFC 5280 requires true on
a CA certificate that carries them.

### spec.issuancePolicy.baselineValues.nameConstraints.permittedDnsNames

`[]string`

DNS names certificates below this CA may carry.

### spec.issuancePolicy.baselineValues.nameConstraints.excludedDnsNames

`[]string`

DNS names certificates below this CA must not carry.

### spec.issuancePolicy.baselineValues.nameConstraints.permittedIpRanges

`[]string`

IP ranges (CIDR) certificates below this CA may carry.

### spec.issuancePolicy.baselineValues.nameConstraints.excludedIpRanges

`[]string`

IP ranges (CIDR) certificates below this CA must not carry.

### spec.issuancePolicy.baselineValues.nameConstraints.permittedEmailAddresses

`[]string`

Email addresses, hosts, or ".domain" suffixes that may appear.

### spec.issuancePolicy.baselineValues.nameConstraints.excludedEmailAddresses

`[]string`

Email addresses, hosts, or ".domain" suffixes that must not appear.

### spec.issuancePolicy.baselineValues.nameConstraints.permittedUris

`[]string`

URI hosts or ".domain" suffixes that may appear.

### spec.issuancePolicy.baselineValues.nameConstraints.excludedUris

`[]string`

URI hosts or ".domain" suffixes that must not appear.

### spec.publishingOptions

`GcpPrivateCaPoolPublishingOptions`

What the pool's authorities publish. Omit to publish nothing.

### spec.publishingOptions.publishCaCert

`bool`

Publish each authority's CA certificate and put its URL in the
Authority Information Access extension of issued certificates.

### spec.publishingOptions.publishCrl

`bool`

Publish each authority's CRL (rebuilt daily and shortly after a
revocation; each expires after 7 days) and put its URL in the CRL
Distribution Points extension.

### spec.publishingOptions.encodingFormat

`string`

How published CA certificates and CRLs are encoded: "PEM" (the
default) or "DER".

- rule: encoding_format must be PEM or DER

### spec.kmsKeyName

`string | valueFrom`

The Cloud KMS key that encrypts the subject, SANs, and PEM certificate
of stored certificates at rest -- a GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
in the pool's region. CA Service's service agent needs
cryptoKeyEncrypterDecrypter on it. Empty keeps Google's encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Labels on the pool. The platform attribution labels are added on
top and win on a key conflict.

### spec.deletionPolicy

`string`

What happens to the pool when this resource is destroyed:
  "" / "DELETE" -- deleted (the provider's default); Google refuses while the
                   pool still holds an authority, even one in its 30-day
                   soft delete (destroy authorities with skip_grace_period
                   first)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPrivateCaPool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/caPools/{ca_pool_id}. |
| `status.outputs.ca_pool_id` | `string` | The pool's ID (the last segment of name). |
| `status.outputs.location` | `string` | The region the pool lives in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpManagedKafkaCluster | `spec.tlsConfig.caPools` | `status.outputs.name` |
| GcpPrivateCaCertificate | `spec.pool` | `status.outputs.name` |
| GcpPrivateCaCertificateAuthority | `spec.pool` | `status.outputs.name` |
| GcpRedisCluster | `spec.serverCaPool` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
