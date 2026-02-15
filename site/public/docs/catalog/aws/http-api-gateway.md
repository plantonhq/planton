---
title: "HTTP API Gateway"
description: "HTTP API Gateway deployment documentation"
icon: "package"
order: 100
componentName: "awshttpapigateway"
---

# AWS HTTP API Gateway

Deploys an AWS API Gateway HTTP API (v2) with a bundled stage, routes with inline integrations, and optional JWT or Lambda authorizers. HTTP APIs offer lower latency and cost compared to REST APIs, with native support for Lambda proxy integration, HTTP proxy integration, CORS, and automatic deployments.

## What Gets Created

When you deploy an AwsHttpApiGateway resource, OpenMCF provisions:

- **HTTP API** — an API Gateway v2 HTTP API with configured CORS, description, and endpoint settings
- **Stage** — a deployment stage (defaults to `$default` with auto-deploy enabled), with optional access logging and throttling
- **Integrations** — deduplicated backend integrations (Lambda proxy or HTTP proxy) shared across routes with identical configuration
- **Routes** — API routes mapping request patterns (e.g., `GET /users`, `POST /orders/{id}`, `$default`) to their corresponding integrations
- **Authorizers** — optional JWT or Lambda (REQUEST) authorizers referenced by routes for request authorization

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **AWS region** specified in provider config or environment
- **Lambda functions** already deployed if using `AWS_PROXY` integrations
- **HTTP endpoints** reachable if using `HTTP_PROXY` integrations
- **Cognito User Pool or OIDC provider** if using JWT authorizers
- **Lambda authorizer function** if using REQUEST authorizers

## Quick Start

Create a file `api.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: my-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsHttpApiGateway.my-api
spec:
  routes:
    - routeKey: "$default"
      integration:
        integrationType: "AWS_PROXY"
        integrationUri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:my-function"
```

Deploy:

```shell
openmcf apply -f api.yaml
```

