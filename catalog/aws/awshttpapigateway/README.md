# AWS HTTP API Gateway

The **AwsHttpApiGateway** component provides a declarative way to deploy AWS API Gateway HTTP APIs (v2) with bundled stages, routes, integrations, and optional authorizers. HTTP APIs are optimized for Lambda proxy and HTTP proxy integrations, offering lower latency and cost compared to REST APIs.

## Overview

AWS API Gateway HTTP APIs (API Gateway v2) are designed for building low-latency, cost-effective REST APIs and HTTP proxy APIs. They support:

- **Lambda proxy integration** — Direct invocation of Lambda functions with automatic request/response transformation
- **HTTP proxy integration** — Forward requests to upstream HTTP endpoints, publicly or privately through a VPC link
- **AWS service integrations** — First-class routes into SQS, EventBridge, Step Functions, Kinesis, and AppConfig with no Lambda glue
- **JWT authorization** — Native integration with Cognito, Auth0, or any OIDC provider
- **Lambda (CUSTOM) authorizers** — Custom authorization logic via Lambda functions
- **Automatic deployments** — Changes to routes and integrations are automatically deployed to the stage
- **Native CORS support** — Built-in CORS configuration without custom integration responses

This kind bundles the API, a single stage, routes with inline integrations, and optional authorizers into one declarative resource. The underlying IaC modules create and wire together the necessary API Gateway resources automatically. Custom domains are the separate `AwsHttpApiDomain` component (a domain outlives any one API and maps many APIs); VPC links are the separate `AwsHttpApiVpcLink` component (one link is shared by many APIs).

## When to Use

Use **AwsHttpApiGateway** when you need to:

- Expose Lambda functions as HTTP endpoints
- Create REST APIs with Lambda backend
- Proxy HTTP requests to upstream services
- Implement JWT-based authentication with Cognito or Auth0
- Build cost-effective APIs (HTTP APIs are up to 70% cheaper than REST APIs)
- Deploy APIs with automatic stage deployments

**When not to use:**

- WebSocket APIs (a separate protocol surface with its own route/response model)
- APIs requiring API keys and usage plans (a REST API feature; use the AwsRestApiUsagePlan kind, or JWT/IAM/Lambda authorizers on HTTP APIs)

**Custom domains** are configured with the `AwsHttpApiDomain` component, which maps one or more APIs (by `api_id`) onto an owned domain with an ACM certificate. **Private backends** are reached through an `AwsHttpApiVpcLink` referenced from the integration's `connection_id`.

## Prerequisites

- **AWS credentials** configured via environment variables, IAM instance profile, or Planton provider config
- **AWS region** specified in provider config or environment
- **Lambda functions** (for AWS_PROXY integrations) or **HTTP endpoints** (for HTTP_PROXY integrations) already deployed
- **Cognito User Pool** (for JWT authorization) or **Lambda authorizer function** (for REQUEST authorization) if using authorizers

## Quick Start

Create a minimal HTTP API with a single route to a Lambda function:

```yaml
apiVersion: aws.planton.dev/v1alpha1
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
| `region` | `string` | **Yes** | AWS region for the API |
| `description` | `string` | No | Human-readable description of the API (max 1024 characters) |
| `api_version` | `string` | No | Informational version label surfaced in the console and OpenAPI exports (max 64 characters) |
| `cors_configuration` | `AwsHttpApiGatewayCorsConfig` | No | CORS configuration for cross-origin requests |
| `disable_execute_api_endpoint` | `bool` | No | Disable the default execute-api endpoint (set to true when an AwsHttpApiDomain fronts this API) |
| `ip_address_type` | `string` | No | `ipv4` or `dualstack` (AWS defaults new APIs to dualstack) |
| `stage` | `AwsHttpApiGatewayStageConfig` | No | Stage configuration (defaults to "$default" with auto_deploy=true) |
| `routes` | `AwsHttpApiGatewayRoute[]` | **Yes** | API routes mapping request patterns to backend integrations (at least one required; route keys must be unique) |
| `authorizers` | `AwsHttpApiGatewayAuthorizer[]` | No | Named authorizers referenced by routes (names must be unique) |

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
| `auto_deploy` | `bool` | No | Enable automatic deployment when routes/integrations change. Defaults to **true** whenever omitted; set explicitly to `false` to manage deployments outside this resource |
| `description` | `string` | No | Human-readable stage description (max 1024 characters) |
| `access_log` | `AwsHttpApiGatewayAccessLogConfig` | No | Access logging configuration for CloudWatch Logs |
| `default_throttle` | `AwsHttpApiGatewayThrottleConfig` | No | Default throttling settings applied to all routes |
| `detailed_metrics_enabled` | `bool` | No | Emit per-route CloudWatch metrics as the stage default (extra CloudWatch cost) |
| `route_settings` | `AwsHttpApiGatewayRouteSettings[]` | No | Per-route overrides of throttling and detailed metrics (each entry's `route_key` must match a defined route) |
| `stage_variables` | `map<string, string>` | No | Stage variables passed to integrations |

### AwsHttpApiGatewayRouteSettings

Overrides stage defaults for a single route.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `route_key` | `string` | **Yes** | The route to override, exactly as defined in `routes` (e.g., "GET /search") |
| `throttling_burst_limit` | `int32` | No | Burst limit for this route (zero inherits the stage default) |
| `throttling_rate_limit` | `double` | No | Rate limit for this route (zero inherits the stage default) |
| `detailed_metrics_enabled` | `bool` | No | Emit detailed metrics for this route regardless of the stage default |

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
| `authorization_type` | `string` | No | Authorization type: "NONE" (default), "JWT", "AWS_IAM", or "CUSTOM" (Lambda authorizer) |
| `authorizer_name` | `string` | No | Name of the authorizer to bind ("JWT" routes reference JWT authorizers; "CUSTOM" routes reference REQUEST authorizers) |
| `authorization_scopes` | `string[]` | No | OAuth 2.0 scopes required for JWT authorization |
| `operation_name` | `string` | No | Stable operationId for OpenAPI exports and generated clients (max 64 characters) |

### AwsHttpApiGatewayIntegration

Defines the backend target for a route. Three backend shapes are supported: Lambda proxy, HTTP proxy (public or private through a VPC link), and direct AWS service integrations.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `integration_type` | `string` | **Yes** | Integration type: "AWS_PROXY" (Lambda or service integration) or "HTTP_PROXY" (HTTP endpoint) |
| `integration_uri` | `StringValueOrRef` | Conditional | Target for proxy integrations (Lambda function ARN / HTTP URL / private ALB-NLB listener ARN). Required unless `integration_subtype` is set, in which case it must be omitted |
| `integration_subtype` | `string` | No | AWS service action (e.g., "SQS-SendMessage", "EventBridge-PutEvents", "StepFunctions-StartExecution"). Requires "AWS_PROXY" type and `credentials_arn`; action parameters go in `request_parameters` |
| `payload_format_version` | `string` | No | "2.0" (recommended, Lambda only) or "1.0". Defaults to "2.0" for Lambda; service integrations are fixed at "1.0" by AWS |
| `integration_method` | `string` | No | HTTP method for integration request (defaults to route's HTTP method for HTTP_PROXY) |
| `timeout_milliseconds` | `int32` | No | Integration timeout in milliseconds (50-30000, default: 30000) |
| `connection_type` | `string` | No | "INTERNET" (default) or "VPC_LINK" for private integrations (requires `connection_id` and HTTP_PROXY) |
| `connection_id` | `StringValueOrRef` | Conditional | The AwsHttpApiVpcLink to route through; required with "VPC_LINK", forbidden otherwise |
| `credentials_arn` | `StringValueOrRef` | Conditional | IAM role API Gateway assumes to call the target; required for service integrations |
| `request_parameters` | `map<string, string>` | No | Parameter mappings (proxy) or service-action parameters (subtype integrations) |
| `response_parameters` | `AwsHttpApiGatewayResponseParameters[]` | No | Response transforms keyed by backend status code |
| `tls_server_name_to_verify` | `string` | No | SNI override for private integrations whose internal ALB serves a public-domain certificate |
| `description` | `string` | No | Human-readable integration description (max 1024 characters) |

### AwsHttpApiGatewayResponseParameters

Transforms the response returned to the caller for one backend status code.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status_code` | `string` | **Yes** | The backend status code the mappings apply to (e.g., "500") |
| `mappings` | `map<string, string>` | **Yes** | Mapping instructions (e.g., "overwrite:statuscode" → "503", "append:header.x-request" → "$context.requestId") |

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
| `audiences` | `string[]` | **Yes** | Expected audiences (e.g., Cognito app client ID). HTTP APIs validate both the `iss` and `aud` claims, so at least one audience is required |

