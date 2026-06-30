# AzureApplicationGateway Terraform Module

Terraform IaC module for provisioning an Azure Application Gateway with backend pools,
HTTP settings, listeners, routing rules, health probes, SSL certificates, and
optional WAF configuration.

## Architecture

Uses a single `azurerm_application_gateway` resource with dynamic blocks for all
repeated sub-components. This matches the Terraform provider's structure where the
Application Gateway is a monolithic resource.

## Usage

```hcl
module "app_gateway" {
  source = "./path/to/module"

  metadata = {
    name = "my-agw"
  }

  spec = {
    region         = "eastus"
    resource_group = "my-rg"
    name           = "my-agw"
    subnet_id      = "/subscriptions/.../subnets/agw-subnet"
    public_ip_id   = "/subscriptions/.../publicIPAddresses/agw-pip"
    sku            = "Standard_v2"

    backend_address_pools = [{
      name  = "default"
      fqdns = ["backend.contoso.com"]
    }]

    backend_http_settings = [{
      name     = "http-settings"
      port     = 80
      protocol = "Http"
    }]

    http_listeners = [{
      name     = "http-listener"
      port     = 80
      protocol = "Http"
    }]

    request_routing_rules = [{
      name                        = "http-rule"
      http_listener_name          = "http-listener"
      backend_address_pool_name   = "default"
      backend_http_settings_name  = "http-settings"
      priority                    = 100
    }]
  }
}
```

## Key Design Notes

- **Single resource**: One `azurerm_application_gateway` with dynamic blocks
- **Auto-derived names**: Gateway IP config, frontend IP config, and frontend port
  names are derived in `locals.tf`
- **V2 SKU only**: Standard_v2 or WAF_v2
- **Basic routing**: Hardcoded `rule_type = "Basic"` (no path-based routing)
- **SSL via Key Vault**: Certificates use `key_vault_secret_id`
- **WAF**: OWASP 3.2 rule set when `waf_enabled = true`
- **Autoscale vs capacity**: Mutually exclusive, determined by `autoscale` presence

## Outputs

| Output | Description |
|--------|-------------|
| `app_gateway_id` | Azure Resource Manager ID of the Application Gateway |
| `app_gateway_name` | Name of the Application Gateway |
