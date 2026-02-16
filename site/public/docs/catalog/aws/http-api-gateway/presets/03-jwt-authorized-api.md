---
title: "JWT Authorized API with Cognito"
description: "This preset creates an HTTP API Gateway with JWT authorization using Amazon Cognito User Pool. Routes can be public (no auth) or protected (require valid JWT token). Protected routes can also require..."
type: "preset"
rank: "03"
presetSlug: "03-jwt-authorized-api"
componentSlug: "http-api-gateway"
componentTitle: "HTTP API Gateway"
provider: "aws"
icon: "package"
order: 3
---

# JWT Authorized API with Cognito

This preset creates an HTTP API Gateway with JWT authorization using Amazon Cognito User Pool. Routes can be public (no auth) or protected (require valid JWT token). Protected routes can also require specific OAuth scopes.

## When to Use

- APIs requiring user authentication
- APIs integrated with Cognito User Pools
- APIs with both public and protected endpoints
- APIs needing fine-grained access control via OAuth scopes

## Key Configuration Choices

- **JWT authorizer** — Validates JSON Web Tokens from Cognito
- **Mixed authorization** — Public route (`/public`) and protected routes (`/profile`, `/orders`)
- **OAuth scopes** — Route-level scope requirements (`orders:write`)
- **CORS with credentials** — Enables browser-based authenticated requests
- **Cognito integration** — Native Cognito User Pool support

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-frontend-url>` | URL of your frontend application | Your frontend deployment URL |
| `<public-handler-lambda-arn>` | ARN of Lambda for public endpoint | `AwsLambda` status outputs |
| `<profile-handler-lambda-arn>` | ARN of Lambda for profile endpoint | `AwsLambda` status outputs |
| `<orders-handler-lambda-arn>` | ARN of Lambda for orders endpoint | `AwsLambda` status outputs |
| `<cognito-issuer-url>` | Cognito User Pool issuer URL | Format: `https://cognito-idp.{region}.amazonaws.com/{userPoolId}` |
| `<cognito-app-client-id>` | Cognito App Client ID | AWS Console → Cognito → User Pools → Your Pool → App Integration → App Clients |

## Finding Cognito Values

### User Pool ID and Issuer URL

1. Go to AWS Console → Cognito → User Pools
2. Select your User Pool
3. Go to "General Settings"
4. Copy the "User pool ID" (format: `us-east-1_ABC123XYZ`)
5. Construct issuer URL: `https://cognito-idp.{region}.amazonaws.com/{userPoolId}`

### App Client ID

1. In your User Pool, go to "App Integration" → "App Clients"
2. Copy the "Client ID" (format: `1a2b3c4d5e6f7g8h9i0j`)

## Route Authorization Levels

- **`GET /public`** — No authorization (`authorization_type: "NONE"`)
- **`GET /profile`** — Requires valid JWT token (no scopes)
- **`POST /orders`** — Requires valid JWT token AND `orders:write` scope

## JWT Token Format

Clients must send JWT tokens in the Authorization header:

```
Authorization: Bearer <jwt-token>
```

The JWT token must:
- Be issued by the Cognito User Pool (valid `iss` claim)
- Have the correct audience (`aud` claim matches App Client ID)
- Not be expired
- Include required scopes (if route specifies `authorization_scopes`)

## Lambda Function Accessing User Claims

Lambda functions receive JWT claims in the request context:

```javascript
exports.handler = async (event) => {
  // Access user ID from JWT claims
  const userId = event.requestContext.authorizer?.claims?.sub;
  const email = event.requestContext.authorizer?.claims?.email;
  
  return {
    statusCode: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ userId, email })
  };
};
```

## OAuth Scopes

Routes can require specific OAuth scopes:

```yaml
authorization_scopes:
  - "orders:write"
  - "admin:access"
```

The JWT token must include all required scopes in the `scope` claim. If scopes are missing, API Gateway returns `403 Forbidden`.

## Using Resource References

Reference Lambda functions and other resources:

```yaml
integration_uri:
  valueFrom:
    kind: AwsLambda
    name: "profile-handler"
    fieldPath: "status.outputs.function_arn"
```

## Testing the API

### Public Endpoint (No Auth)
```bash
curl https://<api-id>.execute-api.<region>.amazonaws.com/public
```

### Protected Endpoint (With JWT)
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  https://<api-id>.execute-api.<region>.amazonaws.com/profile
```

### Protected Endpoint (With Scopes)
```bash
curl -X POST \
  -H "Authorization: Bearer <jwt-token-with-orders:write-scope>" \
  https://<api-id>.execute-api.<region>.amazonaws.com/orders
```

## Related Presets

- **01-default-route-lambda** — Use instead for simple APIs without authentication
- **02-multi-route-lambda** — Use instead for multi-route APIs without authentication
