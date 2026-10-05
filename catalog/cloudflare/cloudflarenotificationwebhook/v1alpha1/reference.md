# CloudflareNotificationWebhook

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareNotificationWebhookSpec registers a webhook destination for
notification policies: an HTTPS endpoint (Slack, Google Chat, Datadog,
Discord, Feishu, Opsgenie, Splunk, or any generic receiver) that
CloudflareNotificationPolicy resources deliver alerts to. Cloudflare
infers the destination's type from the URL and reports it as an output.
A plain CRUD object -- real create, update, delete.

## Example

```yaml
# Complete example manifest for CloudflareNotificationWebhook.
# Registers a Slack incoming-webhook URL as an alert destination.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareNotificationWebhook
metadata:
  name: ops-slack
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: ops-slack
  url: "https://hooks.slack.com/services/REPLACE/WITH/WEBHOOK"
  secret:
    value: "REPLACE_WITH_SHARED_SECRET"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.url` | `string` | yes |  |  |
| `spec.secret` | `string \| valueFrom` (sensitive) |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the webhook destination belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.name

`string` · required

The destination's name, shown in the dashboard's destinations list.

- rule: {"required":true}

### spec.url

`string` · required

The endpoint URL alerts are POSTed to (for Slack/Google Chat/Discord
and kin, the integration's incoming-webhook URL; for Datadog/Splunk/
Opsgenie, the vendor's intake endpoint). Cloudflare sends a TEST POST
to this URL at registration and rejects the create (HTTP 422) unless
the endpoint answers 2xx -- the receiver must be live and accepting
POSTs BEFORE this resource deploys; a GET-only or not-yet-deployed
endpoint can never register.

- rule: {"required":true}

### spec.secret

`string | valueFrom` · sensitive

Optional shared secret sent with each delivery so the receiver can
authenticate Cloudflare (for Datadog/Splunk/Opsgenie this carries the
vendor API key). WRITE-ONLY: Cloudflare never returns it in any API
response -- it cannot be read back, imported, or drift-detected.
Provide a managed-secret reference; the platform resolves it
just-in-time at deploy.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareNotificationWebhook, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.webhook_id` | `string` | The Cloudflare-assigned UUID of the webhook destination (what notification policies reference). |
| `status.outputs.type` | `string` | The destination type Cloudflare inferred from the URL: one of datadog, discord, feishu, gchat, generic, opsgenie, slack, splunk. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareNotificationPolicy | `spec.mechanisms.webhookIds` | `status.outputs.webhook_id` |

## See Also

- [Overview](../README.md)
