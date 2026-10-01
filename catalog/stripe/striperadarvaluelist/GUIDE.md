# Stripe Radar Value List Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Radar List Security Notes

### A List Can Let Fraud Through

An allow list is as powerful as a block list: a rule that trusts `@trusted_customers` skips review for everyone on it. Treat changes to allow lists with the same care as a change to the rule itself.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, delete the list and its items | Radar: Write | `POST` and `DELETE` on `/v1/radar/value_lists` and `/v1/radar/value_list_items` |
| Read | Radar: Write | Write includes the read every refresh makes |

## Judgment
## Lists Hold Values, Rules Hold Decisions

Radar rules are written in the Dashboard and name a list as `@alias` -- "Block if :ip_country: in @blocked_countries". The list is what this kind declares; the rule is not declarable. So a policy change that only moves a value is a manifest change, and a change to what happens is a Dashboard change.

## What Changes and Destroy Do

- **Items are their own objects**: adding a value creates an item, removing one deletes it. An item is never edited, only replaced.
- **Stripe lowers the case of string, email and country values**: `KP` is stored as `kp` and `Fraud@Example.COM` as `fraud@example.com`, so Radar matches them without case. The module sends them already lowered, and `item_ids` stays keyed by the value as you wrote it. Values that differ only by case are one item and are refused. A `case_sensitive_string` list, and lists of ids, fingerprints, card BINs or IP addresses, keep the case you send.
- **`itemType` replaces everything**: the list and every item are recreated.
- **Destroy deletes the list and its items**: Stripe refuses while a Radar rule still uses the list. Remove or edit the rule in the Dashboard first.

## Cost

The list costs nothing. Custom rules that reference a list need Radar for Fraud Teams, which Stripe prices per screened transaction. Radar's default rules, included with payments, cannot reference custom lists.

## Importing a List Made in the Dashboard

Import the list by its id (`rsl_...`) and each item by its own id (`rsli_...`); `GET /v1/radar/value_list_items?value_list=<list id>` lists them. Every item the Dashboard holds must be in `items`, or the next apply deletes it.

## Traps

- **Large lists are many objects**: each item is one resource in the module, so a list of thousands plans and applies slowly.
- **A list that cannot be read fails the plan**: the provider does not treat a missing list as gone. A list deleted in the Dashboard takes its items with it, so recover by forgetting both, `tofu state rm stripe_radar_value_list.this stripe_radar_value_list_item.this`, then apply, which creates a new list with every item. Forgetting the list alone leaves the items in state, and the apply fails reading them.
