# AwsHttpApiGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsHttpApiGatewaySpec defines the desired configuration for an AWS API Gateway HTTP API.

HTTP APIs (API Gateway v2) are optimized for building low-latency, cost-effective
REST APIs and HTTP proxy APIs. They support Lambda proxy integration, HTTP proxy
integration, first-class AWS service integrations (SQS, EventBridge, Step Functions,
Kinesis, AppConfig), private integrations through VPC links, automatic deployments,
native CORS, and JWT/IAM/Lambda authorization.

This component bundles the API, a single stage, routes with inline integrations,
and optional authorizers into one declarative resource. The IaC modules create the
underlying API Gateway resources (api, stage, integrations, routes, authorizers)
and wire them together automatically.

Key design choices:
- HTTP APIs only (WebSocket APIs are a separate protocol surface with their own
  route/response model and would be their own kind).
- Routes carry inline integration config; the module deduplicates shared backends.
- A single stage (defaults to "$default" with auto-deploy) since Planton resources
  are already environment-scoped.
- Authorizers are named and referenced by routes for clean separation.
- Custom domains are the AwsHttpApiDomain kind (a domain outlives any one API
  and maps many APIs); VPC links are the AwsHttpApiVpcLink kind (one link is
  shared by many APIs and owns its own network attachment).
- API keys and usage plans are a REST API feature (the AwsRestApiUsagePlan
  kind); HTTP APIs do not support them -- use JWT/IAM/Lambda authorizers.

Credentials, region, and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiGateway
metadata:
  name: test-http-api
spec:
  region: us-west-2
  description: Test HTTP API Gateway
  routes:
    - route_key: "$default"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:test-function"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.apiVersion` | `string` |  |  |  |
