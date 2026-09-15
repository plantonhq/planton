# AwsBedrockAgentCoreRuntime

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsBedrockAgentCoreRuntimeSpec defines the desired configuration for an
Amazon Bedrock AgentCore agent runtime - a serverless, session-isolated
execution environment that hosts YOUR agent code (any framework:
LangGraph, CrewAI, Strands, plain Python/Node) behind AWS-managed
scaling, identity, and networking.

The runtime hosts one immutable artifact per version: either a container
image (any language, ECR URI) or an AWS-managed code bundle (Python/Node
entrypoint in S3). Every spec change creates a new runtime VERSION;
endpoints pin or float across versions, so live traffic moves only when
an endpoint says so. Creating a runtime is free - AWS bills per-second
for CPU/memory only while sessions execute.

## Example

```yaml
# Canonical AwsBedrockAgentCoreRuntime example (hack/dev manifest and
# refgen Example source): a code-bundle runtime exercising every arm --
# VPC networking, JWT authorization with custom claims and a private
# endpoint, filesystems, named endpoints, and the resource policy.
# Literal ARNs/ids stand in for composed references so the offline
# `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreRuntime
metadata:
  name: support-agent-runtime
  id: support-agent-runtime
  org: test-org
  env: dev
spec:
  region: us-west-2
  runtimeName: support_agent
  description: Hosts the support agent's LangGraph service
  roleArn:
    value: arn:aws:iam::123456789012:role/agentcore-runtime-role
  artifact:
    code:
      runtime: PYTHON_3_13
      entryPoint:
        - main.py
      s3:
        bucket:
          value: my-agent-code-bucket
        prefix: bundles/support-agent.zip
  network:
    mode: VPC
    vpcConfig:
      subnets:
        - value: subnet-0123456789abcdef0
      securityGroups:
        - value: sg-0123456789abcdef0
  serverProtocol: HTTP
  environmentVariables:
    LOG_LEVEL: info
  lifecycle:
    idleRuntimeSessionTimeoutSeconds: 900
    maxLifetimeSeconds: 3600
  customJwtAuthorizer:
    discoveryUrl: https://accounts.google.com/.well-known/openid-configuration
    allowedAudience:
      - support-agents
    allowedClients:
      - support-client
    customClaims:
      - claimName: org
        valueType: STRING
        matchOperator: EQUALS
        matchValue: acme
    privateEndpoint:
      managedVpc:
        vpcId:
          value: vpc-0123456789abcdef0
        subnetIds:
          - value: subnet-0123456789abcdef0
        securityGroupIds:
          - value: sg-0123456789abcdef0
        endpointIpAddressType: IPV4
    privateEndpointOverrides:
      - domain: idp.internal.example.com
        privateEndpoint:
          selfManagedLattice:
            resourceConfigurationId: rcfg-0123456789abcdef0
  requestHeaderAllowlist:
    - X-Trace-Id
  filesystems:
    - mountPath: /mnt/scratch
      sessionStorage: true
    - mountPath: /mnt/shared
      efsAccessPointArn:
        value: arn:aws:elasticfilesystem:us-west-2:123456789012:access-point/fsap-0123456789abcdef0
  endpoints:
    - name: live
      description: Production traffic (tracks the latest version)
    - name: pinned
      agentRuntimeVersion: "1"
  # Statements carry no Resource member: AWS requires it to be exactly
  # the runtime's own ARN (unknowable pre-create), so the modules inject
  # it into every statement.
  resourcePolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          AWS: arn:aws:iam::210987654321:root
        Action: bedrock-agentcore:InvokeAgentRuntime
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.runtimeName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.artifact` | `AwsBedrockAgentCoreRuntimeArtifact` | yes |  |  |
| `spec.artifact.container` | `AwsBedrockAgentCoreRuntimeContainer` |  |  |  |
| `spec.artifact.container.imageUri` | `string` | yes |  |  |
| `spec.artifact.code` | `AwsBedrockAgentCoreRuntimeCode` |  |  |  |
| `spec.artifact.code.runtime` | `string` |  |  |  |
| `spec.artifact.code.entryPoint` | `[]string` | yes |  |  |
| `spec.artifact.code.s3` | `AwsBedrockAgentCoreRuntimeCodeS3` | yes |  |  |
| `spec.artifact.code.s3.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.artifact.code.s3.prefix` | `string` | yes |  |  |
| `spec.artifact.code.s3.versionId` | `string` |  |  |  |
| `spec.network` | `AwsBedrockAgentCoreRuntimeNetwork` | yes |  |  |
| `spec.network.mode` | `string` |  |  |  |
| `spec.network.vpcConfig` | `AwsBedrockAgentCoreVpcConfig` |  |  |  |
| `spec.network.vpcConfig.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.network.vpcConfig.securityGroups` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.serverProtocol` | `string` |  |  |  |
| `spec.environmentVariables` | `map<string, string>` |  |  |  |
| `spec.lifecycle` | `AwsBedrockAgentCoreRuntimeLifecycle` |  |  |  |
| `spec.lifecycle.idleRuntimeSessionTimeoutSeconds` | `int32` |  |  |  |
| `spec.lifecycle.maxLifetimeSeconds` | `int32` |  |  |  |
| `spec.customJwtAuthorizer` | `AwsBedrockAgentCoreJwtAuthorizer` |  |  |  |
| `spec.customJwtAuthorizer.discoveryUrl` | `string` | yes |  |  |
| `spec.customJwtAuthorizer.allowedAudience` | `[]string` |  |  |  |
| `spec.customJwtAuthorizer.allowedClients` | `[]string` |  |  |  |
| `spec.customJwtAuthorizer.allowedScopes` | `[]string` |  |  |  |
| `spec.customJwtAuthorizer.allowedWorkloads` | `AwsBedrockAgentCoreAllowedWorkloads` |  |  |  |
| `spec.customJwtAuthorizer.allowedWorkloads.workloadIdentities` | `[]string` | yes |  |  |
| `spec.customJwtAuthorizer.allowedWorkloads.hostingEnvironmentArns` | `[]string` | yes |  |  |
| `spec.customJwtAuthorizer.customClaims` | `[]AwsBedrockAgentCoreCustomClaim` |  |  |  |
| `spec.customJwtAuthorizer.customClaims[].claimName` | `string` | yes |  |  |
| `spec.customJwtAuthorizer.customClaims[].valueType` | `string` |  |  |  |
| `spec.customJwtAuthorizer.customClaims[].matchOperator` | `string` |  |  |  |
| `spec.customJwtAuthorizer.customClaims[].matchValue` | `string` |  |  |  |
| `spec.customJwtAuthorizer.customClaims[].matchValues` | `[]string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint` | `AwsBedrockAgentCorePrivateEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc` | `AwsBedrockAgentCoreManagedVpcEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.endpointIpAddressType` | `string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.routingDomain` | `string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.tags` | `map<string, string>` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.selfManagedLattice` | `AwsBedrockAgentCoreLatticeEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpoint.selfManagedLattice.resourceConfigurationId` | `string` | yes |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides` | `[]AwsBedrockAgentCorePrivateEndpointOverride` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].domain` | `string` | yes |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint` | `AwsBedrockAgentCorePrivateEndpoint` | yes |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc` | `AwsBedrockAgentCoreManagedVpcEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.endpointIpAddressType` | `string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.routingDomain` | `string` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.tags` | `map<string, string>` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice` | `AwsBedrockAgentCoreLatticeEndpoint` |  |  |  |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice.resourceConfigurationId` | `string` | yes |  |  |
| `spec.requestHeaderAllowlist` | `[]string` |  |  |  |
| `spec.filesystems` | `[]AwsBedrockAgentCoreRuntimeFilesystem` |  |  |  |
| `spec.filesystems[].mountPath` | `string` | yes |  |  |
| `spec.filesystems[].efsAccessPointArn` | `string \| valueFrom` |  |  | AwsEfsAccessPoint (`status.outputs.access_point_arn`) |
| `spec.filesystems[].s3FilesAccessPointArn` | `string \| valueFrom` |  |  |  |
| `spec.filesystems[].sessionStorage` | `bool` |  |  |  |
| `spec.endpoints` | `[]AwsBedrockAgentCoreRuntimeEndpoint` |  |  |  |
| `spec.endpoints[].name` | `string` | yes |  |  |
| `spec.endpoints[].description` | `string` |  |  |  |
| `spec.endpoints[].agentRuntimeVersion` | `string` |  |  |  |
| `spec.resourcePolicy` | `object` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the agent runtime will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.runtimeName

