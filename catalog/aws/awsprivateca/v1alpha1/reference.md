# AwsPrivateCa

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsPrivateCaSpec defines one AWS Private Certificate Authority -
the managed CA that issues internal TLS certificates (mTLS
services, MSK/EKS client auth, service meshes, IoT fleets) - with
its activation, issued certificates, ACM renewal permission, and
resource policy managed in-line.

ACTIVATION is composed, never user-choreographed: a ROOT CA
self-signs its own certificate at apply (the three-step
CSR->issue->install dance the raw provider makes you wire is done
by the modules), so a fresh ROOT CA comes up ACTIVE and issuing. A
SUBORDINATE CA activates when subordinate_activation names a
parent AwsPrivateCa in this account; without it the CA sits in
PENDING_CERTIFICATE (created, billed, not yet able to issue) until
its certificate is installed out of band - the posture for
external/offline parent CAs.

COST is real and per-hour: a GENERAL_PURPOSE CA bills its monthly
fee (USD 400/month, prorated hourly) from creation to deletion;
SHORT_LIVED_CERTIFICATE mode bills USD 50/month prorated. Deleting
a CA parks it restorable for permanent_deletion_time_in_days
(billing stops at delete).

## Example

```yaml
# Canonical AwsPrivateCa example (hack/dev manifest and refgen
# Example source): a self-activating internal root CA with the ACM
# renewal permission granted and a 7-day restore window - the posture
# for issuing internal TLS through AwsCertManagerCert.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsPrivateCa
metadata:
  name: corp-root-ca
  id: corp-root-ca
  org: test-org
  env: dev
spec:
  region: us-west-2
  type: ROOT
  keyAlgorithm: RSA_2048
  signingAlgorithm: SHA256WITHRSA
  subject:
    commonName: Corp Internal Root CA G1
    organization: Test Org
    country: US
  rootCaValidity:
    type: YEARS
    value: "10"
  acmRenewalPermission: true
  permanentDeletionTimeInDays: 7
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.keyAlgorithm` | `string` |  |  |  |
| `spec.signingAlgorithm` | `string` |  |  |  |
| `spec.subject` | `AwsPrivateCaSubject` | yes |  |  |
| `spec.subject.commonName` | `string` |  |  |  |
| `spec.subject.organization` | `string` |  |  |  |
| `spec.subject.organizationalUnit` | `string` |  |  |  |
| `spec.subject.country` | `string` |  |  |  |
| `spec.subject.state` | `string` |  |  |  |
| `spec.subject.locality` | `string` |  |  |  |
| `spec.usageMode` | `string` |  |  |  |
| `spec.keyStorageSecurityStandard` | `string` |  |  |  |
| `spec.revocation` | `AwsPrivateCaRevocation` |  |  |  |
| `spec.revocation.crl` | `AwsPrivateCaCrl` |  |  |  |
| `spec.revocation.crl.enabled` | `bool` |  |  |  |
| `spec.revocation.crl.expirationInDays` | `int64` |  |  |  |
| `spec.revocation.crl.s3BucketName` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.revocation.crl.s3ObjectAcl` | `string` |  |  |  |
| `spec.revocation.crl.customCname` | `string` |  |  |  |
| `spec.revocation.crl.customPath` | `string` |  |  |  |
| `spec.revocation.ocsp` | `AwsPrivateCaOcsp` |  |  |  |
| `spec.revocation.ocsp.enabled` | `bool` |  |  |  |
| `spec.revocation.ocsp.customCname` | `string` |  |  |  |
| `spec.rootCaValidity` | `AwsPrivateCaValidity` |  |  |  |
| `spec.rootCaValidity.type` | `string` |  |  |  |
| `spec.rootCaValidity.value` | `string` | yes |  |  |
| `spec.subordinateActivation` | `AwsPrivateCaSubordinateActivation` |  |  |  |
| `spec.subordinateActivation.parentCaArn` | `string \| valueFrom` | yes |  | AwsPrivateCa (`status.outputs.certificate_authority_arn`) |
| `spec.subordinateActivation.pathLength` | `int64` |  |  |  |
| `spec.subordinateActivation.validity` | `AwsPrivateCaValidity` | yes |  |  |
| `spec.subordinateActivation.validity.type` | `string` |  |  |  |
| `spec.subordinateActivation.validity.value` | `string` | yes |  |  |
| `spec.issuedCertificates` | `[]AwsPrivateCaIssuedCertificate` |  |  |  |
| `spec.issuedCertificates[].name` | `string` | yes |  |  |
| `spec.issuedCertificates[].csr` | `string` | yes |  |  |
| `spec.issuedCertificates[].signingAlgorithm` | `string` |  |  |  |
| `spec.issuedCertificates[].validity` | `AwsPrivateCaValidity` | yes |  |  |
| `spec.issuedCertificates[].validity.type` | `string` |  |  |  |
| `spec.issuedCertificates[].validity.value` | `string` | yes |  |  |
| `spec.issuedCertificates[].templateArn` | `string` |  |  |  |
| `spec.issuedCertificates[].apiPassthrough` | `string` |  |  |  |
| `spec.acmRenewalPermission` | `bool` |  |  |  |
| `spec.policy` | `string` |  |  |  |
| `spec.permanentDeletionTimeInDays` | `int64` |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |

## Field Details

### spec.region

`string` · required

The AWS region the CA lives in. Example: "us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.type

`string`

The CA's place in the hierarchy: "ROOT" (self-signed at apply -
up and issuing in one deploy) or "SUBORDINATE" (signed by a
parent CA - see subordinate_activation). Fixed for life.

- rule: {"string":{"in":["ROOT","SUBORDINATE"]}}

### spec.keyAlgorithm

`string`

The CA key pair's algorithm. RSA_2048 is the broadly compatible
default choice; EC keys are smaller and faster; ML_DSA_* are the
post-quantum lattice algorithms (pair with the matching ML_DSA
signing level); SM2 exists for China regions only. Fixed for
life.

- rule: {"string":{"in":["RSA_2048","RSA_3072","RSA_4096","EC_prime256v1","EC_secp384r1","EC_secp521r1","ML_DSA_44","ML_DSA_65","ML_DSA_87","SM2"]}}

### spec.signingAlgorithm

`string`

The algorithm the CA signs issued certificates with. Must match
the key family (the validation above). Fixed for life.

- rule: {"string":{"in":["SHA256WITHRSA","SHA384WITHRSA","SHA512WITHRSA","SHA256WITHECDSA","SHA384WITHECDSA","SHA512WITHECDSA","ML_DSA_44","ML_DSA_65","ML_DSA_87","SM3WITHSM2"]}}

### spec.subject

`AwsPrivateCaSubject` · required

The CA certificate's X.500 subject. Fixed for life.

- rule: {"required":true}
- rule: set at least one subject field - common_name alone is the usual choice for an internal CA

### spec.subject.commonName

`string`

The CA's common name, e.g. "corp-root-ca" or
"Acme Internal Root CA G1".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"64"}}

### spec.subject.organization

`string`

Legal organization name, e.g. "Acme Corp".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"64"}}

### spec.subject.organizationalUnit

`string`

Division within the organization, e.g. "Platform Engineering".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"64"}}

### spec.subject.country

`string`

Two-letter ISO country code, e.g. "US".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"len":"2","pattern":"^[A-Za-z]{2}$"}}

### spec.subject.state

`string`

State or province, e.g. "California".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128"}}

### spec.subject.locality

`string`

City, e.g. "San Francisco".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128"}}

### spec.usageMode

`string`

What the CA issues: "GENERAL_PURPOSE" (any validity, revocation
supported - USD 400/month) or "SHORT_LIVED_CERTIFICATE"
(certificates capped at 7 days, no revocation - USD 50/month;
the right mode for high-churn workload identity). Unset means
GENERAL_PURPOSE (the AWS default). Fixed for life - the provider
accepts a change plan but AWS never applies it.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["GENERAL_PURPOSE","SHORT_LIVED_CERTIFICATE"]}}

### spec.keyStorageSecurityStandard

`string`

The HSM standard protecting the CA key. Unset takes the region's
default (FIPS_140_2_LEVEL_3_OR_HIGHER in most regions; some
regions only offer LEVEL_2). Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FIPS_140_2_LEVEL_2_OR_HIGHER","FIPS_140_2_LEVEL_3_OR_HIGHER","CCPC_LEVEL_1_OR_HIGHER"]}}

### spec.revocation

`AwsPrivateCaRevocation`

How the CA publishes revocation state for the certificates it
issued. Omit for no revocation infrastructure (fine for
short-lived certificates and closed systems that distrust by
rotation).

### spec.revocation.crl

`AwsPrivateCaCrl`

Certificate Revocation Lists published to S3.

- rule: an enabled CRL requires s3_bucket_name and expiration_in_days - AWS publishes the signed list to that bucket on that cadence

### spec.revocation.crl.enabled

`bool`

Publish CRLs for this CA. The bucket must grant the ACM PCA
service the documented getBucketAcl/putObject permissions
BEFORE the CA is created - AWS validates at create.

### spec.revocation.crl.expirationInDays

`int64`

Days each published CRL is valid before AWS publishes the next
one (1-5000).

- rule: expiration_in_days must be between 1 and 5000

### spec.revocation.crl.s3BucketName

`string | valueFrom`

The S3 bucket CRLs publish to. Reference an AwsS3Bucket
bucket_id output or pass a literal bucket name.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.revocation.crl.s3ObjectAcl

`string`

The objects' ACL: "PUBLIC_READ" (relying parties fetch the CRL
straight from S3) or "BUCKET_OWNER_FULL_CONTROL" (private
bucket - pair with custom_cname fronting it, e.g. CloudFront).
Unset takes AWS's default (PUBLIC_READ - which requires the
bucket to allow public ACLs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PUBLIC_READ","BUCKET_OWNER_FULL_CONTROL"]}}

### spec.revocation.crl.customCname

`string`

The hostname baked into issued certificates' CRL distribution
point instead of the raw S3 URL (a CDN or vanity host fronting
the bucket).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"253"}}

### spec.revocation.crl.customPath

`string`

A path prefix for the published objects within the bucket.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"253"}}

### spec.revocation.ocsp

`AwsPrivateCaOcsp`

Online Certificate Status Protocol - AWS runs the responder.

### spec.revocation.ocsp.enabled

`bool`

Enable OCSP - issued certificates then carry the responder URL
and relying parties check revocation online.

### spec.revocation.ocsp.customCname

`string`

A vanity hostname for the OCSP responder (you run the
DNS/proxy; AWS still answers).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"253"}}

### spec.rootCaValidity

`AwsPrivateCaValidity`

The validity of a ROOT CA's self-signed certificate. Unset means
10 years (the console's default posture). ROOT only. Read at
activation time ONLY: the issue request is the certificate's
birth certificate (AWS never returns it), so both engines ignore
later edits - changing this after activation does nothing.
Reissuing a CA certificate is an operational act via ACM PCA,
never a manifest edit.

- rule: validity value must be a positive integer for DAYS/MONTHS/YEARS/ABSOLUTE, or an RFC3339 timestamp (e.g. "2036-01-01T00:00:00Z") for END_DATE

### spec.rootCaValidity.type

`string`

How the period is expressed: "YEARS", "MONTHS", "DAYS" (relative
to issuance), "ABSOLUTE" (Unix epoch seconds), or "END_DATE"
(an RFC3339 timestamp).

- rule: {"string":{"in":["YEARS","MONTHS","DAYS","ABSOLUTE","END_DATE"]}}

### spec.rootCaValidity.value

`string` · required

The period's value, per type: "10" with YEARS, "398" with DAYS,
"2036-01-01T00:00:00Z" with END_DATE.

- rule: {"string":{"minLen":"1"}}

### spec.subordinateActivation

`AwsPrivateCaSubordinateActivation`

Activate a SUBORDINATE CA from a parent AwsPrivateCa in this
account: the modules issue this CA's certificate from the parent
(a CA-path-length template) and install it. Omit to leave the
subordinate in PENDING_CERTIFICATE for out-of-band activation by
an external parent.

### spec.subordinateActivation.parentCaArn

`string | valueFrom` · required

The parent CA that signs this one. Reference another
AwsPrivateCa certificate_authority_arn output or pass a literal
ARN. The parent must be ACTIVE and in this account.

- references: AwsPrivateCa (`status.outputs.certificate_authority_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsPrivateCa, name: <that resource's name>, fieldPath: status.outputs.certificate_authority_arn}} -- a bare string does not parse

