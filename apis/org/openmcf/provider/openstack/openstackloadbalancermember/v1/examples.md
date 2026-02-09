# OpenStackLoadBalancerMember Examples

## Minimal Member with Literal Pool ID

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: web-backend-1
spec:
  pool_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  address: "10.0.0.10"
  protocol_port: 8080
```

## Member with Foreign Key Reference to Pool

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: web-backend-1
spec:
  pool_id:
    value_from:
      name: web-pool
  address: "10.0.0.10"
  protocol_port: 8080
```

## Member with Subnet (Cross-Subnet Routing)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: cross-subnet-member
spec:
  pool_id:
    value_from:
      name: web-pool
  address: "10.1.0.10"
  protocol_port: 8080
  subnet_id:
    value_from:
      name: backend-subnet
```

## Member with Custom Weight

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: heavy-backend
spec:
  pool_id:
    value_from:
      name: web-pool
  address: "10.0.0.20"
  protocol_port: 8080
  weight: 10
```

## Draining Member (Weight 0)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: draining-member
spec:
  pool_id:
    value_from:
      name: web-pool
  address: "10.0.0.30"
  protocol_port: 8080
  weight: 0
```

## Disabled Member

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: disabled-member
spec:
  pool_id:
    value_from:
      name: web-pool
  address: "10.0.0.40"
  protocol_port: 8080
  admin_state_up: false
```

## Member with Tags

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: tagged-member
spec:
  pool_id:
    value_from:
      name: web-pool
  address: "10.0.0.10"
  protocol_port: 8080
  tags:
    - "team:platform"
    - "env:staging"
```

## Member with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: regional-member
spec:
  pool_id:
    value_from:
      name: web-pool
  address: "10.0.0.10"
  protocol_port: 8080
  region: "RegionTwo"
```

## Fully Specified Member

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: production-backend-1
  org: acme-corp
  env: production
  labels:
    team: platform
    role: primary
spec:
  pool_id:
    value_from:
      name: production-pool
  address: "10.0.0.10"
  protocol_port: 8080
  subnet_id:
    value_from:
      name: backend-subnet
  weight: 10
  admin_state_up: true
  tags:
    - "production"
    - "managed"
    - "primary"
  region: "RegionOne"
```

## CLI Usage

```bash
# Deploy with provider config file
openmcf apply --manifest member.yaml -p openstack-creds.yaml

# Deploy with stored credentials (auto-resolved)
openmcf apply --manifest member.yaml

# Preview changes
openmcf plan --manifest member.yaml -p openstack-creds.yaml

# Destroy
openmcf destroy --manifest member.yaml -p openstack-creds.yaml
```
