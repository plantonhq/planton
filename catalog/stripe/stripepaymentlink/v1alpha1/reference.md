# StripePaymentLink

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

StripePaymentLinkSpec declares a Stripe-hosted payment page at a public address: the prices it
sells, what it collects from the buyer, and where the buyer goes after paying.

The link takes payments from the moment it is applied: anyone with its address can pay for the
declared prices. Applying it moves no money and creates no customer; buyers do both, later.

A link can never change which prices it sells, and the provider never stores a line item's
quantity. Changing the line items' prices or quantities (their order included), the currency, the Connect fields, consent collection, managed payments, the payment's capture
method or future-usage setting, the subscription description or the shipping options REPLACES
the link: Planton creates the new link first, then deactivates the old one. The new link has a
new address, so read status.outputs.url by reference wherever the address is used; visitors to
the old address see inactive_message. A referenced price that is itself replaced replaces the
link too. Adjustable quantities, optional items and every other field update in place.

Validation can't see the price behind a reference, so three of Stripe's rules are checked at
apply, not here: application_fee_amount needs a link with no recurring prices,
application_fee_percent needs one with a recurring price, and payment_method_collection and
subscription_data need a subscription (a recurring price).

One owner per object: declare a link here only if nothing else creates or edits it. When the
application's own code or the Dashboard owns the account's payment links, leave them there.

Stripe never deletes a payment link. Destroy deactivates it (active = false): its address shows
inactive_message, and payments already taken stay. A link that cannot be read makes the plan
fail (the provider does not treat a missing link as gone); remove it from state and apply
again.

The key needs "Payment Links" write (iac/permissions.yaml).

https://docs.stripe.com/payment-links
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/payment_link

## Example

