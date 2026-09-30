---
title: "Terraform Parity"
description: "Measured parity of the Stripe catalog against the pinned Terraform provider"
icon: "check-circle"
order: 90
---

<!-- GENERATED FILE -- DO NOT EDIT.
     Rendered from the committed provider schemas, the kind registry, the
     Terraform modules, the per-kind provider-parity manifests, the
     dispositions ledger, and the E2E profiles.
     parameters: provider=stripe ga-schema=stripe
     Regenerate: make generate-provider-parity-report -->

# Stripe Terraform Parity

This catalog is **built for 100% Terraform parity**: every configurable
argument of the pinned Terraform provider is representable through a kind,
and every provider resource carries exactly one recorded disposition --
omission is a decision, never an accident. This page is the measurement,
generated from the same accounting that gates the repository's CI. It makes
no achieved-parity claim: a kind counts as PROVEN only when live end-to-end
runs pass on every IaC engine it runs on, and the tables below show exactly
how far that has progressed.

## Measurement baseline

| | |
|---|---|
| Provider schema (parity baseline) | `stripe@0.3.0` |
| Kinds in the catalog | 9 |
| Distinct provider resources consumed | 11 |
| Spec fields authored across all kinds | 175 |
| Module pins on `stripe` | `0.3.0` × 9 |

The GA provider is the parity baseline. Capability that exists only in a
secondary channel (for Google, the `google-beta` provider) enters per kind
through the admission list (`pkg/providerparity/admissions/`), never
wholesale: one entry per resource per kind, with the reason and where its
promotion to the baseline is tracked. The accounting reads the list -- an
admitted resource is measured against its channel's schema, an unadmitted
secondary-channel resource is a finding, and an admitted resource the
baseline serves at the pin is a stale admission.

## The provider block

Parity covers the provider's own configuration block -- credentials,
role-assumption chains, default tags, endpoint overrides, retry tuning --
under the same total-accounting rule as resources: every configurable,
non-deprecated provider-block argument is matched to a provider-config
field, mapped by recorded judgment, owned by the modules by recorded
judgment, or excluded with a recorded reason -- and arguments set inside
catalog modules' own provider blocks must carry that judgment too.

| Provider-block args | Matched | Mapped | Module-owned | Excluded | Open gaps | Accounted |
|---|---|---|---|---|---|---|
| 2 | 2 | 0 | 0 | 0 | 0 | ✅ |

## Depth: per-kind accounting

Every configurable, non-deprecated provider argument of a kind's consumed
resources must be matched to a spec field, mapped by recorded judgment, or
excluded with a recorded reason -- and every spec field must reach provider
surface. **Accounted** means both directions hold with zero unexplained
gaps. **Proven** means live end-to-end runs passed on every IaC engine the kind runs on.

**9 of 9 kinds are at total accounting; 0 proven live.**

| Kind | Provider args | Matched | Mapped | Excluded | Open gaps | Accounted | Proven |
|---|---|---|---|---|---|---|---|
| StripeBillingPortalConfiguration | 7 | 4 | 3 | 0 | 0 | ✅ | — |
| StripeEntitlementFeature | 3 | 3 | 0 | 0 | 0 | ✅ | — |
| StripeEventDestination | 13 | 10 | 1 | 2 | 0 | ✅ | — |
| StripePaymentMethodConfiguration | 62 | 3 | 59 | 0 | 0 | ✅ | — |
| StripePaymentMethodDomain | 2 | 2 | 0 | 0 | 0 | ✅ | — |
| StripePrice | 43 | 31 | 2 | 10 | 0 | ✅ | — |
| StripeProduct | 38 | 16 | 1 | 21 | 0 | ✅ | — |
| StripeRadarValueList | 6 | 4 | 1 | 1 | 0 | ✅ | — |
| StripeWebhookEndpoint | 6 | 6 | 0 | 0 | 0 | ✅ | — |

## Breadth: every GA resource, one disposition

All resources of `stripe@0.3.0` land in exactly one class:

| Disposition | Resources | Meaning |
|---|---|---|
| Modeled | 11 | consumed by a kind's Terraform module today |
| IAM-covered | 0 | per-resource IAM member/binding/policy triplets, covered by the owning kinds' additive `iam_members` fields |
| Composed | 0 | capability covered through an existing kind's surface rather than a kind of its own |
| Planned | 14 | judged to be covered by a planned kind or planned composition, not built yet |
| Deferred | 21 | deliberately not offered, each with the recorded reason |
| Excluded as deprecated | 3 | deprecated or superseded provider surface |
| **Total** | **49** | |

## The enumerated record

The full per-resource record, so the accounting above is verifiable
rather than trusted.

### Modeled (11)

| Resource | Consuming kinds |
|---|---|
| `stripe_billing_portal_configuration` | consumed by StripeBillingPortalConfiguration |
| `stripe_entitlements_feature` | consumed by StripeEntitlementFeature |
| `stripe_payment_method_configuration` | consumed by StripePaymentMethodConfiguration |
| `stripe_payment_method_domain` | consumed by StripePaymentMethodDomain |
| `stripe_price` | consumed by StripePrice |
| `stripe_product` | consumed by StripeProduct |
| `stripe_product_feature` | consumed by StripeProduct |
| `stripe_radar_value_list` | consumed by StripeRadarValueList |
| `stripe_radar_value_list_item` | consumed by StripeRadarValueList |
| `stripe_v2_core_event_destination` | consumed by StripeEventDestination |
| `stripe_webhook_endpoint` | consumed by StripeWebhookEndpoint |

