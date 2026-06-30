# AWS HTTP API Gateway

The **AwsHttpApiGateway** component provides a declarative way to deploy AWS API Gateway HTTP APIs (v2) with bundled stages, routes, integrations, and optional authorizers. HTTP APIs are optimized for Lambda proxy and HTTP proxy integrations, offering lower latency and cost compared to REST APIs.

## Overview

AWS API Gateway HTTP APIs (API Gateway v2) are designed for building low-latency, cost-effective REST APIs and HTTP proxy APIs. They support:

- **Lambda proxy integration** — Direct invocation of Lambda functions with automatic request/response transformation
- **HTTP proxy integration** — Forward requests to upstream HTTP endpoints
- **JWT authorization** — Native integration with Cognito, Auth0, or any OIDC provider
- **Lambda REQUEST authorizers** — Custom authorization logic via Lambda functions
- **Automatic deployments** — Changes to routes and integrations are automatically deployed to the stage
- **Native CORS support** — Built-in CORS configuration without custom integration responses

This component bundles the API, a single stage, routes with inline integrations, and optional authorizers into one declarative resource. The underlying IaC modules create and wire together the necessary API Gateway resources automatically.

## When to Use

Use **AwsHttpApiGateway** when you need to:

- Expose Lambda functions as HTTP endpoints
- Create REST APIs with Lambda backend
- Proxy HTTP requests to upstream services
- Implement JWT-based authentication with Cognito or Auth0
- Build cost-effective APIs (HTTP APIs are up to 70% cheaper than REST APIs)
- Deploy APIs with automatic stage deployments

**When not to use:**

- WebSocket APIs (use `AwsWebSocketApiGateway` instead)
- APIs requiring advanced features like API keys, usage plans, or request/response transformations (use REST APIs)
- APIs needing custom domain configuration (not supported in v1)

## Prerequisites

- **AWS credentials** configured via environment variables, IAM instance profile, or Planton provider config
- **AWS region** specified in provider config or environment
- **Lambda functions** (for AWS_PROXY integrations) or **HTTP endpoints** (for HTTP_PROXY integrations) already deployed
- **Cognito User Pool** (for JWT authorization) or **Lambda authorizer function** (for REQUEST authorization) if using authorizers

## Quick Start

Create a minimal HTTP API with a single route to a Lambda function:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsHttpApiGateway
metadata:
  name: my-api
spec:
  routes:
    - route_key: "$default"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:my-function"
