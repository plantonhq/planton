# Multi-Route Lambda API with CORS

This preset creates an HTTP API Gateway with multiple routes to different Lambda functions, configured with CORS for browser access. API Gateway handles routing, allowing each Lambda to focus on a specific endpoint.

## When to Use

- RESTful APIs with multiple endpoints
- APIs accessed from web browsers (requires CORS)
- APIs where routing logic lives in API Gateway (not Lambda)
- Production APIs with separate Lambda functions per endpoint

## Key Configuration Choices

- **Multiple routes** — Separate routes for each HTTP method and path pattern
- **Path parameters** — `{userId}` syntax for dynamic path segments
- **CORS enabled** — Allows browser-based clients to access the API
- **Credentials support** — `allow_credentials: true` enables cookies and auth headers
- **Preflight caching** — `max_age_seconds: 3600` caches OPTIONS responses for 1 hour

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-frontend-url>` | URL of your frontend application (e.g., "https://app.example.com") | Your frontend deployment URL |
| `<get-users-lambda-arn>` | ARN of Lambda function for GET /users | `AwsLambda` status outputs |
| `<create-user-lambda-arn>` | ARN of Lambda function for POST /users | `AwsLambda` status outputs |
| `<get-user-lambda-arn>` | ARN of Lambda function for GET /users/{userId} | `AwsLambda` status outputs |
| `<update-user-lambda-arn>` | ARN of Lambda function for PUT /users/{userId} | `AwsLambda` status outputs |
| `<delete-user-lambda-arn>` | ARN of Lambda function for DELETE /users/{userId} | `AwsLambda` status outputs |

## Route Patterns

- **`GET /users`** — List all users
- **`POST /users`** — Create a new user
- **`GET /users/{userId}`** — Get a specific user (userId passed as path parameter)
- **`PUT /users/{userId}`** — Update a specific user
- **`DELETE /users/{userId}`** — Delete a specific user

## CORS Configuration

The CORS configuration allows:
- Requests from your frontend URL
- Standard HTTP methods (GET, POST, PUT, DELETE)
- Content-Type and Authorization headers
- Credentials (cookies, auth headers)
- Preflight caching for 1 hour

**To allow multiple origins**, add them to `allow_origins`:
```yaml
allow_origins:
  - "https://app.example.com"
  - "https://staging.example.com"
```

**To allow all origins** (not recommended for production):
```yaml
allow_origins:
  - "*"
```

## Lambda Function Accessing Path Parameters

Lambda functions receive path parameters in the event:

```javascript
exports.handler = async (event) => {
  const userId = event.pathParameters?.userId;
  
  // Use userId to fetch user data
  return {
    statusCode: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ userId, name: "John" })
  };
};
```

## Using Resource References

Reference Lambda functions using `valueFrom`:

```yaml
integration_uri:
  valueFrom:
    kind: AwsLambda
    name: "get-users-function"
    fieldPath: "status.outputs.function_arn"
```

## Related Presets

- **01-default-route-lambda** — Use instead for a single Lambda handling all routes
- **03-jwt-authorized-api** — Use instead when you need JWT authentication
