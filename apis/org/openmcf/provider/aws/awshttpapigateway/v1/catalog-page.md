# AWS HTTP API Gateway

Deploys an AWS API Gateway HTTP API (v2) with bundled stage, routes with inline integrations, and optional JWT/Lambda authorizers. HTTP APIs are optimized for Lambda proxy and HTTP proxy integrations, offering lower latency and cost compared to REST APIs.

## What Gets Created

When you deploy an AwsHttpApiGateway resource, OpenMCF provisions:

- **HTTP API** — an API Gateway v2 HTTP API with configured CORS, description, and endpoint settings
- **Stage** — a deployment stage (defaults to `$default` with auto-deploy enabled)
- **Routes** — API routes mapping request patterns (e.g., `GET /users`, `POST /orders/{id}`) to backend integrations
- **Integrations** — backend integrations (Lambda proxy or HTTP proxy) that process requests
- **Authorizers** — optional JWT or Lambda authorizers for request authorization

## Prerequisites

- **AWS credentials** configured via environment variables, IAM instance profile, or OpenMCF provider config
- **AWS region** specified in provider config or environment
- **Lambda functions** (for AWS_PROXY integrations) or **HTTP endpoints** (for HTTP_PROXY integrations) already deployed
- **Cognito User Pool** (for JWT authorization) or **Lambda authorizer function** (for REQUEST authorization) if using authorizers

## Quick Start

Create a file `api.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
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

Deploy:

```shell
openmcf apply -f api.yaml
```

This creates an HTTP API with a single catch-all route (`$default`) that forwards all requests to the specified Lambda function. The API is automatically deployed to the `$default` stage with auto-deploy enabled.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `routes` | `AwsHttpApiGatewayRoute[]` | API routes mapping request patterns to backend integrations (at least one required) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the API (max 1024 characters) |
| `cors_configuration` | `AwsHttpApiGatewayCorsConfig` | — | CORS configuration for cross-origin requests |
| `disable_execute_api_endpoint` | `bool` | `false` | Disable the default execute-api endpoint (set to true when using custom domains) |
| `stage` | `AwsHttpApiGatewayStageConfig` | `$default` with `auto_deploy=true` | Stage configuration for the deployed API |
| `authorizers` | `AwsHttpApiGatewayAuthorizer[]` | `[]` | Named authorizers referenced by routes |

### Route Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `route_key` | `string` | Yes | Route key defining the request pattern (e.g., `GET /users`, `POST /orders/{id}`, `$default`) |
| `integration` | `AwsHttpApiGatewayIntegration` | Yes | Backend integration that processes requests matching this route |
| `authorization_type` | `string` | No | Authorization type: `"NONE"` (default), `"JWT"`, or `"AWS_IAM"` |
| `authorizer_name` | `string` | No | Name of the authorizer to use (required when `authorization_type` is `"JWT"`) |
| `authorization_scopes` | `string[]` | No | OAuth 2.0 scopes required for JWT authorization |

### Integration Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `integration_type` | `string` | Yes | Integration type: `"AWS_PROXY"` (Lambda) or `"HTTP_PROXY"` (HTTP endpoint) |
| `integration_uri` | `StringValueOrRef` | Yes | Integration URI: Lambda function ARN for AWS_PROXY, HTTP URL for HTTP_PROXY |
| `payload_format_version` | `string` | No | Payload format version: `"2.0"` (recommended) or `"1.0"` (defaults to `"2.0"` when empty) |
| `integration_method` | `string` | No | HTTP method for integration request (defaults to route's HTTP method for HTTP_PROXY) |
| `timeout_milliseconds` | `int32` | No | Integration timeout in milliseconds (50-30000, default: 30000) |

### Authorizer Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Unique name for this authorizer (1-128 characters) |
| `authorizer_type` | `string` | Yes | Authorizer type: `"JWT"` or `"REQUEST"` |
| `jwt_configuration` | `AwsHttpApiGatewayJwtConfig` | No | JWT configuration (required when `authorizer_type` is `"JWT"`) |
| `authorizer_uri` | `StringValueOrRef` | No | Lambda function URI for REQUEST authorizers (required when `authorizer_type` is `"REQUEST"`) |
| `authorizer_credentials_arn` | `StringValueOrRef` | No | IAM role ARN that API Gateway assumes to invoke Lambda authorizer |
| `identity_sources` | `string[]` | No | Identity sources used to extract authorization token (e.g., `"$request.header.Authorization"`) |
| `result_ttl_seconds` | `int32` | No | Time in seconds API Gateway caches authorizer result (0-3600, default: 300 for REQUEST) |
| `enable_simple_responses` | `bool` | No | Enable simple boolean responses from Lambda authorizers |
| `authorizer_payload_format_version` | `string` | No | Payload format version for Lambda authorizer event: `"2.0"` (recommended) or `"1.0"` |

### JWT Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `issuer` | `string` | Yes | Token issuer URL (e.g., Cognito: `"https://cognito-idp.{region}.amazonaws.com/{userPoolId}"`) |
| `audiences` | `string[]` | No | Expected audiences (e.g., Cognito app client ID) |

### CORS Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `allow_origins` | `string[]` | No | Origins allowed to make cross-origin requests (e.g., `"https://example.com"`, `"*"`) |
| `allow_methods` | `string[]` | No | HTTP methods allowed for cross-origin requests (e.g., `"GET"`, `"POST"`, `"OPTIONS"`) |
| `allow_headers` | `string[]` | No | Request headers allowed in cross-origin requests (e.g., `"Content-Type"`, `"Authorization"`) |
| `expose_headers` | `string[]` | No | Response headers exposed to the browser in cross-origin responses |
| `max_age_seconds` | `int32` | No | Maximum time in seconds browsers can cache CORS preflight results (0-86400) |
| `allow_credentials` | `bool` | No | Whether the API supports credentials (cookies, authorization headers) in cross-origin requests |

### Stage Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | No | Stage name (defaults to `"$default"` when empty) |
| `auto_deploy` | `bool` | No | Enable automatic deployment when routes/integrations change (defaults to `true` for `"$default"` stage) |
| `access_log` | `AwsHttpApiGatewayAccessLogConfig` | No | Access logging configuration for CloudWatch Logs |
| `default_throttle` | `AwsHttpApiGatewayThrottleConfig` | No | Default throttling settings applied to all routes |
| `stage_variables` | `map<string, string>` | No | Stage variables passed to integrations |

**Validation rules:**

- `routes` must contain at least one route
- Route `authorization_type` must be `"NONE"`, `"JWT"`, or `"AWS_IAM"` when set
- Routes with `authorization_type` `"JWT"` must specify an `authorizer_name`
- Route `authorizer_name` must reference a defined authorizer
- JWT authorizers must have `jwt_configuration` with a non-empty `issuer`
- REQUEST authorizers must have `authorizer_uri` set
- Integration `integration_type` must be `"AWS_PROXY"` or `"HTTP_PROXY"`
- Integration `payload_format_version` must be `"1.0"` or `"2.0"` when set
- Integration `timeout_milliseconds` must be between 50 and 30000 when set

## Examples

### Minimal API

A single catch-all route to a Lambda function:

```yaml
apiVersion: aws.openmcf.org/v1
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

