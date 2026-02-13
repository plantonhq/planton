# AWS Lambda

Deploys an AWS Lambda function with automatic CloudWatch log group creation, supporting both S3 zip and container image deployment models. The component handles VPC networking, environment variables, KMS encryption, and layer attachment.

## What Gets Created

When you deploy an AwsLambda resource, OpenMCF provisions:

- **Lambda Function** — an `aws_lambda_function` resource configured with the specified runtime, handler, memory, timeout, and code source (S3 zip or ECR container image)
- **CloudWatch Log Group** — a `/aws/lambda/<function_name>` log group with 30-day retention for capturing function execution logs
- **VPC Configuration** — created only when `subnets` or `securityGroups` are provided, attaches ENIs in the specified subnets with the given security groups to allow access to private VPC resources

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An IAM execution role** with the `AWSLambdaBasicExecutionRole` policy (add `AWSLambdaVPCAccessExecutionRole` if using VPC configuration)
- **A deployment artifact** — either a zip archive uploaded to an S3 bucket, or a container image pushed to ECR
- **VPC subnets and security groups** if the function needs access to private resources (e.g., RDS, ElastiCache)

## Quick Start

Create a file `lambda.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsLambda
metadata:
  name: my-function
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsLambda.my-function
spec:
  functionName: my-function
  roleArn: arn:aws:iam::123456789012:role/lambda-exec-role
  codeSourceType: CODE_SOURCE_TYPE_S3
  runtime: nodejs18.x
  handler: index.handler
  s3:
    bucket: my-deploy-bucket
    key: functions/my-function.zip
```

Deploy:

```shell
openmcf apply -f lambda.yaml
```

This creates a Node.js Lambda function using code from the specified S3 bucket, along with a CloudWatch log group for its output.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `functionName` | `string` | Human-readable function name shown in the AWS Console. Must be unique per account/region. | Min length 1 |
| `roleArn` | `StringValueOrRef` | Execution role ARN for the function. Can reference an AwsIamRole resource via `valueFrom`. | Required |
| `codeSourceType` | `enum` | How function code is supplied. Valid values: `CODE_SOURCE_TYPE_S3`, `CODE_SOURCE_TYPE_IMAGE`. | Required |
| `runtime` | `string` | Language runtime identifier (e.g., `nodejs18.x`, `python3.11`, `java21`, `provided.al2`). Required when `codeSourceType` is `CODE_SOURCE_TYPE_S3`. | Conditional |
| `handler` | `string` | Function entrypoint (e.g., `index.handler`, `module.function`, `package.Class::method`). Required when `codeSourceType` is `CODE_SOURCE_TYPE_S3`. | Conditional |
| `s3` | `object` | S3 location of the zip deployment package. Required when `codeSourceType` is `CODE_SOURCE_TYPE_S3`. | Conditional |
| `s3.bucket` | `string` | S3 bucket name containing the deployment package. | Min length 1 |
| `s3.key` | `string` | S3 object key (path) to the zip archive. | Min length 1 |
| `imageUri` | `string` | ECR image URI (e.g., `123456789012.dkr.ecr.us-east-1.amazonaws.com/repo:tag`). Required when `codeSourceType` is `CODE_SOURCE_TYPE_IMAGE`. | Conditional |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Free-form description visible in the AWS Console. |
| `memoryMb` | `int32` | AWS default (128) | Memory allocation in MB. CPU and network scale with this value. Range: 128–10240. |
| `timeoutSeconds` | `int32` | AWS default (3) | Maximum execution time per invocation in seconds. Range: 1–900. |
| `reservedConcurrency` | `int32` | — | Concurrent execution limit. Omit or set to `-1` for the unreserved account pool, `0` to disable invocations, or a positive integer to reserve that many slots. |
| `environment` | `map<string, string>` | `{}` | Key/value environment variables available at runtime. Encrypted at rest when `kmsKeyArn` is set. |
| `subnets` | `StringValueOrRef[]` | `[]` | Subnet IDs for Lambda ENIs. Provide at least two across different AZs for high availability. Can reference AwsVpc resources via `valueFrom`. |
| `securityGroups` | `StringValueOrRef[]` | `[]` | Security group IDs attached to Lambda ENIs. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `architecture` | `enum` | `ARCHITECTURE_UNSPECIFIED` | CPU architecture. Valid values: `ARCHITECTURE_UNSPECIFIED`, `X86_64`, `ARM64`. |
| `layerArns` | `StringValueOrRef[]` | `[]` | Layer ARNs to include. Up to five layers; order matters. |
| `kmsKeyArn` | `StringValueOrRef` | — | Customer-managed KMS key ARN to encrypt environment variables at rest. Can reference an AwsKmsKey resource via `valueFrom`. |
| `s3.objectVersion` | `string` | — | S3 object version to pin a specific artifact when bucket versioning is enabled. |

