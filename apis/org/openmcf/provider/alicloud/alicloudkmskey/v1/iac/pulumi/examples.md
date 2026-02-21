# Examples

Create a YAML file using one of the examples below. Deploy with the OpenMCF Pulumi CLI:

```shell
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Minimal KMS Key

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKmsKey
metadata:
  name: my-key
spec:
  region: cn-hangzhou
```

## Production Encryption Key with Rotation

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKmsKey
metadata:
  name: prod-encryption-key
  org: my-org
  env: production
spec:
  region: cn-shanghai
  description: Production master encryption key
  automaticRotation: true
  rotationInterval: "365d"
  deletionProtection: true
  pendingWindowInDays: 30
  tags:
    team: security
```

## Asymmetric Signing Key

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKmsKey
metadata:
  name: signing-key
spec:
  region: cn-hangzhou
  description: RSA signing key for API verification
  keySpec: RSA_2048
  keyUsage: SIGN/VERIFY
  tags:
    purpose: signing
```
