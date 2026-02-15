---
title: "Zip-Based Lambda Function"
description: "This preset creates a Lambda function deployed from a zip archive stored in S3. It uses the Node.js 18.x runtime with 256 MB memory and a 30-second timeout. This is the most common Lambda deployment..."
type: "preset"
rank: "01"
presetSlug: "01-zip-basic"
componentSlug: "lambda"
componentTitle: "Lambda"
provider: "aws"
icon: "package"
order: 1
---

# Zip-Based Lambda Function

This preset creates a Lambda function deployed from a zip archive stored in S3. It uses the Node.js 18.x runtime with 256 MB memory and a 30-second timeout. This is the most common Lambda deployment model for lightweight event handlers, API endpoints, and automation scripts.

## When to Use

- Event-driven functions triggered by API Gateway, S3, SQS, SNS, or CloudWatch Events
- Lightweight API endpoints or webhook handlers
- Functions written in Node.js, Python, Java, Go, .NET, or Ruby using a standard runtime

## Key Configuration Choices

- **S3 code source** (`codeSourceType: CODE_SOURCE_TYPE_S3`) -- Deployment package (zip) is stored in S3; your CI/CD pipeline uploads the zip and updates the key
- **Node.js 18.x** (`runtime: nodejs18.x`) -- Change to `python3.11`, `java21`, `go1.x`, `dotnet8`, or `provided.al2` for other languages
- **256 MB memory** (`memoryMb: 256`) -- CPU scales proportionally; increase for compute-intensive functions
- **30-second timeout** (`timeoutSeconds: 30`) -- Suitable for API handlers; increase up to 900 seconds for long-running processes
- **No VPC** -- Function runs in Lambda's managed network; add `subnets` and `securityGroups` if VPC access is needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<function-name>` | Lambda function name (must be unique per account/region) | Your function naming convention |
| `<lambda-execution-role-arn>` | IAM role ARN with `AWSLambdaBasicExecutionRole` policy | AWS IAM console or `AwsIamRole` status outputs |
| `<deployment-bucket>` | S3 bucket containing the deployment package | Your CI/CD pipeline configuration |
| `<deployment-package-key>` | S3 object key for the zip file (e.g., `functions/my-function/v1.0.0.zip`) | Your CI/CD pipeline configuration |

## Related Presets

- **02-container-basic** -- Use instead when deploying Lambda from a container image (for custom runtimes or larger dependencies)
