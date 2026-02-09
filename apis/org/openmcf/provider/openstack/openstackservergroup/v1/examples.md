# OpenStackServerGroup Examples

## 1. Minimal Anti-Affinity Group

Spread instances across different hypervisors for high availability.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: db-anti-affinity
spec:
  policy: anti-affinity
```

## 2. Affinity Group for Performance

Co-locate instances on the same hypervisor for low-latency communication.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: cache-affinity
spec:
  policy: affinity
```

## 3. Soft Anti-Affinity for Best-Effort HA

Prefer spreading instances, but don't fail if all hosts are full.
Requires Nova API 2.15+.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: app-soft-spread
spec:
  policy: soft-anti-affinity
```

## 4. Soft Affinity for Best-Effort Co-location

Prefer co-location, but allow placement on other hosts if needed.
Requires Nova API 2.15+.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: worker-soft-affinity
spec:
  policy: soft-affinity
```

## 5. Production Server Group with Full Metadata

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: prod-db-spread
  org: acme-corp
  env: production
  labels:
    team: platform
    service: database
spec:
  policy: anti-affinity
  region: RegionOne
```

## 6. Multi-Region Deployment

Create a server group in a specific region.

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: eu-west-spread
spec:
  policy: anti-affinity
  region: eu-west-1
```

## 7. InfraChart Usage -- Instance References Server Group

Server group created first, then referenced by instances via `value_from`:

```yaml
# Server group
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: app-ha-group
spec:
  policy: anti-affinity

---
# Instance referencing the server group
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: app-node-1
spec:
  flavor_name: m1.large
  image_name: ubuntu-22.04
  server_group_id:
    value_from:
      name: app-ha-group
  networks:
    - uuid:
        value_from:
          name: app-network
```