### spec.subordinateActivation.pathLength

`int64`

How many MORE levels of CA may hang below this one (0-3; picks
the SubordinateCACertificate_PathLen{N} template). 0 - the
common case - means this CA signs only end-entity certificates.

- rule: path_length must be between 0 and 3 - AWS templates cover path lengths 0 through 3

### spec.subordinateActivation.validity

`AwsPrivateCaValidity` · required

The subordinate CA certificate's validity. Must end before the
parent's own certificate expires. Read at activation time ONLY
(the birth-certificate contract - see root_ca_validity): both
engines ignore later edits.

- rule: {"required":true}
- rule: validity value must be a positive integer for DAYS/MONTHS/YEARS/ABSOLUTE, or an RFC3339 timestamp (e.g. "2036-01-01T00:00:00Z") for END_DATE

### spec.subordinateActivation.validity.type

`string`

How the period is expressed: "YEARS", "MONTHS", "DAYS" (relative
to issuance), "ABSOLUTE" (Unix epoch seconds), or "END_DATE"
(an RFC3339 timestamp).

- rule: {"string":{"in":["YEARS","MONTHS","DAYS","ABSOLUTE","END_DATE"]}}

### spec.subordinateActivation.validity.value

`string` · required

The period's value, per type: "10" with YEARS, "398" with DAYS,
"2036-01-01T00:00:00Z" with END_DATE.