```yaml
# The canonical example: a subscription page for one price with a 14-day
# trial, promotion codes, and a redirect to the account's site after payment.
# The price is a literal id here so the manifest plans on its own; a real
# manifest references its StripePrice (see the scenarios).
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentLink
metadata:
  name: pro-monthly-link
  org: e2e-org
  env: testing
spec:
  lineItems:
    - price:
        value: price_pro_monthly
      quantity: 1
      adjustableQuantity:
        enabled: true
        minimum: 1
        maximum: 50
  allowPromotionCodes: true
  afterCompletion:
    redirect:
      url: https://example.com/welcome?session={CHECKOUT_SESSION_ID}
  subscriptionData:
    trialPeriodDays: 14
    trialSettings:
      endBehavior:
        missingPaymentMethod: cancel
  paymentMethodCollection: if_required
  inactiveMessage: This offer has ended.
  metadata:
    campaign: launch
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.lineItems` | `[]StripePaymentLinkLineItem` | yes |  |  |
| `spec.lineItems[].price` | `string \| valueFrom` | yes |  | StripePrice (`status.outputs.id`) |
| `spec.lineItems[].quantity` | `int64` |  |  |  |
| `spec.lineItems[].adjustableQuantity` | `StripePaymentLinkAdjustableQuantity` |  |  |  |
| `spec.lineItems[].adjustableQuantity.enabled` | `bool` |  |  |  |
| `spec.lineItems[].adjustableQuantity.minimum` | `int64` |  |  |  |
| `spec.lineItems[].adjustableQuantity.maximum` | `int64` |  |  |  |
| `spec.optionalItems` | `[]StripePaymentLinkOptionalItem` |  |  |  |
| `spec.optionalItems[].price` | `string \| valueFrom` | yes |  | StripePrice (`status.outputs.id`) |
| `spec.optionalItems[].quantity` | `int64` |  |  |  |
| `spec.optionalItems[].adjustableQuantity` | `StripePaymentLinkAdjustableQuantity` |  |  |  |
| `spec.optionalItems[].adjustableQuantity.enabled` | `bool` |  |  |  |
| `spec.optionalItems[].adjustableQuantity.minimum` | `int64` |  |  |  |
| `spec.optionalItems[].adjustableQuantity.maximum` | `int64` |  |  |  |
| `spec.active` | `bool` |  | `true` |  |
| `spec.inactiveMessage` | `string` |  |  |  |
| `spec.afterCompletion` | `StripePaymentLinkAfterCompletion` |  |  |  |
| `spec.afterCompletion.hostedConfirmation` | `StripePaymentLinkHostedConfirmation` |  |  |  |
| `spec.afterCompletion.hostedConfirmation.customMessage` | `string` |  |  |  |
| `spec.afterCompletion.redirect` | `StripePaymentLinkRedirect` |  |  |  |
| `spec.afterCompletion.redirect.url` | `string` | yes |  |  |
| `spec.allowPromotionCodes` | `bool` |  |  |  |
| `spec.automaticTax` | `StripePaymentLinkAutomaticTax` |  |  |  |
| `spec.automaticTax.enabled` | `bool` |  |  |  |
| `spec.automaticTax.liability` | `StripePaymentLinkAccountRef` |  |  |  |
| `spec.automaticTax.liability.account` | `string` |  |  |  |
| `spec.billingAddressCollection` | `enum` |  |  |  |
| `spec.consentCollection` | `StripePaymentLinkConsentCollection` |  |  |  |
| `spec.consentCollection.paymentMethodReuseAgreement` | `StripePaymentLinkPaymentMethodReuseAgreement` |  |  |  |
| `spec.consentCollection.paymentMethodReuseAgreement.position` | `enum` |  |  |  |
| `spec.consentCollection.promotions` | `enum` |  |  |  |
| `spec.consentCollection.termsOfService` | `enum` |  |  |  |
| `spec.currency` | `string` |  |  |  |
| `spec.customFields` | `[]StripePaymentLinkCustomField` |  |  |  |
| `spec.customFields[].key` | `string` | yes |  |  |
| `spec.customFields[].label` | `string` | yes |  |  |
| `spec.customFields[].optional` | `bool` |  |  |  |
| `spec.customFields[].dropdown` | `StripePaymentLinkDropdown` |  |  |  |
| `spec.customFields[].dropdown.options` | `[]StripePaymentLinkDropdownOption` | yes |  |  |
| `spec.customFields[].dropdown.options[].label` | `string` | yes |  |  |
| `spec.customFields[].dropdown.options[].value` | `string` | yes |  |  |
| `spec.customFields[].dropdown.defaultValue` | `string` |  |  |  |
| `spec.customFields[].numeric` | `StripePaymentLinkTextBounds` |  |  |  |
| `spec.customFields[].numeric.defaultValue` | `string` |  |  |  |
| `spec.customFields[].numeric.minimumLength` | `int64` |  |  |  |
| `spec.customFields[].numeric.maximumLength` | `int64` |  |  |  |
| `spec.customFields[].text` | `StripePaymentLinkTextBounds` |  |  |  |
| `spec.customFields[].text.defaultValue` | `string` |  |  |  |
| `spec.customFields[].text.minimumLength` | `int64` |  |  |  |
| `spec.customFields[].text.maximumLength` | `int64` |  |  |  |
| `spec.customText` | `StripePaymentLinkCustomText` |  |  |  |
| `spec.customText.afterSubmit` | `StripePaymentLinkMessage` |  |  |  |
| `spec.customText.afterSubmit.message` | `string` | yes |  |  |
| `spec.customText.shippingAddress` | `StripePaymentLinkMessage` |  |  |  |
| `spec.customText.shippingAddress.message` | `string` | yes |  |  |
| `spec.customText.submit` | `StripePaymentLinkMessage` |  |  |  |
| `spec.customText.submit.message` | `string` | yes |  |  |
| `spec.customText.termsOfServiceAcceptance` | `StripePaymentLinkMessage` |  |  |  |
| `spec.customText.termsOfServiceAcceptance.message` | `string` | yes |  |  |
| `spec.customerCreation` | `enum` |  |  |  |
| `spec.invoiceCreation` | `StripePaymentLinkInvoiceCreation` |  |  |  |
| `spec.invoiceCreation.enabled` | `bool` |  |  |  |
| `spec.invoiceCreation.invoiceData` | `StripePaymentLinkInvoiceData` |  |  |  |
| `spec.invoiceCreation.invoiceData.accountTaxIds` | `[]string` |  |  |  |
| `spec.invoiceCreation.invoiceData.customFields` | `[]StripePaymentLinkInvoiceCustomField` |  |  |  |
| `spec.invoiceCreation.invoiceData.customFields[].name` | `string` | yes |  |  |
| `spec.invoiceCreation.invoiceData.customFields[].value` | `string` | yes |  |  |
| `spec.invoiceCreation.invoiceData.description` | `string` |  |  |  |
| `spec.invoiceCreation.invoiceData.footer` | `string` |  |  |  |
| `spec.invoiceCreation.invoiceData.issuer` | `StripePaymentLinkAccountRef` |  |  |  |
| `spec.invoiceCreation.invoiceData.issuer.account` | `string` |  |  |  |
| `spec.invoiceCreation.invoiceData.metadata` | `map<string, string>` |  |  |  |
| `spec.invoiceCreation.invoiceData.renderingOptions` | `StripePaymentLinkRenderingOptions` |  |  |  |
| `spec.invoiceCreation.invoiceData.renderingOptions.amountTaxDisplay` | `string` |  |  |  |
| `spec.invoiceCreation.invoiceData.renderingOptions.template` | `string` |  |  |  |
| `spec.managedPayments` | `StripePaymentLinkManagedPayments` |  |  |  |
| `spec.managedPayments.enabled` | `bool` |  |  |  |
| `spec.nameCollection` | `StripePaymentLinkNameCollection` |  |  |  |
| `spec.nameCollection.business` | `StripePaymentLinkNameField` |  |  |  |
| `spec.nameCollection.business.enabled` | `bool` |  |  |  |
| `spec.nameCollection.business.optional` | `bool` |  |  |  |
| `spec.nameCollection.individual` | `StripePaymentLinkNameField` |  |  |  |
| `spec.nameCollection.individual.enabled` | `bool` |  |  |  |
| `spec.nameCollection.individual.optional` | `bool` |  |  |  |
| `spec.paymentIntentData` | `StripePaymentLinkPaymentIntentData` |  |  |  |
| `spec.paymentIntentData.captureMethod` | `enum` |  |  |  |
| `spec.paymentIntentData.description` | `string` |  |  |  |
| `spec.paymentIntentData.metadata` | `map<string, string>` |  |  |  |
| `spec.paymentIntentData.setupFutureUsage` | `enum` |  |  |  |
| `spec.paymentIntentData.statementDescriptor` | `string` |  |  |  |
| `spec.paymentIntentData.statementDescriptorSuffix` | `string` |  |  |  |
| `spec.paymentIntentData.transferGroup` | `string` |  |  |  |
| `spec.paymentMethodCollection` | `enum` |  |  |  |
| `spec.paymentMethodOptions` | `StripePaymentLinkPaymentMethodOptions` |  |  |  |
| `spec.paymentMethodOptions.card` | `StripePaymentLinkCardOptions` |  |  |  |
| `spec.paymentMethodOptions.card.restrictions` | `StripePaymentLinkCardRestrictions` |  |  |  |
| `spec.paymentMethodOptions.card.restrictions.brandsBlocked` | `[]string` |  |  |  |
| `spec.paymentMethodTypes` | `[]string` |  |  |  |
| `spec.phoneNumberCollection` | `StripePaymentLinkEnabled` |  |  |  |
| `spec.phoneNumberCollection.enabled` | `bool` |  |  |  |
| `spec.restrictions` | `StripePaymentLinkRestrictions` |  |  |  |
| `spec.restrictions.completedSessions` | `StripePaymentLinkCompletedSessions` | yes |  |  |
| `spec.restrictions.completedSessions.limit` | `int64` |  |  |  |
| `spec.shippingAddressCollection` | `StripePaymentLinkShippingAddressCollection` |  |  |  |
| `spec.shippingAddressCollection.allowedCountries` | `[]string` | yes |  |  |
| `spec.shippingOptions` | `[]StripePaymentLinkShippingOption` |  |  |  |
| `spec.shippingOptions[].shippingRate` | `string \| valueFrom` | yes |  | StripeShippingRate (`status.outputs.id`) |
| `spec.submitType` | `enum` |  |  |  |
| `spec.subscriptionData` | `StripePaymentLinkSubscriptionData` |  |  |  |
| `spec.subscriptionData.description` | `string` |  |  |  |
| `spec.subscriptionData.invoiceSettings` | `StripePaymentLinkSubscriptionInvoiceSettings` |  |  |  |
| `spec.subscriptionData.invoiceSettings.issuer` | `StripePaymentLinkAccountRef` |  |  |  |
| `spec.subscriptionData.invoiceSettings.issuer.account` | `string` |  |  |  |
| `spec.subscriptionData.metadata` | `map<string, string>` |  |  |  |
| `spec.subscriptionData.trialPeriodDays` | `int64` |  |  |  |
| `spec.subscriptionData.trialSettings` | `StripePaymentLinkTrialSettings` |  |  |  |
| `spec.subscriptionData.trialSettings.endBehavior` | `StripePaymentLinkTrialEndBehavior` | yes |  |  |
| `spec.subscriptionData.trialSettings.endBehavior.missingPaymentMethod` | `enum` |  |  |  |
| `spec.taxIdCollection` | `StripePaymentLinkTaxIdCollection` |  |  |  |
| `spec.taxIdCollection.enabled` | `bool` |  |  |  |
| `spec.taxIdCollection.required` | `enum` |  |  |  |
| `spec.applicationFeeAmount` | `int64` |  |  |  |
| `spec.applicationFeePercent` | `double` |  |  |  |
| `spec.onBehalfOf` | `string` |  |  |  |
| `spec.transferData` | `StripePaymentLinkTransferData` |  |  |  |
| `spec.transferData.destination` | `string` | yes |  |  |
| `spec.transferData.amount` | `int64` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.lineItems

