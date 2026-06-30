# OCI Application Load Balancer

Deploys an OCI Application Load Balancer (Layer 7) with backend sets, listeners, SSL certificates, virtual hostnames, and rule sets as a single atomic deployment unit. Supports HTTP, HTTP/2, TCP, and gRPC protocols with flexible bandwidth, cookie-based session persistence, health checking, and rule-based request/response manipulation including HTTP redirects, header injection, and access control.

## What Gets Created

When you deploy an OciApplicationLoadBalancer resource, Planton provisions:

- **Application Load Balancer** — an `oci_load_balancer_load_balancer` resource in the specified compartment with the configured shape, bandwidth, and subnet placement. Standard Planton freeform tags are applied for resource tracking.
- **Backend Sets** — one or more `oci_load_balancer_backend_set` resources defining load balancing policy, health checking, and optional SSL and session persistence configuration for each group of backends.
- **Backends** — `oci_load_balancer_backend` resources for each backend server defined within a backend set. Created only when backends are specified in the backend set configuration.
- **Listeners** — one or more `oci_load_balancer_listener` resources defining the ports, protocols, and routing to backend sets. Listeners are created after all other sub-resources to ensure correct dependency ordering.
- **Certificates** — `oci_load_balancer_certificate` resources for TLS/SSL termination. Created only when certificates are specified.
- **Hostnames** — `oci_load_balancer_hostname` resources for virtual host-based routing. Created only when hostnames are specified.
- **Rule Sets** — `oci_load_balancer_rule_set` resources for request/response manipulation including HTTP redirects, header injection, access control, and connection limits. Created only when rule sets are specified.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the load balancer will be created — literal value or reference to an OciCompartment resource
- **At least one subnet OCID** for load balancer placement — literal value or reference to an OciSubnet resource. For high availability, provide subnets in two different availability domains
- **Backend server IP addresses and ports** for at least one backend set (backends can be added later, but a load balancer without backends does not route traffic)
- **PEM-encoded TLS certificates** if configuring HTTPS listeners or backend SSL

## Quick Start

