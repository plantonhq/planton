# AwsBedrockAgentCoreGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsBedrockAgentCoreGatewaySpec defines the desired configuration for an
Amazon Bedrock AgentCore gateway - a managed MCP (Model Context
Protocol) front door that turns your existing APIs, Lambda functions,
and MCP servers into tools any MCP-speaking agent can discover and
call through ONE authenticated URL.

The gateway's name is taken from `metadata.name` (letters and digits
with single hyphens, max 100 characters - AWS rejects underscores,
consecutive hyphens, and a leading/trailing hyphen).

Each `targets` entry is one backend exposed through the gateway. AWS
deletes a gateway's targets automatically before the gateway itself at
destroy. Gateways are free to create; AWS bills per tool-call at
runtime.

## Example

```yaml
# Canonical AwsBedrockAgentCoreGateway example (hack/dev manifest and
# refgen Example source): a JWT-authorized MCP gateway exercising every
# target arm -- an agent runtime, an API Gateway stage, a Lambda with an
# inline three-level tool schema (raw-JSON leaves at the bottom), a remote
# MCP server, and inline OpenAPI/Smithy schemas -- plus interceptors, a
# policy engine, and per-target credentials. Literal ARNs/ids stand in for
# composed references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreGateway
metadata:
  name: support-tools-gateway
  id: support-tools-gateway
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: MCP front door for the support agent's tools
  roleArn:
    value: arn:aws:iam::123456789012:role/agentcore-gateway-role
  authorizerType: CUSTOM_JWT
  customJwtAuthorizer:
    discoveryUrl: https://accounts.google.com/.well-known/openid-configuration
    allowedAudience:
      - support-agents
    allowedClients:
      - support-client
    customClaims:
      - claimName: team
        valueType: STRING_ARRAY
        matchOperator: CONTAINS_ANY
        matchValues:
          - support
          - platform
  kmsKeyArn:
    value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
  exposeDebugExceptions: false
  mcp:
    instructions: Tools for querying and managing customer orders.
    enableSemanticSearch: true
    supportedVersions:
      - "2025-03-26"
    sessionTimeoutSeconds: 900
    enableResponseStreaming: true
  interceptors:
    - interceptionPoints:
        - REQUEST
      lambdaArn:
        value: arn:aws:lambda:us-west-2:123456789012:function:rewrite-requests
      passRequestHeaders: true
  policyEngine:
    policyEngineArn:
      value: arn:aws:bedrock-agentcore:us-west-2:123456789012:policy-engine/pe-0123456789
    mode: LOG_ONLY
  targets:
    - name: support-runtime
      description: The support agent runtime behind this gateway
      backend:
        agentcoreRuntime:
          agentRuntimeArn:
            value: arn:aws:bedrock-agentcore:us-west-2:123456789012:runtime/support_agent-Ab1Cd2Ef3G
          qualifier: DEFAULT
      credentials:
        gatewayIamRole:
          service: bedrock-agentcore
    - name: orders-rest-api
      backend:
        apiGateway:
          restApiId: a1b2c3d4e5
          stage: prod
          toolFilters:
            - filterPath: /orders/*
              methods:
                - GET
                - POST
          toolOverrides:
            - path: /orders/{id}
              method: GET
              name: get_order
              description: Fetch one order by its ID.
      credentials:
        callerIamCredentials:
          service: execute-api
    - name: order-tools
      backend:
        lambda:
          lambdaArn:
            value: arn:aws:lambda:us-west-2:123456789012:function:order-tools
          tools:
            - name: list_orders
              description: List a customer's orders with optional filters.
              inputSchema:
                type: object
                description: Order listing filters
                properties:
                  - name: customer_id
                    type: string
                    required: true
                  - name: filters
                    type: array
                    items:
                      type: object
                      properties:
                        - name: field
                          type: string
                          required: true
                        - name: values
                          type: array
                          itemsJson:
                            type: string
              outputSchema:
                type: array
                items:
                  type: object
                  properties:
                    - name: order_id
                      type: string
                    - name: line_items
                      type: array
                      itemsJson:
                        type: object
                        properties:
                          sku:
                            type: string
                          quantity:
                            type: integer
      credentials:
        gatewayIamRole:
          service: lambda
      metadata:
        allowedRequestHeaders:
          - X-Trace-Id
        allowedResponseHeaders:
          - X-Request-Id
    - name: docs-mcp
      backend:
        mcpServer:
          endpoint: https://mcp.example.com/mcp
          listingMode: DYNAMIC
      credentials:
        oauth:
          providerArn:
            value: arn:aws:bedrock-agentcore:us-west-2:123456789012:token-vault/default/oauth2credentialprovider/docs
          scopes:
            - read:docs
          grantType: CLIENT_CREDENTIALS
    - name: billing-openapi
      backend:
        openApiSchema:
          # AWS validates the document at target create: a non-empty
          # `servers` array with an HTTPS URL is required (no call is
          # made to it at create).
          inlinePayload: '{"openapi":"3.0.0","info":{"title":"billing","version":"1"},"servers":[{"url":"https://billing.example.com"}],"paths":{}}'
      credentials:
        apiKey:
          providerArn:
            value: arn:aws:bedrock-agentcore:us-west-2:123456789012:token-vault/default/apikeycredentialprovider/billing
          credentialLocation: HEADER
          credentialParameterName: X-Api-Key
    - name: inventory-smithy
      backend:
        smithyModel:
          s3:
            uri: s3://my-schema-bucket/models/inventory.smithy
      credentials:
        jwtPassthrough: {}
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.authorizerType` | `string` |  |  |  |
| `spec.customJwtAuthorizer` | `AwsBedrockAgentCoreGatewayJwtAuthorizer` |  |  |  |
| `spec.customJwtAuthorizer.discoveryUrl` | `string` | yes |  |  |
| `spec.customJwtAuthorizer.allowedAudience` | `[]string` |  |  |  |
| `spec.customJwtAuthorizer.allowedClients` | `[]string` |  |  |  |
| `spec.customJwtAuthorizer.allowedScopes` | `[]string` |  |  |  |
| `spec.customJwtAuthorizer.allowedWorkloads` | `AwsBedrockAgentCoreGatewayAllowedWorkloads` |  |  |  |
| `spec.customJwtAuthorizer.allowedWorkloads.workloadIdentities` | `[]string` | yes |  |  |
| `spec.customJwtAuthorizer.allowedWorkloads.hostingEnvironmentArns` | `[]string` | yes |  |  |
| `spec.customJwtAuthorizer.customClaims` | `[]AwsBedrockAgentCoreGatewayCustomClaim` |  |  |  |
| `spec.customJwtAuthorizer.customClaims[].claimName` | `string` | yes |  |  |
| `spec.customJwtAuthorizer.customClaims[].valueType` | `string` |  |  |  |
| `spec.customJwtAuthorizer.customClaims[].matchOperator` | `string` |  |  |  |
| `spec.customJwtAuthorizer.customClaims[].matchValue` | `string` |  |  |  |
| `spec.customJwtAuthorizer.customClaims[].matchValues` | `[]string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint` | `AwsBedrockAgentCoreGatewayPrivateEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc` | `AwsBedrockAgentCoreGatewayManagedVpcEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.endpointIpAddressType` | `string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.routingDomain` | `string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.tags` | `map<string, string>` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.selfManagedLattice` | `AwsBedrockAgentCoreGatewayLatticeEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.selfManagedLattice.resourceConfigurationId` | `string` | yes |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides` | `[]AwsBedrockAgentCoreGatewayPrivateEndpointOverride` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].domain` | `string` | yes |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint` | `AwsBedrockAgentCoreGatewayPrivateEndpoint` | yes |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc` | `AwsBedrockAgentCoreGatewayManagedVpcEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.endpointIpAddressType` | `string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.routingDomain` | `string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.tags` | `map<string, string>` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice` | `AwsBedrockAgentCoreGatewayLatticeEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice.resourceConfigurationId` | `string` | yes |  |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.exposeDebugExceptions` | `bool` |  |  |  |
| `spec.mcp` | `AwsBedrockAgentCoreGatewayMcp` |  |  |  |
| `spec.mcp.instructions` | `string` |  |  |  |
| `spec.mcp.enableSemanticSearch` | `bool` |  |  |  |
| `spec.mcp.supportedVersions` | `[]string` |  |  |  |
| `spec.mcp.sessionTimeoutSeconds` | `int64` |  |  |  |
| `spec.mcp.enableResponseStreaming` | `bool` |  |  |  |
| `spec.interceptors` | `[]AwsBedrockAgentCoreGatewayInterceptor` |  |  |  |
| `spec.interceptors[].interceptionPoints` | `[]string` | yes |  |  |
| `spec.interceptors[].lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.interceptors[].passRequestHeaders` | `bool` |  |  |  |
| `spec.policyEngine` | `AwsBedrockAgentCoreGatewayPolicyEngine` |  |  |  |
| `spec.policyEngine.policyEngineArn` | `string \| valueFrom` | yes |  | AwsBedrockAgentCoreIdentity (`status.outputs.policy_engine_arn`) |
| `spec.policyEngine.mode` | `string` |  |  |  |
| `spec.targets` | `[]AwsBedrockAgentCoreGatewayTarget` |  |  |  |
| `spec.targets[].name` | `string` | yes |  |  |
| `spec.targets[].description` | `string` |  |  |  |
| `spec.targets[].backend` | `AwsBedrockAgentCoreGatewayTargetBackend` | yes |  |  |
| `spec.targets[].backend.agentcoreRuntime` | `AwsBedrockAgentCoreGatewayRuntimeTarget` |  |  |  |
| `spec.targets[].backend.agentcoreRuntime.agentRuntimeArn` | `string \| valueFrom` | yes |  | AwsBedrockAgentCoreRuntime (`status.outputs.agent_runtime_arn`) |
| `spec.targets[].backend.agentcoreRuntime.qualifier` | `string` |  |  |  |
| `spec.targets[].backend.apiGateway` | `AwsBedrockAgentCoreGatewayApiGatewayTarget` |  |  |  |
| `spec.targets[].backend.apiGateway.restApiId` | `string` | yes |  |  |
| `spec.targets[].backend.apiGateway.stage` | `string` | yes |  |  |
| `spec.targets[].backend.apiGateway.toolFilters` | `[]AwsBedrockAgentCoreGatewayApiGatewayToolFilter` |  |  |  |
| `spec.targets[].backend.apiGateway.toolFilters[].filterPath` | `string` | yes |  |  |
| `spec.targets[].backend.apiGateway.toolFilters[].methods` | `[]string` | yes |  |  |
| `spec.targets[].backend.apiGateway.toolOverrides` | `[]AwsBedrockAgentCoreGatewayApiGatewayToolOverride` |  |  |  |
| `spec.targets[].backend.apiGateway.toolOverrides[].path` | `string` | yes |  |  |
| `spec.targets[].backend.apiGateway.toolOverrides[].method` | `string` |  |  |  |
| `spec.targets[].backend.apiGateway.toolOverrides[].name` | `string` | yes |  |  |
| `spec.targets[].backend.apiGateway.toolOverrides[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda` | `AwsBedrockAgentCoreGatewayLambdaTarget` |  |  |  |
| `spec.targets[].backend.lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.targets[].backend.lambda.tools` | `[]AwsBedrockAgentCoreGatewayToolDefinition` |  |  |  |
| `spec.targets[].backend.lambda.tools[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].description` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema` | `AwsBedrockAgentCoreGatewaySchemaDefinition` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties` | `[]AwsBedrockAgentCoreGatewaySchemaProperty` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].required` | `bool` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items` | `AwsBedrockAgentCoreGatewaySchemaItems` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items` | `AwsBedrockAgentCoreGatewaySchemaItemsLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items.itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items.propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties` | `[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].required` | `bool` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties` | `[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].required` | `bool` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items` | `AwsBedrockAgentCoreGatewaySchemaItems` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.items` | `AwsBedrockAgentCoreGatewaySchemaItemsLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.items.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.items.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.items.itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.items.propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.properties` | `[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].required` | `bool` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema` | `AwsBedrockAgentCoreGatewaySchemaDefinition` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties` | `[]AwsBedrockAgentCoreGatewaySchemaProperty` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].required` | `bool` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items` | `AwsBedrockAgentCoreGatewaySchemaItems` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items` | `AwsBedrockAgentCoreGatewaySchemaItemsLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items.itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items.propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties` | `[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].required` | `bool` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties` | `[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].required` | `bool` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items` | `AwsBedrockAgentCoreGatewaySchemaItems` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.items` | `AwsBedrockAgentCoreGatewaySchemaItemsLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.items.type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.items.description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.items.itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.items.propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.properties` | `[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].name` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].type` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].description` | `string` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].required` | `bool` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].itemsJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].propertiesJson` | `object` |  |  |  |
| `spec.targets[].backend.lambda.toolsS3` | `AwsBedrockAgentCoreGatewaySchemaS3` |  |  |  |
| `spec.targets[].backend.lambda.toolsS3.uri` | `string` | yes |  |  |
| `spec.targets[].backend.lambda.toolsS3.bucketOwnerAccountId` | `string` |  |  |  |
| `spec.targets[].backend.mcpServer` | `AwsBedrockAgentCoreGatewayMcpServerTarget` |  |  |  |
| `spec.targets[].backend.mcpServer.endpoint` | `string` |  |  |  |
| `spec.targets[].backend.mcpServer.listingMode` | `string` |  |  |  |
| `spec.targets[].backend.openApiSchema` | `AwsBedrockAgentCoreGatewaySchemaTarget` |  |  |  |
| `spec.targets[].backend.openApiSchema.inlinePayload` | `string` |  |  |  |
| `spec.targets[].backend.openApiSchema.s3` | `AwsBedrockAgentCoreGatewaySchemaS3` |  |  |  |
| `spec.targets[].backend.openApiSchema.s3.uri` | `string` | yes |  |  |
| `spec.targets[].backend.openApiSchema.s3.bucketOwnerAccountId` | `string` |  |  |  |
| `spec.targets[].backend.smithyModel` | `AwsBedrockAgentCoreGatewaySchemaTarget` |  |  |  |
| `spec.targets[].backend.smithyModel.inlinePayload` | `string` |  |  |  |
| `spec.targets[].backend.smithyModel.s3` | `AwsBedrockAgentCoreGatewaySchemaS3` |  |  |  |
| `spec.targets[].backend.smithyModel.s3.uri` | `string` | yes |  |  |
| `spec.targets[].backend.smithyModel.s3.bucketOwnerAccountId` | `string` |  |  |  |
| `spec.targets[].credentials` | `AwsBedrockAgentCoreGatewayTargetCredentials` |  |  |  |
| `spec.targets[].credentials.apiKey` | `AwsBedrockAgentCoreGatewayApiKeyCredentials` |  |  |  |
| `spec.targets[].credentials.apiKey.providerArn` | `string \| valueFrom` | yes |  |  |
| `spec.targets[].credentials.apiKey.credentialLocation` | `string` |  |  |  |
| `spec.targets[].credentials.apiKey.credentialParameterName` | `string` |  |  |  |
| `spec.targets[].credentials.apiKey.credentialPrefix` | `string` |  |  |  |
| `spec.targets[].credentials.callerIamCredentials` | `AwsBedrockAgentCoreGatewaySigv4Credentials` |  |  |  |
| `spec.targets[].credentials.callerIamCredentials.service` | `string` |  |  |  |
| `spec.targets[].credentials.callerIamCredentials.region` | `string` |  |  |  |
| `spec.targets[].credentials.gatewayIamRole` | `AwsBedrockAgentCoreGatewaySigv4Credentials` |  |  |  |
| `spec.targets[].credentials.gatewayIamRole.service` | `string` |  |  |  |
| `spec.targets[].credentials.gatewayIamRole.region` | `string` |  |  |  |
| `spec.targets[].credentials.jwtPassthrough` | `AwsBedrockAgentCoreGatewayJwtPassthroughCredentials` |  |  |  |
| `spec.targets[].credentials.oauth` | `AwsBedrockAgentCoreGatewayOauthCredentials` |  |  |  |
| `spec.targets[].credentials.oauth.providerArn` | `string \| valueFrom` | yes |  |  |
| `spec.targets[].credentials.oauth.scopes` | `[]string` | yes |  |  |
| `spec.targets[].credentials.oauth.grantType` | `string` |  |  |  |
| `spec.targets[].credentials.oauth.defaultReturnUrl` | `string` |  |  |  |
| `spec.targets[].credentials.oauth.customParameters` | `map<string, string>` |  |  |  |
| `spec.targets[].metadata` | `AwsBedrockAgentCoreGatewayTargetMetadata` |  |  |  |
| `spec.targets[].metadata.allowedQueryParameters` | `[]string` |  |  |  |
| `spec.targets[].metadata.allowedRequestHeaders` | `[]string` |  |  |  |
| `spec.targets[].metadata.allowedResponseHeaders` | `[]string` |  |  |  |
| `spec.targets[].metadata.enabled` | `bool` |  | `true` |  |
| `spec.targets[].privateEndpoint` | `AwsBedrockAgentCoreGatewayPrivateEndpoint` |  |  |  |
| `spec.targets[].privateEndpoint.managedVpc` | `AwsBedrockAgentCoreGatewayManagedVpcEndpoint` |  |  |  |
| `spec.targets[].privateEndpoint.managedVpc.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.targets[].privateEndpoint.managedVpc.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.targets[].privateEndpoint.managedVpc.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.targets[].privateEndpoint.managedVpc.endpointIpAddressType` | `string` |  |  |  |
| `spec.targets[].privateEndpoint.managedVpc.routingDomain` | `string` |  |  |  |
| `spec.targets[].privateEndpoint.managedVpc.tags` | `map<string, string>` |  |  |  |
| `spec.targets[].privateEndpoint.selfManagedLattice` | `AwsBedrockAgentCoreGatewayLatticeEndpoint` |  |  |  |
| `spec.targets[].privateEndpoint.selfManagedLattice.resourceConfigurationId` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the gateway will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description shown in the AgentCore console (1-200
characters when set). Updates in place.

- rule: {"string":{"maxLen":"200"}}

### spec.roleArn

`string | valueFrom` · required

IAM role the gateway assumes to reach its targets (invoke Lambdas,
call API Gateway stages, sign SigV4 requests). The role must trust
bedrock-agentcore.amazonaws.com. Changing the role updates in place.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.authorizerType

`string`

How inbound callers authenticate to the gateway. AWS_IAM takes SigV4
credentials; CUSTOM_JWT validates OIDC bearer tokens (requires
`custom_jwt_authorizer`); NONE disables inbound auth; AUTHENTICATE_ONLY
validates identity without authorization. Changing it replaces the
gateway.

- rule: {"string":{"in":["AWS_IAM","CUSTOM_JWT","NONE","AUTHENTICATE_ONLY"]}}

### spec.customJwtAuthorizer

`AwsBedrockAgentCoreGatewayJwtAuthorizer`

OIDC token validation rules - required when authorizer_type is
CUSTOM_JWT.

### spec.customJwtAuthorizer.discoveryUrl

`string` · required

The provider's OIDC discovery URL (must serve
/.well-known/openid-configuration).

- rule: {"string":{"minLen":"1"}}

### spec.customJwtAuthorizer.allowedAudience

`[]string`

Accepted "aud" claim values. A token must match at least one entry of
at least one configured allow-list.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.customJwtAuthorizer.allowedClients

`[]string`

Accepted OAuth client IDs (the "client_id"/"azp" claim).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.customJwtAuthorizer.allowedScopes

`[]string`

Accepted OAuth scopes.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.customJwtAuthorizer.allowedWorkloads

`AwsBedrockAgentCoreGatewayAllowedWorkloads`

Restrict which AgentCore workload identities may present tokens.

### spec.customJwtAuthorizer.allowedWorkloads.workloadIdentities

`[]string` · required

Workload identity names allowed to call (1-10).

- rule: {"repeated":{"minItems":"1","maxItems":"10","items":{"string":{"minLen":"1"}}}}

### spec.customJwtAuthorizer.allowedWorkloads.hostingEnvironmentArns

`[]string` · required

Hosting environment ARNs allowed to call (1-10).

- rule: {"repeated":{"minItems":"1","maxItems":"10","items":{"string":{"minLen":"1"}}}}

### spec.customJwtAuthorizer.customClaims

`[]AwsBedrockAgentCoreGatewayCustomClaim`

Additional claim-matching rules a token must satisfy beyond the
standard audience/client/scope checks.

- rule: custom claim must set exactly one of match_value or match_values

### spec.customJwtAuthorizer.customClaims[].claimName

`string` · required

The inbound token claim to inspect (1-255 characters; letters,
digits, and _ . - :).

- rule: {"string":{"minLen":"1","maxLen":"255","pattern":"^[A-Za-z0-9_.:-]+$"}}

### spec.customJwtAuthorizer.customClaims[].valueType

`string`

Whether the claim's value is a single STRING or a STRING_ARRAY.

- rule: {"string":{"in":["STRING","STRING_ARRAY"]}}

### spec.customJwtAuthorizer.customClaims[].matchOperator

`string`

How the claim value is compared: EQUALS (exact), CONTAINS (the value
appears), or CONTAINS_ANY (any of the expected values appears).

- rule: {"string":{"in":["EQUALS","CONTAINS","CONTAINS_ANY"]}}

### spec.customJwtAuthorizer.customClaims[].matchValue

`string`

Expected value when matching a single string.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255","pattern":"^[A-Za-z0-9_.:-]+$"}}

### spec.customJwtAuthorizer.customClaims[].matchValues

`[]string`

Expected values when matching against a list.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"255","pattern":"^[A-Za-z0-9_.:-]+$"}}}}

### spec.customJwtAuthorizer.privateEndpoint

`AwsBedrockAgentCoreGatewayPrivateEndpoint`

Reach a PRIVATE OIDC provider through your VPC instead of the public
internet.

- rule: private endpoint must set exactly one of managed_vpc or self_managed_lattice

### spec.customJwtAuthorizer.privateEndpoint.managedVpc

`AwsBedrockAgentCoreGatewayManagedVpcEndpoint`

AWS manages VPC endpoints in your subnets.

### spec.customJwtAuthorizer.privateEndpoint.managedVpc.vpcId

`string | valueFrom` · required

The VPC to route through.

Containment-exempt: a managed endpoint is a private PATH through the
VPC to a private provider or backend; the gateway is not deployed inside
it. On a diagram it stands where its own network places it, with a
line into the VPC -- the verdict a Lambda's VPC subnets and an
EventBridge pipe's task subnets already carry.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.customJwtAuthorizer.privateEndpoint.managedVpc.subnetIds

`[]string | valueFrom` · required

Subnets for the managed endpoint's network interfaces (at least one).
Containment-exempt for the same reason as the VPC above.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds

`[]string | valueFrom`

Security groups on the endpoint interfaces (max 5).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.customJwtAuthorizer.privateEndpoint.managedVpc.endpointIpAddressType

`string`

Whether the endpoint answers IPV4 or IPV6.

- rule: {"string":{"in":["IPV4","IPV6"]}}

### spec.customJwtAuthorizer.privateEndpoint.managedVpc.routingDomain

`string`

Domain the endpoint routes (3-255 characters). Omitted = derived from
the target.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.customJwtAuthorizer.privateEndpoint.managedVpc.tags

`map<string, string>`

Tags applied to the AWS-managed endpoint resources (the module always
adds the Planton identity tags).

### spec.customJwtAuthorizer.privateEndpoint.selfManagedLattice

`AwsBedrockAgentCoreGatewayLatticeEndpoint`

You bring a VPC Lattice resource configuration.

### spec.customJwtAuthorizer.privateEndpoint.selfManagedLattice.resourceConfigurationId

`string` · required

The Lattice resource-configuration identifier.

- rule: {"string":{"minLen":"1"}}

### spec.customJwtAuthorizer.privateEndpointOverrides

`[]AwsBedrockAgentCoreGatewayPrivateEndpointOverride`

Per-domain overrides of the private endpoint (max 5) - route specific
issuer domains through different private paths.

- rule: {"repeated":{"maxItems":"5"}}

### spec.customJwtAuthorizer.privateEndpointOverrides[].domain

`string` · required

The domain this override captures (1-253 characters).

- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint

`AwsBedrockAgentCoreGatewayPrivateEndpoint` · required

The private path for that domain.

- rule: {"required":true}
- rule: private endpoint must set exactly one of managed_vpc or self_managed_lattice

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc

`AwsBedrockAgentCoreGatewayManagedVpcEndpoint`

AWS manages VPC endpoints in your subnets.

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId

`string | valueFrom` · required

The VPC to route through.

Containment-exempt: a managed endpoint is a private PATH through the
VPC to a private provider or backend; the gateway is not deployed inside
it. On a diagram it stands where its own network places it, with a
line into the VPC -- the verdict a Lambda's VPC subnets and an
EventBridge pipe's task subnets already carry.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.subnetIds

`[]string | valueFrom` · required

Subnets for the managed endpoint's network interfaces (at least one).
Containment-exempt for the same reason as the VPC above.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds

`[]string | valueFrom`

Security groups on the endpoint interfaces (max 5).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.endpointIpAddressType

`string`

Whether the endpoint answers IPV4 or IPV6.

- rule: {"string":{"in":["IPV4","IPV6"]}}

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.routingDomain

`string`

Domain the endpoint routes (3-255 characters). Omitted = derived from
the target.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.tags

`map<string, string>`

Tags applied to the AWS-managed endpoint resources (the module always
adds the Planton identity tags).

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice

`AwsBedrockAgentCoreGatewayLatticeEndpoint`

You bring a VPC Lattice resource configuration.

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice.resourceConfigurationId

`string` · required

The Lattice resource-configuration identifier.

- rule: {"string":{"minLen":"1"}}

### spec.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key encrypting the gateway's data at rest.
Without it, AWS uses a service-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.exposeDebugExceptions

`bool`

Return verbose exception detail (stack traces, backend errors) to
callers - AWS's DEBUG exception level. Leave off in production; error
detail can leak backend internals.

### spec.mcp

`AwsBedrockAgentCoreGatewayMcp`

MCP protocol tuning for the gateway's single protocol (MCP - the
modules send that constant).

### spec.mcp.instructions

`string`

Instructions surfaced to agents describing what this gateway's tools
do (1-2048 characters when set) - write them the way you would brief
the model.

- rule: {"string":{"maxLen":"2048"}}

### spec.mcp.enableSemanticSearch

`bool`

Let agents search tools semantically instead of listing all of them
(AWS's SEMANTIC search type - the modules send that constant on
enable).

### spec.mcp.supportedVersions

`[]string`

MCP protocol versions the gateway accepts. Omitted = AWS defaults.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.mcp.sessionTimeoutSeconds

`int64`

Seconds an MCP session survives between calls (900-28800). Omitted =
AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int64":{"lte":"28800","gte":"900"}}

### spec.mcp.enableResponseStreaming

`bool`

Stream tool responses to callers as they are produced. Explicit false
and omitted both leave AWS's default off - set true to enable.

### spec.interceptors

`[]AwsBedrockAgentCoreGatewayInterceptor`

Lambda interceptors that rewrite requests and/or responses in flight
(max 2 - typically one for REQUEST, one for RESPONSE).

- rule: {"repeated":{"maxItems":"2"}}

### spec.interceptors[].interceptionPoints

`[]string` · required

Where the interceptor runs: on the inbound REQUEST, the outbound
RESPONSE, or both (at least one).

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["REQUEST","RESPONSE"]}}}}

### spec.interceptors[].lambdaArn

`string | valueFrom` · required

The Lambda function AWS invokes at the interception points.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.interceptors[].passRequestHeaders

`bool` · optional (explicit presence)

Whether the original client request headers are passed to the
interceptor. Set it to configure the input explicitly; omitted = AWS
default.

### spec.policyEngine

`AwsBedrockAgentCoreGatewayPolicyEngine`

Evaluate every tool call against a Cedar policy engine before it
reaches the target.

### spec.policyEngine.policyEngineArn

`string | valueFrom` · required

The policy engine to evaluate against.

- references: AwsBedrockAgentCoreIdentity (`status.outputs.policy_engine_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockAgentCoreIdentity, name: <that resource's name>, fieldPath: status.outputs.policy_engine_arn}} -- a bare string does not parse

