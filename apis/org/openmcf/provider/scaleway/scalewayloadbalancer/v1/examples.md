# ScalewayLoadBalancer Examples

## Minimal HTTP Load Balancer

The simplest configuration: a small LB with one backend and one HTTP frontend.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayLoadBalancer
metadata:
  name: web-lb
spec:
  zone: fr-par-1
  type: LB-S
  backends:
    - name: web
      serverIps:
        - 10.0.1.5
        - 10.0.1.6
      forwardPort: 80
      forwardProtocol: http
  frontends:
    - name: http
      inboundPort: 80
      backendName: web
```

---

## HTTPS with Let's Encrypt

Auto-provisioned TLS certificate with HTTP-to-HTTPS setup. Includes both an HTTP frontend (for redirect or plain access) and an HTTPS frontend with a Let's Encrypt certificate.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayLoadBalancer
metadata:
  name: secure-web-lb
spec:
  zone: fr-par-1
  type: LB-GP-M
  backends:
    - name: web
      serverIps:
        - 10.0.1.5
        - 10.0.1.6
      forwardPort: 8080
      forwardProtocol: http
      healthCheck:
        type: http
        uri: /health
        expectedCode: 200
  frontends:
    - name: http
      inboundPort: 80
      backendName: web
    - name: https
      inboundPort: 443
      backendName: web
      certificateNames:
        - example-cert
  certificates:
    - name: example-cert
      letsencrypt:
        commonName: example.com
        subjectAlternativeNames:
          - www.example.com
```

---

## Multi-Service Load Balancer

Two separate backends for web and API traffic, each with its own health check and frontend port.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayLoadBalancer
metadata:
  name: multi-service-lb
spec:
  zone: fr-par-1
  type: LB-GP-M
  backends:
    - name: web
      serverIps:
        - 10.0.1.5
        - 10.0.1.6
      forwardPort: 80
      forwardProtocol: http
      healthCheck:
        type: http
        uri: /health
    - name: api
      serverIps:
        - 10.0.2.10
        - 10.0.2.11
      forwardPort: 8080
      forwardProtocol: http
      forwardPortAlgorithm: leastconn
      healthCheck:
        type: http
        uri: /api/health
        expectedCode: 200
        checkDelay: 10s
  frontends:
    - name: web-http
      inboundPort: 80
      backendName: web
    - name: api-http
      inboundPort: 8080
      backendName: api
```

---

## TCP Load Balancer (Database Proxy)

Layer 4 TCP load balancing for non-HTTP services like PostgreSQL.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayLoadBalancer
metadata:
  name: db-lb
spec:
  zone: fr-par-1
  type: LB-S
  backends:
    - name: postgres
      serverIps:
        - 10.0.3.20
        - 10.0.3.21
      forwardPort: 5432
      forwardProtocol: tcp
      stickySessions: table
      healthCheck:
        type: tcp
        checkDelay: 10s
        checkMaxRetries: 5
  frontends:
    - name: postgres
      inboundPort: 5432
      backendName: postgres
```

---

## Full-Featured Example with All Options

Demonstrates all available spec fields.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayLoadBalancer
metadata:
  name: production-lb
  org: my-org
  env: production
spec:
  zone: fr-par-1
  type: LB-GP-L
  description: Production web application load balancer
  sslCompatibilityLevel: ssl_compatibility_level_modern
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  backends:
    - name: web
      serverIps:
        - 10.0.1.5
        - 10.0.1.6
        - 10.0.1.7
      forwardPort: 8080
      forwardProtocol: http
      forwardPortAlgorithm: leastconn
      stickySessions: cookie
      stickySessionsCookieName: SERVERID
      timeoutConnect: 5s
      timeoutServer: 30s
      onMarkedDownAction: shutdown_sessions
      proxyProtocol: v2
      healthCheck:
        type: http
        uri: /health
        expectedCode: 200
        checkDelay: 5s
        checkTimeout: 3s
        checkMaxRetries: 3
        port: 8081
    - name: api
      serverIps:
        - 10.0.2.10
        - 10.0.2.11
      forwardPort: 3000
      forwardProtocol: http
      forwardPortAlgorithm: roundrobin
      sslBridging: true
      healthCheck:
        type: https
        uri: /api/health
        expectedCode: 200
  frontends:
    - name: http
      inboundPort: 80
      backendName: web
    - name: https
      inboundPort: 443
      backendName: web
      certificateNames:
        - production-cert
      timeoutClient: 60s
      enableHttp3: true
    - name: api-https
      inboundPort: 8443
      backendName: api
      certificateNames:
        - production-cert
  certificates:
    - name: production-cert
      letsencrypt:
        commonName: example.com
        subjectAlternativeNames:
          - www.example.com
          - api.example.com
```

---

## Infra Chart Composition with valueFrom References

Demonstrates how a ScalewayLoadBalancer composes with other resources in an infra chart using `valueFrom` references.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayLoadBalancer
metadata:
  name: app-lb
spec:
  zone: fr-par-1
  type: LB-GP-M
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  backends:
    - name: web
      serverIps:
        - 10.0.1.5
        - 10.0.1.6
      forwardPort: 80
      forwardProtocol: http
  frontends:
    - name: http
      inboundPort: 80
      backendName: web
    - name: https
      inboundPort: 443
      backendName: web
      certificateNames:
        - app-cert
  certificates:
    - name: app-cert
      letsencrypt:
        commonName: app.example.com
```

A downstream `ScalewayDnsRecord` can then reference the LB's IP:

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
      kind: ScalewayLoadBalancer
      name: app-lb
      fieldPath: status.outputs.lb_ip_address
  ttl: 300
```