`string` · required

Runtime name in AWS (1-48 characters; must start with a letter, then
letters, digits, underscore - AWS rejects hyphens here, so the name is
an explicit field rather than metadata.name). Changing it replaces the
runtime.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.description

`string`

Human-readable description shown in the AgentCore console (1-4096
characters when set). Updates in place.

- rule: {"string":{"maxLen":"4096"}}

### spec.roleArn

`string | valueFrom` · required

IAM role the AgentCore service assumes to run the agent (pull the ECR
image or read the S3 code bundle, write logs, call the AWS APIs your
agent uses). The role must trust bedrock-agentcore.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.artifact

`AwsBedrockAgentCoreRuntimeArtifact` · required

What the runtime executes - exactly one of a container image or an
AWS-managed code bundle. Switching arms replaces the runtime; changing
values within an arm creates a new runtime version in place.

- rule: {"required":true}
- rule: artifact must set exactly one of container or code

### spec.artifact.container

`AwsBedrockAgentCoreRuntimeContainer`

Run a container image. Any language/framework; the image must expose
the runtime contract (an HTTP server on the expected port).

### spec.artifact.container.imageUri

`string` · required

ECR image URI the runtime pulls
(e.g. "123456789012.dkr.ecr.us-west-2.amazonaws.com/my-agent:v1").

