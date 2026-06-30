# HetznerCloudDnsZone Terraform Module

Terraform IaC module for provisioning Hetzner Cloud DNS zones with record sets. Supports primary mode (records managed via variables) and secondary mode (records synchronized from an external primary nameserver via zone transfer).

## Structure

```
.
├── main.tf           # Zone and record set resources
├── outputs.tf        # Stack output definitions
├── variables.tf      # Input variable definitions (metadata, spec, hcloud_token)
├── locals.tf         # Standard label computation, record set map for for_each
└── provider.tf       # HetznerCloud provider configuration (~> 1.60)
```

## Resources Created

- `hcloud_zone.this` (always) — the DNS zone with domain name, mode, TTL, labels, delete protection, and (for secondary mode) primary nameserver configuration.
- `hcloud_zone_rrset.this` (0–N, `for_each`) — one per record set entry in `var.spec.record_sets`. Each rrset manages all DNS records for a unique (name, type) pair. Keyed by `"{name}-{lowercase_type}"`.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The Hetzner Cloud numeric ID of the created DNS zone |
| `nameservers` | The authoritative Hetzner nameservers assigned to the zone |

## Usage

### Primary Zone with Records

```bash
terraform init

terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"my-zone"}' \
  -var 'spec={"domain_name":"example.com","mode":"primary","ttl":3600,"record_sets":[{"name":"@","type":"A","records":[{"value":"93.184.216.34"}]}]}'

terraform apply \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"my-zone"}' \
  -var 'spec={"domain_name":"example.com","mode":"primary","ttl":3600,"record_sets":[{"name":"@","type":"A","records":[{"value":"93.184.216.34"}]}]}'
```

### Secondary Zone

```bash
terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"secondary-zone"}' \
  -var 'spec={"domain_name":"example.com","mode":"secondary","primary_nameservers":[{"address":"10.0.0.1","port":53,"tsig_algorithm":"hmac-sha256","tsig_key":"secret"}]}'
```

### Using a .tfvars File

For structured input, use a `.tfvars` file:

```hcl
# terraform.tfvars

hcloud_token = "your-api-token"

metadata = {
  name = "my-zone"
  org  = "acme-corp"
  env  = "production"
}

spec = {
  domain_name       = "example.com"
  mode              = "primary"
  ttl               = 3600
  delete_protection = true

  record_sets = [
    {
      name = "@"
      type = "A"
      ttl  = 300
      records = [
        { value = "93.184.216.34", comment = "web-1" },
        { value = "93.184.216.35", comment = "web-2" },
      ]
    },
    {
      name = "www"
      type = "CNAME"
      records = [
        { value = "example.com." },
      ]
    },
    {
      name = "@"
      type = "MX"
      records = [
        { value = "10 mail.example.com." },
        { value = "20 backup.example.com." },
      ]
    },
  ]
}
```

Then:

```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```

### Inspecting Outputs

```bash
terraform output zone_id
terraform output nameservers
```
