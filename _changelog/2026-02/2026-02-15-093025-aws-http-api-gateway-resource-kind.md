# AWS HTTP API Gateway Resource Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Documentation

## Summary

Added AwsHttpApiGateway as a new cloud resource kind (R05, enum 240) in the AWS provider expansion project. This component creates AWS API Gateway HTTP APIs (v2) with bundled stage, routes with inline integrations, and optional JWT/Lambda authorizers -- the most complex AWS resource in the expansion queue to date.

## Problem Statement / Motivation

AWS API Gateway HTTP APIs are the front door for modern serverless and microservice architectures. The OpenMCF AWS provider lacked coverage for API Gateway, meaning users could deploy Lambda functions (AwsLambda) and event infrastructure (SQS, SNS, EventBridge) but had no declarative way to expose them as HTTP endpoints.

### Pain Points

- No API Gateway component forced users to manage HTTP APIs outside OpenMCF
- Lambda functions deployed via OpenMCF couldn't be wired to HTTP endpoints in infra charts
- The serverless-api infra chart pattern was blocked without this component

## Solution / What's New

A clean, declarative AwsHttpApiGateway component that bundles the API, stage, routes, integrations, and authorizers into one resource with automatic integration deduplication.

### Key Design Decisions

1. **HTTP-only scope** -- WebSocket APIs are a fundamentally different paradigm (different routing, authorization, and integration models). A separate AwsWebSocketApiGateway component will be created later.

2. **Routes with inline integrations** -- Each route carries its own integration config. The IaC module automatically deduplicates: when multiple routes share the same backend (same type + URI + payload version), only one Integration resource is created.

3. **Single stage model** -- Defaults to `$default` with `auto_deploy=true`. OpenMCF resources are already environment-scoped, so multi-stage within one API Gateway is unnecessary.

4. **Named authorizers** -- Authorizers are defined once and referenced by name from routes, supporting both JWT (Cognito/Auth0/OIDC) and Lambda (REQUEST) authorizer types.

5. **No custom domains or VPC links** -- Deferred to v2 following the 80/20 rule.

6. **Payload format 2.0 default** -- AWS recommends 2.0 for HTTP APIs; the TF provider defaults to 1.0 for backward compatibility. We default to 2.0 for better UX.

## Implementation Details

### Proto API (10 messages, 42 fields, 12 CEL validations)

```
AwsHttpApiGatewaySpec
├── AwsHttpApiGatewayCorsConfig (6 fields)
├── AwsHttpApiGatewayStageConfig (5 fields)
│   ├── AwsHttpApiGatewayAccessLogConfig (2 fields)
│   └── AwsHttpApiGatewayThrottleConfig (2 fields)
├── AwsHttpApiGatewayRoute (5 fields)
│   └── AwsHttpApiGatewayIntegration (5 fields)
└── AwsHttpApiGatewayAuthorizer (9 fields)
    └── AwsHttpApiGatewayJwtConfig (2 fields)
```

CEL validations enforce:
- Valid integration types (AWS_PROXY, HTTP_PROXY only)
- Valid authorization types (NONE, JWT, AWS_IAM)
- JWT routes must reference a defined authorizer
- Authorizer names must match defined authorizers (cross-reference validation)
- JWT authorizers require jwt_configuration with issuer
- REQUEST authorizers require authorizer_uri
- Range validations for timeouts, TTLs, CORS max age

### Pulumi Module (7 files, clean separation)

```
module/
├── main.go         -- Orchestrator (provider, api, stage, integrations, authorizers, routes)
├── locals.go       -- Name derivation, tags, integration dedup key function
├── api.go          -- HTTP API with CORS
├── stage.go        -- Stage with access logging and throttling
├── integration.go  -- Deduplicated integrations via key-based map
├── route.go        -- Routes wired to integrations and authorizers
├── authorizer.go   -- JWT and Lambda authorizers
└── outputs.go      -- Output constants
```

Integration deduplication logic:
- Key = `{integrationType}:{integrationUri}:{payloadFormatVersion}`
- First occurrence creates the Integration resource
- Subsequent routes with the same key reuse it
- Routes reference integrations via `integrations/{id}` target format

### Terraform Module (5 files, feature parity)

Uses `for_each` on deduplicated integration maps, authorizer maps, and route maps. Locals compute deduplication keys matching the Pulumi logic.

### Validation Tests

41 spec tests covering:
- 16 happy path scenarios (minimal, multi-route, CORS, JWT, REQUEST, production-ready)
- 25 failure scenarios (missing routes, invalid types, broken cross-references, range violations)

### Stack Outputs

- `api_id` -- API Gateway ID
- `api_endpoint` -- Default execute endpoint URL
- `api_arn` -- API ARN
- `execution_arn` -- For Lambda permissions
- `stage_invoke_url` -- Stage-specific invoke URL
- `stage_name` -- Deployed stage name

## Benefits

- **Serverless API pattern unlocked** -- Users can now build complete serverless APIs (API Gateway + Lambda + DynamoDB) entirely in OpenMCF
- **Infra chart composability** -- `execution_arn` and `api_endpoint` outputs enable downstream wiring in infra charts
- **Clean UX** -- Routes with inline integrations hide the API Gateway resource complexity (12 TF resources) behind a simple declarative spec
- **Integration deduplication** -- Multiple routes to the same Lambda don't create redundant Integration resources

## Impact

- **Users**: Can now deploy production-grade HTTP APIs with JWT authorization, CORS, access logging, and throttling
- **Infra charts**: Enables the planned serverless-api and serverless-event-api infra chart patterns
- **AWS coverage**: 5th new resource kind in the expansion (R05 of 32), completing the serverless tier (SQS, SNS, EventBridge Bus, EventBridge Rule, HTTP API Gateway)

## Related Work

- R01 AwsSqsQueue, R02 AwsSnsTopic, R03 AwsEventBridgeBus, R04 AwsEventBridgeRule (same expansion project)
- Future: AwsWebSocketApiGateway (separate component for WebSocket APIs)
- Future: AwsCognitoUserPool (R12, JWT issuer for API Gateway)
- Future: serverless-api infra chart (T03, depends on this component)

---

**Status**: Production Ready
**Timeline**: Single session
