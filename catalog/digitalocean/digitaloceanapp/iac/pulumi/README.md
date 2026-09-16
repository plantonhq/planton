# DigitalOcean App -- Pulumi Module

Deploys a `digitalocean:index/app:App` from a `DigitalOceanApp` spec. The module maps every component family (service, worker, job, static site, function, in-app database), domains, ingress, alerts, and env vars onto the App Platform spec.

## Pulumi SDK gaps (v4.53.0)

Two spec fields are real and Terraform wires them, but the Pulumi DigitalOcean SDK at v4.53.0 (verified on disk) has no matching args. This module fails the apply with `PARITY-EXCEPTION` if they are set, so both engines always deploy the same app:

- service/worker `livenessHealthCheck`
- `spec.ingress.secureHeader`

Omit them on Pulumi stacks, or deploy those arms with Terraform. `spec.vpc` (a one-element `vpcs` list on the SDK), `spec.maintenance`, `spec.ingress.rule.match.authority`, and alert destinations on app-level and component alerts closed at v4.53.0 and are wired. Slack webhook URLs are wrapped as Pulumi secrets because the SDK does not flag them.

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
