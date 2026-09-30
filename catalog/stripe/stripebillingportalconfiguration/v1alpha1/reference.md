# StripeBillingPortalConfiguration

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeBillingPortalConfigurationSpec declares what Stripe's customer portal lets a customer
do: cancel (and how), switch plans (and to which prices), update payment methods and details,
and see invoice history.

An account holds many configurations and one of them is its default, which Stripe uses when a
portal session names none. This kind always creates a configuration of its own and never
adopts or edits the default: the application names this configuration's id
(status.outputs.id) when it opens a portal session, so every customer sees exactly what is
declared here.

Stripe never deletes a configuration. Destroy deactivates it (active = false) and Stripe keeps
it, inactive, forever; a portal session that names it afterwards is refused. A setting removed
from the manifest is not reset in Stripe -- the provider sends only values that are set -- so
change a setting to its opposite rather than deleting it. A configuration deactivated outside
Planton is reactivated by the next apply; one that cannot be read makes the plan fail until it
is removed from state.

The key needs "Customer portal" write (iac/permissions.yaml).

https://docs.stripe.com/customer-management/configure-portal
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/billing_portal_configuration

## Example

```yaml
# The canonical example: a portal with invoices, payment methods and
# cancellation at period end. The lanes run against the dedicated test
# sandbox only.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeBillingPortalConfiguration
metadata:
  name: customer-portal
  org: e2e-org
  env: testing
spec:
  name: Planton catalog example portal
  features:
    invoiceHistory:
      enabled: true
    paymentMethodUpdate:
      enabled: true
    subscriptionCancel:
      enabled: true
      mode: at_period_end
  businessProfile:
    privacyPolicyUrl: https://example.com/privacy
    termsOfServiceUrl: https://example.com/terms
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` |  |  |  |
| `spec.features` | `StripeBillingPortalFeatures` | yes |  |  |
| `spec.features.customerUpdate` | `StripeBillingPortalCustomerUpdate` |  |  |  |
| `spec.features.customerUpdate.enabled` | `bool` |  |  |  |
| `spec.features.customerUpdate.allowedUpdates` | `[]enum` |  |  |  |
| `spec.features.invoiceHistory` | `StripeBillingPortalInvoiceHistory` |  |  |  |
| `spec.features.invoiceHistory.enabled` | `bool` |  |  |  |
| `spec.features.paymentMethodUpdate` | `StripeBillingPortalPaymentMethodUpdate` |  |  |  |
| `spec.features.paymentMethodUpdate.enabled` | `bool` |  |  |  |
| `spec.features.paymentMethodUpdate.paymentMethodConfiguration` | `string \| valueFrom` |  |  | StripePaymentMethodConfiguration (`status.outputs.id`) |
| `spec.features.subscriptionCancel` | `StripeBillingPortalSubscriptionCancel` |  |  |  |
| `spec.features.subscriptionCancel.enabled` | `bool` |  |  |  |
| `spec.features.subscriptionCancel.mode` | `enum` |  |  |  |
| `spec.features.subscriptionCancel.prorationBehavior` | `enum` |  |  |  |
| `spec.features.subscriptionCancel.cancellationReason` | `StripeBillingPortalCancellationReason` |  |  |  |
| `spec.features.subscriptionCancel.cancellationReason.enabled` | `bool` |  |  |  |
| `spec.features.subscriptionCancel.cancellationReason.options` | `[]enum` | yes |  |  |
| `spec.features.subscriptionUpdate` | `StripeBillingPortalSubscriptionUpdate` |  |  |  |
| `spec.features.subscriptionUpdate.enabled` | `bool` |  |  |  |
| `spec.features.subscriptionUpdate.defaultAllowedUpdates` | `[]enum` |  |  |  |
| `spec.features.subscriptionUpdate.prorationBehavior` | `enum` |  |  |  |
| `spec.features.subscriptionUpdate.billingCycleAnchor` | `enum` |  |  |  |
| `spec.features.subscriptionUpdate.trialUpdateBehavior` | `enum` |  |  |  |
| `spec.features.subscriptionUpdate.products` | `[]StripeBillingPortalProduct` |  |  |  |
| `spec.features.subscriptionUpdate.products[].product` | `string \| valueFrom` | yes |  | StripeProduct (`status.outputs.id`) |
| `spec.features.subscriptionUpdate.products[].prices` | `[]string \| valueFrom` | yes |  | StripePrice (`status.outputs.id`) |
| `spec.features.subscriptionUpdate.products[].adjustableQuantity` | `StripeBillingPortalAdjustableQuantity` |  |  |  |
| `spec.features.subscriptionUpdate.products[].adjustableQuantity.enabled` | `bool` |  |  |  |
| `spec.features.subscriptionUpdate.products[].adjustableQuantity.minimum` | `int64` |  |  |  |
| `spec.features.subscriptionUpdate.products[].adjustableQuantity.maximum` | `int64` |  |  |  |
| `spec.features.subscriptionUpdate.scheduleAtPeriodEnd` | `StripeBillingPortalScheduleAtPeriodEnd` |  |  |  |
| `spec.features.subscriptionUpdate.scheduleAtPeriodEnd.conditions` | `[]StripeBillingPortalScheduleCondition` |  |  |  |
| `spec.features.subscriptionUpdate.scheduleAtPeriodEnd.conditions[].type` | `enum` |  |  |  |
| `spec.businessProfile` | `StripeBillingPortalBusinessProfile` |  |  |  |
| `spec.businessProfile.headline` | `string` |  |  |  |
| `spec.businessProfile.privacyPolicyUrl` | `string` |  |  |  |
| `spec.businessProfile.termsOfServiceUrl` | `string` |  |  |  |
| `spec.defaultReturnUrl` | `string` |  |  |  |
| `spec.loginPage` | `StripeBillingPortalLoginPage` |  |  |  |
| `spec.loginPage.enabled` | `bool` |  |  |  |
| `spec.active` | `bool` |  | `true` |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.name

