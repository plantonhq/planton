# AWS CloudFront Examples

Below are several examples demonstrating how to define an AWS CloudFront component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic CloudFront Distribution

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: basic-cdn
spec:
  region: us-east-1
  enabled: true
  priceClass: PRICE_CLASS_100
  origins:
    - domainName: my-bucket.s3.amazonaws.com
      isDefault: true
  defaultRootObject: index.html
```

This example creates a basic CloudFront distribution:
• Uses S3 bucket as origin with custom domain name.
• Sets price class to PRICE_CLASS_100 for cost optimization.
• Configures index.html as the default root object.
• Marks the origin as default using the isDefault flag.

---

## CloudFront with Custom Domain

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: custom-domain-cdn
spec:
  region: us-east-1
  enabled: true
  aliases:
    - cdn.example.com
    - static.example.com
  certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abcd1234-5678-efgh-ijkl-123456abcdef
  priceClass: PRICE_CLASS_200
  origins:
    - domainName: my-bucket.s3.amazonaws.com
      isDefault: true
  defaultRootObject: index.html
```

This example includes custom domain configuration:
• Uses custom domain aliases (cdn.example.com, static.example.com).
• Requires ACM certificate ARN in us-east-1 region.
• Sets price class to PRICE_CLASS_200 for broader edge location coverage.
• Marks the origin as default using the isDefault flag.

---

## CloudFront with Multiple Origins

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: multi-origin-cdn
spec:
  region: us-east-1
  enabled: true
  priceClass: PRICE_CLASS_ALL
  origins:
    - domainName: my-bucket.s3.amazonaws.com
      isDefault: true
    - domainName: api.example.com
      originPath: /v1
      isDefault: false
    - domainName: cdn.example.com
      isDefault: false
  defaultRootObject: index.html
```

This example demonstrates multiple origins:
• S3 bucket as primary origin for static content (marked as default).
• API origin with custom path for backend services.
• CDN origin for additional content sources.
• Uses PRICE_CLASS_ALL for maximum edge location coverage.
• Exactly one origin must be marked as default.

---

## CloudFront for Static Website

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: static-website-cdn
spec:
  region: us-east-1
  enabled: true
  aliases:
    - www.example.com
    - example.com
  certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/website-cert-1234
  priceClass: PRICE_CLASS_100
  origins:
    - domainName: website-bucket.s3-website-us-east-1.amazonaws.com
      isDefault: true
  defaultRootObject: index.html
```

This example is optimized for static websites:
• Uses S3 website endpoint as origin.
• Configures both apex and www domain aliases.
• Includes ACM certificate for HTTPS support.
• Marks the origin as default using the isDefault flag.

---

## Minimal CloudFront Distribution

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: minimal-cdn
spec:
  region: us-east-1
  enabled: true
  origins:
    - domainName: my-bucket.s3.amazonaws.com
      isDefault: true
```

A minimal configuration with:
• Only required fields specified.
• Uses default CloudFront certificate (no custom domain).
• Default price class and cache behavior.
• Marks the origin as default using the isDefault flag.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the CloudFront distribution is active via the AWS console or by
using the AWS CLI:

```shell
aws cloudfront list-distributions
```

You should see your new CloudFront distribution with its domain name (e.g., d123.cloudfront.net) and any custom aliases configured.


