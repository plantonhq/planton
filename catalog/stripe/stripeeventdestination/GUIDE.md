# Stripe Event Destination Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Event Destination Security Notes

### Verify the Signature, Always

A webhook destination's URL is public. Only deliveries whose `Stripe-Signature` header verifies against `status.outputs.signing_secret` came from Stripe; reject the rest before reading the body. Thin events carry no object data, but a forged one can still trigger a fetch.

### Only the Named Cloud Account Can Receive

An EventBridge destination creates a partner event source in exactly the AWS account and region you name, and an Event Grid destination a partner topic in exactly the subscription and resource group you name. Nobody else can associate them.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, delete | Event Destinations: Write | `POST` and `DELETE` on `/v2/core/event_destinations` |
| Read | Event Destinations: Write | Write includes the read every refresh makes |

## Judgment
## Webhook Endpoint or Event Destination

Both deliver events to a URL. Use a StripeWebhookEndpoint for classic snapshot events to one route -- it is the simpler object, and most applications need nothing more. Use an event destination for thin events (the way Stripe's v2 APIs report), for events from accounts you manage or your organization's members, or to deliver straight into Amazon EventBridge or Azure Event Grid.

## How the Signing Secret Behaves

- **Only at creation**: Stripe returns a webhook destination's signing secret in the create response and never again. The provider keeps it in state from then on.
- **Imported destinations have none**: replace an imported destination to get one.
- **Replacement rotates it**: changing `eventPayload`, `eventsFrom`, `snapshotApiVersion` or the destination block replaces the destination, and the new one has a new secret. The receiving service must take it in the same deployment.

## EventBridge and Event Grid

1. Apply the destination. Stripe creates the partner source; its status is `pending` (EventBridge) or `never_activated` (Event Grid).
2. On the cloud side, associate an event bus with `status.outputs.aws_event_source_name`, or activate the partner topic named by `status.outputs.azure_partner_topic_name`.
3. The source turns `active` or `activated`, and events flow.

## Traps

- **Importing an EventBridge destination replaces it**: Stripe's API does not return the AWS region, so an imported EventBridge destination has none in state, and its first apply replaces it -- with a new partner source the bus must associate again.
- **`status` is observed, not set**: the provider cannot enable or disable a destination. A destination Stripe disabled (`status_disabled_reason`) is re-enabled in the Dashboard after fixing the cause.
- **Removing a value does not clear it**: the provider sends only values that are set, so deleting `description` leaves the old one.
- **A destination that cannot be read fails the plan**: the provider does not treat a missing destination as gone. Recover with `tofu state rm stripe_v2_core_event_destination.this` and an apply, which creates a new one.
