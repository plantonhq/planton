# AzureLoadBalancer Terraform Module

Terraform implementation for the AzureLoadBalancer deployment component.

## Resources Created

| Resource | Type | Description |
|----------|------|-------------|
| Load Balancer | `azurerm_lb` | Standard SKU LB with frontend config |
| Backend Pools | `azurerm_lb_backend_address_pool` | One per pool |
| Health Probes | `azurerm_lb_probe` | One per probe |
| Rules | `azurerm_lb_rule` | One per load balancing rule |

## Usage

```hcl
module "load_balancer" {
  source = "./path/to/module"

  metadata = {
    name = "my-lb"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "prod-rg"
    name           = "my-lb"
    public_ip_id   = "/subscriptions/.../publicIPAddresses/my-pip"

    backend_pools = [
      { name = "default" }
    ]

    health_probes = [
      {
        name     = "http-probe"
        protocol = "Http"
        port     = 80
        request_path = "/health"
      }
    ]

    rules = [
      {
        name              = "http-rule"
        protocol          = "Tcp"
        frontend_port     = 80
        backend_port      = 80
        backend_pool_name = "default"
        probe_name        = "http-probe"
      }
    ]
  }
}
```
