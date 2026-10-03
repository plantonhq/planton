# CloudflareQueue Terraform Module

Terraform IaC module for provisioning a Cloudflare Queue and its optional single consumer — a managed, guaranteed-delivery message queue for Workers.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareQueueSpec
locals.tf     — Resource naming
main.tf       — cloudflare_queue + cloudflare_queue_consumer resources
outputs.tf    — outputs (queue_id, queue_name, created_on, modified_on)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareQueue YAML manifest. For standalone use:

```hcl
module "queue" {
  source = "./path/to/module"

  metadata = {
    name = "orders-queue"
  }

  spec = {
    account_id = "your-account-id"
    queue_name = "orders-queue"
    settings = {
      message_retention_period = 86400
    }
    consumer = {
      type = "http_pull"
      settings = {
        batch_size = 5
      }
    }
  }
}
```

The consumer is a separate provider resource gated on the `consumer` block, so editing consumer settings never recreates the queue. A `worker` (push) consumer requires `consumer.script_name`; an `http_pull` consumer forbids it — external clients pull batches over the REST API instead.

## Outputs

| Name | Description |
|------|-------------|
| `queue_id` | Cloudflare-assigned queue ID |
| `queue_name` | The queue name (echoed; referenced by producer bindings and DLQs) |
| `created_on` | RFC3339 creation timestamp |
| `modified_on` | RFC3339 last-modified timestamp |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
