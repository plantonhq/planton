# AWS HTTP API Gateway: Architecture and Deep Dive

This document provides a comprehensive technical overview of AWS API Gateway HTTP APIs, their architecture, integration patterns, authorization mechanisms, and best practices. It serves as a reference for understanding how the AwsHttpApiGateway component works under the hood.

## Table of Contents

1. [HTTP API Gateway Overview](#http-api-gateway-overview)
2. [Architecture and Components](#architecture-and-components)
3. [Routes and Integrations](#routes-and-integrations)
4. [Authorization Mechanisms](#authorization-mechanisms)
5. [Stages and Deployments](#stages-and-deployments)
6. [Comparison: HTTP API vs REST API](#comparison-http-api-vs-rest-api)
7. [Integration Patterns](#integration-patterns)
8. [Performance and Cost](#performance-and-cost)
9. [Security Considerations](#security-considerations)
10. [Best Practices](#best-practices)

---

## HTTP API Gateway Overview

AWS API Gateway HTTP APIs (API Gateway v2) are a simplified, cost-effective API Gateway service optimized for building RESTful APIs and HTTP proxy APIs. Introduced in 2019, HTTP APIs address many limitations of REST APIs while maintaining compatibility with common integration patterns.

### Key Characteristics

- **Lower Latency**: Up to 60% lower latency compared to REST APIs
- **Lower Cost**: Up to 70% cost reduction compared to REST APIs
- **Simplified Model**: Streamlined API surface focused on common use cases
- **Native Features**: Built-in CORS, automatic deployments, JWT authorization
- **Lambda-First**: Optimized for Lambda proxy integration

### Use Cases

HTTP APIs excel in these scenarios:

1. **Serverless APIs**: Lambda-backed REST APIs
2. **HTTP Proxies**: Adding API Gateway features to existing HTTP services
3. **Microservices Gateways**: Routing to multiple backend services
4. **JWT-Authenticated APIs**: APIs using Cognito, Auth0, or OIDC providers
5. **High-Throughput APIs**: APIs requiring low latency and high request rates

---

## Architecture and Components

### Core Components

An HTTP API Gateway consists of several interconnected components:

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP API Gateway                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐│
│  │   Routes     │───▶│ Integrations │───▶│  Backends     ││
│  │              │    │              │    │  (Lambda/HTTP)││
│  └──────────────┘    └──────────────┘    └──────────────┘│
│         │                    │                             │
│         ▼                    ▼                             │
│  ┌──────────────┐    ┌──────────────┐                    │
│  │ Authorizers  │    │     Stage     │                    │
│  │ (JWT/REQUEST)│    │  (Deployment) │                    │
│  └──────────────┘    └──────────────┘                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### API

The **API** is the top-level container that defines:
- API name and description
- CORS configuration (applied globally)
- Protocol (HTTP or WebSocket)
- Endpoint configuration

Each API has a unique identifier (`api_id`) and ARN.

### Routes

**Routes** define request patterns that match incoming requests. A route consists of:

- **Route Key**: Pattern matching HTTP method and path
  - `GET /users` - Specific method and path
  - `POST /orders/{id}` - Path with parameter
  - `$default` - Catch-all route for unmatched requests
  - `ANY /api/*` - Wildcard matching

- **Integration**: Backend target (Lambda or HTTP endpoint)
- **Authorization**: Authorization type and authorizer reference

Routes are evaluated in order, with the first matching route handling the request.

### Integrations

**Integrations** connect routes to backend services. HTTP APIs support two integration types:

#### AWS_PROXY Integration

Lambda proxy integration where API Gateway:
1. Receives HTTP request
2. Transforms to Lambda event format
3. Invokes Lambda function
4. Transforms Lambda response to HTTP response

**Event Format (v2.0)**:
```json
{
  "version": "2.0",
  "routeKey": "GET /users",
  "rawPath": "/users",
  "rawQueryString": "limit=10",
  "headers": {
    "content-type": "application/json",
    "authorization": "Bearer token"
  },
  "queryStringParameters": {
    "limit": "10"
  },
  "pathParameters": {
    "userId": "123"
  },
  "requestContext": {
    "http": {
      "method": "GET",
      "path": "/users",
      "protocol": "HTTP/1.1"
    },
    "requestId": "abc123",
    "time": "09/Feb/2025:12:00:00 +0000"
  },
  "body": "{\"name\":\"John\"}",
  "isBase64Encoded": false
}
```

#### HTTP_PROXY Integration

HTTP proxy integration where API Gateway:
1. Receives HTTP request
2. Forwards to upstream HTTP endpoint
3. Returns upstream response

Supports path rewriting, header manipulation, and query string forwarding.

### Authorizers

**Authorizers** validate and authorize requests before they reach integrations.

#### JWT Authorizers

JWT authorizers validate JSON Web Tokens from OIDC providers:

1. Extract token from `identity_sources` (typically `$request.header.Authorization`)
2. Validate token signature using issuer's public keys
3. Verify `iss` (issuer) matches configured issuer
4. Verify `aud` (audience) matches configured audiences
5. Check token expiration
6. Extract claims and pass to integration

**Supported Providers**:
- Amazon Cognito User Pools
- Auth0
- Any OIDC-compliant provider

**JWT Claims Available in Integration**:
- `$context.authorizer.claims.*` - All JWT claims
- `$context.authorizer.claims.sub` - User ID
- `$context.authorizer.claims.email` - User email

#### REQUEST Authorizers

Lambda authorizers invoke a Lambda function for custom authorization:

1. Extract identity from `identity_sources`
2. Invoke Lambda authorizer function with request context
3. Lambda returns authorization decision:
   - Simple response: `{"isAuthorized": true, "context": {...}}`
   - IAM policy: Full IAM policy document
4. Cache result for `result_ttl_seconds`
5. Pass context to integration

**Lambda Authorizer Event (v2.0)**:
```json
{
  "version": "2.0",
  "type": "REQUEST",
  "routeArn": "arn:aws:execute-api:...",
  "identitySource": ["Bearer token"],
  "routeKey": "GET /users",
  "rawPath": "/users",
  "headers": {...},
  "requestContext": {...}
}
```

### Stages

**Stages** represent deployment environments. Each stage:
- Has a unique name (e.g., `$default`, `prod`, `dev`)
- Contains deployed routes and integrations
- Has its own invoke URL
- Can have stage-specific settings (logging, throttling, variables)

The `$default` stage is special:
- Automatically created
- Invoke URL matches API endpoint (no stage prefix)
- Supports auto-deploy (changes deploy immediately)

---

## Routes and Integrations

### Route Matching

Routes are matched in the order they are defined:

1. **Exact Match**: `GET /users` matches `GET /users` exactly
2. **Path Parameters**: `GET /users/{userId}` matches `GET /users/123`
3. **Wildcards**: `ANY /api/*` matches any path under `/api`
4. **Default Route**: `$default` matches any unmatched request

**Example**:
```yaml
routes:
  - route_key: "GET /users/{id}"      # Matches GET /users/123
  - route_key: "GET /users"            # Matches GET /users
  - route_key: "ANY /api/*"            # Matches any /api path
  - route_key: "$default"              # Catches everything else
```

### Integration Deduplication

When multiple routes share identical integration configuration (same type, URI, payload format), the IaC modules automatically create a single integration resource and reference it from multiple routes. This reduces resource count and simplifies management.

### Payload Format Versions

#### Version 2.0 (Recommended)

- Simplified event structure
- Direct body access (no base64 encoding)
- Cleaner request context
- Better performance

**Body Access**:
```javascript
// Lambda handler
exports.handler = async (event) => {
  const body = JSON.parse(event.body); // Direct access
  // ...
};
```

#### Version 1.0 (Legacy)

- Compatible with REST API format
- Base64-encoded body for binary content
- Multi-value headers
- More verbose structure

**When to Use 1.0**:
- Migrating from REST API
- Need multi-value headers
- Binary content handling

---

## Authorization Mechanisms

### Authorization Types

#### NONE (Default)

No authorization required. All requests proceed to integration.

**Use Cases**:
- Public APIs
- Health check endpoints
- Public documentation

#### JWT

JSON Web Token authorization using a JWT authorizer.

**Flow**:
1. Client sends request with `Authorization: Bearer <token>` header
2. API Gateway extracts token from `identity_sources`
3. JWT authorizer validates token:
   - Signature verification (using issuer's public keys)
   - Issuer (`iss`) validation
   - Audience (`aud`) validation
   - Expiration check
4. If valid, request proceeds with JWT claims in context
5. If invalid, returns 401 Unauthorized

**OAuth Scopes**:
Routes can require specific OAuth scopes:
```yaml
authorization_scopes:
  - "users:read"
  - "users:write"
```

The JWT must contain all required scopes in the `scope` claim.

#### AWS_IAM

AWS Signature Version 4 (SigV4) authorization. Requires requests to be signed with AWS credentials.

**Use Cases**:
- Service-to-service communication
- AWS SDK/CLI access
- Cross-account access

**Request Signing**:
```bash
aws apigatewayv2 invoke-api \
  --api-id <api-id> \
  --route-key "GET /admin" \
  --region us-east-1
```

---

## Stages and Deployments

### Stage Lifecycle

1. **Create Stage**: Define stage name and configuration
2. **Deploy Routes**: Routes and integrations are deployed to stage
3. **Auto-Deploy**: Changes automatically deploy (if enabled)
4. **Manual Deploy**: Explicit deployment required (if auto-deploy disabled)

### Auto-Deploy

When `auto_deploy: true`:
- Changes to routes, integrations, or authorizers deploy immediately
- No explicit deployment step required
- Suitable for development and simple production setups

When `auto_deploy: false`:
- Changes require explicit deployment
- Better control over when changes go live
- Suitable for production with change management processes

### Access Logging

Access logs capture request/response information:

**Log Format Variables**:
- `$context.requestId` - Unique request ID
- `$context.identity.sourceIp` - Client IP address
- `$context.httpMethod` - HTTP method
- `$context.routeKey` - Matched route key
- `$context.status` - HTTP status code
- `$context.responseLatency` - Total latency
- `$context.integrationLatency` - Backend latency
- `$context.error.message` - Error message

**Example JSON Format**:
```json
{
  "requestId": "abc123",
  "ip": "203.0.113.1",
  "requestTime": "09/Feb/2025:12:00:00 +0000",
  "httpMethod": "GET",
  "routeKey": "GET /users",
  "status": 200,
  "protocol": "HTTP/1.1",
  "responseLatency": 150,
  "integrationLatency": 120
}
```

### Throttling

Throttling limits request rate to protect backends:

- **Burst Limit**: Maximum concurrent requests (token bucket size)
- **Rate Limit**: Steady-state requests per second (token refill rate)

**Example**:
```yaml
default_throttle:
  burst_limit: 5000      # Allow 5000 concurrent requests
  rate_limit: 2000.0     # Refill at 2000 req/sec
```

When limit exceeded, API Gateway returns `429 Too Many Requests`.

### Stage Variables

Stage variables pass environment-specific configuration:

```yaml
stage_variables:
  environment: "production"
  api_version: "v1"
  backend_url: "https://api.example.com"
```

Variables accessible in integration request parameters:
- `$stageVariables.environment`
- `$stageVariables.api_version`

---

## Comparison: HTTP API vs REST API

### When to Use HTTP API

✅ **Use HTTP API when**:
- Building new APIs
- Cost optimization is important
- Low latency is critical
- Lambda proxy integration
- JWT authorization
- Simple routing needs

### When to Use REST API

✅ **Use REST API when**:
- Need API keys and usage plans
- Request/response transformations
- Custom authorizers (non-Lambda)
- API caching
- Request validation
- Existing REST API migration

### Feature Comparison

| Feature | HTTP API | REST API |
|---------|----------|----------|
| **Cost** | Lower (up to 70% cheaper) | Higher |
| **Latency** | Lower (up to 60% faster) | Higher |
| **Lambda Integration** | Optimized | Supported |
| **JWT Authorization** | Native | Via Lambda authorizer |
| **CORS** | Built-in | Custom integration response |
| **API Keys** | ❌ | ✅ |
| **Usage Plans** | ❌ | ✅ |
| **Request Validation** | ❌ | ✅ |
| **Response Caching** | ❌ | ✅ |
| **Request/Response Transformations** | ❌ | ✅ |
| **Custom Domain** | ✅ | ✅ |
| **WebSocket** | ✅ (separate API type) | ❌ |

---

## Integration Patterns

### Lambda Proxy Pattern

Most common pattern for serverless APIs:

```yaml
routes:
  - route_key: "GET /users"
    integration:
      integration_type: "AWS_PROXY"
      integration_uri:
        value: "arn:aws:lambda:us-east-1:123456789012:function:get-users"
      payload_format_version: "2.0"
```

**Lambda Handler**:
```javascript
exports.handler = async (event) => {
  const userId = event.pathParameters?.userId;
  
  return {
    statusCode: 200,
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ userId, name: "John" })
  };
};
```

### HTTP Proxy Pattern

Proxy to upstream HTTP service:

```yaml
routes:
  - route_key: "GET /api/*"
    integration:
      integration_type: "HTTP_PROXY"
      integration_uri:
        value: "https://api.example.com"
```

**Use Cases**:
- Adding API Gateway features to existing APIs
- Rate limiting and throttling
- CORS for legacy APIs
- Request/response logging

### Multi-Route Pattern

Multiple routes to different backends:

```yaml
routes:
  - route_key: "GET /users"
    integration:
      integration_type: "AWS_PROXY"
      integration_uri:
        value: "arn:aws:lambda:...:function:get-users"
  - route_key: "GET /orders"
    integration:
      integration_type: "AWS_PROXY"
      integration_uri:
        value: "arn:aws:lambda:...:function:get-orders"
```

### Authorized Pattern

Routes with JWT authorization:

```yaml
routes:
  - route_key: "GET /profile"
    authorization_type: "JWT"
    authorizer_name: "cognito-authorizer"
    integration:
      integration_type: "AWS_PROXY"
      integration_uri:
        value: "arn:aws:lambda:...:function:get-profile"
authorizers:
  - name: "cognito-authorizer"
    authorizer_type: "JWT"
    jwt_configuration:
      issuer: "https://cognito-idp.us-east-1.amazonaws.com/..."
      audiences: ["app-client-id"]
```

---

## Performance and Cost

### Performance Characteristics

**Latency Breakdown**:
- API Gateway overhead: ~10-50ms
- Lambda cold start: ~100-1000ms (first invocation)
- Lambda warm execution: ~10-100ms
- Total (warm): ~20-150ms
- Total (cold): ~110-1050ms

**Optimization Strategies**:
1. Use payload format 2.0 (lower overhead)
2. Keep Lambda functions warm (provisioned concurrency)
3. Minimize integration timeout
4. Use JWT authorizers (faster than Lambda authorizers)
5. Cache authorizer results

### Cost Structure

**HTTP API Pricing** (as of 2025):
- **API Calls**: $1.00 per million requests
- **Data Transfer**: Standard AWS data transfer rates
- **Authorizer Invocations**: Included in API call cost

**REST API Pricing** (for comparison):
- **API Calls**: $3.50 per million requests
- **Data Transfer**: Standard AWS data transfer rates
- **Cache**: Additional charges

**Cost Optimization**:
1. Use HTTP API instead of REST API (70% cost savings)
2. Cache authorizer results (reduce Lambda invocations)
3. Use JWT authorizers (no Lambda invocation cost)
4. Optimize Lambda execution time (reduce duration charges)
5. Use CloudFront for static content (reduce API Gateway calls)

---

## Security Considerations

### Authentication and Authorization

1. **Public Endpoints**: Use `authorization_type: "NONE"` only for truly public endpoints
2. **JWT Validation**: Always validate issuer and audience
3. **Token Storage**: Never log tokens in access logs
4. **Scope Validation**: Enforce OAuth scopes for fine-grained access
5. **IAM Authorization**: Use for service-to-service communication

### Network Security

1. **HTTPS Only**: API Gateway enforces HTTPS
2. **VPC Integration**: Use VPC links for private backend access
3. **IP Restrictions**: Not natively supported (use WAF)
4. **Rate Limiting**: Configure throttling to prevent abuse

### Data Protection

1. **Sensitive Data**: Never log sensitive data in access logs
2. **PII Handling**: Mask PII in logs and responses
3. **Encryption**: Use TLS 1.2+ (enforced by API Gateway)
4. **Secrets**: Use AWS Secrets Manager for sensitive configuration

### Best Practices

1. **Least Privilege**: Grant minimum permissions to Lambda functions
2. **Input Validation**: Validate inputs in Lambda functions
3. **Error Handling**: Don't expose internal errors to clients
4. **Monitoring**: Enable CloudWatch logging and metrics
5. **Audit**: Review access logs regularly

---

## Best Practices

### API Design

1. **RESTful Routes**: Use standard HTTP methods and resource-based paths
2. **Versioning**: Include API version in path (`/api/v1/users`)
3. **Consistent Naming**: Use consistent naming conventions
4. **Error Responses**: Return consistent error response format
5. **Documentation**: Document API endpoints and request/response formats

### Performance

1. **Payload Format 2.0**: Use version 2.0 for better performance
2. **Timeout Configuration**: Set appropriate timeouts per route
3. **Authorizer Caching**: Cache authorizer results (300+ seconds)
4. **Lambda Optimization**: Optimize Lambda cold starts and execution time
5. **Connection Pooling**: Reuse HTTP connections in HTTP_PROXY integrations

### Reliability

1. **Error Handling**: Implement retry logic in clients
2. **Circuit Breakers**: Use circuit breakers for downstream services
3. **Monitoring**: Set up CloudWatch alarms for errors and latency
4. **Health Checks**: Implement health check endpoints
5. **Graceful Degradation**: Handle backend failures gracefully

### Cost Optimization

1. **HTTP API**: Use HTTP API instead of REST API
2. **Authorizer Caching**: Reduce Lambda authorizer invocations
3. **JWT Authorizers**: Prefer JWT over Lambda authorizers
4. **Throttling**: Configure appropriate throttling limits
5. **Monitoring**: Monitor costs using AWS Cost Explorer

### Security

1. **Authorization**: Always use authorization for protected endpoints
2. **CORS**: Configure CORS restrictively
3. **Input Validation**: Validate all inputs
4. **Logging**: Log security events (failed auth, rate limits)
5. **Updates**: Keep dependencies and Lambda runtimes updated

---

## Additional Resources

- [AWS API Gateway HTTP API Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api.html)
- [API Gateway Pricing](https://aws.amazon.com/api-gateway/pricing/)
- [Lambda Best Practices](https://docs.aws.amazon.com/lambda/latest/dg/best-practices.html)
- [JWT.io](https://jwt.io/) - JWT debugging and validation
- [OpenMCF Documentation](https://openmcf.org/docs)
