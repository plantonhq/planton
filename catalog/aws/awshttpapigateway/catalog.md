# AWS HTTP API Gateway

Deploys an HTTP API on Amazon API Gateway (v2) with route-to-integration wiring, JWT and Lambda authorizers, CORS configuration, and auto-deploying stages. The component bundles the API, stage, routes, integrations, and authorizers into a single declarative resource, with typed references for wiring Lambda backends, VPC links, authorizer roles, and Cognito issuers to resources deployed alongside it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **HTTP API** -- an API Gateway v2 HTTP API with optional CORS settings, description, version label, IP address type (IPv4 or dual-stack), and default-endpoint control
- **Stage** -- a deployment stage (defaults to `$default` with auto-deploy enabled) with optional access logging, throttling, detailed CloudWatch metrics, per-route setting overrides, and stage variables
- **Integrations** -- backend targets for each route: Lambda proxy, HTTP proxy (public, or private through a VPC link), or first-class AWS service actions (SQS, EventBridge, Step Functions, Kinesis, AppConfig) with no Lambda glue; routes sharing the same configuration are automatically deduplicated into a single integration resource
- **Routes** -- request pattern mappings (e.g., `GET /users`, `$default`) wired to their backend integrations with optional authorization and OpenAPI operation names
- **Authorizers** -- created only when authorizers are defined; supports JWT authorizers (Cognito, Auth0, OIDC) and Lambda REQUEST authorizers
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A Lambda function or HTTP endpoint** for each route integration. For Lambda proxy integrations (`AWS_PROXY`), provide the function ARN. For HTTP proxy integrations (`HTTP_PROXY`), provide the upstream URL.
- **A CloudWatch Log Group** (optional) for access logging. Provide the ARN directly or reference an AwsCloudwatchLogGroup Infra Component via ValueFromRef.
- **A Lambda authorizer function** (optional) for REQUEST-type authorizers. Provide the function ARN directly or reference an AwsLambda Infra Component via ValueFromRef.
- **An IAM role** (optional) for Lambda authorizer invocation. Required when using REQUEST authorizers. Provide the ARN directly or reference an AwsIamRole Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS HTTP API Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default Route to Lambda** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiGateway
metadata:
  name: my-api
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  routes:
    - routeKey: "$default"
      integration:
        integrationType: AWS_PROXY
        integrationUri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:my-handler"
```

```shell
planton apply -f http-api-gateway.yaml
```

This creates an HTTP API with a single catch-all route forwarding all requests to a Lambda function, using the `$default` stage with auto-deploy. No CORS or authorization is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire authorizers and access logging to resources deployed in the same InfraPipeline:

```yaml
spec:
  stage:
    accessLog:
      destinationArn:
        valueFrom:
          kind: AwsCloudwatchLogGroup
          name: api-access-logs
          fieldPath: status.outputs.log_group_arn
      format: '{"requestId":"$context.requestId","ip":"$context.identity.sourceIp","method":"$context.httpMethod","status":"$context.status"}'
  authorizers:
    - name: lambda-auth
      authorizerType: REQUEST
      authorizerUri:
        valueFrom:
          kind: AwsLambda
          name: auth-function
          fieldPath: status.outputs.function_arn
      authorizerCredentialsArn:
        valueFrom:
          kind: AwsIamRole
          name: api-authorizer-role
          fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the log group, Lambda function, and IAM role first, then provisions the HTTP API with the resolved values.

## Key Configuration

These are the most important decisions when configuring an HTTP API Gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Route design** -- Each route maps a request pattern (`GET /users`, `POST /orders/{id}`, `$default`) to a backend integration. Use specific routes for API Gateway-level routing, or a single `$default` catch-all to delegate routing to the Lambda function. At least one route is required.

**Integration type** -- `AWS_PROXY` forwards requests to a Lambda function as a proxy event (the `payloadFormatVersion` defaults to `2.0`; `1.0` is legacy and the only format for non-Lambda backends). `HTTP_PROXY` forwards requests to an upstream HTTP endpoint. Setting `integrationSubtype` (e.g. `SQS-SendMessage`, `StepFunctions-StartExecution`) turns an `AWS_PROXY` integration into a direct AWS service action: the action's parameters go in `requestParameters`, `credentialsArn` names the IAM role API Gateway assumes, and no `integrationUri` is set.

