# AwsRestApiGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsRestApiGatewaySpec defines the desired configuration for an AWS
API Gateway REST API (API Gateway v1).

REST APIs are API Gateway's full-featured protocol surface: request/
response transformation with mapping templates, request validation
against JSON Schema models, API keys and usage plans, per-method
caching and throttling, WAF integration, canary-capable stages, EDGE/
REGIONAL/PRIVATE endpoints, and gateway-level response customization.
(HTTP APIs - API Gateway v2 - are the leaner, cheaper alternative and
are the AwsHttpApiGateway component.)

This component bundles the API, its resource/method tree, a single
stage with an explicit deployment, and the API-scoped satellites
(authorizers, models, request validators, gateway responses, the
resource policy, documentation, and an optional generated client
certificate) into one declarative resource.

Key design choices:
- The API definition is EXACTLY ONE of `routes` (typed routes with
  inline integrations; the modules derive the API Gateway resource
  tree from the paths) or `openapi` (an OpenAPI document AWS imports).
- REST APIs deploy by EXPLICIT snapshot, not auto-deploy: the modules
  create one deployment whose trigger hashes the full API definition,
  so every spec change redeploys automatically - the declarative
  behavior a Planton component owes its users.
- A single stage, since Planton resources are already
  environment-scoped. Canary traffic shifting is a deploy-workflow
  surface (it needs two live deployments) and is not modeled.
- Authorizers, models, and validators are named and referenced by
  routes for clean separation.
- Custom domains are the AwsRestApiDomain component (a domain outlives
  any one API and maps many APIs); VPC links are the AwsRestApiVpcLink
  component (one link is shared by many APIs); usage plans and API
  keys are the AwsRestApiUsagePlan component (a plan spans APIs and
  stages).
- The account-level CloudWatch-logging role is a region singleton and
  deliberately not modeled here - stage access logs (`stage.access_log`)
  need no account role, while method-level execution logging
  (`method_settings.logging_level`) requires the account's CloudWatch
  role to be configured once per region.

Credentials, region, and deployment workflow live outside this spec
in stack inputs.

## Example

