# AwsBedrockAgentCoreIdentity

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockAgentCoreIdentitySpec defines the desired configuration for
an Amazon Bedrock AgentCore identity-and-access bundle - the
credentials and authorization control plane your agents, gateways, and
tools authenticate through:

  - `workload_identities`: named identities AgentCore workloads present
    when calling other services;
  - `api_key_credential_providers` / `oauth2_credential_providers`:
    vaulted outbound credentials (AWS stores the secrets; gateway
    targets and harness tools reference the provider ARN, never the
    secret);
  - `policy_engine`: a Cedar authorization engine with its policies -
    gateways evaluate tool calls against it.

Every arm is optional; author the ones this bundle owns (at least
one). The account/region token-vault CMK is deliberately NOT part of
this kind - it is an account-level settings singleton.

## Example

```yaml
# Canonical AwsBedrockAgentCoreIdentity example (hack/dev manifest and
# refgen Example source): the full identity-and-access bundle -- workload
# identities, a vaulted API key, OAuth2 providers (well-known and custom
# vendors), and a Cedar policy engine with policies. The credential
# values are placeholders: real manifests supply managed-secret
# references resolved just-in-time at deploy.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreIdentity
metadata:
  name: support-agent-identity
  id: support-agent-identity
  org: test-org
  env: dev
spec:
  region: us-west-2
  workloadIdentities:
    - name: support-agent
      allowedResourceOauth2ReturnUrls:
        - https://app.example.com/oauth/callback
  apiKeyCredentialProviders:
    - name: docs-api
      apiKey: placeholder-api-key-value
  oauth2CredentialProviders:
    - name: github
      vendor: GITHUB
      clientId: Iv1.placeholder
      clientSecret: placeholder-client-secret
    - name: internal-idp
      vendor: CUSTOM
      clientId: planton-agents
      clientSecret: placeholder-client-secret
      oauthDiscovery:
        discoveryUrl: https://idp.example.com/.well-known/openid-configuration
    - name: legacy-idp
      vendor: CUSTOM
      clientId: legacy-client
      clientSecret: placeholder-client-secret
      oauthDiscovery:
        authorizationServerMetadata:
          issuer: https://legacy.example.com
          authorizationEndpoint: https://legacy.example.com/authorize
          tokenEndpoint: https://legacy.example.com/token
          responseTypes:
            - code
  policyEngine:
    engineName: agent_authz
    description: Tool-call authorization for the support agent
    policies:
      - name: allow_order_reads
        description: Let every principal call the order lookup tool
        cedarStatement: 'permit(principal, action == Action::"get_order", resource is AgentCore::Gateway);'
        validationMode: FAIL_ON_ANY_FINDINGS
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.workloadIdentities` | `[]AwsBedrockAgentCoreWorkloadIdentity` |  |  |  |
| `spec.workloadIdentities[].name` | `string` | yes |  |  |
| `spec.workloadIdentities[].allowedResourceOauth2ReturnUrls` | `[]string` |  |  |  |
| `spec.apiKeyCredentialProviders` | `[]AwsBedrockAgentCoreApiKeyProvider` |  |  |  |
| `spec.apiKeyCredentialProviders[].name` | `string` | yes |  |  |
| `spec.apiKeyCredentialProviders[].apiKey` | `string` (sensitive) | yes |  |  |
| `spec.oauth2CredentialProviders` | `[]AwsBedrockAgentCoreOauth2Provider` |  |  |  |
| `spec.oauth2CredentialProviders[].name` | `string` | yes |  |  |
| `spec.oauth2CredentialProviders[].vendor` | `string` |  |  |  |
| `spec.oauth2CredentialProviders[].clientId` | `string` | yes |  |  |
| `spec.oauth2CredentialProviders[].clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.oauth2CredentialProviders[].oauthDiscovery` | `AwsBedrockAgentCoreOauth2Discovery` |  |  |  |
| `spec.oauth2CredentialProviders[].oauthDiscovery.discoveryUrl` | `string` |  |  |  |
| `spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata` | `AwsBedrockAgentCoreOauth2ServerMetadata` |  |  |  |
| `spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata.issuer` | `string` | yes |  |  |
| `spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata.authorizationEndpoint` | `string` | yes |  |  |
| `spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata.tokenEndpoint` | `string` | yes |  |  |
| `spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata.responseTypes` | `[]string` |  |  |  |
| `spec.policyEngine` | `AwsBedrockAgentCorePolicyEngine` |  |  |  |
| `spec.policyEngine.engineName` | `string` | yes |  |  |
| `spec.policyEngine.description` | `string` |  |  |  |
| `spec.policyEngine.encryptionKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.policyEngine.policies` | `[]AwsBedrockAgentCoreCedarPolicy` |  |  |  |
| `spec.policyEngine.policies[].name` | `string` | yes |  |  |
| `spec.policyEngine.policies[].description` | `string` |  |  |  |
| `spec.policyEngine.policies[].cedarStatement` | `string` | yes |  |  |
| `spec.policyEngine.policies[].validationMode` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the identity resources will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.workloadIdentities

