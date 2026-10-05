# Auth0 Action

Deploys an Auth0 Action — a versioned Node.js function that executes at a chosen point in the Auth0 pipeline — with its trigger, source code, runtime, npm dependencies, and encrypted secrets managed as one Infra Component. Supports all ten trigger types, from post-login token enrichment to custom email and phone providers, and can optionally bind the deployed action into its trigger flow so it starts executing immediately.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Auth0 Action** — a versioned Node.js function registered in Auth0 with the specified trigger, source code, runtime, npm dependencies, and encrypted secrets
- **Auth0 Trigger Binding** — created only when `triggerBinding` is set; attaches the deployed action to its trigger flow so it executes during the corresponding pipeline stage

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** — an active connection in the Connect module with Auth0 domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **Management API scopes** — the M2M application behind the Provider Connection needs `create:actions`, `read:actions`, `update:actions`, and `delete:actions`. Trigger binding rides `update:actions` — no separate trigger scope family is needed. These scopes are tenant-wide: a token holding them can manage every action in the tenant.

## Deploy

### Console

Open the deployment store, find **Auth0 Action**, and click **Deploy**. The creation wizard walks you through trigger selection, action code, runtime configuration, and optional dependencies and secrets. Start from the **Post-Login Custom Claims**, **Pre-Registration Domain Allowlist**, or **Credentials Exchange Audit Log** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Action
metadata:
  name: enrich-tokens
  org: acme-corp
  env: prod
spec:
  supportedTrigger:
    id: post-login
    version: v3
  code: |
    exports.onExecutePostLogin = async (event, api) => {
      api.idToken.setCustomClaim('https://myapp/roles', event.authorization?.roles || []);
    };
  deploy: true
  triggerBinding:
    displayName: Enrich Tokens
```

```shell
planton apply -f auth0-action.yaml
```

This creates a post-login action, deploys it as an immutable version, and binds it to the post-login trigger flow — no dependencies or secrets configured. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an Auth0 Action. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Trigger selection** — `supportedTrigger.id` fixes when the action executes and which event and API objects the code receives: `post-login` for token enrichment and access control, `pre-user-registration` for signup validation, `credentials-exchange` for M2M flow customization. The exported function name must match the trigger (`exports.onExecutePostLogin`, `exports.onExecutePreUserRegistration`, and so on), and `supportedTrigger.version` must pair with the trigger (`v3` for post-login; most others are `v2` or `v1`).

**Deploy and bind** — `deploy: true` creates an immutable version on every apply; if the action is bound, the new version begins executing immediately. `triggerBinding` requires `deploy: true` — you cannot bind an undeployed action, and the manifest is rejected if you try. The binding appends the action to the END of the trigger's flow: when multiple actions share a trigger and order matters, omit `triggerBinding` here and manage the bindings externally.

**Runtime version** — Use `node22` for new actions; `node18` is in maintenance and actions on it should be migrated before end-of-life. When omitted, Auth0 assigns a default based on the trigger version.

**Secrets management** — The `secrets` array makes sensitive values available as `event.secrets.<name>` in action code, encrypted at rest and never returned by the API. Auth0 requires full secret management: if you manage secrets here, include ALL secrets for this action — omitting a previously-set secret deletes it.

**Dependencies** — The `dependencies` array names npm packages Auth0 resolves and bundles at deploy time. Pin versions explicitly (e.g., `1.7.0`) rather than using ranges — an unpinned dependency can change your authentication pipeline's behavior on a redeploy you didn't intend, and every package is third-party code running inside login.

**Latency budget** — Every action in a trigger's flow adds its execution time to authentication latency, and each action is killed at a 20-second timeout. External API calls inside `post-login` are paid by every user on every login — keep them rare and fast.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | Auth0-assigned action identifier | Managing trigger bindings externally, Management API operations |
| `version_id` | Identifier of the deployed immutable version (populated only when `deploy` is true) | Verifying which version is live, rollback via the Management API |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Post-login token enrichment** — Add custom claims (roles, email, organization) to ID and access tokens after authentication, so applications and APIs read identity context from the token instead of calling back to Auth0. Nearly every production tenant using RBAC needs this. Start from the **Post-Login Custom Claims** preset.

**Registration domain allowlist** — Block signups outside approved email domains during pre-registration, with the allowlist held in an action secret so changing it never touches code. The pattern for B2B and internal tools that must not accept public registrations. Start from the **Pre-Registration Domain Allowlist** preset.

**M2M audit logging** — POST every client-credentials exchange (client identity, audience, scopes, IP) to an external audit endpoint, failing gracefully so an audit outage never blocks token issuance. Start from the **Credentials Exchange Audit Log** preset.

## Works With

- [**Auth0 Role**](/infra-catalog/auth0-role) — role assignments populate the `event.authorization.roles` that post-login actions read when enriching tokens
- [**Auth0 Resource Server (API)**](/infra-catalog/auth0-resource-server) — the APIs whose access tokens post-login and credentials-exchange actions enrich and gate
