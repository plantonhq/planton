# OpenStackFloatingIp Examples

## Minimal Allocation-Only

Allocate a floating IP from an external network without associating it to any port.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: web-fip
spec:
  floating_network_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

## Allocation with Foreign Key Reference

Reference an external network managed by OpenMCF using `value_from`.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: app-fip
spec:
  floating_network_id:
    value_from:
      name: public-network
```

## Built-In Port Association (Literal Port ID)

Allocate and immediately associate to a specific port.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: web-fip
spec:
  floating_network_id:
    value_from:
      name: external-network
  port_id:
    value: "b2c3d4e5-f6a7-8901-bcde-f12345678901"
```

## Built-In Port Association (FK Reference)

Associate to a port managed by OpenMCF using `value_from`.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: web-fip
spec:
  floating_network_id:
    value_from:
      name: external-network
  port_id:
    value_from:
      name: web-server-port
```

## Port Association with Fixed IP

When the port has multiple IP addresses, specify which one to map.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: db-fip
spec:
  floating_network_id:
    value_from:
      name: external-network
  port_id:
    value_from:
      name: database-port
  fixed_ip: "10.0.1.5"
```

## Specific IP Address Request

Reserve a particular public IP address (useful for DNS pre-configuration).

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: reserved-fip
spec:
  floating_network_id:
    value_from:
      name: external-network
  address: "203.0.113.42"
```

## Allocate from Specific Subnet

Allocate from a particular subnet within the external network.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: subnet-fip
spec:
  floating_network_id:
    value_from:
      name: external-network
  subnet_id: "c3d4e5f6-a7b8-9012-cdef-123456789012"
```

## With Tags

Apply tags for organization and filtering.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: tagged-fip
spec:
  floating_network_id:
    value_from:
      name: external-network
  tags:
    - "team:platform"
    - "env:production"
    - "managed-by:openmcf"
```

## With Region Override

Deploy to a specific region different from the provider default.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: regional-fip
spec:
  floating_network_id:
    value_from:
      name: external-network
  region: "RegionTwo"
```

## Fully Specified

A complete example with all fields populated.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: prod-web-fip
  org: acme-corp
  env: production
  labels:
    team: platform
    service: web-frontend
spec:
  floating_network_id:
    value_from:
      name: public-network
  port_id:
    value_from:
      name: web-frontend-port
  fixed_ip: "10.0.1.10"
  subnet_id: "ext-subnet-uuid"
  address: "203.0.113.50"
  description: "Production web frontend public IP"
  tags:
    - "production"
    - "web-frontend"
    - "managed-by:openmcf"
  region: "RegionOne"
```
