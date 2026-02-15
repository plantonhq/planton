---
title: "Presets"
description: "Ready-to-deploy configuration presets for HTTP API Gateway"
type: "preset-list"
componentSlug: "http-api-gateway"
componentTitle: "HTTP API Gateway"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-default-route-lambda"
    rank: "01"
    title: "Default Route to Lambda"
    excerpt: "This preset creates the simplest possible HTTP API Gateway: a single catch-all route (`$default`) that forwards all requests to a Lambda function. This is the ideal starting point for serverless APIs..."
  - slug: "02-multi-route-lambda"
    rank: "02"
    title: "Multi-Route Lambda API with CORS"
    excerpt: "This preset creates an HTTP API Gateway with multiple routes to different Lambda functions, configured with CORS for browser access. API Gateway handles routing, allowing each Lambda to focus on a..."
  - slug: "03-jwt-authorized-api"
    rank: "03"
    title: "JWT Authorized API with Cognito"
    excerpt: "This preset creates an HTTP API Gateway with JWT authorization using Amazon Cognito User Pool. Routes can be public (no auth) or protected (require valid JWT token). Protected routes can also require..."
---

# HTTP API Gateway Presets

Ready-to-deploy configuration presets for HTTP API Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
