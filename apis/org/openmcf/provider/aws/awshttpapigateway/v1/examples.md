# AWS HTTP API Gateway Examples

This document provides practical examples for deploying AWS API Gateway HTTP APIs using OpenMCF. After creating one of these YAML manifests, deploy it using the OpenMCF CLI:

```shell
# Using Pulumi
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>

# Using Terraform/OpenTofu
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Example 1: Minimal Single Route to Lambda

A minimal HTTP API with a single catch-all route (`$default`) that forwards all requests to a Lambda function.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: minimal-api
spec:
  region: us-east-1
  routes:
    - route_key: "$default"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:my-function"
```

**Key Points:**
- Uses `$default` route key to catch all unmatched requests
- AWS_PROXY integration type for Lambda proxy integration
- Payload format version defaults to "2.0" (simplified event structure)
- No authorization required (defaults to "NONE")
- Stage defaults to "$default" with auto-deploy enabled

**Use Case:** Simple serverless API where a single Lambda handles all requests.

---

## Example 2: Multi-Route API with CORS

Multiple routes to different Lambda functions with CORS enabled for browser access.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: multi-route-api
spec:
  region: us-east-1
  description: Multi-route API with CORS support
  cors_configuration:
    allow_origins:
      - "https://app.example.com"
      - "https://staging.example.com"
    allow_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allow_headers:
      - "Content-Type"
      - "Authorization"
      - "X-Requested-With"
    expose_headers:
      - "X-Request-Id"
    allow_credentials: true
    max_age_seconds: 3600
  routes:
    - route_key: "GET /users"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:get-users"
        payload_format_version: "2.0"
    - route_key: "POST /users"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:create-user"
        payload_format_version: "2.0"
    - route_key: "GET /users/{userId}"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:get-user-by-id"
        payload_format_version: "2.0"
    - route_key: "PUT /users/{userId}"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:update-user"
        payload_format_version: "2.0"
    - route_key: "DELETE /users/{userId}"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:delete-user"
        payload_format_version: "2.0"
```

**Key Points:**
- Multiple routes with specific HTTP methods and paths
- Path parameters (`{userId}`) are passed to Lambda in the event
- CORS configured for browser access from specific origins
- Credentials allowed for authenticated requests
- Preflight caching set to 1 hour (3600 seconds)

**Use Case:** RESTful API with multiple endpoints accessible from web applications.

---

## Example 3: JWT Authorization with Cognito

Routes protected with JWT authorization using Amazon Cognito User Pool.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: jwt-protected-api
spec:
  region: us-east-1
  description: API with JWT authorization via Cognito
  cors_configuration:
    allow_origins:
      - "https://app.example.com"
    allow_methods:
      - "GET"
      - "POST"
      - "OPTIONS"
    allow_headers:
      - "Content-Type"
      - "Authorization"
    allow_credentials: true
  routes:
    - route_key: "GET /public"
      authorization_type: "NONE"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:public-handler"
    - route_key: "GET /profile"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:get-profile"
    - route_key: "POST /orders"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      authorization_scopes:
        - "orders:write"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:create-order"
  authorizers:
    - name: "cognito-authorizer"
      authorizer_type: "JWT"
      jwt_configuration:
        issuer: "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123XYZ"
        audiences:
          - "1a2b3c4d5e6f7g8h9i0j"
      identity_sources:
        - "$request.header.Authorization"
```

**Key Points:**
- Public route (`/public`) with no authorization
- Protected routes require valid JWT token in Authorization header
- OAuth scopes enforced for specific routes (`orders:write`)
- Cognito issuer URL format: `https://cognito-idp.{region}.amazonaws.com/{userPoolId}`
- Audience is the Cognito App Client ID

**Use Case:** API with public and protected endpoints, using Cognito for user authentication.

**Finding Cognito Values:**
- User Pool ID: AWS Console → Cognito → User Pools → Your Pool → General Settings
- App Client ID: AWS Console → Cognito → User Pools → Your Pool → App Integration → App Clients

---

## Example 4: Lambda REQUEST Authorizer