```yaml
# Canonical AwsRestApiGateway example (hack/dev manifest and refgen
# Example source): a typed-route REST API exercising every coexisting
# arm -- MOCK + HTTP integrations, models, a request validator, a TOKEN
# authorizer, a gateway response, a resource policy, documentation, and
# a stage with method settings. Literal ARNs stand in for composed
# references so the offline `tofu plan` renders every arm. The OpenAPI
# definition source is the XOR alternative (exactly-one with routes)
# and is proven by spec tests, not this manifest.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRestApiGateway
metadata:
  name: orders-api
  id: orders-api
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Orders REST API
  apiKeySource: HEADER
  binaryMediaTypes:
    - application/octet-stream
  minimumCompressionSize: 1024
  endpointConfiguration:
    type: REGIONAL
    ipAddressType: ipv4
  endpointAccessMode: BASIC
  securityPolicy: TLS_1_2
  policy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal: "*"
        Action: execute-api:Invoke
        Resource: "*"
  models:
    - name: Order
      contentType: application/json
      description: An order payload
      schema: '{"type":"object","properties":{"id":{"type":"string"}}}'
  requestValidators:
    - name: body-and-params
      validateRequestBody: true
      validateRequestParameters: true
  authorizers:
    - name: orders-token
      type: TOKEN
      lambdaInvokeUri:
        value: arn:aws:apigateway:us-west-2:lambda:path/2015-03-31/functions/arn:aws:lambda:us-west-2:123456789012:function:orders-auth/invocations
      identitySource: method.request.header.Authorization
      resultTtlSeconds: 300
  gatewayResponses:
    - responseType: UNAUTHORIZED
      statusCode: "401"
      responseTemplates:
        application/json: '{"message":"unauthorized"}'
  documentation:
    parts:
      - properties: '{"description":"Orders API"}'
        location:
          type: API
    publishedVersion:
      version: "1"
      description: First published snapshot
  stage:
    name: prod
    description: Production stage
    xrayTracingEnabled: true
    methodSettings:
      - methodPath: "*/*"
        metricsEnabled: true
        loggingLevel: ERROR
        throttlingBurstLimit: 100
        throttlingRateLimit: 50
  routes:
    - path: /health
      method: GET
      operationName: getHealth
      integration:
        type: MOCK
        requestTemplates:
          application/json: '{"statusCode": 200}'
      responses:
        - statusCode: "200"
          integrationResponseTemplates:
            application/json: '{"ok": true}'
    - path: /orders/{id}
      method: GET
      authorization: CUSTOM
      authorizerName: orders-token
      apiKeyRequired: true
      requestValidatorName: body-and-params
      requestParameters:
        method.request.path.id: true
      requestModels:
        application/json: Order
      integration:
        type: HTTP
        httpMethod: GET
        uri:
          value: https://orders.internal.example.com/orders/{id}
        connectionType: INTERNET
        passthroughBehavior: WHEN_NO_MATCH
        timeoutMilliseconds: 29000
      responses:
        - statusCode: "200"
          responseModels:
            application/json: Order
          integrationResponseTemplates:
            application/json: "$input.json('$')"
        - statusCode: "404"
          selectionPattern: ".*not found.*"
          integrationResponseTemplates:
            application/json: '{"error":"not found"}'
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.apiKeySource` | `string` |  |  |  |
| `spec.binaryMediaTypes` | `[]string` |  |  |  |
| `spec.minimumCompressionSize` | `int32` |  |  |  |
| `spec.disableExecuteApiEndpoint` | `bool` |  |  |  |
| `spec.endpointConfiguration` | `AwsRestApiGatewayEndpointConfiguration` |  |  |  |
| `spec.endpointConfiguration.type` | `string` |  |  |  |
| `spec.endpointConfiguration.ipAddressType` | `string` |  |  |  |
| `spec.endpointConfiguration.vpcEndpointIds` | `[]string \| valueFrom` |  |  | AwsVpcEndpoint (`status.outputs.vpc_endpoint_id`) |
| `spec.endpointAccessMode` | `string` |  |  |  |
| `spec.securityPolicy` | `string` |  |  |  |
| `spec.policy` | `object` |  |  |  |
| `spec.routes` | `[]AwsRestApiGatewayRoute` |  |  |  |
| `spec.routes[].path` | `string` |  |  |  |
| `spec.routes[].method` | `string` |  |  |  |
| `spec.routes[].authorization` | `string` |  |  |  |
| `spec.routes[].authorizerName` | `string` |  |  |  |
| `spec.routes[].authorizationScopes` | `[]string` |  |  |  |
| `spec.routes[].apiKeyRequired` | `bool` |  |  |  |
| `spec.routes[].operationName` | `string` |  |  |  |
| `spec.routes[].requestParameters` | `map<string, bool>` |  |  |  |
| `spec.routes[].requestModels` | `map<string, string>` |  |  |  |
| `spec.routes[].requestValidatorName` | `string` |  |  |  |
| `spec.routes[].integration` | `AwsRestApiGatewayIntegration` | yes |  |  |
| `spec.routes[].integration.type` | `string` |  |  |  |
| `spec.routes[].integration.uri` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.invoke_arn`) |
| `spec.routes[].integration.httpMethod` | `string` |  |  |  |
| `spec.routes[].integration.credentialsArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.routes[].integration.connectionType` | `string` |  |  |  |
| `spec.routes[].integration.vpcLinkId` | `string \| valueFrom` |  |  | AwsRestApiVpcLink (`status.outputs.vpc_link_id`) |
| `spec.routes[].integration.passthroughBehavior` | `string` |  |  |  |
| `spec.routes[].integration.contentHandling` | `string` |  |  |  |
| `spec.routes[].integration.cacheKeyParameters` | `[]string` |  |  |  |
| `spec.routes[].integration.cacheNamespace` | `string` |  |  |  |
| `spec.routes[].integration.requestParameters` | `map<string, string>` |  |  |  |
| `spec.routes[].integration.requestTemplates` | `map<string, string>` |  |  |  |
| `spec.routes[].integration.timeoutMilliseconds` | `int32` |  |  |  |
| `spec.routes[].integration.responseTransferMode` | `string` |  |  |  |
| `spec.routes[].integration.tlsInsecureSkipVerification` | `bool` |  |  |  |
| `spec.routes[].responses` | `[]AwsRestApiGatewayRouteResponse` |  |  |  |
| `spec.routes[].responses[].statusCode` | `string` |  |  |  |
| `spec.routes[].responses[].responseModels` | `map<string, string>` |  |  |  |
| `spec.routes[].responses[].responseParameters` | `map<string, bool>` |  |  |  |
| `spec.routes[].responses[].selectionPattern` | `string` |  |  |  |
| `spec.routes[].responses[].integrationResponseParameters` | `map<string, string>` |  |  |  |
| `spec.routes[].responses[].integrationResponseTemplates` | `map<string, string>` |  |  |  |
| `spec.routes[].responses[].contentHandling` | `string` |  |  |  |
| `spec.openapi` | `AwsRestApiGatewayOpenApiDefinition` |  |  |  |
| `spec.openapi.body` | `string` | yes |  |  |
| `spec.openapi.failOnWarnings` | `bool` |  |  |  |
| `spec.openapi.parameters` | `map<string, string>` |  |  |  |
| `spec.openapi.mode` | `string` |  |  |  |
| `spec.models` | `[]AwsRestApiGatewayModel` |  |  |  |
| `spec.models[].name` | `string` | yes |  |  |
| `spec.models[].contentType` | `string` | yes |  |  |
| `spec.models[].description` | `string` |  |  |  |
| `spec.models[].schema` | `string` |  |  |  |
| `spec.requestValidators` | `[]AwsRestApiGatewayRequestValidator` |  |  |  |
| `spec.requestValidators[].name` | `string` | yes |  |  |
| `spec.requestValidators[].validateRequestBody` | `bool` |  |  |  |
| `spec.requestValidators[].validateRequestParameters` | `bool` |  |  |  |
| `spec.authorizers` | `[]AwsRestApiGatewayAuthorizer` |  |  |  |
| `spec.authorizers[].name` | `string` | yes |  |  |
| `spec.authorizers[].type` | `string` |  |  |  |
| `spec.authorizers[].lambdaInvokeUri` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.invoke_arn`) |
| `spec.authorizers[].credentialsArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.authorizers[].providerArns` | `[]string \| valueFrom` |  |  | AwsCognitoUserPool (`status.outputs.user_pool_arn`) |
| `spec.authorizers[].identitySource` | `string` |  |  |  |
| `spec.authorizers[].identityValidationExpression` | `string` |  |  |  |
| `spec.authorizers[].resultTtlSeconds` | `int32` |  |  |  |
| `spec.gatewayResponses` | `[]AwsRestApiGatewayGatewayResponse` |  |  |  |
| `spec.gatewayResponses[].responseType` | `string` |  |  |  |
| `spec.gatewayResponses[].statusCode` | `string` |  |  |  |
| `spec.gatewayResponses[].responseParameters` | `map<string, string>` |  |  |  |
| `spec.gatewayResponses[].responseTemplates` | `map<string, string>` |  |  |  |
| `spec.stage` | `AwsRestApiGatewayStage` |  |  |  |
| `spec.stage.name` | `string` |  |  |  |
| `spec.stage.description` | `string` |  |  |  |
| `spec.stage.stageVariables` | `map<string, string>` |  |  |  |
| `spec.stage.xrayTracingEnabled` | `bool` |  |  |  |
| `spec.stage.cacheCluster` | `AwsRestApiGatewayCacheCluster` |  |  |  |
| `spec.stage.cacheCluster.enabled` | `bool` |  |  |  |
| `spec.stage.cacheCluster.size` | `string` |  |  |  |
| `spec.stage.clientCertificate` | `AwsRestApiGatewayClientCertificate` |  |  |  |
| `spec.stage.clientCertificate.generate` | `bool` |  |  |  |
| `spec.stage.clientCertificate.existingCertificateId` | `string` |  |  |  |
| `spec.stage.clientCertificate.description` | `string` |  |  |  |
| `spec.stage.accessLog` | `AwsRestApiGatewayAccessLog` |  |  |  |
| `spec.stage.accessLog.destinationArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.stage.accessLog.format` | `string` | yes |  |  |
| `spec.stage.documentationVersion` | `string` |  |  |  |
| `spec.stage.methodSettings` | `[]AwsRestApiGatewayMethodSettings` |  |  |  |
| `spec.stage.methodSettings[].methodPath` | `string` | yes |  |  |
| `spec.stage.methodSettings[].metricsEnabled` | `bool` |  |  |  |
| `spec.stage.methodSettings[].loggingLevel` | `string` |  |  |  |
| `spec.stage.methodSettings[].dataTraceEnabled` | `bool` |  |  |  |
| `spec.stage.methodSettings[].throttlingBurstLimit` | `int32` |  |  |  |
| `spec.stage.methodSettings[].throttlingRateLimit` | `double` |  |  |  |
| `spec.stage.methodSettings[].cachingEnabled` | `bool` |  |  |  |
| `spec.stage.methodSettings[].cacheTtlInSeconds` | `int32` |  |  |  |
| `spec.stage.methodSettings[].cacheDataEncrypted` | `bool` |  |  |  |
| `spec.stage.methodSettings[].requireAuthorizationForCacheControl` | `bool` |  |  |  |
| `spec.stage.methodSettings[].unauthorizedCacheControlHeaderStrategy` | `string` |  |  |  |
| `spec.documentation` | `AwsRestApiGatewayDocumentation` |  |  |  |
| `spec.documentation.parts` | `[]AwsRestApiGatewayDocumentationPart` |  |  |  |
| `spec.documentation.parts[].location` | `AwsRestApiGatewayDocumentationLocation` | yes |  |  |
| `spec.documentation.parts[].location.type` | `string` |  |  |  |
| `spec.documentation.parts[].location.path` | `string` |  |  |  |
| `spec.documentation.parts[].location.method` | `string` |  |  |  |
| `spec.documentation.parts[].location.name` | `string` |  |  |  |
| `spec.documentation.parts[].location.statusCode` | `string` |  |  |  |
| `spec.documentation.parts[].properties` | `string` | yes |  |  |
| `spec.documentation.publishedVersion` | `AwsRestApiGatewayDocumentationVersion` |  |  |  |
| `spec.documentation.publishedVersion.version` | `string` | yes |  |  |
| `spec.documentation.publishedVersion.description` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the API will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description of the API.

