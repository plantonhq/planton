```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: example
spec:
  region: us-west-2
```

CLI:

```bash
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

# Examples

## Minimal manifest (YAML)
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: my-ec2
spec:
  region: us-west-2
  instanceName: web-1
  amiId: ami-0123456789abcdef0
  instanceType: t3.small
  subnetId:
    value: subnet-aaa111
  securityGroupIds:
    - value: sg-000111222
  connectionMethod: SSM
  iamInstanceProfileArn:
    value: arn:aws:iam::123456789012:instance-profile/ssm
  rootVolumeSizeGb: 30
  tags:
    env: prod
```

## Bastion/SSH access (YAML)
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: my-ec2-ssh
spec:
  region: us-west-2
  instanceName: web-ssh
  amiId: ami-0123456789abcdef0
  instanceType: t3.small
  subnetId:
    value: subnet-aaa111
  securityGroupIds:
    - value: sg-000111222
  connectionMethod: BASTION
  keyName: my-keypair
  rootVolumeSizeGb: 40
  tags:
    env: staging
```

## CLI flows
- Validate:
```bash
openmcf validate --manifest ../hack/manifest.yaml
```

- Pulumi preview:
```bash
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .
```

- Pulumi update (apply):
```bash
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir . \
  --yes
```

- Pulumi refresh:
```bash
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .
```

- Pulumi destroy:
```bash
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .
```

Note: Provider credentials are supplied via stack input, not in the spec.
