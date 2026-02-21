# HetznerCloudCertificate Terraform Module

Terraform IaC module for provisioning Hetzner Cloud TLS certificates (uploaded or managed via Let's Encrypt).

## Structure

```
.
├── main.tf           # Certificate resources (uploaded and managed, conditional)
├── outputs.tf        # Stack output definitions (conditional selection)
├── variables.tf      # Input variable definitions (metadata, spec, hcloud_token)
├── locals.tf         # Certificate name, type flags, standard label computation
└── provider.tf       # Hetzner Cloud provider configuration (~> 1.60)
```

## Resources Created

Exactly one of:

- `hcloud_uploaded_certificate` (count 0 or 1) — stores a user-provided PEM certificate chain and private key. Created when `spec.uploaded` is set.
- `hcloud_managed_certificate` (count 0 or 1) — requests a Let's Encrypt certificate for the specified domains. Created when `spec.managed` is set.

## Outputs

| Name | Description |
|------|-------------|
| `certificate_id` | Hetzner Cloud numeric ID of the created certificate |
| `type` | Certificate type: `"uploaded"` or `"managed"` |
| `fingerprint` | SHA256 fingerprint of the certificate |
| `not_valid_before` | Certificate validity start (ISO-8601) |
| `not_valid_after` | Certificate validity end (ISO-8601) |

Each output uses a ternary expression to select from whichever resource was created.

## Usage

### Managed Certificate

```bash
terraform init

terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"api-cert"}' \
  -var 'spec={"managed":{"domain_names":["example.com","www.example.com"]}}'

terraform apply \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"api-cert"}' \
  -var 'spec={"managed":{"domain_names":["example.com","www.example.com"]}}'
```

### Uploaded Certificate

```bash
terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"wildcard-cert"}' \
  -var 'spec={"uploaded":{"certificate":"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----","private_key":"-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"}}'
```

For structured input, use a `.tfvars` file:

```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```