This creates an HTTP API with a single catch-all route (`$default`) that forwards all requests to the specified Lambda function. A `$default` stage with auto-deploy is created automatically.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `routes` | `AwsHttpApiGatewayRoute[]` | API routes mapping request patterns to backend integrations | Minimum 1 item |
| `routes[].routeKey` | `string` | Route key defining the request pattern (e.g., `"GET /users"`, `"POST /orders/{id}"`, `"$default"`) | Non-empty |
| `routes[].integration` | `object` | Backend integration that processes requests matching this route | Required |
| `routes[].integration.integrationType` | `string` | Integration type: `"AWS_PROXY"` (Lambda) or `"HTTP_PROXY"` (HTTP endpoint) | Non-empty; must be `"AWS_PROXY"` or `"HTTP_PROXY"` |
| `routes[].integration.integrationUri` | `StringValueOrRef` | Lambda function ARN for `AWS_PROXY`, upstream HTTP URL for `HTTP_PROXY`. Can reference `AwsLambda` via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the API (max 1024 characters) |
| `disableExecuteApiEndpoint` | `bool` | `false` | Disable the default execute-api endpoint when using a custom domain |
| `corsConfiguration.allowOrigins` | `string[]` | `[]` | Origins allowed to make cross-origin requests (e.g., `"https://example.com"`, `"*"`) |
| `corsConfiguration.allowMethods` | `string[]` | `[]` | HTTP methods allowed for cross-origin requests (e.g., `"GET"`, `"POST"`, `"OPTIONS"`) |
| `corsConfiguration.allowHeaders` | `string[]` | `[]` | Request headers allowed in cross-origin requests (e.g., `"Content-Type"`, `"Authorization"`) |
| `corsConfiguration.exposeHeaders` | `string[]` | `[]` | Response headers exposed to the browser in cross-origin responses |
| `corsConfiguration.maxAgeSeconds` | `int` | `0` | Maximum time in seconds that browsers cache CORS preflight results (0–86400) |
| `corsConfiguration.allowCredentials` | `bool` | `false` | Whether the API supports credentials (cookies, authorization headers) in cross-origin requests |
| `stage.name` | `string` | `"$default"` | Stage name. Named stages (e.g., `"prod"`) append the name to the invoke URL path. |
| `stage.autoDeploy` | `bool` | `true` | Automatic deployment when routes, integrations, or authorizers change |
| `stage.accessLog.destinationArn` | `StringValueOrRef` | — | CloudWatch Log Group ARN for access log delivery. Can reference `AwsCloudwatchLogGroup` via `valueFrom`. |
| `stage.accessLog.format` | `string` | — | Log format template using API Gateway access log variables (e.g., `$context.requestId`) |
| `stage.defaultThrottle.burstLimit` | `int` | `0` | Maximum concurrent requests (token bucket size) |
| `stage.defaultThrottle.rateLimit` | `double` | `0` | Steady-state request rate limit in requests per second |
| `stage.stageVariables` | `map<string, string>` | `{}` | Stage variables passed to integrations as environment-specific configuration |
| `routes[].authorizationType` | `string` | `"NONE"` | Authorization type: `"NONE"`, `"JWT"`, or `"AWS_IAM"` |
| `routes[].authorizerName` | `string` | — | Name of the authorizer to use. Required when `authorizationType` is `"JWT"`. Must match a defined authorizer. |
| `routes[].authorizationScopes` | `string[]` | `[]` | OAuth 2.0 scopes required for JWT authorization |
| `routes[].integration.payloadFormatVersion` | `string` | `"2.0"` | Payload format version for Lambda integrations: `"2.0"` (recommended) or `"1.0"`. Only applies to `AWS_PROXY`. |
| `routes[].integration.integrationMethod` | `string` | — | HTTP method for the integration request. Defaults to the route method for `HTTP_PROXY`. Always `POST` for `AWS_PROXY`. |
| `routes[].integration.timeoutMilliseconds` | `int` | `30000` | Integration timeout in milliseconds (50–30000). Returns 504 if the backend does not respond in time. |
| `authorizers[].name` | `string` | — | Unique authorizer name (1–128 characters). Routes reference authorizers by this name. |
| `authorizers[].authorizerType` | `string` | — | Authorizer type: `"JWT"` or `"REQUEST"` |
| `authorizers[].jwtConfiguration.issuer` | `string` | — | Token issuer URL. Required for JWT authorizers. (e.g., `"https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123"`) |
| `authorizers[].jwtConfiguration.audiences` | `string[]` | `[]` | Expected JWT audiences (e.g., Cognito app client ID) |
| `authorizers[].authorizerUri` | `StringValueOrRef` | — | Lambda function invoke ARN for REQUEST authorizers. Can reference `AwsLambda` via `valueFrom`. |
| `authorizers[].authorizerCredentialsArn` | `StringValueOrRef` | — | IAM role ARN that API Gateway assumes to invoke the Lambda authorizer. Can reference `AwsIamRole` via `valueFrom`. |
| `authorizers[].identitySources` | `string[]` | `[]` | Identity sources for token extraction (e.g., `"$request.header.Authorization"`) |
| `authorizers[].resultTtlSeconds` | `int` | `0` | Time in seconds that API Gateway caches the authorizer result (0–3600) |
| `authorizers[].enableSimpleResponses` | `bool` | `false` | Enable simple `{"isAuthorized": true/false}` responses from Lambda authorizers. Only for REQUEST authorizers. |
| `authorizers[].authorizerPayloadFormatVersion` | `string` | — | Payload format version for the Lambda authorizer event: `"2.0"` (recommended) or `"1.0"`. Only for REQUEST authorizers. |

## Examples

### Multi-Route API with CORS

Multiple routes to different Lambda functions with cross-origin support:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: users-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsHttpApiGateway.users-api
spec:
  description: "Users service API"
  corsConfiguration:
    allowOrigins:
      - "https://app.example.com"
    allowMethods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allowHeaders:
      - "Content-Type"
      - "Authorization"
    maxAgeSeconds: 3600
    allowCredentials: true
  routes:
    - routeKey: "GET /users"
      integration:
        integrationType: "AWS_PROXY"
        integrationUri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:list-users"
    - routeKey: "POST /users"
      integration:
        integrationType: "AWS_PROXY"
        integrationUri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:create-user"
    - routeKey: "GET /users/{id}"
      integration:
        integrationType: "AWS_PROXY"
        integrationUri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:get-user"
