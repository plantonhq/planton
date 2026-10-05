# AwsBedrockAgentCoreTools

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockAgentCoreToolsSpec defines the desired configuration for a
bundle of Amazon Bedrock AgentCore built-in tools - managed, sandboxed
execution environments agents drive at runtime:

  - `browsers`: cloud web browsers (page navigation, form filling,
    scraping) with optional session recording to S3;
  - `browser_profiles`: reusable saved browser state (cookies, logins)
    sessions can start from;
  - `code_interpreters`: sandboxes that run model-written code.

Every arm is optional; author the ones this bundle owns (at least
one). AWS exposes NO update for these resources - every field change
recreates the tool (cheap: tools are configuration shells; AWS bills
only per session at runtime).

## Example

```yaml
# Canonical AwsBedrockAgentCoreTools example (hack/dev manifest and
# refgen Example source): the full tools bundle -- a recorded, signed
# browser with enterprise policies and an mTLS certificate, a VPC
# browser, a saved browser profile, and code interpreters in SANDBOX and
# VPC postures. Literal ARNs/ids stand in for composed references so the
# offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreTools
metadata:
  name: support-agent-tools
  id: support-agent-tools
  org: test-org
  env: dev
spec:
  region: us-west-2
  browsers:
    - name: research_browser
      description: Web research sessions with recording
      executionRoleArn:
        value: arn:aws:iam::123456789012:role/agentcore-tools-role
      network:
        mode: PUBLIC
      signingEnabled: true
      recording:
        enabled: true
        s3Location:
          bucket:
            value: my-recordings-bucket
          prefix: browser-sessions/
      enterprisePolicies:
        - type: MANAGED
          s3:
            bucket:
              value: my-policies-bucket
            prefix: chrome/policy.json
      certificates:
        - secretArn:
            value: arn:aws:secretsmanager:us-west-2:123456789012:secret:mtls-client-cert-AbCdEf
    - name: internal_browser
      description: Reaches intranet sites through the VPC
      network:
        mode: VPC
        vpcConfig:
          subnets:
            - value: subnet-0123456789abcdef0
          securityGroups:
            - value: sg-0123456789abcdef0
  browserProfiles:
    - name: logged_in_docs
      description: Saved session state with the docs-site login
  codeInterpreters:
    - name: python_sandbox
      description: Runs model-written analysis code with no network
      network:
        mode: SANDBOX
    - name: vpc_interpreter
      description: Reaches private data stores from executed code
      executionRoleArn:
        value: arn:aws:iam::123456789012:role/agentcore-tools-role
      network:
        mode: VPC
        vpcConfig:
          subnets:
            - value: subnet-0123456789abcdef0
          securityGroups:
            - value: sg-0123456789abcdef0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.browsers` | `[]AwsBedrockAgentCoreBrowser` |  |  |  |