- rule: {"string":{"maxLen":"1024"}}

### spec.apiKeySource

`string`

Where API Gateway reads API keys from on requests to methods with
api_key_required: the X-Api-Key HEADER (default) or the identity an
AUTHORIZER returns.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["HEADER","AUTHORIZER"]}}

### spec.binaryMediaTypes

`[]string`

Media types treated as binary (e.g. "image/png",
"application/octet-stream", or the "*/*" wildcard). Payloads with
other content types are handled as UTF-8 text.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.minimumCompressionSize

`int32` · optional (explicit presence)

Compress responses at or above this size in bytes (0-10485760).
0 compresses everything. Omitted = compression disabled on create,
UNCHANGED on update (the provider attribute is Computed, so a null
keeps whatever is set). Set -1 to explicitly turn compression back
off after it was enabled - AWS's own clear sentinel.

- rule: {"int32":{"lte":10485760,"gte":-1}}

### spec.disableExecuteApiEndpoint

`bool`

Disable the default execute-api endpoint. Set to true when a custom
domain (AwsRestApiDomain) fronts this API to prevent callers from
bypassing the domain (and its TLS policy / WAF) via the default
endpoint.

### spec.endpointConfiguration

`AwsRestApiGatewayEndpointConfiguration`

Endpoint type and addressing. Omitted = a REGIONAL endpoint (the
right default for almost every new API; EDGE routes through
CloudFront, PRIVATE is reachable only through VPC endpoints).

- rule: PRIVATE endpoints require ip_address_type 'dualstack' (or omit it for the AWS default)
- rule: vpc_endpoint_ids apply only to PRIVATE endpoint types

### spec.endpointConfiguration.type

`string`

The endpoint type. REGIONAL serves from this region (front it with
your own CloudFront/domain as needed); EDGE provisions a managed
CloudFront distribution; PRIVATE is reachable only through interface
VPC endpoints.

- rule: {"string":{"in":["REGIONAL","EDGE","PRIVATE"]}}

### spec.endpointConfiguration.ipAddressType

`string`

Endpoint addressing: "ipv4" or "dualstack". Omitted = the AWS
default. PRIVATE endpoints require dualstack.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ipv4","dualstack"]}}

### spec.endpointConfiguration.vpcEndpointIds

`[]string | valueFrom`

VPC endpoints associated with a PRIVATE API - the interface
endpoints callers invoke it through (they also generate the
Route 53 aliases).

