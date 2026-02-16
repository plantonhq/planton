# GcpKmsKey — Examples

## Example 1: Standard CMEK Key with Auto-Rotation

The most common use case: a symmetric encryption key for customer-managed encryption
across GCP services like BigQuery, CloudSQL, GCS, and Spanner.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKey
metadata:
  name: prod-cmek-key
spec:
  keyRingId:
    value: "projects/my-prod-project/locations/us-central1/keyRings/prod-encryption"
  keyName: cmek-data-key
  rotationPeriod: "7776000s"  # 90 days
```

**Notes:**
- `purpose` defaults to `ENCRYPT_DECRYPT` when not specified
- Algorithm defaults to `GOOGLE_SYMMETRIC_ENCRYPTION` with `SOFTWARE` protection
- The 90-day rotation period automatically creates a new key version and sets it as primary

## Example 2: HSM-Protected Key for Compliance

For environments requiring FIPS 140-2 Level 3 compliance, use Cloud HSM to protect
the key material in dedicated hardware security modules.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKey
metadata:
  name: compliance-cmek-key
spec:
  keyRingId:
    value: "projects/my-prod-project/locations/us-central1/keyRings/compliance-keys"
  keyName: hsm-cmek-key
  rotationPeriod: "7776000s"  # 90 days
  versionTemplate:
    algorithm: GOOGLE_SYMMETRIC_ENCRYPTION
    protectionLevel: HSM
```

**Notes:**
- HSM keys are significantly more expensive than software keys
- Protection level is immutable -- you cannot switch between HSM and SOFTWARE after creation
- The key ring must be in a region that supports Cloud HSM

## Example 3: Asymmetric Signing Key

For signing artifacts, JWTs, or verifying code integrity. Asymmetric keys do not
support automatic rotation -- key versions are managed explicitly.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKey
metadata:
  name: artifact-signer
spec:
  keyRingId:
    value: "projects/my-prod-project/locations/us-central1/keyRings/signing-keys"
  keyName: artifact-signing-key
  purpose: ASYMMETRIC_SIGN
  versionTemplate:
    algorithm: EC_SIGN_P256_SHA256
```

**Notes:**
- `EC_SIGN_P256_SHA256` provides a good balance of security and performance
- For stronger security, consider `EC_SIGN_P384_SHA384` or RSA-based algorithms
- No `rotationPeriod` -- asymmetric key version rotation is handled manually

## Example 4: Cross-Resource Reference (Infra Chart Pattern)

When composing resources in an infra chart, use `valueFrom` to reference the
key ring from another OpenMCF resource instead of hardcoding the path.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKey
metadata:
  name: data-cmek-key
spec:
  keyRingId:
    valueFrom:
      kind: GcpKmsKeyRing
      name: prod-key-ring
      fieldPath: status.outputs.key_ring_id
  keyName: data-encryption-key
  rotationPeriod: "7776000s"
  destroyScheduledDuration: "2592000s"  # 30 days (explicit)
```

**Notes:**
- `valueFrom` creates a dependency edge -- the key ring is provisioned before this key
- The key ring's `key_ring_id` output provides the fully qualified path automatically
- `destroyScheduledDuration` of 30 days is the GCP default; shown here for documentation

## Example 5: Key Without Initial Version

For advanced workflows where key material will be imported from an external source
or versions will be created separately.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKey
metadata:
  name: import-ready-key
spec:
  keyRingId:
    value: "projects/my-prod-project/locations/global/keyRings/import-keys"
  keyName: external-import-key
  skipInitialVersionCreation: true
```

**Notes:**
- The key is created but has no active versions -- it cannot encrypt or sign until
  a version is created
- Useful when key material comes from an external HSM or key management system
- Versions must be created using `google_kms_crypto_key_version` or an import job
