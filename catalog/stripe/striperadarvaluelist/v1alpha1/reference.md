# StripeRadarValueList

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

StripeRadarValueListSpec declares a Radar list -- blocked countries, trusted customers, known
fraudulent emails -- and every item in it. Radar rules name the list by its alias
("Block if :ip_country: in @blocked_countries"), so the list is where the values live and the
rule is where the decision lives.

Items are validated against the list's item type before Stripe sees them: a country list takes
two-letter codes, an email list takes email addresses. Each item is its own object in Stripe:
adding an item creates it, removing one deletes it, and an item can never be edited, only
replaced. Changing item_type REPLACES the list and every item in it.

Destroy deletes the list and its items. Stripe refuses to delete a list a Radar rule still uses,
and rules are not declarable here: remove the rule in the Dashboard first. A list deleted outside
Planton makes the next plan fail (the provider does not treat a missing list as gone); remove it
from state and apply again.

The key needs "Radar" write (iac/permissions.yaml). A list costs nothing, but writing custom
rules that use it needs Radar for Fraud Teams, which Stripe prices per screened transaction.

https://docs.stripe.com/radar/lists
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/radar_value_list

## Example

```yaml
# The canonical example: a list of countries a Radar rule blocks.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeRadarValueList
metadata:
  name: blocked-countries
  org: e2e-org
  env: testing
spec:
  alias: blocked_countries
  name: Blocked countries
  itemType: country
  items:
    - KP
    - IR
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.alias` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.itemType` | `enum` |  |  |  |
| `spec.items` | `[]string` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.alias

`string` · required

alias is the name Radar rules reference the list by, after an @ ("blocked_countries"). It
updates in place, and every rule that names the old alias must be changed with it.

- rule: {"required":true}

### spec.name

`string` · required

name is the list's human-readable name in the Dashboard. It updates in place.

- rule: {"required":true}

### spec.itemType

`enum`

item_type is what the list holds. Unset, Stripe uses string (any text, compared without case).
Changing it REPLACES the list and every item.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `item_type_unspecified`
- `account` -- account holds connected account ids (acct_...).
- `card_bin` -- card_bin holds card number prefixes.
- `card_fingerprint` -- card_fingerprint holds Stripe card fingerprints.
- `case_sensitive_string` -- case_sensitive_string holds text compared with case.
- `country` -- country holds two-letter ISO country codes.
- `crypto_fingerprint` -- crypto_fingerprint holds Stripe crypto wallet fingerprints.
- `customer_id` -- customer_id holds customer ids (cus_...).
- `email` -- email holds email addresses.
- `ip_address` -- ip_address holds IP addresses.
- `sepa_debit_fingerprint` -- sepa_debit_fingerprint holds Stripe SEPA Debit fingerprints.
- `string` -- string holds any text, compared without case (Stripe's default).
- `us_bank_account_fingerprint` -- us_bank_account_fingerprint holds Stripe US bank account fingerprints.

### spec.items

`[]string`

items are the list's values, each once. Adding an item creates it in Stripe and removing one
deletes it.

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the list. A key removed here is removed in
Stripe on the next apply; removing the whole map leaves Stripe's keys in place.

## Validation Rules

- `spec.items.country`: a country list holds two-letter ISO country codes in capitals, such as US or DE
- `spec.items.email`: an email list holds email addresses, such as fraud@example.com
- `spec.items.ip_address`: an ip_address list holds IPv4 or IPv6 addresses, such as 203.0.113.7
- `spec.items.card_bin`: a card_bin list holds card BINs: the card number's first 6 to 8 digits
- `spec.items.customer_id`: a customer_id list holds Stripe customer ids (cus_...)
- `spec.items.account`: an account list holds Stripe account ids (acct_...)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeRadarValueList, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the list's Stripe id (rsl_...). |
| `status.outputs.alias` | `string` | alias is the name Radar rules reference the list by. |
| `status.outputs.item_ids` | `map<string, string>` | item_ids maps each item's value to its Stripe id (rsli_...), the id Stripe needs to import or delete that item. |

## See Also

- [Overview](../README.md)
