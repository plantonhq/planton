# GcpBinaryAuthorizationAttestor

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBinaryAuthorizationAttestorSpec is a Binary Authorization attestor
(`google_binary_authorization_attestor`) with the Artifact Analysis note
its attestations are stored under (`google_container_analysis_note`).

An attestor is a trusted signer: a build or QA pipeline signs an image
digest with a private key and records the signature (an attestation) as
an occurrence of the attestor's note; at deploy time a
GcpBinaryAuthorizationPolicy rule that requires the attestor admits the
image only if one of the attestor's public keys verifies such a
signature. The signatures themselves are written by the pipelines, not by
this block.

The note: set note to have this block create the attestor's own note in
its project (Google's recommended shape -- one note per attestor), or
attestation_authority_note.note_reference to use a note that already
exists. Exactly one. For a created note the module also grants the
attestor's service account (delegation_service_account_email)
roles/containeranalysis.notes.occurrences.viewer on it, which Google
requires before the attestor can read attestations. For a referenced
note, grant it yourself.

A policy in another project needs roles/binaryauthorization.attestorsVerifier
on the attestor for that project's Binary Authorization service agent.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBinaryAuthorizationAttestor
metadata:
  name: built-by-ci
spec:
  description: Images built from main by the CI pipeline
  note:
    humanReadableName: CI build pipeline
  attestationAuthorityNote:
    publicKeys:
      - pkixPublicKey:
          kmsKeyVersion:
            value: projects/my-gcp-project/locations/global/keyRings/binauthz/cryptoKeys/ci-signer/cryptoKeyVersions/1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.attestorName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.note` | `GcpBinaryAuthorizationAttestorNote` |  |  |  |
| `spec.note.noteName` | `string` |  |  |  |
| `spec.note.humanReadableName` | `string` | yes |  |  |
| `spec.note.shortDescription` | `string` |  |  |  |
| `spec.note.longDescription` | `string` |  |  |  |
| `spec.note.expirationTime` | `string` |  |  |  |
| `spec.note.relatedNoteNames` | `[]string` |  |  |  |
| `spec.note.relatedUrl` | `[]GcpBinaryAuthorizationAttestorNoteRelatedUrl` |  |  |  |
| `spec.note.relatedUrl[].url` | `string` | yes |  |  |
| `spec.note.relatedUrl[].label` | `string` |  |  |  |
| `spec.attestationAuthorityNote` | `GcpBinaryAuthorizationAttestorAuthorityNote` |  |  |  |
| `spec.attestationAuthorityNote.noteReference` | `string` |  |  |  |
| `spec.attestationAuthorityNote.publicKeys` | `[]GcpBinaryAuthorizationAttestorPublicKey` |  |  |  |
| `spec.attestationAuthorityNote.publicKeys[].id` | `string` |  |  |  |
| `spec.attestationAuthorityNote.publicKeys[].comment` | `string` |  |  |  |
| `spec.attestationAuthorityNote.publicKeys[].asciiArmoredPgpPublicKey` | `string` |  |  |  |
| `spec.attestationAuthorityNote.publicKeys[].pkixPublicKey` | `GcpBinaryAuthorizationAttestorPkixPublicKey` |  |  |  |
| `spec.attestationAuthorityNote.publicKeys[].pkixPublicKey.publicKeyPem` | `string` |  |  |  |
| `spec.attestationAuthorityNote.publicKeys[].pkixPublicKey.signatureAlgorithm` | `string` |  |  |  |
| `spec.attestationAuthorityNote.publicKeys[].pkixPublicKey.kmsKeyVersion` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.initial_version_name`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the attestor (and a created note) lives in: a literal
project ID or a GcpProject reference. Empty means the provider's
default project. The module enables binaryauthorization.googleapis.com
and containeranalysis.googleapis.com there.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.attestorName

`string`

The attestor's ID, unique in its project, e.g. "built-by-ci" or
"qa-approved". Defaults to metadata.name. Immutable: a new ID replaces
the attestor, and policies naming the old one must be re-applied.

### spec.description

`string`

What the attestor vouches for; shown when choosing attestors.

### spec.note

`GcpBinaryAuthorizationAttestorNote`

The note to create for this attestor. Mutually exclusive with
attestation_authority_note.note_reference.

### spec.note.noteName

`string`

The note's ID in the project. Defaults to "{attestor_name}-note".
Immutable.

### spec.note.humanReadableName

`string` · required

The signer's name as people read it, e.g. "Production build pipeline"
(Google's attestation_authority.hint). Required.

- rule: {"required":true}

### spec.note.shortDescription

`string`

A one-sentence description of the note.

### spec.note.longDescription

`string`

A longer description: what signing by this attestor means.

### spec.note.expirationTime

`string`

When the note expires, as an RFC 3339 timestamp, e.g.
"2027-01-01T00:00:00Z". Empty: never.

### spec.note.relatedNoteNames

`[]string`

Other notes this one relates to: projects/{project}/notes/{note}.

- rule: {"repeated":{"unique":true}}

### spec.note.relatedUrl

`[]GcpBinaryAuthorizationAttestorNoteRelatedUrl`

Links to more information, e.g. the signing pipeline's runbook.

### spec.note.relatedUrl[].url

`string` · required

The URL. Required.

- rule: {"required":true}

### spec.note.relatedUrl[].label

`string`

The link's label.

### spec.attestationAuthorityNote

`GcpBinaryAuthorizationAttestorAuthorityNote`

The attestor's note reference and public keys.

### spec.attestationAuthorityNote.noteReference

`string`

An existing ATTESTATION_AUTHORITY note: projects/{project}/notes/{note},
or a bare note ID in the attestor's project. Leave empty when note
creates one. Immutable: a new reference replaces the attestor.

### spec.attestationAuthorityNote.publicKeys

`[]GcpBinaryAuthorizationAttestorPublicKey`

The public keys that verify attestations; one verifying key is enough.
Rotate by adding the new key, re-signing, then removing the old one.
Empty: the attestor never verifies anything, so every rule requiring
it denies.

- rule: set exactly one of ascii_armored_pgp_public_key or pkix_public_key
- rule: leave id empty for a PGP key: Google computes it from the key's fingerprint

### spec.attestationAuthorityNote.publicKeys[].id

`string`

The key's ID, which every signature must name exactly. For a PKIX key
it must be an RFC 3986 URI; empty takes Google's default (a digest of
the key), or for a Cloud KMS key the module sets
//cloudkms.googleapis.com/v1/{key version}, the ID gcloud and Google's
signing tools use. Leave empty for a PGP key.

### spec.attestationAuthorityNote.publicKeys[].comment

`string`

A note about the key, e.g. who holds the private half.

### spec.attestationAuthorityNote.publicKeys[].asciiArmoredPgpPublicKey

`string`

An ASCII-armored PGP public key, the whole output of
`gpg --export --armor signer@example.com`.

### spec.attestationAuthorityNote.publicKeys[].pkixPublicKey

`GcpBinaryAuthorizationAttestorPkixPublicKey`

A PKIX public key: a PEM, or a Cloud KMS signing key version.

- rule: set exactly one of public_key_pem or kms_key_version
- rule: signature_algorithm is required with public_key_pem and comes from the key version with kms_key_version

### spec.attestationAuthorityNote.publicKeys[].pkixPublicKey.publicKeyPem

`string`

A PEM-encoded public key (RFC 7468 SubjectPublicKeyInfo), for a key
pair held outside Cloud KMS.

### spec.attestationAuthorityNote.publicKeys[].pkixPublicKey.signatureAlgorithm

`string`

The algorithm signatures use with public_key_pem; it must match the
key. Google's names: EC_SIGN_P256_SHA256, EC_SIGN_P384_SHA384,
EC_SIGN_P521_SHA512 (the ECDSA_* names are the same algorithms),
RSA_SIGN_PKCS1_{2048,3072,4096}_SHA256, RSA_SIGN_PKCS1_4096_SHA512,
RSA_SIGN_PSS_{2048,3072,4096}_SHA256, RSA_SIGN_PSS_4096_SHA512 (and
their RSA_PSS_* aliases), and the post-quantum ML_DSA_65.

- rule: signature_algorithm must be one of Google's Binary Authorization signature algorithms

### spec.attestationAuthorityNote.publicKeys[].pkixPublicKey.kmsKeyVersion

`string | valueFrom`

A Cloud KMS asymmetric-sign key version whose public key verifies the
signatures: a GcpKmsKey reference (the version Google creates with the
key) or a literal
projects/*/locations/*/keyRings/*/cryptoKeys/*/cryptoKeyVersions/*.
Both modules read the version's public key and algorithm from Cloud
KMS, so the pipeline signs with the private half that never leaves
KMS. The key's purpose must be ASYMMETRIC_SIGN.

- references: GcpKmsKey (`status.outputs.initial_version_name`)
- rule: kms_key_version must be projects/*/locations/*/keyRings/*/cryptoKeys/*/cryptoKeyVersions/*
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.initial_version_name}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroying this block does, for the attestor and a created note:
  "" / "DELETE" -- both are deleted; policies naming the attestor stop
                   admitting images that only it signed
  "PREVENT"     -- destroy fails
  "ABANDON"     -- both leave management and stay in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `note_xor_note_reference`: set exactly one of note (create the attestor's note) or attestation_authority_note.note_reference (use an existing note)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBinaryAuthorizationAttestor, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.attestor_id` | `string` | Full resource name: projects/{project}/attestors/{attestor_name}. The form a GcpBinaryAuthorizationPolicy rule's require_attestations_by takes. |
| `status.outputs.attestor_name` | `string` | The attestor's ID in its project. |
| `status.outputs.note_reference` | `string` | The note attestations are stored under: projects/{project}/notes/{note}. Signing pipelines create their attestation occurrences against it. |
| `status.outputs.delegation_service_account_email` | `string` | The service account the attestor reads attestations with. It needs roles/containeranalysis.notes.occurrences.viewer on the note (granted by the module for a note it creates). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.attestationAuthorityNote.publicKeys[].pkixPublicKey.kmsKeyVersion` | GcpKmsKey | `status.outputs.initial_version_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBinaryAuthorizationPolicy | `spec.defaultAdmissionRule.requireAttestationsBy` | `status.outputs.attestor_id` |
| GcpBinaryAuthorizationPolicy | `spec.clusterAdmissionRules[].requireAttestationsBy` | `status.outputs.attestor_id` |

## See Also

- [Overview](../README.md)
