# ScalewayInstanceSecurityGroup Examples

## Minimal Example

A security group with default settings (accept-all for both directions). Useful as a starting point or for development environments where you don't need strict firewalling.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: dev-sg
spec:
  zone: fr-par-1
```

## Web Server (Allowlist Model)

A production web server security group that drops all inbound traffic by default and only accepts HTTP, HTTPS, and SSH from a restricted IP range.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: web-server-sg
  org: acme-corp
  env: production
spec:
  zone: fr-par-1
  description: "Web server firewall -- HTTP/HTTPS from anywhere, SSH from office only"
  inbound_default_policy: drop
  inbound_rules:
    - action: accept
      protocol: TCP
      port_range: "80"
      ip_range: "0.0.0.0/0"
    - action: accept
      protocol: TCP
      port_range: "443"
      ip_range: "0.0.0.0/0"
    - action: accept
      protocol: TCP
      port_range: "22"
      ip_range: "203.0.113.0/24"
    - action: accept
      protocol: ICMP
      ip_range: "0.0.0.0/0"
```

## Database Server

A security group for a database instance that only accepts connections from the application tier's CIDR and SSH from a bastion host.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: database-sg
  org: acme-corp
  env: production
spec:
  zone: fr-par-1
  description: "Database firewall -- PostgreSQL from app tier, SSH from bastion"
  inbound_default_policy: drop
  inbound_rules:
    - action: accept
      protocol: TCP
      port_range: "5432"
      ip_range: "10.0.1.0/24"
    - action: accept
      protocol: TCP
      port_range: "22"
      ip_range: "10.0.0.5/32"
```

## Kubernetes Worker Nodes

A security group for Kubernetes worker nodes that allows the API server, NodePort range, and internal cluster communication.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: k8s-workers-sg
  org: acme-corp
  env: staging
spec:
  zone: nl-ams-1
  description: "Kubernetes worker nodes -- API server, NodePorts, cluster internal"
  inbound_default_policy: drop
  inbound_rules:
    # Kubernetes API server
    - action: accept
      protocol: TCP
      port_range: "6443"
      ip_range: "0.0.0.0/0"
    # NodePort range for services
    - action: accept
      protocol: TCP
      port_range: "30000-32767"
      ip_range: "0.0.0.0/0"
    # HTTP ingress
    - action: accept
      protocol: TCP
      port_range: "80"
      ip_range: "0.0.0.0/0"
    # HTTPS ingress
    - action: accept
      protocol: TCP
      port_range: "443"
      ip_range: "0.0.0.0/0"
    # Internal cluster communication (kubelet, etcd, flannel, etc.)
    - action: accept
      protocol: ANY
      ip_range: "10.0.0.0/8"
    # ICMP for network debugging
    - action: accept
      protocol: ICMP
      ip_range: "0.0.0.0/0"
```

## Deny-All with Strict Egress Control

A locked-down security group that controls both inbound AND outbound traffic. Useful for compliance-sensitive workloads that must not reach arbitrary internet destinations.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: restricted-sg
  org: acme-corp
  env: production
spec:
  zone: fr-par-1
  description: "Locked-down SG -- strict ingress and egress"
  inbound_default_policy: drop
  outbound_default_policy: drop
  inbound_rules:
    - action: accept
      protocol: TCP
      port_range: "443"
      ip_range: "10.0.0.0/8"
  outbound_rules:
    # DNS resolution
    - action: accept
      protocol: UDP
      port_range: "53"
      ip_range: "0.0.0.0/0"
    - action: accept
      protocol: TCP
      port_range: "53"
      ip_range: "0.0.0.0/0"
    # HTTPS to external APIs
    - action: accept
      protocol: TCP
      port_range: "443"
      ip_range: "0.0.0.0/0"
    # NTP
    - action: accept
      protocol: UDP
      port_range: "123"
      ip_range: "0.0.0.0/0"
```

## Stateless Security Group

A stateless security group for advanced use cases (e.g., high-throughput packet forwarding or network appliances). In stateless mode, you must define explicit rules for BOTH directions of traffic.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: stateless-sg
spec:
  zone: fr-par-1
  description: "Stateless firewall for network appliance"
  stateful: false
  inbound_default_policy: drop
  outbound_default_policy: drop
  inbound_rules:
    - action: accept
      protocol: TCP
      port_range: "80"
      ip_range: "0.0.0.0/0"
  outbound_rules:
    # Must explicitly allow return traffic in stateless mode
    - action: accept
      protocol: TCP
      port_range: "1024-65535"
      ip_range: "0.0.0.0/0"
```

## Infra Chart valueFrom Reference

When used in an infra chart, the security group is referenced by an Instance resource via `valueFrom`. The infra chart template wires the dependency automatically:

```yaml
# In the infra chart template -- ScalewayInstance resource
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: "{{ values.env }}-web-server"
spec:
  zone: "{{ values.zone }}"
  type: DEV1-S
  security_group_id:
    valueFrom:
      kind: ScalewayInstanceSecurityGroup
      name: "{{ values.env }}-web-sg"
      fieldPath: status.outputs.security_group_id
```

## SMTP-Enabled Security Group

A security group that allows outbound SMTP traffic. Only use this if your Scaleway account is authorized for email sending.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: email-relay-sg
spec:
  zone: fr-par-1
  description: "Email relay -- SMTP enabled"
  enable_default_security: false
  inbound_default_policy: drop
  inbound_rules:
    - action: accept
      protocol: TCP
      port_range: "25"
      ip_range: "10.0.0.0/8"
    - action: accept
      protocol: TCP
      port_range: "587"
      ip_range: "10.0.0.0/8"
```
