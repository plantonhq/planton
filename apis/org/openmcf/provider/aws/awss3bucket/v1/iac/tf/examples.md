# AWS S3 Bucket Examples

Below are several examples demonstrating how to define an AWS S3 Bucket component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic Private S3 Bucket

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: my-private-bucket
spec:
  isPublic: false
  region: "us-east-1"
```

This example creates a basic private S3 bucket:
• Private access with all public access blocked.
• Located in us-east-1 region.
• Default ownership controls (BucketOwnerPreferred).
• Suitable for secure data storage.

---

## Public S3 Bucket for Static Website

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: my-public-website-bucket
spec:
  isPublic: true
  region: "us-west-2"
```

This example creates a public S3 bucket:
• Public read access enabled.
• Located in us-west-2 region.
• Suitable for static website hosting.
• Public ACL applied for web access.

---

## S3 Bucket for Application Data

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: app-data-bucket
spec:
  isPublic: false
  region: "eu-west-1"
```

This example creates a private application data bucket:
• Private access for application data.
• Located in eu-west-1 region.
• Secure storage for application assets.
• Suitable for user uploads and application data.

---

## S3 Bucket for Backup Storage

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: backup-storage-bucket
spec:
  isPublic: false
  region: "us-east-1"
```

This example creates a backup storage bucket:
• Private access for secure backup storage.
• Located in us-east-1 region.
• Suitable for database backups and archives.
• Secure storage for critical data.

---

## S3 Bucket for Media Storage

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: media-storage-bucket
spec:
  isPublic: true
  region: "us-east-1"
```

This example creates a public media storage bucket:
• Public read access for media files.
• Located in us-east-1 region.
• Suitable for images, videos, and public media.
• CDN-friendly configuration.

---

## S3 Bucket for Log Storage

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: log-storage-bucket
spec:
  isPublic: false
  region: "us-west-2"
```

This example creates a log storage bucket:
• Private access for log files.
• Located in us-west-2 region.
• Suitable for application and server logs.
• Secure storage for audit trails.

---

## S3 Bucket for Development Environment

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: dev-assets-bucket
spec:
  isPublic: false
  region: "us-east-1"
```

This example creates a development assets bucket:
• Private access for development assets.
• Located in us-east-1 region.
• Suitable for development and testing data.
• Secure storage for development resources.

---

## S3 Bucket for Production Environment

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: prod-assets-bucket
spec:
  isPublic: false
  region: "us-east-1"
```

This example creates a production assets bucket:
• Private access for production assets.
• Located in us-east-1 region.
• Suitable for production application data.
• Secure storage for business-critical assets.

---

## S3 Bucket for Content Distribution

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: cdn-content-bucket
spec:
  isPublic: true
  region: "us-east-1"
```

This example creates a CDN content bucket:
• Public read access for CDN integration.
• Located in us-east-1 region.
• Suitable for CloudFront distribution origin.
• Optimized for content delivery networks.

---

## S3 Bucket for Analytics Data

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: analytics-data-bucket
spec:
  isPublic: false
  region: "us-east-1"
```

This example creates an analytics data bucket:
• Private access for analytics data.
• Located in us-east-1 region.
• Suitable for data lake and analytics storage.
• Secure storage for business intelligence data.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the S3 bucket is active via the AWS console or by
using the AWS CLI:

```shell
aws s3 ls s3://<your-bucket-name>
```

For detailed bucket information:

```shell
aws s3api get-bucket-location --bucket <your-bucket-name>
```

To check bucket access control:

```shell
aws s3api get-bucket-acl --bucket <your-bucket-name>
```

This will show the S3 bucket details including location, access control, and configuration information.
