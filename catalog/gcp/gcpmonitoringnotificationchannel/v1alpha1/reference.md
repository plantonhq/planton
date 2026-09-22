# GcpMonitoringNotificationChannel

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpMonitoringNotificationChannelSpec defines a Cloud Monitoring
notification channel — the delivery endpoint (an email address, a Slack
channel, a PagerDuty service, a webhook, an SMS number, or a Pub/Sub
topic) that alerting policies notify when incidents open, close, or gain
new violations.

A channel is pure configuration: creating one sends nothing on its own.
Alert policies reference the channel by its server-assigned resource name
(the `channel_name` stack output) in their notification_channels list —
that reference is the composition edge charts wire.

Channel behavior is driven by `type` plus type-specific configuration in
`channel_labels` (non-secret settings such as the email address) and
`sensitive_labels` (credentials such as a Slack auth token). GCP validates
the configuration server-side against the chosen type's schema — an
unknown or missing configuration key fails at apply time with the API's
own message.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringNotificationChannel
metadata:
  name: my-sample-channel
spec:
  # GCP project that owns the channel.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Delivery mechanism; GCP validates the value and its configuration
  # keys server-side (email/sms/slack/pagerduty/webhook_tokenauth/
  # webhook_basicauth/pubsub, ...).
  type: email

  # Shown in the console and in notification footers; omit to default to
  # metadata.name.
  displayName: On-call email

  # Who owns the endpoint and why it exists.
  description: Primary paging channel for the platform team

  # Type-specific NON-SECRET configuration. Which keys apply depends on
  # type (email -> email_address; slack -> channel_name; webhook -> url).
  # Credentials belong in sensitiveLabels, never here.
  channelLabels:
    email_address: oncall@example.com

  # Whether notifications are forwarded (default true). Disabled channels
  # keep configuration and references but deliver nothing.
  enabled: true

  # If true, deletion proceeds even while alert policies reference the
  # channel (they lose it in the same operation). Default false — the
  # safe posture.
  forceDelete: false

  # User metadata labels, merged with Planton's platform labels.
  labels:
    team: platform

  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.type` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.channelLabels` | `map<string, string>` |  |  |  |
| `spec.sensitiveLabels` | `GcpMonitoringNotificationChannelSensitiveLabels` |  |  |  |
| `spec.sensitiveLabels.authToken` | `string` (sensitive) |  |  |  |
| `spec.sensitiveLabels.password` | `string` (sensitive) |  |  |  |
| `spec.sensitiveLabels.serviceKey` | `string` (sensitive) |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.forceDelete` | `bool` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the notification channel.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.type

`string` · required

The channel type — which delivery mechanism this channel uses. GCP
validates the value server-side against its live channel-type catalog
(there is no fixed client-side list; new types appear without provider
releases). Common values:
  email              -- channel_labels: email_address
  sms                -- channel_labels: number (E.164, e.g. +15551234567)
  slack              -- channel_labels: channel_name (e.g. #alerts);
                        sensitive_labels: auth_token
  pagerduty          -- sensitive_labels: service_key
  webhook_tokenauth  -- channel_labels: url; the token rides the URL
  webhook_basicauth  -- channel_labels: url, username;
                        sensitive_labels: password
  pubsub             -- channel_labels: topic (projects/{p}/topics/{t})
Immutable in practice for most types: GCP updates a channel's type only
by delete-and-recreate semantics on the provider side.

- rule: {"required":true}

### spec.displayName

`string`

Human-readable name shown in the Cloud Monitoring console and in
notification footers. Defaults to metadata.name when left empty.
Limited to 512 Unicode characters by the API.

- rule: {"string":{"maxLen":"512"}}

### spec.description

`string`

Why this channel exists and who owns the endpoint — write it for the
operator triaging a 3am page. Limited to 1024 bytes by the API.

- rule: {"string":{"maxLen":"1024"}}

### spec.channelLabels

`map<string, string>`

Type-specific, NON-SECRET configuration keys (maps to the provider's
`labels` argument — distinct from the `labels` field below, which is
user metadata). Which keys apply depends on `type`; see the type list
above. GCP rejects unknown keys for the chosen type at apply time.
Credentials (auth_token, password, service_key) are refused here by
validation — they belong in sensitive_labels.

### spec.sensitiveLabels

`GcpMonitoringNotificationChannelSensitiveLabels`

Credentials for channel types that authenticate to an external service.
Each field is a secret: the platform stores it as a managed-secret
reference and resolves it just-in-time at deploy — it never sits in
plaintext in the control plane.

### spec.sensitiveLabels.authToken

`string` · sensitive

OAuth token for the slack channel type (from the Slack app
installation).

### spec.sensitiveLabels.password

`string` · sensitive

HTTP basic-auth password for the webhook_basicauth channel type.

### spec.sensitiveLabels.serviceKey

`string` · sensitive

Integration/service key for the pagerduty channel type (from the
PagerDuty service's integration settings).

### spec.enabled

`bool` · optional (explicit presence)

Whether notifications are forwarded to the described channel (default
true). A disabled channel keeps its configuration and its references
from alert policies but delivers nothing — the safe way to silence an
endpoint temporarily without rewiring policies. Both IaC engines send
the value explicitly so behavior is identical regardless of engine.

- default: `true`

### spec.forceDelete

`bool`

If true, deleting the channel proceeds even when alert policies still
reference it — GCP removes the channel from those policies in the same
operation. If false (default), the delete FAILS while references exist,
which is the safer posture: a dangling policy silently loses its
delivery endpoint.

### spec.labels

`map<string, string>`

User labels attached to the channel for organizing and identifying it
(maps to the provider's user_labels), merged with Planton's platform
labels (which win on key conflicts). Keys and values may contain only
lowercase letters, numerals, underscores, and dashes; keys must begin
with a letter.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the channel is deleted (fails while alert policies still
               reference it unless force_delete is true)
  "PREVENT" -- destroy FAILS; protects the paging path of a production
               alerting setup from accidental teardown
  "ABANDON" -- the channel is removed from management but keeps
               delivering notifications in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `credentials_never_in_channel_labels`: auth_token, password, and service_key are credentials — set them in sensitive_labels, never in channel_labels

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpMonitoringNotificationChannel, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.channel_name` | `string` | The server-assigned resource name of the channel. Format: projects/{project}/notificationChannels/{channel_id} This is THE composition handle: alert policies reference the channel by exactly this value in their notification_channels list. |
| `status.outputs.verification_status` | `string` | Whether the channel has been verified (SMS and email channels require verification before they deliver). One of: VERIFICATION_STATUS_UNSPECIFIED, UNVERIFIED, VERIFIED. Channel types that need no verification report UNSPECIFIED — that is normal, not an error. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBillingBudget | `spec.notifications.monitoringNotificationChannels` | `status.outputs.channel_name` |
| GcpMonitoringAlertPolicy | `spec.notificationChannels` | `status.outputs.channel_name` |
| GcpMonitoringAlertPolicy | `spec.alertStrategy.notificationChannelStrategy[].notificationChannelNames` | `status.outputs.channel_name` |

## See Also

- [Overview](../README.md)