```

### JWT-Protected API with Cognito

Routes guarded by a JWT authorizer backed by Amazon Cognito:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: secure-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsHttpApiGateway.secure-api
spec:
  description: "Secured API with Cognito JWT authorization"
  corsConfiguration:
    allowOrigins:
      - "https://app.example.com"
    allowMethods:
      - "GET"
      - "POST"
    allowHeaders:
      - "Content-Type"
      - "Authorization"
  routes:
    - routeKey: "GET /profile"
      authorizationType: "JWT"
      authorizerName: "cognito-auth"
      authorizationScopes:
        - "openid"
        - "profile"
      integration:
        integrationType: "AWS_PROXY"
        integrationUri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:get-profile"
    - routeKey: "POST /orders"
      authorizationType: "JWT"
      authorizerName: "cognito-auth"
      authorizationScopes:
        - "openid"
        - "orders:write"
      integration:
        integrationType: "AWS_PROXY"
        integrationUri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:create-order"
    - routeKey: "GET /health"
      integration:
        integrationType: "AWS_PROXY"
        integrationUri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:healthcheck"
  authorizers:
    - name: "cognito-auth"
      authorizerType: "JWT"
      jwtConfiguration:
        issuer: "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123XYZ"
        audiences:
          - "1a2b3c4d5e6f7g8h9i0j"
      identitySources:
        - "$request.header.Authorization"
```

### HTTP Proxy with Stage Configuration and Throttling

An HTTP proxy API with access logging and default throttle settings:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: proxy-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsHttpApiGateway.proxy-api
spec:
  description: "HTTP proxy to internal microservices"
  stage:
    name: "$default"
    autoDeploy: true
    accessLog:
      destinationArn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/apigateway/proxy-api"
      format: '{"requestId":"$context.requestId","ip":"$context.identity.sourceIp","method":"$context.httpMethod","path":"$context.routeKey","status":"$context.status","latency":"$context.responseLatency"}'
    defaultThrottle:
      burstLimit: 500
      rateLimit: 1000
    stageVariables:
      backendHost: "internal.example.com"
  routes:
    - routeKey: "$default"
      integration:
        integrationType: "HTTP_PROXY"
        integrationUri:
          value: "https://internal.example.com"
        integrationMethod: "ANY"
        timeoutMilliseconds: 10000
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding ARNs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: ref-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsHttpApiGateway.ref-api
spec:
  routes:
    - routeKey: "GET /items"
      authorizationType: "JWT"
      authorizerName: "lambda-auth"
      integration:
        integrationType: "AWS_PROXY"
        integrationUri:
          valueFrom:
            kind: AwsLambda
            name: get-items-fn
            fieldPath: "status.outputs.function_arn"
  authorizers:
    - name: "lambda-auth"
      authorizerType: "REQUEST"
      authorizerUri:
        valueFrom:
          kind: AwsLambda
          name: auth-fn
          fieldPath: "status.outputs.function_arn"
      authorizerCredentialsArn:
        valueFrom:
          kind: AwsIamRole
          name: apigw-invoke-role
          fieldPath: "status.outputs.role_arn"
      identitySources:
        - "$request.header.Authorization"
      enableSimpleResponses: true
      authorizerPayloadFormatVersion: "2.0"
      resultTtlSeconds: 300
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `api_id` | `string` | The API Gateway API identifier, used for constructing resource ARNs |
| `api_endpoint` | `string` | The default endpoint URL: `https://{api-id}.execute-api.{region}.amazonaws.com` |
| `api_arn` | `string` | The Amazon Resource Name (ARN) of the API |
| `execution_arn` | `string` | The execution ARN prefix used in Lambda resource-based policies: `arn:aws:execute-api:{region}:{account-id}:{api-id}` |
| `stage_invoke_url` | `string` | The invoke URL for the deployed stage (includes stage name for named stages) |
| `stage_name` | `string` | The name of the deployed stage (e.g., `"$default"`, `"prod"`) |

## Related Components

- [AwsLambda](/docs/catalog/aws/lambda) — Lambda functions used as backend integrations via AWS_PROXY
- [AwsIamRole](/docs/catalog/aws/iam-role) — IAM roles for Lambda authorizer invocation credentials
- [AwsCloudwatchLogGroup](/docs/catalog/aws/awscloudwatchloggroup) — CloudWatch Log Groups for API access logging
- [AwsWebSocketApiGateway](/docs/catalog/aws/awswebsocketapigateway) — WebSocket APIs for real-time bidirectional communication
