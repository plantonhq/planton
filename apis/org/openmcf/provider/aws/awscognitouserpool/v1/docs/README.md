# AWS Cognito User Pool -- Architecture and Design

## Overview

Amazon Cognito User Pools is a fully managed user directory that provides sign-up, sign-in, and token-based authentication for web and mobile applications. It implements the OAuth 2.0 and OpenID Connect (OIDC) standards, enabling applications to authenticate users and receive JWTs (JSON Web Tokens) without managing any identity infrastructure.

This component bundles three tightly coupled resources into a single declarative unit: the user pool (directory), app clients (application configurations), and an optional domain (hosted UI endpoints).

## Architecture

### Identity Model

Cognito offers two mutually exclusive identity models, and the choice is **permanent** (ForceNew):

**Username Attributes** (`username_attributes`): The specified attribute(s) become the username. With `["email"]`, users sign in with their email address, and Cognito auto-generates an internal UUID as the "sub" claim. This is the most common model for consumer-facing applications.

**Alias Attributes** (`alias_attributes`): Users have a separate, explicit username (typically auto-generated or user-chosen), and the alias attributes serve as alternative sign-in identifiers. With `["email", "preferred_username"]`, users can sign in with either their email, preferred username, or their unique username. This model is common for gaming, social, or enterprise applications where usernames are meaningful.

**Neither**: If both are omitted, users must provide a unique username at sign-up. This is the least common model but useful when usernames are managed externally.

### Authentication Flows

Cognito supports several authentication flows controlled by `explicit_auth_flows` on the app client:

- **USER_SRP_AUTH**: Secure Remote Password protocol. The password never leaves the client. Recommended for all applications.
- **REFRESH_TOKEN_AUTH**: Refresh token exchange. Always recommended to enable session continuity.
- **USER_PASSWORD_AUTH**: Direct username/password transmission. Less secure; only for server-side applications behind TLS.
- **ADMIN_USER_PASSWORD_AUTH**: Admin-initiated authentication. For server-side applications that authenticate users on behalf of themselves.
- **CUSTOM_AUTH**: Custom authentication flow using Lambda triggers (define/create/verify challenge). For biometrics, passwordless, or multi-step flows.

### Token Model

Cognito issues three JWTs per successful authentication:

- **ID Token**: Contains user identity claims (email, name, custom attributes). Used by applications to identify the user. Default TTL: 1 hour.
- **Access Token**: Contains OAuth scopes and group memberships. Used for API authorization. Default TTL: 1 hour.
- **Refresh Token**: Used to obtain new ID/access tokens without re-authentication. Default TTL: 30 days. Cannot exceed 10 years.

### MFA Architecture

Cognito supports three MFA enforcement levels:

- **OFF**: No MFA. Users authenticate with password only.
- **OPTIONAL**: Users can opt in to MFA during sign-in. Once enrolled, MFA is required on subsequent sign-ins.
- **ON**: All users must enroll in MFA. Authentication fails without a valid MFA code.

This component supports TOTP (Time-based One-Time Password) via authenticator apps. SMS-based MFA is excluded from v1 as it requires a separate IAM role and SNS configuration.

### Email Delivery

Cognito sends emails for verification codes, password resets, and invitation messages. Two modes:

- **COGNITO_DEFAULT**: Cognito's built-in email service. Limited to 50 emails/day in sandbox mode. No setup required. Suitable for development and low-volume applications.
- **DEVELOPER**: Routes emails through your SES verified identity. No daily limit (subject to SES limits). Required for production applications.

### Custom Attributes

Cognito supports up to 50 custom attributes per pool, automatically prefixed with `custom:`. Key constraints:

- **Data types**: String, Number, DateTime, Boolean
- **Immutable flags**: The `mutable` and `required` settings cannot be changed after the attribute is added (ForceNew at the attribute level, not the pool level)
- **Schema is append-only**: You can add new attributes but cannot remove or rename existing ones

### Lambda Triggers

Cognito invokes Lambda functions at 10 lifecycle points, enabling deep customization:

- **Pre Sign-Up**: Validate/auto-confirm users, block sign-ups
- **Pre/Post Authentication**: Custom validation, logging, analytics
- **Post Confirmation**: Welcome emails, provisioning, downstream system sync
- **Pre Token Generation**: Add/remove/modify JWT claims (e.g., inject tenant_id)
- **Custom Message**: Customize verification and invitation email/SMS content
- **User Migration**: Import users on-the-fly from an external identity store
- **Custom Auth Challenge**: Define/create/verify custom authentication challenges

### Domain and Hosted UI

The domain configuration enables Cognito's hosted sign-in UI and standard OAuth 2.0 endpoints (`/oauth2/authorize`, `/oauth2/token`, `/oauth2/userInfo`). Two types:

- **Cognito Prefix Domain**: `{prefix}.auth.{region}.amazoncognito.com`. Free, no DNS setup. Good for development and internal tools.
- **Custom Domain**: Your own domain (e.g., `auth.example.com`). Requires an ACM certificate in **us-east-1** (Cognito uses CloudFront). Outputs the CloudFront distribution ARN for creating a Route53 alias record.

## App Client Design

App clients are the bridge between applications and the user pool. Each client represents a distinct application with its own:

- **Secret**: Server-side apps use a client secret for the Authorization Code grant. SPAs and mobile apps must not have a secret (`generate_secret: false`).
- **OAuth flows**: Authorization Code (`code`) for most apps, Client Credentials for M2M, Implicit is deprecated.
- **Scopes**: Standard OIDC scopes (`openid`, `email`, `profile`) plus custom scopes from resource servers.
- **Token TTLs**: Each client can have different access/ID/refresh token lifetimes.

## Infra Chart Composability

This component produces outputs that enable downstream resource wiring:

**As a JWT issuer for API Gateway**:
```
spec.authorizers[].jwtConfiguration.issuer -> status.outputs.user_pool_endpoint
spec.authorizers[].jwtConfiguration.audiences -> status.outputs.client_ids.{name}
```

**As environment variables for ECS/EKS deployments**:
```
COGNITO_USER_POOL_ID -> status.outputs.user_pool_id
COGNITO_CLIENT_ID -> status.outputs.client_ids.{name}
COGNITO_DOMAIN -> status.outputs.user_pool_domain
```

## Cost Model

Cognito pricing is based on Monthly Active Users (MAUs):

- **Free tier**: 50,000 MAUs (LITE tier) or 10,000 MAUs (ESSENTIALS tier)
- **Beyond free tier**: ~$0.0055/MAU (varies by tier and volume)
- **Advanced security features**: Additional cost per MAU
- **Custom domain**: No additional Cognito cost (ACM certificates are free)

## Security Best Practices

1. Always enable `transit_encryption` (TLS) for token endpoints
2. Use SRP auth (`ALLOW_USER_SRP_AUTH`) -- passwords never leave the client
3. Enable token revocation for production clients
4. Set `preventUserExistenceErrors: ENABLED` to prevent user enumeration
5. Use `DEVELOPER` email mode for production (SES) to avoid 50/day sandbox limit
6. Enable deletion protection for production pools
7. MFA `OPTIONAL` at minimum for production; `ON` for sensitive applications
