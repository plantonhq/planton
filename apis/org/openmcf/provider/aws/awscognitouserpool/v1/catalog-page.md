# AWS Cognito User Pool

Deploys an AWS Cognito User Pool with bundled app clients and an optional hosted UI domain. Provides managed user directory services, password-based authentication with configurable MFA, email verification, custom user attributes, and Lambda trigger hooks -- enabling OAuth 2.0 / OIDC token-based authentication for web and mobile applications.

## What Gets Created

When you deploy an AwsCognitoUserPool resource, OpenMCF provisions:

- **Cognito User Pool** -- an `aws_cognito_user_pool` resource with the configured identity model, password policy, MFA settings, email delivery, and optional Lambda triggers
- **App Client(s)** -- one `aws_cognito_user_pool_client` per entry in `spec.clients`, each with its own OAuth flows, scopes, token validity, and security settings
- **User Pool Domain** -- created only when `spec.domain` is set, an `aws_cognito_user_pool_domain` that enables the hosted sign-in UI and OAuth2 endpoints (Authorization, Token, UserInfo)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An ACM certificate in us-east-1** if configuring a custom domain (Cognito uses CloudFront for custom domains)
- **A verified SES identity** if using `emailConfiguration.emailSendingAccount: DEVELOPER` for production email volumes
- **Lambda function(s)** with `cognito-idp.amazonaws.com` invoke permission if configuring Lambda triggers

## Quick Start

Create a file `cognito.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: my-auth
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsCognitoUserPool.my-auth
spec:
  region: us-east-1
  usernameAttributes:
    - email
  autoVerifiedAttributes:
    - email
  clients:
    - name: web-app
      explicitAuthFlows:
        - ALLOW_USER_SRP_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
```

Deploy:

```shell
openmcf apply -f cognito.yaml
```

