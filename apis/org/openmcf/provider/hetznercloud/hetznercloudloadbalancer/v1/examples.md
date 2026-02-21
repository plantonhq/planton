# HetznerCloudLoadBalancer Examples

## Minimal HTTP Load Balancer

The simplest configuration: an lb11 load balancer in Falkenstein with a single HTTP service on port 80 forwarding to one server target. The server is referenced by literal ID.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: web-lb
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: http
  serverTargets:
    - serverId:
        value: "12345"
```

---

## HTTPS with TLS Termination

An HTTPS load balancer that terminates TLS using a certificate managed by a `HetznerCloudCertificate` resource. HTTP traffic on port 80 is automatically redirected to HTTPS. A custom health check verifies that the backend application responds on `/health`.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: web-https-lb
  org: acme-corp
  env: staging
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: https
      destinationPort: 8080
      http:
        certificateIds:
          - valueFrom:
              kind: HetznerCloudCertificate
              name: web-cert
              fieldPath: status.outputs.certificate_id
        redirectHttp: true
      healthCheck:
        protocol: http
        port: 8080
        interval: 10
        timeout: 5
        retries: 3
        http:
          path: /health
  serverTargets:
    - serverId:
        value: "12345"
    - serverId:
        value: "12346"
```

---

## Multi-Service with Sticky Sessions

A load balancer exposing both HTTPS (port 443) and a TCP service (port 5432 for database connections). The HTTPS service uses cookie-based sticky sessions with a custom cookie name and lifetime. Multiple servers are referenced by literal ID.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: app-lb
  org: acme-corp
  env: production
spec:
  loadBalancerType: lb21
  location: fsn1
  algorithm: least_connections
  services:
    - protocol: https
      destinationPort: 8443
      http:
        certificateIds:
          - value: "99001"
        stickySessions: true
        cookieName: APPSESSION
        cookieLifetime: 3600
        redirectHttp: true
      healthCheck:
        protocol: http
        port: 8443
        http:
          path: /ready
          statusCodes:
            - "200"
    - protocol: tcp
      listenPort: 5432
      destinationPort: 5432
      healthCheck:
        protocol: tcp
        port: 5432
        interval: 10
        timeout: 5
        retries: 2
  serverTargets:
    - serverId:
        value: "11111"
    - serverId:
        value: "22222"
    - serverId:
        value: "33333"
```

---

## Private Network Load Balancer

A load balancer attached to a private Hetzner Cloud network. Targets receive traffic over the private network instead of the public internet. The load balancer gets a fixed private IP and keeps its public interface enabled for client-facing traffic.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: internal-lb
  org: acme-corp
  env: production
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: http
      destinationPort: 8080
      healthCheck:
        protocol: http
        port: 8080
        http:
          path: /health
  serverTargets:
    - serverId:
        value: "11111"
      usePrivateIp: true
    - serverId:
        value: "22222"
      usePrivateIp: true
  network:
    networkId:
      value: "55555"
    ip: "10.0.1.100"
    enablePublicInterface: true
```

---

## Full-Featured Production Load Balancer

A production-grade load balancer using all three target types, HTTPS with TLS termination, private networking, and `valueFrom` references to other OpenMCF components. Demonstrates the complete feature set: label selector targets for auto-scaling, an IP target for an external service, custom health checks, sticky sessions, and delete protection.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: prod-lb
  org: acme-corp
  env: production
spec:
  loadBalancerType: lb31
  location: fsn1
  algorithm: least_connections
  deleteProtection: true
  services:
    - protocol: https
      destinationPort: 8080
      http:
        certificateIds:
          - valueFrom:
              kind: HetznerCloudCertificate
              name: prod-cert
              fieldPath: status.outputs.certificate_id
        redirectHttp: true
        stickySessions: true
        cookieName: PRODSESSION
        cookieLifetime: 1800
      healthCheck:
        protocol: http
        port: 8080
        interval: 10
        timeout: 5
        retries: 3
        http:
          path: /health
          statusCodes:
            - "200"
    - protocol: tcp
      listenPort: 6379
      destinationPort: 6379
      proxyprotocol: true
      healthCheck:
        protocol: tcp
        port: 6379
        interval: 10
        timeout: 5
        retries: 2
  serverTargets:
    - serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: web-01
          fieldPath: status.outputs.server_id
      usePrivateIp: true
    - serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: web-02
          fieldPath: status.outputs.server_id
      usePrivateIp: true
  labelSelectorTargets:
    - selector: "env=production,role=web"
      usePrivateIp: true
  ipTargets:
    - ip: "203.0.113.50"
  network:
    networkId:
      valueFrom:
        kind: HetznerCloudNetwork
        name: main-vpc
        fieldPath: status.outputs.network_id
    ip: "10.0.1.200"
    enablePublicInterface: true
```
