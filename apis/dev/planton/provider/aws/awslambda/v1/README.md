# AWSLambda

AWS Lambda lets you run code without provisioning or managing servers. This resource models an AWS Lambda function with common settings (runtime/handler or container image, memory/timeout, environment, VPC networking, layers, and encryption) and exposes observable outputs like function ARN and log group name.

## Spec fields (summary)
- function_name: Name of the function (unique per account/region)
- description: Free-text description for the function
- role_arn: Execution role as value-or-reference (defaults to AwsIamRole → status.outputs.role_arn)
- runtime: Language/runtime for zip/S3 code (ignored for container image)
- handler: Entrypoint for zip/S3 code (ignored for container image)
- memory_mb: Memory in MB (128–10240)
- timeout_seconds: Max execution time in seconds (1–900)
- reserved_concurrency: -1 (unreserved pool), 0 (disabled), or positive integer
- environment: Key/value environment variables
- subnets: Value-or-reference to VPC subnets (defaults to AwsVpc)
- security_groups: Value-or-reference to security groups (defaults to AwsSecurityGroup → status.outputs.security_group_id)
- architecture: X86_64 or ARM64
- layer_arns: Value-or-reference to Lambda layers (no defaults yet)
- kms_key_arn: Value-or-reference to KMS key (defaults to AwsKmsKey → status.outputs.key_arn)
- code_source_type: S3 or IMAGE
- s3: S3 bucket/key (and optional version) for zip package
- image_uri: ECR image URI for container-based code

## Stack outputs
- function_arn: Full ARN of the function
- function_name: Final function name
- log_group_name: CloudWatch Logs group (e.g., /aws/lambda/<name>)
- role_arn: Execution role ARN
- layer_arns: Attached layer ARNs

## How it works
This resource is orchestrated by the Planton CLI as part of a stack-update. The CLI validates your manifest, generates stack inputs, and invokes IaC backends in this repo:
- Pulumi (Go modules under iac/pulumi)
- Terraform (modules under iac/tf)

Credentials and region live in stack input (provider_credential), not in the spec.

## References
- AWS Lambda: https://docs.aws.amazon.com/lambda/latest/dg/welcome.html
- Lambda runtimes: https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html
- Container images: https://docs.aws.amazon.com/lambda/latest/dg/images-create.html
- VPC config: https://docs.aws.amazon.com/lambda/latest/dg/configuration-vpc.html
- Environment variables & encryption: https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars.html
