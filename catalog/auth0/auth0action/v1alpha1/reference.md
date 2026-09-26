# Auth0Action

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0ActionSpec defines the configuration for an Auth0 Action.
Actions are secure, tenant-specific, versioned functions written in Node.js
that execute at certain points during the Auth0 runtime. They are used to
customize and extend Auth0's capabilities with custom logic.

This spec covers the 80/20 use case for managing Auth0 Actions:
- Post-login token enrichment with custom claims
- Pre-registration validation (e.g., email domain allowlists)
- Credentials-exchange customization for M2M flows
- Custom phone/email provider integration

https://auth0.com/docs/customize/actions
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/action
https://www.pulumi.com/registry/packages/auth0/api-docs/action/

## Example

```yaml
# Test manifest for Auth0Action
# This file is used for local testing and development

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Action
metadata:
  name: test-post-login-action
  org: test-organization
  env: development
  labels:
    team: platform
    purpose: testing
spec:
  supported_trigger:
    id: post-login
    version: v3
  code: |
    exports.onExecutePostLogin = async (event, api) => {
      const namespace = 'https://test.example.com';
      api.idToken.setCustomClaim(`${namespace}/email`, event.user.email);
      api.accessToken.setCustomClaim(`${namespace}/roles`, event.authorization?.roles || []);
    };
  runtime: node22
  deploy: true
  trigger_binding:
    display_name: Test Post Login Action
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.supportedTrigger` | `Auth0ActionSupportedTrigger` | yes |  |  |
| `spec.supportedTrigger.id` | `string` | yes |  |  |
| `spec.supportedTrigger.version` | `string` | yes |  |  |
| `spec.code` | `string` | yes |  |  |
| `spec.runtime` | `string` |  |  |  |
| `spec.deploy` | `bool` |  |  |  |
| `spec.dependencies` | `[]Auth0ActionDependency` |  |  |  |
| `spec.dependencies[].name` | `string` | yes |  |  |
| `spec.dependencies[].version` | `string` | yes |  |  |
| `spec.secrets` | `[]Auth0ActionSecret` |  |  |  |
| `spec.secrets[].name` | `string` | yes |  |  |
| `spec.secrets[].value` | `string` (sensitive) | yes |  |  |
| `spec.triggerBinding` | `Auth0ActionTriggerBinding` |  |  |  |
| `spec.triggerBinding.displayName` | `string` |  |  |  |

## Field Details

### spec.supportedTrigger

`Auth0ActionSupportedTrigger` · required

supported_trigger defines the single trigger this action targets.
Auth0 limits each action to exactly one trigger.

The trigger determines when the action executes in the Auth0 pipeline
and which event/api objects are available to the action code.

https://auth0.com/docs/customize/actions/flows-and-triggers

- rule: {"required":true}

### spec.supportedTrigger.id

`string` · required

id is the trigger type identifier.
Determines when in the Auth0 pipeline this action executes.

Available triggers:
- "post-login": After user authenticates (most common)
- "credentials-exchange": During client_credentials or token exchange
- "pre-user-registration": Before a new user is created
- "post-user-registration": After a new user is created
- "post-change-password": After a user changes their password
- "send-phone-message": When Auth0 needs to send an SMS/voice message
- "password-reset-post-challenge": After password reset verification
- "custom-email-provider": Custom email delivery
- "custom-phone-provider": Custom phone/SMS delivery
- "custom-token-exchange": Custom token exchange flow

https://auth0.com/docs/customize/actions/flows-and-triggers

- rule: {"required":true,"string":{"in":["post-login","credentials-exchange","pre-user-registration","post-user-registration","post-change-password","send-phone-message","password-reset-post-challenge","custom-email-provider","custom-phone-provider","custom-token-exchange"]}}

### spec.supportedTrigger.version

`string` · required

version is the trigger API version. This determines which runtime versions
are compatible and which event/api objects are available.

Common versions:
- "v3" for post-login (current)
- "v2" for credentials-exchange
- "v2" for pre-user-registration
- "v2" for post-user-registration
- "v2" for post-change-password
- "v2" for send-phone-message
- "v1" for password-reset-post-challenge
- "v1" for custom-email-provider
- "v1" for custom-phone-provider
- "v1" for custom-token-exchange

https://registry.terraform.io/providers/auth0/auth0/latest/docs/guides/action_triggers

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.code

`string` · required

code is the Node.js source code of the action.
The exported function name must match the trigger type:
  - post-login: exports.onExecutePostLogin
  - credentials-exchange: exports.onExecuteCredentialsExchange
  - pre-user-registration: exports.onExecutePreUserRegistration
  - post-user-registration: exports.onExecutePostUserRegistration
  - post-change-password: exports.onExecutePostChangePassword
  - send-phone-message: exports.onExecuteSendPhoneMessage
  - password-reset-post-challenge: exports.onExecutePasswordResetPostChallenge
  - custom-email-provider: exports.onExecuteCustomEmailProvider
  - custom-phone-provider: exports.onExecuteCustomPhoneProvider
  - custom-token-exchange: exports.onExecuteCustomTokenExchange