Create a file `load-balancer.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciApplicationLoadBalancer
metadata:
  name: web-lb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciApplicationLoadBalancer.web-lb
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

Deploy:

```shell
planton apply -f load-balancer.yaml
```

This creates a public flexible-shape load balancer with 10–100 Mbps bandwidth, one backend set using round-robin distribution with HTTP health checks on `/health`, two backend servers on port 8080, and one HTTP listener on port 80. The load balancer OCID and assigned IP addresses are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the load balancer will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `shape` | `string` | Load balancer shape. Use `"flexible"` for configurable bandwidth (recommended). Deprecated fixed shapes (`"100Mbps"`, `"400Mbps"`, `"8000Mbps"`) are accepted for backward compatibility. | Minimum 1 character |
| `subnetIds` | `StringValueOrRef[]` | OCIDs of subnets where the load balancer will be provisioned. For regional load balancers (recommended), provide subnets in two different availability domains for high availability. Changing subnets after creation forces recreation. Can reference OciSubnet resources via `valueFrom`. | Minimum 1 item |
| `backendSets` | `BackendSet[]` | Backend sets defining groups of backend servers with load balancing policies and health checking. Each listener routes traffic to exactly one default backend set. See [backendSet fields](#backendset-fields). | Minimum 1 item |
| `listeners` | `Listener[]` | Listeners defining ports and protocols on which the load balancer accepts client connections. See [listener fields](#listener-fields). | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name for the load balancer shown in the OCI Console. |
| `shapeDetails` | `ShapeDetails` | — | Bandwidth configuration for flexible-shape load balancers. Required when shape is `"flexible"`. Ignored for fixed shapes. See [shapeDetails fields](#shapedetails-fields). |
| `isPrivate` | `bool` | `false` | When true, creates a private load balancer not accessible from the public internet. Private load balancers receive only private IP addresses. Changing after creation forces recreation. |
| `networkSecurityGroupIds` | `StringValueOrRef[]` | — | OCIDs of network security groups applied to the load balancer. Can reference OciSecurityGroup resources via `valueFrom`. |
| `isDeleteProtectionEnabled` | `bool` | `false` | When true, prevents accidental deletion. Must be explicitly disabled before the load balancer can be deleted. |
| `ipMode` | `string` | `"IPV4"` | IP version mode. Accepted values: `"IPV4"`, `"IPV6"`. |
| `reservedIps` | `ReservedIp[]` | — | Pre-created reserved public IPs to assign. When omitted, OCI assigns ephemeral public IPs. See [reservedIp fields](#reservedip-fields). |
| `isRequestIdEnabled` | `bool` | `false` | When true, the load balancer adds a request ID header to each request for tracing and debugging. |
| `requestIdHeader` | `string` | OCI default | Custom header name for the request ID. Only effective when `isRequestIdEnabled` is true. |
| `certificates` | `Certificate[]` | — | TLS/SSL certificates for HTTPS termination. Referenced by name in listener and backend set SSL configurations. See [certificate fields](#certificate-fields). |
| `hostnames` | `Hostname[]` | — | Virtual hostnames for host-based routing. Referenced by name in listener configurations. See [hostname fields](#hostname-fields). |
| `ruleSets` | `RuleSet[]` | — | Rule sets for request/response manipulation. Referenced by name in listener configurations. See [ruleSet fields](#ruleset-fields). |

### shapeDetails Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `minimumBandwidthInMbps` | `int32` | Minimum bandwidth in Mbps. The load balancer always provides at least this bandwidth. | 10–8000 |
| `maximumBandwidthInMbps` | `int32` | Maximum bandwidth in Mbps. The load balancer can burst up to this bandwidth. Must be >= `minimumBandwidthInMbps`. | 10–8000 |

### reservedIp Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `id` | `string` | OCID of the reserved public IP to assign to the load balancer. | Minimum 1 character |

### backendSet Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name for this backend set within the load balancer. Listeners reference backend sets by this name. | Minimum 1 character |
| `policy` | `enum` | Load balancing policy. Values: `round_robin`, `least_connections`, `ip_hash`. | Required (cannot be unspecified) |
| `healthChecker` | `HealthChecker` | Health checker monitoring backend availability. The load balancer removes unhealthy backends from rotation until they pass again. See [healthChecker fields](#healthchecker-fields). | Required |
| `backends` | `Backend[]` | Backend servers in this set. When omitted, the backend set is created without backends (useful when backends are added dynamically). See [backend fields](#backend-fields). | Optional |
| `sslConfiguration` | `SslConfiguration` | SSL configuration for encrypting traffic between the load balancer and backend servers (backend SSL / re-encryption). See [sslConfiguration fields](#sslconfiguration-fields). | Optional |
| `backendMaxConnections` | `int32` | Maximum simultaneous connections per backend. When omitted, connections are unlimited. | Optional |
| `lbCookieSessionPersistence` | `LbCookieSessionPersistenceConfig` | Load-balancer-managed cookie persistence. The LB injects and tracks a cookie to pin clients to backends. Mutually exclusive with `appCookieSessionPersistence`. See [lbCookieSessionPersistenceConfig fields](#lbcookiesessionpersistenceconfig-fields). | Optional |
| `appCookieSessionPersistence` | `SessionPersistenceConfig` | Application-managed cookie persistence. The LB reads an existing application cookie for backend affinity. Mutually exclusive with `lbCookieSessionPersistence`. See [sessionPersistenceConfig fields](#sessionpersistenceconfig-fields). | Optional |

### healthChecker Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `protocol` | `enum` | Health check protocol. Values: `http`, `tcp`. | Required (cannot be unspecified) |
| `port` | `int32` | Port to probe. When omitted or 0, uses the backend's traffic port. | Optional |
| `urlPath` | `string` | URL path for HTTP health checks (e.g., `"/health"`). Required when protocol is `http`. Ignored for `tcp`. | Optional |
| `returnCode` | `int32` | Expected HTTP status code (e.g., `200`). When omitted, any 2xx status is healthy. | Optional |
| `responseBodyRegex` | `string` | Regex pattern to match against the response body. When omitted, body content is not checked. | Optional |
| `intervalMs` | `int32` | Interval between health checks in milliseconds. | Default: 30000 |
| `timeoutInMillis` | `int32` | Maximum wait time for a response in milliseconds. | Default: 3000 |
| `retries` | `int32` | Consecutive failures before marking a backend unhealthy. | Default: 3 |
| `isForcePlainText` | `bool` | When true, forces plain text health checks even when the backend set has SSL configured. | Optional |

### backend Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ipAddress` | `string` | IP address of the backend server. | Minimum 1 character |
| `port` | `int32` | Port on which the backend listens for traffic. | 1–65535 |
| `weight` | `int32` | Relative weight for traffic distribution. Higher weights receive proportionally more traffic. | Default: 1 |
| `backup` | `bool` | When true, this backend only receives traffic when all non-backup backends are unhealthy. | Optional |
| `drain` | `bool` | When true, existing connections complete but no new connections are sent to this backend. | Optional |
| `offline` | `bool` | When true, the backend is taken offline. No traffic is sent to offline backends. | Optional |
| `maxConnections` | `int32` | Maximum simultaneous connections to this backend. When omitted, the backend set's `backendMaxConnections` applies. | Optional |

