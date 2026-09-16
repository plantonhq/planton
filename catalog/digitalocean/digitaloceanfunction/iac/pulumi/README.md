# DigitalOcean Function -- Pulumi Module

Deploys a `digitalocean:index/app:App` with a single functions component from a `DigitalOceanFunction` spec. There is no standalone Functions resource.

The app name is `spec.appName` (2-32 characters, unique in the account -- never `metadata.name`, which the API would reject for length). The functions component name is `spec.functionName`. Source (`git` / `github` / `gitlab` / `bitbucket`) is actually set on the component — a Functions deploy with an empty source cannot build. `sourceDirectory` is omitted from the app spec when unset so App Platform reads `project.yml` from the repository root.

Runtime, memory, timeout, and schedules are read by App Platform from `project.yml`. They are not Pulumi args.

Component alert destinations (emails, Slack webhooks) are wired -- the SDK carries them since pulumi-digitalocean v4.53.0. Webhook URLs are wrapped as Pulumi secrets because the SDK does not flag them.

## Prerequisites

- Pulumi CLI 3.x
- Go 1.21+
- `DIGITALOCEAN_TOKEN`

## Outputs

| Output | Description |
|--------|-------------|
| `function_id` | App UUID that hosts the functions component |
| `https_endpoint` | Public HTTPS URL |
| `default_hostname` | Default `ondigitalocean.app` hostname |

## Usage

```go
package main

import (
    digitaloceanfunctionv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanfunction/v1alpha1"
    "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanfunction/iac/pulumi/module"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        return module.Resources(ctx, stackInput)
    })
}
```

The Planton runner supplies `stackInput`. See the kind [README](../../README.md) and [GUIDE](../../GUIDE.md).
