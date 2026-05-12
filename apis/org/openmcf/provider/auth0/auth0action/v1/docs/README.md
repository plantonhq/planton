# Auth0Action — Research Documentation

## What Are Auth0 Actions?

Auth0 Actions are server-side Node.js functions that run at specific extension points (triggers) in the Auth0 pipeline. They replace the deprecated Rules and Hooks system with a more structured, versioned, and testable approach to customizing authentication flows.

Actions are:
- **Tenant-scoped**: Each action belongs to a single Auth0 tenant.
- **Versioned**: Each deployment creates an immutable version. Rollback is possible.
- **Trigger-specific**: Each action targets exactly one trigger (e.g., post-login).
- **Sandboxed**: Actions run in Auth0's Node.js sandbox with access to `event` and `api` objects.

## Trigger Types and Versions

Auth0 provides 10 trigger types. Each trigger has a version that determines the event/api contract available to action code.

| Trigger | Current Version | Event Object | API Object | Common Uses |
|---|---|---|---|---|
| `post-login` | v3 | User, connection, request, session | Token claims, MFA, deny | Token enrichment, conditional MFA |
| `credentials-exchange` | v2 | Client, request | Token claims, deny | M2M token enrichment, audit |
| `pre-user-registration` | v2 | User, connection | Deny, set metadata | Registration gating |
| `post-user-registration` | v2 | User, connection | (none) | Notifications, provisioning |
| `post-change-password` | v2 | User, connection | (none) | Alerts, audit logs |
| `send-phone-message` | v2 | Message options | (none) | Custom SMS/voice providers |
| `password-reset-post-challenge` | v1 | User, request | (none) | Post-reset actions |
| `custom-email-provider` | v1 | Message options | (none) | Custom email delivery |
| `custom-phone-provider` | v1 | Message options | (none) | Custom phone delivery |
| `custom-token-exchange` | v1 | Request, client | Token claims, deny | Federated token exchange |

## Action Lifecycle

1. **Create**: Define the action (name, code, trigger, runtime, dependencies, secrets).
2. **Build**: Auth0 compiles the action and installs npm dependencies. Status: `building` → `built` or `failed`.
3. **Deploy**: Create an immutable version from the built action. Only deployed actions can be bound to triggers.
4. **Bind**: Attach the deployed action to its trigger. Multiple actions can be bound to the same trigger; execution order matters.
5. **Execute**: Auth0 invokes the action during the corresponding pipeline stage.
6. **Update**: Modify code/config → re-build → re-deploy. The new version replaces the old one in bound triggers.

## Runtime Versions

| Runtime | Node.js Version | Status | Notes |
|---|---|---|---|
| `node18` | 18.x LTS | Maintenance | Still supported but not recommended for new actions |
| `node22` | 22.x LTS | Current | Recommended for all new actions |
| `node12`, `node16` | 12.x, 16.x | Deprecated | Available in Terraform provider for legacy compatibility |

If `runtime` is omitted, Auth0 assigns a default based on the trigger version. For most triggers with v2+, this defaults to `node18` (the Terraform provider maps `node18` to the internal `node18-actions` runtime string).

## Dependencies and Secrets

**Dependencies** are npm packages installed during the build phase. Specify name and exact version (or semver range). Common packages: `axios`, `lodash`, `twilio`, `@sendgrid/mail`.

**Secrets** are encrypted key-value pairs accessible at runtime via `event.secrets.<name>`. Auth0 encrypts secrets at rest and never returns values through the Management API. The IaC provider manages secrets as a full set — adding or removing a secret in config is an all-or-nothing operation.

## Trigger Bindings

After deployment, an action must be **bound** to a trigger to execute. Bindings are ordered — the execution sequence within a trigger flow is determined by binding order.

The Terraform provider offers two binding resources:
- `auth0_trigger_action`: Appends a single action to a trigger (no ordering control).
- `auth0_trigger_actions`: Manages the full ordered list of actions for a trigger.

The Pulumi provider mirrors this with `auth0.TriggerAction` and `auth0.TriggerActions`.

In this OpenMCF component, trigger binding is modeled as an optional inline field (`trigger_binding`) for the 80/20 case: create, deploy, and bind in a single manifest. For complex multi-action ordering, omit `trigger_binding` and manage bindings externally.

## Provider Parity

| Feature | Terraform | Pulumi | OpenMCF Spec |
|---|---|---|---|
| Create action | `auth0_action` | `auth0.Action` | `Auth0ActionSpec` |
| Deploy on create/update | `deploy = true` | `Deploy: true` | `deploy: true` |
| Bind to trigger | `auth0_trigger_action` | `auth0.TriggerAction` | `trigger_binding` (inline) |
| npm dependencies | `dependencies` block | `Dependencies` | `dependencies` list |
| Secrets | `secrets` block | `Secrets` | `secrets` list |
| Runtime | `runtime` | `Runtime` | `runtime` |
| Modules | `modules` block | Not available | Not supported (80/20) |

The Terraform `modules` attribute (for Auth0 Action Modules, a marketplace feature) is intentionally excluded from the OpenMCF spec. Action Modules are a separate Auth0 concept and rarely used in IaC-managed deployments.

## References

- [Auth0 Actions Overview](https://auth0.com/docs/customize/actions)
- [Actions Coding Guidelines](https://auth0.com/docs/customize/actions/action-coding-guidelines)
- [Flows and Triggers](https://auth0.com/docs/customize/actions/flows-and-triggers)
- [Terraform auth0_action](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/action)
- [Terraform auth0_trigger_action](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/trigger_action)
- [Pulumi auth0.Action](https://www.pulumi.com/registry/packages/auth0/api-docs/action/)
- [Pulumi auth0.TriggerAction](https://www.pulumi.com/registry/packages/auth0/api-docs/triggeraction/)
