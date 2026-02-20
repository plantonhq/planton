# OciKmsKey Examples

## AES-256 Key (Default HSM)

A 256-bit AES symmetric key in an HSM — the most common choice for encrypting Block Volumes, Object Storage, and databases:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: encryption-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciKmsKey.encryption-key
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  managementEndpoint:
    value: "https://xxx-management.kms.us-ashburn-1.oraclecloud.com"
  keyShape:
    algorithm: aes
    length: 32
```

## RSA-4096 Signing Key with Auto-Rotation

A 4096-bit RSA key for signing and verification, with automatic rotation every 90 days starting on a specific date. References vault and compartment via `valueFrom`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: signing-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciKmsKey.signing-key
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-security
      fieldPath: status.outputs.compartmentId
  managementEndpoint:
    valueFrom:
      kind: OciKmsVault
      name: prod-vault
      fieldPath: status.outputs.managementEndpoint
  keyShape:
    algorithm: rsa
    length: 512
  isAutoRotationEnabled: true
  autoKeyRotationDetails:
    rotationIntervalInDays: 90
    timeOfScheduleStart: "2026-04-01T00:00:00Z"
```

## ECDSA P-256 Key with Software Protection

An ECDSA P-256 key for lightweight digital signatures with software-based protection:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: ecdsa-p256
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciKmsKey.ecdsa-p256
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  managementEndpoint:
    value: "https://xxx-management.kms.us-ashburn-1.oraclecloud.com"
  keyShape:
    algorithm: ecdsa
    length: 32
    curveId: nist_p256
  protectionMode: software
```

## External Key (BYOK/EKMS)

A key backed by a third-party HSM for regulatory BYOK requirements. The vault must be of type `external`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: byok-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciKmsKey.byok-key
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  managementEndpoint:
    value: "https://xxx-management.kms.us-ashburn-1.oraclecloud.com"
  keyShape:
    algorithm: aes
    length: 32
  protectionMode: external
  externalKeyReference:
    externalKeyId: "ekm-key-uuid-12345"
```

## Common Operations

### Enable auto-rotation on an existing key

Set `isAutoRotationEnabled: true` and optionally provide `autoKeyRotationDetails` with `rotationIntervalInDays`. Re-apply the manifest. OCI begins rotating the key on the configured schedule.

### Retrieve the key OCID for use by other resources

After deploying the key, other OpenMCF resources (OciBlockVolume, OciObjectStorageBucket) can reference it via `valueFrom`:

```yaml
kmsKeyId:
  valueFrom:
    kind: OciKmsKey
    name: encryption-key
    fieldPath: status.outputs.keyId
```

## Best Practices

1. **Use AES-256 for data-at-rest encryption** — symmetric keys are faster and suitable for block/object/database encryption.
2. **Use RSA or ECDSA for signing** — asymmetric keys when you need sign/verify operations.
3. **Use HSM protection for production** — FIPS 140-2 Level 3 compliance; use software only when cost is a primary concern.
4. **Enable auto-rotation for compliance** — many regulatory frameworks require periodic key rotation.
5. **Use `valueFrom` to reference vaults** — avoids hardcoding management endpoints and ensures correct dependency ordering.
