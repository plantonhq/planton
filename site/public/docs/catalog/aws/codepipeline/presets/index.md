---
title: "Presets"
description: "Ready-to-deploy configuration presets for CodePipeline"
type: "preset-list"
componentSlug: "codepipeline"
componentTitle: "CodePipeline"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-github-source-codebuild"
    rank: "01"
    title: "GitHub Source + CodeBuild (CI Pipeline)"
    excerpt: "This preset creates a V2 pipeline that fetches source code from a GitHub repository via CodeStar Connection and runs a CodeBuild project. A git push trigger automatically executes the pipeline when..."
  - slug: "02-ecr-ecs-deploy"
    rank: "02"
    title: "ECR Source + ECS Deploy (Container Deployment Pipeline)"
    excerpt: "This preset creates a V2 pipeline that triggers when a new Docker image is pushed to an ECR repository, runs a CodeBuild step to prepare the deployment bundle (generating `imagedefinitions.json`),..."
  - slug: "03-s3-lambda-deploy"
    rank: "03"
    title: "S3 Source + Lambda Deploy (Serverless Deployment Pipeline)"
    excerpt: "This preset creates a V2 pipeline that triggers when a deployment package is uploaded to S3 and invokes a Lambda function to perform the deployment. This is the standard serverless deployment..."
---

# CodePipeline Presets

Ready-to-deploy configuration presets for CodePipeline. Each preset is a complete manifest you can copy, customize, and deploy.
