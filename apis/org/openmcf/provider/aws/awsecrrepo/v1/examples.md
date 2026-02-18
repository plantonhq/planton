# Create using CLI

Create a YAML file using the examples shown below. After the YAML file is created, use the following command to apply:

```shell
planton apply -f <yaml-path>
```

# Basic Example

This basic example creates an AWS ECR Repository with secure defaults: immutable tags, image scanning enabled, and AES256 encryption.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: my-basic-repo
spec:
  region: us-east-1
  repository_name: my-app/backend
  image_immutable: true
```

# Example with Lifecycle Policy

This example adds lifecycle policies for cost control - expires untagged images after 7 days and keeps only the last 50 images.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: my-app-with-lifecycle
spec:
  region: us-east-1
  repository_name: my-app/frontend
  image_immutable: true
  scan_on_push: true
  lifecycle_policy:
    expire_untagged_after_days: 7
    max_image_count: 50
```

# Example with KMS Encryption

This example uses customer-managed KMS encryption for compliance requirements.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: my-compliant-repo
spec:
  region: us-east-1
  repository_name: my-org/secure-service
  image_immutable: true
  scan_on_push: true
  encryption_type: KMS
  kms_key_id: arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012
  lifecycle_policy:
    expire_untagged_after_days: 14
    max_image_count: 30
```

# Example with Environment Variables

This example uses environment variables to parameterize the ECR configuration.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: ${REPO_NAME}
spec:
  region: us-east-1
  repository_name: ${ORG_NAME}/${SERVICE_NAME}
  image_immutable: ${IMAGE_IMMUTABLE}
  scan_on_push: ${SCAN_ON_PUSH}
  encryption_type: ${ENCRYPTION_TYPE}
  lifecycle_policy:
    expire_untagged_after_days: ${EXPIRE_UNTAGGED_DAYS}
    max_image_count: ${MAX_IMAGE_COUNT}
```

In this example, replace the placeholders like `${REPO_NAME}` with your actual environment variable names or values.

# Example with Environment Secrets

The below example assumes that the secrets are managed by Planton Cloud's [AWS Secrets Manager](https://buf.build/openmcf/apis/docs/main:ai.planton.code2cloud.v1.aws.awssecretsmanager) deployment module.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: my-secret-repo
spec:
  region: us-east-1
  repository_name: my-org/confidential-service
  image_immutable: true
  scan_on_push: true
  encryption_type: KMS
  kms_key_id: ${awssm-my-org-prod-aws-secrets.ecr-kms-key-arn}
  lifecycle_policy:
    expire_untagged_after_days: 14
    max_image_count: 30
```

In this example:

- **kms_key_id** is retrieved from AWS Secrets Manager.
- The value before the dot (`awssm-my-org-prod-aws-secrets`) is the ID of the AWS Secrets Manager resource on Planton Cloud.
- The value after the dot (`ecr-kms-key-arn`) is the name of the secret within that resource.

# Production-Ready Example

This comprehensive example demonstrates production best practices:
- Immutable tags for stability
- Image scanning enabled for security
- KMS encryption for compliance
- Lifecycle policies for cost control

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: production-service-repo
spec:
  region: us-east-1
  repository_name: my-company/production-api
  image_immutable: true
  scan_on_push: true
  encryption_type: KMS
  kms_key_id: arn:aws:kms:us-east-1:123456789012:key/prod-ecr-key
  force_delete: false
  lifecycle_policy:
    expire_untagged_after_days: 1
    max_image_count: 100
```

# Development Environment Example

This example is suitable for development environments with more lenient policies:
- Mutable tags for flexibility
- Shorter retention for faster cleanup
- AES256 encryption (no KMS required)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: dev-service-repo
spec:
  region: us-east-1
  repository_name: my-company/dev-api
  image_immutable: false
  scan_on_push: true
  encryption_type: AES256
  force_delete: true
  lifecycle_policy:
    expire_untagged_after_days: 3
    max_image_count: 20
```

---

These examples illustrate various configurations of the `AwsEcrRepo` API resource, demonstrating how to define ECR repositories with different security postures, lifecycle policies, and encryption configurations.

**Key Production Recommendations:**
- Always enable `image_immutable: true` for production to prevent tag overwrites
- Keep `scan_on_push: true` (default) for security vulnerability detection
- Use lifecycle policies to control costs (untagged images can accumulate quickly)
- Use `encryption_type: KMS` when compliance requires auditable key management
- Set `force_delete: false` (default) to prevent accidental repository deletion

Please ensure that you replace placeholder values like repository names, KMS key ARNs, and environment variable names with your actual configuration details.
