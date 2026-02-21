# HetznerCloudServer Examples

## Minimal Server

The simplest configuration: a shared x86 server running Ubuntu 24.04 in Falkenstein. No SSH keys, no networking customization, no protections. The server receives auto-assigned public IPv4 and IPv6 addresses.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: dev-box
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
```

---

## Server with SSH Key and Cloud-Init

A server with an SSH key for access and a cloud-init script that installs Nginx on first boot. The SSH key is referenced by name (a literal string value). Cloud-init runs once at creation — changing `userData` forces server replacement.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: web-01
  org: acme-corp
  env: staging
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  sshKeys:
    - value: "deploy-key"
  userData: |
    #!/bin/bash
    apt-get update && apt-get install -y nginx
    systemctl enable nginx
    systemctl start nginx
```

---

## Server with Firewall and Private Network

A server attached to a private network and protected by a firewall, using `valueFrom` references to other OpenMCF components. This pattern is typical in infra-chart compositions where resources are deployed together.

The firewall controls inbound traffic at the infrastructure level. The private network provides internal communication between servers without traversing the public internet.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: app-01
  org: acme-corp
  env: production
spec:
  serverType: cpx21
  image: debian-12
  location: fsn1
  sshKeys:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: prod-key
        fieldPath: status.outputs.ssh_key_id
  firewallIds:
    - valueFrom:
        kind: HetznerCloudFirewall
        name: web-firewall
        fieldPath: status.outputs.firewall_id
  networks:
    - networkId:
        valueFrom:
          kind: HetznerCloudNetwork
          name: main-vpc
          fieldPath: status.outputs.network_id
      ip: "10.0.1.10"
```

The `valueFrom` references establish dependency edges in the deployment DAG — the server waits for the SSH key, firewall, and network to be created before provisioning.

---

## HA Server with Placement Group and Protections

A production server with anti-affinity scheduling, automatic backups, and all protection flags enabled. The `keepDisk` flag preserves the ability to downgrade the server type later by preventing irreversible disk upgrades.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: db-primary
  org: acme-corp
  env: production
  labels:
    role: database
spec:
  serverType: ccx13
  image: ubuntu-24.04
  location: fsn1
  sshKeys:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: prod-key
        fieldPath: status.outputs.ssh_key_id
  placementGroupId:
    valueFrom:
      kind: HetznerCloudPlacementGroup
      name: db-spread
      fieldPath: status.outputs.placement_group_id
  firewallIds:
    - valueFrom:
        kind: HetznerCloudFirewall
        name: db-firewall
        fieldPath: status.outputs.firewall_id
  networks:
    - networkId:
        valueFrom:
          kind: HetznerCloudNetwork
          name: main-vpc
          fieldPath: status.outputs.network_id
  userData: |
    #cloud-config
    package_update: true
    packages:
      - postgresql-16
      - postgresql-client-16
  backups: true
  keepDisk: true
  deleteProtection: true
  rebuildProtection: true
  shutdownBeforeDeletion: true
```

---

## Full-Featured Server with Primary IPs and rDNS

A fully configured server demonstrating every feature: stable public IPv4 and IPv6 via Primary IPs, private networking, SSH keys, firewall, placement group, cloud-init, backups, protections, and reverse DNS.

This example shows the complete infra-chart composition pattern. Each companion resource is referenced via `valueFrom`, creating a DAG where the server is provisioned only after all its dependencies exist.

**Note on rDNS:** This example uses `dnsPtr` to set reverse DNS on the server's IPv4. This works because `publicNet` sets `ipv4Enabled: true` without an `ipv4` Primary IP reference — the server receives an auto-assigned IPv4. If `publicNet.ipv4` referenced a Primary IP, the `dnsPtr` field should not be used; manage rDNS on the `HetznerCloudPrimaryIp` component instead.

Companion resource manifests (deployed in the same infra chart):

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: prod-key
  org: acme-corp
  env: production
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..."
```

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: web-spread
  org: acme-corp
  env: production
spec:
  type: spread
```

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: web-firewall
  org: acme-corp
  env: production
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "443"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "10.0.0.0/8"
```

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: main-vpc
  org: acme-corp
  env: production
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      ipRange: "10.0.1.0/24"
      networkZone: eu-central
```

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: web-ipv6
  org: acme-corp
  env: production
spec:
  type: ipv6
  location: fsn1
  deleteProtection: true
```

The server manifest referencing all companion resources:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: web-prod-01
  org: acme-corp
  env: production
  labels:
    role: web-frontend
spec:
  serverType: cpx31
  image: ubuntu-24.04
  location: fsn1
  sshKeys:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: prod-key
        fieldPath: status.outputs.ssh_key_id
  placementGroupId:
    valueFrom:
      kind: HetznerCloudPlacementGroup
      name: web-spread
      fieldPath: status.outputs.placement_group_id
  firewallIds:
    - valueFrom:
        kind: HetznerCloudFirewall
        name: web-firewall
        fieldPath: status.outputs.firewall_id
  publicNet:
    ipv4Enabled: true
    ipv6Enabled: true
    ipv6:
      valueFrom:
        kind: HetznerCloudPrimaryIp
        name: web-ipv6
        fieldPath: status.outputs.primary_ip_id
  networks:
    - networkId:
        valueFrom:
          kind: HetznerCloudNetwork
          name: main-vpc
          fieldPath: status.outputs.network_id
      ip: "10.0.1.20"
      aliasIps:
        - "10.0.1.21"
  userData: |
    #!/bin/bash
    apt-get update && apt-get install -y nginx certbot python3-certbot-nginx
    systemctl enable nginx
  backups: true
  keepDisk: true
  deleteProtection: true
  rebuildProtection: true
  shutdownBeforeDeletion: true
  dnsPtr: web-prod-01.example.com
```

The `valueFrom` references ensure:
1. All dependencies are created before the server is provisioned
2. Correct numeric IDs are passed without manual lookup
3. Replacing a dependency (e.g., rotating an SSH key) propagates to the server on the next apply
