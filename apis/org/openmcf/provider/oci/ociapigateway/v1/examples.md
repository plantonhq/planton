# OciApiGateway Examples

## Simple HTTP Proxy

A public gateway proxying all requests to a backend service:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciApiGateway
metadata:
  name: proxy-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciApiGateway.proxy-api
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  endpointType: public
  subnetId:
    value: "ocid1.subnet.oc1..example"
  deployment:
    pathPrefix: "/api"
    routes:
      - path: "/"
        backend:
          type: http
          url: "https://backend.example.com:8080"
```

## Serverless API with JWT and CORS

A public gateway routing to OCI Functions with JWT authentication via a remote JWKS endpoint and CORS for browser clients:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciApiGateway
metadata:
  name: fn-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciApiGateway.fn-api
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  endpointType: public
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: public-subnet
      fieldPath: status.outputs.subnetId
  deployment:
    pathPrefix: "/fn/v1"
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
        isAnonymousAccessAllowed: true
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
      - path: "/public-data"
        methods:
          - "GET"
        backend:
          type: oracle_functions
          functionId: "ocid1.fnfunc.oc1..example-public"
        authorization:
          type: anonymous
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

## Private Gateway with Rate Limiting

A VCN-internal gateway with per-IP rate limiting and NSG binding for internal microservice communication:

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
      - path: "/ingest"
        methods:
          - "POST"
        backend:
          type: http
          url: "http://ingest-service.internal:8080"
          headers:
            - name: "X-Source"
              value: "api-gateway"
```

## Scope-Based Route Authorization

A gateway with JWT authentication where different routes require different OAuth2 scopes:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciApiGateway
metadata:
  name: scoped-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciApiGateway.scoped-api
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  endpointType: public
  subnetId:
    value: "ocid1.subnet.oc1..example"
  deployment:
    pathPrefix: "/api/v2"
    requestPolicies:
      authentication:
        issuers:
          - "https://idcs-abc.identity.oraclecloud.com/"
        audiences:
          - "https://api.example.com"
        publicKeys:
          type: remote_jwks
          uri: "https://idcs-abc.identity.oraclecloud.com/admin/v1/SigningCert/jwk"
        verifyClaims:
          - key: "email_verified"
            values:
              - "true"
            isRequired: true
    routes:
      - path: "/users"
        methods:
          - "GET"
        backend:
          type: http
          url: "https://user-service.example.com"
        authorization:
          type: any_of
          allowedScope:
            - "users:read"
      - path: "/users"
        methods:
          - "POST"
          - "PUT"
        backend:
          type: http
          url: "https://user-service.example.com"
        authorization:
          type: any_of
          allowedScope:
            - "users:write"
```

## Common Operations

### Add a new route

Append a new entry to `deployment.routes` and re-apply. Routes are evaluated in order, so place more specific paths before catch-all patterns.

### Update backend timeouts

Modify `connectTimeoutInSeconds`, `readTimeoutInSeconds`, or `sendTimeoutInSeconds` on the route's backend and re-apply.

### Enable logging on an existing deployment

Add a `loggingPolicies` block to the deployment and re-apply. Access logs capture every request; execution logs capture gateway processing details.

## Best Practices

1. **Use stock responses for health checks** — avoids routing health probes to backend services.
2. **Enable access logging in production** — essential for debugging and audit trails.
3. **Use remote JWKS for authentication** — automatic key rotation without redeploying the gateway.
4. **Set explicit timeouts** — prevents long-running backend requests from exhausting gateway resources.
5. **Use `valueFrom` references** for `subnetId` and `compartmentId` — avoids hardcoding OCIDs.
6. **Order routes from specific to general** — first match wins; place `/users/{id}` before `/users`.
7. **Use per-route authorization** — mix anonymous and authenticated routes in the same deployment.
