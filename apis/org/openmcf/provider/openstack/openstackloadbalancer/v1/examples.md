# OpenStackLoadBalancer Examples

## Minimal Load Balancer with Literal Subnet ID

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: dev-lb
spec:
  vip_subnet_id:
    value: "e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d"
```

## Load Balancer with Foreign Key Reference to Subnet

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: app-lb
spec:
  vip_subnet_id:
    value_from:
      name: app-subnet
```

## Load Balancer with VIP Address

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: static-vip-lb
spec:
  vip_subnet_id:
    value_from:
      name: app-subnet
  vip_address: "10.0.0.100"
  description: "Load balancer with a specific VIP address"
```

## Load Balancer with Flavor

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: premium-lb
spec:
  vip_subnet_id:
    value_from:
      name: prod-subnet
  flavor_id: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  description: "Premium flavor load balancer with higher bandwidth limits"
```

## Load Balancer with Tags

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: tagged-lb
spec:
  vip_subnet_id:
    value_from:
      name: dev-subnet
  tags:
    - "team:platform"
    - "env:staging"
    - "managed-by:openmcf"
```

## Load Balancer with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: regional-lb
spec:
  vip_subnet_id:
    value_from:
      name: regional-subnet
  region: "RegionTwo"
```

## Fully Specified Load Balancer

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: production-lb
  org: acme-corp
  env: production
  labels:
    team: platform
    cost-center: engineering
spec:
  vip_subnet_id:
    value_from:
      name: production-subnet
  vip_address: "10.0.0.100"
  description: "Production Octavia load balancer for ACME Corp"
  admin_state_up: true
  flavor_id: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  tags:
    - "production"
    - "managed"
    - "acme"
  region: "RegionOne"
```

## CLI Usage

```bash
# Deploy with provider config file
openmcf apply --manifest loadbalancer.yaml -p openstack-creds.yaml

# Deploy with stored credentials (auto-resolved)
openmcf apply --manifest loadbalancer.yaml

# Preview changes
openmcf plan --manifest loadbalancer.yaml -p openstack-creds.yaml

# Destroy
openmcf destroy --manifest loadbalancer.yaml -p openstack-creds.yaml
```
