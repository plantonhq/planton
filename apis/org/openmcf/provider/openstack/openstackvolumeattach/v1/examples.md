# OpenStackVolumeAttach Examples

## Minimal: Attach with Literal UUIDs

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: data-attach
spec:
  instance_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  volume_id:
    value: "b2c3d4e5-f6a7-8901-bcde-f12345678901"
```

## With FK References (InfraChart Pattern)

```yaml
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

## With Explicit Device Path

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: db-data-attach
spec:
  instance_id:
    value_from:
      name: db-server
  volume_id:
    value_from:
      name: db-data
  device: "/dev/vdb"
```

## With Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: remote-attach
spec:
  instance_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  volume_id:
    value: "b2c3d4e5-f6a7-8901-bcde-f12345678901"
  region: "RegionTwo"
```

## Full InfraChart Context: Volume + Instance + Attachment

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
# 2. Create the instance
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
# 3. Attach the volume (DAG: waits for both to be created)
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

## Multiple Volumes on One Instance

```yaml
# First volume attachment
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: data-attach
spec:
  instance_id:
    value_from:
      name: db-server
  volume_id:
    value_from:
      name: db-data
  device: "/dev/vdb"
---
# Second volume attachment (logs)
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: logs-attach
spec:
  instance_id:
    value_from:
      name: db-server
  volume_id:
    value_from:
      name: db-logs
  device: "/dev/vdc"
```
