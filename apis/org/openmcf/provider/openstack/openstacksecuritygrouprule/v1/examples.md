# OpenStackSecurityGroupRule Examples

## 1. Allow SSH from anywhere (simplest rule)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-ssh
spec:
  security_group_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  port_range_min: 22
  port_range_max: 22
  remote_ip_prefix: "0.0.0.0/0"
```

## 2. Allow HTTPS from private network

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-https-private
spec:
  security_group_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  port_range_min: 443
  port_range_max: 443
  remote_ip_prefix: "10.0.0.0/8"
  description: "Allow HTTPS from RFC1918 private address space"
```

## 3. Allow all egress IPv4

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: egress-all-ipv4
spec:
  security_group_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  direction: egress
  ethertype: IPv4
```

## 4. Allow ICMP ping

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-ping
spec:
  security_group_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  direction: ingress
  ethertype: IPv4
  protocol: icmp
  port_range_min: 8
  port_range_max: 0
  remote_ip_prefix: "0.0.0.0/0"
  description: "Allow ICMP Echo Request (ping)"
```

## 5. Cross-SG rule with value_from (InfraChart DAG pattern)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-ssh-from-bastion
spec:
  security_group_id:
    value_from:
      name: app-sg
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  port_range_min: 22
  port_range_max: 22
  remote_group_id:
    value_from:
      name: bastion-sg
  description: "Allow SSH from bastion hosts"
```

## 6. Self-referencing rule (allow internal SG traffic)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-internal-traffic
spec:
  security_group_id:
    value_from:
      name: app-sg
  direction: ingress
  ethertype: IPv4
  remote_group_id:
    value_from:
      name: app-sg
  description: "Allow all traffic between instances in the same SG"
```

## 7. UDP port range (DNS)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-dns-udp
spec:
  security_group_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  direction: ingress
  ethertype: IPv4
  protocol: udp
  port_range_min: 53
  port_range_max: 53
  remote_ip_prefix: "192.168.0.0/16"
  description: "Allow DNS queries from internal network"
```

## 8. IPv6 egress rule

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: egress-all-ipv6
spec:
  security_group_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  direction: egress
  ethertype: IPv6
  description: "Allow all egress IPv6 traffic"
```

## 9. Kubernetes NodePort range

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-k8s-nodeports
spec:
  security_group_id:
    value_from:
      name: k8s-worker-sg
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  port_range_min: 30000
  port_range_max: 32767
  remote_ip_prefix: "0.0.0.0/0"
  description: "Allow Kubernetes NodePort range"
```

## 10. Region override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-ssh-regiontwo
spec:
  security_group_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  port_range_min: 22
  port_range_max: 22
  remote_ip_prefix: "0.0.0.0/0"
  region: "RegionTwo"
```