## Outputs

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
apiVersion: aws.planton.dev/v1alpha1
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
apiVersion: aws.planton.dev/v1alpha1
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

## AWS Service Integration Example

Route requests straight into SQS and Step Functions with no Lambda glue:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiGateway
metadata:
  name: ingest-api
spec:
  region: us-east-1
  routes:
    - route_key: "POST /enqueue"
      integration:
        integration_type: "AWS_PROXY"
        integration_subtype: "SQS-SendMessage"
        credentials_arn:
          valueFrom:
            kind: AwsIamRole
            name: apigw-service-role
            fieldPath: status.outputs.role_arn
        request_parameters:
          QueueUrl: "https://sqs.us-east-1.amazonaws.com/123456789012/orders"
          MessageBody: "$request.body"
    - route_key: "POST /workflows"
      integration:
        integration_type: "AWS_PROXY"
        integration_subtype: "StepFunctions-StartExecution"
        credentials_arn:
          valueFrom:
            kind: AwsIamRole
            name: apigw-service-role
            fieldPath: status.outputs.role_arn
        request_parameters:
          StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:orders"
          Input: "$request.body"
```

## Private Integration Example

Proxy to an internal ALB through a VPC link:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiGateway
metadata:
  name: private-backend-api
spec:
  region: us-east-1
  routes:
    - route_key: "$default"
      integration:
        integration_type: "HTTP_PROXY"
        integration_uri:
          value: "arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/internal/50dc6c495c0c9188/f2f7dc8efc522ab2"
        connection_type: "VPC_LINK"
        connection_id:
          valueFrom:
            kind: AwsHttpApiVpcLink
            name: private-services-link
            fieldPath: status.outputs.vpc_link_id
```

## Related Kinds

- [AwsLambda](/docs/catalog/aws/awslambda) — Lambda functions used as backend integrations
- [AwsHttpApiVpcLink](/docs/catalog/aws/awshttpapivpclink) — VPC links for private integrations
- [AwsHttpApiDomain](/docs/catalog/aws/awshttpapidomain) — Custom domains mapping APIs under owned DNS names
- [AwsStepFunction](/docs/catalog/aws/awsstepfunction) — State machines invoked via service integrations
- [AwsSqsQueue](/docs/catalog/aws/awssqsqueue) — Queues fed via SQS-SendMessage service integrations
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — IAM roles for authorizers and service integrations
- [AwsCloudwatchLogGroup](/docs/catalog/aws/awscloudwatchloggroup) — CloudWatch Log Groups for access logging

## Deliberately Omitted

- **API keys / usage plans**: Not supported by HTTP APIs (a REST API feature); use JWT/IAM/Lambda authorizers.
- **WebSocket-only stage knobs** (`logging_level`, `data_trace_enabled`): silently ignored by AWS for HTTP APIs, so modeling them would be dishonest surface.
- **Quick-create fields** (`route_key`/`target`/`credentials_arn` on the API): produce AWS-managed route/stage/integration objects that cannot be managed declaratively; the folded routes cover the same outcome honestly.
- **OpenAPI `body` import**: conflicts with the declarative routes fold (two owners for the same routes); revisit on concrete pull.

## Additional Resources

- [AWS API Gateway HTTP API Documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api.html)
- See `GUIDE.md` for architecture deep-dive

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