This creates a user pool where users sign in with their email address, email is auto-verified on sign-up, and a single app client supports SRP authentication with refresh tokens.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | The AWS region where the Cognito User Pool will be created (e.g., `us-east-1`). | Required |
| `clients` | `AwsCognitoUserPoolClient[]` | App clients that authenticate against this pool. At least one required. | Minimum 1 item |
| `clients[].name` | `string` | Client name, used as key in `client_ids` and `client_secrets` output maps. Must be unique across all clients. | 1-128 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `usernameAttributes` | `string[]` | `[]` | Attributes used as username: `"email"`, `"phone_number"`. Mutually exclusive with `aliasAttributes`. ForceNew. |
| `aliasAttributes` | `string[]` | `[]` | Alias identifiers: `"email"`, `"phone_number"`, `"preferred_username"`. Mutually exclusive with `usernameAttributes`. ForceNew. |
| `usernameCaseSensitive` | `bool` | `false` | Case-sensitive usernames. ForceNew. |
| `passwordPolicy.minimumLength` | `int` | 8 (AWS default) | Minimum password length. Range: 6-99. |
| `passwordPolicy.requireLowercase` | `bool` | `false` | Require lowercase letter. |
| `passwordPolicy.requireUppercase` | `bool` | `false` | Require uppercase letter. |
| `passwordPolicy.requireNumbers` | `bool` | `false` | Require digit. |
| `passwordPolicy.requireSymbols` | `bool` | `false` | Require special character. |
| `passwordPolicy.temporaryPasswordValidityDays` | `int` | 7 (AWS default) | Days until admin-created temporary passwords expire. Range: 0-365. |
| `mfaConfiguration` | `string` | `"OFF"` | MFA enforcement: `"OFF"`, `"OPTIONAL"`, `"ON"`. |
| `softwareTokenMfaEnabled` | `bool` | `false` | Enable TOTP authenticator app MFA. Requires `mfaConfiguration` not `"OFF"`. |
| `autoVerifiedAttributes` | `string[]` | `[]` | Auto-verify on sign-up: `"email"`, `"phone_number"`. |
| `accountRecoveryMechanisms` | `object[]` | `[]` | Recovery methods with `.name` (`"verified_email"`, `"verified_phone_number"`, `"admin_only"`) and `.priority` (1-2). |
| `emailConfiguration.emailSendingAccount` | `string` | `"COGNITO_DEFAULT"` | Email mode: `"COGNITO_DEFAULT"` (50/day sandbox) or `"DEVELOPER"` (SES). |
| `emailConfiguration.sourceArn` | `StringValueOrRef` | — | SES identity ARN. Required for `"DEVELOPER"` mode. |
| `emailConfiguration.fromEmailAddress` | `string` | — | "From" address for DEVELOPER mode. |
| `emailConfiguration.replyToEmailAddress` | `string` | — | Reply-to address. |
| `emailConfiguration.configurationSet` | `string` | — | SES configuration set for delivery metrics. |
| `allowAdminCreateUserOnly` | `bool` | `false` | Disable self-registration. |
| `deletionProtection` | `bool` | `false` | Prevent accidental pool deletion. |
| `customAttributes` | `object[]` | `[]` | Custom user attributes. See README for schema fields. |
| `lambdaConfig.preSignUp` | `StringValueOrRef` | — | Lambda ARN for pre-sign-up hook. Can reference AwsLambda via `valueFrom`. |
| `lambdaConfig.preAuthentication` | `StringValueOrRef` | — | Lambda ARN for pre-authentication hook. |
| `lambdaConfig.postAuthentication` | `StringValueOrRef` | — | Lambda ARN for post-authentication hook. |
| `lambdaConfig.postConfirmation` | `StringValueOrRef` | — | Lambda ARN for post-confirmation hook. |
| `lambdaConfig.preTokenGeneration` | `StringValueOrRef` | — | Lambda ARN for pre-token-generation hook. |
| `lambdaConfig.customMessage` | `StringValueOrRef` | — | Lambda ARN for custom message hook. |
| `lambdaConfig.userMigration` | `StringValueOrRef` | — | Lambda ARN for user migration hook. |
| `lambdaConfig.defineAuthChallenge` | `StringValueOrRef` | — | Lambda ARN for define-auth-challenge hook. |
| `lambdaConfig.createAuthChallenge` | `StringValueOrRef` | — | Lambda ARN for create-auth-challenge hook. |
| `lambdaConfig.verifyAuthChallengeResponse` | `StringValueOrRef` | — | Lambda ARN for verify-auth-challenge-response hook. |
| `clients[].generateSecret` | `bool` | `false` | Generate client secret. ForceNew. True for server-side apps, false for SPAs/mobile. |
| `clients[].allowedOauthFlowsUserPoolClient` | `bool` | `false` | Enable OAuth flows for this client. |
| `clients[].allowedOauthFlows` | `string[]` | `[]` | OAuth grant types: `"code"`, `"implicit"`, `"client_credentials"`. |
| `clients[].allowedOauthScopes` | `string[]` | `[]` | OAuth scopes: `"openid"`, `"email"`, `"profile"`, custom scopes. |
| `clients[].callbackUrls` | `string[]` | `[]` | OAuth redirect URIs after authentication. |
| `clients[].logoutUrls` | `string[]` | `[]` | Redirect URIs after sign-out. |
| `clients[].defaultRedirectUri` | `string` | — | Default callback URL (must be in `callbackUrls`). |
| `clients[].supportedIdentityProviders` | `string[]` | — | Identity providers: `"COGNITO"`, social provider names. |
| `clients[].explicitAuthFlows` | `string[]` | `[]` | Auth APIs: `"ALLOW_USER_SRP_AUTH"`, `"ALLOW_REFRESH_TOKEN_AUTH"`, etc. |
| `clients[].accessTokenValidityMinutes` | `int` | 60 | Access token TTL in minutes. Range: 5-1440. |
| `clients[].idTokenValidityMinutes` | `int` | 60 | ID token TTL in minutes. Range: 5-1440. |
| `clients[].refreshTokenValidityDays` | `int` | 30 | Refresh token TTL in days. Range: 1-3650. |
| `clients[].enableTokenRevocation` | `bool` | `false` | Revoke tokens on sign-out. |
| `clients[].preventUserExistenceErrors` | `string` | — | `"ENABLED"` prevents user enumeration attacks. `"LEGACY"` for backward compatibility. |
| `domain.domain` | `string` | — | Cognito prefix (e.g., `"myapp-auth"`) or custom FQDN (e.g., `"auth.example.com"`). ForceNew. |
| `domain.certificateArn` | `StringValueOrRef` | — | ACM cert ARN for custom domains. Required when domain contains a dot. Must be in us-east-1. Can reference AwsCertManagerCert via `valueFrom`. |

## Examples

### Email Sign-In with OAuth Hosted UI

A web application using the Cognito-hosted sign-in page with Authorization Code flow:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: webapp-auth
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: webapp
    pulumi.openmcf.org/stack.name: staging.AwsCognitoUserPool.webapp-auth
spec:
  region: us-east-1
  usernameAttributes:
    - email
  passwordPolicy:
    minimumLength: 10
    requireLowercase: true
    requireUppercase: true
    requireNumbers: true
  autoVerifiedAttributes:
    - email
  accountRecoveryMechanisms:
    - name: verified_email
      priority: 1
  clients:
    - name: web-app
      allowedOauthFlowsUserPoolClient: true
      allowedOauthFlows:
        - code
      allowedOauthScopes:
        - openid
        - email
        - profile
      callbackUrls:
        - https://staging.example.com/callback
      logoutUrls:
        - https://staging.example.com/logout
      explicitAuthFlows:
        - ALLOW_USER_SRP_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
      enableTokenRevocation: true
      preventUserExistenceErrors: ENABLED
  domain:
    domain: acme-staging-auth
