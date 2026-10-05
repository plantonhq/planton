# AWS Bedrock AgentCore Identity

Deploys an Amazon Bedrock AgentCore identity-and-access bundle — the credentials and authorization control plane your agents, gateways, and tools authenticate through. It vaults API keys and OAuth2 clients as credential providers that consumers reference by ARN (never by value), names the workload identities AgentCore workloads present, and hosts a Cedar policy engine that gateways evaluate every tool call against. Each arm is optional; the bundle needs at least one.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions, per named entry:

- **Workload Identity** — one per `workloadIdentities` entry, with its OAuth2 return-URL allow-list for user-delegated token flows
- **API Key Credential Provider** — one per `apiKeyCredentialProviders` entry; AWS stores the key in Secrets Manager under the service's token vault
- **OAuth2 Credential Provider** — one per `oauth2CredentialProviders` entry: a well-known vendor (GOOGLE, GITHUB, MICROSOFT, SALESFORCE, SLACK) with AWS-known endpoints, or CUSTOM with your own OIDC discovery; the client secret is vaulted the same way
- **Cedar Policy Engine** — created only when `policyEngine` is set, plus one **Cedar Policy** per `policies` entry — each policy is a structural child of its engine, created after it and destroyed before it

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AgentCore identity permissions (`bedrock-agentcore:CreateWorkloadIdentity`, `CreateApiKeyCredentialProvider`, `CreateOauth2CredentialProvider`, `CreatePolicyEngine`, `CreatePolicy` and their siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Org secrets for every credential value** — `apiKey`, `clientId`, and `clientSecret` are sensitive fields carrying `$secret/<slug>` references resolved just-in-time at deploy, never literals. Create the org secrets before applying the manifest.

### AWS Account

- Bedrock AgentCore available in the target region
- The real OAuth client registered with its vendor (only for `oauth2CredentialProviders` — AWS vaults the credentials, it does not create them)

## Deploy

### Console

Open the deployment store, find **AWS Bedrock AgentCore Identity**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the bundle arms — workload identities, credential providers, and the policy engine. Start from the **Vaulted Credentials** preset in the [Presets](#presets) tab for the credential-provider shape, or the **Cedar Authorization** preset for a policy engine with permit/forbid rules.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreIdentity
metadata:
  name: agent-fleet-identity
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  apiKeyCredentialProviders:
    - name: docs-api
      apiKey: $secret/docs-api-key
  oauth2CredentialProviders:
    - name: github
      vendor: GITHUB
      clientId: $secret/github-client-id
      clientSecret: $secret/github-client-secret
```

```shell
planton apply -f agentcore-identity.yaml
```

This vaults one API key and one GitHub OAuth client as credential providers — the values come from managed secrets resolved just-in-time at deploy, and consumers will reference the provider ARNs. An Infra Job tracks the provisioning in real time.

### InfraChart

When the bundle's policy engine encrypts with a customer-managed key deployed in the same chart, wire the key reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  policyEngine:
    engineName: agent_authz
    encryptionKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: agentcore-policies-key
        fieldPath: status.outputs.key_arn
    policies:
      - name: allow_read_tools
        cedarStatement: 'permit(principal, action in [Action::"get_order"], resource is AgentCore::Gateway);'
```

The InfraPipeline resolves the dependency graph, creates the KMS key first, then provisions the engine encrypting with it.

## Key Configuration

These are the most important decisions when configuring an identity bundle. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Reference providers, never secrets** — consumers (gateway targets, harness tools) take the provider ARN from the output maps; the secret itself lives in AWS's token vault. Rotation is an in-place update: change the org secret's value and re-apply — consumers never change. This indirection is the whole point of the component.

**Names are identity and force replacement** — renaming a workload identity, credential provider, engine, or policy replaces it, and in-flight token grants against the old provider name die. Rename during quiet windows. Engine and policy names additionally reject hyphens (a letter, then letters/digits/underscores).

**The vendor field is the discriminator** — the five well-known OAuth vendors carry AWS-known endpoints, so no discovery is needed; CUSTOM pairs exactly with `oauthDiscovery` (a discovery URL, or spelled-out server metadata for vendors without a discovery document). The manifest validation enforces the pairing.

**Always constrain the Cedar resource** — AWS rejects a fully-wildcard resource at CreatePolicy ("a wildcard resource was detected") regardless of validation mode. Scope every statement to a specific `AgentCore::Gateway` entity or at least the type: `resource is AgentCore::Gateway`.

**Prefer FAIL_ON_ANY_FINDINGS** — Cedar's static analysis catches policies that can never match. IGNORE_ALL_FINDINGS is for deliberate forward-references to entities that do not exist yet, not a default.

**Return URLs are validated server-side** — CreateWorkloadIdentity rejects reserved and documentation TLDs in `allowedResourceOauth2ReturnUrls` ("Invalid redirect url" — `.test` is rejected); real-TLD HTTPS URLs and `http://localhost` URLs are accepted.

**Scope one bundle per trust domain** — one Identity per team or agent fleet keeps the blast radius of a leaked provider-ARN grant small and the Cedar policy set readable. The account/region token-vault encryption key is deliberately not part of this kind — it is an account-level singleton owned by the AWS Bedrock AgentCore Token Vault component.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `policyEngine.encryptionKeyArn` | `status.outputs.key_arn` |

Credential values (`apiKey`, `clientId`, `clientSecret`) travel as `$secret/<slug>` managed-secret references rather than foreign keys.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `api_key_provider_arns` | Credential provider ARNs keyed by provider name | Gateway target `apiKey` credentials |
| `oauth2_provider_arns` | Credential provider ARNs keyed by provider name | Gateway target `oauth` credentials; harness gateway-tool OAuth |
| `policy_engine_arn` | The Cedar policy engine's ARN | A gateway's `policyEngine` attachment (LOG_ONLY or ENFORCE) |
| `workload_identity_arns` | Workload identity ARNs keyed by entry name | JWT authorizers' `allowedWorkloads` restrictions on other AgentCore resources |
| `api_key_secret_arns` / `oauth2_client_secret_arns` | Secrets Manager secret ARNs holding each vaulted value, keyed by provider name | IAM policies granting vault-secret access; locating the AWS-managed entries |

`policy_engine_id` and `policy_ids` are also exported; they identify the engine and its policies for operational tooling rather than feeding composition.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Vaulted outbound credentials** — API key and OAuth2 providers holding everything your gateways and harnesses present to backends. Consumers wire the provider ARN from the output maps, so rotating any credential is a secret-value change and a re-apply — no gateway or harness edit, ever. Start from the **Vaulted Credentials** preset.

**Tool-call authorization with Cedar** — a policy engine with permit rules for read-only tools and forbid rules for mutating ones, attached to a gateway in LOG_ONLY until the decision log looks right, then flipped to ENFORCE. Start from the **Cedar Authorization** preset.

**One bundle per agent fleet** — a workload identity, that fleet's credential providers, and its policy engine in one resource. The bundle boundary is the trust boundary: granting access to one fleet's provider ARNs exposes nothing of another's.

## Works With

- [**AWS Bedrock AgentCore Gateway**](/infra-catalog/aws-bedrock-agent-core-gateway) — consumes provider ARNs for target credentials and the policy engine ARN for tool-call authorization
- [**AWS Bedrock AgentCore Evaluation**](/infra-catalog/aws-bedrock-agent-core-evaluation) — harness gateway tools authenticate through this bundle's OAuth2 providers
- [**AWS Bedrock AgentCore Token Vault**](/infra-catalog/aws-bedrock-agent-core-token-vault) — the account/region vault whose encryption key governs the secrets this bundle stores
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for the policy engine's policies at rest