- rule: {"string":{"minLen":"1"}}

### spec.issuedCertificates

`[]AwsPrivateCaIssuedCertificate`

Certificates issued directly from this CA at apply time, keyed
by name - the declarative surface for long-lived internal
certificates (service mTLS pairs, MSK client certs). Bring your
own CSR (the private key never touches AWS or this manifest).
The CA must be able to issue: a ROOT or activated SUBORDINATE.

### spec.issuedCertificates[].name

`string` · required

The certificate's name - the for_each key on both engines and
the key in the issued_certificate_arns output map.

- rule: {"string":{"minLen":"1","maxLen":"63","pattern":"^[a-z0-9]([a-z0-9-]*[a-z0-9])?$"}}

### spec.issuedCertificates[].csr

`string` · required

The PEM certificate signing request. Generate it wherever the
private key lives (openssl req -new ...) - the key itself never
touches AWS or this manifest.

- rule: csr must be a PEM-encoded certificate signing request (-----BEGIN CERTIFICATE REQUEST-----)
- rule: {"string":{"minLen":"1"}}

### spec.issuedCertificates[].signingAlgorithm

`string`

The signing algorithm for THIS certificate. Must match the CA
key's family (same rule as the CA's own signing_algorithm).

- rule: {"string":{"in":["SHA256WITHRSA","SHA384WITHRSA","SHA512WITHRSA","SHA256WITHECDSA","SHA384WITHECDSA","SHA512WITHECDSA","ML_DSA_44","ML_DSA_65","ML_DSA_87","SM3WITHSM2"]}}