```

### Production with MFA and Multiple Clients

A hardened production pool with optional MFA, SES email, a public SPA client and a confidential server client:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: prod-auth
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.AwsCognitoUserPool.prod-auth
spec:
  region: us-east-1
  usernameAttributes:
    - email
  passwordPolicy:
    minimumLength: 12
    requireLowercase: true
    requireUppercase: true
    requireNumbers: true
    requireSymbols: true
    temporaryPasswordValidityDays: 3
  mfaConfiguration: OPTIONAL
  softwareTokenMfaEnabled: true
  autoVerifiedAttributes:
    - email
  accountRecoveryMechanisms:
    - name: verified_email
      priority: 1
  emailConfiguration:
    emailSendingAccount: DEVELOPER
    sourceArn: "arn:aws:ses:us-east-1:123456789012:identity/noreply@example.com"
    fromEmailAddress: "Acme <noreply@example.com>"
  deletionProtection: true
  clients:
    - name: web-spa
      allowedOauthFlowsUserPoolClient: true
      allowedOauthFlows:
        - code
      allowedOauthScopes:
        - openid
        - email
        - profile
      callbackUrls:
        - https://app.example.com/callback
      logoutUrls:
        - https://app.example.com/logout
      explicitAuthFlows:
        - ALLOW_USER_SRP_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
      accessTokenValidityMinutes: 60
      idTokenValidityMinutes: 60
      refreshTokenValidityDays: 30
      enableTokenRevocation: true
      preventUserExistenceErrors: ENABLED
    - name: api-server
      generateSecret: true
      allowedOauthFlowsUserPoolClient: true
      allowedOauthFlows:
        - client_credentials
      allowedOauthScopes:
        - api/read
        - api/write
      accessTokenValidityMinutes: 30
      enableTokenRevocation: true
  domain:
    domain: acme-prod-auth
```

### API Gateway JWT Integration (valueFrom Pattern)

Shows how an AwsHttpApiGateway references this pool's endpoint and client ID for JWT authorization:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: my-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.AwsHttpApiGateway.my-api
spec:
  routes:
    - routeKey: "GET /users"
      integration:
        integrationType: AWS_PROXY
        integrationUri:
          valueFrom:
            kind: AwsLambda
            name: get-users
            fieldPath: status.outputs.function_arn
      authorizationType: JWT
      authorizerName: cognito
  authorizers:
    - name: cognito
      authorizerType: JWT
      jwtConfiguration:
        issuer:
          valueFrom:
            kind: AwsCognitoUserPool
            name: prod-auth
            fieldPath: status.outputs.user_pool_endpoint
        audiences:
          - valueFrom:
              kind: AwsCognitoUserPool
              name: prod-auth
              fieldPath: status.outputs.client_ids.web-spa
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `user_pool_id` | `string` | User pool identifier (e.g., `us-east-1_Ab1Cd2EfG`). Used in SDK configurations and IAM policies. |
| `user_pool_arn` | `string` | User pool ARN. Used in IAM policies and cross-service permissions. |
| `user_pool_endpoint` | `string` | OIDC issuer endpoint URL. Used as the JWT `issuer` in API Gateway authorizers. |
| `user_pool_domain` | `string` | Full hosted UI domain URL. Empty when no domain is configured. |
| `cloudfront_distribution_arn` | `string` | CloudFront distribution ARN for custom domains. Used for Route53 alias records. Empty for prefix domains. |
| `client_ids` | `map<string, string>` | Map of client name to client ID. Access specific clients via `client_ids.{name}`. |
| `client_secrets` | `map<string, string>` | Map of client name to client secret. Only populated for clients with `generateSecret: true`. Sensitive. |

## Related Components

- [AWS Lambda](/docs/catalog/aws/lambda) -- Lambda functions for Cognito triggers (pre-sign-up, post-confirmation, pre-token-generation)
- [AWS HTTP API Gateway](/docs/catalog/aws/http-api-gateway) -- uses `user_pool_endpoint` as JWT issuer and `client_ids` as audiences for JWT authorization
- [AWS ACM Certificate](/docs/catalog/aws/cert-manager-cert) -- ACM certificate for custom domains (must be in us-east-1)
- [AWS IAM Role](/docs/catalog/aws/iam-role) -- execution roles for Lambda triggers
