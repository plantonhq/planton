# OCI API Gateway

Deploys an Oracle Cloud Infrastructure API Gateway with a bundled API deployment. The gateway provides the managed network endpoint (public or private) while the deployment defines the API specification: routes with HTTP, OCI Functions, or stock-response backends, optional JWT authentication, CORS, rate limiting, and per-route authorization.

## What Gets Created

When you deploy an OciApiGateway resource, OpenMCF provisions:

- **API Gateway** — an `apigateway.Gateway` resource in the specified compartment and subnet with configurable endpoint type (public or private), optional TLS certificate, and optional NSG bindings.
- **API Deployment** — an `apigateway.Deployment` resource on the gateway with a path prefix, route definitions, and optional request policies (JWT authentication, CORS, rate limiting). The deployment depends on the gateway and is always created.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the gateway and deployment will be created — either a literal value or a reference to an OciCompartment resource
- **A subnet OCID** — for public gateways, this must be a public subnet. Either a literal value or via `valueFrom` referencing an OciSubnet resource
- **An OCI Certificates service certificate OCID** (optional) — if terminating TLS on the gateway
- **OCI Functions function OCIDs** (for Functions backends) — if routing to serverless functions
- **A JWKS endpoint URL or static keys** (for JWT authentication) — if enabling token-based authentication

## Quick Start

Create a file `api-gateway.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciApiGateway
metadata:
  name: my-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciApiGateway.my-api
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  endpointType: endpoint_type_public
  subnetId:
    value: "ocid1.subnet.oc1..example"
  deployment:
    pathPrefix: "/api/v1"
    routes:
      - path: "/health"
        methods:
          - "GET"
        backend:
          type: stock_response
          status: 200
          body: '{"status":"ok"}'
```

Deploy:

```shell
openmcf apply -f api-gateway.yaml
```

This creates a public API gateway with a single health-check route that returns a stock response. The gateway OCID, hostname, and deployment endpoint URL are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the gateway and deployment will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `endpointType` | `enum` | Whether the gateway is internet-facing or VCN-internal. Values: `public`, `private`. Immutable after creation. | Required, not `unspecified` |
| `subnetId` | `StringValueOrRef` | OCID of the subnet for the gateway. Public gateways require a public subnet. Immutable after creation. Can reference an OciSubnet resource via `valueFrom`. | Required |
| `deployment` | `Deployment` | The API deployment defining routes, backends, and policies. | Required |
| `deployment.pathPrefix` | `string` | URL path prefix for all routes (e.g., `"/api/v1"`). Must start with `"/"`. Immutable after creation. | Min length 1, starts with `/` |
| `deployment.routes` | `Route[]` | Route definitions. Routes are evaluated in order; first match wins. | Min 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | metadata name | Display name for the gateway. |
| `certificateId` | `string` | — | OCID of an OCI Certificates service certificate for TLS termination. |
| `networkSecurityGroupIds` | `StringValueOrRef[]` | — | NSGs applied to the gateway. Can reference OciSecurityGroup resources via `valueFrom`. |
| `deployment.displayName` | `string` | `"{gatewayName}-deployment"` | Display name for the deployment. |
| `deployment.loggingPolicies` | `LoggingPolicies` | — | Access and execution logging configuration. |
| `deployment.requestPolicies` | `RequestPolicies` | — | Deployment-level JWT authentication, CORS, and rate limiting. |

### Route

| Field | Type | Description |
|-------|------|-------------|
| `path` | `string` | URL path pattern (e.g., `"/users/{userId}"`, `"/health"`). |
| `methods` | `string[]` | HTTP methods this route handles (e.g., `["GET", "POST"]`). When empty, all methods are accepted. |
| `backend` | `Backend` | Backend that processes matching requests. Required. |
| `authorization` | `RouteAuthorization` | Per-route authorization policy. |
| `loggingPolicies` | `LoggingPolicies` | Per-route logging override. |

### Backend

| Field | Type | Description |
|-------|------|-------------|
| `type` | `enum` | Backend type. Values: `http`, `oracle_functions`, `stock_response`. |
| `url` | `string` | Target URL for HTTP backends. Required when `type` is `http`. |
| `functionId` | `string` | OCID of the OCI function. Required when `type` is `oracle_functions`. |
| `status` | `int32` | HTTP status code for stock responses. |
| `body` | `string` | Response body for stock responses. |
| `connectTimeoutInSeconds` | `float` | Connection timeout. Applicable to `http` and `oracle_functions`. |
| `readTimeoutInSeconds` | `float` | Read timeout. Applicable to `http` and `oracle_functions`. |
| `sendTimeoutInSeconds` | `float` | Send timeout. Applicable to `http` and `oracle_functions`. |
| `isSslVerifyDisabled` | `bool` | Skip TLS verification for HTTP backends. |
| `headers` | `BackendHeader[]` | Custom headers added to backend requests. |

### RequestPolicies (Deployment-Level)

| Field | Type | Description |
|-------|------|-------------|
| `authentication` | `Authentication` | JWT-based authentication. See below. |
| `cors` | `CorsPolicy` | Cross-Origin Resource Sharing policy. See below. |
| `rateLimiting` | `RateLimiting` | Rate limiting policy. See below. |

### Authentication (JWT)

