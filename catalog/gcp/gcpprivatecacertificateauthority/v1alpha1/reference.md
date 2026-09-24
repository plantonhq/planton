# GcpPrivateCaCertificateAuthority

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpPrivateCaCertificateAuthoritySpec defines a certificate authority in a CA pool
(`google_privateca_certificate_authority`) -- a root that signs itself, or
a subordinate signed by another authority (in CA Service, by reference)
or by an outside CA. A pool issues through every enabled authority in it,
so rotation is: add a new authority, enable it, disable the old one.

Lifecycle Google enforces:
  - a self-signed root is ENABLED after create, or STAGED when
    desired_state says so (trusted but not yet issuing);
  - a subordinate signed by reference is activated and enabled on create;
    one signed by an outside CA needs pem_ca_certificate and
    subordinate_config.pem_issuer_chain; with neither it waits in
    AWAITING_USER_ACTIVATION;
  - destroy disables it and schedules deletion after 30 days (restorable
    until then) unless skip_grace_period; its pool cannot be deleted
    until the authority is gone.

Everything that defines the certificate -- config, key_spec, type,
lifetime, gcs_bucket -- is immutable: a change replaces the authority.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificateAuthority
metadata:
  name: root-ca
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  pool:
    value: projects/my-gcp-project/locations/us-central1/caPools/root-pool
  certificateAuthorityId: root-ca-2026
  type: SELF_SIGNED
  config:
    subjectConfig:
      subject:
        commonName: Example Root CA
        organization: Example
        countryCode: US
    x509Config:
      caOptions:
        isCa: true
        maxIssuerPathLength: 1
      keyUsage:
        baseKeyUsage:
          certSign: true
          crlSign: true
      nameConstraints:
        critical: true
        permittedDnsNames:
          - internal.example.com
  keySpec:
    algorithm: EC_P384_SHA384
  lifetime: 630720000s
  labels:
    tier: root
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.pool` | `string \| valueFrom` | yes |  | GcpPrivateCaPool (`status.outputs.name`) |
| `spec.certificateAuthorityId` | `string` |  |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.config` | `GcpPrivateCaCertificateAuthorityConfig` | yes |  |  |
| `spec.config.subjectConfig` | `GcpPrivateCaCertificateAuthoritySubjectConfig` | yes |  |  |
| `spec.config.subjectConfig.subject` | `GcpPrivateCaCertificateAuthoritySubject` | yes |  |  |
| `spec.config.subjectConfig.subject.commonName` | `string` | yes |  |  |
| `spec.config.subjectConfig.subject.countryCode` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.organization` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.organizationalUnit` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.locality` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.province` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.streetAddress` | `string` |  |  |  |
| `spec.config.subjectConfig.subject.postalCode` | `string` |  |  |  |
| `spec.config.subjectConfig.subjectAltName` | `GcpPrivateCaCertificateAuthoritySubjectAltName` |  |  |  |
| `spec.config.subjectConfig.subjectAltName.dnsNames` | `[]string` |  |  |  |
| `spec.config.subjectConfig.subjectAltName.uris` | `[]string` |  |  |  |
| `spec.config.subjectConfig.subjectAltName.emailAddresses` | `[]string` |  |  |  |
| `spec.config.subjectConfig.subjectAltName.ipAddresses` | `[]string` |  |  |  |
| `spec.config.subjectKeyId` | `string` |  |  |  |
| `spec.config.x509Config` | `GcpPrivateCaCertificateAuthorityX509Parameters` | yes |  |  |
| `spec.config.x509Config.keyUsage` | `GcpPrivateCaCertificateAuthorityKeyUsage` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage` | `GcpPrivateCaCertificateAuthorityBaseKeyUsage` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.digitalSignature` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.contentCommitment` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.keyEncipherment` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.dataEncipherment` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.keyAgreement` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.certSign` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.crlSign` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.encipherOnly` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.baseKeyUsage.decipherOnly` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage` | `GcpPrivateCaCertificateAuthorityExtendedKeyUsage` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.serverAuth` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.clientAuth` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.codeSigning` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.emailProtection` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.timeStamping` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.extendedKeyUsage.ocspSigning` | `bool` |  |  |  |
| `spec.config.x509Config.keyUsage.unknownExtendedKeyUsages` | `[]GcpPrivateCaCertificateAuthorityObjectId` |  |  |  |
| `spec.config.x509Config.keyUsage.unknownExtendedKeyUsages[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.config.x509Config.caOptions` | `GcpPrivateCaCertificateAuthorityCaOptions` |  |  |  |
| `spec.config.x509Config.caOptions.isCa` | `bool` |  |  |  |
| `spec.config.x509Config.caOptions.maxIssuerPathLength` | `int32` |  |  |  |
| `spec.config.x509Config.policyIds` | `[]GcpPrivateCaCertificateAuthorityObjectId` |  |  |  |
| `spec.config.x509Config.policyIds[].objectIdPath` | `[]int32` | yes |  |  |
| `spec.config.x509Config.aiaOcspServers` | `[]string` |  |  |  |
| `spec.config.x509Config.additionalExtensions` | `[]GcpPrivateCaCertificateAuthorityX509Extension` |  |  |  |
| `spec.config.x509Config.additionalExtensions[].objectId` | `GcpPrivateCaCertificateAuthorityObjectId` | yes |  |  |
| `spec.config.x509Config.additionalExtensions[].objectId.objectIdPath` | `[]int32` | yes |  |  |
| `spec.config.x509Config.additionalExtensions[].critical` | `bool` |  |  |  |
| `spec.config.x509Config.additionalExtensions[].value` | `string` | yes |  |  |
| `spec.config.x509Config.nameConstraints` | `GcpPrivateCaCertificateAuthorityNameConstraints` |  |  |  |
| `spec.config.x509Config.nameConstraints.critical` | `bool` |  |  |  |
| `spec.config.x509Config.nameConstraints.permittedDnsNames` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.excludedDnsNames` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.permittedIpRanges` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.excludedIpRanges` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.permittedEmailAddresses` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.excludedEmailAddresses` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.permittedUris` | `[]string` |  |  |  |
| `spec.config.x509Config.nameConstraints.excludedUris` | `[]string` |  |  |  |
| `spec.keySpec` | `GcpPrivateCaCertificateAuthorityKeySpec` | yes |  |  |
| `spec.keySpec.algorithm` | `string` |  |  |  |
| `spec.keySpec.cloudKmsKeyVersion` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.primary_version_name`) |
| `spec.lifetime` | `string` |  |  |  |
| `spec.subordinateConfig` | `GcpPrivateCaCertificateAuthoritySubordinateConfig` |  |  |  |
| `spec.subordinateConfig.certificateAuthority` | `string \| valueFrom` |  |  | GcpPrivateCaCertificateAuthority (`status.outputs.name`) |
| `spec.subordinateConfig.pemIssuerChain` | `[]string` |  |  |  |
| `spec.pemCaCertificate` | `string` |  |  |  |
| `spec.gcsBucket` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.bucket_name`) |
| `spec.userDefinedAccessUrls` | `GcpPrivateCaCertificateAuthorityUserDefinedAccessUrls` |  |  |  |
| `spec.userDefinedAccessUrls.aiaIssuingCertificateUrls` | `[]string` |  |  |  |
| `spec.userDefinedAccessUrls.crlAccessUrls` | `[]string` |  |  |  |
| `spec.desiredState` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.skipGracePeriod` | `bool` |  |  |  |
| `spec.ignoreActiveCertificatesOnDeletion` | `bool` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the authority lives in (its pool's project): a literal project ID or a GcpProject reference.
If omitted, the provider's default project is used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The CA Service region, e.g. "us-central1" -- its pool's region. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.pool

`string | valueFrom` · required

The CA pool the authority lives in. A GcpPrivateCaPool reference resolves to its full
resource path (name output); a literal takes the full path
projects/{project}/locations/{location}/caPools/{id} or the bare pool
ID. The modules derive the bare ID Google's resource expects; the pool
must be in this resource's project and location. Immutable.

- references: GcpPrivateCaPool (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPrivateCaPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.certificateAuthorityId

`string`

The authority's ID, unique in its parent: letters, digits, "-" and "_",
at most 63 characters. Defaults to metadata.name. Immutable.

- rule: certificate_authority_id must be 1-63 letters, digits, '-' or '_'

### spec.type

`string`

SELF_SIGNED (a root; the default) or SUBORDINATE (signed by another
authority -- set subordinate_config). Immutable.

- rule: type must be SELF_SIGNED or SUBORDINATE

### spec.config

`GcpPrivateCaCertificateAuthorityConfig` · required

The authority's own certificate. Immutable.

- rule: {"required":true}
- rule: an authority's x509_config.ca_options.is_ca must be set (true for a CA certificate)

### spec.config.subjectConfig

`GcpPrivateCaCertificateAuthoritySubjectConfig` · required

The CA certificate's subject.

- rule: {"required":true}

### spec.config.subjectConfig.subject

`GcpPrivateCaCertificateAuthoritySubject` · required

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

`GcpPrivateCaCertificateAuthoritySubjectAltName`

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

A custom Subject Key Identifier, lowercase hex. Only to keep the SKI of
a CA first created outside CA Service; omit to let Google derive it.

### spec.config.x509Config

`GcpPrivateCaCertificateAuthorityX509Parameters` · required

The CA certificate's X.509 fields. ca_options.is_ca is required; a CA
sets is_ca true, key_usage.base_key_usage.cert_sign and crl_sign, and
usually a max_issuer_path_length (0 for an authority that issues only
leaf certificates).

- rule: {"required":true}

### spec.config.x509Config.keyUsage

`GcpPrivateCaCertificateAuthorityKeyUsage`

What the certificate's key may be used for (the key usage and extended
key usage extensions). Omit for no key-usage statement; both usage
groups are optional and an empty group states nothing.

### spec.config.x509Config.keyUsage.baseKeyUsage

`GcpPrivateCaCertificateAuthorityBaseKeyUsage`

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

`GcpPrivateCaCertificateAuthorityExtendedKeyUsage`

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

`[]GcpPrivateCaCertificateAuthorityObjectId`

Extended key usages with no named field here, each an OID path.

### spec.config.x509Config.keyUsage.unknownExtendedKeyUsages[].objectIdPath

`[]int32` · required

The OID's arcs, most significant first.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"gte":0}}}}

### spec.config.x509Config.caOptions

`GcpPrivateCaCertificateAuthorityCaOptions`

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

`[]GcpPrivateCaCertificateAuthorityObjectId`

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

`[]GcpPrivateCaCertificateAuthorityX509Extension`

Custom X.509 extensions, each an OID, a base64 DER value, and whether
a relying party that does not understand it must reject the
certificate (critical).

### spec.config.x509Config.additionalExtensions[].objectId

`GcpPrivateCaCertificateAuthorityObjectId` · required

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

`GcpPrivateCaCertificateAuthorityNameConstraints`

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

### spec.keySpec

`GcpPrivateCaCertificateAuthorityKeySpec` · required

The authority's signing key. Immutable.

- rule: {"required":true}
- rule: key_spec is exactly one of algorithm or cloud_kms_key_version

### spec.keySpec.algorithm

`string`

A Google-managed Cloud HSM key of this algorithm:
  EC_P256_SHA256, EC_P384_SHA384 -- small, fast, widely supported
  RSA_PKCS1_2048_SHA256, RSA_PKCS1_3072_SHA256, RSA_PKCS1_4096_SHA256
  RSA_PSS_2048_SHA256, RSA_PSS_3072_SHA256, RSA_PSS_4096_SHA256

- rule: algorithm must be one of the RSA_PSS_*, RSA_PKCS1_*, or EC_P256_SHA256 / EC_P384_SHA384 values

### spec.keySpec.cloudKmsKeyVersion

`string | valueFrom`

A Cloud KMS key version you own, with an asymmetric-sign purpose -- a
GcpKmsKey reference (its primary version) or a literal
projects/*/locations/*/keyRings/*/cryptoKeys/*/cryptoKeyVersions/*.
CA Service's service agent needs signerVerifier and viewer on the key.
Enterprise pools only (Google's tier rule).

- references: GcpKmsKey (`status.outputs.primary_version_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.primary_version_name}} -- a bare string does not parse