### sslConfiguration Fields

Used in both backend sets (backend-facing SSL / re-encryption) and listeners (client-facing SSL termination).

| Field | Type | Description |
|-------|------|-------------|
| `certificateIds` | `string[]` | OCIDs of OCI Certificate Service certificates. Preferred over `certificateName` for managed certificate lifecycle. |
| `certificateName` | `string` | Name of a certificate defined in this load balancer's `certificates` list. |
| `cipherSuiteName` | `string` | Cipher suite for SSL negotiation. Example: `"oci-default-ssl-cipher-suite-v1"`. |
| `protocols` | `string[]` | TLS protocol versions to accept. Example: `["TLSv1.2", "TLSv1.3"]`. |
| `serverOrderPreference` | `string` | Cipher order preference. Values: `"ENABLED"` (server preference), `"DISABLED"` (client preference). |
| `trustedCertificateAuthorityIds` | `string[]` | OCIDs of trusted CA certificates for verifying backend server certificates (backend SSL) or client certificates (mutual TLS on listener). |
| `verifyDepth` | `int32` | Maximum certificate chain verification depth. Default: 5. |
| `verifyPeerCertificate` | `bool` | When true, verifies the peer's certificate. For backend SSL: verifies the backend server certificate. For listener SSL with mutual TLS: verifies the client certificate. |
| `hasSessionResumption` | `bool` | When true, enables TLS session resumption for improved performance. Only applicable in listener SSL context. Ignored for backend set SSL. |

### lbCookieSessionPersistenceConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `cookieName` | `string` | Cookie name. When omitted, OCI generates a default name. |
| `disableFallback` | `bool` | When true, clients without the cookie are rejected instead of being assigned to a new backend. |
| `domain` | `string` | Domain attribute for the Set-Cookie header. When omitted, applies to the request domain. |
| `isHttpOnly` | `bool` | When true, the cookie is marked HttpOnly (not accessible to JavaScript). |
| `isSecure` | `bool` | When true, the cookie is marked Secure (only sent over HTTPS). |
| `maxAgeInSeconds` | `int32` | Cookie lifetime in seconds. When omitted or 0, the cookie is a session cookie (expires when the browser closes). |
| `path` | `string` | Path attribute for the Set-Cookie header. Default: `"/"`. |

### sessionPersistenceConfig Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `cookieName` | `string` | Name of the application cookie used for session affinity. | Minimum 1 character |
| `disableFallback` | `bool` | When true, clients without the cookie are rejected instead of being assigned to a new backend. | Optional |

