---
title: "Default Route to Lambda"
description: "This preset creates the simplest possible HTTP API Gateway: a single catch-all route (`$default`) that forwards all requests to a Lambda function. This is the ideal starting point for serverless APIs..."
type: "preset"
rank: "01"
presetSlug: "01-default-route-lambda"
componentSlug: "http-api-gateway"
componentTitle: "HTTP API Gateway"
provider: "aws"
icon: "package"
order: 1
---

# Default Route to Lambda

This preset creates the simplest possible HTTP API Gateway: a single catch-all route (`$default`) that forwards all requests to a Lambda function. This is the ideal starting point for serverless APIs where a single Lambda handles all routing logic internally.

## When to Use

- Simple serverless APIs with a single Lambda function
- Prototyping and development
- APIs where routing logic lives in the Lambda function
- Starting point before adding more routes

## Key Configuration Choices

- **`$default` route key** — Catches all unmatched requests, allowing the Lambda to handle routing internally
- **AWS_PROXY integration** — Lambda proxy integration optimized for HTTP APIs
- **Payload format version 2.0** — Defaults to 2.0 for simplified event structure (no base64 encoding, direct body access)
- **No authorization** — Public API (add authorization as needed)
- **Auto-deploy stage** — Changes deploy immediately (default `$default` stage)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<lambda-function-arn>` | ARN of the Lambda function to invoke | AWS Lambda console or `AwsLambda` status outputs (`status.outputs.function_arn`) |

## Lambda Function Requirements

Your Lambda function should handle all routes and return HTTP responses:

```javascript
exports.handler = async (event) => {
  const route = event.routeKey;
  const path = event.rawPath;
  
  // Handle different routes
  if (path === '/users') {
    return {
      statusCode: 200,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ users: [...] })
    };
  }
  
  // Default response
  return {
    statusCode: 404,
    body: JSON.stringify({ error: "Not found" })
  };
};
```

## Using Resource References

Instead of hardcoding the Lambda ARN, reference an `AwsLambda` resource:

```yaml
integration_uri:
  valueFrom:
    kind: AwsLambda
    name: "my-function"
    fieldPath: "status.outputs.function_arn"
```

## Related Presets

- **02-multi-route-lambda** — Use instead when you want API Gateway to handle routing with multiple routes
- **03-jwt-authorized-api** — Use instead when you need JWT authentication