`[]StripePaymentLinkLineItem` · required

line_items are the prices the link sells, 1 to 20, in the order shown. Changing a price, a
quantity or the order REPLACES the link; adjustable quantities update in place.

- rule: {"repeated":{"minItems":"1","maxItems":"20"}}

### spec.lineItems[].price

`string | valueFrom` · required

price is the price sold (price_...). Reference a StripePrice. Changing it REPLACES the link.

- references: StripePrice (`status.outputs.id`)
- rule: price is a Stripe price id (price_...), or a reference to a StripePrice
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripePrice, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.lineItems[].quantity

`int64`

quantity is how many are sold, at least 1. Changing it REPLACES the link: the provider never
stores it, so an update in place could never be detected or sent.

- rule: {"int64":{"gte":"1"}}

### spec.lineItems[].adjustableQuantity

`StripePaymentLinkAdjustableQuantity`

adjustable_quantity lets the buyer change the quantity. It updates in place.

- rule: adjustable_quantity keeps minimum <= maximum

### spec.lineItems[].adjustableQuantity.enabled

`bool`

enabled lets the buyer change the quantity.

### spec.lineItems[].adjustableQuantity.minimum

`int64` · optional (explicit presence)

minimum is the smallest quantity the buyer may choose.

- rule: {"int64":{"gte":"0"}}