https://auth0.com/docs/customize/actions/write-your-first-action

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.runtime

`string`

runtime is the Node.js runtime version for this action.
If omitted, Auth0 assigns a default based on the trigger version.

- "node18": Node.js 18 (maintenance, not recommended for new actions)
- "node22": Node.js 22 (current, recommended)

https://auth0.com/docs/customize/actions/action-coding-guidelines#supported-node-versions

- rule: {"string":{"in":["","node18","node22"]}}

### spec.deploy

`bool`

deploy controls whether the action is automatically deployed after creation or update.
A deployed action creates an immutable version. If the action is bound to a trigger,
the new version begins executing immediately.

Set to true (recommended) for most use cases. Set to false only when you want to
stage the action code without activating it.

Default: true

### spec.dependencies

`[]Auth0ActionDependency`

dependencies is a list of third-party npm modules this action depends on.
Auth0 installs these packages when building the action.

Example:
  dependencies:
    - name: lodash
      version: "4.17.21"
    - name: axios
      version: "1.6.0"

https://auth0.com/docs/customize/actions/manage-dependencies

- rule: Each dependency needs a package name. Specify the npm package name, e.g., 'lodash' or 'axios'.
- rule: Each dependency needs a version. Use a semver range like '4.17.21', '^1.0.0', or 'latest'.

### spec.dependencies[].name

`string` · required

name is the npm package name.
Example: "lodash", "axios", "@sendgrid/mail"

- rule: {"required":true}

### spec.dependencies[].version

`string` · required

version is the npm package version or semver range.
Example: "4.17.21", "^1.0.0", "latest"

- rule: {"required":true}

### spec.secrets

`[]Auth0ActionSecret`

secrets is a list of key-value secrets available to the action at runtime via
event.secrets.<name>. Secrets are encrypted at rest and never returned by the API.

Use secrets for API keys, connection strings, tokens, and other sensitive values.
Avoid hardcoding secrets in the action code.

Important: Auth0 requires full secret management -- if you manage secrets here,
you must include ALL secrets for this action. Omitting a previously-set secret
will cause it to be deleted.

Example:
  secrets:
    - name: SLACK_WEBHOOK_URL
      value: $secret/slack-webhook-url

https://auth0.com/docs/customize/actions/write-your-first-action#add-a-secret

- rule: Each secret needs a name. This becomes the key you reference in action code as event.secrets.<name>.
- rule: Each secret needs a value: a $secret/<slug> reference to the managed secret that holds it (the literal only on a deploy without the platform).

### spec.secrets[].name

`string` · required

name is the secret key, referenced in code as event.secrets.<name>.
Must be unique within the action. Convention: UPPER_SNAKE_CASE.
Example: "SLACK_WEBHOOK_URL", "API_KEY"

- rule: {"required":true}

### spec.secrets[].value

`string` · required · sensitive

value is the secret content, and a secret field: on Planton it takes only a reference to a
managed secret (`$secret/<slug>`), which the runner resolves at deploy, so the secret is
never stored with the resource; a deploy without the platform takes the literal. Auth0
encrypts it at rest and never returns it from its API.
Example: "$secret/slack-webhook-url"

- rule: {"required":true}

### spec.triggerBinding

`Auth0ActionTriggerBinding`

trigger_binding optionally binds this action to its supported trigger after deployment.
When set, the action is both deployed and attached to the trigger flow so that it
executes during the corresponding Auth0 pipeline stage.

When omitted, the action is created (and optionally deployed) but NOT bound to
any trigger. This is useful when trigger ordering is managed separately.

Note: the action is appended to the end of the trigger's flow. To control
execution order across multiple actions, manage bindings externally.

### spec.triggerBinding.displayName

`string`

display_name is the label shown for this action within the trigger flow.
If omitted, defaults to the action's metadata.name.

Example: "Enrich Token Claims", "Validate Email Domain"

## Validation Rules

- `trigger_binding_requires_deploy`: An action must be deployed before it can be bound to a trigger. Either remove trigger_binding or set deploy to true.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0Action, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the unique identifier of the Auth0 action. This is assigned by Auth0 and used to reference the action in trigger bindings and API calls. |
| `status.outputs.name` | `string` | name is the name of the action as configured in metadata.name. |
| `status.outputs.version_id` | `string` | version_id is the identifier of the currently deployed version. Only populated when deploy is true. Each deployment creates a new immutable version with a unique ID. |
| `status.outputs.runtime` | `string` | runtime is the resolved Node.js runtime version. Reflects the actual runtime used, which may differ from the requested value if Auth0 assigns a default. |

## See Also

- [Overview](../README.md)
