# OpenStackSecurityGroup Examples

## Minimal Security Group (No Rules)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: empty-sg
spec: {}
```

## Web Server Security Group

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: web-sg
spec:
  description: "Web server security group"
  rules:
    - key: allow-http
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 80
      port_range_max: 80
      remote_ip_prefix: "0.0.0.0/0"
    - key: allow-https
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 443
      port_range_max: 443
      remote_ip_prefix: "0.0.0.0/0"
```

## SSH Bastion Security Group (Zero Trust)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: bastion-sg
spec:
  description: "SSH bastion with zero-trust baseline"
  delete_default_rules: true
  rules:
    - key: allow-ssh
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 22
      port_range_max: 22
      remote_ip_prefix: "203.0.113.0/24"
      description: "SSH from office network only"
    - key: egress-dns
      direction: egress
      ethertype: IPv4
      protocol: udp
      port_range_min: 53
      port_range_max: 53
      remote_ip_prefix: "0.0.0.0/0"
      description: "Allow DNS resolution"
    - key: egress-https
      direction: egress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 443
      port_range_max: 443
      remote_ip_prefix: "0.0.0.0/0"
      description: "Allow HTTPS outbound"
```

## Database Security Group (Restricted Access)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: db-sg
spec:
  description: "PostgreSQL database security group"
  rules:
    - key: allow-postgres-from-app
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 5432
      port_range_max: 5432
      remote_ip_prefix: "10.0.1.0/24"
      description: "Allow PostgreSQL from app subnet"
    - key: allow-postgres-from-monitoring
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 5432
      port_range_max: 5432
      remote_group_id: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
      description: "Allow PostgreSQL from monitoring security group"
```

## ICMP Monitoring Security Group

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: icmp-sg
spec:
  description: "Allow ICMP for monitoring and diagnostics"
  rules:
    - key: allow-icmp-echo-request
      direction: ingress
      ethertype: IPv4
      protocol: icmp
      port_range_min: 8
      port_range_max: 0
      remote_ip_prefix: "10.0.0.0/8"
      description: "Allow ping from internal network"
    - key: allow-icmp-echo-reply
      direction: ingress
      ethertype: IPv4
      protocol: icmp
      port_range_min: 0
      port_range_max: 0
      remote_ip_prefix: "10.0.0.0/8"
      description: "Allow ping replies from internal network"
    - key: allow-all-icmp-egress
      direction: egress
      ethertype: IPv4
      protocol: icmp
      remote_ip_prefix: "0.0.0.0/0"
```

## Kubernetes Node Security Group

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: k8s-node-sg
  org: acme-corp
  env: production
spec:
  description: "Kubernetes worker node security group"
  delete_default_rules: true
  rules:
    - key: allow-kubelet
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 10250
      port_range_max: 10250
      remote_ip_prefix: "10.0.0.0/16"
      description: "Kubelet API"
    - key: allow-nodeport
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 30000
      port_range_max: 32767
      remote_ip_prefix: "0.0.0.0/0"
      description: "NodePort services"
    - key: egress-all-ipv4
      direction: egress
      ethertype: IPv4
      description: "Allow all outbound IPv4"
    - key: egress-all-ipv6
      direction: egress
      ethertype: IPv6
      description: "Allow all outbound IPv6"
  tags:
    - "kubernetes"
    - "production"
  region: RegionOne
```

## Stateless High-Performance Security Group

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: high-perf-sg
spec:
  description: "Stateless SG for high-throughput workloads"
  stateful: false
  delete_default_rules: true
  rules:
    - key: ingress-app-port
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 8080
      port_range_max: 8080
      remote_ip_prefix: "10.0.0.0/8"
    - key: egress-app-return
      direction: egress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 1024
      port_range_max: 65535
      remote_ip_prefix: "10.0.0.0/8"
      description: "Return traffic (stateless requires explicit egress)"
```

## Dual-Stack (IPv4 + IPv6) Security Group

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: dual-stack-sg
spec:
  description: "Dual-stack web server"
  rules:
    - key: allow-http-ipv4
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 80
      port_range_max: 80
      remote_ip_prefix: "0.0.0.0/0"
    - key: allow-http-ipv6
      direction: ingress
      ethertype: IPv6
      protocol: tcp
      port_range_min: 80
      port_range_max: 80
      remote_ip_prefix: "::/0"
    - key: allow-https-ipv4
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 443
      port_range_max: 443
      remote_ip_prefix: "0.0.0.0/0"
    - key: allow-https-ipv6
      direction: ingress
      ethertype: IPv6
      protocol: tcp
      port_range_min: 443
      port_range_max: 443
      remote_ip_prefix: "::/0"
```

## Port Range Security Group

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: port-range-sg
spec:
  description: "Service with ephemeral port range"
  rules:
    - key: allow-app-ports
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 8000
      port_range_max: 8100
      remote_ip_prefix: "10.0.0.0/16"
      description: "Application ports 8000-8100"
    - key: allow-udp-range
      direction: ingress
      ethertype: IPv4
      protocol: udp
      port_range_min: 5000
      port_range_max: 5010
      remote_ip_prefix: "10.0.0.0/16"
      description: "UDP service ports"
```
