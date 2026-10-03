# Azure Application Insights Standard Web Test

Deploys an Application Insights Standard Web Test -- a synthetic availability monitor that issues an HTTP request to a URL from one or more Azure test-agent locations on a schedule and records whether the endpoint responded correctly. It is how you prove an endpoint is reachable and healthy from the OUTSIDE -- and, paired with a Monitor metric alert on the test's availability, how you get paged when it is not. The spec covers the probe request, the response assertions (status code, SSL certificate lifetime, body content), the schedule, and the Azure test-agent locations the probe runs from.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Standard Web Test** -- a `Microsoft.Insights/webTests` resource bound to the referenced Application Insights component, configured with the probe request (URL, method, body, headers), the response assertions (status code, SSL certificate, body content), the schedule (frequency, timeout, retry), and the geo-locations it runs from
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the web test will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **An Application Insights component** to store the test's results and host its availability metric. Provide the component ARM ID directly or reference an AzureApplicationInsights Infra Component via ValueFromRef.
- **A publicly reachable endpoint** -- Azure's test agents probe from the public internet; an endpoint behind an IP allowlist needs Azure's published test-agent ranges admitted.

## Deploy

### Console

Open the deployment store, find **Azure Application Insights Standard Web Test**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Multi-Region Endpoint Availability** preset in the [Presets](#presets) tab to pre-populate a multi-region availability probe.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureApplicationInsightsStandardWebTest
metadata:
  name: checkout-availability
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  name: checkout-api-availability
  region: eastus
  applicationInsightsId:
    valueFrom:
      name: checkout-api-insights
  request:
    url: https://api.example.com/healthz
  geoLocations:
    - us-va-ash-azr
    - emea-nl-ams-azr
    - apac-sg-sin-azr
```

```shell
planton apply -f web-test.yaml
```

This creates a web test on Azure's defaults: a GET probe every 5 minutes from the three listed locations, a 30-second timeout, and any 200 response passing. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the test to a resource group and Application Insights component deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  applicationInsightsId:
    valueFrom:
      kind: AzureApplicationInsights
      name: checkout-api-insights
      fieldPath: status.outputs.application_insights_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and Application Insights component first, then provisions the web test with the resolved values.

## Key Configuration

These are the most important decisions when configuring a standard web test. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Probe request** -- The required `request` block carries the URL (must be publicly reachable), the HTTP method (empty deploys Azure's GET default; the closed set is GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS), an optional body for the verbs that carry one, redirect and dependent-request behavior, and extra headers. Probe a dedicated health endpoint -- stable, unauthenticated, and honest about the service's real dependencies.

**Geo locations** -- The `geoLocations` list (at least one required) names the Azure test-agent locations the probe runs FROM -- location IDs like `us-va-ash-azr`, not region names. Run from 3+ locations across continents so a single region's network trouble does not read as an outage.

**Response validation** -- The optional `validationRules` block raises the bar for "healthy": an exact status code (assert 301 to watch a redirect, 401 to prove an endpoint stays protected), the SSL certificate check with a remaining-lifetime floor (1-365 days -- expiry becomes an early warning, not an outage), and a body content match (pass-if-found asserts healthy content; the default fail-if-found watches for an error string). SSL assertions require an https URL, and the lifetime floor rides the SSL check -- Azure silently drops it when the check is off.

**Schedule** -- `frequency` is Azure's closed set: 300, 600, or 900 seconds (blank deploys the 5-minute default). `timeout` (blank = 30 seconds) doubles as a worst-case latency assertion. `retryEnabled` filters transient blips out of the availability signal; `enabled: false` keeps the definition but stops the runs.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureApplicationInsights** | `applicationInsightsId` | `status.outputs.application_insights_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `web_test_id` | Azure Resource Manager ID of the web test | Monitor metric alerts on the test's availability -- how a failed probe becomes a page |
| `web_test_name` | Name of the web test | Azure CLI references, portal navigation |
| `synthetic_monitor_id` | The synthetic monitor identity Azure assigns | Correlating availability results in Application Insights queries |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Multi-region availability probe** -- a GET against a health endpoint every 5 minutes from three continents, with retry on and the SSL lifetime floor set. The standard outside-in check for anything customer-facing. Start from the **Multi-Region Endpoint Availability** preset.

**Content check** -- a single-location probe asserting the response body carries a healthy marker (`"status":"ok"`), catching the error page that still ships a 200. Start from the **Response-Content Health Check** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the web test is created
- [**Azure Application Insights**](/infra-catalog/azure-application-insights) -- stores the test's results and hosts its availability metric
- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- the workspace behind the Application Insights component, where availability results are queryable with KQL
