# Examples

## Minimal manifest (YAML)
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: my-cdn
spec:
  region: us-east-1
  enabled: true
  priceClass: PRICE_CLASS_100
  origins:
    - domainName: bucket.s3.amazonaws.com
      isDefault: true
  defaultRootObject: index.html
```

## Multiple origins manifest (YAML)
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: my-cdn
spec:
  region: us-east-1
  enabled: true
  aliases:
    - cdn.example.com
  certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012
  priceClass: PRICE_CLASS_100
  origins:
    - domainName: bucket.s3.amazonaws.com
      originPath: /assets
      isDefault: true
    - domainName: api.example.com
      isDefault: false
  defaultRootObject: index.html
```

## CLI flows
- Validate:
```bash
openmcf validate --manifest ./manifest.yaml
```

- Pulumi deploy:
```bash
openmcf pulumi up --manifest ./manifest.yaml --stack org/project/stack
```

- Terraform deploy:
```bash
openmcf terraform apply --manifest ./manifest.yaml --stack org/project/stack
```

Note: Provider credentials are supplied via stack input, not in the spec.