Custom authorization logic using a Lambda function authorizer.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: lambda-authorizer-api
spec:
  region: us-east-1
  description: API with Lambda REQUEST authorizer
  routes:
    - route_key: "GET /data"
      authorization_type: "JWT"
      authorizer_name: "custom-authorizer"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:get-data"
    - route_key: "POST /data"
      authorization_type: "JWT"
      authorizer_name: "custom-authorizer"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:create-data"
  authorizers:
    - name: "custom-authorizer"
      authorizer_type: "REQUEST"
      authorizer_uri:
        value: "arn:aws:lambda:us-east-1:123456789012:function:authorizer-function"
      authorizer_credentials_arn:
        value: "arn:aws:iam::123456789012:role/api-gateway-authorizer-role"
      identity_sources:
        - "$request.header.Authorization"
        - "$request.querystring.apiKey"
      result_ttl_seconds: 300
      enable_simple_responses: true
      authorizer_payload_format_version: "2.0"
```

**Key Points:**
- REQUEST authorizer invokes Lambda function for authorization
- IAM role required for API Gateway to invoke authorizer Lambda
- Multiple identity sources (header and query string)
- Simple responses enabled (returns `{"isAuthorized": true/false}`)
- Authorizer result cached for 5 minutes (300 seconds)

**Use Case:** Custom authorization logic (API keys, custom tokens, complex permission checks).

**Lambda Authorizer Response Format (Simple Responses):**
```json
{
  "isAuthorized": true,
  "context": {
    "userId": "user123",
    "role": "admin"
  }
}
```

---

## Example 5: Production-Ready API with Logging and Throttling

Complete production configuration with access logging, throttling, CORS, and JWT authorization.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: production-api
spec:
  region: us-east-1
  description: Production API with comprehensive configuration
  cors_configuration:
    allow_origins:
      - "https://app.example.com"
    allow_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "PATCH"
      - "OPTIONS"
    allow_headers:
      - "Content-Type"
      - "Authorization"
      - "X-Requested-With"
      - "X-Request-Id"
    expose_headers:
      - "X-Request-Id"
      - "X-Rate-Limit-Remaining"
    allow_credentials: true
    max_age_seconds: 3600
  stage:
    name: "$default"
    auto_deploy: true
    access_log:
      destination_arn:
        valueFrom:
          kind: AwsCloudwatchLogGroup
          name: "api-gateway-logs"
          fieldPath: "status.outputs.log_group_arn"
      format: '{"requestId":"$context.requestId","ip":"$context.identity.sourceIp","requestTime":"$context.requestTime","httpMethod":"$context.httpMethod","routeKey":"$context.routeKey","status":"$context.status","protocol":"$context.protocol","responseLatency":"$context.responseLatency","integrationLatency":"$context.integrationLatency","integrationStatus":"$context.integrationStatus","errorMessage":"$context.error.message","errorMessageString":"$context.error.messageString"}'
    default_throttle:
      burst_limit: 5000
      rate_limit: 2000.0
    stage_variables:
      environment: "production"
      version: "v1"
  routes:
    - route_key: "GET /health"
      authorization_type: "NONE"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "health-check-function"
            fieldPath: "status.outputs.function_arn"
        payload_format_version: "2.0"
        timeout_milliseconds: 3000
    - route_key: "GET /api/v1/users"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "get-users-function"
            fieldPath: "status.outputs.function_arn"
        payload_format_version: "2.0"
        timeout_milliseconds: 5000
    - route_key: "POST /api/v1/users"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      authorization_scopes:
        - "users:write"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "create-user-function"
            fieldPath: "status.outputs.function_arn"
        payload_format_version: "2.0"
        timeout_milliseconds: 10000
    - route_key: "GET /api/v1/users/{userId}"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "get-user-function"
            fieldPath: "status.outputs.function_arn"
        payload_format_version: "2.0"
    - route_key: "PUT /api/v1/users/{userId}"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      authorization_scopes:
        - "users:write"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "update-user-function"
            fieldPath: "status.outputs.function_arn"
        payload_format_version: "2.0"
    - route_key: "DELETE /api/v1/users/{userId}"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      authorization_scopes:
        - "users:admin"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "delete-user-function"
            fieldPath: "status.outputs.function_arn"
        payload_format_version: "2.0"
  authorizers:
    - name: "cognito-authorizer"
      authorizer_type: "JWT"
      jwt_configuration:
        issuer: "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123XYZ"
        audiences:
          - "1a2b3c4d5e6f7g8h9i0j"
      identity_sources:
        - "$request.header.Authorization"
```