- rule: {"string":{"minLen":"1"}}

### spec.artifact.code

`AwsBedrockAgentCoreRuntimeCode`

Run an AWS-managed code bundle - hand AWS your Python/Node source in
S3 and an entrypoint; AWS supplies the base environment (no image to
build or maintain).

### spec.artifact.code.runtime

`string`

Managed base runtime for the code bundle.

- rule: {"string":{"in":["PYTHON_3_10","PYTHON_3_11","PYTHON_3_12","PYTHON_3_13","PYTHON_3_14","NODE_22"]}}

### spec.artifact.code.entryPoint

`[]string` · required

Entrypoint command relative to the bundle root (1-2 elements, e.g.
["main.py"] or ["src/app.py", "handler"]).

- rule: {"repeated":{"minItems":"1","maxItems":"2","items":{"string":{"minLen":"1","maxLen":"128"}}}}

### spec.artifact.code.s3

`AwsBedrockAgentCoreRuntimeCodeS3` · required

S3 location of the code bundle.

- rule: {"required":true}

### spec.artifact.code.s3.bucket

`string | valueFrom` · required

Bucket holding the code bundle.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.artifact.code.s3.prefix

`string` · required

Object key (or key prefix) of the code within the bucket (1-1024
characters).

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.artifact.code.s3.versionId

`string`

Pin an explicit object version on a versioned bucket. Omitted = the
current version.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024"}}

### spec.network

`AwsBedrockAgentCoreRuntimeNetwork` · required

How the runtime reaches the network. Required by AWS.

- rule: {"required":true}
- rule: vpc_config is required when mode is VPC and forbidden otherwise

### spec.network.mode

`string`