### Planned (14)

| Resource | Recorded reason |
|---|---|
| `stripe_billing_alert` | judged as a planned composition into StripeBillingMeter (a usage threshold alert keyed by its meter, cannot be updated; removing one only removes it from state and Stripe keeps it active) |
| `stripe_billing_meter` | judged as a planned StripeBillingMeter kind (how usage events are aggregated for usage-based prices), carrying its alerts; destroy deactivates it |
| `stripe_coupon` | judged as a planned StripeCoupon kind (a reusable discount: an amount or percentage off, for a duration) |
| `stripe_file` | judged as a planned StripeFile kind (an uploaded file other objects reference by id: a business logo, a card design image, dispute evidence); destroy only removes it from state |
| `stripe_file_link` | judged as a planned composition into StripeFile (a public link keyed by its file, dropped from state with it) |
| `stripe_issuing_personalization_design` | judged as a planned StripeIssuingPersonalizationDesign kind (the printed design of physical issued cards: card logo, carrier text, physical bundle); destroy only removes it from state |
| `stripe_payment_link` | judged as a planned StripePaymentLink kind (a hosted page selling fixed prices; applying it moves no money, but it is a live URL that takes payments); destroy deactivates it |
| `stripe_promotion_code` | judged as a planned StripePromotionCode kind (a customer-facing code for a coupon, with its own active state, redemption limits and expiry; campaigns mint codes over a coupon's life); destroy deactivates it |
| `stripe_shipping_rate` | judged as a planned StripeShippingRate kind (a shipping option offered at checkout); destroy deactivates it |
| `stripe_tax_rate` | judged as a planned StripeTaxRate kind (a manual tax rate; its percentage and inclusiveness cannot change, so a change is a new rate); destroy deactivates it |
| `stripe_tax_registration` | judged as a planned StripeTaxRegistration kind (where the account is registered to collect tax, for Stripe Tax); destroy only removes it from state, and Stripe keeps collecting |
| `stripe_terminal_configuration` | judged as a planned StripeTerminalConfiguration kind (reader behavior per device model: tipping, splash screens, offline mode) |
| `stripe_terminal_location` | judged as a planned StripeTerminalLocation kind (a physical place readers are registered to) |
| `stripe_terminal_reader` | judged as a planned StripeTerminalReader kind (a card reader registered with a single-use registration code, so a replacement needs a new code) |

### Deferred (21)

| Resource | Recorded reason |
|---|---|
| `stripe_billing_credit_grant` | grants a customer billing credit, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_charge` | applying it charges a card immediately, and destroy refunds nothing; a record of business, not declared configuration (the provider also calls direct charges a legacy flow) |
| `stripe_climate_order` | a purchase paid from the merchant balance, and destroy cancels nothing; money movement, not declared configuration |
| `stripe_credit_note` | can refund money on apply, and destroy voids nothing; a record of business, not declared configuration |
| `stripe_customer` | a customer is a record of business the account's own flows create and own, not declared configuration |
| `stripe_customer_balance_transaction` | a ledger credit or debit applied to a customer's next invoice, and destroy reverses nothing; an event, not declared state |
| `stripe_invoice` | a bill to a customer that Stripe can finalize and collect automatically; a record of business, not declared configuration |
| `stripe_invoice_item` | adds a charge to a customer's next invoice; a record of business, not declared configuration |
| `stripe_issuing_card` | an issued card that can spend once active, and destroy only removes it from state while the card stays live; money movement, not declared configuration |
| `stripe_issuing_cardholder` | a person or company authorized to spend on issued cards, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_issuing_dispute` | a dispute over a card transaction, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_payment_intent` | with confirm set, applying it takes payment, and destroy neither cancels nor refunds it; a record of business, not declared configuration |
| `stripe_payment_method` | a customer's payment instrument, and destroy does not detach it; a record of business, not declared configuration |
| `stripe_person` | an identity record (with personal data) of a Connect account's representative or owner; a record of business, not declared configuration |
| `stripe_quote` | a sales offer to a customer, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_setup_intent` | saves a customer's payment credentials, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_subscription` | bills a customer on a schedule, and destroy only removes it from state while Stripe keeps billing; a record of business, not declared configuration |
| `stripe_subscription_item` | part of a subscription, which is not offered; changing one changes what a customer is billed |
| `stripe_subscription_schedule` | future billing changes for a customer, and destroy only removes it from state while the schedule keeps running; a record of business |
| `stripe_tax_id` | part of a customer record, which is not offered |
| `stripe_treasury_financial_account` | an account that holds money, and destroy does not close it; money movement, not declared configuration |

### Excluded as deprecated (3)

| Resource | Recorded reason |
|---|---|
| `stripe_apple_pay_domain` | Stripe's Apple Pay guide registers domains as payment method domains, which cover Apple Pay with every other wallet method (StripePaymentMethodDomain); no guide uses the Apple Pay domain endpoint, though Stripe's API spec does not flag it |
| `stripe_plan` | the provider's own description: the Prices API replaces the Plans API; StripePrice covers recurring prices |
| `stripe_source` | the provider's own description: Stripe does not recommend the deprecated Sources API; payment methods replace it |