- references: AwsVpcEndpoint (`status.outputs.vpc_endpoint_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpcEndpoint, name: <that resource's name>, fieldPath: status.outputs.vpc_endpoint_id}} -- a bare string does not parse

### spec.endpointAccessMode

`string`

Restrict how the API's endpoints resolve callers: BASIC (default)
or STRICT (rejects requests whose Host header does not match the
invoked endpoint - hardens against host-header confusion).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BASIC","STRICT"]}}

### spec.securityPolicy

`string`

Minimum TLS version and cipher policy on the API's default
endpoint. Omitted = the AWS default for the endpoint type. The
SecurityPolicy_TLS13_* values are the 2025-09 policy family
(FIPS/PFS/PQ variants); TLS_1_0 and TLS_1_2 are the legacy names.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TLS_1_0","TLS_1_2","SecurityPolicy_TLS13_1_3_2025_09","SecurityPolicy_TLS13_1_3_FIPS_2025_09","SecurityPolicy_TLS13_1_2_PFS_PQ_2025_09","SecurityPolicy_TLS13_1_2_FIPS_PQ_2025_09","SecurityPolicy_TLS13_1_2_FIPS_PFS_PQ_2025_09","SecurityPolicy_TLS13_1_2_PQ_2025_09","SecurityPolicy_TLS13_1_2_2021_06","SecurityPolicy_TLS13_2025_EDGE","SecurityPolicy_TLS12_PFS_2025_EDGE","SecurityPolicy_TLS12_2018_EDGE"]}}

### spec.policy

`object`

Resource policy controlling who may invoke the API (IAM policy
document as structured YAML/JSON). REQUIRED in practice for PRIVATE
endpoints (allow the VPC endpoints); also used to IP-allowlist or
cross-account-share public APIs.

### spec.routes

`[]AwsRestApiGatewayRoute`

Typed routes with inline integrations. The modules derive the API
Gateway resource tree from the route paths, then create the method,
integration, and response resources per route.

- rule: responses entries must have unique status_code values

### spec.routes[].path

`string`

The request path, starting with "/" ("/" itself addresses the root
resource). Path parameters use braces ("/users/{id}"); a
greedy-proxy tail is "{proxy+}" ("/files/{proxy+}"). The modules
derive the API Gateway resource tree from these paths - intermediate
segments need no route of their own.

- rule: {"string":{"pattern":"^/"}}

### spec.routes[].method

`string`

The HTTP method ("ANY" matches every method).

- rule: {"string":{"in":["ANY","DELETE","GET","HEAD","OPTIONS","PATCH","POST","PUT"]}}

### spec.routes[].authorization

`string`

How callers authorize: "NONE" (default), "AWS_IAM" (SigV4),
"CUSTOM" (a Lambda authorizer), or "COGNITO_USER_POOLS".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","AWS_IAM","CUSTOM","COGNITO_USER_POOLS"]}}

### spec.routes[].authorizerName

`string`

Name of the authorizer (from `authorizers`) securing this route.
Required when authorization is "CUSTOM" or "COGNITO_USER_POOLS".

### spec.routes[].authorizationScopes

`[]string`

OAuth scopes a Cognito access token must carry (COGNITO_USER_POOLS
routes; with scopes set, only access tokens are accepted - ID
tokens are rejected).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.routes[].apiKeyRequired

`bool`

Require a valid API key (from an AwsRestApiUsagePlan) on this
route.

### spec.routes[].operationName

`string`

Operation name surfaced in exports as the operationId - useful when
generated client SDKs need stable method names.

- rule: {"string":{"maxLen":"64"}}

### spec.routes[].requestParameters

`map<string, bool>`

Method request parameters API Gateway should know about, mapped to
whether each is REQUIRED. Keys use the method-request form:
"method.request.header.X-Tenant", "method.request.querystring.page",
"method.request.path.id". Required parameters are enforced when the
route's validator validates parameters.

### spec.routes[].requestModels

`map<string, string>`

Request body models per content type (e.g. "application/json" ->
"OrderInput"). Values reference `models` entries by name (or the
AWS built-ins "Empty"/"Error"). Enforced when the route's validator
validates bodies.

### spec.routes[].requestValidatorName

`string`

Name of the request validator (from `request_validators`) applied
to this route.

### spec.routes[].integration

`AwsRestApiGatewayIntegration` · required

The backend integration processing matched requests.

- rule: {"required":true}
- rule: MOCK integrations take no uri; every other integration type requires one
- rule: AWS_PROXY (Lambda proxy) integrations use http_method 'POST' (or omit it and the modules send POST)
- rule: http_method is required for HTTP, HTTP_PROXY, and AWS integrations (AWS_PROXY defaults to POST; MOCK takes none)
- rule: integrations with connection_type 'VPC_LINK' must set vpc_link_id (the link to route through), and vpc_link_id must not be set otherwise
- rule: integrations with connection_type 'VPC_LINK' must use integration type 'HTTP' or 'HTTP_PROXY' - REST VPC links front NLB-based HTTP backends
- rule: timeout_milliseconds above 300000 requires response_transfer_mode 'STREAM' (BUFFERED responses cap at 300000)

### spec.routes[].integration.type

`string`

The integration type.

- rule: {"string":{"in":["MOCK","HTTP","HTTP_PROXY","AWS","AWS_PROXY"]}}

### spec.routes[].integration.uri

`string | valueFrom`

The backend target. For AWS_PROXY (and non-proxy Lambda AWS
integrations) this is the function's API Gateway INVOKE ARN - the
AwsLambda invoke_arn output carries exactly this form. For other
AWS integrations it is the service's apigateway path ARN
(arn:aws:apigateway:{region}:{service}:path/...). For HTTP/
HTTP_PROXY it is the endpoint URL (with a VPC link, the private
resource's URL). MOCK integrations take no target.

- references: AwsLambda (`status.outputs.invoke_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.invoke_arn}} -- a bare string does not parse

### spec.routes[].integration.httpMethod

`string`

The HTTP method API Gateway uses toward the backend. Lambda
invocations (AWS_PROXY, and AWS with a Lambda target) are always
POST. Omitted for MOCK.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ANY","DELETE","GET","HEAD","OPTIONS","PATCH","POST","PUT"]}}

### spec.routes[].integration.credentialsArn

`string | valueFrom`

IAM role API Gateway assumes to call the backend. Required for AWS
service integrations; for Lambda, the alternative is a resource
policy on the function allowing apigateway.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.routes[].integration.connectionType

`string`

How the request reaches the backend: "INTERNET" (default) or
"VPC_LINK" (through an AwsRestApiVpcLink to an NLB-fronted private
service).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["INTERNET","VPC_LINK"]}}

### spec.routes[].integration.vpcLinkId

`string | valueFrom`

The VPC link to route through. Required when connection_type is
"VPC_LINK"; must be omitted otherwise.

