# AWS ECR Repository Examples

Below are several examples demonstrating how to define an AWS ECR Repository component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic ECR Repository

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: basic-ecr-repo
spec:
  region: us-east-1
  repositoryName: my-app
```

This example creates a basic ECR repository:
• Uses default encryption (AES256).
• Mutable image tags (can be overwritten).
• No force delete protection.
• Simple repository name for easy identification.

---

## ECR Repository with Immutable Tags

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: immutable-ecr-repo
spec:
  region: us-east-1
  repositoryName: production-app
  imageImmutable: true
  encryptionType: AES256
```

This example enforces immutable tags:
• Prevents image tags from being overwritten.
• Uses AWS-managed encryption (AES256).
• Suitable for production environments.
• Ensures image version consistency.

---

## ECR Repository with KMS Encryption

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: kms-encrypted-ecr-repo
spec:
  region: us-east-1
  repositoryName: secure-app
  imageImmutable: true
  encryptionType: KMS
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abcd1234-5678-efgh-ijkl-123456abcdef
```

This example uses customer-managed KMS encryption:
• Uses custom KMS key for encryption.
• Immutable tags for security.
• Compliant with strict security requirements.
• Full control over encryption keys.

---

## ECR Repository with Force Delete

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: force-delete-ecr-repo
spec:
  region: us-west-2
  repositoryName: development-app
  imageImmutable: false
  forceDelete: true
  encryptionType: AES256
```

This example allows force deletion:
• Can delete repository even with images.
• Mutable tags for development flexibility.
• Useful for development/testing environments.
• Automatically removes all images on deletion.

---

## Production ECR Repository

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: production-ecr-repo
spec:
  region: us-east-1
  repositoryName: production-microservice
  imageImmutable: true
  encryptionType: KMS
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/production-ecr-key
  forceDelete: false
```

This example is production-ready:
• Immutable tags prevent accidental overwrites.
• Customer-managed KMS encryption.
• Force delete disabled for data protection.
• Descriptive repository name for organization.

---

## Development ECR Repository

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: development-ecr-repo
spec:
  region: us-west-2
  repositoryName: dev-app
  imageImmutable: false
  encryptionType: AES256
  forceDelete: true
```

This example is optimized for development:
• Mutable tags for iterative development.
• AWS-managed encryption for simplicity.
• Force delete enabled for easy cleanup.
• Simple naming convention.

---

## Multi-Environment ECR Repository

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: multi-env-ecr-repo
spec:
  region: us-east-1
  repositoryName: github.com/myorg/myapp
  imageImmutable: true
  encryptionType: AES256
  forceDelete: false
```

This example uses organization naming:
• Follows GitHub-style naming convention.
• Immutable tags for consistency.
• Standard encryption for most use cases.
• Force delete disabled for safety.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the ECR repository is active via the AWS console or by
using the AWS CLI:

```shell
aws ecr describe-repositories --repository-names <your-repository-name>
```

You should see your new ECR repository with its configuration details, including repository URL, ARN, and encryption settings.
Use the repository URL to push and pull Docker images to/from your ECR repository.

