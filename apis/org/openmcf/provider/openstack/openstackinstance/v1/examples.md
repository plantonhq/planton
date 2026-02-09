# OpenStackInstance Examples

## 1. Minimal Instance (Image Boot)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: dev-vm
spec:
  flavor_name: m1.small
  image_name: ubuntu-22.04
  networks:
    - uuid:
        value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

## 2. Instance with Flavor ID

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: by-flavor-id
spec:
  flavor_id: "12345678-abcd-efgh-ijkl-mnopqrstuvwx"
  image_name: centos-9
  networks:
    - uuid:
        value: "network-uuid"
```

## 3. Instance with SSH Keypair

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: ssh-enabled-vm
spec:
  flavor_name: m1.medium
  image_name: ubuntu-22.04
  key_pair:
    value: my-ssh-key
  networks:
    - uuid:
        value: "network-uuid"
```

## 4. Instance with Security Groups (InfraChart)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: secured-vm
spec:
  flavor_name: m1.medium
  image_name: ubuntu-22.04
  key_pair:
    value_from:
      name: dev-keypair
  networks:
    - uuid:
        value_from:
          name: dev-network
  security_groups:
    - value_from:
        name: ssh-sg
    - value_from:
        name: web-sg
```

## 5. Instance with Pre-provisioned Port

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: stable-ip-vm
spec:
  flavor_name: m1.large
  image_name: ubuntu-22.04
  networks:
    - port:
        value_from:
          name: app-port
      access_network: true
```

## 6. Boot from Volume (Persistent Root Disk)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: persistent-vm
spec:
  flavor_name: m1.large
  networks:
    - uuid:
        value_from:
          name: app-network
  block_device:
    - source_type: image
      uuid: "glance-image-uuid"
      destination_type: volume
      boot_index: 0
      volume_size: 50
      delete_on_termination: true
      volume_type: SSD
```

## 7. Instance with Additional Data Volume

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: vm-with-data-disk
spec:
  flavor_name: m1.xlarge
  image_name: ubuntu-22.04
  networks:
    - uuid:
        value: "network-uuid"
  block_device:
    - source_type: image
      uuid: "image-uuid"
      destination_type: volume
      boot_index: 0
      volume_size: 30
      delete_on_termination: true
    - source_type: blank
      destination_type: volume
      boot_index: -1
      volume_size: 200
      volume_type: SSD
      delete_on_termination: false
```

## 8. Instance with Cloud-Init User Data

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: auto-configured-vm
spec:
  flavor_name: m1.medium
  image_name: ubuntu-22.04
  config_drive: true
  networks:
    - uuid:
        value: "network-uuid"
  user_data: |
    #cloud-config
    package_update: true
    packages:
      - nginx
      - docker.io
    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
```

## 9. Instance with Server Group (Anti-Affinity HA)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: db-replica-1
spec:
  flavor_name: m1.xlarge
  image_name: ubuntu-22.04
  server_group_id:
    value_from:
      name: db-anti-affinity
  networks:
    - uuid:
        value_from:
          name: db-network
  availability_zone: nova
```

## 10. Instance with Multiple Networks

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: multi-nic-vm
spec:
  flavor_name: m1.large
  image_name: ubuntu-22.04
  networks:
    - uuid:
        value_from:
          name: public-network
      access_network: true
    - uuid:
        value_from:
          name: private-network
      fixed_ip_v4: "10.0.1.100"
```

## 11. Fully-Specified Production Instance

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: prod-app-server
  org: acme-corp
  env: production
  labels:
    team: platform
    service: api-gateway
spec:
  flavor_name: m1.xlarge
  key_pair:
    value_from:
      name: prod-keypair
  networks:
    - port:
        value_from:
          name: app-port
      access_network: true
  security_groups:
    - value_from:
        name: web-sg
    - value_from:
        name: monitoring-sg
  block_device:
    - source_type: image
      uuid: "golden-image-uuid"
      destination_type: volume
      boot_index: 0
      volume_size: 50
      delete_on_termination: true
      volume_type: SSD
    - source_type: blank
      destination_type: volume
      boot_index: -1
      volume_size: 500
      volume_type: SSD
      delete_on_termination: false
  user_data: |
    #cloud-config
    package_update: true
    packages:
      - nginx
      - prometheus-node-exporter
  metadata:
    role: api-gateway
    environment: production
    managed_by: planton
  config_drive: true
  server_group_id:
    value_from:
      name: app-ha-group
  availability_zone: nova
  tags:
    - production
    - managed
    - api-gateway
  region: RegionOne
```
