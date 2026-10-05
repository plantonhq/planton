# AWS Bedrock AgentCore Tools

Deploys a bundle of Amazon Bedrock AgentCore built-in tools — managed, sandboxed execution environments agents drive at runtime: cloud browsers with optional S3 session recording and Chrome enterprise policies, reusable browser profiles holding saved cookies and logins, and code interpreters that run model-written code. Each arm is optional (the bundle needs at least one), and AWS exposes no update for any of the three — every field change recreates the tool. Tools are configuration shells; AWS bills per session at runtime.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions, per named entry:

- **Browser** — one per `browsers` entry: a managed cloud browser with PUBLIC or VPC egress, optional traffic signing, S3 session recording, Chrome enterprise policy files, and mTLS client certificates from Secrets Manager
- **Browser Profile** — one per `browserProfiles` entry: reusable saved browser state (cookies, logins) sessions can start from
- **Code Interpreter** — one per `codeInterpreters` entry: a sandbox for model-written code with SANDBOX (no network), PUBLIC, or VPC networking and an optional execution role for the AWS APIs the code calls

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AgentCore tool permissions (`bedrock-agentcore:CreateBrowser`, `CreateBrowserProfile`, `CreateCodeInterpreter` and their siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Bedrock AgentCore available in the target region
- An IAM role trusting `bedrock-agentcore.amazonaws.com` with matching S3 or Secrets Manager access, wired as `executionRoleArn` (only for browsers with `recording`, `enterprisePolicies`, or `certificates`, and for interpreters whose code calls AWS APIs)
- An S3 bucket the role may write to (only for browser session recording) and the policy JSON objects (only for enterprise policies)
- Subnets and security groups (only for VPC network mode)

## Deploy

### Console

Open the deployment store, find **AWS Bedrock AgentCore Tools**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the three tool arms. Start from the **Sandboxed Code Interpreter** preset in the [Presets](#presets) tab for the safest compute shape, or the **Recorded Research Browser** preset for a browser with a full audit trail.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreTools
metadata:
  name: agent-sandbox
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  codeInterpreters:
    - name: python_sandbox
      description: Runs model-written analysis code with no network access
      network:
        mode: SANDBOX
```

```shell
planton apply -f agentcore-tools.yaml
```

This creates one code interpreter whose sandbox has no network access and no AWS credentials — model-written code can compute but cannot reach anything. An Infra Job tracks the provisioning in real time.

### InfraChart

When a recorded browser deploys alongside its recordings bucket in one chart, wire the bucket reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  browsers:
    - name: research_browser
      executionRoleArn:
        valueFrom:
          kind: AwsIamRole
          name: agent-tools-role
          fieldPath: status.outputs.role_arn
      network:
        mode: PUBLIC
      recording:
        enabled: true
        s3Location:
          bucket:
            valueFrom:
              kind: AwsS3Bucket
              name: agent-recordings
              fieldPath: status.outputs.bucket_id
          prefix: browser-sessions/
```

The InfraPipeline resolves the dependency graph, deploys the role and bucket first, then creates the browser recording into them.

## Key Configuration

These are the most important decisions when configuring a tools bundle. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Everything is replace-on-change** — AWS exposes no update for browsers, profiles, or interpreters: every field change recreates the tool. Recreates are cheap (tools are configuration shells), but sessions are not — in-flight sessions on the old tool finish while new sessions land on the replacement, so plan rollouts so long-running sessions drain before you drop the old tool's grants.

**Default code to SANDBOX** — model-written code gets network access only when the task genuinely needs it, and then prefer VPC mode so egress rides your routing and security groups. An interpreter with no `executionRoleArn` also has no AWS credentials — the safest posture. Note the modes differ per tool deliberately: SANDBOX exists only for interpreters; browsers take PUBLIC or VPC.

**Recording is your audit trail** — for browsers touching authenticated sites, enable `recording` to S3 and lifecycle the prefix. Replay is how you answer "what did the agent actually do?" — and `signingEnabled` lets sites verify traffic comes from an AWS-managed browser.

**Profiles hold credentials — treat them like credentials** — a saved login in a browser profile is a live session for whoever starts a browser from it. Scope profile use narrowly.

**Enterprise policies are the browser kill switch** — Chrome policy files from S3 control downloads, extensions, and URL allow-lists. MANAGED enforces; RECOMMENDED merely suggests defaults the session can change. For guardrails, that distinction is the whole decision.

**Names take the AgentCore charset** — a letter first, then letters, digits, and underscores; CreateBrowser rejects hyphens server-side. Each name is the for_each key on both engines and the key in the corresponding output map, so renaming an entry replaces that tool.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `browsers[].executionRoleArn`, `codeInterpreters[].executionRoleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `browsers[].recording.s3Location.bucket`, `browsers[].enterprisePolicies[].s3.bucket` | `status.outputs.bucket_id` |
| **AwsSecretsManagerSecret** | `browsers[].certificates[].secretArn`, `codeInterpreters[].certificates[].secretArn` | `status.outputs.secret_arn` |
| **AwsSubnet** | `network.vpcConfig.subnets` (browsers and interpreters) | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `network.vpcConfig.securityGroups` (browsers and interpreters) | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `browser_arns` | Browser ARNs keyed by each `browsers` entry's name | AgentCore Evaluation harness `agentcoreBrowser` tools (reference the map key explicitly) |
| `code_interpreter_arns` | Interpreter ARNs keyed by each `codeInterpreters` entry's name | AgentCore Evaluation harness `agentcoreCodeInterpreter` tools |
| `browser_ids` / `code_interpreter_ids` | The tools' unique identifiers, keyed the same way | Agent code starting browser and interpreter sessions through the data plane |
| `browser_profile_ids` / `browser_profile_arns` | Profile identifiers keyed by each `browserProfiles` entry's name | Starting browser sessions from saved state |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Sandboxed compute** — one code interpreter in SANDBOX mode with no execution role: agents run analysis and transformation code with zero blast radius. The default answer for "the agent needs to run code". Start from the **Sandboxed Code Interpreter** preset.

**Audited research browser** — a PUBLIC browser with signing enabled, session recording to a lifecycle-managed S3 prefix, and a saved profile for portals the agent must stay logged into. The shape for agents doing real web work where you need replay. Start from the **Recorded Research Browser** preset.

**Governed enterprise browser** — VPC egress plus MANAGED Chrome policies (URL allow-lists, download bans) and mTLS certificates for internal sites. Tool traffic rides your network controls, and the policy files are the enforcement point.

## Works With

- [**AWS Bedrock AgentCore Evaluation**](/infra-catalog/aws-bedrock-agent-core-evaluation) — harnesses drive these browsers and interpreters as tools via the ARN output maps
- [**AWS Bedrock AgentCore Runtime**](/infra-catalog/aws-bedrock-agent-core-runtime) — hosted agents start tool sessions against this bundle at runtime
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role for recordings, policies, certificates, and code-called AWS APIs
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — session recordings and enterprise policy files
- [**AWS Secrets Manager Secret**](/infra-catalog/aws-secrets-manager-secret) — mTLS client certificates the tools present
- [**AWS Subnet**](/infra-catalog/aws-subnet) — session network placement in VPC mode
- [**AWS Security Group**](/infra-catalog/aws-security-group) — session network rules in VPC mode
