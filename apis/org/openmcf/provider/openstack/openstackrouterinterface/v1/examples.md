# OpenStackRouterInterface Examples

## 1. Basic Router-Subnet Attachment (Literal UUIDs)

Attach a router to a subnet using their UUIDs directly.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: main-router-subnet
spec:
  router_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  subnet_id:
    value: "b2c3d4e5-f6a7-8901-bcde-f12345678901"
```

## 2. InfraChart Usage (Both value_from)

Reference both router and subnet by name in an InfraChart. The DAG engine resolves the UUIDs.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: dev-router-subnet
spec:
  router_id:
    value_from:
      name: dev-router
  subnet_id:
    value_from:
      name: dev-subnet
```

## 3. Mixed Mode: Router Literal, Subnet Reference

Use a pre-existing router UUID but reference a subnet managed in the same InfraChart.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: shared-router-to-new-subnet
spec:
  router_id:
    value: "existing-router-uuid"
  subnet_id:
    value_from:
      name: team-subnet
```

## 4. Mixed Mode: Router Reference, Subnet Literal

Reference a router managed in the InfraChart but use a pre-existing subnet UUID.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: new-router-to-shared-subnet
spec:
  router_id:
    value_from:
      name: team-router
  subnet_id:
    value: "existing-subnet-uuid"
```

## 5. With Region Override

Attach in a specific region (overriding the provider default).

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: regional-attachment
spec:
  router_id:
    value: "router-uuid"
  subnet_id:
    value: "subnet-uuid"
  region: RegionTwo
```

## 6. Developer Environment Pattern

Typical usage in the `openstack/developer-environment` InfraChart -- connects the developer's isolated subnet to the edge router.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: dev-env-router-interface
spec:
  router_id:
    value_from:
      name: dev-edge-router
  subnet_id:
    value_from:
      name: dev-subnet
```

## 7. Multiple Subnets on One Router

Attach multiple subnets to the same router by creating multiple router interface resources.

```yaml
# First subnet
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: router-app-subnet
spec:
  router_id:
    value_from:
      name: main-router
  subnet_id:
    value_from:
      name: app-subnet
---
# Second subnet
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: router-db-subnet
spec:
  router_id:
    value_from:
      name: main-router
  subnet_id:
    value_from:
      name: db-subnet
```

## 8. Fully-Specified with Rich Metadata

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: prod-router-interface
  org: acme-corp
  env: production
  labels:
    team: platform
    tier: networking
spec:
  router_id:
    value: "prod-router-uuid"
  subnet_id:
    value: "prod-subnet-uuid"
  region: RegionOne
```