## Examples

### Container Image Lambda

Deploy a Lambda function from a container image stored in ECR:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsLambda
metadata:
  name: image-function
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsLambda.image-function
spec:
  functionName: image-function
  roleArn: arn:aws:iam::123456789012:role/lambda-exec-role
  codeSourceType: CODE_SOURCE_TYPE_IMAGE
  imageUri: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-func:latest
  architecture: ARM64
  memoryMb: 512
  timeoutSeconds: 30
```

### VPC-Connected Lambda with Environment Variables

A function that accesses private VPC resources such as an RDS database:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsLambda
metadata:
  name: vpc-function
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsLambda.vpc-function
spec:
  functionName: vpc-function
  roleArn: arn:aws:iam::123456789012:role/lambda-vpc-role
  codeSourceType: CODE_SOURCE_TYPE_S3
  runtime: python3.11
  handler: app.handler
  s3:
    bucket: deploy-artifacts
    key: functions/vpc-function-v2.zip
  memoryMb: 256
  timeoutSeconds: 60
  environment:
    DB_HOST: mydb.cluster-abc123.us-east-1.rds.amazonaws.com
    DB_NAME: appdb
  subnets:
    - subnet-private-az1
    - subnet-private-az2
  securityGroups:
    - sg-lambda-vpc
```

### Production Lambda with KMS Encryption and Layers

Full-featured configuration with encrypted environment variables, layers, reserved concurrency, and pinned S3 artifact version:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsLambda
metadata:
  name: prod-function
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsLambda.prod-function
spec:
  functionName: prod-function
  description: Production order processing function
  roleArn: arn:aws:iam::123456789012:role/prod-lambda-role
  codeSourceType: CODE_SOURCE_TYPE_S3
  runtime: java21
  handler: com.example.OrderHandler::handleRequest
  s3:
    bucket: prod-deploy-artifacts
    key: functions/order-processor-1.4.0.zip
    objectVersion: abc123def456
  architecture: ARM64
  memoryMb: 2048
  timeoutSeconds: 300
  reservedConcurrency: 50
  environment:
    STAGE: production
    TABLE_NAME: orders
  kmsKeyArn: arn:aws:kms:us-east-1:123456789012:key/mrk-abc123
  layerArns:
    - arn:aws:lambda:us-east-1:123456789012:layer:common-utils:3
    - arn:aws:lambda:us-east-1:123456789012:layer:monitoring:7
  subnets:
    - subnet-prod-az1
    - subnet-prod-az2
  securityGroups:
    - sg-prod-lambda
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding ARNs and IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsLambda
metadata:
  name: ref-function
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsLambda.ref-function
spec:
  functionName: ref-function
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: lambda-role
      field: status.outputs.role_arn
  codeSourceType: CODE_SOURCE_TYPE_S3
  runtime: nodejs18.x
  handler: index.handler
  s3:
    bucket: deploy-bucket
    key: functions/ref-function.zip
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: lambda-key
      field: status.outputs.key_arn
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[1].id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: lambda-sg
        field: status.outputs.security_group_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `function_arn` | `string` | Full ARN of the Lambda function |
| `function_name` | `string` | Final name of the Lambda function |
| `log_group_name` | `string` | CloudWatch Logs log group name (e.g., `/aws/lambda/<function_name>`) |
| `role_arn` | `string` | Execution role ARN that the function assumes |
| `layer_arns` | `string[]` | Layer ARNs attached to the function (present only when layers are configured) |

## Related Components

- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides the execution role for the function
- [AwsVpc](/docs/catalog/aws/awsvpc) — provides subnets for VPC-connected functions
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — controls network access for Lambda ENIs
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — encrypts environment variables at rest
- [AwsS3Bucket](/docs/catalog/aws/awss3bucket) — hosts the deployment package for zip-based code