| `spec.corsConfiguration` | `AwsHttpApiGatewayCorsConfig` |  |  |  |
| `spec.corsConfiguration.allowOrigins` | `[]string` |  |  |  |
| `spec.corsConfiguration.allowMethods` | `[]string` |  |  |  |
| `spec.corsConfiguration.allowHeaders` | `[]string` |  |  |  |
| `spec.corsConfiguration.exposeHeaders` | `[]string` |  |  |  |
| `spec.corsConfiguration.maxAgeSeconds` | `int32` |  |  |  |
| `spec.corsConfiguration.allowCredentials` | `bool` |  |  |  |
| `spec.disableExecuteApiEndpoint` | `bool` |  |  |  |
| `spec.ipAddressType` | `string` |  |  |  |
| `spec.stage` | `AwsHttpApiGatewayStageConfig` |  |  |  |
| `spec.stage.name` | `string` |  |  |  |
| `spec.stage.autoDeploy` | `bool` |  |  |  |
| `spec.stage.description` | `string` |  |  |  |
| `spec.stage.accessLog` | `AwsHttpApiGatewayAccessLogConfig` |  |  |  |
| `spec.stage.accessLog.destinationArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.stage.accessLog.format` | `string` | yes |  |  |
| `spec.stage.defaultThrottle` | `AwsHttpApiGatewayThrottleConfig` |  |  |  |
| `spec.stage.defaultThrottle.burstLimit` | `int32` |  |  |  |
| `spec.stage.defaultThrottle.rateLimit` | `double` |  |  |  |
| `spec.stage.detailedMetricsEnabled` | `bool` |  |  |  |
| `spec.stage.routeSettings` | `[]AwsHttpApiGatewayRouteSettings` |  |  |  |
| `spec.stage.routeSettings[].routeKey` | `string` | yes |  |  |
| `spec.stage.routeSettings[].throttlingBurstLimit` | `int32` |  |  |  |
| `spec.stage.routeSettings[].throttlingRateLimit` | `double` |  |  |  |
| `spec.stage.routeSettings[].detailedMetricsEnabled` | `bool` |  |  |  |
| `spec.stage.stageVariables` | `map<string, string>` |  |  |  |
| `spec.routes` | `[]AwsHttpApiGatewayRoute` | yes |  |  |
| `spec.routes[].routeKey` | `string` | yes |  |  |
| `spec.routes[].integration` | `AwsHttpApiGatewayIntegration` | yes |  |  |
| `spec.routes[].integration.integrationType` | `string` | yes |  |  |
| `spec.routes[].integration.integrationUri` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.routes[].integration.integrationSubtype` | `string` |  |  |  |
| `spec.routes[].integration.payloadFormatVersion` | `string` |  |  |  |
| `spec.routes[].integration.integrationMethod` | `string` |  |  |  |
| `spec.routes[].integration.timeoutMilliseconds` | `int32` |  |  |  |
| `spec.routes[].integration.connectionType` | `string` |  |  |  |
| `spec.routes[].integration.connectionId` | `string \| valueFrom` |  |  | AwsHttpApiVpcLink (`status.outputs.vpc_link_id`) |
| `spec.routes[].integration.credentialsArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.routes[].integration.requestParameters` | `map<string, string>` |  |  |  |
| `spec.routes[].integration.responseParameters` | `[]AwsHttpApiGatewayResponseParameters` |  |  |  |
| `spec.routes[].integration.responseParameters[].statusCode` | `string` |  |  |  |
| `spec.routes[].integration.responseParameters[].mappings` | `map<string, string>` | yes |  |  |
| `spec.routes[].integration.tlsServerNameToVerify` | `string` |  |  |  |
| `spec.routes[].integration.description` | `string` |  |  |  |
| `spec.routes[].authorizationType` | `string` |  |  |  |
| `spec.routes[].authorizerName` | `string` |  |  |  |
| `spec.routes[].authorizationScopes` | `[]string` |  |  |  |
| `spec.routes[].operationName` | `string` |  |  |  |
| `spec.authorizers` | `[]AwsHttpApiGatewayAuthorizer` |  |  |  |
| `spec.authorizers[].name` | `string` | yes |  |  |
| `spec.authorizers[].authorizerType` | `string` | yes |  |  |
| `spec.authorizers[].jwtConfiguration` | `AwsHttpApiGatewayJwtConfig` |  |  |  |
| `spec.authorizers[].jwtConfiguration.issuer` | `string \| valueFrom` | yes |  | AwsCognitoUserPool (`status.outputs.issuer`) |
| `spec.authorizers[].jwtConfiguration.audiences` | `[]string \| valueFrom` |  |  | AwsCognitoUserPoolClient (`status.outputs.client_id`) |
| `spec.authorizers[].authorizerUri` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.authorizers[].authorizerCredentialsArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.authorizers[].identitySources` | `[]string` |  |  |  |
| `spec.authorizers[].resultTtlSeconds` | `int32` |  |  |  |
| `spec.authorizers[].enableSimpleResponses` | `bool` |  |  |  |
| `spec.authorizers[].authorizerPayloadFormatVersion` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description of the API (max 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.apiVersion

`string`

A version identifier for the API (max 64 characters). Purely informational
metadata surfaced in the AWS console and exports -- it does not create
stages or affect routing.

- rule: {"string":{"maxLen":"64"}}

### spec.corsConfiguration

`AwsHttpApiGatewayCorsConfig`

CORS configuration for cross-origin requests. When not set, no CORS
headers are returned by the API.

### spec.corsConfiguration.allowOrigins

`[]string`

Origins allowed to make cross-origin requests (e.g., "https://example.com", "*").

### spec.corsConfiguration.allowMethods

`[]string`

HTTP methods allowed for cross-origin requests (e.g., "GET", "POST", "OPTIONS").

### spec.corsConfiguration.allowHeaders

`[]string`

Request headers allowed in cross-origin requests (e.g., "Content-Type", "Authorization").

### spec.corsConfiguration.exposeHeaders

`[]string`

Response headers exposed to the browser in cross-origin responses.

### spec.corsConfiguration.maxAgeSeconds

`int32`

Maximum time in seconds that browsers can cache CORS preflight results.
Reduces the number of preflight OPTIONS requests. Range: 0-86400.

- rule: {"int32":{"lte":86400,"gte":0}}

### spec.corsConfiguration.allowCredentials

`bool`

Whether the API supports credentials (cookies, authorization headers)
in cross-origin requests.

### spec.disableExecuteApiEndpoint

`bool`

Disable the default execute-api endpoint. Set to true when a custom domain
(AwsHttpApiDomain) fronts this API to prevent callers from bypassing the
domain (and its TLS policy / WAF) via the default endpoint.

### spec.ipAddressType

`string`

IP address type for the API's default endpoint.
- "ipv4": Resolve the endpoint to IPv4 addresses only.
- "dualstack": Resolve to both IPv4 and IPv6.
When omitted, AWS defaults new APIs to dualstack. Changing the value
updates the endpoint in place.

### spec.stage

`AwsHttpApiGatewayStageConfig`

Stage configuration for the deployed API. When not set, a "$default" stage
with auto_deploy=true is created automatically, which is the recommended
configuration for most HTTP APIs.

### spec.stage.name

`string`

Stage name. Defaults to "$default" when empty, which is the recommended
configuration for HTTP APIs. Named stages (e.g., "prod", "dev") append
the stage name to the invoke URL path.

### spec.stage.autoDeploy

`bool` · optional (explicit presence)

Enable automatic deployment when routes, integrations, or authorizers
change. When omitted, the modules default to true -- the configuration
that makes a declarative spec self-applying. Set explicitly to false to
require deployments to be created outside this resource (an advanced
pattern; changes to routes then have no effect until a deployment is
published).

### spec.stage.description

`string`

Human-readable description of the stage (max 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.stage.accessLog

`AwsHttpApiGatewayAccessLogConfig`

Access logging configuration. When set, API Gateway streams access logs
to the specified CloudWatch Log Group.

### spec.stage.accessLog.destinationArn

`string | valueFrom` · required

CloudWatch Log Group ARN for access log delivery. Accepts a direct ARN
or a reference to an AwsCloudwatchLogGroup resource.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.stage.accessLog.format

`string` · required

Log format template. Uses API Gateway access log variables (e.g.,
$context.requestId, $context.identity.sourceIp, $context.httpMethod).

Common JSON format:
  {"requestId":"$context.requestId","ip":"$context.identity.sourceIp",
   "method":"$context.httpMethod","path":"$context.routeKey",
   "status":"$context.status","latency":"$context.responseLatency"}

- rule: {"string":{"minLen":"1"}}

### spec.stage.defaultThrottle

`AwsHttpApiGatewayThrottleConfig`

Default throttling settings applied to all routes unless overridden
per route in route_settings.

### spec.stage.defaultThrottle.burstLimit

`int32`

Maximum number of concurrent requests allowed (burst). API Gateway uses
the token bucket algorithm where burst_limit is the bucket size.

- rule: {"int32":{"gte":0}}

### spec.stage.defaultThrottle.rateLimit

`double`

Steady-state request rate limit (requests per second). API Gateway uses
this as the token refill rate.

### spec.stage.detailedMetricsEnabled

`bool`

Emit detailed CloudWatch metrics for all routes (per-route dimensions for
count, latency, and errors). Applies as the stage default; can be
overridden per route in route_settings. Detailed metrics carry additional
CloudWatch cost.

### spec.stage.routeSettings

`[]AwsHttpApiGatewayRouteSettings`

Per-route overrides of throttling and detailed metrics. Each entry targets
one route (by its route_key) and overrides the stage defaults for that
route only -- e.g. a lower rate limit on an expensive search route, or
detailed metrics on just the checkout path.

### spec.stage.routeSettings[].routeKey

`string` · required

The route to override, addressed by its route_key exactly as defined in
routes (e.g. "GET /search", "$default").

- rule: {"string":{"minLen":"1"}}

### spec.stage.routeSettings[].throttlingBurstLimit

`int32`

Maximum number of concurrent requests for this route (token bucket size).
Zero inherits the stage default.

- rule: {"int32":{"gte":0}}

### spec.stage.routeSettings[].throttlingRateLimit

`double`

Steady-state request rate limit for this route (requests per second).
Zero inherits the stage default.

### spec.stage.routeSettings[].detailedMetricsEnabled

`bool`

Emit detailed CloudWatch metrics for this route regardless of the stage
default.

### spec.stage.stageVariables

`map<string, string>`

Stage variables passed to integrations. These act as environment-specific
configuration values accessible in integration request parameters.

### spec.routes

`[]AwsHttpApiGatewayRoute` · required

API routes mapping request patterns to backend integrations. Each route
specifies a route key (e.g., "GET /users", "$default") and an inline
integration that defines the backend target.

When multiple routes share the same integration configuration (same type,
URI, and payload format), the IaC modules automatically deduplicate and
create a single integration resource.

At least one route is required -- an API without routes has no function.

- rule: {"repeated":{"minItems":"1"}}

### spec.routes[].routeKey

`string` · required

Route key defining the request pattern to match. Format: "{METHOD} {PATH}"
for specific routes (e.g., "GET /users", "POST /orders/{id}") or "$default"
for a catch-all route that handles unmatched requests.

- rule: {"string":{"minLen":"1"}}

### spec.routes[].integration

`AwsHttpApiGatewayIntegration` · required

Backend integration that processes requests matching this route.

- rule: {"required":true}

### spec.routes[].integration.integrationType

`string` · required

Integration type. Valid values:
- "AWS_PROXY": Lambda proxy integration or (with integration_subtype)
  a first-class AWS service integration.
- "HTTP_PROXY": HTTP proxy integration -- API Gateway forwards the
  request to an upstream HTTP endpoint.

- rule: {"string":{"minLen":"1"}}

### spec.routes[].integration.integrationUri

`string | valueFrom`

Integration URI for proxy integrations. For AWS_PROXY (Lambda) this is
the Lambda function ARN; for HTTP_PROXY this is the upstream URL (for
private integrations through a VPC link, the ALB/NLB listener ARN or
Cloud Map service ARN). Accepts a direct value or a reference to another
resource's output. Must be omitted for AWS service integrations
(integration_subtype set) -- their target is expressed in
request_parameters instead.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.routes[].integration.integrationSubtype

`string`

AWS service integration subtype. Selects a first-class service action
that API Gateway invokes directly -- no Lambda glue required. Known
subtypes include: "EventBridge-PutEvents", "SQS-SendMessage",
"SQS-ReceiveMessage", "SQS-DeleteMessage", "SQS-PurgeQueue",
"Kinesis-PutRecord", "StepFunctions-StartExecution",
"StepFunctions-StartSyncExecution", "StepFunctions-StopExecution",
"AppConfig-GetConfiguration". The action's parameters (e.g. QueueUrl and
MessageBody for SQS-SendMessage, StateMachineArn and Input for
StepFunctions-StartExecution) are supplied via request_parameters, and
credentials_arn must grant API Gateway permission to call the action.

- rule: {"string":{"maxLen":"128"}}

### spec.routes[].integration.payloadFormatVersion

`string`

Payload format version for Lambda integrations. Controls the format of the
event sent to the Lambda function.
- "2.0" (recommended): Simplified event structure with direct body access.
- "1.0": Legacy format with multi-value headers and base64 encoding.
Defaults to "2.0" when empty. Only applicable to AWS_PROXY integrations.
AWS service integrations (integration_subtype) always use "1.0".

### spec.routes[].integration.integrationMethod

`string`

HTTP method used for the integration request. Valid values: "ANY",
"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" (uppercase).
Defaults to the route's HTTP method for HTTP_PROXY integrations. For
AWS_PROXY (Lambda) integrations, this is always POST regardless of the
value set here.

### spec.routes[].integration.timeoutMilliseconds

`int32`

Integration timeout in milliseconds. If the backend does not respond within
this duration, API Gateway returns a 504 Gateway Timeout.
Range: 50-30000 (50ms to 30s). AWS default: 30000 (30s).
Leave at 0 to use the AWS default.

### spec.routes[].integration.connectionType

`string`

Connection type for the integration.
- "INTERNET" (default): Route to the target over the public internet.
- "VPC_LINK": Route through a VPC link to a private ALB, NLB, or Cloud
  Map service inside a VPC. Requires connection_id and integration_type
  HTTP_PROXY.

### spec.routes[].integration.connectionId

`string | valueFrom`

The VPC link to route through for private integrations. Accepts a direct
VPC link ID or a reference to an AwsHttpApiVpcLink resource. Required
when connection_type is "VPC_LINK"; must be omitted otherwise.

- references: AwsHttpApiVpcLink (`status.outputs.vpc_link_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsHttpApiVpcLink, name: <that resource's name>, fieldPath: status.outputs.vpc_link_id}} -- a bare string does not parse

### spec.routes[].integration.credentialsArn

`string | valueFrom`

IAM role that API Gateway assumes to invoke the integration target.
Required for AWS service integrations (integration_subtype set) -- the
role must trust apigateway.amazonaws.com and grant the service action
(e.g. sqs:SendMessage on the target queue). Optional for Lambda proxy
integrations, which normally authorize through the function's resource
policy instead.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.routes[].integration.requestParameters

`map<string, string>`

Parameter mappings applied to the integration request.

For proxy integrations (Lambda / HTTP), keys are mapping instructions of
the form "append:header.<name>", "overwrite:header.<name>",
"remove:querystring.<name>", "overwrite:path", etc., and values are
static strings or context expressions (e.g. "$context.requestId",
"$request.header.Authorization").

For AWS service integrations (integration_subtype set), keys are the
service action's parameter names (e.g. "QueueUrl", "MessageBody" for
SQS-SendMessage; "StateMachineArn", "Input" for
StepFunctions-StartExecution) and values are static strings or request
expressions like "$request.body".

### spec.routes[].integration.responseParameters

`[]AwsHttpApiGatewayResponseParameters`

Response parameter mappings, keyed by the backend status code they apply
to. Each entry transforms the response API Gateway returns to the caller
for that status code -- e.g. overwriting the status code or injecting a
header. Supported by proxy and service integrations on HTTP APIs.

### spec.routes[].integration.responseParameters[].statusCode

`string`

The backend response status code these mappings apply to (e.g. "403",
"500"). API Gateway accepts 200-599 -- informational (1xx) codes cannot
carry response overrides.

- rule: {"string":{"pattern":"^[2-5][0-9][0-9]$"}}

### spec.routes[].integration.responseParameters[].mappings

`map<string, string>` · required

Mapping instructions applied to the response. Keys are instructions such
as "overwrite:statuscode", "append:header.<name>",
"overwrite:header.<name>", "remove:header.<name>"; values are static
strings or context expressions (e.g. "$context.requestId").

- rule: {"map":{"minPairs":"1"}}

### spec.routes[].integration.tlsServerNameToVerify

`string`

Server name TLS verification for private HTTP_PROXY integrations. When
set, API Gateway verifies the target's certificate against this server
name (SNI) instead of the resolved address -- required when a private
ALB terminates TLS with a certificate issued for the public domain name.
Maps to the integration's tls_config block.

### spec.routes[].integration.description

`string`

Human-readable description of the integration (max 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.routes[].authorizationType

`string`

Authorization type for this route. Valid values:
- "NONE" (default): No authorization required.
- "JWT": JSON Web Token authorization using a JWT authorizer.
- "AWS_IAM": AWS IAM authorization using SigV4 signatures.
- "CUSTOM": Lambda authorization using a REQUEST authorizer.

### spec.routes[].authorizerName

`string`

Name of the authorizer to use for this route. Must match the name of an
authorizer defined in the `authorizers` field. Required when
`authorization_type` is "JWT" (bind a JWT authorizer) or "CUSTOM" (bind
a REQUEST authorizer).

### spec.routes[].authorizationScopes

`[]string`

OAuth 2.0 scopes required for JWT authorization. The request must include
all specified scopes to be authorized. Only applicable when
`authorization_type` is "JWT".

### spec.routes[].operationName

`string`

Operation name for this route (max 64 characters). Surfaced in OpenAPI
exports as the operationId -- useful when generated client SDKs need
stable method names.

- rule: {"string":{"maxLen":"64"}}

### spec.authorizers

`[]AwsHttpApiGatewayAuthorizer`

Named authorizers that can be referenced by routes. Define JWT authorizers
for Cognito/Auth0/OIDC integration, or Lambda (REQUEST) authorizers for
custom authorization logic. Routes reference authorizers by name.

### spec.authorizers[].name

`string` · required

Unique name for this authorizer. Routes reference authorizers by this name.
Must be unique across all authorizers in the spec.

- rule: {"string":{"minLen":"1","maxLen":"128"}}

### spec.authorizers[].authorizerType

`string` · required

Authorizer type. Valid values:
- "JWT": Validates a JSON Web Token using the configured issuer and audiences.
- "REQUEST": Invokes a Lambda function that returns an authorization decision.

- rule: {"string":{"minLen":"1"}}

### spec.authorizers[].jwtConfiguration

`AwsHttpApiGatewayJwtConfig`

JWT configuration. Required when authorizer_type is "JWT".
Configures the token issuer and expected audiences for JWT validation.

### spec.authorizers[].jwtConfiguration.issuer

`string | valueFrom` · required

Token issuer URL. API Gateway validates that the JWT's "iss" claim matches
this value. Accepts a direct issuer URL -- for Cognito,
"https://cognito-idp.{region}.amazonaws.com/{userPoolId}"; for Auth0,
"https://{domain}/" -- or a reference to an AwsCognitoUserPool resource
(its issuer output carries exactly this URL).

- references: AwsCognitoUserPool (`status.outputs.issuer`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.issuer}} -- a bare string does not parse

### spec.authorizers[].jwtConfiguration.audiences

`[]string | valueFrom`

Expected audiences. API Gateway validates that the JWT's "aud" claim matches
one of these values. Each entry accepts a direct value -- for Cognito, an
app client ID -- or a reference to an AwsCognitoUserPoolClient resource;
literals and references can be mixed.

- references: AwsCognitoUserPoolClient (`status.outputs.client_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPoolClient, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.authorizers[].authorizerUri

`string | valueFrom`

Lambda function URI for REQUEST authorizers. This is the Lambda function
invoke ARN. Required when authorizer_type is "REQUEST".

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.authorizers[].authorizerCredentialsArn

`string | valueFrom`

IAM role ARN that API Gateway assumes to invoke the Lambda authorizer.
Only applicable to REQUEST authorizers.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.authorizers[].identitySources

`[]string`

Identity sources used to extract the authorization token or context.
For JWT authorizers: "$request.header.Authorization" (typical).
For REQUEST authorizers: varies by implementation.

### spec.authorizers[].resultTtlSeconds

`int32`

Time in seconds that API Gateway caches the authorizer result.
Range: 0-3600. Set to 0 to disable caching.
AWS default: 300 for REQUEST authorizers with identity sources.

### spec.authorizers[].enableSimpleResponses

`bool`

Enable simple boolean responses from Lambda authorizers.
When true, the Lambda returns {"isAuthorized": true/false} instead of an
IAM policy document. Simpler to implement for most use cases.
Only applicable to REQUEST authorizers.

### spec.authorizers[].authorizerPayloadFormatVersion

`string`

Payload format version for the Lambda authorizer event.
- "2.0" (recommended): Simplified event with direct access to headers/query params.
- "1.0": Legacy format.
Only applicable to REQUEST authorizers.

## Validation Rules

- `ip_address_type_valid`: ip_address_type must be 'ipv4' or 'dualstack' when set
- `route_keys_unique`: route_key values must be unique across routes -- two routes with the same key would conflict in API Gateway
- `route_authorization_type_valid`: route authorization_type must be 'NONE', 'JWT', 'AWS_IAM', or 'CUSTOM' when set
- `authorized_route_requires_authorizer_name`: routes with authorization_type 'JWT' or 'CUSTOM' must specify an authorizer_name
- `route_authorizer_name_must_exist`: route authorizer_name must match a defined authorizer name
- `route_authorizer_type_matches`: a route's authorization_type must match its authorizer: 'JWT' routes reference JWT authorizers, 'CUSTOM' routes reference REQUEST (Lambda) authorizers
- `authorizer_names_unique`: authorizer names must be unique -- routes reference authorizers by name
- `authorizer_type_valid`: authorizer authorizer_type must be 'JWT' or 'REQUEST'
- `jwt_authorizer_requires_config`: JWT authorizers must have jwt_configuration with an issuer and at least one audience
- `request_authorizer_requires_uri`: REQUEST authorizers must have authorizer_uri set
- `integration_type_valid`: route integration integration_type must be 'AWS_PROXY' or 'HTTP_PROXY'
- `integration_uri_per_mode`: integration_uri is required for Lambda and HTTP proxy integrations, and must be omitted when integration_subtype selects an AWS service action (the action's parameters go in request_parameters instead)
- `integration_subtype_requires_aws_proxy`: integration_subtype (AWS service integrations like SQS-SendMessage or StepFunctions-StartExecution) requires integration_type 'AWS_PROXY'
- `integration_subtype_requires_credentials`: AWS service integrations (integration_subtype set) require credentials_arn -- an IAM role API Gateway assumes to call the service
- `integration_connection_type_valid`: integration connection_type must be 'INTERNET' or 'VPC_LINK' when set
- `integration_vpc_link_coupling`: integrations with connection_type 'VPC_LINK' must set connection_id (the VPC link to route through), and connection_id must not be set otherwise
- `integration_vpc_link_requires_http_proxy`: integrations with connection_type 'VPC_LINK' must use integration_type 'HTTP_PROXY' -- private integrations proxy HTTP traffic to an ALB, NLB, or Cloud Map service inside the VPC
- `payload_format_version_valid`: route integration payload_format_version must be '1.0' or '2.0' when set
- `payload_format_version_lambda_only`: payload_format_version '2.0' applies only to Lambda proxy integrations -- HTTP_PROXY and AWS service integrations (integration_subtype) are fixed at '1.0' by AWS
- `integration_timeout_range`: route integration timeout_milliseconds must be between 50 and 30000 when set
- `integration_method_valid`: route integration integration_method must be one of 'ANY', 'DELETE', 'GET', 'HEAD', 'OPTIONS', 'PATCH', 'POST', 'PUT' when set
- `authorizer_ttl_range`: authorizer result_ttl_seconds must be between 0 and 3600 when set
- `authorizer_payload_format_version_valid`: authorizer authorizer_payload_format_version must be '1.0' or '2.0' when set
- `stage_route_settings_target_defined_routes`: stage.route_settings route_key values must match a route defined in routes

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsHttpApiGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.api_id` | `string` | The API Gateway API identifier. Used for constructing resource ARNs and referencing the API in other AWS services. |
| `status.outputs.api_endpoint` | `string` | The default endpoint URL of the API. Format: https://{api-id}.execute-api.{region}.amazonaws.com Clients use this URL to invoke the API when no custom domain is configured. |
| `status.outputs.api_arn` | `string` | The Amazon Resource Name (ARN) of the API. Used for IAM policies and resource-based permissions. |
| `status.outputs.execution_arn` | `string` | The execution ARN prefix for the API. Used in Lambda resource-based policies to grant API Gateway permission to invoke Lambda functions. Format: arn:aws:execute-api:{region}:{account-id}:{api-id} Append "/*/*" for all stages and routes, or "/{stage}/{method}/{path}" for specific permissions. |
| `status.outputs.stage_invoke_url` | `string` | The invoke URL for the deployed stage. For the "$default" stage this is the same as api_endpoint. For named stages the URL includes the stage name: https://{api-id}.execute-api.{region}.amazonaws.com/{stage-name} |
| `status.outputs.stage_name` | `string` | The name of the deployed stage (e.g., "$default", "prod"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.stage.accessLog.destinationArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.routes[].integration.integrationUri` | AwsLambda | `status.outputs.function_arn` |
| `spec.routes[].integration.connectionId` | AwsHttpApiVpcLink | `status.outputs.vpc_link_id` |
| `spec.routes[].integration.credentialsArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.authorizers[].jwtConfiguration.issuer` | AwsCognitoUserPool | `status.outputs.issuer` |
| `spec.authorizers[].jwtConfiguration.audiences` | AwsCognitoUserPoolClient | `status.outputs.client_id` |
| `spec.authorizers[].authorizerUri` | AwsLambda | `status.outputs.function_arn` |
| `spec.authorizers[].authorizerCredentialsArn` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsHttpApiDomain | `spec.apiMappings[].apiId` | `status.outputs.api_id` |

## See Also

- [Overview](../README.md)