- references: AwsRestApiVpcLink (`status.outputs.vpc_link_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRestApiVpcLink, name: <that resource's name>, fieldPath: status.outputs.vpc_link_id}} -- a bare string does not parse

### spec.routes[].integration.passthroughBehavior

`string`

What happens to requests whose content type matches no request
template (non-proxy integrations): "WHEN_NO_MATCH" passes the body
through, "WHEN_NO_TEMPLATES" passes through only when NO template
is defined at all, "NEVER" rejects with 415. Omitted = the AWS
default (WHEN_NO_MATCH).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["WHEN_NO_MATCH","WHEN_NO_TEMPLATES","NEVER"]}}

### spec.routes[].integration.contentHandling

`string`

Convert the request body between text and binary before it reaches
the backend. Omitted = passthrough (no conversion).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CONVERT_TO_BINARY","CONVERT_TO_TEXT"]}}

### spec.routes[].integration.cacheKeyParameters

`[]string`

Request parameters cached as cache keys for this method (e.g.
"method.request.querystring.page") - responses are cached per
distinct combination when the stage cache is on.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.routes[].integration.cacheNamespace

`string`

Cache namespace shared across methods that should share cached
responses. Omitted = AWS scopes the cache to this method's
resource.

### spec.routes[].integration.requestParameters

`map<string, string>`

Integration request parameter mappings - keys like
"integration.request.header.X-Backend-Auth", values are static
strings ("'value'") or method-request expressions
("method.request.header.Authorization").

### spec.routes[].integration.requestTemplates

`map<string, string>`

VTL mapping templates transforming the request body, per content
type (non-proxy integrations).

### spec.routes[].integration.timeoutMilliseconds

`int32`

Backend timeout in milliseconds. 50-300000 with BUFFERED responses
(the default; AWS default 29000), up to 900000 with STREAM.
Omitted = the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":900000,"gte":50}}

### spec.routes[].integration.responseTransferMode

`string`

How API Gateway relays the backend response: "BUFFERED" (default)
assembles it, "STREAM" relays chunks as they arrive (enables
response streaming and the longer 900s timeout ceiling).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BUFFERED","STREAM"]}}

### spec.routes[].integration.tlsInsecureSkipVerification

`bool`

Skip TLS certificate verification toward the backend (HTTP
integrations to endpoints with untrusted certificates - a
last-resort knob; prefer fixing the backend certificate).

### spec.routes[].responses

`[]AwsRestApiGatewayRouteResponse`

Typed method responses with their integration response mappings.
Required for non-proxy integrations (AWS/HTTP/MOCK - API Gateway
only returns statuses declared here); proxy integrations pass the
backend response through and need none.

### spec.routes[].responses[].statusCode

`string`

The status code API Gateway may return (e.g. "200", "404").

- rule: {"string":{"pattern":"^[1-5][0-9][0-9]$"}}

### spec.routes[].responses[].responseModels

`map<string, string>`

Response body models per content type, referencing `models` names
(or the AWS built-ins "Empty"/"Error") - documents the response
shape in exports.

### spec.routes[].responses[].responseParameters

`map<string, bool>`

Response headers this method may return, mapped to whether each is
required. Keys use the method-response form:
"method.response.header.Access-Control-Allow-Origin".

### spec.routes[].responses[].selectionPattern

`string`

Regex selecting which backend responses map to this status code -
matched against the backend status code (HTTP) or the Lambda error
message (AWS). The empty pattern is the DEFAULT mapping every
unmatched response falls through to; exactly one response should
leave it empty.

### spec.routes[].responses[].integrationResponseParameters

`map<string, string>`

Integration response header mappings - keys like
"method.response.header.Access-Control-Allow-Origin", values are
static strings ("'*'") or integration-response expressions.

### spec.routes[].responses[].integrationResponseTemplates

`map<string, string>`

VTL templates transforming the backend response body, per content
type.

### spec.routes[].responses[].contentHandling

`string`

Convert the response body between text and binary. Omitted =
passthrough.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CONVERT_TO_BINARY","CONVERT_TO_TEXT"]}}

### spec.openapi

`AwsRestApiGatewayOpenApiDefinition`

An OpenAPI 3.0 (or Swagger 2.0) document AWS imports as the API
definition - routes, integrations (x-amazon-apigateway-* extensions),
models, and authorizers all come from the document.

### spec.openapi.body

`string` · required

The OpenAPI 3.0 (or Swagger 2.0) document, as JSON or YAML text.
Integrations are declared with the x-amazon-apigateway-* extensions.

- rule: {"string":{"minLen":"1"}}

### spec.openapi.failOnWarnings

`bool`

Fail the deployment when AWS reports import warnings (unknown
extensions, ignored fields) instead of silently continuing.

### spec.openapi.parameters

`map<string, string>`

OpenAPI import parameters (e.g. {"endpointConfigurationTypes":
"REGIONAL"} or {"basepath": "prepend"}).

### spec.openapi.mode

`string`

How re-imports apply the document: "overwrite" (default - the
document is the whole truth and replaces the API definition) or
"merge" (the document is layered onto the existing definition).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["overwrite","merge"]}}

### spec.models

`[]AwsRestApiGatewayModel`

JSON Schema models validating and documenting payloads. Routes
reference models by name in request_models/response_models. The
AWS built-ins "Empty" and "Error" exist on every API without being
defined here.

### spec.models[].name

`string` · required

Model name (alphanumeric). Routes reference it in request_models/
response_models; mapping templates address it via $ref.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z0-9]+$"}}

### spec.models[].contentType

`string` · required

The content type the model describes (e.g. "application/json").
Immutable in AWS - changing it replaces the model.

- rule: {"string":{"minLen":"1"}}

### spec.models[].description

`string`

What the model represents.

- rule: {"string":{"maxLen":"1024"}}

### spec.models[].schema

`string`

The JSON Schema document (draft 4), as JSON text.

### spec.requestValidators

`[]AwsRestApiGatewayRequestValidator`

Request validators routes opt into by name - reject malformed
bodies and/or missing required parameters before the backend is
invoked.

### spec.requestValidators[].name

`string` · required

Validator name routes reference in request_validator_name.

- rule: {"string":{"minLen":"1"}}

### spec.requestValidators[].validateRequestBody

`bool`

Validate request bodies against the method's request_models.

### spec.requestValidators[].validateRequestParameters

`bool`

Validate that required parameters (request_parameters entries set
to true) are present.

### spec.authorizers

`[]AwsRestApiGatewayAuthorizer`

Named authorizers routes reference: Lambda TOKEN/REQUEST authorizers
or Cognito user pools.

- rule: TOKEN and REQUEST authorizers require lambda_invoke_uri (and take no provider_arns); COGNITO_USER_POOLS authorizers require provider_arns (and take no lambda_invoke_uri)
- rule: identity_validation_expression applies only to TOKEN authorizers

### spec.authorizers[].name

`string` · required

Unique name for this authorizer. Routes reference authorizers by
this name, and it is the key in the `authorizer_ids` output map.

- rule: {"string":{"minLen":"1","maxLen":"128"}}

### spec.authorizers[].type

`string`

The authorizer kind: "TOKEN" (Lambda, single bearer token),
"REQUEST" (Lambda, full request parameters), or
"COGNITO_USER_POOLS" (Cognito ID/access tokens). Switching to or
from COGNITO_USER_POOLS replaces the authorizer in AWS.

- rule: {"string":{"in":["TOKEN","REQUEST","COGNITO_USER_POOLS"]}}

### spec.authorizers[].lambdaInvokeUri

`string | valueFrom`

The Lambda function evaluating authorization, for TOKEN/REQUEST
authorizers. This is the function's API Gateway INVOKE ARN
(arn:aws:apigateway:{region}:lambda:path/2015-03-31/functions/
{function-arn}/invocations) - the AwsLambda invoke_arn output
carries exactly this form.

- references: AwsLambda (`status.outputs.invoke_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.invoke_arn}} -- a bare string does not parse

