# StripeWebhookEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeWebhookEndpointSpec is where the Stripe account the provider connection's key belongs
to delivers its events: a URL and the event types it receives.

Stripe signs every delivery with the endpoint's signing secret, and returns that secret only
when the endpoint is created. The deployment captures it into status.outputs.secret, where the
service that verifies deliveries reads it by reference. An endpoint imported from one made in
the Dashboard carries no secret; replace it to get one.

Destroy deletes the endpoint from Stripe. An endpoint deleted outside Planton makes the next
plan fail (the provider does not treat a missing endpoint as gone); remove it from state and
apply again to recreate it.

The key needs "Webhook Endpoints" write (iac/permissions.yaml).

https://docs.stripe.com/api/webhook_endpoints
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/webhook_endpoint

## Example

```yaml
# The canonical example: an endpoint delivering three billing events. The
# lanes run against the dedicated test sandbox only; Stripe does not check
# that the URL answers when the endpoint is created.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeWebhookEndpoint
metadata:
  name: billing-events
  org: e2e-org
  env: testing
spec:
  url: https://example.com/stripe
  description: Planton catalog example endpoint
  enabledEvents:
    - checkout.session.completed
    - invoice.paid
    - charge.refunded
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.url` | `string` | yes |  |  |
| `spec.enabledEvents` | `[]string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |
| `spec.apiVersion` | `string` |  |  |  |
| `spec.connect` | `bool` |  |  |  |

## Field Details

### spec.url

`string` · required

url is where Stripe sends each event, as an HTTP POST. Live mode requires HTTPS; a sandbox
accepts HTTP for local testing through a tunnel. Changing it updates the endpoint in place
and keeps its signing secret.

- rule: url is an https:// (or, in a sandbox, http://) address
- rule: {"required":true,"string":{"uri":true}}

### spec.enabledEvents

`[]string` · required

enabled_events are the event types delivered to this endpoint, in Stripe's spelling
("checkout.session.completed", "invoice.paid", "charge.refunded"). "*" delivers every event
except those that must be selected by name. List only the events the receiving service
handles: every other delivery is traffic it must acknowledge and ignore.

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"pattern":"^(\\*|[a-z0-9_]+(\\.[a-z0-9_]+)+)$"}}}}

### spec.description

`string`

description is a note shown beside the endpoint in the Dashboard.

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the endpoint. A key removed here is removed
in Stripe on the next apply.

### spec.apiVersion

`string`

api_version pins the Stripe API version events are rendered in for this endpoint
("2025-09-30.clover"). Unset, Stripe renders events in the account's default version. Changing
it REPLACES the endpoint, and a new endpoint gets a new signing secret: the service that
verifies deliveries must take the new secret in the same deployment. Leave it unset unless
the receiving code parses fields that differ between versions.

- rule: api_version is a Stripe API version such as 2025-09-30.clover

### spec.connect

`bool`

connect delivers events from the account's connected accounts (Stripe Connect) instead of
its own. Changing it REPLACES the endpoint, which rotates the signing secret.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeWebhookEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the endpoint's Stripe id (we_...). |
| `status.outputs.secret` | `string` (sensitive) | secret is the signing secret (whsec_...) the receiving service verifies each delivery's Stripe-Signature header with. Stripe returns it only when the endpoint is created, so it is empty for an imported endpoint, and a replacement (api_version or connect changed) yields a new one. |
| `status.outputs.status` | `string` | status is "enabled" or "disabled" as Stripe reports it. Stripe disables an endpoint whose deliveries keep failing; this kind does not toggle it. |
| `status.outputs.url` | `string` | url is the address events are delivered to. |
| `status.outputs.application` | `string` | application is the Connect application that created the endpoint, when one did. |

## See Also

- [Overview](../README.md)