### spec.issuedCertificates[].validity

`AwsPrivateCaValidity` · required

How long the certificate is valid. A SHORT_LIVED_CERTIFICATE
CA caps this at 7 days.

- rule: {"required":true}
- rule: validity value must be a positive integer for DAYS/MONTHS/YEARS/ABSOLUTE, or an RFC3339 timestamp (e.g. "2036-01-01T00:00:00Z") for END_DATE

### spec.issuedCertificates[].validity.type

`string`

How the period is expressed: "YEARS", "MONTHS", "DAYS" (relative
to issuance), "ABSOLUTE" (Unix epoch seconds), or "END_DATE"
(an RFC3339 timestamp).

- rule: {"string":{"in":["YEARS","MONTHS","DAYS","ABSOLUTE","END_DATE"]}}

### spec.issuedCertificates[].validity.value

`string` · required

The period's value, per type: "10" with YEARS, "398" with DAYS,
"2036-01-01T00:00:00Z" with END_DATE.

- rule: {"string":{"minLen":"1"}}

### spec.issuedCertificates[].templateArn

`string`

The AWS certificate template shaping the certificate's X.509
profile (key usages, path length). Unset means
EndEntityCertificate/V1 - a plain TLS certificate. Templates are
partition-scoped static ARNs like
"arn:aws:acm-pca:::template/EndEntityClientAuthCertificate/V1".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:[a-z0-9-]+:acm-pca:::template/.+$"}}

