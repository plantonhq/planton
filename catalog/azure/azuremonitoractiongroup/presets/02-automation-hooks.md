# Automation Hooks (Webhook + Function)

This preset creates a machine-only action group: an Entra-authenticated webhook into an incident platform and an HTTP-triggered Azure Function for programmatic remediation. Alert rules typically reference this group alongside an on-call group.

## When to Use

- Incident platforms (PagerDuty, Opsgenie, custom) consuming Azure alerts by webhook
- Auto-remediation flows (restart, scale, failover) triggered by alerts

## Key Configuration Choices

- **Entra-authenticated webhook** (`aadAuth`) -- the keyless posture: the call authenticates as an Entra application instead of a secret baked into the URL; configure the receiving app to accept the token audience first
- **Common alert schema everywhere** -- machines should always parse the one consistent payload
- **Function by FK** -- `functionAppResourceId` resolves from the `AzureFunctionApp` output when composed in an Infra Chart

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `my-observability-rg` | Resource group holding the group | `AzureResourceGroup` status outputs |
| `11111111-2222-3333-4444-555555555555` (objectId) | The Entra application the webhook authenticates as | Entra ID -> App registrations -> the app's Object ID |
| `my-ops-functions` / `HandleAlert` | The remediation function | `AzureFunctionApp` status outputs + your function name |

## Related Presets

- **01-oncall-team** -- The human channel to pair with automation
- **03-role-fanout** -- Role-based notification without address lists
