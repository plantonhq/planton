# AzurePublicIp Examples

## Minimal Configuration

The simplest possible Public IP -- just a name, region, and resource group.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: my-pip
spec:
  region: eastus
  resource_group: my-rg
  name: my-public-ip
```

## With DNS Label

Create a Public IP with a stable DNS name at `myapp.eastus.cloudapp.azure.com`.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: app-pip
  org: mycompany
  env: development
spec:
  region: eastus
  resource_group: dev-network-rg
  name: dev-app-pip
  domain_name_label: myapp
```

## Zone-Redundant (Production)

A zone-redundant Public IP spread across all three availability zones for maximum
resilience. Recommended for production workloads.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: prod-gateway-pip
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-network-rg
  name: prod-gateway-pip
  domain_name_label: prod-gateway
  zones:
    - "1"
    - "2"
    - "3"
```

## With Extended Idle Timeout

A Public IP with a 30-minute idle timeout for long-lived connections like WebSocket
or gRPC streaming.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: grpc-pip
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group: prod-app-rg
  name: prod-grpc-pip
  idle_timeout_in_minutes: 30
  zones:
    - "1"
    - "2"
    - "3"
```

## Zonal (Pinned to Single Zone)

A Public IP pinned to a specific availability zone. Use when the attached resource
(e.g., a VM) is also in a specific zone.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: vm-pip
spec:
  region: eastus
  resource_group: my-rg
  name: vm-zone1-pip
  zones:
    - "1"
```

## Infra Chart Wiring: Enterprise Network Foundation

This example shows how a Public IP fits into an enterprise network infrastructure,
with proper `valueFrom` wiring between resources.

### Resource Group (Layer 0)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: network-rg
  org: mycompany
  env: production
spec:
  name: prod-network-rg
  region: eastus
```

### Public IP (Layer 1) -- references resource group

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: gateway-pip
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-gateway-pip
  domain_name_label: prod-gateway
  zones:
    - "1"
    - "2"
    - "3"
  idle_timeout_in_minutes: 10
```

### NAT Gateway (Layer 2) -- references Public IP

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNatGateway
metadata:
  name: nat-gw
  org: mycompany
  env: production
spec:
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  # NAT Gateway can reference this Public IP's ID
  # (shown conceptually -- actual field depends on AzureNatGateway spec)
```

## Multiple Public IPs in One Environment

Enterprise architectures often need multiple Public IPs for different purposes.

### Load Balancer Public IP

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: lb-pip
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-lb-pip
  zones:
    - "1"
    - "2"
    - "3"
```

### Application Gateway Public IP

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: appgw-pip
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-appgw-pip
  domain_name_label: prod-appgw
  zones:
    - "1"
    - "2"
    - "3"
```

## Best Practices

1. **Use zone-redundant in production** -- always specify `zones: ["1", "2", "3"]`
   for production Public IPs to survive availability zone failures

2. **Use DNS labels for stable names** -- `domain_name_label` provides a human-readable
   FQDN that's useful for CNAME records and documentation

3. **Tune idle timeout for your workload** -- the default 4 minutes works for HTTP
   traffic but is too aggressive for WebSocket, gRPC, or database connections

4. **Naming convention** -- use a pattern like `{env}-{purpose}-pip`
   (e.g., `prod-gateway-pip`, `dev-lb-pip`)

5. **One Public IP per purpose** -- don't share Public IPs between unrelated resources.
   Public IPs are cheap; the cost of a $3.65/month static IP is negligible compared
   to the debugging cost of shared-IP routing issues
