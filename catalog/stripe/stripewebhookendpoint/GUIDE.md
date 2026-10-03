# Stripe Webhook Endpoint Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Webhook Endpoint Security Notes

### Verify the Signature, Always

Anyone can POST to a public URL. Only deliveries whose `Stripe-Signature` header verifies against `status.outputs.secret` came from Stripe; reject the rest before reading the body.

### The Secret Moves by Reference

The secret exists only in the create response. The deployment captures it into a sensitive output, and the receiving service reads it from there by reference. Nobody copies it from the Dashboard, and nobody pastes it into a record.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, delete | Webhook Endpoints: Write | `POST` and `DELETE` on `/v1/webhook_endpoints` |
| Read | Webhook Endpoints: Write | Write includes the read every refresh makes |

## Judgment
## How the Secret Behaves

- **Only at creation**: Stripe returns the signing secret in the create response and never again. The provider keeps it in state from then on.
- **Imported endpoints have none**: an endpoint made in the Dashboard and imported reports an empty secret. Replace it (a new endpoint, the same URL and events) to get one.
- **Replacement rotates it**: changing `apiVersion` or `connect` replaces the endpoint, and the new one has a new secret. The receiving service must take the new secret in the same deployment, or it rejects every delivery.

## Moving From a Hand-Made Endpoint Without Losing Events

1. Declare a new endpoint with the same URL and events. While both exist, Stripe delivers every event to both.
2. Point the receiving service at the new endpoint's `status.outputs.secret` and redeploy it. It now accepts the new endpoint's deliveries and rejects the old one's, which Stripe retries.
3. Delete the hand-made endpoint in the Dashboard.

Handle each event idempotently by its Stripe event id, so an event delivered twice during the switch is acted on once.

## Traps

- **A missing endpoint fails the plan**: the provider does not treat an endpoint deleted outside Planton as gone. Recover with `tofu state rm stripe_webhook_endpoint.this` and an apply, which creates a new endpoint with a new secret.
- **`status` is observed, not set**: Stripe disables an endpoint whose deliveries keep failing, and this kind cannot turn it back on. Fix the route; re-enable the endpoint in the Dashboard.
- **Live mode needs HTTPS**: a sandbox accepts `http://` for local testing through a tunnel; live mode refuses it.
