# OpenStackVolume Examples

## Minimal: Blank 10GB Volume

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: data-vol
spec:
  size: 10
```

## With Volume Type and Description

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: db-data
spec:
  size: 100
  description: "PostgreSQL data volume"
  volume_type: "SSD"
```

## With Availability Zone

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: app-storage
spec:
  size: 50
  availability_zone: "az-1"
  volume_type: "HDD"
```

## Bootable Volume from Image (Literal UUID)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: boot-vol
spec:
  size: 20
  image_id:
    value: "c3d4e5f6-a7b8-9012-cdef-123456789012"
```

## Bootable Volume from Image (FK Reference)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: boot-vol
spec:
  size: 20
  image_id:
    value_from:
      name: ubuntu-22-04
```

## Clone from Existing Volume

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: db-data-clone
spec:
  size: 100
  source_vol_id: "b2c3d4e5-f6a7-8901-bcde-f12345678901"
```

## Restore from Snapshot

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: db-data-restored
spec:
  size: 100
  snapshot_id: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

## With Metadata

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: monitored-vol
  org: acme-corp
  env: production
spec:
  size: 200
  volume_type: "SSD"
  metadata:
    backup: "daily"
    team: "platform"
    purpose: "database"
```

## With Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: remote-vol
spec:
  size: 50
  region: "RegionTwo"
```

## InfraChart Context: Volume with Instance Attachment

```yaml
# 1. Create the volume
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: app-data
spec:
  size: 100
  volume_type: "SSD"
---
# 2. Create the instance (references network, keypair, etc.)
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: app-server
spec:
  flavor_name: "m1.large"
  image_name: "ubuntu-22.04"
  key_pair:
    value_from:
      name: dev-key
  networks:
    - uuid:
        value_from:
          name: dev-network
  security_groups:
    - value_from:
        name: dev-sg
---
# 3. Attach the volume to the instance (DAG-visible join)
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: app-data-attach
spec:
  instance_id:
    value_from:
      name: app-server
  volume_id:
    value_from:
      name: app-data
```
