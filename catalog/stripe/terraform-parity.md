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
| Kinds in the catalog | 16 |
| Distinct provider resources consumed | 20 |
| Spec fields authored across all kinds | 323 |
| Module pins on `stripe` | `0.3.0` × 16 |

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

**16 of 16 kinds are at total accounting; 14 proven live.**

| Kind | Provider args | Matched | Mapped | Excluded | Open gaps | Accounted | Proven |
|---|---|---|---|---|---|---|---|
| StripeBillingMeter | 10 | 6 | 2 | 2 | 0 | ✅ | ✅ tofu |
| StripeBillingPortalConfiguration | 7 | 4 | 3 | 0 | 0 | ✅ | — |
| StripeCoupon | 12 | 9 | 2 | 1 | 0 | ✅ | ✅ tofu |
| StripeEntitlementFeature | 3 | 3 | 0 | 0 | 0 | ✅ | ✅ tofu |
| StripeEventDestination | 13 | 10 | 1 | 2 | 0 | ✅ | ✅ tofu |
| StripePaymentLink | 32 | 13 | 19 | 0 | 0 | ✅ | ✅ tofu |
| StripePaymentMethodConfiguration | 62 | 3 | 59 | 0 | 0 | ✅ | ✅ tofu |
| StripePaymentMethodDomain | 2 | 2 | 0 | 0 | 0 | ✅ | ✅ tofu |
| StripePrice | 43 | 31 | 2 | 10 | 0 | ✅ | ✅ tofu |
| StripeProduct | 38 | 16 | 1 | 21 | 0 | ✅ | ✅ tofu |
| StripePromotionCode | 14 | 10 | 2 | 2 | 0 | ✅ | — |
| StripeRadarValueList | 6 | 4 | 1 | 1 | 0 | ✅ | ✅ tofu |
| StripeShippingRate | 15 | 13 | 0 | 2 | 0 | ✅ | ✅ tofu |
| StripeTaxRate | 10 | 10 | 0 | 0 | 0 | ✅ | ✅ tofu |
| StripeTaxRegistration | 4 | 3 | 1 | 0 | 0 | ✅ | ✅ tofu |
| StripeWebhookEndpoint | 6 | 6 | 0 | 0 | 0 | ✅ | ✅ tofu |

## Breadth: every GA resource, one disposition

All resources of `stripe@0.3.0` land in exactly one class:

| Disposition | Resources | Meaning |
|---|---|---|
| Modeled | 19 | consumed by a kind's Terraform module today |
| IAM-covered | 0 | per-resource IAM member/binding/policy triplets, covered by the owning kinds' additive `iam_members` fields |
| Composed | 0 | capability covered through an existing kind's surface rather than a kind of its own |
| Planned | 0 | judged to be covered by a planned kind or planned composition, not built yet |
| Deferred | 27 | deliberately not offered, each with the recorded reason |
| Excluded as deprecated | 3 | deprecated or superseded provider surface |
| **Total** | **49** | |

## The enumerated record

The full per-resource record, so the accounting above is verifiable
rather than trusted.

### Modeled (19)

| Resource | Consuming kinds |
|---|---|
| `stripe_billing_alert` | consumed by StripeBillingMeter |
| `stripe_billing_meter` | consumed by StripeBillingMeter |
| `stripe_billing_portal_configuration` | consumed by StripeBillingPortalConfiguration |
| `stripe_coupon` | consumed by StripeCoupon |
| `stripe_entitlements_feature` | consumed by StripeEntitlementFeature |
| `stripe_payment_link` | consumed by StripePaymentLink |
| `stripe_payment_method_configuration` | consumed by StripePaymentMethodConfiguration |
| `stripe_payment_method_domain` | consumed by StripePaymentMethodDomain |
| `stripe_price` | consumed by StripePrice |
| `stripe_product` | consumed by StripeProduct |
| `stripe_product_feature` | consumed by StripeProduct |
| `stripe_promotion_code` | consumed by StripePromotionCode |
| `stripe_radar_value_list` | consumed by StripeRadarValueList |
| `stripe_radar_value_list_item` | consumed by StripeRadarValueList |
| `stripe_shipping_rate` | consumed by StripeShippingRate |
| `stripe_tax_rate` | consumed by StripeTaxRate |
| `stripe_tax_registration` | consumed by StripeTaxRegistration |
| `stripe_v2_core_event_destination` | consumed by StripeEventDestination |
| `stripe_webhook_endpoint` | consumed by StripeWebhookEndpoint |

