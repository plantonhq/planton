# Azure Front Door Endpoint

Deploys a public Front Door endpoint -- the `*.azurefd.net` hostname clients hit at Microsoft's edge. The endpoint belongs to an AzureFrontDoorProfile and is the attachment point for routes and, later, custom domains; endpoints are many-per-profile with independent lifecycles, so one profile commonly fronts several applications, each behind its own endpoint. Renaming the endpoint replaces it and changes the generated hostname, so every DNS record pointing at the old name breaks.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Endpoint** -- a named child of the profile with a generated, globally unique public hostname (`{endpointName}-{hash}.z01.azurefd.net`)
- **Enabled state** -- whether the endpoint accepts traffic (default on; disable for maintenance without deleting the hostname)
- **Azure Tags** -- Planton-derived resource tags merged with user tags and applied to the endpoint

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Front Door Profile** the endpoint nests under. Reference an AzureFrontDoorProfile Infra Component via ValueFromRef, or provide the profile ARM ID directly.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Endpoint**, and click **Deploy**. The wizard walks you through the parent profile, the endpoint name (2–46 characters -- shorter than other Front Door names because Azure builds the public hostname from it), and the enabled switch. Start from the **Production Endpoint** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorEndpoint
metadata:
  name: my-web-endpoint
  org: acme-corp
  env: prod
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  endpointName: my-web
```

```shell
planton apply -f front-door-endpoint.yaml
```

This creates an enabled endpoint under the profile with a generated `my-web-{hash}.z01.azurefd.net` hostname, ready for routes to attach. An Infra Job tracks the provisioning in real time.

### InfraChart

In an InfraChart, wire the endpoint to its profile through `valueFrom`:

```yaml
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  endpointName: my-web
```

The InfraPipeline resolves the dependency graph, provisioning the profile before the endpoint that references it.

## Key Configuration

These are the most important decisions when configuring a Front Door endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Endpoint name** -- 2–46 characters; letters, digits, and hyphens; must start and end with a letter or digit. Becomes the prefix of the generated public hostname, so pick something recognizable in browser address bars and DNS records. ForceNew: renaming replaces the endpoint and changes the hostname -- every DNS record pointing at the old name breaks.

**One endpoint per application** -- endpoints are many-per-profile with independent lifecycles. Splitting applications across endpoints (rather than piling routes onto one) keeps each hostname's blast radius contained: a disabled or replaced endpoint takes down one app, not the whole profile.

**Enabled** -- the traffic switch. Leave on for production; set false for a maintenance hold or a dark launch without deleting the DNS target. Disabling stops requests at the edge (clients get errors from Front Door; the backend sees nothing) and flipping it back is a fast in-place update. Defaults to true when omitted.

**Tags** -- free-form tags merged over the Planton-derived resource tags (organization, environment, resource id); a user tag with the same key wins. Azure Policy enforces them and Microsoft Cost Management groups by them. Updatable in place.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureFrontDoorProfile** | `profileId` | `status.outputs.profile_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint_id` | ARM resource ID of the endpoint | AzureFrontDoorRoute.`endpointId`; AzureFrontDoorSecurityPolicy WAF association |
| `endpoint_name` | The endpoint's name | Operator tooling |
| `host_name` | The generated `*.azurefd.net` hostname | AzureDnsRecord CNAME/alias targets; AzureFrontDoorCustomDomain validation |

## Common Patterns

**One endpoint per fronted application** -- the common shape: a single profile carries several endpoints, one per app, and each endpoint anchors the routes (`/api/*`, `/static/*`) that split its hostname's traffic across origin groups. Start from the **Production Endpoint** preset.

**Dark launch with a pre-provisioned endpoint** -- create the endpoint with `enabled: false`: it is fully provisioned and the `host_name` output exists immediately, so CNAME records and routes can be prepared ahead of the cutover, and launch day is a single in-place flip of `enabled`. Start from the **Pre-Provisioned (Disabled) Endpoint** preset.

**Maintenance hold without teardown** -- disabling an endpoint stops traffic at the edge while every route, domain attachment, and the generated hostname survive intact. The trade against deleting: clients receive errors from Front Door during the hold, so pair the flip with your maintenance-window communication.

## Works With

- [**Azure Front Door Profile**](/infra-catalog/azure-front-door-profile) -- the parent container referenced by `profileId`; its SKU tier and identity govern every child
- [**Azure Front Door Route**](/infra-catalog/azure-front-door-route) -- attaches URL patterns to this endpoint and forwards them to an origin group
- [**Azure Front Door Custom Domain**](/infra-catalog/azure-front-door-custom-domain) -- serves a branded hostname that CNAMEs onto this endpoint's `host_name` output
- [**Azure Front Door Security Policy**](/infra-catalog/azure-front-door-security-policy) -- associates a WAF policy with this endpoint through its `endpoint_id`
- [**Azure DNS Record**](/infra-catalog/azure-dns-record) -- CNAMEs or aliases a custom hostname at the generated `host_name`
