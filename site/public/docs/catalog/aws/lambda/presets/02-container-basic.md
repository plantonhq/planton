---
title: "Container-Based Lambda Function"
description: "This preset creates a Lambda function deployed from a container image in ECR. The runtime and handler are defined by the image's CMD/ENTRYPOINT, not by Lambda configuration. This is ideal for..."
type: "preset"
rank: "02"
presetSlug: "02-container-basic"
componentSlug: "lambda"
componentTitle: "Lambda"
provider: "aws"
icon: "package"
order: 2
---

# Container-Based Lambda Function

This preset creates a Lambda function deployed from a container image in ECR. The runtime and handler are defined by the image's CMD/ENTRYPOINT, not by Lambda configuration. This is ideal for functions with large dependencies, custom runtimes, or existing Docker-based build pipelines.

## When to Use

- Functions with dependencies exceeding the 250 MB zip limit (ML models, large SDKs, native binaries)
- Custom runtimes not available as Lambda managed runtimes (Rust, C++, custom interpreters)
- Teams with existing Docker build pipelines who want to reuse their container workflow
- Functions that need a consistent local development experience using Docker

## Key Configuration Choices

- **Container image code** (`codeSourceType: CODE_SOURCE_TYPE_IMAGE`) -- Runtime and handler are defined in the Docker image, not in Lambda config
- **512 MB memory** (`memoryMb: 512`) -- Container images typically need more memory; increase for ML or data processing workloads
- **60-second timeout** (`timeoutSeconds: 60`) -- Longer default since container functions often handle heavier workloads
- **No runtime/handler** -- These are defined in the container image's CMD/ENTRYPOINT and are ignored by Lambda
- **No VPC** -- Function runs in Lambda's managed network; add `subnets` and `securityGroups` if VPC access is needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<function-name>` | Lambda function name (must be unique per account/region) | Your function naming convention |
| `<lambda-execution-role-arn>` | IAM role ARN with `AWSLambdaBasicExecutionRole` policy | AWS IAM console or `AwsIamRole` status outputs |
| `<ecr-image-uri>` | ECR image URI (e.g., `123456789012.dkr.ecr.us-east-1.amazonaws.com/my-function:latest`) | AWS ECR console or `AwsEcrRepo` status outputs |

## Related Presets

- **01-zip-basic** -- Use instead for lightweight functions deployed from S3 zip archives
