# OpenStackLoadBalancerMonitor Examples

## Minimal HTTP Monitor

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: http-health
spec:
  pool_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  type: "HTTP"
  delay: 5
  timeout: 10
  max_retries: 3
```

## HTTP Monitor with Foreign Key Reference

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: http-health
spec:
  pool_id:
    value_from:
      name: web-pool
  type: "HTTP"
  delay: 5
  timeout: 10
  max_retries: 3
```

## HTTP Monitor with URL Path and Expected Codes

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: http-healthz
spec:
  pool_id:
    value_from:
      name: web-pool
  type: "HTTP"
  delay: 10
  timeout: 5
  max_retries: 3
  url_path: "/healthz"
  http_method: "GET"
  expected_codes: "200"
```

## HTTPS Monitor

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: https-health
spec:
  pool_id:
    value_from:
      name: secure-pool
  type: "HTTPS"
  delay: 10
  timeout: 5
  max_retries: 3
  url_path: "/health"
  http_method: "GET"
  expected_codes: "200-299"
```

## PING Monitor

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: ping-check
spec:
  pool_id:
    value_from:
      name: backend-pool
  type: "PING"
  delay: 10
  timeout: 5
  max_retries: 3
```

## TCP Monitor

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: tcp-check
spec:
  pool_id:
    value_from:
      name: tcp-pool
  type: "TCP"
  delay: 5
  timeout: 3
  max_retries: 3
```

## Monitor with Max Retries Down

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: sensitive-monitor
spec:
  pool_id:
    value_from:
      name: web-pool
  type: "HTTP"
  delay: 5
  timeout: 3
  max_retries: 3
  max_retries_down: 2
  url_path: "/ready"
  http_method: "GET"
  expected_codes: "200"
```

## Monitor with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: regional-monitor
spec:
  pool_id:
    value_from:
      name: regional-pool
  type: "HTTP"
  delay: 10
  timeout: 5
  max_retries: 3
  region: "RegionTwo"
```

## Fully Specified HTTP Monitor

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: production-monitor
  org: acme-corp
  env: production
  labels:
    team: platform
    purpose: health-check
spec:
  pool_id:
    value_from:
      name: production-pool
  type: "HTTP"
  delay: 5
  timeout: 10
  max_retries: 3
  max_retries_down: 2
  url_path: "/healthz"
  http_method: "GET"
  expected_codes: "200"
  admin_state_up: true
  region: "RegionOne"
```

## UDP-CONNECT Monitor

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: udp-check
spec:
  pool_id:
    value_from:
      name: udp-pool
  type: "UDP-CONNECT"
  delay: 10
  timeout: 5
  max_retries: 3
```

## CLI Usage

```bash
# Deploy with provider config file
openmcf apply --manifest monitor.yaml -p openstack-creds.yaml

# Deploy with stored credentials (auto-resolved)
openmcf apply --manifest monitor.yaml

# Preview changes
openmcf plan --manifest monitor.yaml -p openstack-creds.yaml

# Destroy
openmcf destroy --manifest monitor.yaml -p openstack-creds.yaml
```
