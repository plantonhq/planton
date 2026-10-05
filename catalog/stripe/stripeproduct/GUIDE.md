# Stripe Product Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Product Security Notes

### Everything Here Is Public

A product's name, description, images and marketing lines are shown to anyone who reaches a Stripe-hosted page that sells it. Put nothing in them you would not put on your pricing page; metadata is not shown to customers.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, archive | Products: Write | `POST` on `/v1/products`; the module never deletes a product |
| Attach and detach features | Entitlements: Write (only with `features`) | `POST` and `DELETE` on `/v1/products/{id}/features` |
| Read | Products: Write | Write includes the read every refresh makes |

## Judgment
## One Owner per Object

Declare a product here only if nothing else creates or edits it. When your application's own code, a billing system, or the Dashboard already owns your price list, leave its products and prices there: two owners overwrite each other on every apply, and neither notices until a customer is charged the wrong amount. Declared products and prices suit a catalog you review like code; a catalog your application generates belongs to your application.

## What Destroy and Changes Do

- **Destroy archives**: the product can no longer be bought, existing subscriptions keep billing, and Stripe keeps it.
- **`type` replaces**: changing between `service` and `good` archives the product and creates a new one; every StripePrice that references it is replaced with it.
- **Features are links**: adding a feature attaches it; removing one deletes the link, and customers subscribed to the product lose that entitlement. An archived feature cannot be attached.
- **A removed setting stays in Stripe**: the provider sends only values that are set, so deleting `description` from the manifest leaves the old one. Change it to the new value instead.

## Importing a Product Made in the Dashboard

Import the product by its id (`prod_...`) and each feature link by `<product id>/<link id>`; the links are listed by `GET /v1/products/<product id>/features`. Once imported, the manifest owns the product: stop editing it in the Dashboard.

## Traps

- **No hidden prices**: the provider can create a price inside a product, but that price would be invisible to Planton, never archived on destroy, and its changes never detected. Declare a StripePrice that names the product instead.
- **Service-only fields**: `statementDescriptor` and `unitLabel` apply only to a service; the spec refuses them on a good.
- **A product that cannot be read fails the plan**: the provider does not treat a missing product as gone. Recover with `tofu state rm stripe_product.this` and an apply, which creates a new one.