**Key Points:**
- Comprehensive CORS configuration for production web app
- Access logging to CloudWatch Logs with detailed JSON format
- Throttling: 2000 requests/second steady-state, 5000 burst
- Health check endpoint without authorization
- Resource references using `valueFrom` for Lambda functions and Log Group
- OAuth scopes for fine-grained authorization
- Different timeout values per route based on expected processing time
- Stage variables for environment-specific configuration

**Use Case:** Production API requiring observability, rate limiting, and fine-grained access control.

---

## Example 6: HTTP Proxy Integration

Proxy requests to an upstream HTTP service instead of Lambda.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: http-proxy-api
spec:
  region: us-east-1
  description: API Gateway proxying to upstream HTTP service
  routes:
    - route_key: "GET /api/*"
      integration:
        integration_type: "HTTP_PROXY"
        integration_uri:
          value: "https://api.example.com"
        integration_method: "GET"
        timeout_milliseconds: 10000
    - route_key: "POST /api/*"
      integration:
        integration_type: "HTTP_PROXY"
        integration_uri:
          value: "https://api.example.com"
        integration_method: "POST"
        timeout_milliseconds: 10000
    - route_key: "PUT /api/*"
      integration:
        integration_type: "HTTP_PROXY"
        integration_uri:
          value: "https://api.example.com"
        integration_method: "PUT"
        timeout_milliseconds: 10000
