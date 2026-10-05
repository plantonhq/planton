# AWS REST API Gateway

Deploys an Amazon API Gateway REST API (API Gateway v1) — the API, its resource and method tree, a single stage with an explicit deployment, and the API-scoped satellites — as one declarative resource. REST APIs are API Gateway's full-featured surface: mapping templates, JSON Schema request validation, API keys, per-method caching and throttling, WAF integration, and EDGE, REGIONAL, or PRIVATE endpoints. The API definition is exactly one of typed `routes` (the modules derive the resource tree from the paths) or an imported `openapi` document; HTTP APIs, the leaner v2 alternative, are the AWS HTTP API Gateway kind.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **REST API** — from typed `routes` or the `openapi` document, with the chosen endpoint type, binary media types, compression threshold, TLS security policy, and resource policy
- **Resource tree, methods, and integrations** — derived level-by-level from the route paths (up to five segments; intermediate segments need no route of their own), with one method, integration, and set of typed responses per route
- **Deployment and Stage** — one explicit deployment whose trigger hashes the full API definition, so every spec change redeploys, and one stage (named "prod" when `stage` is omitted) with optional cache cluster, access logging, X-Ray tracing, and per-method settings
- **Named satellites** — authorizers (Lambda TOKEN/REQUEST or Cognito), JSON Schema models, request validators, gateway-response customizations, and documentation parts with an optional published version
- **Client Certificate** — created only when `stage.clientCertificate.generate` is true; AWS generates the key material and the PEM is exported for backend trust configuration

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with API Gateway control-plane permissions (`apigateway:POST` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Backend targets already exist: a Lambda function for `AWS_PROXY`/`AWS` integrations (its resource policy must allow `apigateway.amazonaws.com`, or wire `credentialsArn`), an HTTP endpoint for `HTTP`/`HTTP_PROXY`, or an NLB-fronted service behind a REST API VPC Link for private backends
- A Lambda invoke ARN for TOKEN/REQUEST authorizers, or a Cognito user pool for `COGNITO_USER_POOLS` authorizers
- Interface VPC endpoints (only for `endpointConfiguration.type: PRIVATE`)
- The account's region-level CloudWatch role configured once per region (only for `methodSettings.loggingLevel` execution logging — stage access logs need no account role)

## Deploy

### Console

Open the deployment store, find **AWS REST API Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the route definitions, and the stage. Start from the **Mock Health API** preset in the [Presets](#presets) tab to prove the tree without a backend, or the **Lambda Proxy Orders API** preset for the standard Lambda pattern.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRestApiGateway
metadata:
  name: orders-api
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: Orders REST API
  routes:
    - path: /orders
      method: POST
      apiKeyRequired: true
      integration:
        type: AWS_PROXY
        uri:
          valueFrom:
            kind: AwsLambda
            name: orders-handler
            fieldPath: status.outputs.invoke_arn
    - path: /health
      method: GET
      integration:
        type: MOCK
        requestTemplates:
          application/json: '{"statusCode":200}'
      responses:
        - statusCode: "200"
          integrationResponseTemplates:
            application/json: '{"ok":true}'
```

```shell
planton apply -f rest-api.yaml
```

This creates a REGIONAL REST API with a Lambda-proxied `POST /orders` requiring an API key, a backend-free `GET /health` answered by a MOCK integration, and a `prod` stage serving the deployment. An Infra Job tracks the provisioning in real time.

### InfraChart

When the API deploys alongside its Lambda backend in one chart, wire the invoke ARN via ValueFromRef:

```yaml
spec:
  region: us-west-2
  routes:
    - path: /orders
      method: POST
      integration:
        type: AWS_PROXY
        uri:
          valueFrom:
            kind: AwsLambda
            name: orders-handler
            fieldPath: status.outputs.invoke_arn
```

The InfraPipeline resolves the dependency graph, deploys the function first, then creates the API integrating with it.

## Key Configuration

These are the most important decisions when configuring a REST API. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Typed routes or OpenAPI, never both** — AWS would accept both and silently overwrite; the spec enforces exactly one. Prefer typed routes when Planton is the source of truth; import `openapi` when the contract already lives in a document. With an OpenAPI body, expect a reconciliation pass in the apply log: AWS's import wipes settings configured outside the document and the apply re-applies them — it is not drift.

**Every spec change redeploys** — REST APIs deploy by explicit snapshot, not auto-deploy, and the module hashes the full definition into the deployment trigger. Adopting an existing API re-deploys it once on the first reconcile (the trigger is engine-side metadata AWS never stores) — a behavioral no-op when the definitions match.

**Endpoint type is the exposure decision** — REGIONAL is right for almost every new API; EDGE provisions a managed CloudFront distribution; PRIVATE is reachable only through interface VPC endpoints and is an API nobody can call until the resource `policy` admits those endpoints. When a custom domain fronts the API, set `disableExecuteApiEndpoint` so callers cannot bypass the domain's TLS policy and WAF via the default endpoint.

**MOCK integrations are first-class** — API Gateway answers from the mapping templates alone, with no backend. Use them for health checks and contract stubs; they prove the tree before any Lambda exists. Non-proxy integrations (AWS, HTTP, MOCK) return only the statuses declared in `responses` — proxy integrations pass the backend response through and need none.

**Keep TOKEN authorizers on a short TTL** — `resultTtlSeconds` caches the authorizer Lambda's decision (default 300); a stolen token stays valid that long. `identityValidationExpression` rejects malformed tokens before the Lambda is even invoked.

**Enabling the stage cache can block an apply for up to 90 minutes** — AWS provisions a dedicated cache cluster when `cacheCluster` turns on or resizes, and the cluster bills hourly by size tier while enabled. Caching still applies only to methods that enable `cachingEnabled` in `methodSettings`. Plan cache changes for a window where a long apply is acceptable.

**Compression cannot be turned off by removing the field** — once `minimumCompressionSize` is set, unsetting it keeps compression on (AWS treats absent as "no change"). Set it to `-1` to explicitly disable compression again.

**Large APIs apply serially, not flakily** — AWS rejects concurrent method-response writes on one API, so the engines serialize them with retries. A big route and response surface makes applies slower; nothing is wrong.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsLambda** | `routes[].integration.uri`, `authorizers[].lambdaInvokeUri` | `status.outputs.invoke_arn` |
| **AwsIamRole** | `routes[].integration.credentialsArn`, `authorizers[].credentialsArn` | `status.outputs.role_arn` |
| **AwsRestApiVpcLink** | `routes[].integration.vpcLinkId` | `status.outputs.vpc_link_id` |
| **AwsCognitoUserPool** | `authorizers[].providerArns` | `status.outputs.user_pool_arn` |
| **AwsVpcEndpoint** | `endpointConfiguration.vpcEndpointIds` | `status.outputs.vpc_endpoint_id` |
| **AwsCloudwatchLogGroup** | `stage.accessLog.destinationArn` | `status.outputs.log_group_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rest_api_id` | The REST API ID (the `{restapi-id}` in invoke URLs) | REST API Domain base-path mappings; Usage Plan `apiStages`; AgentCore Gateway `apiGateway` targets |
| `stage_name` | The stage serving the API | Domain mappings and usage plan stages pair it with `rest_api_id` |
| `invoke_url` | The stage invoke URL | Client configuration and smoke tests against the default endpoint |
| `execution_arn` | The execution ARN prefix | Lambda resource policies and IAM invoke statements scoping who may call |
| `stage_arn` | The stage ARN | WAF web ACL association |
| `client_certificate_pem` | The generated client certificate's PEM body (only with `clientCertificate.generate`) | Backend trust stores verifying calls came through this API |

The remaining outputs — `rest_api_arn`, `root_resource_id`, `deployment_id`, `client_certificate_id`, and the `resource_ids`, `authorizer_ids`, `model_ids`, `request_validator_ids`, `documentation_part_ids`, and route/response map families — exist for import derivations and operational tooling rather than composition wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Backend-free health API** — a MOCK route answering from templates. It costs nothing to run, proves the deployment pipeline end to end, and is the canonical stub while backends are still being built. Start from the **Mock Health API** preset.

**Lambda proxy with API keys** — `AWS_PROXY` routes handing the raw request to a function, `apiKeyRequired` on the mutating methods, and an AWS REST API Usage Plan metering the keys. The workhorse serverless shape. Start from the **Lambda Proxy Orders API** preset.

**Private backend through a VPC link** — `HTTP_PROXY` integrations with `connectionType: VPC_LINK` routing through an AWS REST API VPC Link to an NLB-fronted internal service. The API is the only public surface; the service never leaves the VPC.

**Validating, transforming facade** — non-proxy `AWS`/`HTTP` integrations with JSON Schema `models`, a request validator rejecting malformed bodies before the backend is invoked, and VTL mapping templates reshaping requests and responses. The shape for fronting legacy backends with a clean contract.

## Works With

- [**AWS Lambda**](/infra-catalog/aws-lambda) — proxy and non-proxy backends, and TOKEN/REQUEST authorizer functions
- [**AWS REST API Domain**](/infra-catalog/aws-rest-api-domain) — custom domains mapping onto this API's `rest_api_id` and `stage_name`
- [**AWS REST API Usage Plan**](/infra-catalog/aws-rest-api-usage-plan) — API keys, quotas, and throttling attached to this API's stage
- [**AWS REST API VPC Link**](/infra-catalog/aws-rest-api-vpc-link) — the path to NLB-fronted private backends
- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) — token validation for `COGNITO_USER_POOLS` authorizers
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — stage access log delivery
- [**AWS WAF Web ACL**](/infra-catalog/aws-waf-web-acl) — associates with the `stage_arn` to filter traffic
- [**AWS HTTP API Gateway**](/infra-catalog/aws-http-api-gateway) — the leaner API Gateway v2 alternative for plain proxy workloads
- [**AWS Bedrock AgentCore Gateway**](/infra-catalog/aws-bedrock-agent-core-gateway) — fronts this API's stage as MCP tools for agents