### spec.lineItems[].adjustableQuantity.maximum

`int64` · optional (explicit presence)

maximum is the largest quantity the buyer may choose.

- rule: {"int64":{"gte":"1"}}

### spec.optionalItems

`[]StripePaymentLinkOptionalItem`

optional_items are up to 10 extra prices the buyer may add at checkout. They update in place.

- rule: {"repeated":{"maxItems":"10"}}

### spec.optionalItems[].price

`string | valueFrom` · required

price is the price offered (price_...). Reference a StripePrice.

- references: StripePrice (`status.outputs.id`)
- rule: price is a Stripe price id (price_...), or a reference to a StripePrice
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripePrice, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.optionalItems[].quantity

`int64`

quantity is the quantity added when the buyer takes the item, at least 1.

- rule: {"int64":{"gte":"1"}}

### spec.optionalItems[].adjustableQuantity

`StripePaymentLinkAdjustableQuantity`

adjustable_quantity lets the buyer change the quantity.

- rule: adjustable_quantity keeps minimum <= maximum

### spec.optionalItems[].adjustableQuantity.enabled

`bool`

enabled lets the buyer change the quantity.

### spec.optionalItems[].adjustableQuantity.minimum

`int64` · optional (explicit presence)

minimum is the smallest quantity the buyer may choose.

- rule: {"int64":{"gte":"0"}}

### spec.optionalItems[].adjustableQuantity.maximum

`int64` · optional (explicit presence)

maximum is the largest quantity the buyer may choose.

- rule: {"int64":{"gte":"1"}}

### spec.active

`bool` · optional (explicit presence)

active is whether the link accepts payments. Setting it false deactivates the link without
destroying the resource; its address shows inactive_message.

- default: `true`

### spec.inactiveMessage

`string`

inactive_message is shown at the link's address while it is deactivated, up to 500 characters.

- rule: {"string":{"maxLen":"500"}}

### spec.afterCompletion

`StripePaymentLinkAfterCompletion`

after_completion is what the buyer sees after paying. Unset, Stripe's confirmation page.

### spec.afterCompletion.hostedConfirmation

`StripePaymentLinkHostedConfirmation`