`string`

name is a label for the configuration, shown only to the account's team.

### spec.features

`StripeBillingPortalFeatures` · required

features are what a customer can do in the portal. Every feature is off unless enabled here.

- rule: {"required":true}

### spec.features.customerUpdate

`StripeBillingPortalCustomerUpdate`

customer_update lets a customer edit their own details.

### spec.features.customerUpdate.enabled

`bool`

enabled turns the feature on.

### spec.features.customerUpdate.allowedUpdates

`[]enum`

allowed_updates are the details a customer may edit. Empty with the feature enabled, the
portal allows none.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `allowed_update_unspecified`
- `address`
- `email`
- `name`
- `phone`
- `shipping`
- `tax_id`

### spec.features.invoiceHistory

`StripeBillingPortalInvoiceHistory`

invoice_history lets a customer see and download past invoices.

### spec.features.invoiceHistory.enabled

`bool`

enabled turns the feature on.

### spec.features.paymentMethodUpdate

`StripeBillingPortalPaymentMethodUpdate`

payment_method_update lets a customer add and change payment methods.

### spec.features.paymentMethodUpdate.enabled

`bool`

enabled turns the feature on.

### spec.features.paymentMethodUpdate.paymentMethodConfiguration

`string | valueFrom`

payment_method_configuration is the payment-method configuration (pmc_...) that decides
which methods a customer may add in the portal. Reference a StripePaymentMethodConfiguration
so the portal offers the same methods as checkout. Unset, the account's default
payment-method configuration applies.