```

Deploy using Planton:

```bash
planton apply -f api.yaml
```

## Spec Fields

### AwsHttpApiGatewaySpec

The root specification for the HTTP API Gateway.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | `string` | No | Human-readable description of the API (max 1024 characters) |
| `cors_configuration` | `AwsHttpApiGatewayCorsConfig` | No | CORS configuration for cross-origin requests |
| `disable_execute_api_endpoint` | `bool` | No | Disable the default execute-api endpoint (set to true when using custom domains) |
| `stage` | `AwsHttpApiGatewayStageConfig` | No | Stage configuration (defaults to "$default" with auto_deploy=true) |
| `routes` | `AwsHttpApiGatewayRoute[]` | **Yes** | API routes mapping request patterns to backend integrations (at least one required) |
| `authorizers` | `AwsHttpApiGatewayAuthorizer[]` | No | Named authorizers referenced by routes |

### AwsHttpApiGatewayCorsConfig

Configures cross-origin resource sharing (CORS) for the API.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `allow_origins` | `string[]` | No | Origins allowed to make cross-origin requests (e.g., "https://example.com", "*") |
| `allow_methods` | `string[]` | No | HTTP methods allowed for cross-origin requests (e.g., "GET", "POST", "OPTIONS") |
| `allow_headers` | `string[]` | No | Request headers allowed in cross-origin requests (e.g., "Content-Type", "Authorization") |
| `expose_headers` | `string[]` | No | Response headers exposed to the browser in cross-origin responses |
| `max_age_seconds` | `int32` | No | Maximum time in seconds browsers can cache CORS preflight results (0-86400) |
| `allow_credentials` | `bool` | No | Whether the API supports credentials (cookies, authorization headers) in cross-origin requests |

### AwsHttpApiGatewayStageConfig

Configures the deployment stage for the API.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | No | Stage name (defaults to "$default" when empty) |
| `auto_deploy` | `bool` | No | Enable automatic deployment when routes/integrations change (defaults to true for "$default" stage) |
| `access_log` | `AwsHttpApiGatewayAccessLogConfig` | No | Access logging configuration for CloudWatch Logs |
| `default_throttle` | `AwsHttpApiGatewayThrottleConfig` | No | Default throttling settings applied to all routes |
| `stage_variables` | `map<string, string>` | No | Stage variables passed to integrations |

### AwsHttpApiGatewayAccessLogConfig

Configures access logging for the API stage.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `destination_arn` | `StringValueOrRef` | **Yes** | CloudWatch Log Group ARN for access log delivery |
| `format` | `string` | **Yes** | Log format template using API Gateway access log variables |

### AwsHttpApiGatewayThrottleConfig

Configures request throttling for the API stage.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `burst_limit` | `int32` | No | Maximum number of concurrent requests allowed (burst) |
| `rate_limit` | `double` | No | Steady-state request rate limit (requests per second) |

### AwsHttpApiGatewayRoute

Maps a request pattern to a backend integration.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `route_key` | `string` | **Yes** | Route key defining the request pattern (e.g., "GET /users", "POST /orders/{id}", "$default") |
| `integration` | `AwsHttpApiGatewayIntegration` | **Yes** | Backend integration that processes requests matching this route |
| `authorization_type` | `string` | No | Authorization type: "NONE" (default), "JWT", or "AWS_IAM" |
| `authorizer_name` | `string` | No | Name of the authorizer to use (required when authorization_type is "JWT") |
| `authorization_scopes` | `string[]` | No | OAuth 2.0 scopes required for JWT authorization |

### AwsHttpApiGatewayIntegration

Defines the backend target for a route.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `integration_type` | `string` | **Yes** | Integration type: "AWS_PROXY" (Lambda) or "HTTP_PROXY" (HTTP endpoint) |
| `integration_uri` | `StringValueOrRef` | **Yes** | Integration URI: Lambda function ARN for AWS_PROXY, HTTP URL for HTTP_PROXY |
| `payload_format_version` | `string` | No | Payload format version: "2.0" (recommended) or "1.0" (defaults to "2.0" when empty) |
| `integration_method` | `string` | No | HTTP method for integration request (defaults to route's HTTP method for HTTP_PROXY) |
| `timeout_milliseconds` | `int32` | No | Integration timeout in milliseconds (50-30000, default: 30000) |

### AwsHttpApiGatewayAuthorizer

Defines a named authorizer referenced by routes.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | **Yes** | Unique name for this authorizer (1-128 characters) |
| `authorizer_type` | `string` | **Yes** | Authorizer type: "JWT" or "REQUEST" |
| `jwt_configuration` | `AwsHttpApiGatewayJwtConfig` | No | JWT configuration (required when authorizer_type is "JWT") |
| `authorizer_uri` | `StringValueOrRef` | No | Lambda function URI for REQUEST authorizers (required when authorizer_type is "REQUEST") |
| `authorizer_credentials_arn` | `StringValueOrRef` | No | IAM role ARN that API Gateway assumes to invoke Lambda authorizer |
| `identity_sources` | `string[]` | No | Identity sources used to extract authorization token (e.g., "$request.header.Authorization") |
| `result_ttl_seconds` | `int32` | No | Time in seconds API Gateway caches authorizer result (0-3600, default: 300 for REQUEST) |
| `enable_simple_responses` | `bool` | No | Enable simple boolean responses from Lambda authorizers |
| `authorizer_payload_format_version` | `string` | No | Payload format version for Lambda authorizer event: "2.0" (recommended) or "1.0" |

### AwsHttpApiGatewayJwtConfig

Configures JWT validation for a JWT authorizer.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `issuer` | `string` | **Yes** | Token issuer URL (e.g., Cognito: "https://cognito-idp.{region}.amazonaws.com/{userPoolId}") |
| `audiences` | `string[]` | No | Expected audiences (e.g., Cognito app client ID) |

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `api_id` | `string` | The API Gateway API identifier |
| `api_endpoint` | `string` | The default endpoint URL: `https://{api-id}.execute-api.{region}.amazonaws.com` |
| `api_arn` | `string` | The Amazon Resource Name (ARN) of the API |
| `execution_arn` | `string` | The execution ARN prefix: `arn:aws:execute-api:{region}:{account-id}:{api-id}` |
| `stage_invoke_url` | `string` | The invoke URL for the deployed stage |
| `stage_name` | `string` | The name of the deployed stage (e.g., "$default") |

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsHttpApiGateway
metadata:
  name: simple-api
spec:
  routes:
    - route_key: "$default"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:my-function"
```

## Production-Ready Example

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsHttpApiGateway
metadata:
  name: production-api
spec:
  description: Production API for user management
  cors_configuration:
    allow_origins:
      - "https://app.example.com"
    allow_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allow_headers:
      - "Content-Type"
      - "Authorization"
    allow_credentials: true
    max_age_seconds: 3600
  stage:
    name: "$default"
    auto_deploy: true
    access_log:
      destination_arn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/apigateway/production-api"
      format: '{"requestId":"$context.requestId","ip":"$context.identity.sourceIp","method":"$context.httpMethod","path":"$context.routeKey","status":"$context.status","latency":"$context.responseLatency"}'
    default_throttle:
      burst_limit: 5000
      rate_limit: 2000.0
  routes:
    - route_key: "GET /users"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "get-users-function"
            fieldPath: "status.outputs.function_arn"
        payload_format_version: "2.0"
        timeout_milliseconds: 5000
    - route_key: "POST /users"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      authorization_scopes:
        - "users:write"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "create-user-function"
            fieldPath: "status.outputs.function_arn"
        payload_format_version: "2.0"
    - route_key: "GET /health"
      authorization_type: "NONE"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:health-check"
  authorizers:
    - name: "cognito-authorizer"
      authorizer_type: "JWT"
      jwt_configuration:
        issuer: "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123XYZ"
        audiences:
          - "1a2b3c4d5e6f7g8h9i0j"
      identity_sources:
        - "$request.header.Authorization"
```

## Related Components

- [AwsLambda](/docs/catalog/aws/awslambda) — Lambda functions used as backend integrations
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — IAM roles for Lambda authorizers
- [AwsCloudwatchLogGroup](/docs/catalog/aws/awscloudwatchloggroup) — CloudWatch Log Groups for access logging
- [AwsWebSocketApiGateway](/docs/catalog/aws/awswebsocketapigateway) — WebSocket APIs (separate component)

## Additional Resources

- [AWS API Gateway HTTP API Documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api.html)
- [Planton Documentation](https://planton.dev/docs)
- See `examples.md` for more detailed examples
- See `docs/README.md` for architecture deep-dive
