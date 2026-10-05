# AWS Bedrock AgentCore Runtime

Deploys an Amazon Bedrock AgentCore agent runtime — a serverless, session-isolated execution environment that hosts your agent code (any framework: LangGraph, CrewAI, Strands, plain Python or Node) behind AWS-managed scaling, identity, and networking. The runtime executes one immutable artifact per version — a container image from ECR or an AWS-managed code bundle from S3 — and every spec change creates a new version, with named endpoints deciding when live traffic moves. Billing is per-second for CPU and memory only while sessions execute.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Agent Runtime** — the runtime itself, named by `runtimeName`, executing the container image or code bundle with the chosen network mode, server protocol, environment variables, session lifecycle, filesystem mounts, and optional inbound JWT authorization
- **Runtime Endpoint** — one per `endpoints` entry: a named serving endpoint that floats on the latest version (version omitted) or pins an explicit one. AWS also maintains an implicit DEFAULT endpoint on every runtime
- **Resource Policy** — applied to the runtime's own ARN, created only when `resourcePolicy` is set; both engines inject the deployed runtime's ARN into every statement's Resource

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AgentCore control-plane permissions (`bedrock-agentcore:CreateAgentRuntime` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Bedrock AgentCore available in the target region
- An IAM role trusting `bedrock-agentcore.amazonaws.com` with pull/read access to your artifact — ECR pull for `artifact.container`, S3 read for `artifact.code` — wired as `roleArn`
- The artifact itself: an ECR image exposing the runtime contract (an HTTP server on the expected port), or a Python/Node bundle in S3
- Subnets and security groups (only for `network.mode: VPC`)

## Deploy

### Console

Open the deployment store, find **AWS Bedrock AgentCore Runtime**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the artifact source, and networking. Start from the **Code-Bundle Agent** preset in the [Presets](#presets) tab to run Python source straight from S3 with no image to build, or the **Container Agent in a VPC** preset for a container reaching private resources.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreRuntime
metadata:
  name: support-agent
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  runtimeName: support_agent
  description: Hosts the support agent's Python service
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: agentcore-runtime-role
      fieldPath: status.outputs.role_arn
  artifact:
    code:
      runtime: PYTHON_3_13
      entryPoint:
        - main.py
      s3:
        bucket:
          valueFrom:
            kind: AwsS3Bucket
            name: agent-code-bundles
            fieldPath: status.outputs.bucket_id
        prefix: bundles/support-agent.zip
  network:
    mode: PUBLIC
  endpoints:
    - name: live
      description: Production traffic (tracks the latest version)
```

```shell
planton apply -f agentcore-runtime.yaml
```

This creates a code-bundle runtime running `main.py` on the managed Python 3.13 base with AWS-managed outbound internet, plus a `live` endpoint floating on the latest version. An Infra Job tracks the provisioning in real time.

### InfraChart

When the runtime deploys alongside its execution role and code bucket in one chart, wire both references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  runtimeName: support_agent
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: agentcore-runtime-role
      fieldPath: status.outputs.role_arn
  artifact:
    code:
      runtime: PYTHON_3_13
      entryPoint:
        - main.py
      s3:
        bucket:
          valueFrom:
            kind: AwsS3Bucket
            name: agent-code-bundles
            fieldPath: status.outputs.bucket_id
        prefix: bundles/support-agent.zip
  network:
    mode: PUBLIC
```

The InfraPipeline resolves the dependency graph, deploys the role and bucket first, then provisions the runtime on them.

## Key Configuration

These are the most important decisions when configuring an agent runtime. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Ship changes through versions, promote through endpoints** — edit the spec freely: sessions in flight finish on their version, and live traffic moves only when a floating endpoint picks up the new version or you re-point a pinned one. This is the built-in staging/production split — a `live` endpoint pinned to a known-good version and a `staging` endpoint floating on latest, over one runtime.

**The artifact arm is replace-on-switch** — exactly one of `container` or `code`; switching between them replaces the runtime (new identity, new ARN), while changing values within an arm creates a new version in place. The code arm trades image builds for AWS's managed Python/Node base; the container arm runs any language.

**runtimeName has a strict charset** — a letter first, then letters, digits, and underscores; hyphens are rejected, which is why the name is an explicit field rather than derived from `metadata.name`. Changing it replaces the runtime.

**The role is the agent's blast radius** — it needs artifact access plus whatever AWS APIs your agent calls. Scope it per agent; never share one broad role across agents.

**PUBLIC mode still gates data-plane calls** — it controls the session's outbound network only; inbound invocation is always authenticated by IAM (SigV4) or, when `customJwtAuthorizer` is set, OIDC bearer tokens matching your audience, client, and claim rules. Use VPC mode when the agent must reach private resources.

**Session scratch is ephemeral by design** — `sessionStorage` mounts vanish when the session ends. Use EFS access-point mounts for durable cross-session state; mount paths take the `/mnt/<name>` shape and are the mount identity.

**Never author a Resource member in resource-policy statements** — AWS requires each statement's Resource to be exactly the runtime's own ARN, an AWS-suffixed value that does not exist until create. Both IaC engines inject it; authors write only Effect, Principal, Action, and conditions.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `artifact.code.s3.bucket` | `status.outputs.bucket_id` |
| **AwsSubnet** | `network.vpcConfig.subnets` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `network.vpcConfig.securityGroups` | `status.outputs.security_group_id` |
| **AwsEfsAccessPoint** | `filesystems[].efsAccessPointArn` | `status.outputs.access_point_arn` |

The JWT authorizer's private endpoint additionally takes **AwsVpc**, **AwsSubnet**, and **AwsSecurityGroup** references when the OIDC provider is reached through your VPC.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `agent_runtime_arn` | The runtime's ARN — the canonical key for IAM policies | AgentCore Gateway `agentcoreRuntime` targets; Evaluation harness runtime environments |
| `agent_runtime_id` | The unique runtime identifier | Data-plane invocation (`InvokeAgentRuntime`) and CLI operations |
| `endpoint_arns` | Endpoint ARNs keyed by each `endpoints` entry's name — an endpoint's AWS identity is its name | Callers invoking a specific endpoint qualifier |
| `agent_runtime_version` | The current version number; every spec change advances it | Pinning an endpoint to a known-good version |

`workload_identity_arn` — the identity the hosted agent presents when calling AgentCore services — is also exported; it mostly serves JWT `allowedWorkloads` restrictions on other AgentCore resources.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Code bundle, no image pipeline** — Python or Node source zipped to S3, an entrypoint, and AWS's managed base runtime. The fastest path from agent code to a hosted agent: shipping a new version is uploading a new bundle and applying. Start from the **Code-Bundle Agent** preset.

**Container agent in a VPC** — your own image with VPC networking, session lifecycle ceilings, and ephemeral scratch storage. The shape for agents that call private databases and internal APIs, or need a language the managed base does not cover. Start from the **Container Agent in a VPC** preset.

**MCP server as a runtime** — `serverProtocol: MCP` hosts an MCP server instead of a plain HTTP agent; front it with an AgentCore gateway target to give other agents one authenticated URL to its tools.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role the service assumes to pull the artifact and run the agent
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — holds code bundles for the managed-runtime artifact arm
- [**AWS Bedrock AgentCore Gateway**](/infra-catalog/aws-bedrock-agent-core-gateway) — fronts this runtime as an MCP tool via `agentcoreRuntime` targets, and supplies tools the hosted agent calls
- [**AWS Bedrock AgentCore Memory**](/infra-catalog/aws-bedrock-agent-core-memory) — the memory hosted agents write events to and query records from
- [**AWS Bedrock AgentCore Evaluation**](/infra-catalog/aws-bedrock-agent-core-evaluation) — harnesses pin their runtime environment to this runtime for repeatable benches
- [**AWS Subnet**](/infra-catalog/aws-subnet) — session network placement in VPC mode
- [**AWS Security Group**](/infra-catalog/aws-security-group) — session network rules in VPC mode
- [**AWS EFS Access Point**](/infra-catalog/aws-efs-access-point) — durable cross-session filesystem mounts
