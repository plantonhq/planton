# StripeEventDestination

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeEventDestinationSpec declares where a Stripe account sends its events through Stripe's
v2 event destinations: a webhook URL, an Amazon EventBridge event bus, or an Azure Event Grid
partner topic, with either thin events (a small notification the receiver fetches details
for) or snapshot events (the full object, as classic webhooks deliver it).

When to use which: for classic snapshot webhooks to one URL, a StripeWebhookEndpoint is the
simpler kind; use an event destination for thin events, for events from other accounts or an
organization, or to deliver into AWS or Azure.

The destination's type is the block that is set (webhook_endpoint, amazon_eventbridge or
azure_event_grid); the module sends Stripe the matching type.

A webhook destination's signing secret exists only at creation. The deployment captures it
into status.outputs.signing_secret, where the receiving service reads it by reference; an
imported destination carries none. An EventBridge or Event Grid destination stays pending until
its partner source is associated on the cloud side: status.outputs names the source to
associate.

Destroy deletes the destination. The provider cannot enable or disable a destination, so
status is observed only. A destination deleted outside Planton makes the next plan fail (the
provider does not treat a missing destination as gone); remove it from state and apply again.

The key needs "Event Destinations" write (iac/permissions.yaml).

https://docs.stripe.com/event-destinations
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/v2_core_event_destination

## Example

```yaml
# The canonical example: thin billing events to a webhook route. The lanes
# run against the dedicated test sandbox only; Stripe does not check that the
# URL answers when the destination is created.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeEventDestination
metadata:
  name: billing-thin-events
  org: e2e-org
  env: testing
spec:
  name: Billing thin events
  eventPayload: thin
  enabledEvents:
    - v1.billing.meter.error_report_triggered
    - v1.billing.meter.no_meter_found
  webhookEndpoint:
    url: https://example.com/stripe/thin
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.eventPayload` | `enum` |  |  |  |
| `spec.enabledEvents` | `[]string` | yes |  |  |
| `spec.eventsFrom` | `[]string` |  |  |  |
| `spec.snapshotApiVersion` | `string` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |
| `spec.webhookEndpoint` | `StripeEventDestinationWebhookEndpoint` |  |  |  |
| `spec.webhookEndpoint.url` | `string` | yes |  |  |
| `spec.amazonEventbridge` | `StripeEventDestinationAmazonEventBridge` |  |  |  |
| `spec.amazonEventbridge.awsAccountId` | `string` |  |  |  |
| `spec.amazonEventbridge.awsRegion` | `string` |  |  |  |
| `spec.azureEventGrid` | `StripeEventDestinationAzureEventGrid` |  |  |  |
| `spec.azureEventGrid.azureSubscriptionId` | `string` |  |  |  |
| `spec.azureEventGrid.azureResourceGroupName` | `string` | yes |  |  |
| `spec.azureEventGrid.azureRegion` | `string` | yes |  |  |

## Field Details

### spec.name

`string` · required

name identifies the destination in the Dashboard. It updates in place.

- rule: {"required":true}

### spec.description

`string`

description says what the destination is for. It updates in place.

### spec.eventPayload

`enum`

event_payload is what each event carries. Changing it REPLACES the destination.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `event_payload_unspecified`
- `snapshot` -- snapshot carries the full object, as classic webhooks do.
- `thin` -- thin carries a small notification; the receiver fetches the details it needs.

### spec.enabledEvents

`[]string` · required

enabled_events are the event types delivered, in Stripe's spelling: snapshot events like
"invoice.paid", thin events like "v1.billing.meter.error_report_triggered". List only the events
the receiver handles. They update in place.

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"pattern":"^(\\*|[a-z0-9_]+(\\.[a-z0-9_]+)+)$"}}}}

### spec.eventsFrom

`[]string`