### spec.authorizers[].credentialsArn

`string | valueFrom`

IAM role API Gateway assumes to invoke the authorizer Lambda.
Omitted = the Lambda's resource policy must allow
apigateway.amazonaws.com instead.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.authorizers[].providerArns

`[]string | valueFrom`

Cognito user pools whose tokens the authorizer accepts, for
COGNITO_USER_POOLS authorizers.

Containment-exempt: the API authorizes callers AGAINST the pool; it is
not deployed into it. On a diagram the pool is a room its clients and
identity providers live in, and the API stands outside with a line in
-- the same verdict the HTTP API's JWT issuer and the ALB listener's
Cognito action already carry.

- references: AwsCognitoUserPool (`status.outputs.user_pool_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_arn}} -- a bare string does not parse

### spec.authorizers[].identitySource

`string`

Where the token/identity lives on the request. Omitted = the AWS
default "method.request.header.Authorization". REQUEST authorizers
may list several sources comma-separated
("method.request.header.Auth, method.request.querystring.Name").

### spec.authorizers[].identityValidationExpression

`string`

Regex the incoming token must match BEFORE the authorizer is
invoked (TOKEN authorizers) - rejects malformed tokens without
paying for a Lambda call.

### spec.authorizers[].resultTtlSeconds

`int32` · optional (explicit presence)

Seconds API Gateway caches authorizer results (0-3600; 0 disables
caching). Omitted = the AWS default of 300.

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.gatewayResponses

`[]AwsRestApiGatewayGatewayResponse`

Customize the responses API Gateway itself generates (4XX/5XX
classes, throttling, missing auth, WAF blocks) - inject headers or
rewrite bodies before they reach callers.

### spec.gatewayResponses[].responseType

`string`

Which gateway-generated response to customize.

- rule: {"string":{"in":["DEFAULT_4XX","DEFAULT_5XX","RESOURCE_NOT_FOUND","UNAUTHORIZED","INVALID_API_KEY","ACCESS_DENIED","AUTHORIZER_FAILURE","AUTHORIZER_CONFIGURATION_ERROR","INVALID_SIGNATURE","EXPIRED_TOKEN","MISSING_AUTHENTICATION_TOKEN","INTEGRATION_FAILURE","INTEGRATION_TIMEOUT","API_CONFIGURATION_ERROR","UNSUPPORTED_MEDIA_TYPE","BAD_REQUEST_PARAMETERS","BAD_REQUEST_BODY","REQUEST_TOO_LARGE","THROTTLED","QUOTA_EXCEEDED","WAF_FILTERED"]}}

### spec.gatewayResponses[].statusCode

`string`

Override the response's status code (e.g. return "404" for
MISSING_AUTHENTICATION_TOKEN to hide the API's existence). Omitted =
the type's default code.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[1-5][0-9][0-9]$"}}

### spec.gatewayResponses[].responseParameters

`map<string, string>`

Header mappings on the customized response - keys like
"gatewayresponse.header.Access-Control-Allow-Origin", values are
static strings ("'*'") or context expressions.

### spec.gatewayResponses[].responseTemplates

`map<string, string>`

Body templates per content type (e.g. "application/json" ->
"{\"message\":$context.error.messageString}").

### spec.stage

`AwsRestApiGatewayStage`

The stage serving the deployed API. Omitted = a stage named "prod"
with defaults.

- rule: method_settings entries must have unique method_path values

### spec.stage.name

`string`

Stage name (letters, digits, hyphens, underscores; it appears in
the invoke URL path). Omitted = "prod".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128","pattern":"^[a-zA-Z0-9_-]+$"}}

### spec.stage.description

`string`

Human-readable description of the stage.

- rule: {"string":{"maxLen":"1024"}}

### spec.stage.stageVariables

`map<string, string>`

Stage variables integrations can read (e.g. per-environment
backend hosts referenced as ${stageVariables.backendHost}).

### spec.stage.xrayTracingEnabled

`bool`

Trace requests with AWS X-Ray.

### spec.stage.cacheCluster

`AwsRestApiGatewayCacheCluster`

The stage response cache. Provisioning a cache cluster bills
hourly by size while enabled; per-method caching behavior is tuned
in method_settings.

### spec.stage.cacheCluster.enabled

`bool`

Enable the cache cluster. Enabling or resizing can take up to 90
minutes while AWS provisions it; caching still applies only to
methods that enable caching_enabled in method_settings.

### spec.stage.cacheCluster.size

`string`

Cache size in GB - one of AWS's fixed tiers. Larger tiers bill
more per hour.

- rule: {"string":{"in":["0.5","1.6","6.1","13.5","28.4","58.2","118","237"]}}

### spec.stage.clientCertificate

`AwsRestApiGatewayClientCertificate`

TLS client certificate API Gateway presents to HTTP backends, so
they can verify calls really come through the API.

- rule: set exactly one of generate or existing_certificate_id
- rule: description applies only to a generated certificate

### spec.stage.clientCertificate.generate

`bool`

Generate a certificate with this API (AWS creates and rotates the
key material; the PEM is exported for backend trust
configuration).

### spec.stage.clientCertificate.existingCertificateId

`string`

Use an existing API Gateway client certificate by ID.

### spec.stage.clientCertificate.description

`string`

Description on the generated certificate.

- rule: {"string":{"maxLen":"1024"}}

### spec.stage.accessLog

`AwsRestApiGatewayAccessLog`

Access logging to CloudWatch.

### spec.stage.accessLog.destinationArn

`string | valueFrom` · required

CloudWatch Log Group ARN for access log delivery.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.stage.accessLog.format

`string` · required

Log format template using $context variables.

Common JSON format:
  {"requestId":"$context.requestId","ip":"$context.identity.sourceIp",
   "method":"$context.httpMethod","path":"$context.resourcePath",
   "status":"$context.status","latency":"$context.responseLatency"}

- rule: {"string":{"minLen":"1"}}

### spec.stage.documentationVersion

`string`

The published documentation version this stage serves. Must match
documentation.published_version.version.

### spec.stage.methodSettings

`[]AwsRestApiGatewayMethodSettings`

Per-method overrides of logging, metrics, throttling, and caching.

### spec.stage.methodSettings[].methodPath

`string` · required

The methods these settings target: "*/*" for every method, or
"{resource-path}/{method}" ("users/GET",
"orders/{id}/items/POST").

- rule: {"string":{"minLen":"1"}}

### spec.stage.methodSettings[].metricsEnabled

`bool` · optional (explicit presence)

Emit per-method CloudWatch metrics (billed per metric).

### spec.stage.methodSettings[].loggingLevel

`string`

Execution logging level for these methods. Requires the account's
region-level CloudWatch role to be configured (a region singleton
outside this component).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["OFF","ERROR","INFO"]}}

### spec.stage.methodSettings[].dataTraceEnabled

`bool` · optional (explicit presence)

Log full request/response bodies (INFO logging) - invaluable for
debugging, unsafe for production PII.

### spec.stage.methodSettings[].throttlingBurstLimit

`int32` · optional (explicit presence)

Token-bucket burst for these methods. Omitted = the account
default.

- rule: {"int32":{"gte":0}}

### spec.stage.methodSettings[].throttlingRateLimit

`double` · optional (explicit presence)

Steady-state request rate (requests/second). Omitted = the account
default.

- rule: {"double":{"gte":0}}

### spec.stage.methodSettings[].cachingEnabled

`bool` · optional (explicit presence)

Cache responses for these methods (requires the stage
cache_cluster).

### spec.stage.methodSettings[].cacheTtlInSeconds

`int32` · optional (explicit presence)

Cached-response TTL in seconds.

- rule: {"int32":{"gte":0}}

### spec.stage.methodSettings[].cacheDataEncrypted

`bool` · optional (explicit presence)

Encrypt cached responses.

### spec.stage.methodSettings[].requireAuthorizationForCacheControl

`bool` · optional (explicit presence)

Only authorized callers may send Cache-Control headers that bypass
or invalidate the cache.

### spec.stage.methodSettings[].unauthorizedCacheControlHeaderStrategy

`string`

What happens to unauthorized cache-control requests:
fail, succeed with a warning header, or succeed silently.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FAIL_WITH_403","SUCCEED_WITH_RESPONSE_HEADER","SUCCEED_WITHOUT_RESPONSE_HEADER"]}}

### spec.documentation

`AwsRestApiGatewayDocumentation`

API documentation: typed documentation parts plus an optional
published version served on the stage.

### spec.documentation.parts

`[]AwsRestApiGatewayDocumentationPart`

Documentation parts - each attaches a properties document to one
API element.

### spec.documentation.parts[].location

`AwsRestApiGatewayDocumentationLocation` · required

Where the documentation attaches.

- rule: {"required":true}

### spec.documentation.parts[].location.type

`string`

The element type being documented.

- rule: {"string":{"in":["API","AUTHORIZER","MODEL","RESOURCE","METHOD","PATH_PARAMETER","QUERY_PARAMETER","REQUEST_HEADER","REQUEST_BODY","RESPONSE","RESPONSE_HEADER","RESPONSE_BODY"]}}

### spec.documentation.parts[].location.path

`string`

The resource path the element lives under (e.g. "/users/{id}").
Omitted = matches every path.

### spec.documentation.parts[].location.method

`string`

The HTTP method the element belongs to. Omitted = matches every
method.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ANY","DELETE","GET","HEAD","OPTIONS","PATCH","POST","PUT"]}}

### spec.documentation.parts[].location.name

`string`

The named element (parameter/header name, authorizer or model
name) for types that address one.

### spec.documentation.parts[].location.statusCode

`string`

The status code for RESPONSE* types. Omitted = matches every
status.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([1-5][0-9][0-9]|\\*)$"}}

### spec.documentation.parts[].properties

`string` · required

The documentation content as a JSON object (OpenAPI documentation
fields, e.g. {"description":"Lists users","summary":"..."}).

- rule: {"string":{"minLen":"1"}}

### spec.documentation.publishedVersion

`AwsRestApiGatewayDocumentationVersion`

Publish the parts as a documentation version (REST APIs serve
documentation from published versions, not live parts). The stage's
documentation_version selects it.

### spec.documentation.publishedVersion.version

`string` · required

The version identifier (e.g. "1.0.0"). Publishing a changed
version replaces the published snapshot.

- rule: {"string":{"minLen":"1"}}

### spec.documentation.publishedVersion.description

`string`

What this documentation version covers.

- rule: {"string":{"maxLen":"1024"}}

## Validation Rules

- `definition_exactly_one_source`: define the API with exactly one of routes (typed) or openapi (imported document)
- `route_path_method_unique`: route path+method pairs must be unique - two routes with the same path and method would conflict in API Gateway
- `route_path_depth_max_five`: route paths support at most five segments - use a greedy-proxy tail ({proxy+}) for deeper hierarchies
- `model_names_unique`: model names must be unique - routes reference models by name
- `request_validator_names_unique`: request validator names must be unique - routes reference validators by name
- `authorizer_names_unique`: authorizer names must be unique - routes reference authorizers by name
- `gateway_response_types_unique`: gateway_responses entries must have unique response_type values
- `authorized_route_requires_authorizer_name`: routes with authorization 'CUSTOM' or 'COGNITO_USER_POOLS' must specify an authorizer_name
- `route_authorizer_name_must_exist`: route authorizer_name must match a defined authorizer name
- `route_authorizer_type_matches`: a route's authorization must match its authorizer: 'CUSTOM' routes reference TOKEN or REQUEST authorizers, 'COGNITO_USER_POOLS' routes reference COGNITO_USER_POOLS authorizers
- `authorization_scopes_cognito_only`: authorization_scopes apply only to routes with authorization 'COGNITO_USER_POOLS'
- `route_validator_name_must_exist`: route request_validator_name must match a defined request validator name
- `route_request_models_must_exist`: route request_models values must reference a defined model name or the AWS built-ins 'Empty'/'Error'
- `route_response_models_must_exist`: route response model values must reference a defined model name or the AWS built-ins 'Empty'/'Error'
- `stage_documentation_version_requires_documentation`: stage.documentation_version requires documentation.published_version with the same version - the stage can only serve a version this spec publishes

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRestApiGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.rest_api_id` | `string` | The REST API ID (the {restapi-id} in invoke URLs and import IDs). |
| `status.outputs.rest_api_arn` | `string` | The REST API ARN. |
| `status.outputs.execution_arn` | `string` | The execution ARN prefix (arn:aws:execute-api:{region}:{account}:{restapi-id}) - the base Lambda resource policies and IAM invoke statements scope to. |
| `status.outputs.root_resource_id` | `string` | The root ("/") resource ID - the parent new resources attach under. |
| `status.outputs.stage_name` | `string` | The stage name serving the API. |
| `status.outputs.stage_arn` | `string` | The stage ARN (the WAF web-ACL association target). |
| `status.outputs.invoke_url` | `string` | The stage invoke URL (https://{restapi-id}.execute-api.{region}.amazonaws.com/{stage}). |
| `status.outputs.deployment_id` | `string` | The deployment ID the stage currently serves. |
| `status.outputs.client_certificate_id` | `string` | The generated client certificate's ID (set only when stage.client_certificate.generate is true). |
| `status.outputs.client_certificate_pem` | `string` | The generated client certificate's PEM body - backends add it to their trust store to verify calls come through the API. |
| `status.outputs.resource_ids` | `map<string, string>` | API Gateway resource IDs keyed by route path (every derived tree node, e.g. "/users" and "/users/{id}"). |
| `status.outputs.authorizer_ids` | `map<string, string>` | Authorizer IDs keyed by each `authorizers` entry's name. |
| `status.outputs.model_ids` | `map<string, string>` | Model IDs keyed by each `models` entry's name. |
| `status.outputs.request_validator_ids` | `map<string, string>` | Request validator IDs keyed by each `request_validators` entry's name. |
| `status.outputs.documentation_part_ids` | `map<string, string>` | Documentation part IDs keyed by each `documentation.parts` entry's position ("0", "1", ...) - the order they are declared in the spec. |
| `status.outputs.route_resource_ids` | `map<string, string>` | API Gateway resource IDs keyed by ROUTE key ("GET /users/{id}") - the same keys the method and integration instances use, so import derivations (and chart wiring) can resolve each route's resource ID blind. Root-path routes map to the root resource ID. |
| `status.outputs.route_methods` | `map<string, string>` | Each route's HTTP method keyed by route key ("GET /users/{id}" -> "GET"). A config echo: the route key is composite, so import derivations need the method as its own addressable value. |
| `status.outputs.response_resource_ids` | `map<string, string>` | API Gateway resource IDs keyed by ROUTE-RESPONSE key ("GET /users/{id}\|200") - the method-response and integration-response instances' keys. |
| `status.outputs.response_methods` | `map<string, string>` | Each route-response's HTTP method keyed by route-response key ("GET /users/{id}\|200" -> "GET"). A config echo for import derivations. |
| `status.outputs.response_status_codes` | `map<string, string>` | Each route-response's status code keyed by route-response key ("GET /users/{id}\|200" -> "200"). A config echo for import derivations. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.endpointConfiguration.vpcEndpointIds` | AwsVpcEndpoint | `status.outputs.vpc_endpoint_id` |
| `spec.routes[].integration.uri` | AwsLambda | `status.outputs.invoke_arn` |
| `spec.routes[].integration.credentialsArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.routes[].integration.vpcLinkId` | AwsRestApiVpcLink | `status.outputs.vpc_link_id` |
| `spec.authorizers[].lambdaInvokeUri` | AwsLambda | `status.outputs.invoke_arn` |
| `spec.authorizers[].credentialsArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.authorizers[].providerArns` | AwsCognitoUserPool | `status.outputs.user_pool_arn` |
| `spec.stage.accessLog.destinationArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsRestApiDomain | `spec.basePathMappings[].restApiId` | `status.outputs.rest_api_id` |
| AwsRestApiDomain | `spec.basePathMappings[].stageName` | `status.outputs.stage_name` |
| AwsRestApiUsagePlan | `spec.apiStages[].restApiId` | `status.outputs.rest_api_id` |
| AwsRestApiUsagePlan | `spec.apiStages[].stageName` | `status.outputs.stage_name` |

## See Also

- [Overview](../README.md)
