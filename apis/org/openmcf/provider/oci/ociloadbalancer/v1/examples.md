# OCI Application Load Balancer Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure Application Load Balancers using the OpenMCF API. Each example demonstrates a different use case, progressing from a minimal HTTP load balancer to a fully configured production deployment with HTTPS, session persistence, rule sets, virtual hostname routing, and infrastructure references.

## Table of Contents

- [Example 1: Minimal HTTP Load Balancer](#example-1-minimal-http-load-balancer)
- [Example 2: HTTPS with SSL Termination and HTTP Redirect](#example-2-https-with-ssl-termination-and-http-redirect)
- [Example 3: Weighted Backends with Backup and Draining](#example-3-weighted-backends-with-backup-and-draining)
- [Example 4: Multi-Domain Virtual Hostname Routing](#example-4-multi-domain-virtual-hostname-routing)
- [Example 5: Session Persistence with LB-Managed Cookie](#example-5-session-persistence-with-lb-managed-cookie)
- [Example 6: Full Production Load Balancer](#example-6-full-production-load-balancer)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal HTTP Load Balancer

**Use Case:** A public HTTP load balancer distributing traffic across two backend servers with round-robin policy — the simplest path to load-balanced traffic on OCI.

**Configuration:**
- **Shape:** Flexible (10–100 Mbps)
- **Backend sets:** 1 (round robin, HTTP health check)
- **Backends:** 2
- **Listeners:** 1 (HTTP on port 80)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLoadBalancer
metadata:
  name: web-lb
  org: my-org
  env: dev
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciLoadBalancer.web-lb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  shape: "flexible"
  shapeDetails:
    minimumBandwidthInMbps: 10
    maximumBandwidthInMbps: 100
  subnetIds:
    - value: "ocid1.subnet.oc1.iad.example"
  backendSets:
    - name: "web-backend"
      policy: round_robin
      healthChecker:
        protocol: http
        urlPath: "/health"
        port: 8080
      backends:
        - ipAddress: "10.0.1.10"
          port: 8080
        - ipAddress: "10.0.1.11"
          port: 8080
  listeners:
    - name: "http-listener"
      port: 80
      protocol: http
      defaultBackendSetName: "web-backend"
```

**Deploy with OpenMCF CLI:**

```shell
openmcf apply -f web-lb.yaml
```

**What happens:**
- A flexible-shape load balancer is created with 10–100 Mbps bandwidth in the specified subnet.
- The backend set uses round-robin distribution, cycling through each backend for consecutive requests.
- HTTP health checks probe `/health` on port 8080 every 30 seconds (default). Backends that fail 3 consecutive checks (default) are removed from rotation.
- The HTTP listener on port 80 forwards all traffic to `web-backend`.
- OCI assigns an ephemeral public IP to the load balancer. The load balancer OCID and IP addresses are exported as stack outputs.

---

## Example 2: HTTPS with SSL Termination and HTTP Redirect

**Use Case:** A production web application requiring HTTPS with SSL termination at the load balancer. HTTP traffic on port 80 is automatically redirected to HTTPS on port 443 using a rule set.

**Configuration:**
- **Shape:** Flexible (10–100 Mbps)
- **Backend sets:** 1 (round robin, HTTP health check)
- **Listeners:** 2 (HTTP redirect on 80, HTTPS on 443)
- **Certificates:** 1 (PEM-encoded)
- **Rule sets:** 1 (HTTP-to-HTTPS redirect)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLoadBalancer
metadata:
  name: secure-web
  org: acme
  env: staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: staging.OciLoadBalancer.secure-web
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  shape: "flexible"
  shapeDetails:
    minimumBandwidthInMbps: 10
    maximumBandwidthInMbps: 100
  subnetIds:
    - value: "ocid1.subnet.oc1.iad.example"
  certificates:
    - certificateName: "app-cert"
      publicCertificate: |
        -----BEGIN CERTIFICATE-----
        MIIBkTCB+wIJALxyz...
        -----END CERTIFICATE-----
      privateKey: |
        -----BEGIN RSA PRIVATE KEY-----
        MIIEpAIBAAKCAQEA...
        -----END RSA PRIVATE KEY-----
  backendSets:
    - name: "web-backend"
      policy: round_robin
      healthChecker:
        protocol: http
        urlPath: "/health"
        port: 8080
      backends:
        - ipAddress: "10.0.1.10"
          port: 8080
        - ipAddress: "10.0.1.11"
          port: 8080
  ruleSets:
    - name: "http-to-https"
      items:
        - action: redirect
          redirectUri:
            protocol: "https"
            host: "{host}"
            port: 443
            path: "{path}"
            query: "{query}"
          responseCode: 301
  listeners:
    - name: "http-redirect"
      port: 80
      protocol: http
      defaultBackendSetName: "web-backend"
      ruleSetNames:
        - "http-to-https"
    - name: "https-listener"
      port: 443
      protocol: http
      defaultBackendSetName: "web-backend"
      sslConfiguration:
        certificateName: "app-cert"
        protocols:
          - "TLSv1.2"
          - "TLSv1.3"
```

**What happens:**
- A certificate named `app-cert` is uploaded to the load balancer. The certificate PEM data is stored in the load balancer resource (not in OCI Certificate Service).
- The `http-to-https` rule set returns a 301 redirect to `https://{host}:443/{path}?{query}` for every request. Template variables `{host}`, `{path}`, and `{query}` preserve the original request values.
- The HTTP listener on port 80 applies the redirect rule set — all HTTP requests receive a 301 redirect before reaching any backend.
- The HTTPS listener on port 443 terminates SSL using `app-cert` and forwards decrypted traffic to `web-backend`. The listener protocol is `http` (not a separate `https` value) with `sslConfiguration` enabling SSL termination.
- Only TLSv1.2 and TLSv1.3 are accepted. Clients attempting TLSv1.0 or TLSv1.1 connections are rejected.
- The `defaultBackendSetName` on the HTTP redirect listener is required even though the redirect intercepts all traffic before it reaches a backend.

**Certificate management note:**
For production deployments, prefer OCI Certificate Service managed certificates referenced via `sslConfiguration.certificateIds` instead of uploading PEM certificates directly. Managed certificates support automatic renewal and centralized lifecycle management.

---

## Example 3: Weighted Backends with Backup and Draining

**Use Case:** A load balancer with weighted traffic distribution for canary deployments, a backup backend for failover, and a draining backend for graceful maintenance.

**Configuration:**
- **Shape:** Flexible (10–400 Mbps)
- **Backend sets:** 1 (round robin with weights)
- **Backends:** 4 (2 primary, 1 canary, 1 backup)
- **Listeners:** 1 (HTTP on port 80)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLoadBalancer
metadata:
  name: canary-lb
  org: acme
  env: staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: staging.OciLoadBalancer.canary-lb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  shape: "flexible"
  shapeDetails:
    minimumBandwidthInMbps: 10
    maximumBandwidthInMbps: 400
  subnetIds:
    - value: "ocid1.subnet.oc1.iad.example"
  backendSets:
    - name: "app-backend"
      policy: round_robin
      healthChecker:
        protocol: http
        urlPath: "/healthz"
        port: 8080
        returnCode: 200
        intervalMs: 10000
        timeoutInMillis: 3000
        retries: 3
      backends:
        - ipAddress: "10.0.1.10"
          port: 8080
          weight: 5
        - ipAddress: "10.0.1.11"
          port: 8080
          weight: 5
        - ipAddress: "10.0.1.20"
          port: 8080
          weight: 1
        - ipAddress: "10.0.1.30"
          port: 8080
          weight: 1
          backup: true
  listeners:
    - name: "http-listener"
      port: 80
      protocol: http
      defaultBackendSetName: "app-backend"
```

**What happens:**
- Traffic is distributed by weight: the two primary backends (weight 5 each) receive approximately 45% each, and the canary backend (weight 1) receives approximately 10%. This ratio provides a controlled rollout with minimal blast radius.
- The backup backend (weight 1, `backup: true`) is in standby — it receives traffic only when all non-backup backends are unhealthy. This provides a last-resort failover path.
- Health checks probe `/healthz` on port 8080 every 10 seconds. Backends must return HTTP 200. After 3 consecutive failures (30 seconds), a backend is removed from rotation. After 3 consecutive successes, it is re-added.
- To drain a backend for maintenance, update the manifest to set `drain: true` on that backend and re-apply. Existing connections complete, but no new connections are routed to the draining backend.

**Draining a backend for maintenance:**

```yaml
      backends:
        - ipAddress: "10.0.1.10"
          port: 8080
          weight: 5
          drain: true
```

After re-applying, existing connections to `10.0.1.10` complete normally, but all new traffic goes to the remaining backends. When maintenance is complete, remove `drain: true` and re-apply.

---

## Example 4: Multi-Domain Virtual Hostname Routing

**Use Case:** A single load balancer serving two applications — `api.example.com` and `app.example.com` — on the same IP address. Each domain routes to a different backend set using virtual hostname routing.

**Configuration:**
- **Shape:** Flexible (100–400 Mbps)
- **Backend sets:** 2 (one per domain)
- **Hostnames:** 2
- **Listeners:** 2 (one per domain, same port)
- **Subnets:** 2 (cross-AD for high availability)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLoadBalancer
metadata:
  name: multi-domain-lb
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.OciLoadBalancer.multi-domain-lb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  shape: "flexible"
  shapeDetails:
    minimumBandwidthInMbps: 100
    maximumBandwidthInMbps: 400
  subnetIds:
    - value: "ocid1.subnet.oc1.iad.example-ad1"
    - value: "ocid1.subnet.oc1.iad.example-ad2"
  hostnames:
    - name: "api-host"
      hostname: "api.example.com"
    - name: "app-host"
      hostname: "app.example.com"
  backendSets:
    - name: "api-backend"
      policy: least_connections
      healthChecker:
        protocol: http
        urlPath: "/healthz"
        port: 3000
      backends:
        - ipAddress: "10.0.1.20"
          port: 3000
        - ipAddress: "10.0.1.21"
          port: 3000
    - name: "app-backend"
      policy: round_robin
      healthChecker:
        protocol: http
        urlPath: "/ready"
        port: 8080
      backends:
        - ipAddress: "10.0.2.10"
          port: 8080
        - ipAddress: "10.0.2.11"
          port: 8080
  listeners:
    - name: "api-listener"
      port: 80
      protocol: http
      defaultBackendSetName: "api-backend"
      hostnameNames:
        - "api-host"
    - name: "app-listener"
      port: 80
      protocol: http
      defaultBackendSetName: "app-backend"
      hostnameNames:
        - "app-host"
```

**What happens:**
- Both listeners use port 80, but each is filtered by a different hostname. When a request arrives with `Host: api.example.com`, the `api-listener` handles it and routes to `api-backend`. Requests with `Host: app.example.com` go to `app-listener` and route to `app-backend`.
- The API backend uses `least_connections` policy — ideal for long-lived API requests where connection duration varies. The web app backend uses `round_robin` — suitable for short, uniform HTTP requests.
- Two subnets in different availability domains provide cross-AD redundancy. If one AD becomes unavailable, the load balancer continues operating from the other subnet.
- Requests that do not match any configured hostname are not handled by either listener. To catch unmatched requests, add a listener without `hostnameNames` as a default/catch-all.

**Adding HTTPS to multi-domain:**
To add HTTPS, define certificates for each domain (or use a wildcard certificate), add HTTPS listeners with `sslConfiguration`, and optionally add HTTP redirect rule sets per listener.

---

## Example 5: Session Persistence with LB-Managed Cookie

**Use Case:** A stateful web application requiring client sessions to be pinned to the same backend server. The load balancer manages the session cookie, injecting a `Set-Cookie` header with configurable security attributes.

**Configuration:**
- **Shape:** Flexible (10–100 Mbps)
- **Backend sets:** 1 (round robin with LB cookie persistence)
- **Listeners:** 1 (HTTP on port 80)
- **Session persistence:** LB-managed cookie

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLoadBalancer
metadata:
  name: stateful-app
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.OciLoadBalancer.stateful-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  shape: "flexible"
  shapeDetails:
    minimumBandwidthInMbps: 10
    maximumBandwidthInMbps: 100
  subnetIds:
    - value: "ocid1.subnet.oc1.iad.example"
  backendSets:
    - name: "app-backend"
      policy: round_robin
      healthChecker:
        protocol: http
        urlPath: "/health"
        port: 8080
      backends:
        - ipAddress: "10.0.1.10"
          port: 8080
        - ipAddress: "10.0.1.11"
          port: 8080
        - ipAddress: "10.0.1.12"
          port: 8080
      lbCookieSessionPersistence:
        cookieName: "LB_SESSION"
        isHttpOnly: true
        isSecure: true
        maxAgeInSeconds: 3600
        path: "/"
  listeners:
    - name: "http-listener"
      port: 80
      protocol: http
      defaultBackendSetName: "app-backend"
```

**What happens:**
- On the first request from a client, the load balancer selects a backend using round-robin and injects a `Set-Cookie: LB_SESSION=<value>; HttpOnly; Secure; Path=/; Max-Age=3600` header in the response.
- On subsequent requests, the load balancer reads the `LB_SESSION` cookie and routes the request to the same backend that served the original request.
- `isHttpOnly: true` prevents JavaScript access to the cookie, mitigating XSS-based session theft.
- `isSecure: true` ensures the cookie is only sent over HTTPS connections. When using HTTP (not HTTPS), omit this flag or clients will not send the cookie back.
- `maxAgeInSeconds: 3600` sets a 1-hour cookie lifetime. After expiration, the next request is load-balanced normally and a new cookie is issued. Set to 0 or omit for session cookies (expire when the browser closes).
- If the pinned backend becomes unhealthy, the load balancer selects a new backend and issues a new cookie. Set `disableFallback: true` to reject requests when the pinned backend is unavailable instead of failing over.

**Application-managed cookie alternative:**
If the backend application already sets a session cookie (e.g., `JSESSIONID`), use `appCookieSessionPersistence` instead:

```yaml
      appCookieSessionPersistence:
        cookieName: "JSESSIONID"
```

The load balancer reads the existing `JSESSIONID` cookie to determine backend affinity instead of injecting its own cookie. The two persistence modes are mutually exclusive per backend set.

---

## Example 6: Full Production Load Balancer

**Use Case:** A production load balancer combining HTTPS with SSL termination, HTTP-to-HTTPS redirect, security response headers, LB-managed session persistence, delete protection, request ID tracing, NSGs, cross-AD subnets, and infrastructure references via `valueFrom`.

**Configuration:**
- **Shape:** Flexible (100–400 Mbps)
- **Backend sets:** 1 (round robin with LB cookie, custom health check tuning)
- **Backends:** 3 (2 primary, 1 backup)
- **Listeners:** 2 (HTTP redirect, HTTPS with hostname filtering)
- **Certificates:** 1
- **Hostnames:** 1
- **Rule sets:** 2 (HTTP redirect, security headers)
- **Session persistence:** LB-managed cookie
- **Infrastructure references:** Compartment, subnets, NSG via `valueFrom`

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciLoadBalancer
metadata:
  name: prod-lb
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.OciLoadBalancer.prod-lb
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  shape: "flexible"
  shapeDetails:
    minimumBandwidthInMbps: 100
    maximumBandwidthInMbps: 400
  subnetIds:
    - valueFrom:
        kind: OciSubnet
        name: public-subnet-ad1
        fieldPath: status.outputs.subnetId
    - valueFrom:
        kind: OciSubnet
        name: public-subnet-ad2
        fieldPath: status.outputs.subnetId
  isDeleteProtectionEnabled: true
  isRequestIdEnabled: true
  requestIdHeader: "X-Request-ID"
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciNetworkSecurityGroup
        name: lb-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  certificates:
    - certificateName: "prod-cert"
      publicCertificate: |
        -----BEGIN CERTIFICATE-----
        MIIBkTCB+wIJALxyz...
        -----END CERTIFICATE-----
      caCertificate: |
        -----BEGIN CERTIFICATE-----
        MIIBkTCB+wIJALabc...
        -----END CERTIFICATE-----
      privateKey: |
        -----BEGIN RSA PRIVATE KEY-----
        MIIEpAIBAAKCAQEA...
        -----END RSA PRIVATE KEY-----
  backendSets:
    - name: "web-backend"
      policy: round_robin
      healthChecker:
        protocol: http
        urlPath: "/health"
        port: 8080
        returnCode: 200
        intervalMs: 10000
        timeoutInMillis: 3000
        retries: 3
      backends:
        - ipAddress: "10.0.1.10"
          port: 8080
          weight: 3
        - ipAddress: "10.0.1.11"
          port: 8080
          weight: 3
        - ipAddress: "10.0.1.12"
          port: 8080
          weight: 1
          backup: true
      backendMaxConnections: 1000
      lbCookieSessionPersistence:
        cookieName: "LB_SESSION"
        isHttpOnly: true
        isSecure: true
        maxAgeInSeconds: 3600
        path: "/"
  hostnames:
    - name: "www-host"
      hostname: "www.example.com"
  ruleSets:
    - name: "http-to-https"
      items:
        - action: redirect
          redirectUri:
            protocol: "https"
            host: "{host}"
            port: 443
            path: "{path}"
            query: "{query}"
          responseCode: 301
    - name: "security-headers"
      items:
        - action: add_http_response_header
          header: "Strict-Transport-Security"
          value: "max-age=31536000; includeSubDomains"
        - action: add_http_response_header
          header: "X-Content-Type-Options"
          value: "nosniff"
        - action: add_http_response_header
          header: "X-Frame-Options"
          value: "DENY"
  listeners:
    - name: "http-redirect"
      port: 80
      protocol: http
      defaultBackendSetName: "web-backend"
      ruleSetNames:
        - "http-to-https"
    - name: "https-listener"
      port: 443
      protocol: http
      defaultBackendSetName: "web-backend"
      sslConfiguration:
        certificateName: "prod-cert"
        protocols:
          - "TLSv1.2"
          - "TLSv1.3"
        cipherSuiteName: "oci-default-ssl-cipher-suite-v1"
        serverOrderPreference: "ENABLED"
      hostnameNames:
        - "www-host"
      ruleSetNames:
        - "security-headers"
      connectionConfiguration:
        idleTimeoutInSeconds: 300
```

**What happens:**
- Infrastructure references use `valueFrom` to compose with OciCompartment, OciSubnet, and OciNetworkSecurityGroup. This enables the load balancer to be deployed as part of an infra chart where compartments, subnets, and NSGs are created by upstream components.
- Two subnets in different availability domains provide cross-AD redundancy for the regional load balancer.
- Delete protection prevents accidental deletion — the protection must be explicitly disabled before the load balancer can be removed.
- Request ID tracing injects an `X-Request-ID` header into every request, enabling end-to-end request correlation across the load balancer, backend applications, and logging systems.
- The NSG controls which traffic can reach the load balancer (e.g., allow TCP 80/443 from the internet, deny everything else).
- HTTPS listener terminates SSL with the `prod-cert` certificate, accepting only TLSv1.2 and TLSv1.3. Server order preference is enabled, meaning the load balancer's cipher preference takes priority over the client's.
- The `security-headers` rule set injects `Strict-Transport-Security`, `X-Content-Type-Options`, and `X-Frame-Options` response headers on every HTTPS response.
- Health checks probe every 10 seconds with a 3-second timeout. Backends must return HTTP 200 from `/health` on port 8080. After 3 consecutive failures (30 seconds), a backend is removed. After 3 consecutive successes, it is re-added.
- Backend connection limit is set to 1000 per backend via `backendMaxConnections`, preventing any single backend from being overwhelmed.
- The backup backend receives traffic only when both primary backends are unhealthy, providing a last-resort failover path.
- LB-managed cookie session persistence pins clients to backends for 1 hour with HttpOnly and Secure cookie flags.
- Idle connection timeout is 300 seconds (5 minutes) — connections idle longer than this are closed by the load balancer.

---

## Common Operations

### Get Load Balancer Status

After deploying, retrieve the load balancer OCID and IP addresses from stack outputs:

```shell
# Pulumi
pulumi stack output load_balancer_id
pulumi stack output ip_addresses

# Terraform
terraform output load_balancer_id
terraform output ip_addresses
```

### View Load Balancer Details

Use the OCI CLI to inspect the load balancer:

```shell
LB_ID=$(pulumi stack output load_balancer_id)

oci lb load-balancer get \
  --load-balancer-id "$LB_ID" \
  --query 'data.{state:"lifecycle-state",shape:"shape-name",ips:"ip-addresses[*].ip-address"}' \
  --output table
```

### Check Backend Health

View the health status of backends in a backend set:

```shell
LB_ID=$(pulumi stack output load_balancer_id)

oci lb backend-health get \
  --load-balancer-id "$LB_ID" \
  --backend-set-name "web-backend" \
  --backend-name "10.0.1.10:8080" \
  --query 'data.{status:status,checks:"health-check-results[*].{timestamp:timestamp,result:health-check-status}"}' \
  --output table
```

### Use Outputs in Downstream Resources

The `load_balancer_id` and `ip_addresses` outputs can be referenced by downstream resources. For example, an OciDnsRecord pointing to the load balancer's IP:

```yaml
spec:
  loadBalancerId:
    valueFrom:
      kind: OciLoadBalancer
      name: prod-lb
      fieldPath: status.outputs.loadBalancerId
```

### Update Load Balancer Configuration

To add backends, change health check settings, add listeners, or modify rule sets, update the manifest and re-apply:

```shell
openmcf apply -f prod-lb.yaml
```

Changes to backend sets, listeners, certificates, hostnames, and rule sets are applied in-place. Changes to `subnetIds`, `isPrivate`, or `shape` force load balancer recreation.

### Temporarily Disable a Backend

To take a backend offline without removing it from the manifest, set `offline: true` and re-apply:

```yaml
      backends:
        - ipAddress: "10.0.1.10"
          port: 8080
          offline: true
```

No traffic is sent to offline backends. Remove `offline: true` and re-apply to bring the backend back online.

---

## Best Practices

### Shape and Bandwidth Selection

| Workload Type | Min Bandwidth | Max Bandwidth | Rationale |
|---------------|--------------|---------------|-----------|
| Dev/test environment | 10 Mbps | 10 Mbps | Minimum cost. Sufficient for light testing traffic. |
| Small web application | 10 Mbps | 100 Mbps | Handles traffic spikes up to 100 Mbps without over-provisioning minimum. |
| Production web service | 100 Mbps | 400 Mbps | Baseline for production with headroom for traffic bursts. |
| High-traffic application | 400 Mbps | 8000 Mbps | Large-scale applications with sustained high throughput. |

- **Always use the `"flexible"` shape.** Fixed shapes (`"100Mbps"`, `"400Mbps"`, `"8000Mbps"`) are deprecated by Oracle.
- **Set `minimumBandwidthInMbps` to your baseline.** You pay for the minimum bandwidth even during low-traffic periods.
- **Set `maximumBandwidthInMbps` to your peak.** The load balancer scales up to this limit during traffic spikes and scales back down during quiet periods.

### Health Check Tuning

| Parameter | Dev/Test | Production | Notes |
|-----------|----------|------------|-------|
| `intervalMs` | 30000 | 10000 | More frequent in production for faster detection. |
| `timeoutInMillis` | 10000 | 3000 | Tight timeouts catch hung processes quickly. |
| `retries` | 1 | 3 | Higher in production to avoid false positives from transient network issues. |
| `returnCode` | — (any 2xx) | 200 | Explicit in production ensures the application is truly healthy, not just returning a 204 or 202. |

- **Use HTTP health checks for application-level health.** HTTP checks verify the application can process requests. Use a dedicated health endpoint (e.g., `/health` or `/healthz`) that checks database connectivity, cache availability, and other dependencies.
- **Use TCP health checks for non-HTTP backends.** TCP checks verify the port is accepting connections — suitable for raw TCP backends or when the application does not expose an HTTP health endpoint.
- **Set `isForcePlainText: true` when backends use self-signed certificates.** This forces health checks over plain HTTP even when backend SSL is configured, avoiding health check failures due to certificate validation.

### SSL Configuration

| Setting | Recommended Value | Rationale |
|---------|------------------|-----------|
| `protocols` | `["TLSv1.2", "TLSv1.3"]` | TLSv1.0 and TLSv1.1 are deprecated and vulnerable. |
| `cipherSuiteName` | `"oci-default-ssl-cipher-suite-v1"` | OCI's default suite balances security and compatibility. |
| `serverOrderPreference` | `"ENABLED"` | Server-side cipher preference prevents clients from negotiating weak ciphers. |

- **Prefer OCI Certificate Service over uploaded PEM certificates.** Managed certificates (`certificateIds`) support automatic renewal and centralized lifecycle management.
- **Rotate uploaded certificates before expiry.** When using PEM certificates, update the certificate in the manifest and re-apply before the certificate expires. The load balancer replaces the certificate without downtime.
- **Enable backend SSL for end-to-end encryption.** Add `sslConfiguration` on the backend set when compliance requires encrypted traffic between the load balancer and backends, not just between clients and the load balancer.

### Rule Set Patterns

**HTTP-to-HTTPS redirect** — the most common rule set:

```yaml
ruleSets:
  - name: "http-to-https"
    items:
      - action: redirect
        redirectUri:
          protocol: "https"
          host: "{host}"
          port: 443
          path: "{path}"
          query: "{query}"
        responseCode: 301
```

**Security response headers** — inject headers without modifying backend applications:

```yaml
ruleSets:
  - name: "security-headers"
    items:
      - action: add_http_response_header
        header: "Strict-Transport-Security"
        value: "max-age=31536000; includeSubDomains"
      - action: add_http_response_header
        header: "X-Content-Type-Options"
        value: "nosniff"
      - action: add_http_response_header
        header: "X-Frame-Options"
        value: "DENY"
```

**HTTP method restriction** — allow only GET and POST, reject others with 405:

```yaml
ruleSets:
  - name: "method-restriction"
    items:
      - action: control_access_using_http_methods
        allowedMethods:
          - "GET"
          - "POST"
          - "HEAD"
          - "OPTIONS"
        statusCode: 405
```

### High Availability

- **Use two subnets in different availability domains.** Regional load balancers operate across ADs. If one AD is unavailable, the load balancer continues from the other.
- **Use reserved public IPs for DNS stability.** Ephemeral IPs change when the load balancer is recreated. Reserved IPs persist across recreations, preventing DNS record updates.
- **Enable delete protection on production load balancers.** Set `isDeleteProtectionEnabled: true` to prevent accidental deletion during routine operations.

### Session Persistence

| Mode | When to Use | Trade-off |
|------|------------|-----------|
| LB cookie (`lbCookieSessionPersistence`) | Backend applications are stateless but need sticky routing for performance (e.g., local caches). | The load balancer injects a cookie — no application changes needed. |
| App cookie (`appCookieSessionPersistence`) | Backend applications already manage session cookies (e.g., `JSESSIONID`, `session_id`). | No additional cookie overhead, but the application must set the cookie before the load balancer can route by it. |
| No persistence | Fully stateless backends where any instance can handle any request. | Best for stateless APIs, microservices, and horizontally scaled services. |

- **Set `isSecure: true` and `isHttpOnly: true` for LB cookies in production.** This prevents JavaScript access and ensures the cookie is only sent over HTTPS.
- **Use `disableFallback: false` (default) unless session loss is unacceptable.** With fallback enabled, clients whose pinned backend is down are transparently routed to a healthy backend and issued a new cookie. With fallback disabled, the request is rejected.
