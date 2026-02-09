# OpenStackNetworkPort Examples

## Minimal Port (auto-assigned IP from any subnet)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: simple-port
spec:
  network_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

## Port with Specific Subnet and IP

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: web-server-port
spec:
  network_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  fixed_ips:
    - subnet_id:
        value: "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      ip_address: "192.168.1.100"
```

## Port with InfraChart FK References (value_from)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: app-port
spec:
  network_id:
    value_from:
      name: my-network
  fixed_ips:
    - subnet_id:
        value_from:
          name: my-subnet
  security_group_ids:
    - value_from:
        name: app-sg
    - value_from:
        name: ssh-sg
```

## Multi-Homed Port (Two Subnets)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: dual-stack-port
spec:
  network_id:
    value: "net-uuid"
  fixed_ips:
    - subnet_id:
        value: "subnet-v4-uuid"
      ip_address: "192.168.1.50"
    - subnet_id:
        value: "subnet-v6-uuid"
```

## Port with Multiple Security Groups

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: secured-port
spec:
  network_id:
    value: "net-uuid"
  security_group_ids:
    - value: "sg-web-uuid"
    - value: "sg-monitoring-uuid"
    - value: "sg-bastion-uuid"
```

## Port with No Security Groups (Zero-Trust Bypass)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: lb-vip-port
spec:
  network_id:
    value: "net-uuid"
  no_security_groups: true
  description: "Load balancer VIP port - no SGs needed"
```

## Port with Specific MAC Address

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: bonded-port
spec:
  network_id:
    value: "net-uuid"
  mac_address: "fa:16:3e:aa:bb:cc"
  fixed_ips:
    - subnet_id:
        value: "subnet-uuid"
```

## Port with Port Security Disabled

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: appliance-port
spec:
  network_id:
    value: "net-uuid"
  port_security_enabled: false
  no_security_groups: true
  description: "Network appliance port - port security disabled"
```

## Port with Admin State Down

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: standby-port
spec:
  network_id:
    value: "net-uuid"
  admin_state_up: false
  description: "Standby port - administratively down until failover"
```

## Fully-Specified Production Port

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetworkPort
metadata:
  name: prod-web-server
  org: acme-corp
  env: production
  labels:
    team: platform
    service: web
spec:
  network_id:
    value_from:
      name: prod-network
  fixed_ips:
    - subnet_id:
        value_from:
          name: prod-subnet
      ip_address: "10.0.1.100"
  security_group_ids:
    - value_from:
        name: web-sg
    - value_from:
        name: monitoring-sg
  admin_state_up: true
  port_security_enabled: true
  description: "Production web server primary NIC"
  tags:
    - managed-by-planton
    - environment-production
  region: RegionOne
```
