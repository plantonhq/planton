---
title: "Presets"
description: "Ready-to-deploy configuration presets for Lambda"
type: "preset-list"
componentSlug: "lambda"
componentTitle: "Lambda"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-zip-basic"
    rank: "01"
    title: "Zip-Based Lambda Function"
    excerpt: "This preset creates a Lambda function deployed from a zip archive stored in S3. It uses the Node.js 18.x runtime with 256 MB memory and a 30-second timeout. This is the most common Lambda deployment..."
  - slug: "02-container-basic"
    rank: "02"
    title: "Container-Based Lambda Function"
    excerpt: "This preset creates a Lambda function deployed from a container image in ECR. The runtime and handler are defined by the image's CMD/ENTRYPOINT, not by Lambda configuration. This is ideal for..."
---

# Lambda Presets

Ready-to-deploy configuration presets for Lambda. Each preset is a complete manifest you can copy, customize, and deploy.