### listener Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name for this listener within the load balancer. | Minimum 1 character |
| `port` | `int32` | Port on which the listener accepts connections. | 1–65535 |
| `protocol` | `enum` | Listener protocol. Values: `http`, `http2`, `tcp`, `grpc`. | Required (cannot be unspecified) |
| `defaultBackendSetName` | `string` | Name of the backend set that receives traffic from this listener. Must match a backend set defined in `backendSets`. | Minimum 1 character |
| `sslConfiguration` | `SslConfiguration` | SSL configuration for client-facing encryption (SSL termination). Required for HTTPS listeners. See [sslConfiguration fields](#sslconfiguration-fields). | Optional |
| `connectionConfiguration` | `ConnectionConfiguration` | Connection timeout and proxy protocol settings. See [connectionConfiguration fields](#connectionconfiguration-fields). | Optional |
| `hostnameNames` | `string[]` | Names of hostname resources in this load balancer's `hostnames` list. When set, the listener only handles requests matching these hostnames (virtual host routing). | Optional |
| `ruleSetNames` | `string[]` | Names of rule set resources in this load balancer's `ruleSets` list. Rule sets are applied in the order specified. | Optional |
| `routingPolicyName` | `string` | Name of a routing policy for content-based routing. Routing policies are managed outside this component. | Optional |

### connectionConfiguration Fields

| Field | Type | Description |
|-------|------|-------------|
| `idleTimeoutInSeconds` | `int64` | Maximum idle time in seconds before the load balancer closes the connection. Applies to both client-side and backend-side connections. |
| `backendTcpProxyProtocolVersion` | `int32` | Proxy protocol version (1 or 2). When set, the load balancer prepends proxy protocol headers to backend connections, allowing backends to see the original client IP. |

### certificate Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `certificateName` | `string` | Unique name for this certificate within the load balancer. SSL configurations reference certificates by this name. | Minimum 1 character |
| `caCertificate` | `string` | PEM-encoded CA certificate chain. | Optional |
| `publicCertificate` | `string` | PEM-encoded public certificate. | Optional |
| `privateKey` | `string` | PEM-encoded private key. Sensitive. | Optional |
| `passphrase` | `string` | Passphrase for an encrypted private key. Sensitive. | Optional |

### hostname Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name for this hostname resource within the load balancer. Listeners reference hostnames by this name. | Minimum 1 character |
| `hostname` | `string` | Fully qualified domain name (FQDN) to match against the HTTP Host header. Example: `"app.example.com"`. | Minimum 1 character |

### ruleSet Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name for this rule set within the load balancer. Listeners reference rule sets by this name. | Minimum 1 character |
| `items` | `RuleSetItem[]` | Rules in this set. See [ruleSetItem fields](#rulesetitem-fields). | Minimum 1 item |

### ruleSetItem Fields

The `action` field determines which other fields are relevant for each rule:

| Action | Relevant Fields |
|--------|-----------------|
| `add_http_request_header` | `header`, `value` |
| `add_http_response_header` | `header`, `value` |
| `extend_http_request_header_value` | `header`, `prefix`, `suffix` |
| `extend_http_response_header_value` | `header`, `prefix`, `suffix` |
| `remove_http_request_header` | `header` |
| `remove_http_response_header` | `header` |
| `redirect` | `redirectUri`, `responseCode`, `conditions` |
| `allow` | `conditions` |
| `control_access_using_http_methods` | `allowedMethods`, `statusCode` |
| `http_header` | `areInvalidCharactersAllowed`, `httpLargeHeaderSizeInKb` |
| `ip_based_max_connections` | `defaultMaxConnections`, `ipMaxConnections` |

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `action` | `enum` | Rule action type. See the action-to-field mapping above. | Required (cannot be unspecified) |
| `header` | `string` | HTTP header name. Used by header add/remove/extend actions. | Optional |
| `value` | `string` | Header value. Used by `add_http_request_header` and `add_http_response_header`. | Optional |
| `prefix` | `string` | Prefix to prepend to an existing header value. Used by extend actions. | Optional |
| `suffix` | `string` | Suffix to append to an existing header value. Used by extend actions. | Optional |
| `redirectUri` | `RedirectUri` | Redirect URI template. Used by the `redirect` action. See [redirectUri fields](#redirecturi-fields). | Optional |
| `responseCode` | `int32` | HTTP response code for the redirect (e.g., 301, 302, 307, 308). Used by the `redirect` action. | Optional |
| `conditions` | `RuleSetItemCondition[]` | Conditions that must be met for this rule to apply. Used by `redirect` and `allow` actions. See [ruleSetItemCondition fields](#rulesetitemcondition-fields). | Optional |
| `allowedMethods` | `string[]` | Allowed HTTP methods. Requests using other methods receive the `statusCode` response. Used by `control_access_using_http_methods`. | Optional |
| `statusCode` | `int32` | HTTP status code returned when access is denied (e.g., 403, 405). Used by `control_access_using_http_methods`. | Optional |
| `areInvalidCharactersAllowed` | `bool` | When true, allows invalid characters in HTTP headers. Used by `http_header`. | Optional |
| `httpLargeHeaderSizeInKb` | `int32` | Maximum HTTP header size in KB. Used by `http_header`. | Optional |
| `defaultMaxConnections` | `int32` | Default maximum connections per IP when no specific IP rule matches. Used by `ip_based_max_connections`. | Optional |
| `ipMaxConnections` | `IpMaxConnection[]` | Per-IP connection limits. Used by `ip_based_max_connections`. See [ipMaxConnection fields](#ipmaxconnection-fields). | Optional |
| `description` | `string` | Description of the rule. | Optional |

### ruleSetItemCondition Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `attributeName` | `string` | Attribute to evaluate. Values: `"PATH"`, `"SOURCE_IP_ADDRESS"`, `"SOURCE_VCN_ID"`, `"SOURCE_VCN_IP_ADDRESS"`. | Minimum 1 character |
| `attributeValue` | `string` | Value to match against the attribute. | Minimum 1 character |
| `operator` | `string` | Matching operator. Values: `"EXACT_MATCH"`, `"FORCE_LONGEST_PREFIX_MATCH"`, `"PREFIX_MATCH"`, `"SUFFIX_MATCH"`. | Default: `"EXACT_MATCH"` |

### redirectUri Fields

| Field | Type | Description |
|-------|------|-------------|
| `protocol` | `string` | Target protocol. Use `"{protocol}"` to preserve the original request's protocol. |
| `host` | `string` | Target hostname. Use `"{host}"` to preserve the original request's hostname. |
| `port` | `int32` | Target port. Use `0` to preserve the original request's port. |
| `path` | `string` | Target path. Use `"{path}"` to preserve the original request's path. |
| `query` | `string` | Target query string. Use `"{query}"` to preserve the original request's query string. |

### ipMaxConnection Fields

| Field | Type | Description |
|-------|------|-------------|
| `ipAddresses` | `string[]` | IP addresses to apply this connection limit to. |
| `maxConnections` | `int32` | Maximum simultaneous connections allowed from these IPs. |

## Examples

### Minimal HTTP Load Balancer

A public HTTP load balancer distributing traffic across two backends with round-robin policy and HTTP health checking:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciApplicationLoadBalancer
metadata:
  name: web-lb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciApplicationLoadBalancer.web-lb
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

### HTTPS with SSL Termination and HTTP Redirect

An HTTPS load balancer with a certificate for SSL termination. The HTTP listener on port 80 redirects all traffic to HTTPS using a rule set:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciApplicationLoadBalancer
metadata:
  name: secure-web
  org: acme
  env: staging
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: platform
    pulumi.planton.dev/stack.name: staging.OciApplicationLoadBalancer.secure-web
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

### Multi-Domain with Virtual Hostname Routing

A load balancer serving two applications on the same IP address using virtual hostname routing. Each hostname routes to a different backend set:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciApplicationLoadBalancer
metadata:
  name: multi-domain-lb
  org: acme
  env: prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: platform
    pulumi.planton.dev/stack.name: prod.OciApplicationLoadBalancer.multi-domain-lb
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

### Full Production Load Balancer

A production load balancer with HTTPS, HTTP-to-HTTPS redirect, security headers, LB-managed session persistence, delete protection, request tracing, NSGs, and infrastructure references via `valueFrom`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciApplicationLoadBalancer
metadata:
  name: prod-lb
  org: acme
  env: prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: platform
    pulumi.planton.dev/stack.name: prod.OciApplicationLoadBalancer.prod-lb
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
  isPrivate: false
  isDeleteProtectionEnabled: true
  isRequestIdEnabled: true
  requestIdHeader: "X-Request-ID"
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciSecurityGroup
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

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `load_balancer_id` | `string` | OCID of the load balancer. |
| `ip_addresses` | `string` | Comma-separated IP addresses assigned to the load balancer. For public load balancers, includes the public IP(s). For private load balancers, includes the private IP(s). |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/ocisubnet) — provides subnets referenced by `subnetIds` via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — manages network security rules applied to the load balancer (`networkSecurityGroupIds`) via `valueFrom`
- [OciNetworkLoadBalancer](/docs/catalog/oci/ocinetworkloadbalancer) — Layer 4 (TCP/UDP) load balancing with source IP preservation, for workloads that do not need HTTP-level features