hosted_confirmation shows Stripe's confirmation page, optionally with a custom message
(hosted_confirmation: {} is Stripe's page as it is).

### spec.afterCompletion.hostedConfirmation.customMessage

`string`

custom_message replaces the page's default message.

### spec.afterCompletion.redirect

`StripePaymentLinkRedirect`

redirect sends the buyer to the account's site.

### spec.afterCompletion.redirect.url

`string` · required

url is where the buyer goes. {CHECKOUT_SESSION_ID} in it is replaced with the session's id.

- rule: url is an https:// or http:// address
- rule: {"required":true}

### spec.allowPromotionCodes

`bool` · optional (explicit presence)

allow_promotion_codes lets the buyer enter a promotion code.

### spec.automaticTax

`StripePaymentLinkAutomaticTax`

automatic_tax calculates tax with Stripe Tax.

### spec.automaticTax.enabled

`bool`

enabled turns Stripe Tax on for the link.

### spec.automaticTax.liability

`StripePaymentLinkAccountRef`

liability is the account liable for the tax. Unset, the account itself.

### spec.automaticTax.liability.account

`string`

account is the connected account (acct_...). Empty, the account the link belongs to.

- rule: account is a Stripe connected account id (acct_...)

### spec.billingAddressCollection

`enum`

billing_address_collection is whether the buyer's billing address is always collected.
Unset, auto: only when needed.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `billing_address_collection_unspecified`
- `auto` -- auto collects it only when needed.
- `required` -- required always collects it.

### spec.consentCollection

`StripePaymentLinkConsentCollection`

consent_collection asks the buyer to agree to terms or promotional emails. Changing it
REPLACES the link.

### spec.consentCollection.paymentMethodReuseAgreement

`StripePaymentLinkPaymentMethodReuseAgreement`

payment_method_reuse_agreement is where the agreement to save the payment method is shown.

### spec.consentCollection.paymentMethodReuseAgreement.position

`enum`

position is auto (Stripe shows it where needed) or hidden (the account shows its own).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `position_unspecified`
- `auto`
- `hidden`

### spec.consentCollection.promotions

`enum`

promotions asks the buyer to agree to promotional emails. Unset, none.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `promotions_unspecified`
- `auto` -- auto asks where the law requires it.
- `none`

### spec.consentCollection.termsOfService

`enum`

terms_of_service requires the buyer to accept the account's terms (set in the Dashboard's
public details). Unset, none.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `stripe_payment_link_terms_of_service_unspecified`
- `none`
- `required`

### spec.currency

`string`

currency is the lowercase three-letter ISO code the link charges in, when its prices offer
several ("usd"). Changing it REPLACES the link.

- rule: currency is a lowercase three-letter ISO code, such as usd

### spec.customFields

`[]StripePaymentLinkCustomField`

custom_fields are up to 3 extra questions the buyer answers.

- rule: {"repeated":{"maxItems":"3"}}

### spec.customFields[].key

`string` · required

key identifies the answer in the Checkout session, up to 200 letters and digits.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9]{1,200}$"}}

### spec.customFields[].label

`string` · required

label is the question as the buyer sees it, up to 50 characters.

- rule: {"required":true,"string":{"maxLen":"50"}}

### spec.customFields[].optional

`bool` · optional (explicit presence)

optional lets the buyer skip the question.

### spec.customFields[].dropdown

`StripePaymentLinkDropdown`

dropdown makes the answer a choice from a list.

### spec.customFields[].dropdown.options

`[]StripePaymentLinkDropdownOption` · required

options are the choices, 1 to 200.

- rule: {"repeated":{"minItems":"1","maxItems":"200"}}

### spec.customFields[].dropdown.options[].label

`string` · required

label is the choice as the buyer sees it, up to 100 characters.

- rule: {"required":true,"string":{"maxLen":"100"}}

### spec.customFields[].dropdown.options[].value

`string` · required

value is what the Checkout session records, up to 100 letters and digits.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9]{1,100}$"}}

### spec.customFields[].dropdown.defaultValue

`string`

default_value is the value of the option selected at first.

### spec.customFields[].numeric

`StripePaymentLinkTextBounds`

numeric makes the answer a number (numeric: {} with no bounds is enough to choose it).

- rule: minimum_length is at most maximum_length

### spec.customFields[].numeric.defaultValue

`string`

default_value is the answer filled in at first.

### spec.customFields[].numeric.minimumLength

`int64` · optional (explicit presence)

minimum_length is the shortest answer, 0 to 255.

- rule: {"int64":{"lte":"255","gte":"0"}}

### spec.customFields[].numeric.maximumLength

`int64` · optional (explicit presence)

maximum_length is the longest answer, 1 to 255.

- rule: {"int64":{"lte":"255","gte":"1"}}

### spec.customFields[].text

`StripePaymentLinkTextBounds`

text makes the answer free text (text: {} with no bounds is enough to choose it).

- rule: minimum_length is at most maximum_length

### spec.customFields[].text.defaultValue

`string`

default_value is the answer filled in at first.

### spec.customFields[].text.minimumLength

`int64` · optional (explicit presence)

minimum_length is the shortest answer, 0 to 255.

- rule: {"int64":{"lte":"255","gte":"0"}}