### Multi-Route API with CORS

Multiple routes to different Lambda functions with CORS enabled:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: multi-route-api
spec:
  cors_configuration:
    allow_origins:
      - "https://app.example.com"
    allow_methods:
      - "GET"
      - "POST"
      - "OPTIONS"
    allow_headers:
      - "Content-Type"
      - "Authorization"
    allow_credentials: true
  routes:
    - route_key: "GET /users"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:get-users"
    - route_key: "POST /users"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:create-user"
```

### JWT Authorization with Cognito

Routes protected with JWT authorization using Amazon Cognito:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: jwt-api
spec:
  routes:
    - route_key: "GET /profile"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:get-profile"
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

### Resource References

Reference Lambda functions and other resources using `valueFrom`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: ref-api
spec:
  routes:
    - route_key: "GET /users"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "get-users-function"
            fieldPath: "status.outputs.function_arn"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `api_id` | `string` | The API Gateway API identifier |
| `api_endpoint` | `string` | The default endpoint URL: `https://{api-id}.execute-api.{region}.amazonaws.com` |
| `api_arn` | `string` | The Amazon Resource Name (ARN) of the API |
| `execution_arn` | `string` | The execution ARN prefix: `arn:aws:execute-api:{region}:{account-id}:{api-id}` |
| `stage_invoke_url` | `string` | The invoke URL for the deployed stage |
| `stage_name` | `string` | The name of the deployed stage (e.g., `"$default"`) |

## Related Components

- [AwsLambda](/docs/catalog/aws/awslambda) — Lambda functions used as backend integrations
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — IAM roles for Lambda authorizers
- [AwsCloudwatchLogGroup](/docs/catalog/aws/awscloudwatchloggroup) — CloudWatch Log Groups for access logging
- [AwsWebSocketApiGateway](/docs/catalog/aws/awswebsocketapigateway) — WebSocket APIs (separate component)

## Additional Resources

- [README.md](./README.md) — Comprehensive component documentation
- [examples.md](./examples.md) — Detailed examples with explanations
- [docs/README.md](./docs/README.md) — Architecture deep-dive and best practices
- [AWS API Gateway HTTP API Documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api.html)
- [OpenMCF Documentation](https://openmcf.org/docs)