**Private integrations** -- Set `connectionType: VPC_LINK` with `connectionId` referencing an AwsHttpApiVpcLink to reach a private ALB, NLB, or Cloud Map service inside a VPC over `HTTP_PROXY`. `tlsServerNameToVerify` overrides certificate verification (SNI) when the private target terminates TLS with a public-domain certificate.

**Authorization** -- Routes support `NONE` (public), `JWT` (Cognito, Auth0, or any OIDC provider), `AWS_IAM` (SigV4), and `CUSTOM` (a Lambda REQUEST authorizer decides per call). JWT authorizers validate tokens at the edge without invoking Lambda; their `issuer` can reference an AwsCognitoUserPool and each audience an AwsCognitoUserPoolClient. Routes bind authorizers by name, and the authorization type must match the authorizer's kind (JWT routes bind JWT authorizers, CUSTOM routes bind REQUEST authorizers).

**CORS** -- Configure `corsConfiguration` when the API is called from web browsers on a different domain. Specify allowed origins, methods, headers, and credential support. Omit entirely for server-to-server APIs.

**Throttling and metrics** -- Set `stage.defaultThrottle` with `burstLimit` and `rateLimit` to protect backend services from traffic spikes, and `stage.detailedMetricsEnabled` for per-route CloudWatch dimensions. `stage.routeSettings` overrides either per route (e.g. a lower rate on an expensive search route) — each entry must target a declared route key.

**Response shaping** -- `responseParameters` on an integration transforms what callers receive per backend status code (overwrite the status, inject or strip headers); `requestParameters` on proxy integrations applies the same mapping instructions to the upstream request.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsLambda** | `routes[].integration.integrationUri` | `status.outputs.function_arn` |
| **AwsHttpApiVpcLink** (optional) | `routes[].integration.connectionId` | `status.outputs.vpc_link_id` |
| **AwsIamRole** (optional) | `routes[].integration.credentialsArn` | `status.outputs.role_arn` |
| **AwsCloudwatchLogGroup** (optional) | `stage.accessLog.destinationArn` | `status.outputs.log_group_arn` |
| **AwsLambda** (optional) | `authorizers[].authorizerUri` | `status.outputs.function_arn` |
| **AwsIamRole** (optional) | `authorizers[].authorizerCredentialsArn` | `status.outputs.role_arn` |
| **AwsCognitoUserPool** (optional) | `authorizers[].jwtConfiguration.issuer` | `status.outputs.issuer` |
| **AwsCognitoUserPoolClient** (optional) | `authorizers[].jwtConfiguration.audiences[]` | `status.outputs.client_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `api_id` | API Gateway API identifier | Resource ARN construction, CloudWatch metrics |
| `api_endpoint` | Default invoke URL | Client configuration, DNS CNAME records |
| `api_arn` | Amazon Resource Name | IAM policies, resource-based permissions |
| `execution_arn` | Execution ARN prefix | Lambda resource-based policies for invoke permissions |
| `stage_invoke_url` | Stage-specific invoke URL | Client configuration, CloudFront origins |
| `stage_name` | Deployed stage name | Logging filters, stage-specific monitoring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single Lambda catch-all** -- A `$default` route forwarding all requests to one Lambda function. The Lambda handles routing internally. Ideal for prototyping or frameworks like Express.js or FastAPI running in Lambda. Start from the **Default Route to Lambda** preset.

**Multi-route REST API** -- Separate routes per HTTP method and path with CORS enabled for browser access. Each route targets a dedicated Lambda function. Suited for production APIs with clear endpoint separation. Start from the **Multi-Route Lambda API with CORS** preset.

**JWT-protected API** -- Routes secured with a JWT authorizer validating tokens from Cognito, Auth0, or any OIDC provider. Mix public and protected routes with OAuth scope requirements. Start from the **JWT Authorized API with Cognito** preset.

## Works With

- [**AWS Lambda**](/infra-catalog/aws-lambda) -- provides backend functions for route integrations and REQUEST authorizers
- [**AWS HTTP API VPC Link**](/infra-catalog/aws-http-api-vpc-link) -- provides the network attachment for private integrations to ALB/NLB/Cloud Map targets inside a VPC
- [**AWS HTTP API Domain**](/infra-catalog/aws-http-api-domain) -- fronts this API with a custom domain; its API mappings reference this API's `api_id` and `stage_name` outputs
- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) -- provides the JWT issuer for token validation
- [**AWS Cognito User Pool Client**](/infra-catalog/aws-cognito-user-pool-client) -- provides app client IDs as JWT audiences
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides invocation roles for Lambda authorizers and AWS service actions
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- provides the destination for API access logs