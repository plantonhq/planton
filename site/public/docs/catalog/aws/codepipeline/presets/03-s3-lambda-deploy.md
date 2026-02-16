---
title: "S3 Source + Lambda Deploy (Serverless Deployment Pipeline)"
description: "This preset creates a V2 pipeline that triggers when a deployment package is uploaded to S3 and invokes a Lambda function to perform the deployment. This is the standard serverless deployment..."
type: "preset"
rank: "03"
presetSlug: "03-s3-lambda-deploy"
componentSlug: "codepipeline"
componentTitle: "CodePipeline"
provider: "aws"
icon: "package"
order: 3
---

# S3 Source + Lambda Deploy (Serverless Deployment Pipeline)

This preset creates a V2 pipeline that triggers when a deployment package is uploaded to S3 and invokes a Lambda function to perform the deployment. This is the standard serverless deployment pipeline for teams that package Lambda functions externally and want automated deployments when a new package appears in S3.

## When to Use

- Serverless applications deployed as Lambda function packages
- Teams that build and package Lambda functions in an external CI system (GitHub Actions, Jenkins, etc.) and upload the zip to S3
- Simple deployment flows where a deployer Lambda function handles the update logic (updating function code, publishing versions, shifting aliases)
- Lightweight pipelines that don't need a build stage

## Key Configuration Choices

- **S3 source** (`provider: S3`) — Pipeline triggers when the specified S3 object is updated. For automatic detection, either enable `PollForSourceChanges: "true"` or configure a CloudTrail + EventBridge rule (recommended) to detect S3 `PutObject` events
- **Lambda invoke** (`category: Invoke`, `provider: Lambda`) — Invokes a deployer Lambda function that receives the artifact and performs the deployment logic. The `UserParameters` string is passed to the function's event payload
- **SUPERSEDED execution** — For serverless deploys, the latest package should always win; SUPERSEDED mode ensures only the newest version is deployed
- **PollForSourceChanges: false** — Polling is disabled by default. Configure an EventBridge rule for the S3 bucket to trigger the pipeline efficiently (see AWS documentation for S3 source actions)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<codepipeline-service-role-arn>` | IAM role ARN with permissions for S3 and Lambda | AWS IAM console or `AwsIamRole` status outputs |
| `<pipeline-artifacts-bucket-name>` | S3 bucket name for pipeline artifact storage | AWS S3 console or `AwsS3Bucket` status outputs |
| `<lambda-packages-bucket>` | S3 bucket containing the Lambda deployment packages | AWS S3 console or `AwsS3Bucket` status outputs |
| `<path/to/function-package.zip>` | S3 object key for the Lambda deployment package (e.g., `releases/my-function/latest.zip`) | Your CI pipeline's upload path |
| `<deployer-lambda-function-name>` | Name of the Lambda function that performs the deployment (updates target function code, publishes version, shifts alias) | AWS Lambda console or `AwsLambda` status outputs |
| `<target-lambda-function-name>` | Name of the Lambda function being deployed (passed via `UserParameters`) | AWS Lambda console or `AwsLambda` status outputs |

## Related Presets

- **01-github-source-codebuild** — Use when source is a GitHub repo and you need a build stage
- **02-ecr-ecs-deploy** — Use for container deployments to ECS instead of Lambda