`[]AwsBedrockAgentCoreWorkloadIdentity`

Named identities AgentCore workloads (runtimes, gateways) present
when calling other services.

### spec.workloadIdentities[].name

`string` · required

Identity name in AWS (3-255 characters; letters, digits, and
_ . -). The for_each key on both engines and the key in the
`workload_identity_arns` output map. Changing it replaces the
identity.

- rule: {"string":{"minLen":"3","maxLen":"255","pattern":"^[A-Za-z0-9_.-]+$"}}

### spec.workloadIdentities[].allowedResourceOauth2ReturnUrls

`[]string`

OAuth2 redirect URLs this workload may use in user-delegated token
flows (the browser must land on an allow-listed URL). AWS validates
each URL server-side at CreateWorkloadIdentity and rejects
reserved/documentation TLDs ("Invalid redirect url" -- .test
rejected, live-caught 2026-08-14); real-TLD https URLs and
http://localhost URLs are accepted.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.apiKeyCredentialProviders

`[]AwsBedrockAgentCoreApiKeyProvider`

Vaulted API keys for outbound calls (AWS stores each key in Secrets
Manager under the service's token vault).

### spec.apiKeyCredentialProviders[].name

`string` · required

Provider name in AWS (1-128 characters; letters, digits, hyphen,
underscore). The for_each key on both engines and the key in the
`api_key_provider_arns` output map. Changing it replaces the
provider.

- rule: {"string":{"minLen":"1","maxLen":"128","pattern":"^[a-zA-Z0-9\\-_]+$"}}

### spec.apiKeyCredentialProviders[].apiKey

`string` · required · sensitive

The API key to vault. Handled as a sensitive value end to end: use a
secret reference in manifests (resolved just-in-time at deploy),
never a literal. Rotating the key updates in place.

- rule: {"string":{"minLen":"1"}}

### spec.oauth2CredentialProviders

`[]AwsBedrockAgentCoreOauth2Provider`

Vaulted OAuth2 client credentials for outbound token flows.

- rule: oauth_discovery is required when vendor is CUSTOM and forbidden otherwise

### spec.oauth2CredentialProviders[].name

`string` · required

Provider name in AWS (1-128 characters; letters, digits, hyphen,
underscore). The for_each key on both engines and the key in the
`oauth2_provider_arns` output map. Changing it replaces the provider.

- rule: {"string":{"minLen":"1","maxLen":"128","pattern":"^[a-zA-Z0-9\\-_]+$"}}

### spec.oauth2CredentialProviders[].vendor

`string`

The OAuth vendor. The well-known vendors (GOOGLE, GITHUB, MICROSOFT,
SALESFORCE, SLACK) carry AWS-known endpoints - no discovery needed;
CUSTOM requires `oauth_discovery`. Changing the vendor replaces the
provider.

- rule: {"string":{"in":["CUSTOM","GITHUB","GOOGLE","MICROSOFT","SALESFORCE","SLACK"]}}

### spec.oauth2CredentialProviders[].clientId

`string` · required

The OAuth2 client ID (1-256 characters). Sensitive: use a secret
reference in manifests, never a literal.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.oauth2CredentialProviders[].clientSecret

`string` · required · sensitive

The OAuth2 client secret (1-2048 characters). Sensitive: use a
secret reference in manifests (resolved just-in-time at deploy),
never a literal. Rotating it updates in place.

- rule: {"string":{"minLen":"1","maxLen":"2048"}}

### spec.oauth2CredentialProviders[].oauthDiscovery

`AwsBedrockAgentCoreOauth2Discovery`

How to reach a CUSTOM vendor's endpoints - required exactly when
vendor is CUSTOM.

- rule: set exactly one of discovery_url or authorization_server_metadata

### spec.oauth2CredentialProviders[].oauthDiscovery.discoveryUrl

`string`

The vendor's OIDC discovery URL (must serve
/.well-known/openid-configuration).

### spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata

`AwsBedrockAgentCoreOauth2ServerMetadata`

Or spell the endpoints out when the vendor has no discovery
document.

### spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata.issuer

`string` · required

The token issuer URL (the "iss" claim value).