| `spec.browsers[].name` | `string` | yes |  |  |
| `spec.browsers[].description` | `string` |  |  |  |
| `spec.browsers[].executionRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.browsers[].network` | `AwsBedrockAgentCoreBrowserNetwork` | yes |  |  |
| `spec.browsers[].network.mode` | `string` |  |  |  |
| `spec.browsers[].network.vpcConfig` | `AwsBedrockAgentCoreToolVpcConfig` |  |  |  |
| `spec.browsers[].network.vpcConfig.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.browsers[].network.vpcConfig.securityGroups` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.browsers[].signingEnabled` | `bool` |  |  |  |
| `spec.browsers[].recording` | `AwsBedrockAgentCoreBrowserRecording` |  |  |  |
| `spec.browsers[].recording.enabled` | `bool` |  |  |  |
| `spec.browsers[].recording.s3Location` | `AwsBedrockAgentCoreToolS3Location` |  |  |  |
| `spec.browsers[].recording.s3Location.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.browsers[].recording.s3Location.prefix` | `string` | yes |  |  |
| `spec.browsers[].enterprisePolicies` | `[]AwsBedrockAgentCoreBrowserEnterprisePolicy` |  |  |  |
| `spec.browsers[].enterprisePolicies[].type` | `string` |  |  |  |
| `spec.browsers[].enterprisePolicies[].s3` | `AwsBedrockAgentCoreToolS3Object` | yes |  |  |
| `spec.browsers[].enterprisePolicies[].s3.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.browsers[].enterprisePolicies[].s3.prefix` | `string` | yes |  |  |
| `spec.browsers[].enterprisePolicies[].s3.versionId` | `string` |  |  |  |
| `spec.browsers[].certificates` | `[]AwsBedrockAgentCoreToolCertificate` |  |  |  |
| `spec.browsers[].certificates[].secretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.browserProfiles` | `[]AwsBedrockAgentCoreBrowserProfile` |  |  |  |
| `spec.browserProfiles[].name` | `string` | yes |  |  |
| `spec.browserProfiles[].description` | `string` |  |  |  |
| `spec.codeInterpreters` | `[]AwsBedrockAgentCoreCodeInterpreter` |  |  |  |
| `spec.codeInterpreters[].name` | `string` | yes |  |  |
| `spec.codeInterpreters[].description` | `string` |  |  |  |
| `spec.codeInterpreters[].executionRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.codeInterpreters[].network` | `AwsBedrockAgentCoreCodeInterpreterNetwork` | yes |  |  |
| `spec.codeInterpreters[].network.mode` | `string` |  |  |  |
| `spec.codeInterpreters[].network.vpcConfig` | `AwsBedrockAgentCoreToolVpcConfig` |  |  |  |
| `spec.codeInterpreters[].network.vpcConfig.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.codeInterpreters[].network.vpcConfig.securityGroups` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.codeInterpreters[].certificates` | `[]AwsBedrockAgentCoreToolCertificate` |  |  |  |
| `spec.codeInterpreters[].certificates[].secretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region where the tools will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.browsers

`[]AwsBedrockAgentCoreBrowser`

Managed cloud browsers.

### spec.browsers[].name

`string` · required

Browser name in AWS (1-48 characters; letter first, then letters,
digits, underscore - CreateBrowser rejects hyphens server-side,
live-caught 2026-08-14). The for_each key on both engines and the
key in the `browser_ids` output map.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.browsers[].description

`string`

Human-readable description.

### spec.browsers[].executionRoleArn

`string | valueFrom`

IAM role the browser assumes for its AWS touchpoints (writing
session recordings to S3, reading enterprise policies and
certificates). Must trust bedrock-agentcore.amazonaws.com. Required
in practice when `recording`, `enterprise_policies`, or
`certificates` are set.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.browsers[].network

`AwsBedrockAgentCoreBrowserNetwork` · required

How browser sessions reach the network. Required by AWS.

- rule: {"required":true}
- rule: vpc_config is required when mode is VPC and forbidden otherwise

### spec.browsers[].network.mode

`string`

PUBLIC gives sessions AWS-managed outbound internet; VPC attaches
sessions to your subnets.

- rule: {"string":{"in":["PUBLIC","VPC"]}}

### spec.browsers[].network.vpcConfig

`AwsBedrockAgentCoreToolVpcConfig`

VPC placement - required when mode is VPC.

### spec.browsers[].network.vpcConfig.subnets

`[]string | valueFrom` · required

Subnets the session network interfaces attach to (at least one).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.browsers[].network.vpcConfig.securityGroups

`[]string | valueFrom` · required

Security groups applied to the session network interfaces (at least
one).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.browsers[].signingEnabled

`bool` · optional (explicit presence)

Cryptographically sign browser traffic so sites can verify it comes
from an AWS-managed browser. Omitted = AWS default.

### spec.browsers[].recording

`AwsBedrockAgentCoreBrowserRecording`

Record browser sessions to S3 for replay and audit.

### spec.browsers[].recording.enabled

`bool` · optional (explicit presence)

Whether recording is on. Omitted = AWS default.

### spec.browsers[].recording.s3Location

`AwsBedrockAgentCoreToolS3Location`

Where recordings land.

### spec.browsers[].recording.s3Location.bucket

`string | valueFrom` · required

The destination bucket. The tool's execution role must be allowed to
write to it.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.browsers[].recording.s3Location.prefix

`string` · required

Key prefix within the bucket (e.g. "browser-recordings/").

- rule: {"string":{"minLen":"1"}}

### spec.browsers[].enterprisePolicies

`[]AwsBedrockAgentCoreBrowserEnterprisePolicy`

Chrome enterprise policy files applied to every session (max 100).

- rule: {"repeated":{"maxItems":"100"}}

### spec.browsers[].enterprisePolicies[].type

`string`

Policy class: MANAGED (enforced) or RECOMMENDED (defaults the user
can change). Omitted = AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["MANAGED","RECOMMENDED"]}}

### spec.browsers[].enterprisePolicies[].s3

`AwsBedrockAgentCoreToolS3Object` · required

S3 object holding the policy JSON.

- rule: {"required":true}

### spec.browsers[].enterprisePolicies[].s3.bucket

`string | valueFrom` · required

The bucket holding the object.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.browsers[].enterprisePolicies[].s3.prefix

`string` · required

Object key (1-1024 characters).

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.browsers[].enterprisePolicies[].s3.versionId

`string`

Pin an explicit object version on a versioned bucket. Omitted = the
current version.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024"}}

### spec.browsers[].certificates

`[]AwsBedrockAgentCoreToolCertificate`

Client certificates (in Secrets Manager) the browser presents to
mTLS-protected sites (max 200).

- rule: {"repeated":{"maxItems":"200"}}

### spec.browsers[].certificates[].secretArn

`string | valueFrom` · required

The Secrets Manager secret holding the certificate material.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.browserProfiles

`[]AwsBedrockAgentCoreBrowserProfile`

Reusable saved browser state (cookies, logins).

### spec.browserProfiles[].name

`string` · required

Profile name in AWS (1-48 characters; letter first, then letters,
digits, underscore). The for_each key on both engines and the key in
the `browser_profile_ids` output map.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.browserProfiles[].description

`string`

Human-readable description (1-4096 characters when set).

- rule: {"string":{"maxLen":"4096"}}

### spec.codeInterpreters

`[]AwsBedrockAgentCoreCodeInterpreter`

Managed code-execution sandboxes.

### spec.codeInterpreters[].name

`string` · required

Interpreter name in AWS (1-48 characters; letter first, then
letters, digits, underscore). The for_each key on both engines and
the key in the `code_interpreter_ids` output map.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.codeInterpreters[].description

`string`

Human-readable description (1-4096 characters when set).

- rule: {"string":{"maxLen":"4096"}}

### spec.codeInterpreters[].executionRoleArn

`string | valueFrom`

IAM role the interpreter assumes for the AWS APIs the executed code
calls. Must trust bedrock-agentcore.amazonaws.com. Omitted = the
sandbox has no AWS credentials (the safest posture).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.codeInterpreters[].network

`AwsBedrockAgentCoreCodeInterpreterNetwork` · required

How the sandbox reaches the network. Required by AWS.

- rule: {"required":true}
- rule: vpc_config is required when mode is VPC and forbidden otherwise

### spec.codeInterpreters[].network.mode

`string`

SANDBOX blocks all network access (the safest for untrusted code);
PUBLIC gives AWS-managed outbound internet; VPC attaches the sandbox
to your subnets.

- rule: {"string":{"in":["SANDBOX","PUBLIC","VPC"]}}

### spec.codeInterpreters[].network.vpcConfig

`AwsBedrockAgentCoreToolVpcConfig`

VPC placement - required when mode is VPC.

### spec.codeInterpreters[].network.vpcConfig.subnets

`[]string | valueFrom` · required

Subnets the session network interfaces attach to (at least one).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.codeInterpreters[].network.vpcConfig.securityGroups

`[]string | valueFrom` · required

Security groups applied to the session network interfaces (at least
one).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.codeInterpreters[].certificates

`[]AwsBedrockAgentCoreToolCertificate`

Client certificates (in Secrets Manager) presented to
mTLS-protected endpoints the code calls (max 200).

- rule: {"repeated":{"maxItems":"200"}}

### spec.codeInterpreters[].certificates[].secretArn

`string | valueFrom` · required

The Secrets Manager secret holding the certificate material.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

## Validation Rules

- `at_least_one_tool`: set at least one of browsers, browser_profiles, or code_interpreters
- `browser_names_unique`: browsers entries must have unique names
- `browser_profile_names_unique`: browser_profiles entries must have unique names
- `code_interpreter_names_unique`: code_interpreters entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockAgentCoreTools, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.browser_ids` | `map<string, string>` | Browser IDs keyed by each `browsers` entry's name. |
| `status.outputs.browser_arns` | `map<string, string>` | Browser ARNs keyed by each `browsers` entry's name. |
| `status.outputs.browser_profile_ids` | `map<string, string>` | Browser profile IDs keyed by each `browser_profiles` entry's name. |
| `status.outputs.browser_profile_arns` | `map<string, string>` | Browser profile ARNs keyed by each `browser_profiles` entry's name. |
| `status.outputs.code_interpreter_ids` | `map<string, string>` | Code interpreter IDs keyed by each `code_interpreters` entry's name. |
| `status.outputs.code_interpreter_arns` | `map<string, string>` | Code interpreter ARNs keyed by each `code_interpreters` entry's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.browsers[].executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.browsers[].network.vpcConfig.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.browsers[].network.vpcConfig.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.browsers[].recording.s3Location.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.browsers[].enterprisePolicies[].s3.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.browsers[].certificates[].secretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.codeInterpreters[].executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.codeInterpreters[].network.vpcConfig.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.codeInterpreters[].network.vpcConfig.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.codeInterpreters[].certificates[].secretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |

## See Also

- [Overview](../README.md)
