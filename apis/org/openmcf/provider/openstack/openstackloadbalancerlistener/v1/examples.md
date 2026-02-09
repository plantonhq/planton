# OpenStackLoadBalancerListener Examples

## Minimal HTTP Listener with Literal Load Balancer ID

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: http-listener
spec:
  loadbalancer_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  protocol: "HTTP"
  protocol_port: 80
```

## Listener with Foreign Key Reference to Load Balancer

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: http-listener
spec:
  loadbalancer_id:
    value_from:
      name: app-lb
  protocol: "HTTP"
  protocol_port: 80
```

## HTTPS Pass-Through Listener

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: https-passthrough
spec:
  loadbalancer_id:
    value_from:
      name: app-lb
  protocol: "HTTPS"
  protocol_port: 443
  description: "HTTPS pass-through listener (TLS terminated at backend)"
```

## TERMINATED_HTTPS Listener with TLS Termination

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: tls-listener
spec:
  loadbalancer_id:
    value_from:
      name: app-lb
  protocol: "TERMINATED_HTTPS"
  protocol_port: 443
  default_tls_container_ref: "https://barbican.example.com/v1/secrets/cert-abc-123"
  insert_headers:
    X-Forwarded-For: "true"
    X-Forwarded-Proto: "true"
  description: "TLS termination at load balancer"
```

## TCP Listener for Database Traffic

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: mysql-listener
spec:
  loadbalancer_id:
    value_from:
      name: db-lb
  protocol: "TCP"
  protocol_port: 3306
  allowed_cidrs:
    - "10.0.0.0/8"
  description: "MySQL TCP listener restricted to internal network"
```

## Listener with Connection Limit and Allowed CIDRs

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: api-listener
spec:
  loadbalancer_id:
    value_from:
      name: api-lb
  protocol: "HTTP"
  protocol_port: 8080
  connection_limit: 10000
  allowed_cidrs:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
  description: "API listener with connection limit and CIDR restrictions"
```

## Listener with Tags

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: tagged-listener
spec:
  loadbalancer_id:
    value_from:
      name: app-lb
  protocol: "HTTP"
  protocol_port: 80
  tags:
    - "team:platform"
    - "env:staging"
    - "managed-by:openmcf"
```

## Fully Specified Listener

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: production-https-listener
  org: acme-corp
  env: production
  labels:
    team: platform
    cost-center: engineering
spec:
  loadbalancer_id:
    value_from:
      name: production-lb
  protocol: "TERMINATED_HTTPS"
  protocol_port: 443
  description: "Production HTTPS listener with TLS termination for ACME Corp"
  connection_limit: 50000
  default_tls_container_ref: "https://barbican.example.com/v1/secrets/prod-cert-xyz"
  insert_headers:
    X-Forwarded-For: "true"
    X-Forwarded-Proto: "true"
    X-Forwarded-Port: "true"
  allowed_cidrs:
    - "10.0.0.0/8"
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
openmcf apply --manifest listener.yaml -p openstack-creds.yaml

# Deploy with stored credentials (auto-resolved)
openmcf apply --manifest listener.yaml

# Preview changes
openmcf plan --manifest listener.yaml -p openstack-creds.yaml

# Destroy
openmcf destroy --manifest listener.yaml -p openstack-creds.yaml
```
