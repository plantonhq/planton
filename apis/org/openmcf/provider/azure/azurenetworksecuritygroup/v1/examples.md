# AzureNetworkSecurityGroup Examples

## Minimal: Empty NSG (Azure Defaults Only)

An NSG with no user-defined rules relies entirely on Azure's implicit defaults:
allow VNet-to-VNet, allow load balancer probes, deny all other inbound, allow all outbound.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: default-nsg
spec:
  region: eastus
  resource_group: my-rg
  name: default-nsg
```

---

## Web Tier: Allow HTTP/HTTPS, Deny Everything Else

A typical web-tier NSG that allows inbound HTTP and HTTPS traffic from the internet
and explicitly denies all other inbound traffic.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: web-tier-nsg
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-network-rg
  name: web-tier-nsg
  security_rules:
    - name: allow-https
      description: "Allow HTTPS from internet"
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destination_port_range: "443"
    - name: allow-http
      description: "Allow HTTP from internet (redirect to HTTPS)"
      priority: 200
      direction: Inbound
      access: Allow
      protocol: Tcp
      destination_port_range: "80"
    - name: deny-all-inbound
      description: "Explicit deny-all as safety net"
      priority: 4096
      direction: Inbound
      access: Deny
      protocol: "*"
      destination_port_range: "*"
```

---

## App Tier: Allow Traffic from Web Tier Only

An application-tier NSG that only allows traffic from the web-tier subnet on specific
ports, blocking all other inbound traffic.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: app-tier-nsg
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-network-rg
  name: app-tier-nsg
  security_rules:
    - name: allow-web-to-app
      description: "Allow traffic from web tier on app port"
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      source_address_prefix: "10.0.1.0/24"
      destination_port_range: "8080"
    - name: allow-health-checks
      description: "Allow Azure Load Balancer health probes"
      priority: 200
      direction: Inbound
      access: Allow
      protocol: Tcp
      source_address_prefix: AzureLoadBalancer
      destination_port_range: "8080"
    - name: deny-all-inbound
      priority: 4096
      direction: Inbound
      access: Deny
      protocol: "*"
      destination_port_range: "*"
```

---

## Data Tier: Database Access from App Tier

A data-tier NSG that restricts database access to the application subnet and
management jump box, with no direct internet access.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: data-tier-nsg
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-network-rg
  name: data-tier-nsg
  security_rules:
    - name: allow-postgres-from-app
      description: "Allow PostgreSQL from app tier"
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      source_address_prefix: "10.0.2.0/24"
      destination_port_range: "5432"
    - name: allow-redis-from-app
      description: "Allow Redis from app tier"
      priority: 200
      direction: Inbound
      access: Allow
      protocol: Tcp
      source_address_prefix: "10.0.2.0/24"
      destination_port_range: "6380"
    - name: allow-ssh-from-mgmt
      description: "Allow SSH from management subnet"
      priority: 300
      direction: Inbound
      access: Allow
      protocol: Tcp
      source_address_prefix: "10.0.255.0/24"
      destination_port_range: "22"
    - name: deny-all-inbound
      priority: 4096
      direction: Inbound
      access: Deny
      protocol: "*"
      destination_port_range: "*"
```

---

## Enterprise: Multi-Source Rule with Address Prefixes (Plural)

When a rule needs to allow traffic from multiple non-contiguous CIDR blocks,
use the plural `source_address_prefixes` field instead of creating multiple rules.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: mgmt-nsg
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-network-rg
  name: mgmt-nsg
  security_rules:
    - name: allow-ssh-from-vpn
      description: "Allow SSH from corporate VPN ranges"
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destination_port_range: "22"
      source_address_prefixes:
        - "10.100.0.0/16"
        - "172.16.0.0/12"
        - "192.168.1.0/24"
    - name: allow-rdp-from-vpn
      description: "Allow RDP from corporate VPN ranges"
      priority: 200
      direction: Inbound
      access: Allow
      protocol: Tcp
      destination_port_range: "3389"
      source_address_prefixes:
        - "10.100.0.0/16"
        - "172.16.0.0/12"
    - name: deny-all-inbound
      priority: 4096
      direction: Inbound
      access: Deny
      protocol: "*"
      destination_port_range: "*"
```

---

## Infra Chart: Using StringValueOrRef for Resource Group

In infra charts, the resource group is typically a reference to a dynamically
created AzureResourceGroup resource.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: "{{ values.env }}-web-nsg"
spec:
  region: "{{ values.region }}"
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: "{{ values.env }}-network-rg"
      fieldPath: status.outputs.resource_group_name
  name: "{{ values.env }}-web-nsg"
  security_rules:
    - name: allow-https
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destination_port_range: "443"
```
