# OpenStackSubnet Examples

## Minimal Subnet with Literal Network ID

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: dev-subnet
spec:
  network_id:
    value: "e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d"
  cidr: "192.168.1.0/24"
```

## Subnet with Foreign Key Reference to Network

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: app-subnet
spec:
  network_id:
    value_from:
      name: my-network
  cidr: "10.0.0.0/16"
```

## Subnet with DNS Nameservers

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: dns-subnet
spec:
  network_id:
    value_from:
      name: dev-network
  cidr: "192.168.10.0/24"
  dns_nameservers:
    - "8.8.8.8"
    - "8.8.4.4"
```

## Subnet with Custom Gateway

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: gateway-subnet
spec:
  network_id:
    value_from:
      name: prod-network
  cidr: "10.10.0.0/24"
  gateway_ip: "10.10.0.254"
```

## Isolated Subnet (No Gateway)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: storage-subnet
spec:
  network_id:
    value_from:
      name: storage-network
  cidr: "172.16.0.0/24"
  no_gateway: true
  description: "Isolated storage network -- no routing needed"
```

## Subnet with DHCP Disabled

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: static-subnet
spec:
  network_id:
    value_from:
      name: mgmt-network
  cidr: "10.20.0.0/24"
  enable_dhcp: false
  description: "Management subnet with static IPs only"
```

## Subnet with Allocation Pools

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: pooled-subnet
spec:
  network_id:
    value_from:
      name: app-network
  cidr: "10.0.0.0/16"
  gateway_ip: "10.0.0.1"
  allocation_pools:
    - start: "10.0.1.0"
      end: "10.0.1.255"
    - start: "10.0.2.0"
      end: "10.0.2.255"
  description: "Subnet with reserved ranges -- 10.0.0.x reserved for network appliances"
```

## IPv6 Subnet

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: ipv6-subnet
spec:
  network_id:
    value_from:
      name: dual-stack-network
  cidr: "2001:db8::/64"
  ip_version: 6
```

## Subnet with Tags

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: tagged-subnet
spec:
  network_id:
    value_from:
      name: dev-network
  cidr: "192.168.50.0/24"
  tags:
    - "team:platform"
    - "env:staging"
    - "managed-by:openmcf"
```

## Subnet with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: regional-subnet
spec:
  network_id:
    value_from:
      name: multi-region-network
  cidr: "10.100.0.0/24"
  region: "RegionTwo"
```

## Fully Specified Subnet

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: production-subnet
  org: acme-corp
  env: production
  labels:
    team: platform
    cost-center: engineering
spec:
  network_id:
    value_from:
      name: production-network
  cidr: "10.0.0.0/16"
  ip_version: 4
  gateway_ip: "10.0.0.1"
  enable_dhcp: true
  dns_nameservers:
    - "10.0.0.2"
    - "10.0.0.3"
  allocation_pools:
    - start: "10.0.1.0"
      end: "10.0.127.255"
    - start: "10.0.128.0"
      end: "10.0.254.255"
  description: "Production subnet for ACME Corp -- serves all application workloads"
  tags:
    - "production"
    - "managed"
    - "acme"
  region: "RegionOne"
```

## CLI Usage

```bash
# Deploy with provider config file
openmcf apply --manifest subnet.yaml -p openstack-creds.yaml

# Deploy with stored credentials (auto-resolved)
openmcf apply --manifest subnet.yaml

# Preview changes
openmcf plan --manifest subnet.yaml -p openstack-creds.yaml

# Destroy
openmcf destroy --manifest subnet.yaml -p openstack-creds.yaml
```
