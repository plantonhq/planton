# DigitalOcean Certificate -- Terraform Module

Deploys a `digitalocean_certificate` from a `DigitalOceanCertificate` spec: the certificate's stable name plus exactly one source branch -- Let's Encrypt domains or custom PEM material. DigitalOcean's `type` argument is derived from whichever branch is set. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanCertificate`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)
- For Let's Encrypt: every domain managed by DigitalOcean DNS in the same account

## Usage

```hcl
module "certificate" {
  source = "./path/to/module"

  metadata = {
    name = "web-cert"
  }

  spec = {
    certificate_name = "web-cert"
    lets_encrypt = {
      domains = ["example.com", "*.example.com"]
    }
  }

  digitalocean_token = var.digitalocean_token
}
```

## Behavior notes

- The oneof arrives as two optional objects; exactly one is non-null (validation enforces it upstream) and the module derives `type` from the set branch -- a mismatched type/branch combination is unrepresentable.
- Every argument is create-only; `create_before_destroy` makes any replacement zero-downtime for load balancers referencing the certificate by name.
- The resource id (and the `certificate_id` output) is the certificate NAME at this pin -- a Let's Encrypt certificate's UUID rotates on every auto-renewal.
- The provider stores only hashes of PEM material; the API never returns it (recorded config-only import tolerances for all three PEM fields).

## Outputs

Exactly the kind's output contract, identical to the Pulumi module: `certificate_id`, `expiry_rfc3339`.
