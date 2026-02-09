# OpenStackLoadBalancerPool Examples

## Minimal Pool with Literal Listener ID

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: web-pool
spec:
  listener_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  protocol: "HTTP"
  lb_method: "ROUND_ROBIN"
```

## Pool with Foreign Key Reference to Listener

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: app-pool
spec:
  listener_id:
    value_from:
      name: http-listener
  protocol: "HTTP"
  lb_method: "LEAST_CONNECTIONS"
```

## Pool with Session Persistence (SOURCE_IP)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: sticky-pool
spec:
  listener_id:
    value_from:
      name: http-listener
  protocol: "HTTP"
  lb_method: "ROUND_ROBIN"
  persistence:
    type: "SOURCE_IP"
```

## Pool with App Cookie Persistence

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: cookie-pool
spec:
  listener_id:
    value_from:
      name: https-listener
  protocol: "HTTPS"
  lb_method: "ROUND_ROBIN"
  persistence:
    type: "APP_COOKIE"
    cookie_name: "JSESSIONID"
```

## TCP Pool with Source IP Stickiness

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: tcp-pool
spec:
  listener_id:
    value_from:
      name: tcp-listener
  protocol: "TCP"
  lb_method: "SOURCE_IP"
  description: "TCP pool with client IP stickiness"
```

## Pool with Tags

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: tagged-pool
spec:
  listener_id:
    value_from:
      name: http-listener
  protocol: "HTTP"
  lb_method: "ROUND_ROBIN"
  tags:
    - "team:platform"
    - "env:staging"
    - "managed-by:openmcf"
```

## Pool with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: regional-pool
spec:
  listener_id:
    value_from:
      name: regional-listener
  protocol: "HTTP"
  lb_method: "ROUND_ROBIN"
  region: "RegionTwo"
```

## Pool in Disabled State

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: maintenance-pool
spec:
  listener_id:
    value_from:
      name: http-listener
  protocol: "HTTP"
  lb_method: "ROUND_ROBIN"
  admin_state_up: false
  description: "Pool disabled for maintenance"
```

## Fully Specified Pool

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: production-pool
  org: acme-corp
  env: production
  labels:
    team: platform
    cost-center: engineering
spec:
  listener_id:
    value_from:
      name: production-listener
  protocol: "HTTP"
  lb_method: "LEAST_CONNECTIONS"
  persistence:
    type: "APP_COOKIE"
    cookie_name: "JSESSIONID"
  description: "Production backend pool for ACME Corp web application"
  admin_state_up: true
  tags:
    - "production"
    - "managed"
    - "acme"
  region: "RegionOne"
```

## CLI Usage

```bash
# Deploy with provider config file
openmcf apply --manifest pool.yaml -p openstack-creds.yaml

# Deploy with stored credentials (auto-resolved)
openmcf apply --manifest pool.yaml

# Preview changes
openmcf plan --manifest pool.yaml -p openstack-creds.yaml

# Destroy
openmcf destroy --manifest pool.yaml -p openstack-creds.yaml
```