- rule: {"string":{"minLen":"1"}}

### spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata.authorizationEndpoint

`string` · required

The authorization endpoint (user-delegated flows).

- rule: {"string":{"minLen":"1"}}

### spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata.tokenEndpoint

`string` · required

The token endpoint (all flows).

- rule: {"string":{"minLen":"1"}}

### spec.oauth2CredentialProviders[].oauthDiscovery.authorizationServerMetadata.responseTypes

`[]string`

OAuth response types the server supports. Omitted = AWS defaults.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.policyEngine

`AwsBedrockAgentCorePolicyEngine`

A Cedar policy engine with its policies.

- rule: policies entries must have unique names

### spec.policyEngine.engineName

`string` · required

Engine name in AWS (1-48 characters; letter first, then letters,
digits, underscore - AWS rejects hyphens here). Changing it replaces
the engine.

- rule: {"string":{"minLen":"1","maxLen":"48","pattern":"^[A-Za-z][A-Za-z0-9_]*$"}}

### spec.policyEngine.description

`string`

Human-readable description (1-4096 characters when set).

- rule: {"string":{"maxLen":"4096"}}

### spec.policyEngine.encryptionKeyArn

`string | valueFrom`

Customer-managed KMS key encrypting the engine's policies at rest.
Without it, AWS uses a service-managed key. Changing it replaces the
engine.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.policyEngine.policies

`[]AwsBedrockAgentCoreCedarPolicy`

The engine's Cedar policies.

### spec.policyEngine.policies[].name

`string` · required

Policy name in AWS (1-48 characters; letter first, then letters,
digits, underscore). The for_each key on both engines and the key in
the `policy_ids` output map. Changing it replaces the policy.

- rule: {"string":{"minLen":"1","maxLen":"48","pattern":"^[A-Za-z][A-Za-z0-9_]*$"}}

### spec.policyEngine.policies[].description

`string`

Human-readable description (1-4096 characters when set).

- rule: {"string":{"maxLen":"4096"}}

### spec.policyEngine.policies[].cedarStatement

`string` · required

The Cedar policy statement (permit/forbid rules over principals,
actions, and resources). AWS rejects a fully-wildcard resource at
CreatePolicy ("a wildcard resource was detected", live-caught
2026-08-14): constrain the resource to a specific
AgentCore::Gateway entity or to the type, e.g.
`permit(principal, action, resource is AgentCore::Gateway);`.

- rule: {"string":{"minLen":"1"}}

### spec.policyEngine.policies[].validationMode

`string`

What AWS does with static-analysis findings on the statement:
FAIL_ON_ANY_FINDINGS rejects a policy with findings (the safe
default); IGNORE_ALL_FINDINGS stores it anyway. Omitted = AWS
default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FAIL_ON_ANY_FINDINGS","IGNORE_ALL_FINDINGS"]}}

## Validation Rules

- `at_least_one_arm`: set at least one of workload_identities, api_key_credential_providers, oauth2_credential_providers, or policy_engine
- `workload_identity_names_unique`: workload_identities entries must have unique names
- `api_key_provider_names_unique`: api_key_credential_providers entries must have unique names
- `oauth2_provider_names_unique`: oauth2_credential_providers entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockAgentCoreIdentity, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.workload_identity_arns` | `map<string, string>` | Workload identity ARNs keyed by each `workload_identities` entry's name. |
| `status.outputs.api_key_provider_arns` | `map<string, string>` | Credential provider ARNs keyed by each `api_key_credential_providers` entry's name - the value a gateway target's api_key credentials consume. |
| `status.outputs.api_key_secret_arns` | `map<string, string>` | Secrets Manager secret ARNs holding each vaulted API key, keyed by provider name. |
| `status.outputs.oauth2_provider_arns` | `map<string, string>` | Credential provider ARNs keyed by each `oauth2_credential_providers` entry's name - the value a gateway target's oauth credentials consume. |
| `status.outputs.oauth2_client_secret_arns` | `map<string, string>` | Secrets Manager secret ARNs holding each vaulted OAuth client secret, keyed by provider name. |
| `status.outputs.policy_engine_id` | `string` | The policy engine's unique identifier (empty when the bundle has no engine). |
| `status.outputs.policy_engine_arn` | `string` | The policy engine's ARN - the value a gateway's policy_engine configuration consumes. |
| `status.outputs.policy_ids` | `map<string, string>` | Cedar policy IDs keyed by each `policy_engine.policies` entry's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.policyEngine.encryptionKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgentCoreGateway | `spec.policyEngine.policyEngineArn` | `status.outputs.policy_engine_arn` |

## See Also

- [Overview](../README.md)
