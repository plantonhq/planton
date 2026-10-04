# Multi-Region Endpoint Availability

This preset creates a standard web test that pings a public endpoint every 5
minutes from three regions, asserts a 200 with a valid SSL certificate at
least 30 days from expiry, and retries a failed run before counting it. It
is the everyday "is my site up, from the outside" monitor. Pair it with a
metric alert on the availability metric to get paged on sustained failures.

## When to Use

- Any public HTTP(S) endpoint whose uptime you must prove and be alerted on
- Catching SSL certificates about to expire before they do
- Distinguishing a real outage from a single-region network blip (three
  locations)

## Key Configuration Choices

- **Three `geoLocations`** -- a failure from one region alone is likely a
  network blip; a failure from all three is a real outage
- **`retryEnabled: true`** -- suppresses false alarms from transient
  single-run failures
- **`sslCertRemainingLifetime: 30`** -- fails the test (and alerts) 30 days
  before the certificate expires, giving you time to renew

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<resource-group-name>` | The resource group to create the test in | The resource group's `status.outputs.resource_group_name` |
| `<application-insights>` | The AzureApplicationInsights component storing results | Your component's Planton resource name |
| `https://www.example.com/health` | The URL to monitor | Your public endpoint |

## Downstream Wiring

Alert on the test's availability with a metric alert:

```yaml
spec:
  webTestAvailabilityCriteria:
    webTestId:
      valueFrom:
        name: my-endpoint-web-test
    componentId:
      valueFrom:
        name: <application-insights>
```
