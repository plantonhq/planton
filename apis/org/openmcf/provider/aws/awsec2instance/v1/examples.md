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
openmcf validate --manifest ./manifest.yaml
```

- Pulumi deploy:
```bash
openmcf pulumi update --manifest ./manifest.yaml --stack <org>/<project>/<stack> --module-dir <path> --yes
```

- Terraform deploy:
```bash
openmcf tofu apply --manifest ./manifest.yaml --auto-approve
```

Note: Provider credentials are supplied via stack input, not in the spec.
