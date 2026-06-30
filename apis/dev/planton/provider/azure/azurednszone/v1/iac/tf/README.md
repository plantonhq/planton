# Azure DNS Zone - Terraform Module

This Terraform module provisions an Azure DNS Zone with support for all common DNS record types.

## Overview

This module creates:
- An Azure DNS Zone in a specified resource group
- DNS records (A, AAAA, CNAME, MX, TXT, NS, CAA, SRV, PTR) as configured
- Proper tagging for resource organization and tracking

## Prerequisites

- Terraform >= 1.0
- Azure CLI configured with appropriate credentials
- An existing Azure Resource Group
- Domain registered with a registrar (Azure DNS is not a registrar)

## Usage

### Basic Example

```hcl
module "dns_zone" {
  source = "./path/to/module"

  metadata = {
    name = "example.com"
    id   = "dns-001"
    org  = "myorg"
    env  = "production"
  }

  spec = {
    zone_name      = "example.com"
    resource_group = "prod-network-rg"
    
    records = [
      {
        name        = "@"
        record_type = "A"
        values      = ["203.0.113.10"]
        ttl_seconds = 3600
      },
      {
        name        = "www"
        record_type = "CNAME"
        values      = ["example.com."]
        ttl_seconds = 3600
      }
    ]
  }
}
```

### Complete Example with Email Records

```hcl
module "dns_zone_with_email" {
  source = "./path/to/module"

  metadata = {
    name = "contoso.com"
    id   = "dns-contoso-prod"
    org  = "contoso"
    env  = "production"
  }

  spec = {
    zone_name      = "contoso.com"
    resource_group = "prod-network-rg"
    
    records = [
      # Web hosting
      {
        name        = "@"
        record_type = "A"
        values      = ["198.51.100.45"]
        ttl_seconds = 300
      },
      {
        name        = "www"
        record_type = "CNAME"
        values      = ["contoso.com."]
        ttl_seconds = 300
      },
      
      # Mail servers
      {
        name        = "@"
        record_type = "MX"
        values      = ["mail.contoso.com.", "backup-mail.contoso.com."]
        ttl_seconds = 3600
      },
      
      # Mail server A records
      {
        name        = "mail"
        record_type = "A"
        values      = ["198.51.100.100"]
        ttl_seconds = 300
      },
      {
        name        = "backup-mail"
        record_type = "A"
        values      = ["198.51.100.101"]
        ttl_seconds = 300
      },
      
      # Email authentication
      {
        name        = "@"
        record_type = "TXT"
        values      = ["v=spf1 include:mail.contoso.com -all"]
        ttl_seconds = 300
      },
      {
        name        = "_dmarc"
        record_type = "TXT"
        values      = ["v=DMARC1; p=reject; rua=mailto:dmarc@contoso.com"]
        ttl_seconds = 300
      },
      
      # Certificate authority authorization
      {
        name        = "@"
        record_type = "CAA"
        values      = ["letsencrypt.org"]
        ttl_seconds = 86400
      }
    ]
  }
}
```

### API Gateway / Microservices Example

```hcl
module "api_dns_zone" {
  source = "./path/to/module"

  metadata = {
    name = "api.example.com"
    env  = "production"
  }

  spec = {
    zone_name      = "api.example.com"
    resource_group = "prod-network-rg"
    
    records = [
      # API Gateway
      {
        name        = "@"
        record_type = "A"
        values      = ["203.0.113.50"]
        ttl_seconds = 60
      },
      
      # Individual microservices
      {
        name        = "users"
        record_type = "A"
        values      = ["203.0.113.51"]
        ttl_seconds = 60
      },
      {
        name        = "orders"
        record_type = "A"
        values      = ["203.0.113.52"]
        ttl_seconds = 60
      },
      {
        name        = "payments"
        record_type = "A"
        values      = ["203.0.113.53"]
        ttl_seconds = 60
      }
    ]
  }
}
```

