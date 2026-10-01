# The Payment Link's Choices Are Declared Choices

**Date**: October 1, 2026
**Type**: Fix
**Components**: StripePaymentLink, StripeCoupon

## Summary

**A payment link's custom question type and its after-payment behavior are now protobuf `oneof`s.** A custom field is a dropdown, a number or a text answer. After paying, the buyer sees Stripe's confirmation page or is redirected. Both were three or two sibling blocks held to "set exactly one" by a counting rule, which the spec forge flow says becomes a `oneof`. The schema now states the choice itself, as the event destination's receivers already do. Every reader sees one choice, not a rule to work out. A form that writes only what a person authored keeps a chosen-but-empty arm (`numeric: {}`, `hostedConfirmation: {}`) as the choice it is.

**Nothing changes on the wire.** Field names and numbers are unchanged, so manifests, presets, scenarios and the OpenTofu module read exactly as before.

**A coupon's other-currency amounts replace it.** The kind's header and its Ten Off Once preset said they update in place. The field comment and the live proof say Stripe refuses a new amount for a currency a coupon already has, so a change replaces the coupon. Both now say so.

## What Changed

- **StripePaymentLink:**
  - `StripePaymentLinkAfterCompletion.behavior` is a required oneof of `hosted_confirmation` and `redirect`, replacing the `after_completion.one_behavior` rule;
  - `StripePaymentLinkCustomField.type` is a required oneof of `dropdown`, `numeric` and `text`, replacing the `custom_field.one_type` rule;
  - the spec test moves onto the oneofs and pins Stripe's confirmation page with no message;
  - the module's comment on how the spec models Stripe's type discriminators follows.
- **StripeCoupon:** the header comment and the Ten Off Once preset say a change to `currencyOptions` replaces the coupon.
- **Generated:** both kinds' Go stubs, references and the proto-docs index.
