# AWS Bedrock AgentCore Gateway

Deploys an Amazon Bedrock AgentCore gateway — a managed MCP (Model Context Protocol) front door that turns your existing APIs, Lambda functions, and MCP servers into tools any MCP-speaking agent can discover and call through one authenticated URL. Inbound callers authenticate with AWS IAM (SigV4) or OIDC bearer tokens; each target carries its own outbound credentials, and every tool call can be evaluated against a Cedar policy engine before it reaches the backend. The gateway costs nothing until agents call tools — billing follows tool calls at runtime.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MCP Gateway** — the gateway itself, named from `metadata.name`, with the chosen inbound authorizer, optional KMS encryption, MCP protocol tuning (instructions, semantic tool search, session timeout, response streaming), up to two Lambda interceptors, and an optional Cedar policy engine attachment
- **Gateway Target** — one per `targets` entry: an AgentCore runtime, an API Gateway REST API stage, a Lambda function with explicit tool schemas, a remote MCP server, or tools derived from an OpenAPI 3 or Smithy schema. AWS deletes a gateway's targets automatically before the gateway itself at destroy

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AgentCore control-plane permissions (`bedrock-agentcore:CreateGateway` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Bedrock AgentCore available in the target region
- An IAM role trusting `bedrock-agentcore.amazonaws.com` that can reach your targets — invoke Lambdas, call API Gateway stages, sign SigV4 requests — wired as `roleArn`
- An OIDC provider serving `/.well-known/openid-configuration` (only for `authorizerType: CUSTOM_JWT`)
- AgentCore Identity credential providers holding backend API keys or OAuth clients (only for targets with `apiKey` or `oauth` outbound credentials)

## Deploy

### Console

Open the deployment store, find **AWS Bedrock AgentCore Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the region and inbound authorizer, and the target backends. Start from the **OpenAPI Tools Gateway** preset in the [Presets](#presets) tab for the workhorse shape — a REST API's schema turned into tools — or the **JWT Gateway with an OAuth Backend** preset for OIDC inbound auth with vaulted outbound credentials.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreGateway
metadata:
  name: orders-tools-gateway
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: MCP front door for the orders API
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: agentcore-gateway-role
      fieldPath: status.outputs.role_arn
  authorizerType: AWS_IAM
  mcp:
    instructions: Tools for querying and managing customer orders.
    enableSemanticSearch: true
  targets:
    - name: orders-api
      description: The orders REST API as MCP tools
      backend:
        openApiSchema:
          s3:
            uri: s3://acme-schema-bucket/apis/orders-openapi.json
      credentials:
        gatewayIamRole:
          service: execute-api
```

```shell
planton apply -f agentcore-gateway.yaml
```

This creates an IAM-authenticated MCP gateway with one target that derives a tool per route from the OpenAPI document and signs backend calls with the gateway's role. An Infra Job tracks the provisioning in real time.

### InfraChart

When the gateway deploys alongside its execution role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: agentcore-gateway-role
      fieldPath: status.outputs.role_arn
  authorizerType: AWS_IAM
```

The InfraPipeline resolves the dependency graph, creates the role first, then provisions the gateway assuming it.

## Key Configuration

These are the most important decisions when configuring a gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**authorizerType is a one-way door** — changing the inbound authorizer replaces the gateway, which changes `gateway_url` and breaks every configured MCP client. Decide between AWS_IAM (SigV4 callers) and CUSTOM_JWT (OIDC bearer tokens, requires `customJwtAuthorizer`) before anything points at the URL. NONE disables inbound auth entirely — an open tool endpoint is rarely what production wants.

**Exactly one backend arm per target** — an AgentCore runtime, an API Gateway stage, a Lambda with explicit tool schemas, a remote MCP server, or an OpenAPI/Smithy schema. OpenAPI documents are validated server-side at target create: the document must carry a non-empty `servers` array whose URL uses HTTPS, or the target lands FAILED with named validation errors. Nothing calls the URL at create — it just has to be present and HTTPS.

**Outbound credentials belong in AgentCore Identity** — API keys and OAuth clients live as Identity credential providers, and targets reference the provider ARN. Rotating a credential never touches the gateway. `jwtPassthrough` and `callerIamCredentials` instead forward the caller's own identity to the backend — pick per target, at most one credential arm each.

**Tool descriptions are model-facing prose** — the description on every tool, schema property, and the gateway's `mcp.instructions` is what the model reads to decide when and how to call. Write them like documentation for a sharp intern, not like code comments; a vague description is a tool the agent misuses or ignores.

**Turn on semantic search past a dozen tools** — without `enableSemanticSearch`, agents list every tool into their context window on connect. With it, they query for the relevant ones. Large API Gateway or OpenAPI targets can derive dozens of tools from routes, so this matters earlier than expected.

**Remote MCP servers: DEFAULT vs DYNAMIC listing** — DEFAULT snapshots the server's tools at create and sync, so the server must be reachable at deploy time. DYNAMIC queries the server live on every list — use it when the tool set changes or availability at deploy time is not guaranteed.

**Roll out Cedar in LOG_ONLY first** — ENFORCE blocks denied tool calls immediately, mid-conversation. LOG_ONLY records what policies would decide without blocking; read the decisions, then flip the mode. And leave `exposeDebugExceptions` off in production — verbose exception detail leaks backend internals to callers.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** | `kmsKeyArn` | `status.outputs.key_arn` |
| **AwsLambda** | `interceptors[].lambdaArn`, `targets[].backend.lambda.lambdaArn` | `status.outputs.function_arn` |
| **AwsBedrockAgentCoreIdentity** | `policyEngine.policyEngineArn` | `status.outputs.policy_engine_arn` |
| **AwsBedrockAgentCoreRuntime** | `targets[].backend.agentcoreRuntime.agentRuntimeArn` | `status.outputs.agent_runtime_arn` |

Target credential providers (`credentials.apiKey.providerArn`, `credentials.oauth.providerArn`) also commonly reference an **AwsBedrockAgentCoreIdentity** resource's provider-ARN output maps, and private endpoints for JWT providers and backends take **AwsVpc**, **AwsSubnet**, and **AwsSecurityGroup** references.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `gateway_url` | The MCP URL agents connect to | MCP client configuration — AgentCore runtime agents, IDEs, any MCP-speaking client |
| `gateway_arn` | The gateway's ARN — the canonical key for IAM policies | AgentCore Evaluation harness `agentcoreGateway` tools; IAM policies scoping who may invoke |
| `gateway_id` | The unique gateway identifier | CLI and API operations addressing the gateway |
| `target_ids` | Target IDs keyed by each `targets` entry's name | Operational tooling addressing individual targets |

`workload_identity_arn` — the workload identity AWS created for the gateway — is also exported; it mostly serves JWT `allowedWorkloads` restrictions on other AgentCore resources rather than composition wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**REST API as tools** — an OpenAPI 3 schema target (inline or from S3) with `gatewayIamRole` credentials signing for `execute-api`. The gateway derives one tool per route; tool filters narrow the surface and tool overrides rename or re-describe routes the model gets wrong. Start from the **OpenAPI Tools Gateway** preset.

**OIDC gateway with vaulted outbound credentials** — `authorizerType: CUSTOM_JWT` validating your identity provider's tokens inbound, with targets authenticating outbound through AgentCore Identity OAuth providers. Neither side's secrets ever appear in the manifest. Start from the **JWT Gateway with an OAuth Backend** preset.

**Lambda tool target** — one function fulfilling several explicitly-schema'd tools. The typed schema tree nests three levels, then bottoms out in raw-JSON leaves (`itemsJson` / `propertiesJson`) — exactly where AWS's own configuration surface stops. Best when the tools are purpose-built for agents rather than derived from an existing API.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role the gateway assumes to reach its targets, wired via `roleArn`
- [**AWS Lambda**](/infra-catalog/aws-lambda) — tool-fulfilling functions and request/response interceptors
- [**AWS Bedrock AgentCore Identity**](/infra-catalog/aws-bedrock-agent-core-identity) — vaulted API keys, OAuth providers, and the Cedar policy engine the gateway evaluates against
- [**AWS Bedrock AgentCore Runtime**](/infra-catalog/aws-bedrock-agent-core-runtime) — agent runtimes exposed as tools through `agentcoreRuntime` targets, and the agents that consume this gateway's `gateway_url`
- [**AWS Bedrock AgentCore Evaluation**](/infra-catalog/aws-bedrock-agent-core-evaluation) — harnesses call tools through the gateway via its `gateway_arn`
- [**AWS REST API Gateway**](/infra-catalog/aws-rest-api-gateway) — REST API stages fronted as tools through `apiGateway` targets
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for the gateway's data at rest