### spec.customFields[].text.maximumLength

`int64` · optional (explicit presence)

maximum_length is the longest answer, 1 to 255.

- rule: {"int64":{"lte":"255","gte":"1"}}

### spec.customText

`StripePaymentLinkCustomText`

custom_text is extra text shown on the page.

### spec.customText.afterSubmit

`StripePaymentLinkMessage`

after_submit is shown after the payment button.

### spec.customText.afterSubmit.message

`string` · required

message is the text, up to 1200 characters.

- rule: {"required":true,"string":{"maxLen":"1200"}}

### spec.customText.shippingAddress

`StripePaymentLinkMessage`

shipping_address is shown beside the shipping address form.

### spec.customText.shippingAddress.message

`string` · required

message is the text, up to 1200 characters.

- rule: {"required":true,"string":{"maxLen":"1200"}}

### spec.customText.submit

`StripePaymentLinkMessage`

submit is shown beside the payment button.

### spec.customText.submit.message

`string` · required

message is the text, up to 1200 characters.

- rule: {"required":true,"string":{"maxLen":"1200"}}

### spec.customText.termsOfServiceAcceptance

`StripePaymentLinkMessage`

terms_of_service_acceptance replaces the terms acceptance text.

### spec.customText.termsOfServiceAcceptance.message

`string` · required

message is the text, up to 1200 characters.

- rule: {"required":true,"string":{"maxLen":"1200"}}

### spec.customerCreation

`enum`

customer_creation is whether a one-time payment creates a Stripe customer. Unset,
if_required. A subscription always creates one.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `customer_creation_unspecified`
- `always` -- always creates a customer.
- `if_required` -- if_required creates one only when needed.

### spec.invoiceCreation

`StripePaymentLinkInvoiceCreation`

invoice_creation sends a post-purchase invoice for one-time payments.

### spec.invoiceCreation.enabled

`bool`

enabled sends the invoice.

### spec.invoiceCreation.invoiceData

`StripePaymentLinkInvoiceData`

invoice_data configures the invoice.

### spec.invoiceCreation.invoiceData.accountTaxIds

`[]string`

account_tax_ids are the account's tax ids shown on the invoice (txi_...).

- rule: {"repeated":{"unique":true}}

### spec.invoiceCreation.invoiceData.customFields

`[]StripePaymentLinkInvoiceCustomField`

custom_fields are up to 4 name-value lines shown on the invoice.

- rule: {"repeated":{"maxItems":"4"}}

### spec.invoiceCreation.invoiceData.customFields[].name

`string` · required

name is the line's label, up to 40 characters.

- rule: {"required":true,"string":{"maxLen":"40"}}

### spec.invoiceCreation.invoiceData.customFields[].value

`string` · required

value is the line's value, up to 140 characters.

- rule: {"required":true,"string":{"maxLen":"140"}}

### spec.invoiceCreation.invoiceData.description

`string`

description is the invoice's description.

### spec.invoiceCreation.invoiceData.footer

`string`

footer is the invoice's footer.

### spec.invoiceCreation.invoiceData.issuer

`StripePaymentLinkAccountRef`

issuer is the account that issues the invoice. Unset, the account itself.

### spec.invoiceCreation.invoiceData.issuer.account

`string`

account is the connected account (acct_...). Empty, the account the link belongs to.

- rule: account is a Stripe connected account id (acct_...)

### spec.invoiceCreation.invoiceData.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the invoice.

### spec.invoiceCreation.invoiceData.renderingOptions

`StripePaymentLinkRenderingOptions`

rendering_options control the invoice's PDF.

### spec.invoiceCreation.invoiceData.renderingOptions.amountTaxDisplay

`string`

amount_tax_display is how amounts show tax on the PDF: exclude_tax, or
include_inclusive_tax. Unset, Stripe's default.

### spec.invoiceCreation.invoiceData.renderingOptions.template

`string`

template is the invoice rendering template to use (inrtem_...).

### spec.managedPayments

`StripePaymentLinkManagedPayments`

managed_payments sells through Stripe's managed payments. Changing it REPLACES the link.

### spec.managedPayments.enabled

`bool` · optional (explicit presence)

enabled sells through managed payments.

### spec.nameCollection

`StripePaymentLinkNameCollection`

name_collection collects the buyer's name, as an individual or a business.

### spec.nameCollection.business

`StripePaymentLinkNameField`

business collects a business name.

### spec.nameCollection.business.enabled

`bool`

enabled shows the field.

### spec.nameCollection.business.optional

`bool` · optional (explicit presence)