### Deferred (27)

| Resource | Recorded reason |
|---|---|
| `stripe_billing_credit_grant` | grants a customer billing credit, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_charge` | applying it charges a card immediately, and destroy refunds nothing; a record of business, not declared configuration (the provider also calls direct charges a legacy flow) |
| `stripe_climate_order` | a purchase paid from the merchant balance, and destroy cancels nothing; money movement, not declared configuration |
| `stripe_credit_note` | can refund money on apply, and destroy voids nothing; a record of business, not declared configuration |
| `stripe_customer` | a customer is a record of business the account's own flows create and own, not declared configuration |
| `stripe_customer_balance_transaction` | a ledger credit or debit applied to a customer's next invoice, and destroy reverses nothing; an event, not declared state |
| `stripe_file` | an uploaded file is an input other objects name by id, never a setting declared for its own sake, and the provider uploads it from a path on the machine running the engine; destroy only removes it from state |
| `stripe_file_link` | a public link to an uploaded file, which is not offered |
| `stripe_invoice` | a bill to a customer that Stripe can finalize and collect automatically; a record of business, not declared configuration |
| `stripe_invoice_item` | adds a charge to a customer's next invoice; a record of business, not declared configuration |
| `stripe_issuing_card` | an issued card that can spend once active, and destroy only removes it from state while the card stays live; money movement, not declared configuration |
| `stripe_issuing_cardholder` | a person or company authorized to spend on issued cards, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_issuing_dispute` | a dispute over a card transaction, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_issuing_personalization_design` | the printed design of issued cards needs an approved Issuing program, and its card logo needs a file of purpose issuing_logo that the provider's file upload cannot create; destroy only removes it from state |
| `stripe_payment_intent` | with confirm set, applying it takes payment, and destroy neither cancels nor refunds it; a record of business, not declared configuration |
| `stripe_payment_method` | a customer's payment instrument, and destroy does not detach it; a record of business, not declared configuration |
| `stripe_person` | an identity record (with personal data) of a Connect account's representative or owner; a record of business, not declared configuration |
| `stripe_quote` | a sales offer to a customer, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_setup_intent` | saves a customer's payment credentials, and destroy only removes it from state; a record of business, not declared configuration |
| `stripe_subscription` | bills a customer on a schedule, and destroy only removes it from state while Stripe keeps billing; a record of business, not declared configuration |
| `stripe_subscription_item` | part of a subscription, which is not offered; changing one changes what a customer is billed |
| `stripe_subscription_schedule` | future billing changes for a customer, and destroy only removes it from state while the schedule keeps running; a record of business |
| `stripe_tax_id` | part of a customer record, which is not offered |
| `stripe_terminal_configuration` | in-store reader behavior (tipping, splash screens, offline mode, Wi-Fi); it serves only physical checkout fleets, and its images and certificates are uploaded files, which are not offered either |
| `stripe_terminal_location` | a store's address that in-person readers register to; it serves only businesses with physical checkout, which configure their shop fleet in the Dashboard or their own point-of-sale software |
| `stripe_terminal_reader` | enrolls a physical card reader with a code the device shows once and the provider never stores, so the declaration can never be applied again or recreated without a person at the reader; hardware enrollment, not declared configuration |
| `stripe_treasury_financial_account` | an account that holds money, and destroy does not close it; money movement, not declared configuration |

### Excluded as deprecated (3)

| Resource | Recorded reason |
|---|---|
| `stripe_apple_pay_domain` | Stripe's Apple Pay guide registers domains as payment method domains, which cover Apple Pay with every other wallet method (StripePaymentMethodDomain); no guide uses the Apple Pay domain endpoint, though Stripe's API spec does not flag it |
| `stripe_plan` | the provider's own description: the Prices API replaces the Plans API; StripePrice covers recurring prices |
| `stripe_source` | the provider's own description: Stripe does not recommend the deprecated Sources API; payment methods replace it |
