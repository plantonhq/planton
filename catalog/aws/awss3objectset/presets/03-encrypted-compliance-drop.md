# Encrypted Compliance Drop

This preset writes an audit artifact under WORM (write-once-read-many) retention: customer-managed KMS encryption, a SHA256 upload checksum, and COMPLIANCE-mode Object Lock that nobody — including the account root — can shorten. The target bucket must have been created with Object Lock enabled (an immutable create-time bucket setting on `AwsS3Bucket`).

## When to Use

- Storing audit reports, attestations, or evidence artifacts that regulations require to be tamper-proof
- Pinning a specific artifact to a customer-managed KMS key that differs from the bucket's default encryption
- Any object whose integrity must be independently verifiable (the stored SHA256 checksum is retrievable via GetObjectAttributes)

## Key Configuration Choices

- **`objectLockMode: COMPLIANCE`** -- The object version genuinely cannot be deleted until the retain-until date passes; destroy fails until then. Use `GOVERNANCE` instead if privileged principals should be able to override.
- **Per-object KMS override** (`serverSideEncryption: aws:kms` + `kmsKey` reference) -- Diverges from the bucket's default encryption for this artifact only; the deploying principal needs `kms:GenerateDataKey` on the key
- **`bucketKeyEnabled: true`** -- Batches KMS requests to cut KMS costs for the encrypted object
- **`checksumAlgorithm: SHA256`** -- Stores a full-object checksum for compliance regimes that mandate SHA-based integrity proofs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region where the bucket lives | Must match the bucket's region |
| `<object-lock-enabled-bucket-name>` | Bucket created with Object Lock enabled | AWS S3 console (Object Lock is a create-time setting) |
| `<report-name>` | Artifact identifier used in the key and content | Your audit workflow |
| `<kms-key-resource-name>` | Name of the `AwsKmsKey` resource whose ARN encrypts the object | Your infra stack's resource list |
| `objectLockRetainUntilDate` (example value) | RFC 3339 timestamp until which the version is retained — replace `2033-01-01T00:00:00Z` with your retention deadline | Your retention policy |
