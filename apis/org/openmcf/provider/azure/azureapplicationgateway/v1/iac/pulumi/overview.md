# AzureApplicationGateway Pulumi Module -- Architecture Overview

## Resource Flow

```
Stack Input (AzureApplicationGatewayStackInput)
  │
  ├── target: AzureApplicationGateway (api + spec + metadata)
  └── provider_config: AzureProviderConfig (credentials)
        │
        ▼
  initializeLocals()
  ├── Extracts resource group name via .GetValue()
  ├── Derives gateway IP config name: "{name}-gw-ip-config"
  ├── Derives frontend IP config name: "{name}-frontend-ip-config"
  ├── Builds Azure tags from metadata
  └── Returns Locals struct
        │
        ▼
  Resources()
  ├── Creates Azure provider (auth via service principal)
  ├── Builds all nested argument blocks:
  │   ├── SKU: Name + Tier (same value), Capacity or Autoscale
  │   ├── GatewayIpConfigurations: [{name, subnet_id}]
  │   ├── FrontendIpConfigurations: [{name, public_ip_address_id}]
  │   ├── FrontendPorts: derived from listeners [{name}-port, port]
  │   ├── BackendAddressPools: [{name, fqdns, ip_addresses}]
  │   ├── BackendHttpSettings: [{name, port, protocol, affinity, timeout, probe}]
  │   ├── HttpListeners: [{name, frontend_ip, frontend_port, protocol, host, ssl_cert}]
  │   ├── RequestRoutingRules: [{name, rule_type=Basic, listener, pool, settings, priority}]
  │   ├── Probes (if any): [{name, protocol, path, host, interval, timeout, threshold}]
  │   ├── SslCertificates (if any): [{name, key_vault_secret_id}]
  │   ├── Identity (if identity_ids): {type=UserAssigned, identity_ids}
  │   └── WafConfiguration (if waf_enabled): {enabled, firewall_mode, OWASP 3.2}
  ├── Creates SINGLE network.ApplicationGateway resource
  │   └── All sub-components are nested arguments (not separate resources)
  └── Exports outputs:
      ├── app_gateway_id (ARM resource ID)
      └── app_gateway_name
```

## Design Decisions

### Single Monolithic Resource

Unlike the Load Balancer (which creates 4 separate Pulumi resources), the Application
Gateway is a single `network.NewApplicationGateway` resource with all sub-components
as nested argument blocks. This matches the Terraform provider's structure where
`azurerm_application_gateway` is one resource with many nested blocks.

### V2 SKU Only

Only Standard_v2 and WAF_v2 SKUs are supported. V1 SKUs are legacy and lack
autoscale, zone redundancy, and modern features. The SKU `name` and `tier`
use the same value (e.g., both set to "Standard_v2").

### Auto-Derived Internal Names

Three internal Azure names are automatically derived:
- Gateway IP config: `"{name}-gw-ip-config"` (associates App GW with subnet)
- Frontend IP config: `"{name}-frontend-ip-config"` (binds public IP)
- Frontend ports: `"{listener_name}-port"` (binds port numbers)

Users never need to specify these names. The IaC module handles the internal plumbing.

### Capacity vs Autoscale

Mutually exclusive. When autoscale is configured:
- SKU capacity is nil (not set)
- AutoscaleConfiguration block is added with min/max

When fixed capacity:
- SKU capacity is set to the spec value
- No AutoscaleConfiguration block

### Basic Routing Only

All routing rules use `rule_type = "Basic"`. Path-based routing (which requires
url_path_map and path_rule blocks) is excluded per 80/20 scoping and can be
added as a v2 enhancement.

### SSL via Key Vault Only

SSL certificates use `KeyVaultSecretId` to reference certificates stored in
Azure Key Vault. This avoids putting PFX data or passwords in manifests.
The Application Gateway must have a user-assigned identity with GET permission
on the Key Vault certificate.

### WAF Configuration

WAF is only configured when `waf_enabled` is true and SKU is "WAF_v2".
The OWASP rule set type and version (3.2) are hardcoded as sensible defaults.
The firewall mode (Detection/Prevention) is configurable.

### Default Handling

Fields with OpenMCF defaults (cookie_based_affinity="Disabled", request_timeout=30,
probe interval=30, timeout=30, unhealthy_threshold=3, waf_mode="Prevention",
enable_http2=false, waf_enabled=false, capacity=2) are resolved by OpenMCF
middleware before the Pulumi module runs. The module uses `.GetXxx()` to access
the resolved values.