### spec.policyEngine.mode

`string`

LOG_ONLY records what policies WOULD decide without blocking calls
(the safe rollout mode); ENFORCE blocks denied calls.

- rule: {"string":{"in":["LOG_ONLY","ENFORCE"]}}

### spec.targets

`[]AwsBedrockAgentCoreGatewayTarget`

The backends this gateway exposes as MCP tools.

### spec.targets[].name

`string` · required

Target name (letters and digits with single hyphens, max 100
characters). The for_each key on both engines and the key in the
`target_ids` output map; also the target name in AWS. Changing it
replaces the target.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][-]?){1,100}$"}}

### spec.targets[].description

`string`

Human-readable description (1-200 characters when set).

- rule: {"string":{"maxLen":"200"}}

### spec.targets[].backend

`AwsBedrockAgentCoreGatewayTargetBackend` · required

The backend this target fronts - exactly one arm.

- rule: {"required":true}
- rule: backend must set exactly one of agentcore_runtime, api_gateway, lambda, mcp_server, open_api_schema, or smithy_model

### spec.targets[].backend.agentcoreRuntime

`AwsBedrockAgentCoreGatewayRuntimeTarget`

Front an AgentCore agent runtime over plain HTTP.

### spec.targets[].backend.agentcoreRuntime.agentRuntimeArn

