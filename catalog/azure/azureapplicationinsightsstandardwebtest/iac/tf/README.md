# AzureApplicationInsightsStandardWebTest - Terraform Module

Terraform implementation for the AzureApplicationInsightsStandardWebTest
kind.

## Resources Created

- `azurerm_application_insights_standard_web_test.main` -- the synthetic
  availability test, attached to the referenced Application Insights
  component and executed from the configured Azure locations

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.application_insights_id` | The Application Insights component the test reports into |
| `spec.request` | The probe: `url` is required; verb, body, headers, and redirect-following are presence-guarded so unspecified specs deploy Azure's defaults (GET, follow redirects) |
| `spec.geo_locations` | Azure location IDs the test runs FROM (required) |
| `spec.validation_rules` | Optional status-code, SSL, and content-match assertions, realized as nested blocks only when set |
| `spec.frequency` / `spec.timeout` | Cadence and per-run budget; unset deploys Azure's 5-minute / 30-second defaults |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- Optional request/validation fields are omitted (not defaulted) when the
  spec leaves them unset, so both engines send identical request bodies.
- The `synthetic_monitor_id` output is Azure-assigned; the `web_test_id`
  output is the seam an `AzureMonitorMetricAlert` references through its
  `web_test_id` criteria.

## Usage

```hcl
module "web_test" {
  source = "./path/to/module"

  metadata = { name = "homepage-availability" }
  spec = {
    resource_group          = "observability-rg"
    region                  = "eastus"
    name                    = "homepage-availability"
    application_insights_id = "/subscriptions/.../components/app-insights"
    geo_locations           = ["us-ca-sjc-azr", "emea-nl-ams-azr"]
    request = {
      url = "https://example.com/healthz"
    }
  }
}
```
