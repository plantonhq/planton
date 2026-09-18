# DigitalOcean App -- Pulumi Module

Deploys a `digitalocean:index/app:App` from a `DigitalOceanApp` spec. The module maps every component family (service, worker, job, static site, function, in-app database), domains, ingress, alerts, and env vars onto the App Platform spec.

## SDK shape notes

The full spec surface is wired; there is no Pulumi SDK gap on this resource at the pin (`pulumi-digitalocean/sdk/v4 v4.79.1`). `spec.vpc` is a one-element `vpcs` list on the SDK. `spec.ingress.secureHeader` sits on `AppSpecIngressArgs` (ingress-wide, beside `rules`), not on a rule. Service and worker `livenessHealthCheck` reuse the readiness check's field shape on their own SDK types. Slack webhook URLs are wrapped as Pulumi secrets because the SDK does not flag them.

## Prerequisites

- Pulumi CLI 3.x
- Go 1.21+
- `DIGITALOCEAN_TOKEN`

## Outputs

| Output | Description |
|--------|-------------|
| `app_id` | App UUID |
| `default_hostname` | Default `ondigitalocean.app` hostname (scheme stripped from `DefaultIngress`) |
| `live_url` | Public URL including protocol |
| `live_domain` | Live hostname without scheme |
| `active_deployment_id` | Currently live deployment UUID |

App Platform has no tag surface; computed labels are not sent.

## Usage

```go
package main

import (
    digitaloceanappv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanapp/v1alpha1"
    "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanapp/iac/pulumi/module"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        return module.Resources(ctx, stackInput)
    })
}
```

The Planton runner supplies `stackInput`. See the kind [README](../../README.md) and [GUIDE](../../GUIDE.md) for the spec.