events_from is whose events are routed here: "@self" (this account), "@accounts" (accounts
this account manages), "@organization_members" (accounts linked to its organization), or
"@organization_members/@accounts" (the accounts those members manage). Unset, Stripe routes
this account's own events. Changing it REPLACES the destination.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^@[a-z_]+(/@[a-z_]+)?$"}}}}

### spec.snapshotApiVersion

`string`

snapshot_api_version is the Stripe API version snapshot events are rendered in
("2025-09-30.clover"). Only for event_payload snapshot; unset, the account's default version.
Changing it REPLACES the destination.

- rule: snapshot_api_version is a Stripe API version such as 2025-09-30.clover

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the destination.

### spec.webhookEndpoint

`StripeEventDestinationWebhookEndpoint`

webhook_endpoint is the URL a webhook destination POSTs events to.

### spec.webhookEndpoint.url

`string` · required

url is where Stripe POSTs each event. Live mode requires HTTPS. It updates in place and keeps
the signing secret.

- rule: url is an https:// (or, in a sandbox, http://) address
- rule: {"required":true,"string":{"uri":true}}

### spec.amazonEventbridge

`StripeEventDestinationAmazonEventBridge`

amazon_eventbridge is the AWS account and region an EventBridge destination creates its
partner event source in. Changing it REPLACES the destination.

### spec.amazonEventbridge.awsAccountId

`string`

aws_account_id is the 12-digit AWS account the partner event source appears in.

- rule: {"string":{"pattern":"^[0-9]{12}$"}}

### spec.amazonEventbridge.awsRegion

`string`

aws_region is the AWS region the partner event source appears in ("us-east-1").

- rule: {"string":{"pattern":"^[a-z]{2}(-[a-z]+)+-[0-9]+$"}}

### spec.azureEventGrid

`StripeEventDestinationAzureEventGrid`

azure_event_grid is the Azure subscription, resource group and region an Event Grid
destination creates its partner topic in. Changing it REPLACES the destination.

### spec.azureEventGrid.azureSubscriptionId

`string`

azure_subscription_id is the Azure subscription the partner topic appears in.

- rule: {"string":{"uuid":true}}

### spec.azureEventGrid.azureResourceGroupName

`string` · required

azure_resource_group_name is the resource group the partner topic appears in.

- rule: {"required":true}

### spec.azureEventGrid.azureRegion

`string` · required

azure_region is the Azure region of the partner topic ("eastus").

- rule: {"required":true}

## Validation Rules

- `spec.snapshot_api_version_needs_snapshot`: snapshot_api_version renders snapshot events: set event_payload snapshot, or remove it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeEventDestination, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the destination's Stripe id (ed_...). |
| `status.outputs.status` | `string` | status is "enabled" or "disabled" as Stripe reports it. |
| `status.outputs.status_disabled_reason` | `string` | status_disabled_reason says why a disabled destination is disabled: "user", "no_aws_event_source_exists" or "no_azure_partner_topic_exists". |
| `status.outputs.signing_secret` | `string` (sensitive) | signing_secret is a webhook destination's signing secret, the one the receiving service verifies each delivery's Stripe-Signature header with. Stripe returns it only when the destination is created, so it is empty for an imported destination, and a replacement yields a new one. |
| `status.outputs.aws_event_source_arn` | `string` | aws_event_source_arn is the ARN of the partner event source an EventBridge destination created. |
| `status.outputs.aws_event_source_name` | `string` | aws_event_source_name is that event source's name (aws.partner/stripe.com/...), the value an EventBridge event bus associates with to start receiving events. |
| `status.outputs.aws_event_source_status` | `string` | aws_event_source_status is "pending" until the source is associated with an event bus, then "active". |
| `status.outputs.azure_partner_topic_name` | `string` | azure_partner_topic_name is the partner topic an Event Grid destination created. |
| `status.outputs.azure_partner_topic_status` | `string` | azure_partner_topic_status is "never_activated" until the partner topic is activated in Azure, then "activated". |

## See Also

- [Overview](../README.md)
