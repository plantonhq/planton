# OpenStackFloatingIpAssociate Examples

## Minimal Association (Literal Values)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: web-fip-assoc
spec:
  floating_ip:
    value: "203.0.113.42"
  port_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

## InfraChart FK References (value_from)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: web-fip-assoc
spec:
  floating_ip:
    value_from:
      name: web-floating-ip
  port_id:
    value_from:
      name: web-server-port
```

## Association with Specific Fixed IP (Multi-IP Port)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: web-fip-assoc
spec:
  floating_ip:
    value: "203.0.113.42"
  port_id:
    value: "port-uuid-1234"
  fixed_ip: "192.168.1.100"
```

## Association with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: regional-fip-assoc
spec:
  floating_ip:
    value_from:
      name: regional-fip
  port_id:
    value_from:
      name: regional-port
  region: RegionTwo
```

## Fully-Specified Production Association

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: prod-web-fip-assoc
  org: acme-corp
  env: production
  labels:
    team: platform
    service: web
spec:
  floating_ip:
    value_from:
      name: prod-web-fip
  port_id:
    value_from:
      name: prod-web-port
  fixed_ip: "10.0.1.100"
  region: RegionOne
```

## Using Floating IP UUID Instead of Address

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: uuid-based-assoc
spec:
  floating_ip:
    value: "c3d4e5f6-a7b8-9012-cdef-123456789012"
  port_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

## InfraChart DAG Example (Full Context)

This shows the three components working together in an InfraChart:

```yaml
# Step 1: Allocate a floating IP (no port association)
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: web-fip
spec:
  floating_network_id:
    value: "external-net-uuid"
---
# Step 2: Create a port on the tenant network
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: web-port
spec:
  network_id:
    value_from:
      name: tenant-network
  fixed_ips:
    - subnet_id:
        value_from:
          name: tenant-subnet
---
# Step 3: Associate the floating IP with the port (DAG-visible)
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: web-fip-assoc
spec:
  floating_ip:
    value_from:
      name: web-fip
  port_id:
    value_from:
      name: web-port
```