| Field | Type | Description |
|-------|------|-------------|
| `issuers` | `string[]` | Allowed token issuers (iss claim). |
| `audiences` | `string[]` | Allowed audiences (aud claim). |
| `tokenHeader` | `string` | HTTP header containing the token. Defaults to `"Authorization"`. |
| `tokenQueryParam` | `string` | Query parameter containing the token. |
| `tokenAuthScheme` | `string` | Auth scheme prefix (e.g., `"Bearer"`). Defaults to `"Bearer"`. |
| `maxClockSkewInSeconds` | `float` | Clock skew tolerance for exp/nbf/iat claims. |
| `isAnonymousAccessAllowed` | `bool` | Allow unauthenticated requests (routes can still enforce authorization). |
| `publicKeys` | `PublicKeys` | Public key configuration for signature verification. Required. |
| `verifyClaims` | `VerifyClaim[]` | Additional claims to verify. |

### CorsPolicy

| Field | Type | Description |
|-------|------|-------------|
| `allowedOrigins` | `string[]` | Allowed origins (e.g., `["*"]` or `["https://app.example.com"]`). Min 1. |
| `allowedMethods` | `string[]` | Allowed HTTP methods. |
| `allowedHeaders` | `string[]` | Allowed request headers. |
| `exposedHeaders` | `string[]` | Response headers exposed to the browser. |
| `isAllowCredentialsEnabled` | `bool` | Allow credentials (cookies, auth headers). |
| `maxAgeInSeconds` | `int32` | Preflight response cache duration. |

### RateLimiting

| Field | Type | Description |
|-------|------|-------------|
| `rateInRequestsPerSecond` | `int32` | Maximum requests per second. Must be > 0. |
| `rateKey` | `enum` | Grouping key. Values: `client_ip` (per-IP), `total` (aggregate). |

## Examples

### Public Gateway with HTTP Backend

A public gateway proxying requests to an upstream service:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciApiGateway
metadata:
  name: api-proxy
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciApiGateway.api-proxy
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  endpointType: endpoint_type_public
  subnetId:
    value: "ocid1.subnet.oc1..example"
  deployment:
    pathPrefix: "/api/v1"
    routes:
      - path: "/users"
        methods:
          - "GET"
          - "POST"
        backend:
          type: http
          url: "https://backend.example.com:8080"
      - path: "/users/{userId}"
        methods:
          - "GET"
          - "PUT"
          - "DELETE"
        backend:
          type: http
          url: "https://backend.example.com:8080"
```

### Functions Backend with JWT Authentication

A gateway routing to OCI Functions with JWT authentication via remote JWKS and CORS for a browser client:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciApiGateway
metadata:
  name: serverless-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciApiGateway.serverless-api
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  endpointType: endpoint_type_public
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: public-subnet
      fieldPath: status.outputs.subnetId
  deployment:
    pathPrefix: "/fn"
    loggingPolicies:
      accessLog:
        isEnabled: true
      executionLog:
        isEnabled: true
        logLevel: info
    requestPolicies:
      authentication:
        issuers:
          - "https://auth.example.com/"
        audiences:
          - "https://api.example.com"
        publicKeys:
          type: remote_jwks
          uri: "https://auth.example.com/.well-known/jwks.json"
          maxCacheDurationInHours: 24
      cors:
        allowedOrigins:
          - "https://app.example.com"
        allowedMethods:
          - "GET"
          - "POST"
          - "OPTIONS"
        allowedHeaders:
          - "Authorization"
          - "Content-Type"
        maxAgeInSeconds: 3600
    routes:
      - path: "/process"
        methods:
          - "POST"
        backend:
          type: oracle_functions
          functionId: "ocid1.fnfunc.oc1..example"
        authorization:
          type: authentication_only
      - path: "/health"
        methods:
          - "GET"
        backend:
          type: stock_response
          status: 200
          body: '{"status":"ok"}'
        authorization:
          type: anonymous
```

### Private Gateway with Rate Limiting

A VCN-internal gateway with rate limiting for internal microservice communication:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciApiGateway
metadata:
  name: internal-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciApiGateway.internal-api
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  endpointType: private
  subnetId:
    value: "ocid1.subnet.oc1..example"
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: api-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  deployment:
    pathPrefix: "/internal"
    requestPolicies:
      rateLimiting:
        rateInRequestsPerSecond: 100
        rateKey: client_ip
    routes:
      - path: "/data"
        methods:
          - "GET"
        backend:
          type: http
          url: "http://data-service.internal:8080"
          connectTimeoutInSeconds: 5
          readTimeoutInSeconds: 30
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `gateway_id` | `string` | OCID of the API gateway |
| `hostname` | `string` | Hostname assigned by OCI (e.g., `"abc123.apigateway.us-ashburn-1.oci.customer-oci.com"`) |
| `deployment_endpoint` | `string` | Full endpoint URL (gateway hostname + deployment path prefix) |

## Related Components

- [OciSubnet](/docs/catalog/oci/ocisubnet) — provides the subnet referenced by `subnetId` via `valueFrom`
- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — provides NSGs referenced by `networkSecurityGroupIds` via `valueFrom`
- [OciFunctionsApplication](/docs/catalog/oci/ocifunctionsapplication) — hosts the functions invoked by `oracle_functions` backends
- [OciDnsRecord](/docs/catalog/oci/ocidnsrecord) — creates CNAME records pointing to the gateway hostname