### spec.lifetime

`string`

How long the CA certificate is valid. Defaults to 315360000s (10
years); roots typically 10-20 years, subordinates shorter. Issued
certificates never outlive it. Immutable.

- rule: lifetime must be a duration in seconds with an 's' suffix, e.g. 86400s

### spec.subordinateConfig

`GcpPrivateCaCertificateAuthoritySubordinateConfig`

A subordinate's issuer. Updatable in place, but the authority must
continue to validate against it.

- rule: subordinate_config is exactly one of certificate_authority or pem_issuer_chain

### spec.subordinateConfig.certificateAuthority

`string | valueFrom`

The CA Service authority that signs this one -- a
GcpPrivateCaCertificateAuthority reference (its full name) or a
literal projects/*/locations/*/caPools/*/certificateAuthorities/*,
often in another pool. The modules have it sign this authority's CSR
and activate it on create.

- references: GcpPrivateCaCertificateAuthority (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPrivateCaCertificateAuthority, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.subordinateConfig.pemIssuerChain

`[]string`

An external issuer's PEM certificate chain, issuer first, not
including this authority's own certificate. With pem_ca_certificate,
it activates the authority from an outside CA.

### spec.pemCaCertificate

`string`

The CA certificate an outside CA signed from this authority's CSR
(read the CSR from Google after the first apply), PEM. With
subordinate_config.pem_issuer_chain, it activates the authority.

### spec.gcsBucket

`string | valueFrom`

The Cloud Storage bucket the authority publishes its CA certificate
and CRLs to -- a GcpGcsBucket reference or a bare bucket name (no
gs://). CA Service's service agent needs write access. Empty lets
Google create a managed bucket. Immutable.

- references: GcpGcsBucket (`status.outputs.bucket_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.userDefinedAccessUrls

`GcpPrivateCaCertificateAuthorityUserDefinedAccessUrls`

URLs advertised in place of the Cloud Storage ones.

### spec.userDefinedAccessUrls.aiaIssuingCertificateUrls

`[]string`

Where the issuer CA certificate may be downloaded (Authority
Information Access extension).

### spec.userDefinedAccessUrls.crlAccessUrls

`[]string`

Where the CRL may be fetched (CRL Distribution Points extension).

### spec.desiredState

`string`

The state to hold the authority in:
  "" / "ENABLED" -- issuing (the default after create)
  "STAGED"       -- created trusted but not issuing; only at create
  "DISABLED"     -- not issuing; only after create (the provider
                    refuses DISABLED on a new authority)

- rule: desired_state must be ENABLED, DISABLED, or STAGED

### spec.deletionProtection

`bool` · optional (explicit presence)

Guard against destroying the authority. Defaults to true: a destroy
fails until the manifest sets false and is applied.

### spec.skipGracePeriod

`bool`

Delete immediately on destroy instead of after Google's 30-day soft
delete. Irreversible; required when the pool is destroyed in the same
run.

### spec.ignoreActiveCertificatesOnDeletion

`bool`

Allow destroying an authority whose issued certificates have not
expired or been revoked. They stay valid until they expire.

### spec.labels

`map<string, string>`

Labels on the authority. The platform attribution labels are added on
top and win on a key conflict.

### spec.deletionPolicy

`string`

What happens to the authority when this resource is destroyed:
  "" / "DELETE" -- deleted (the provider's default) -- subject to
                   deletion_protection and the soft delete
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.activation_needs_subordinate`: subordinate_config and pem_ca_certificate apply only to type SUBORDINATE
- `spec.third_party_needs_issuer_chain`: pem_ca_certificate needs subordinate_config.pem_issuer_chain (activation from an outside CA)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPrivateCaCertificateAuthority, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/caPools/{pool}/certificateAuthorities/{id}. |
| `status.outputs.certificate_authority_id` | `string` | The authority's ID (the last segment of name). |
| `status.outputs.state` | `string` | The authority's state: ENABLED, DISABLED, STAGED, or AWAITING_USER_ACTIVATION. |
| `status.outputs.pem_ca_certificate` | `string` | The authority's own CA certificate, PEM -- the trust anchor for a root. |
| `status.outputs.pem_ca_certificates` | `[]string` | The authority's certificate chain, PEM, its own certificate first and the root last. |
| `status.outputs.ca_certificate_access_url` | `string` | Where Google publishes the CA certificate. |
| `status.outputs.crl_access_urls` | `[]string` | Where Google publishes the authority's CRLs. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.pool` | GcpPrivateCaPool | `status.outputs.name` |
| `spec.keySpec.cloudKmsKeyVersion` | GcpKmsKey | `status.outputs.primary_version_name` |
| `spec.subordinateConfig.certificateAuthority` | GcpPrivateCaCertificateAuthority | `status.outputs.name` |
| `spec.gcsBucket` | GcpGcsBucket | `status.outputs.bucket_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpPrivateCaCertificate | `spec.certificateAuthority` | `status.outputs.name` |
| GcpPrivateCaCertificateAuthority | `spec.subordinateConfig.certificateAuthority` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
