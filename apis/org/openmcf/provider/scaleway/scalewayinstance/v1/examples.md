# ScalewayInstance Examples

## Minimal Development Instance

The simplest configuration: a small development instance with a public IP for direct SSH access.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: dev-server
spec:
  zone: fr-par-1
  type: DEV1-S
  image: ubuntu_jammy
  publicIp: {}
```

---

## Production Instance (Private Only, No Public IP)

A production instance behind a Load Balancer with no public IP, custom security group, and Private Network attachment.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: web-server-01
  org: my-org
  env: production
spec:
  zone: fr-par-1
  type: PRO2-M
  image: ubuntu_jammy
  securityGroupId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  rootVolume:
    sizeInGb: 50
    volumeType: sbs_volume
  state: started
  protected: true
```

---

## Instance with Cloud-Init Bootstrapping

An instance that installs Docker and starts a web application on first boot.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: app-server
spec:
  zone: fr-par-1
  type: DEV1-M
  image: ubuntu_jammy
  publicIp: {}
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  cloudInit: |
    #!/bin/bash
    set -euo pipefail

    # Install Docker
    apt-get update
    apt-get install -y docker.io docker-compose-plugin
    systemctl enable --now docker

    # Pull and run the application
    docker run -d --restart=always -p 80:8080 myapp:latest
```

---

## Instance with Additional Local Volumes

An instance with extra local SSD storage for data processing.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: data-processor
spec:
  zone: fr-par-1
  type: GP1-S
  image: ubuntu_jammy
  publicIp: {}
  rootVolume:
    sizeInGb: 30
    volumeType: l_ssd
  additionalVolumes:
    - name: data-volume
      volumeType: l_ssd
      sizeInGb: 100
    - name: scratch-space
      volumeType: scratch
      sizeInGb: 50
  cloudInit: |
    #!/bin/bash
    # Mount additional volumes
    mkfs.ext4 /dev/vdb
    mkdir -p /data
    mount /dev/vdb /data
    echo '/dev/vdb /data ext4 defaults 0 0' >> /etc/fstab
```

---

## Bastion Host

A small instance with a public IP and restrictive security group, used as a jump host for SSH access to private instances.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: bastion
  env: production
spec:
  zone: fr-par-1
  type: DEV1-S
  image: ubuntu_jammy
  publicIp:
    reverseDns: bastion.example.com
  securityGroupId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  protected: true
  cloudInit: |
    #!/bin/bash
    # Harden SSH configuration
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
    sed -i 's/#MaxAuthTries 6/MaxAuthTries 3/' /etc/ssh/sshd_config
    systemctl restart sshd
```

---

## Full-Featured Example with All Options

Demonstrates all available spec fields.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: production-web-01
  org: my-org
  env: production
spec:
  zone: fr-par-1
  type: PRO2-M
  image: ubuntu_jammy
  publicIp:
    reverseDns: web-01.example.com
  securityGroupId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  rootVolume:
    sizeInGb: 50
    volumeType: sbs_volume
    deleteOnTermination: true
    sbsIops: 15000
  additionalVolumes:
    - name: app-data
      volumeType: l_ssd
      sizeInGb: 100
  cloudInit: |
    #!/bin/bash
    set -euo pipefail
    apt-get update && apt-get upgrade -y
    apt-get install -y docker.io
    systemctl enable --now docker
  state: started
  protected: true
```

---

## Infra Chart Composition with valueFrom References

Demonstrates how a ScalewayInstance composes with other resources in an infra chart using `valueFrom` references for the security group and Private Network.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: app-server
spec:
  zone: fr-par-1
  type: PRO2-S
  image: ubuntu_jammy
  securityGroupId:
    valueFrom:
      kind: ScalewayInstanceSecurityGroup
      name: web-sg
      fieldPath: status.outputs.security_group_id
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  rootVolume:
    sizeInGb: 30
    volumeType: sbs_volume
  cloudInit: |
    #!/bin/bash
    apt-get update && apt-get install -y docker.io
```

A downstream `ScalewayDnsRecord` can reference the instance's public IP:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: app-dns
spec:
  dnsZoneId:
    valueFrom:
      kind: ScalewayDnsZone
      name: example-zone
  name: app
  type: A
  data:
    valueFrom:
      kind: ScalewayInstance
      name: app-server
      fieldPath: status.outputs.public_ip_address
  ttl: 300
```

A `ScalewayLoadBalancer` can reference the instance's private IP as a backend:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayLoadBalancer
metadata:
  name: app-lb
spec:
  zone: fr-par-1
  type: LB-S
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  backends:
    - name: web
      serverIps:
        - "10.0.1.5"  # ScalewayInstance private IP (hardcoded in this example)
      forwardPort: 80
      forwardProtocol: http
  frontends:
    - name: http
      inboundPort: 80
      backendName: web
```

---

## Stopped Instance (Pre-Provisioned)

An instance provisioned but not started. Useful for capacity reservation or scheduled activation.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: standby-server
spec:
  zone: fr-par-1
  type: PRO2-S
  image: ubuntu_jammy
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  state: stopped
```