## Inputs

### metadata

Object containing resource metadata:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Resource name (typically the domain name) |
| `id` | string | No | - | Unique identifier for the resource |
| `org` | string | No | - | Organization name for tagging |
| `env` | string | No | - | Environment (dev, staging, production) |
| `labels` | map(string) | No | {} | Additional labels |
| `tags` | list(string) | No | [] | Additional tags |
| `version` | object | No | - | Version information |

### spec

Object containing DNS zone specification:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `zone_name` | string | Yes | - | DNS zone name (e.g., "example.com") |
| `resource_group` | string | Yes | - | Azure Resource Group name |
| `records` | list(object) | No | [] | List of DNS records to create |

### spec.records

Each record object has the following fields:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Record name ("@" for zone apex, "www" for subdomain) |
| `record_type` | string | Yes | - | DNS record type (A, AAAA, CNAME, MX, TXT, NS, CAA, SRV, PTR) |
| `values` | list(string) | Yes | - | Record values (IPs, hostnames, text, etc.) |
| `ttl_seconds` | number | No | 60 | Time-to-live in seconds |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | string | Azure Resource Manager ID of the DNS zone |
| `zone_name` | string | Name of the DNS zone |
| `nameservers` | list(string) | List of Azure nameservers assigned to the zone |
| `max_number_of_record_sets` | number | Maximum record sets allowed in the zone |
| `number_of_record_sets` | number | Current number of record sets in the zone |

## DNS Record Types

### Supported Record Types

- **A**: IPv4 address records
- **AAAA**: IPv6 address records
- **CNAME**: Canonical name (alias) records
- **MX**: Mail exchange records
- **TXT**: Text records (SPF, DKIM, DMARC, domain verification)
- **NS**: Name server records (for subdomain delegation)
- **CAA**: Certificate authority authorization records
- **SRV**: Service locator records
- **PTR**: Pointer records (reverse DNS)

### Record-Specific Notes

**CNAME Records:**
- CNAME values should end with a dot (e.g., "target.example.com.")
- Cannot coexist with other record types at the same name
- Cannot be created at the zone apex (@)

**MX Records:**
- Values should be hostnames ending with a dot
- Default preference/priority is 10
- Multiple values create multiple MX records with the same priority

**TXT Records:**
- Used for SPF, DKIM, DMARC, and domain verification
- Each value creates a separate TXT record

**CAA Records:**
- Controls which certificate authorities can issue certificates
- Default flags: 0, tag: "issue"
- Example values: "letsencrypt.org", "digicert.com"

## Post-Deployment Steps

After Terraform creates the DNS zone:

1. **Retrieve Nameservers:**
   ```bash
   terraform output nameservers
   ```

2. **Update Domain Registrar:**
   - Log into your domain registrar (GoDaddy, Namecheap, etc.)
   - Replace current nameservers with Azure nameservers from output
   - Save changes

3. **Verify Delegation:**
   ```bash
   dig NS example.com
   # or
   nslookup -type=NS example.com
   ```

4. **Test DNS Resolution:**
   ```bash
   dig @ns1-01.azure-dns.com example.com
   ```

5. **Wait for Propagation:**
   - DNS changes can take up to 48 hours to propagate globally
   - Most changes propagate within 1-4 hours

## Best Practices

### TTL Values

- **Low TTLs (60-300s)**: Use for records that change frequently or during migrations
- **Medium TTLs (3600s)**: Standard for most records
- **High TTLs (86400s)**: Use for stable records like CAA or NS

### Migration Strategy

When migrating from another DNS provider:

1. Lower TTLs at old provider 48 hours before migration
2. Create zone and records in Azure (this module)
3. Verify all records with `dig @ns1-01.azure-dns.com domain.com`
4. Update nameservers at registrar
5. Monitor for issues for 24-48 hours
6. Raise TTLs back to normal values

### Security

