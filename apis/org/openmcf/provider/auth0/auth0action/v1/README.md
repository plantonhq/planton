# Auth0Action

Manages an [Auth0 Action](https://auth0.com/docs/customize/actions) — a secure, tenant-specific Node.js function that executes at a designated point in the Auth0 pipeline. Actions customize authentication flows, enrich tokens, enforce policies, and integrate with external systems.

## When to Use

- **Token enrichment**: Add custom claims to ID or access tokens after login.
- **Registration gating**: Restrict user sign-ups by email domain, IP, or metadata.
- **M2M customization**: Modify client-credentials exchange behavior.
- **External notifications**: Send Slack/email alerts on login or password change.
- **Custom MFA**: Enforce conditional multi-factor authentication.
- **Custom providers**: Deliver SMS or email through non-default channels.

## Quick Start

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Action
metadata:
  name: enrich-token-claims
  org: acme-corp
  env: production
spec:
  supported_trigger:
    id: post-login
    version: v3
  code: |
    exports.onExecutePostLogin = async (event, api) => {
      const namespace = 'https://myapp.example.com';
      api.idToken.setCustomClaim(`${namespace}/roles`, event.authorization?.roles || []);
      api.accessToken.setCustomClaim(`${namespace}/email`, event.user.email);
    };
  deploy: true
  trigger_binding:
    display_name: Enrich Token Claims
```

## Supported Triggers

| Trigger ID | Description | Export Function |
|---|---|---|
| `post-login` | After user authenticates | `onExecutePostLogin` |
| `credentials-exchange` | Client credentials / token exchange | `onExecuteCredentialsExchange` |
| `pre-user-registration` | Before new user is created | `onExecutePreUserRegistration` |
| `post-user-registration` | After new user is created | `onExecutePostUserRegistration` |
| `post-change-password` | After password change | `onExecutePostChangePassword` |
| `send-phone-message` | SMS/voice delivery | `onExecuteSendPhoneMessage` |
| `password-reset-post-challenge` | After password reset verification | `onExecutePasswordResetPostChallenge` |
| `custom-email-provider` | Custom email delivery | `onExecuteCustomEmailProvider` |
| `custom-phone-provider` | Custom phone/SMS delivery | `onExecuteCustomPhoneProvider` |
| `custom-token-exchange` | Custom token exchange flow | `onExecuteCustomTokenExchange` |

## Key Behaviors

- **deploy**: When `true` (recommended), the action is compiled and a new immutable version is created. Bound triggers immediately use the new version.
- **trigger_binding**: When set, the action is both deployed and attached to its trigger. When omitted, the action is created/deployed but not bound (useful for external binding management).
- **secrets**: All secrets must be declared together. Omitting a previously-configured secret causes its deletion.
- **runtime**: Node.js 22 (recommended) or Node.js 18. Omit to use Auth0's default for the trigger version.

## Outputs

| Output | Description |
|---|---|
| `id` | Auth0 action identifier |
| `name` | Action name |
| `version_id` | Deployed version ID (when deploy=true) |
| `runtime` | Resolved Node.js runtime |

## Auth0 Documentation

- [Actions overview](https://auth0.com/docs/customize/actions)
- [Write your first action](https://auth0.com/docs/customize/actions/write-your-first-action)
- [Flows and triggers](https://auth0.com/docs/customize/actions/flows-and-triggers)
- [Manage dependencies](https://auth0.com/docs/customize/actions/manage-dependencies)
- [Action coding guidelines](https://auth0.com/docs/customize/actions/action-coding-guidelines)