optional lets the buyer leave it empty.

### spec.nameCollection.individual

`StripePaymentLinkNameField`

individual collects a person's name.

### spec.nameCollection.individual.enabled

`bool`

enabled shows the field.

### spec.nameCollection.individual.optional

`bool` · optional (explicit presence)

optional lets the buyer leave it empty.

### spec.paymentIntentData

`StripePaymentLinkPaymentIntentData`

payment_intent_data configures the payment a one-time purchase creates.

### spec.paymentIntentData.captureMethod

`enum`

capture_method is when the funds are captured. Unset, Stripe's default. Changing it REPLACES
the link.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `capture_method_unspecified`
- `automatic`
- `automatic_async`
- `manual` -- manual authorizes now and captures later.

### spec.paymentIntentData.description

`string`

description is the payment's description.

### spec.paymentIntentData.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the payment.

### spec.paymentIntentData.setupFutureUsage

`enum`

setup_future_usage saves the payment method for later payments. Changing it REPLACES the
link.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `setup_future_usage_unspecified`
- `off_session`
- `on_session`

### spec.paymentIntentData.statementDescriptor

`string`

statement_descriptor is the text on the buyer's card statement for a non-card payment, up
to 22 characters.

- rule: {"string":{"maxLen":"22"}}

### spec.paymentIntentData.statementDescriptorSuffix

`string`

statement_descriptor_suffix is appended to the account's statement descriptor for a card
payment, up to 22 characters in total.

- rule: {"string":{"maxLen":"22"}}

### spec.paymentIntentData.transferGroup

`string`

transfer_group groups the payment's Connect transfers.

### spec.paymentMethodCollection

`enum`

payment_method_collection is whether a subscription collects a payment method when the first
invoice is free (a trial). Unset, always.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `stripe_payment_link_payment_method_collection_unspecified`
- `always` -- always collects a payment method.
- `if_required` -- if_required collects one only when something is due.

### spec.paymentMethodOptions

`StripePaymentLinkPaymentMethodOptions`

payment_method_options configure individual payment methods.

### spec.paymentMethodOptions.card

`StripePaymentLinkCardOptions`

card configures card payments.

### spec.paymentMethodOptions.card.restrictions

`StripePaymentLinkCardRestrictions`

restrictions limit which cards are accepted.

### spec.paymentMethodOptions.card.restrictions.brandsBlocked

`[]string`

brands_blocked are the card brands refused ("american_express", "discover_global_network",
"mastercard", "visa").

- rule: {"repeated":{"unique":true}}

### spec.paymentMethodTypes

`[]string`

payment_method_types are the payment methods offered ("card", "link"). Unset, Stripe chooses
from the account's payment method settings, which is recommended.

- rule: {"repeated":{"unique":true}}

### spec.phoneNumberCollection

`StripePaymentLinkEnabled`

phone_number_collection collects the buyer's phone number.

### spec.phoneNumberCollection.enabled

`bool`

enabled turns it on.

### spec.restrictions

`StripePaymentLinkRestrictions`

restrictions stop the link after a number of completed payments.

### spec.restrictions.completedSessions

`StripePaymentLinkCompletedSessions` · required

completed_sessions deactivates the link after a number of completed payments.

- rule: {"required":true}

### spec.restrictions.completedSessions.limit

`int64`

limit is the number of completed payments after which the link deactivates itself.

- rule: {"int64":{"gte":"1"}}

### spec.shippingAddressCollection

`StripePaymentLinkShippingAddressCollection`

shipping_address_collection collects a shipping address in the listed countries.

### spec.shippingAddressCollection.allowedCountries

`[]string` · required

allowed_countries are the two-letter ISO codes of the countries shipped to ("US", "DE").

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"pattern":"^[A-Z]{2}$"}}}}

### spec.shippingOptions

`[]StripePaymentLinkShippingOption`

shipping_options are the shipping rates the buyer chooses from (shr_...). Reference a
StripeShippingRate. Changing them REPLACES the link.

### spec.shippingOptions[].shippingRate

`string | valueFrom` · required

shipping_rate is the rate (shr_...). Reference a StripeShippingRate.

- references: StripeShippingRate (`status.outputs.id`)
- rule: shipping_rate is a Stripe shipping rate id (shr_...), or a reference to a StripeShippingRate
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripeShippingRate, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.submitType

`enum`

submit_type is the payment button's label. Unset, auto (pay, or subscribe for a subscription).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `stripe_payment_link_submit_type_unspecified`
- `auto`
- `book`
- `donate`
- `pay`
- `subscribe`