PUBLIC gives sessions AWS-managed outbound internet; VPC attaches
sessions to your subnets so they reach private resources (and the
internet only through your VPC's routing).

- rule: {"string":{"in":["PUBLIC","VPC"]}}

### spec.network.vpcConfig

`AwsBedrockAgentCoreVpcConfig`

VPC placement - required when mode is VPC.

### spec.network.vpcConfig.subnets

`[]string | valueFrom` · required

Subnets the session network interfaces attach to (at least one).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.network.vpcConfig.securityGroups

`[]string | valueFrom` · required

Security groups applied to the session network interfaces (at least
one).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.serverProtocol

`string`

Protocol the hosted agent server speaks: HTTP (plain request/response,
the AWS default when omitted), MCP (Model Context Protocol server),
A2A (agent-to-agent), or AGUI.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["HTTP","MCP","A2A","AGUI"]}}

### spec.environmentVariables

`map<string, string>`

Environment variables injected into every session of the runtime.

### spec.lifecycle

`AwsBedrockAgentCoreRuntimeLifecycle`

Session lifetime tuning. Omitted = AWS defaults.

### spec.lifecycle.idleRuntimeSessionTimeoutSeconds

`int32`

Seconds an idle session survives before AWS reclaims it. Omitted =
AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.lifecycle.maxLifetimeSeconds

`int32`

Hard ceiling in seconds on any session's total lifetime, idle or not.
Omitted = AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.customJwtAuthorizer

`AwsBedrockAgentCoreJwtAuthorizer`

Inbound JWT authorization for the runtime's endpoints. Omitted = the
runtime accepts AWS IAM (SigV4) callers only.

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

`AwsBedrockAgentCoreAllowedWorkloads`

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

`[]AwsBedrockAgentCoreCustomClaim`

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

`AwsBedrockAgentCorePrivateEndpoint`

Reach a PRIVATE OIDC provider through your VPC instead of the public
internet.

- rule: private endpoint must set exactly one of managed_vpc or self_managed_lattice

### spec.customJwtAuthorizer.privateEndpoint.managedVpc

`AwsBedrockAgentCoreManagedVpcEndpoint`

AWS manages VPC endpoints in your subnets.

### spec.customJwtAuthorizer.privateEndpoint.managedVpc.vpcId

`string | valueFrom` · required

The VPC to route through.

Containment-exempt: a managed endpoint is a private PATH through the
VPC to a private provider or backend; the runtime is not deployed inside
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

`AwsBedrockAgentCoreLatticeEndpoint`

You bring a VPC Lattice resource configuration.

### spec.customJwtAuthorizer.privateEndpoint.selfManagedLattice.resourceConfigurationId

`string` · required

The Lattice resource-configuration identifier.

- rule: {"string":{"minLen":"1"}}

### spec.customJwtAuthorizer.privateEndpointOverrides

`[]AwsBedrockAgentCorePrivateEndpointOverride`

Per-domain overrides of the private endpoint (max 5) - route specific
issuer domains through different private paths.

- rule: {"repeated":{"maxItems":"5"}}

### spec.customJwtAuthorizer.privateEndpointOverrides[].domain

`string` · required

The domain this override captures (1-253 characters).

- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint

`AwsBedrockAgentCorePrivateEndpoint` · required

The private path for that domain.

- rule: {"required":true}
- rule: private endpoint must set exactly one of managed_vpc or self_managed_lattice

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc

`AwsBedrockAgentCoreManagedVpcEndpoint`

AWS manages VPC endpoints in your subnets.

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId

`string | valueFrom` · required

The VPC to route through.

Containment-exempt: a managed endpoint is a private PATH through the
VPC to a private provider or backend; the runtime is not deployed inside
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

`AwsBedrockAgentCoreLatticeEndpoint`

You bring a VPC Lattice resource configuration.

### spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice.resourceConfigurationId

`string` · required

The Lattice resource-configuration identifier.

- rule: {"string":{"minLen":"1"}}

### spec.requestHeaderAllowlist

`[]string`

HTTP request headers the runtime forwards to your agent code beyond
the AWS defaults (e.g. a tracing header your framework reads).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.filesystems

`[]AwsBedrockAgentCoreRuntimeFilesystem`

Filesystems mounted into every session (max 5). Each mount is exactly
one of an EFS access point, an S3 files access point, or AWS-managed
per-session scratch storage.