```

**Key Points:**
- HTTP_PROXY integration type for upstream HTTP services
- Wildcard route (`/api/*`) proxies all paths under `/api`
- Integration method matches route method by default
- Useful for API Gateway features (CORS, throttling, logging) in front of existing APIs

**Use Case:** Adding API Gateway features (CORS, throttling, logging) to existing HTTP APIs without modifying backend.

---

## Example 7: Mixed Authorization Types

API with public routes, JWT-protected routes, and AWS IAM authorization.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: mixed-auth-api
spec:
  region: us-east-1
  description: API with multiple authorization types
  routes:
    - route_key: "GET /public"
      authorization_type: "NONE"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:public-handler"
    - route_key: "GET /user-data"
      authorization_type: "JWT"
      authorizer_name: "cognito-authorizer"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:user-data-handler"
    - route_key: "POST /admin"
      authorization_type: "AWS_IAM"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          value: "arn:aws:lambda:us-east-1:123456789012:function:admin-handler"
  authorizers:
    - name: "cognito-authorizer"
      authorizer_type: "JWT"
      jwt_configuration:
        issuer: "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123XYZ"
        audiences:
          - "1a2b3c4d5e6f7g8h9i0j"
      identity_sources:
        - "$request.header.Authorization"
```

**Key Points:**
- Public route (`/public`) with no authorization
- JWT-protected route (`/user-data`) for end-user access
- AWS_IAM authorization (`/admin`) requires SigV4 signing (used by AWS SDKs, CLI)
- Different authorization types per route based on use case

**Use Case:** API serving both public content, authenticated users, and AWS service-to-service calls.

---

## Resource References Example

Demonstrates OpenMCF's foreign key system for automatic resource linking.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: ref-api
spec:
  region: us-east-1
  stage:
    access_log:
      destination_arn:
        valueFrom:
          kind: AwsCloudwatchLogGroup
          name: "api-logs"
          fieldPath: "status.outputs.log_group_arn"
  routes:
    - route_key: "GET /users"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "get-users-function"
            fieldPath: "status.outputs.function_arn"
    - route_key: "POST /users"
      integration:
        integration_type: "AWS_PROXY"
        integration_uri:
          valueFrom:
            kind: AwsLambda
            name: "create-user-function"
            fieldPath: "status.outputs.function_arn"
  authorizers:
    - name: "lambda-authorizer"
      authorizer_type: "REQUEST"
      authorizer_uri:
        valueFrom:
          kind: AwsLambda
          name: "authorizer-function"
          fieldPath: "status.outputs.function_arn"
      authorizer_credentials_arn:
        valueFrom:
          kind: AwsIamRole
          name: "api-gateway-authorizer-role"
          fieldPath: "status.outputs.role_arn"
```

**Key Points:**
- Lambda function ARNs automatically referenced from `AwsLambda` resources
- CloudWatch Log Group ARN automatically referenced from `AwsCloudwatchLogGroup` resource
- IAM role ARN automatically referenced from `AwsIamRole` resource
- Enables dependency management and resource linking
- Reduces manual ARN/ID management
- Ensures correct deployment order

---

## Verification Commands

After deploying your HTTP API, verify it using these commands:

### AWS CLI Verification

```shell
# Get API details
aws apigatewayv2 get-api --api-id <api-id>

# List routes
aws apigatewayv2 get-routes --api-id <api-id>

# Get stage details
aws apigatewayv2 get-stage --api-id <api-id> --stage-name "$default"

# Test API endpoint
curl https://<api-id>.execute-api.<region>.amazonaws.com/users
```

### Testing with Authentication

```shell
# Test JWT-protected endpoint
curl -H "Authorization: Bearer <jwt-token>" \
  https://<api-id>.execute-api.<region>.amazonaws.com/users

# Test AWS IAM authorization (requires AWS credentials)
aws apigatewayv2 invoke-api \
  --api-id <api-id> \
  --route-key "GET /admin" \
  --region <region>
```

---

## Best Practices

1. **Use Payload Format Version 2.0**: Simplified event structure with direct body access
2. **Enable Access Logging**: Essential for debugging and monitoring in production
3. **Configure Throttling**: Protect backend from traffic spikes
4. **Use Resource References**: Leverage `valueFrom` for automatic dependency management
5. **Set Appropriate Timeouts**: Match timeout to expected Lambda execution time
6. **Enable CORS for Web Apps**: Configure CORS when API is accessed from browsers
7. **Use JWT for User Authentication**: Prefer JWT authorizers over Lambda authorizers when possible (lower latency, lower cost)
8. **Cache Authorizer Results**: Set `result_ttl_seconds` to reduce Lambda invocations
9. **Use Stage Variables**: Pass environment-specific configuration to integrations
10. **Monitor API Metrics**: Use CloudWatch metrics for API Gateway performance

---

## Common Issues and Solutions

### CORS Errors

**Symptoms**: Browser shows CORS errors when calling API

**Solutions**:
- Ensure `cors_configuration` is set with correct `allow_origins`
- Include "OPTIONS" in `allow_methods` for preflight requests
- Set `allow_credentials: true` if sending cookies/auth headers
- Verify `allow_headers` includes all headers sent by client

### 401 Unauthorized

**Symptoms**: Requests to protected routes return 401

**Solutions**:
- Verify JWT token is valid and not expired
- Check issuer URL matches Cognito User Pool
- Ensure audience matches App Client ID
- Verify `identity_sources` matches where token is sent (header vs query string)

### 504 Gateway Timeout

**Symptoms**: API returns 504 after 30 seconds

**Solutions**:
- Increase `timeout_milliseconds` on integration (max 30000ms)
- Optimize Lambda function execution time
- Check Lambda function logs for errors
- Verify Lambda has sufficient memory allocation

### Integration Not Found

**Symptoms**: Route returns error about missing integration

**Solutions**:
- Verify `integration_uri` is correct Lambda ARN or HTTP URL
- Check Lambda function exists and is in same region
- Ensure API Gateway has permission to invoke Lambda (check Lambda resource policy)

---

## Additional Resources

- [AWS API Gateway HTTP API Documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api.html)
- [API Gateway Access Logging](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-logging.html)
- [JWT Authorizers](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-jwt-authorizer.html)
- [Lambda Authorizers](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-lambda-authorizer.html)
- [OpenMCF Documentation](https://openmcf.org/docs)