### spec.subscriptionData

`StripePaymentLinkSubscriptionData`

subscription_data configures the subscription a recurring price creates.

### spec.subscriptionData.description

`string`

description is the subscription's description, shown in the customer portal and on invoices.
Changing it REPLACES the link.

### spec.subscriptionData.invoiceSettings

`StripePaymentLinkSubscriptionInvoiceSettings`

invoice_settings configure the subscription's invoices.

### spec.subscriptionData.invoiceSettings.issuer

`StripePaymentLinkAccountRef`

issuer is the account that issues the invoices. Unset, the account itself.

### spec.subscriptionData.invoiceSettings.issuer.account

`string`

account is the connected account (acct_...). Empty, the account the link belongs to.

- rule: account is a Stripe connected account id (acct_...)

### spec.subscriptionData.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on each subscription.

### spec.subscriptionData.trialPeriodDays

`int64` · optional (explicit presence)

trial_period_days is the free trial each subscription starts with, at least 1 day.

- rule: {"int64":{"gte":"1"}}

### spec.subscriptionData.trialSettings

`StripePaymentLinkTrialSettings`

trial_settings are what happens when a trial ends.

### spec.subscriptionData.trialSettings.endBehavior

`StripePaymentLinkTrialEndBehavior` · required

end_behavior is what happens at the trial's end.

- rule: {"required":true}

### spec.subscriptionData.trialSettings.endBehavior.missingPaymentMethod

`enum`

missing_payment_method is what happens when the trial ends without a payment method.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `missing_payment_method_unspecified`
- `cancel` -- cancel cancels the subscription.
- `create_invoice` -- create_invoice creates an invoice the customer must pay.
- `pause` -- pause pauses the subscription until a payment method is added.

### spec.taxIdCollection

`StripePaymentLinkTaxIdCollection`

tax_id_collection collects the buyer's tax id.

### spec.taxIdCollection.enabled

`bool`

enabled shows the tax id field.

### spec.taxIdCollection.required

`enum`

required is whether a business buyer must give one. Unset, never.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `required_unspecified`
- `if_supported` -- if_supported requires one where Stripe can validate it.
- `never`

### spec.applicationFeeAmount

`int64` · optional (explicit presence)

application_fee_amount is the platform's fee, in the smallest currency unit, taken from each
one-time payment and kept by the Connect platform. Only for links with no recurring prices.
Changing it REPLACES the link.

- rule: {"int64":{"gte":"0"}}

### spec.applicationFeePercent

`double` · optional (explicit presence)

application_fee_percent is the platform's share of each subscription payment, 0 to 100, kept
by the Connect platform. Only for links with a recurring price. Changing it REPLACES the link.

- rule: {"double":{"lte":100,"gte":0}}

### spec.onBehalfOf

`string`

on_behalf_of is the connected account (acct_...) the payments are made on behalf of: it
becomes the settlement merchant, and its name and statement descriptor are shown to the buyer.
Changing it REPLACES the link.

- rule: on_behalf_of is a Stripe connected account id (acct_...)

### spec.transferData

`StripePaymentLinkTransferData`

transfer_data sends each payment's funds to a connected account. Changing it REPLACES the
link.

### spec.transferData.destination

`string` · required

destination is the connected account that receives the funds (acct_...).

- rule: destination is a Stripe connected account id (acct_...)
- rule: {"required":true}

### spec.transferData.amount

`int64` · optional (explicit presence)

amount is how much of each payment is transferred, in the smallest currency unit. Unset, the
whole payment less fees.

- rule: {"int64":{"gte":"0"}}

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the link. It updates in place.

## Validation Rules

- `spec.items_at_most_twenty`: a link sells at most 20 prices: line_items and optional_items together are at most 20
- `spec.one_application_fee`: a link takes one kind of platform fee: application_fee_amount (one-time) or application_fee_percent (subscription), not both

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripePaymentLink, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the payment link's Stripe id (plink_...). It changes when the link is replaced. |
| `status.outputs.url` | `string` | url is the page's public address (https://buy.stripe.com/...). A site's configuration reads it by reference, so it follows a replacement on the same apply. |
| `status.outputs.active` | `bool` | active is false once the link is deactivated (by destroy, by spec.active, by a completed- session limit, or in the Dashboard). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.lineItems[].price` | StripePrice | `status.outputs.id` |
| `spec.optionalItems[].price` | StripePrice | `status.outputs.id` |
| `spec.shippingOptions[].shippingRate` | StripeShippingRate | `status.outputs.id` |

## See Also

- [Overview](../README.md)
