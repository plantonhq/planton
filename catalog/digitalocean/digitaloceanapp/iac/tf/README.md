# DigitalOcean App -- Terraform Module

Deploys a `digitalocean_app` from a `DigitalOceanApp` spec. Dynamic blocks cover every component family, domains, ingress, alerts, env vars, VPC, and maintenance. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanApp`). Do not hand-edit it. The API token lives in `credentials.tf` so the generated file stays generator-owned.

## Prerequisites

- OpenTofu or Terraform 1.0+
- DigitalOcean API token (`digitalocean_token`)

## Usage

```hcl
module "app" {
  source = "./path/to/module"

  metadata = {
    name = "demo-app"
  }

  spec = {
    app_name = "demo-app"
    region   = "nyc"
    services = [
      {
        name = "web"
        image = {
          registry_type = "docker_hub"
          registry      = "library"
          repository    = "nginx"
          tag           = "latest"
        }
        instance_size_slug = "basic-xxs"
        instance_count     = 1
        http_port          = 80
      }
    ]
  }

  digitalocean_token = var.digitalocean_token
}

output "live_url" {
  value = module.app.live_url
}
```

Enum strings that are empty or end in `_unspecified` are omitted. Image `registry_type` is upper-cased on the wire (`docker_hub` → `DOCKER_HUB`). Region is sent as-is (`nyc3`).

## Outputs

| Output | Description |
|--------|-------------|
| `app_id` | App UUID (import id) |
| `default_hostname` | Default `ondigitalocean.app` hostname |
| `live_url` | Public URL including protocol |
| `live_domain` | Live hostname without scheme |
| `active_deployment_id` | Currently live deployment UUID |

See the kind [README](../../README.md) and [GUIDE](../../GUIDE.md) for the spec, including the two arms only this engine wires at the current Pulumi SDK (service/worker liveness probes and ingress `secure_header`).
