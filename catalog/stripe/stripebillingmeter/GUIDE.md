# Stripe Billing Meter Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API is HTTPS-only, including the meter-event calls your application makes
- **Usage data**: your application sends it to Stripe with its own key; this component stores none of it

## Billing Meter Security Notes

### The Event Name Is Not a Secret

The event name tells Stripe which meter an event counts toward. Reporting usage needs a Stripe key with write access to meter events, which belongs to your application, not to this component. Keep that key in your secrets manager, never beside the event name.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, deactivate the meter | Billing Meters: Write | `POST` on `/v1/billing/meters` and `/v1/billing/meters/{id}/deactivate`; there is no delete |
| Create an alert | Alerts: Write (label proven on the first least-privilege run) | `POST` on `/v1/billing/alerts`; there is no update or delete |
| Read | Write on each | Write includes the read every refresh makes |

## Judgment
## One Owner per Object

Declare a meter here only if nothing else creates or edits it. A meter your application creates per tenant belongs to your application; the handful of meters your pricing is built on suit a chart you review like code.

## How Your Application Learns the Event Name

Declare the meter, the metered StripePrice that names it, and your application's deployment in one chart. The deployment reads the event name from the meter's `status.outputs.event_name` by reference, into an environment variable. When a change replaces the meter, the price and the deployment follow on the same apply, so the name your application sends and the name the meter counts are always the same.

## Changing a Meter

Only `displayName` changes in place. Changing `eventName`, `defaultAggregation`, `customerMapping`, `valueSettings` or `eventTimeWindow` replaces the meter:

1. Planton deactivates the old meter; it stops accepting events.
2. Planton creates the new meter, with a new id.
3. Every metered StripePrice that references the meter is replaced on the same apply, because a price's meter can never change. Subscriptions on the old price keep billing what the old meter counted until you move them.

## Alerts Are Forgotten, Not Deleted

Stripe's alerts can't be updated, and the provider has no delete for them:

- **Changing an alert** (its threshold, customer or title) creates a new alert. The old one stays active in Stripe.
- **Removing an alert**, or destroying the meter, leaves it active in Stripe.

Archive a stale alert in the Dashboard, or with Stripe's archive call (`POST /v1/billing/alerts/{id}/archive`). The `alert_ids` output lists every alert this meter declared.

## What Destroy Does

Destroy deactivates the meter; Stripe never deletes one. It stops accepting events and keeps the usage it recorded. Its alerts stay active.

## Importing a Meter Made in the Dashboard

Import the meter by its id (`mtr_...`), and each alert by its own id. Every field marked as replacing must match the Dashboard's meter exactly, or the first apply replaces it.

## Traps

- **A metered price needs a meter**: a StripePrice with `usageType: metered` and no meter is refused before anything runs.
- **Stale alerts keep firing**: after you change a threshold, both the old and the new alert send events until you archive the old one.
- **A meter that cannot be read fails the plan**: the provider does not treat a missing meter as gone. Recover with `tofu state rm stripe_billing_meter.this` and an apply, which creates a new one.
