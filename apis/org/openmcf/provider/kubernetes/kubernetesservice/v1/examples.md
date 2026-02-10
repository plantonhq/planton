# KubernetesService Examples

Complete, copy-paste ready examples for deploying Kubernetes Services with OpenMCF.

## 1. Minimal ClusterIP Service

The simplest service configuration. Exposes pods internally within the cluster.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: backend-api
spec:
  namespace: default
  name: backend-api
  selector:
    app: backend
  ports:
    - port: 80
      target_port: "8080"
```

```bash
openmcf pulumi up --manifest minimal-clusterip.yaml
```

## 2. Multi-Port ClusterIP Service

A service exposing multiple ports, common for applications serving HTTP and gRPC.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: api-gateway
spec:
  namespace: production
  name: api-gateway
  selector:
    app: api-gateway
    tier: frontend
  ports:
    - name: http
      port: 80
      target_port: "8080"
    - name: grpc
      port: 9090
      target_port: "9090"
      protocol: TCP
  labels:
    team: platform
    environment: production
```

```bash
openmcf pulumi up --manifest multi-port.yaml
```

## 3. NodePort Service

Exposes the service on a static port on each cluster node. Useful for development and non-cloud environments.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: web-nodeport
spec:
  namespace: staging
  name: web-nodeport
  type: node_port
  selector:
    app: web
  ports:
    - name: http
      port: 80
      target_port: "8080"
      node_port: 30080
  external_traffic_policy: local
```

```bash
openmcf pulumi up --manifest nodeport.yaml
```

## 4. AWS Network Load Balancer

A production LoadBalancer service using AWS NLB with source IP preservation and IP-based access control.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: production-lb
spec:
  namespace: production
  name: production-lb
  type: load_balancer
  selector:
    app: web
    tier: frontend
  ports:
    - name: https
      port: 443
      target_port: "8443"
  external_traffic_policy: local
  session_affinity: client_ip
  load_balancer_source_ranges:
    - "10.0.0.0/8"
    - "203.0.113.0/24"
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
    service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
    service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: "true"
```

```bash
openmcf pulumi up --manifest aws-nlb.yaml
```

## 5. GCP Internal Load Balancer

An internal LoadBalancer service for GCP, accessible only within the VPC.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: internal-api
spec:
  namespace: production
  name: internal-api
  type: load_balancer
  selector:
    app: internal-api
  ports:
    - name: http
      port: 80
      target_port: "8080"
  annotations:
    cloud.google.com/load-balancer-type: "Internal"
    networking.gke.io/internal-load-balancer-allow-global-access: "true"
```

```bash
openmcf pulumi up --manifest gcp-ilb.yaml
```

## 6. ExternalName Service

Maps a service to an external DNS name. Useful for referencing external databases or third-party APIs.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: external-database
spec:
  namespace: production
  name: external-database
  type: external_name
  external_dns_name: my-database.us-east-1.rds.amazonaws.com
```

```bash
openmcf pulumi up --manifest external-name.yaml
```

Pods can now connect to `external-database.production.svc.cluster.local` and traffic will be routed to the RDS endpoint.

## 7. Headless Service for StatefulSet

A headless service (clusterIP: None) that enables direct pod DNS resolution, commonly used with StatefulSets.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: cassandra-headless
spec:
  namespace: data
  name: cassandra
  headless: true
  selector:
    app: cassandra
  ports:
    - name: cql
      port: 9042
      target_port: "9042"
    - name: internode
      port: 7000
      target_port: "7000"
```

```bash
openmcf pulumi up --manifest headless.yaml
```

Individual pods are addressable as `cassandra-0.cassandra.data.svc.cluster.local`.

## 8. DNS Service with UDP and TCP

A service exposing both UDP and TCP ports, typical for DNS services.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: custom-dns
spec:
  namespace: kube-system
  name: custom-dns
  selector:
    app: coredns
  ports:
    - name: dns-udp
      port: 53
      target_port: "53"
      protocol: UDP
    - name: dns-tcp
      port: 53
      target_port: "53"
      protocol: TCP
```

```bash
openmcf pulumi up --manifest dns-service.yaml
```

## Deployment Commands

All examples can be deployed with either Pulumi or Terraform:

```bash
# Deploy with Pulumi
openmcf pulumi up --manifest <manifest-file>

# Preview changes before deploying
openmcf pulumi preview --manifest <manifest-file>

# Deploy with Terraform
openmcf tofu apply --manifest <manifest-file>

# Plan changes before deploying
openmcf tofu plan --manifest <manifest-file>

# Tear down
openmcf pulumi down --manifest <manifest-file>
openmcf tofu destroy --manifest <manifest-file>
```
