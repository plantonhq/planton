# StripeBillingMeter

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeBillingMeterSpec declares a usage meter: the event name the application sends to Stripe
for each unit of use (an API request, a GB stored), how those events add up per customer, and
alerts when a customer's usage crosses a threshold. A metered StripePrice names the meter to
bill what it counts.

Only the display name can change in Stripe. Changing the event name, time window, customer
mapping, aggregation or value key REPLACES the meter: Planton deactivates the old one and
creates the new one, and every metered price that references it is replaced with it. The
application must then send the new meter's events, so read the event name from
status.outputs.event_name by reference rather than repeating it.

Destroy deactivates the meter (status inactive): it stops accepting events and Stripe keeps it.
A meter that cannot be read makes the plan fail (the provider does not treat a missing meter
as gone); remove it from state and apply again.

Alerts are forgotten, not deleted: Stripe's alerts can't be updated or deleted through the
provider, so changing an alert replaces it and leaves the old one active in Stripe, and
removing an alert or destroying the meter leaves every alert active. Archive a stale alert in
the Dashboard, or with Stripe's archive call (POST /v1/billing/alerts/{id}/archive).

The key needs "Billing Meters" write, and "Alerts" write when alerts are declared
(iac/permissions.yaml).

https://docs.stripe.com/billing/subscriptions/usage-based/recording-usage
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/billing_meter

## Example

```yaml
# The canonical example: API requests summed per customer, with an alert at
# 10,000.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeBillingMeter
metadata:
  name: api-requests
  org: e2e-org
  env: testing
spec:
  displayName: API requests
  eventName: api_requests
  defaultAggregation:
    formula: sum
  customerMapping:
    eventPayloadKey: stripe_customer_id
  valueSettings:
    eventPayloadKey: value
  alerts:
    - title: 10k requests
      gte: 10000
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.displayName` | `string` | yes |  |  |
| `spec.eventName` | `string` | yes |  |  |
| `spec.defaultAggregation` | `StripeBillingMeterDefaultAggregation` | yes |  |  |
| `spec.defaultAggregation.formula` | `enum` |  |  |  |
| `spec.customerMapping` | `StripeBillingMeterCustomerMapping` |  |  |  |
| `spec.customerMapping.eventPayloadKey` | `string` | yes |  |  |
| `spec.valueSettings` | `StripeBillingMeterValueSettings` |  |  |  |
| `spec.valueSettings.eventPayloadKey` | `string` | yes |  |  |
| `spec.eventTimeWindow` | `enum` |  |  |  |
| `spec.alerts` | `[]StripeBillingMeterAlert` |  |  |  |
| `spec.alerts[].title` | `string` | yes |  |  |
| `spec.alerts[].gte` | `int64` |  |  |  |
| `spec.alerts[].customer` | `string` |  |  |  |
| `spec.alerts[].recurrence` | `enum` |  |  |  |

## Field Details

### spec.displayName

`string` · required

display_name is the meter's name in the Dashboard and on invoices ("API requests"). It is the
only field that updates in place.

- rule: {"required":true}

### spec.eventName

`string` · required

event_name is the name the application sends with each usage event ("api_requests").
Changing it REPLACES the meter.

- rule: {"required":true}

### spec.defaultAggregation

`StripeBillingMeterDefaultAggregation` · required

default_aggregation is how the events in a billing period add up to the usage billed.
Changing it REPLACES the meter.

- rule: {"required":true}

### spec.defaultAggregation.formula

`enum`

formula is the aggregation.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `formula_unspecified`
- `count` -- count counts the events.
- `sum` -- sum adds up the events' values.
- `last` -- last takes the last event's value.

### spec.customerMapping

`StripeBillingMeterCustomerMapping`

customer_mapping is where in each event Stripe finds the customer. Unset, Stripe reads the
stripe_customer_id payload key. Changing it REPLACES the meter.

### spec.customerMapping.eventPayloadKey

`string` · required

event_payload_key is the payload key that holds the Stripe customer id
("stripe_customer_id"). Stripe maps by customer id, the only mapping it offers.

- rule: {"required":true}

### spec.valueSettings

`StripeBillingMeterValueSettings`

value_settings is where in each event Stripe finds the amount of use. Unset, Stripe reads the
value payload key. Changing it REPLACES the meter.

### spec.valueSettings.eventPayloadKey

`string` · required

event_payload_key is the payload key that holds the numeric value ("value", "bytes").

- rule: {"required":true}

### spec.eventTimeWindow

`enum`

event_time_window is the period events are pre-aggregated over for reporting. Unset, events
are not pre-aggregated. Changing it REPLACES the meter.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `event_time_window_unspecified`
- `day`
- `hour`

### spec.alerts

`[]StripeBillingMeterAlert`

alerts notify the account (a billing.alert.triggered event) when a customer's usage on this
meter reaches a threshold. Each is keyed by its title, unique within the meter. An alert
can't be changed: changing one creates a new alert and leaves the old one active in Stripe,
and removing one (or destroying the meter) leaves it active too.

### spec.alerts[].title

`string` · required

title is the alert's name, shown in the Dashboard and in the event, unique within the meter.
It keys the alert: changing it creates a new alert and leaves the old one active in Stripe.

- rule: {"required":true}

### spec.alerts[].gte

`int64`

gte is the usage, in the meter's units, at which the alert fires.

- rule: {"int64":{"gt":"0"}}

### spec.alerts[].customer

`string`

customer limits the alert to one customer's usage (cus_...). Unset, it watches every
customer.

- rule: customer is a Stripe customer id (cus_...)

### spec.alerts[].recurrence

`enum`

recurrence is how often the alert can fire. Unset, one_time: once per customer, the only
recurrence Stripe offers.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `recurrence_unspecified`
- `one_time` -- one_time fires once per customer.

## Validation Rules

- `spec.alerts.unique_title`: each alert has its own title: rename or remove the repeated one

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeBillingMeter, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the meter's Stripe id (mtr_...), the value a metered StripePrice's recurring.meter references. It changes when the meter is replaced. |
| `status.outputs.event_name` | `string` | event_name is the name the application sends with each usage event. An application's deployment reads it by reference, so it follows a replacement. |
| `status.outputs.status` | `string` | status is active, or inactive once the meter is deactivated (by destroy or in the Dashboard). |
| `status.outputs.alert_ids` | `map<string, string>` | alert_ids maps each alert's title to its Stripe id. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| StripePrice | `spec.recurring.meter` | `status.outputs.id` |

## See Also

- [Overview](../README.md)