### spec.issuedCertificates[].apiPassthrough

`string`

Dynamic X.509 extensions (JSON ApiPassthrough document - SANs,
custom key usages) for templates in the ApiPassthrough family.

- rule: api_passthrough must be a valid JSON ApiPassthrough document
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.acmRenewalPermission

`bool`

Grant ACM permission to auto-renew the certificates it requested
from this CA (the IssueCertificate/GetCertificate/
ListPermissions grant to acm.amazonaws.com). Set this whenever
AwsCertManagerCert resources are issued from this CA - without
it their renewals fail silently at expiry.

### spec.policy

`string`

The CA's resource policy (JSON) - grants OTHER accounts or an
organization the right to issue from this CA via RAM sharing.

- rule: policy must be a valid JSON policy document
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.permanentDeletionTimeInDays

`int64`

Days the deleted CA stays restorable, 7-30. Unset means 30 (the
provider default). Billing stops at delete either way; a shorter
window frees the CA's name-and-subject slot sooner. E2E and
ephemeral environments should use 7. FIXED AT CREATE: AWS stores
nothing (only the delete call consumes it) and the provider
cannot apply a window-only change (an empty update the API
rejects), so both engines read this at create and ignore later
edits - the delete uses the create-time window.

- rule: permanent_deletion_time_in_days must be between 7 and 30 (unset defaults to 30)

### spec.enabled

`bool` · optional (explicit presence)

Whether the CA accepts issue requests. Set false to PAUSE
issuance without deleting (existing certificates stay valid and
revocation keeps publishing; billing continues). Only an ACTIVE
CA can be disabled.

- default: `true`

## Validation Rules

- `spec.signing_algorithm_matches_key_family`: signing_algorithm must match key_algorithm's family - RSA_* keys sign *WITHRSA, EC_* keys sign *WITHECDSA, ML_DSA_* keys sign with their exact ML_DSA level, SM2 signs SM3WITHSM2
- `spec.root_ca_validity_root_only`: root_ca_validity applies only when type is ROOT - a subordinate's certificate validity comes from its parent via subordinate_activation.validity
- `spec.subordinate_activation_subordinate_only`: subordinate_activation applies only when type is SUBORDINATE - a ROOT CA self-activates
- `spec.short_lived_forbids_revocation`: SHORT_LIVED_CERTIFICATE mode does not support revocation - drop the crl/ocsp configuration (short-lived certificates expire within 7 days instead of being revoked)
- `spec.issued_certificate_names_unique`: issued certificate names must be unique within the CA

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsPrivateCa, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.certificate_authority_arn` | `string` | The CA's ARN - the join key every consumer uses (MSK's TLS client-auth CA list, AwsCertManagerCert's issuing CA, a subordinate's parent reference). |
| `status.outputs.certificate_authority_id` | `string` | The CA's id (the UUID tail of the ARN). |
| `status.outputs.ca_certificate` | `string` | The CA's certificate, PEM - what relying parties add to their trust stores. |
| `status.outputs.ca_certificate_chain` | `string` | The CA's certificate chain, PEM (empty for a ROOT - it IS the anchor; a subordinate chains up to its root). |
| `status.outputs.ca_csr` | `string` | The CA's certificate signing request, PEM - what an EXTERNAL parent CA signs when activating a subordinate out of band. |
| `status.outputs.issued_certificate_arns` | `map<string, string>` | Issued certificates' ARNs keyed by each entry's name - each value is that certificate's import ID. |
| `status.outputs.activation_certificate_arn` | `string` | The ARN of the CA's own activation certificate (the root's self-signed certificate, or the parent-issued subordinate certificate) - empty until activated. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.revocation.crl.s3BucketName` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.subordinateActivation.parentCaArn` | AwsPrivateCa | `status.outputs.certificate_authority_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsCertManagerCert | `spec.certificateAuthorityArn` | `status.outputs.certificate_authority_arn` |
| AwsMskCluster | `spec.authentication.tlsCertificateAuthorityArns` | `status.outputs.certificate_authority_arn` |
| AwsPrivateCa | `spec.subordinateActivation.parentCaArn` | `status.outputs.certificate_authority_arn` |

## See Also

- [Overview](../README.md)
