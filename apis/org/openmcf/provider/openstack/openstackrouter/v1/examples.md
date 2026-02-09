# OpenStackRouter Examples

## 1. Internal-Only Router

Minimal router for east-west traffic between subnets (no external access).

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: internal-router
spec: {}
```

## 2. Router with External Gateway (Literal UUID)

Connect to a pre-existing external network using its UUID.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: edge-router
spec:
  external_network_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

## 3. Router with SNAT Enabled

Explicitly enable Source NAT for tenant-to-internet traffic.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: snat-router
spec:
  external_network_id:
    value: "ext-net-uuid"
  enable_snat: true
```

## 4. Router with SNAT Disabled

Disable SNAT -- useful when tenants have public IP ranges or use floating IPs exclusively.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: no-snat-router
spec:
  external_network_id:
    value: "ext-net-uuid"
  enable_snat: false
```

## 5. Distributed Virtual Router (DVR)

Enable DVR mode for distributed routing on compute nodes.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: dvr-router
spec:
  external_network_id:
    value: "ext-net-uuid"
  distributed: true
```

## 6. Router with Specific External IP

Request a specific IP on the external network.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: fixed-ip-router
spec:
  external_network_id:
    value: "ext-net-uuid"
  external_fixed_ips:
    - subnet_id: "ext-subnet-uuid"
      ip_address: "203.0.113.50"
```

## 7. Router with Multiple External IPs

Allocate IPs from multiple subnets on the external network (dual-stack or multi-subnet).

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: multi-ip-router
spec:
  external_network_id:
    value: "ext-net-uuid"
  external_fixed_ips:
    - subnet_id: "ipv4-subnet-uuid"
    - subnet_id: "ipv6-subnet-uuid"
```

## 8. InfraChart Usage (value_from Reference)

Reference an OpenStackNetwork resource by name in an InfraChart.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: dev-router
spec:
  external_network_id:
    value_from:
      name: public-network
```

## 9. Router with Tags and Description

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: tagged-router
spec:
  external_network_id:
    value: "ext-net-uuid"
  description: "Edge router for development team"
  tags:
    - team:platform
    - env:dev
```

## 10. Router with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: regional-router
spec:
  external_network_id:
    value: "ext-net-uuid"
  region: RegionTwo
```

## 11. Admin-Down Router

Create a router in administratively disabled state.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: disabled-router
spec:
  admin_state_up: false
```

## 12. Fully-Specified Production Router

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: prod-edge-router
  org: acme-corp
  env: production
  labels:
    team: platform
spec:
  external_network_id:
    value: "ext-net-uuid"
  admin_state_up: true
  enable_snat: true
  distributed: false
  external_fixed_ips:
    - subnet_id: "ext-subnet-uuid"
      ip_address: "203.0.113.50"
  description: "Production edge router for ACME Corp"
  tags:
    - production
    - managed
  region: RegionOne
```
