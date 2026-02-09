# OpenStackNetwork Examples

## Minimal Private Network

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: dev-network
spec: {}
```

## Network with Description and MTU

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: overlay-network
spec:
  description: "VXLAN overlay network for development workloads"
  mtu: 1450
```

## Network with DNS Domain

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: dns-enabled-network
spec:
  description: "Network with DNS integration for auto-assigned DNS names"
  dns_domain: "dev.internal.example.com."
```

## Shared Network (Admin-Only)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: shared-services
spec:
  description: "Shared network visible to all tenants"
  shared: true
```

## External Provider Network (Admin-Only)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: external-net
spec:
  description: "External network for floating IP allocation and router gateways"
  external: true
  shared: true
```

## Network with Port Security Disabled

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: promiscuous-network
spec:
  description: "Network with port security disabled for monitoring workloads"
  port_security_enabled: false
```

## Network with Tags

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: tagged-network
spec:
  tags:
    - "team:platform"
    - "env:staging"
    - "managed-by:openmcf"
```

## Network with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: regional-network
spec:
  region: "RegionTwo"
```

## Fully Specified Network

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: production-network
  org: acme-corp
  env: production
  labels:
    team: platform
    cost-center: engineering
spec:
  description: "Production application network for ACME Corp"
  admin_state_up: true
  mtu: 1500
  dns_domain: "prod.acme.internal."
  port_security_enabled: true
  tags:
    - "production"
    - "managed"
    - "acme"
  region: "RegionOne"
```

## CLI Usage

```bash
# Deploy with provider config file
openmcf apply --manifest network.yaml -p openstack-creds.yaml

# Deploy with stored credentials (auto-resolved)
openmcf apply --manifest network.yaml

# Preview changes
openmcf plan --manifest network.yaml -p openstack-creds.yaml

# Destroy
openmcf destroy --manifest network.yaml -p openstack-creds.yaml
```