- rule: {"repeated":{"maxItems":"5"}}
- rule: filesystem must set exactly one of efs_access_point_arn, s3_files_access_point_arn, or session_storage

### spec.filesystems[].mountPath

`string` · required

Where the filesystem appears inside sessions: /mnt/<one-level>
(6-200 characters).

- rule: {"string":{"minLen":"6","maxLen":"200","pattern":"^/mnt/[a-zA-Z0-9._-]+/?$"}}

### spec.filesystems[].efsAccessPointArn

`string | valueFrom`

Mount an EFS access point (durable, shared across sessions).
Reference an AwsEfsAccessPoint access_point_arn output or pass a
literal ARN.

- references: AwsEfsAccessPoint (`status.outputs.access_point_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEfsAccessPoint, name: <that resource's name>, fieldPath: status.outputs.access_point_arn}} -- a bare string does not parse

### spec.filesystems[].s3FilesAccessPointArn

`string | valueFrom`

Mount an S3 files access point.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.filesystems[].sessionStorage

`bool`

AWS-managed per-session scratch storage (wiped when the session
ends).

### spec.endpoints

`[]AwsBedrockAgentCoreRuntimeEndpoint`

Named serving endpoints for this runtime. Each endpoint either floats
on the latest version (omit version) or pins an explicit one - this is
how you run DEFAULT/staging/prod traffic splits over one runtime.
AWS also maintains an implicit DEFAULT endpoint on every runtime.

### spec.endpoints[].name

`string` · required

Endpoint name (1-48 characters; letter first, then letters, digits,
underscore). The for_each key on both engines and the key in the
`endpoint_arns` output map; also the endpoint name in AWS.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.endpoints[].description

`string`

Human-readable description (1-256 characters when set).

- rule: {"string":{"maxLen":"256"}}

### spec.endpoints[].agentRuntimeVersion

`string`

Pin the endpoint to an explicit runtime version ("1", "2", ...).
Omitted = the endpoint tracks the runtime's latest version.

### spec.resourcePolicy

`object`

Resource-based policy applied to the runtime's own ARN (an IAM policy
document as structured JSON) - grant other accounts or services
permission to invoke this runtime. Omitted = no resource policy.
Author statements WITHOUT a Resource member: AWS requires each
statement's Resource to be exactly the runtime's own ARN (an
AWS-suffixed value no author can know before create; PutResourcePolicy
400 otherwise, live-caught 2026-08-14), so both modules set Resource
on every statement to the runtime ARN they deploy.

## Validation Rules

- `endpoint_names_unique`: endpoints entries must have unique names
- `filesystem_mount_paths_unique`: filesystems entries must have unique mount_paths

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockAgentCoreRuntime, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.agent_runtime_id` | `string` | The unique runtime identifier (e.g. "my_agent-AbC1dEf2Gh"). |
| `status.outputs.agent_runtime_arn` | `string` | The Amazon Resource Name of the runtime - the canonical key for IAM policies and gateway HTTP targets. |
| `status.outputs.agent_runtime_version` | `string` | The runtime's current version number ("1", "2", ...). Every spec change that touches the artifact or configuration advances it. |
| `status.outputs.workload_identity_arn` | `string` | ARN of the workload identity AWS created for this runtime (the identity the hosted agent presents when calling AgentCore services). |
| `status.outputs.endpoint_arns` | `map<string, string>` | Endpoint ARNs keyed by each `endpoints` entry's name. An endpoint's AWS identity IS its name (there is no separate endpoint ID). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.artifact.code.s3.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.network.vpcConfig.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.network.vpcConfig.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.filesystems[].efsAccessPointArn` | AwsEfsAccessPoint | `status.outputs.access_point_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].runtimeEnvironment.agentRuntimeArn` | `status.outputs.agent_runtime_arn` |
| AwsBedrockAgentCoreGateway | `spec.targets[].backend.agentcoreRuntime.agentRuntimeArn` | `status.outputs.agent_runtime_arn` |

## See Also

- [Overview](../README.md)