`string | valueFrom` · required

The agent runtime to invoke.

- references: AwsBedrockAgentCoreRuntime (`status.outputs.agent_runtime_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockAgentCoreRuntime, name: <that resource's name>, fieldPath: status.outputs.agent_runtime_arn}} -- a bare string does not parse

### spec.targets[].backend.agentcoreRuntime.qualifier

`string`

Endpoint or version qualifier to invoke (e.g. "DEFAULT" or an
endpoint name). Omitted = the runtime's default.

### spec.targets[].backend.apiGateway

`AwsBedrockAgentCoreGatewayApiGatewayTarget`

Front an API Gateway REST API - the gateway derives one MCP tool per
route.

### spec.targets[].backend.apiGateway.restApiId

`string` · required

The REST API's identifier.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.apiGateway.stage

`string` · required

The deployed stage to call (e.g. "prod").

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.apiGateway.toolFilters

`[]AwsBedrockAgentCoreGatewayApiGatewayToolFilter`

Only expose routes matching these filters. Omitted = every route
becomes a tool.

### spec.targets[].backend.apiGateway.toolFilters[].filterPath

`string` · required

Route path pattern to include (e.g. "/orders/*").

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.apiGateway.toolFilters[].methods

`[]string` · required

HTTP methods to include on the matched paths (at least one).

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["GET","POST","PUT","PATCH","DELETE","HEAD","OPTIONS"]}}}}

### spec.targets[].backend.apiGateway.toolOverrides

`[]AwsBedrockAgentCoreGatewayApiGatewayToolOverride`

Rename or re-describe specific routes' derived tools.

### spec.targets[].backend.apiGateway.toolOverrides[].path

`string` · required

The route path the override applies to.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.apiGateway.toolOverrides[].method

`string`

The route method the override applies to.

- rule: {"string":{"in":["GET","POST","PUT","PATCH","DELETE","HEAD","OPTIONS"]}}

### spec.targets[].backend.apiGateway.toolOverrides[].name

`string` · required

The tool name the model sees.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.apiGateway.toolOverrides[].description

`string`

The tool description the model reads - the better the description,
the better the model's tool use.

### spec.targets[].backend.lambda

`AwsBedrockAgentCoreGatewayLambdaTarget`

Front a Lambda function - you define the tools and their JSON
schemas; the gateway invokes the function with the tool call.

- rule: lambda target must set exactly one of tools or tools_s3
- rule: tool names must be unique

### spec.targets[].backend.lambda.lambdaArn

`string | valueFrom` · required

The Lambda function that fulfills every tool in this target.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.targets[].backend.lambda.tools

`[]AwsBedrockAgentCoreGatewayToolDefinition`

The tools the function implements, defined inline (name, description,
JSON-schema input/output).

### spec.targets[].backend.lambda.tools[].name

`string` · required

Tool name the model calls.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].description

`string` · required

What the tool does, shown to the model (required by AWS) - the better
the description, the better the model's tool use.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].inputSchema

`AwsBedrockAgentCoreGatewaySchemaDefinition` · required

JSON schema of the tool's input (required by AWS).

- rule: {"required":true}
- rule: a schema node takes properties or items, not both
- rule: property names must be unique

### spec.targets[].backend.lambda.tools[].inputSchema.type

`string`

JSON type at the schema root.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.description

`string`

What this value means, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.properties

`[]AwsBedrockAgentCoreGatewaySchemaProperty`

Named properties when type is "object".

- rule: a schema node takes properties or items, not both
- rule: nested property names must be unique

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].name

`string` · required

Property name.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].type

`string`

JSON type of the property.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].description

`string`

What the property means, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].required

`bool`

Whether the model must supply this property.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items

`AwsBedrockAgentCoreGatewaySchemaItems`

Element schema when type is "array".

- rule: a schema node takes properties or items, not both
- rule: nested property names must be unique

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.type

`string`

JSON type of the elements.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.description

`string`

What the elements mean, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items

`AwsBedrockAgentCoreGatewaySchemaItemsLeaf`

Element schema when the elements are arrays (leaf depth).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items.type

`string`

JSON type of the elements.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items.description

`string`

What the elements mean, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items.itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.items.propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties

`[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf`

Named properties when the elements are objects (leaf depth).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].name

`string` · required

Property name.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].type

`string`

JSON type of the property.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].description

`string`

What the property means, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].required

`bool`

Whether the model must supply this property.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].items.properties[].propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties

`[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf`

Nested properties when type is "object" (leaf depth - deeper shapes
go in each entry's raw-JSON fields).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].name

`string` · required

Property name.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].type

`string`

JSON type of the property.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].description

`string`

What the property means, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].required

`bool`

Whether the model must supply this property.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].inputSchema.properties[].properties[].propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].inputSchema.items

`AwsBedrockAgentCoreGatewaySchemaItems`

Element schema when type is "array".

- rule: a schema node takes properties or items, not both
- rule: nested property names must be unique

### spec.targets[].backend.lambda.tools[].inputSchema.items.type

`string`

JSON type of the elements.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.items.description

`string`

What the elements mean, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.items.items

`AwsBedrockAgentCoreGatewaySchemaItemsLeaf`

Element schema when the elements are arrays (leaf depth).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].inputSchema.items.items.type

`string`

JSON type of the elements.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.items.items.description

`string`

What the elements mean, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.items.items.itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].inputSchema.items.items.propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].inputSchema.items.properties

`[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf`

Named properties when the elements are objects (leaf depth).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].name

`string` · required

Property name.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].type

`string`

JSON type of the property.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].description

`string`

What the property means, shown to the model.

### spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].required

`bool`

Whether the model must supply this property.

### spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].inputSchema.items.properties[].propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].outputSchema

`AwsBedrockAgentCoreGatewaySchemaDefinition`

JSON schema of the tool's output. Omitted = unconstrained.

- rule: a schema node takes properties or items, not both
- rule: property names must be unique

### spec.targets[].backend.lambda.tools[].outputSchema.type

`string`

JSON type at the schema root.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.description

`string`

What this value means, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.properties

`[]AwsBedrockAgentCoreGatewaySchemaProperty`

Named properties when type is "object".

- rule: a schema node takes properties or items, not both
- rule: nested property names must be unique

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].name

`string` · required

Property name.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].type

`string`

JSON type of the property.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].description

`string`

What the property means, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].required

`bool`

Whether the model must supply this property.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items

`AwsBedrockAgentCoreGatewaySchemaItems`

Element schema when type is "array".

- rule: a schema node takes properties or items, not both
- rule: nested property names must be unique

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.type

`string`

JSON type of the elements.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.description

`string`

What the elements mean, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items

`AwsBedrockAgentCoreGatewaySchemaItemsLeaf`

Element schema when the elements are arrays (leaf depth).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items.type

`string`

JSON type of the elements.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items.description

`string`

What the elements mean, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items.itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.items.propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties

`[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf`

Named properties when the elements are objects (leaf depth).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].name

`string` · required

Property name.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].type

`string`

JSON type of the property.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].description

`string`

What the property means, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].required

`bool`

Whether the model must supply this property.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].items.properties[].propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties

`[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf`

Nested properties when type is "object" (leaf depth - deeper shapes
go in each entry's raw-JSON fields).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].name

`string` · required

Property name.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].type

`string`

JSON type of the property.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].description

`string`

What the property means, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].required

`bool`

Whether the model must supply this property.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].outputSchema.properties[].properties[].propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].outputSchema.items

`AwsBedrockAgentCoreGatewaySchemaItems`

Element schema when type is "array".

- rule: a schema node takes properties or items, not both
- rule: nested property names must be unique

### spec.targets[].backend.lambda.tools[].outputSchema.items.type

`string`

JSON type of the elements.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.items.description

`string`

What the elements mean, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.items.items

`AwsBedrockAgentCoreGatewaySchemaItemsLeaf`

Element schema when the elements are arrays (leaf depth).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].outputSchema.items.items.type

`string`

JSON type of the elements.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.items.items.description

`string`

What the elements mean, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.items.items.itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].outputSchema.items.items.propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.tools[].outputSchema.items.properties

`[]AwsBedrockAgentCoreGatewaySchemaPropertyLeaf`

Named properties when the elements are objects (leaf depth).

- rule: set at most one of items_json or properties_json

### spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].name

`string` · required

Property name.

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].type

`string`

JSON type of the property.

- rule: {"string":{"in":["string","number","integer","boolean","object","array"]}}

### spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].description

`string`

What the property means, shown to the model.

### spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].required

`bool`

Whether the model must supply this property.

### spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].itemsJson

`object`

Deeper array-element schema as a raw JSON-schema object.

### spec.targets[].backend.lambda.tools[].outputSchema.items.properties[].propertiesJson

`object`

Deeper object properties as a raw JSON-schema properties object.

### spec.targets[].backend.lambda.toolsS3

`AwsBedrockAgentCoreGatewaySchemaS3`

Or point at an S3 object holding the tool definitions.

### spec.targets[].backend.lambda.toolsS3.uri

`string` · required

S3 URI of the document (e.g. "s3://my-bucket/tools/schema.json").

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.lambda.toolsS3.bucketOwnerAccountId

`string`

Expected bucket-owner account ID (cross-account safety check).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{12}$"}}

### spec.targets[].backend.mcpServer

`AwsBedrockAgentCoreGatewayMcpServerTarget`

Front an existing remote MCP server.

### spec.targets[].backend.mcpServer.endpoint

`string`

The server's HTTPS endpoint.

- rule: {"string":{"pattern":"^https://.*"}}

### spec.targets[].backend.mcpServer.listingMode

`string`

How the gateway lists the server's tools: DEFAULT snapshots them at
create/sync; DYNAMIC queries the server live on every list. Omitted =
AWS default (DEFAULT).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DEFAULT","DYNAMIC"]}}

### spec.targets[].backend.openApiSchema

`AwsBedrockAgentCoreGatewaySchemaTarget`

Derive tools from an OpenAPI 3 schema describing an HTTP API.

- rule: set exactly one of inline_payload or s3

### spec.targets[].backend.openApiSchema.inlinePayload

`string`

The schema document inline (JSON or YAML for OpenAPI; Smithy IDL or
JSON AST for Smithy). AWS validates OpenAPI content server-side when
the target creates: the document must carry a non-empty `servers`
array whose URL uses HTTPS, or the target lands FAILED with named
validation errors (live-caught 2026-08-14; nothing calls the URL at
create).

### spec.targets[].backend.openApiSchema.s3

`AwsBedrockAgentCoreGatewaySchemaS3`

Or point at an S3 object holding the schema document.

### spec.targets[].backend.openApiSchema.s3.uri

`string` · required

S3 URI of the document (e.g. "s3://my-bucket/tools/schema.json").

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.openApiSchema.s3.bucketOwnerAccountId

`string`

Expected bucket-owner account ID (cross-account safety check).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{12}$"}}

### spec.targets[].backend.smithyModel

`AwsBedrockAgentCoreGatewaySchemaTarget`

Derive tools from a Smithy model describing an API.

- rule: set exactly one of inline_payload or s3

### spec.targets[].backend.smithyModel.inlinePayload

`string`

The schema document inline (JSON or YAML for OpenAPI; Smithy IDL or
JSON AST for Smithy). AWS validates OpenAPI content server-side when
the target creates: the document must carry a non-empty `servers`
array whose URL uses HTTPS, or the target lands FAILED with named
validation errors (live-caught 2026-08-14; nothing calls the URL at
create).

### spec.targets[].backend.smithyModel.s3

`AwsBedrockAgentCoreGatewaySchemaS3`

Or point at an S3 object holding the schema document.

### spec.targets[].backend.smithyModel.s3.uri

`string` · required

S3 URI of the document (e.g. "s3://my-bucket/tools/schema.json").

- rule: {"string":{"minLen":"1"}}

### spec.targets[].backend.smithyModel.s3.bucketOwnerAccountId

`string`

Expected bucket-owner account ID (cross-account safety check).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{12}$"}}

### spec.targets[].credentials

`AwsBedrockAgentCoreGatewayTargetCredentials`

How the GATEWAY authenticates to this backend (outbound credentials).
Omitted = the gateway's own IAM role without SigV4 service signing.

### spec.targets[].credentials.apiKey

`AwsBedrockAgentCoreGatewayApiKeyCredentials`

Send an API key from an AgentCore Identity api-key credential
provider.

### spec.targets[].credentials.apiKey.providerArn

`string | valueFrom` · required

The AgentCore Identity api-key credential provider holding the key.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.targets[].credentials.apiKey.credentialLocation

`string`

Where the key travels: HEADER or QUERY_PARAMETER. Omitted = AWS
default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["HEADER","QUERY_PARAMETER"]}}

### spec.targets[].credentials.apiKey.credentialParameterName

`string`

The header or query-parameter name carrying the key (e.g.
"X-Api-Key").

### spec.targets[].credentials.apiKey.credentialPrefix

`string`

Prefix prepended to the key value (e.g. "Bearer ").

### spec.targets[].credentials.callerIamCredentials

`AwsBedrockAgentCoreGatewaySigv4Credentials`

Forward the CALLER's own IAM credentials to the backend (SigV4).

- rule: region requires service

### spec.targets[].credentials.callerIamCredentials.service

`string`

The AWS service name to sign for (e.g. "bedrock-agentcore",
"execute-api", "lambda"). Required when forwarding the caller's
credentials; optional for the gateway role (omit for plain
IAM-role-based auth without SigV4 signing).

### spec.targets[].credentials.callerIamCredentials.region

`string`

Region used for signing. Omitted = the gateway's region. Only
meaningful when `service` is set.

### spec.targets[].credentials.gatewayIamRole

`AwsBedrockAgentCoreGatewaySigv4Credentials`

Sign requests with the GATEWAY's IAM role (SigV4).

- rule: region requires service

### spec.targets[].credentials.gatewayIamRole.service

`string`

The AWS service name to sign for (e.g. "bedrock-agentcore",
"execute-api", "lambda"). Required when forwarding the caller's
credentials; optional for the gateway role (omit for plain
IAM-role-based auth without SigV4 signing).

### spec.targets[].credentials.gatewayIamRole.region

`string`

Region used for signing. Omitted = the gateway's region. Only
meaningful when `service` is set.

### spec.targets[].credentials.jwtPassthrough

`AwsBedrockAgentCoreGatewayJwtPassthroughCredentials`

Pass the caller's inbound JWT straight through to the backend.

### spec.targets[].credentials.oauth

`AwsBedrockAgentCoreGatewayOauthCredentials`

Obtain an OAuth token from an AgentCore Identity oauth2 credential
provider.

- rule: default_return_url is required when grant_type is AUTHORIZATION_CODE

### spec.targets[].credentials.oauth.providerArn

`string | valueFrom` · required

The AgentCore Identity oauth2 credential provider to obtain tokens
from.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.targets[].credentials.oauth.scopes

`[]string` · required

OAuth scopes to request (at least one).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.targets[].credentials.oauth.grantType

`string`

The grant flow: CLIENT_CREDENTIALS (machine-to-machine, the common
case) or AUTHORIZATION_CODE (user-delegated; requires
default_return_url). Omitted = AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CLIENT_CREDENTIALS","AUTHORIZATION_CODE","TOKEN_EXCHANGE"]}}

### spec.targets[].credentials.oauth.defaultReturnUrl

`string`

Where the user's browser lands after authorizing - required by AWS
when grant_type is AUTHORIZATION_CODE.

### spec.targets[].credentials.oauth.customParameters

`map<string, string>`

Extra provider-specific token-request parameters.

### spec.targets[].metadata

`AwsBedrockAgentCoreGatewayTargetMetadata`

Which caller metadata (headers, query parameters) propagates through
the gateway to the backend and back.

### spec.targets[].metadata.allowedQueryParameters

`[]string`

URL query parameters propagated from the caller to the backend.

- rule: {"repeated":{"maxItems":"10","items":{"string":{"minLen":"1"}}}}

### spec.targets[].metadata.allowedRequestHeaders

`[]string`

HTTP headers propagated from the caller to the backend.

- rule: {"repeated":{"maxItems":"10","items":{"string":{"minLen":"1"}}}}

### spec.targets[].metadata.allowedResponseHeaders

`[]string`

HTTP headers propagated from the backend response to the caller.

- rule: {"repeated":{"maxItems":"10","items":{"string":{"minLen":"1"}}}}

### spec.targets[].metadata.enabled

`bool` · optional (explicit presence)

Whether metadata propagation is on. Unset means on: declaring the block
has always meant propagating, and this switch lets a manifest say the
opposite out loud.

- default: `true`

### spec.targets[].privateEndpoint

`AwsBedrockAgentCoreGatewayPrivateEndpoint`

Reach a PRIVATE backend through your VPC instead of the public
internet.

- rule: private endpoint must set exactly one of managed_vpc or self_managed_lattice

### spec.targets[].privateEndpoint.managedVpc

`AwsBedrockAgentCoreGatewayManagedVpcEndpoint`

AWS manages VPC endpoints in your subnets.

### spec.targets[].privateEndpoint.managedVpc.vpcId

`string | valueFrom` · required

The VPC to route through.

Containment-exempt: a managed endpoint is a private PATH through the
VPC to a private provider or backend; the gateway is not deployed inside
it. On a diagram it stands where its own network places it, with a
line into the VPC -- the verdict a Lambda's VPC subnets and an
EventBridge pipe's task subnets already carry.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.targets[].privateEndpoint.managedVpc.subnetIds

`[]string | valueFrom` · required

Subnets for the managed endpoint's network interfaces (at least one).
Containment-exempt for the same reason as the VPC above.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.targets[].privateEndpoint.managedVpc.securityGroupIds

`[]string | valueFrom`

Security groups on the endpoint interfaces (max 5).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.targets[].privateEndpoint.managedVpc.endpointIpAddressType

`string`

Whether the endpoint answers IPV4 or IPV6.

- rule: {"string":{"in":["IPV4","IPV6"]}}

### spec.targets[].privateEndpoint.managedVpc.routingDomain

`string`

Domain the endpoint routes (3-255 characters). Omitted = derived from
the target.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.targets[].privateEndpoint.managedVpc.tags

`map<string, string>`

Tags applied to the AWS-managed endpoint resources (the module always
adds the Planton identity tags).

### spec.targets[].privateEndpoint.selfManagedLattice

`AwsBedrockAgentCoreGatewayLatticeEndpoint`

You bring a VPC Lattice resource configuration.

### spec.targets[].privateEndpoint.selfManagedLattice.resourceConfigurationId

`string` · required

The Lattice resource-configuration identifier.

- rule: {"string":{"minLen":"1"}}

## Validation Rules

- `custom_jwt_requires_authorizer`: custom_jwt_authorizer is required when authorizer_type is CUSTOM_JWT
- `target_names_unique`: targets entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockAgentCoreGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.gateway_id` | `string` | The unique gateway identifier (e.g. "my-gateway-abc123de45"). |
| `status.outputs.gateway_arn` | `string` | The Amazon Resource Name of the gateway - the canonical key for IAM policies and harness gateway tools. |
| `status.outputs.gateway_url` | `string` | The MCP URL agents connect to (e.g. "https://my-gateway-abc123de45.gateway.bedrock-agentcore.us-west-2.amazonaws.com/mcp"). |
| `status.outputs.workload_identity_arn` | `string` | ARN of the workload identity AWS created for this gateway. |
| `status.outputs.target_ids` | `map<string, string>` | Target IDs keyed by each `targets` entry's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.interceptors[].lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.policyEngine.policyEngineArn` | AwsBedrockAgentCoreIdentity | `status.outputs.policy_engine_arn` |
| `spec.targets[].backend.agentcoreRuntime.agentRuntimeArn` | AwsBedrockAgentCoreRuntime | `status.outputs.agent_runtime_arn` |
| `spec.targets[].backend.lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.targets[].privateEndpoint.managedVpc.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.targets[].privateEndpoint.managedVpc.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.targets[].privateEndpoint.managedVpc.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].tools[].agentcoreGateway.gatewayArn` | `status.outputs.gateway_arn` |

## See Also

- [Overview](../README.md)
