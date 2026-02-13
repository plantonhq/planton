# AzureLoadBalancer Examples

## Minimal Public Load Balancer

A simple public-facing LB with one backend pool, one TCP probe, and one HTTP rule.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: web-lb
spec:
  region: eastus
  resource_group: prod-rg
  name: web-lb
  public_ip_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/publicIPAddresses/web-pip
  backend_pools:
    - name: default
  health_probes:
    - name: http-probe
      protocol: Http
      port: 80
      request_path: /health
  rules:
    - name: http-rule
      protocol: Tcp
      frontend_port: 80
      backend_port: 80
      backend_pool_name: default
      probe_name: http-probe
```

## Internal Load Balancer with Static IP

An internal LB for backend service traffic within a VNet, using a static private IP for DNS stability.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: api-internal-lb
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group: prod-network-rg
  name: api-internal-lb
  subnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/app-subnet
  private_ip_address: "10.0.2.100"
  backend_pools:
    - name: api-servers
  health_probes:
    - name: api-health
      protocol: Https
      port: 443
      request_path: /api/healthz
      interval_in_seconds: 10
      number_of_probes: 3
  rules:
    - name: https-rule
      protocol: Tcp
      frontend_port: 443
      backend_port: 443
      backend_pool_name: api-servers
      probe_name: api-health
      idle_timeout_in_minutes: 10
```

## Web Tier Load Balancer (HTTP + HTTPS)

Public LB serving both HTTP and HTTPS traffic with a shared health probe.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: web-tier-lb
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-rg
  name: web-tier-lb
  public_ip_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/publicIPAddresses/web-pip
  backend_pools:
    - name: web-servers
  health_probes:
    - name: http-health
      protocol: Http
      port: 80
      request_path: /health
      interval_in_seconds: 10
  rules:
    - name: http-rule
      protocol: Tcp
      frontend_port: 80
      backend_port: 80
      backend_pool_name: web-servers
      probe_name: http-health
    - name: https-rule
      protocol: Tcp
      frontend_port: 443
      backend_port: 443
      backend_pool_name: web-servers
      probe_name: http-health
```

## Multi-Pool Load Balancer

Public LB with separate backend pools for web and API tiers, each with its own probe.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: multi-tier-lb
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-rg
  name: multi-tier-lb
  public_ip_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/publicIPAddresses/multi-pip
  backend_pools:
    - name: web-pool
    - name: api-pool
  health_probes:
    - name: web-probe
      protocol: Http
      port: 80
      request_path: /health
    - name: api-probe
      protocol: Https
      port: 8443
      request_path: /api/ready
  rules:
    - name: web-http
      protocol: Tcp
      frontend_port: 80
      backend_port: 80
      backend_pool_name: web-pool
      probe_name: web-probe
    - name: api-https
      protocol: Tcp
      frontend_port: 8443
      backend_port: 8443
      backend_pool_name: api-pool
      probe_name: api-probe
```

## Enterprise: SQL AlwaysOn with Floating IP

Internal LB configured for SQL Server AlwaysOn availability groups, requiring floating IP (Direct Server Return) and outbound SNAT disabled.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: sql-ao-lb
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-data-rg
  name: sql-ao-lb
  subnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-data-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/data-subnet
  private_ip_address: "10.0.3.200"
  backend_pools:
    - name: sql-nodes
  health_probes:
    - name: sql-probe
      protocol: Tcp
      port: 1433
      interval_in_seconds: 5
      number_of_probes: 2
  rules:
    - name: sql-rule
      protocol: Tcp
      frontend_port: 1433
      backend_port: 1433
      backend_pool_name: sql-nodes
      probe_name: sql-probe
      enable_floating_ip: true
      disable_outbound_snat: true
      idle_timeout_in_minutes: 30
```

## Infra Chart Reference (valueFrom)

Using `valueFrom` references for composition in infra charts, where the public IP and resource group are created by other components.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: chart-lb
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: foundation-rg
      fieldPath: status.outputs.resource_group_name
  name: chart-lb
  public_ip_id:
    valueFrom:
      kind: AzurePublicIp
      name: lb-pip
      fieldPath: status.outputs.public_ip_id
  backend_pools:
    - name: default
  health_probes:
    - name: tcp-probe
      protocol: Tcp
      port: 80
  rules:
    - name: http-rule
      protocol: Tcp
      frontend_port: 80
      backend_port: 80
      backend_pool_name: default
      probe_name: tcp-probe
```

## Internal LB with valueFrom (Infra Chart)

Internal LB in an infra chart that references subnet from another component.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: internal-chart-lb
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: foundation-rg
      fieldPath: status.outputs.resource_group_name
  name: internal-chart-lb
  subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: app-subnet
      fieldPath: status.outputs.subnet_id
  backend_pools:
    - name: app-pool
  health_probes:
    - name: app-health
      protocol: Http
      port: 8080
      request_path: /ready
  rules:
    - name: app-rule
      protocol: Tcp
      frontend_port: 8080
      backend_port: 8080
      backend_pool_name: app-pool
      probe_name: app-health
```