- **Always use CAA records** to restrict certificate issuance
- **Implement SPF, DKIM, DMARC** for email domains
- **Tag resources** with org/env for access control policies
- **Use separate zones** for different environments (dev/staging/prod)

### Tagging Strategy

The module automatically tags all resources with:
- `resource`: "true"
- `resource_id`: From metadata
- `resource_kind`: "azure_dns_zone"
- `resource_name`: From metadata
- `organization`: If provided
- `environment`: If provided

## Troubleshooting

### Zone Creation Fails

**Error:** Resource group not found
- Ensure the resource group exists before running Terraform
- Verify you have permissions to create resources in the resource group

**Error:** Invalid zone name
- Zone names must be valid DNS domain names
- Do not include trailing dots in `zone_name`

### DNS Not Resolving

**Issue:** Records not resolving after delegation
- Wait for DNS propagation (up to 48 hours, usually faster)
- Verify nameservers at registrar match Azure output
- Test with `dig @ns1-01.azure-dns.com domain.com` to query Azure directly

**Issue:** Some records work, others don't
- Check for conflicts (e.g., CNAME at zone apex)
- Verify record values end with dots where required (CNAME, MX)
- Confirm TTL hasn't caused stale caches

### Terraform State Issues

**Issue:** Resource already exists
- Import existing zone: `terraform import azurerm_dns_zone.dns_zone /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/dnszones/{zone}`

## Examples of Common Patterns

### Domain Verification for Azure Services

```hcl
records = [
  {
    name        = "asuid"
    record_type = "TXT"
    values      = ["unique-verification-string"]
    ttl_seconds = 300
  }
]
```

### Subdomain Delegation

```hcl
records = [
  {
    name        = "dev"
    record_type = "NS"
    values      = [
      "ns1-dev.example.com.",
      "ns2-dev.example.com."
    ]
    ttl_seconds = 3600
  }
]
```

### Load Balancer / Multi-Region

```hcl
records = [
  {
    name        = "app"
    record_type = "A"
    values      = [
      "203.0.113.10",  # Region 1
      "198.51.100.20"  # Region 2
    ]
    ttl_seconds = 60  # Low TTL for fast failover
  }
]
```

## Resources Created

This module creates the following Azure resources:

- `azurerm_dns_zone.dns_zone` - The DNS zone itself
- `azurerm_dns_a_record.a_records[*]` - A records (IPv4)
- `azurerm_dns_aaaa_record.aaaa_records[*]` - AAAA records (IPv6)
- `azurerm_dns_cname_record.cname_records[*]` - CNAME records
- `azurerm_dns_mx_record.mx_records[*]` - MX records
- `azurerm_dns_txt_record.txt_records[*]` - TXT records
- `azurerm_dns_ns_record.ns_records[*]` - NS records
- `azurerm_dns_caa_record.caa_records[*]` - CAA records
- `azurerm_dns_srv_record.srv_records[*]` - SRV records
- `azurerm_dns_ptr_record.ptr_records[*]` - PTR records

## Advanced Topics

### DNSSEC

Azure DNS supports DNSSEC signing (GA 2025). To enable:

1. Create zone with this module
2. Enable DNSSEC via Azure CLI:
   ```bash
   az network dns zone update --name example.com --resource-group rg --signing-enabled true
   ```
3. Retrieve DS record and add to registrar

### Integration with Kubernetes

Use **ExternalDNS** to automatically manage records based on Kubernetes Ingresses:

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: myapp.example.com
```

### Monitoring

Enable Azure Monitor for DNS zones:

```hcl
# Add to your Terraform configuration
resource "azurerm_monitor_diagnostic_setting" "dns_logs" {
  name               = "dns-diagnostics"
  target_resource_id = azurerm_dns_zone.dns_zone.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id

  log {
    category = "QueryLog"
    enabled  = true
  }
}
```

## License

This module is part of the Planton infrastructure framework.

## Support

For issues and feature requests, see the main Planton documentation.