- references: StripePaymentMethodConfiguration (`status.outputs.id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: StripePaymentMethodConfiguration, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.features.subscriptionCancel

`StripeBillingPortalSubscriptionCancel`

subscription_cancel lets a customer cancel a subscription.

- rule: proration_behavior always_invoice or create_prorations settles unused time, which only an immediate cancellation leaves: set mode immediately, or proration_behavior none

### spec.features.subscriptionCancel.enabled

`bool`

enabled turns the feature on.

### spec.features.subscriptionCancel.mode

`enum`

mode is when a cancellation takes effect. Unset, Stripe cancels at the end of the period.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `mode_unspecified`
- `at_period_end` -- at_period_end keeps the subscription until the paid period ends.
- `immediately` -- immediately ends the subscription now.

### spec.features.subscriptionCancel.prorationBehavior

`enum`

proration_behavior is how an immediate cancellation settles the unused time. It applies only
with mode immediately.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `proration_behavior_unspecified`
- `always_invoice` -- always_invoice credits the unused time and invoices the result at once.
- `create_prorations` -- create_prorations credits the unused time to the customer's next invoice.
- `none` -- none leaves the unused time unrefunded.

### spec.features.subscriptionCancel.cancellationReason

`StripeBillingPortalCancellationReason`

cancellation_reason asks the customer why they are leaving.

### spec.features.subscriptionCancel.cancellationReason.enabled

`bool`

enabled shows the question.

### spec.features.subscriptionCancel.cancellationReason.options

`[]enum` · required

options are the reasons offered, in Stripe's fixed vocabulary.

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `option_unspecified`
- `customer_service`
- `low_quality`
- `missing_features`
- `other`
- `switched_service`
- `too_complex`
- `too_expensive`
- `unused`

### spec.features.subscriptionUpdate

`StripeBillingPortalSubscriptionUpdate`

subscription_update lets a customer switch plans or quantities.

- rule: switching prices needs the prices to switch between: list at least one product under products, or drop price from default_allowed_updates

### spec.features.subscriptionUpdate.enabled

`bool`

enabled turns the feature on. The portal needs at least one product to switch between.

### spec.features.subscriptionUpdate.defaultAllowedUpdates

`[]enum`

default_allowed_updates are what a customer may change on a subscription.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `default_allowed_update_unspecified`
- `price`
- `promotion_code`
- `quantity`

### spec.features.subscriptionUpdate.prorationBehavior

`enum`

proration_behavior is how a change mid-period settles the difference. Unset, Stripe uses
none.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `proration_behavior_unspecified`
- `always_invoice`
- `create_prorations`
- `none`

### spec.features.subscriptionUpdate.billingCycleAnchor

`enum`

billing_cycle_anchor is whether a change restarts the billing period. Unset, Stripe keeps it
unchanged.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `billing_cycle_anchor_unspecified`
- `now`
- `unchanged`

### spec.features.subscriptionUpdate.trialUpdateBehavior

`enum`

trial_update_behavior is what a change does to a subscription still in trial. Unset, Stripe
ends the trial.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `trial_update_behavior_unspecified`
- `continue_trial`
- `end_trial`

### spec.features.subscriptionUpdate.products

`[]StripeBillingPortalProduct`

products are the products, and their prices, a customer may switch between. Stripe allows up
to ten.

- rule: {"repeated":{"maxItems":"10"}}
- rule: each price is listed once: remove the repeated price id

### spec.features.subscriptionUpdate.products[].product

`string | valueFrom` · required

product is the product a customer may switch to (prod_...). Reference a StripeProduct so the
portal offers exactly the plan declared beside it.

- references: StripeProduct (`status.outputs.id`)
- rule: product is a Stripe product id (prod_...), or a reference to a StripeProduct
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripeProduct, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.features.subscriptionUpdate.products[].prices

`[]string | valueFrom` · required

prices are the product's prices a customer may choose (price_...). Reference each
StripePrice. A price replaced after an amount change has a new id, and a reference follows it
on the portal's next apply.

- references: StripePrice (`status.outputs.id`)
- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"product.prices.format","message":"a price is a Stripe price id (price_...), or a reference to a StripePrice","expression":"!has(this.value) || this.value.startsWith('price_')"}]}}}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripePrice, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.features.subscriptionUpdate.products[].adjustableQuantity

`StripeBillingPortalAdjustableQuantity`

adjustable_quantity lets a customer change the quantity of this product.

- rule: minimum must not exceed maximum

### spec.features.subscriptionUpdate.products[].adjustableQuantity.enabled

`bool`

enabled lets the customer change the quantity.

### spec.features.subscriptionUpdate.products[].adjustableQuantity.minimum

`int64` · optional (explicit presence)

minimum is the smallest quantity allowed.

- rule: {"int64":{"gte":"0"}}

### spec.features.subscriptionUpdate.products[].adjustableQuantity.maximum

`int64` · optional (explicit presence)

maximum is the largest quantity allowed.

- rule: {"int64":{"gte":"1"}}

### spec.features.subscriptionUpdate.scheduleAtPeriodEnd

`StripeBillingPortalScheduleAtPeriodEnd`

schedule_at_period_end defers some changes to the end of the period instead of applying them
at once.

### spec.features.subscriptionUpdate.scheduleAtPeriodEnd.conditions

`[]StripeBillingPortalScheduleCondition`

conditions are the changes that wait for the period to end.

### spec.features.subscriptionUpdate.scheduleAtPeriodEnd.conditions[].type

`enum`

type is the change deferred.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `type_unspecified`
- `decreasing_item_amount` -- decreasing_item_amount defers a change that lowers what the customer pays.
- `shortening_interval` -- shortening_interval defers a change to a shorter billing interval.

### spec.businessProfile

`StripeBillingPortalBusinessProfile`

business_profile is the portal's headline and the policy links it shows.

### spec.businessProfile.headline

`string`

headline is the message shown to customers at the top of the portal.

### spec.businessProfile.privacyPolicyUrl

`string`

privacy_policy_url links the business's privacy policy.

- rule: privacy_policy_url is an https:// or http:// URL

### spec.businessProfile.termsOfServiceUrl

`string`

terms_of_service_url links the business's terms of service.

- rule: terms_of_service_url is an https:// or http:// URL

### spec.defaultReturnUrl

`string`

default_return_url is where the portal's "return" link leads when the session that opened
the portal names no return URL of its own.

- rule: default_return_url is an https:// or http:// URL

### spec.loginPage

`StripeBillingPortalLoginPage`

login_page turns on a shareable portal URL where a customer signs in with their email
(status.outputs.login_page_url), for accounts that do not open sessions from their own
application.

### spec.loginPage.enabled

`bool`

enabled creates the shareable URL; turning it off retires the URL.

### spec.active

`bool` · optional (explicit presence)

active is whether portal sessions may use this configuration. Setting it false deactivates
the configuration without destroying the resource; destroy deactivates it too.

- default: `true`

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the configuration.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeBillingPortalConfiguration, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the configuration's Stripe id (bpc_...), the value an application passes as `configuration` when it creates a portal session. |
| `status.outputs.is_default` | `bool` | is_default is whether this is the account's default configuration. A configuration this kind creates is not the default unless someone makes it so in the Dashboard. |
| `status.outputs.active` | `bool` | active is whether portal sessions may use the configuration. |
| `status.outputs.login_page_url` | `string` | login_page_url is the shareable portal sign-in URL, when login_page is enabled. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.features.paymentMethodUpdate.paymentMethodConfiguration` | StripePaymentMethodConfiguration | `status.outputs.id` |
| `spec.features.subscriptionUpdate.products[].product` | StripeProduct | `status.outputs.id` |
| `spec.features.subscriptionUpdate.products[].prices` | StripePrice | `status.outputs.id` |

## See Also

- [Overview](../README.md)
