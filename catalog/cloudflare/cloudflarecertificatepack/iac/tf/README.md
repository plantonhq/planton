# CloudflareCertificatePack Terraform Module

Terraform IaC module for ordering an advanced edge certificate pack on a Cloudflare zone — a CA-issued certificate covering the zone apex and its hostnames.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareCertificatePackSpec
locals.tf     — Zone id flattening and type default ("advanced")
main.tf       — cloudflare_certificate_pack resource
outputs.tf    — outputs (certificate_pack_id, zone_id)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareCertificatePack YAML manifest. For standalone use:

```hcl
module "certificate_pack" {
  source = "./path/to/module"

  metadata = {
    name = "apex-pack"
  }

  spec = {
    zone_id               = "your-zone-id"
    certificate_authority = "lets_encrypt"
    validation_method     = "txt"
    validity_days         = 90
    hosts                 = ["example.com", "*.example.com"]
  }
}
```

A pack is an order, not an editable object: changing hosts, CA, validation method, or validity days replaces the pack. Import is supported but not recommended — replace the certificate instead. `zone_id` is exported because a pack's API identity is (zone_id, certificate_pack_id).

## Outputs

| Name | Description |
|------|-------------|
| `certificate_pack_id` | Cloudflare-assigned pack ID |
| `zone_id` | The zone the pack was ordered in |

There is no `status` output: issuance is asynchronous (`initializing` → `pending_validation` → `active`), and a point-in-time phase is never a stable output — it flips on the first refresh after the transition and re-plans forever.

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
