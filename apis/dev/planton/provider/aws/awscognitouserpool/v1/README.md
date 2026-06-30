# AWS Cognito User Pool

Deploys an AWS Cognito User Pool with bundled app clients and an optional hosted UI domain. Provides user directory services, password-based authentication, MFA, custom attributes, and Lambda trigger hooks for web and mobile applications.

## When to Use

- You need managed user authentication for a web or mobile application
- You want OAuth 2.0 / OIDC token-based authentication without managing your own identity server
- You need the Cognito hosted UI for login/signup pages
- You need JWT tokens for API Gateway authorization

## What Gets Created

- **Cognito User Pool** -- the user directory with password policy, MFA settings, email verification, and optional Lambda triggers
- **App Client(s)** -- one or more OAuth/OIDC client configurations for applications to authenticate against the pool
- **User Pool Domain** (optional) -- a Cognito-hosted prefix domain or custom domain for the hosted sign-in UI and token endpoints

## Prerequisites

- AWS credentials with permissions for `cognito-idp:*`
- (Optional) A verified SES identity if using DEVELOPER email mode for production sending volumes
- (Optional) Lambda functions if configuring trigger hooks (must grant Cognito `lambda:InvokeFunction` permission)
- (Optional) An ACM certificate in **us-east-1** if configuring a custom domain

## ForceNew Fields (Cannot Change After Creation)

These fields destroy and recreate the user pool or sub-resource if changed:

- `username_attributes` / `alias_attributes` -- the identity model
- `username_case_sensitive` -- case sensitivity of usernames
- Custom attribute `mutable` and `required` flags
- Domain name
- App client `generate_secret`

## Spec Reference

### Identity Model

| Field | Type | Description |
|-------|------|-------------|
| `usernameAttributes` | `string[]` | Sign-in identifiers: `"email"`, `"phone_number"`. Mutually exclusive with `aliasAttributes`. ForceNew. |
| `aliasAttributes` | `string[]` | Alias identifiers: `"email"`, `"phone_number"`, `"preferred_username"`. Mutually exclusive with `usernameAttributes`. ForceNew. |
| `usernameCaseSensitive` | `bool` | Case-sensitive usernames. Default: `false`. ForceNew. |

### Password Policy

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `passwordPolicy.minimumLength` | `int32` | 8 | Minimum password length (6-99) |
| `passwordPolicy.requireLowercase` | `bool` | `false` | Require lowercase letter |
| `passwordPolicy.requireUppercase` | `bool` | `false` | Require uppercase letter |
| `passwordPolicy.requireNumbers` | `bool` | `false` | Require digit |
| `passwordPolicy.requireSymbols` | `bool` | `false` | Require special character |
| `passwordPolicy.temporaryPasswordValidityDays` | `int32` | 7 | Days until temporary passwords expire (0-365) |

### MFA

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mfaConfiguration` | `string` | `"OFF"` | MFA enforcement: `"OFF"`, `"OPTIONAL"`, `"ON"` |
| `softwareTokenMfaEnabled` | `bool` | `false` | Enable TOTP authenticator app MFA |

### Verification and Recovery

| Field | Type | Description |
|-------|------|-------------|
| `autoVerifiedAttributes` | `string[]` | Attributes to auto-verify: `"email"`, `"phone_number"` |
| `accountRecoveryMechanisms` | `object[]` | Recovery methods with `.name` and `.priority` (1-2) |

### Email Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `emailConfiguration.emailSendingAccount` | `string` | `"COGNITO_DEFAULT"` | `"COGNITO_DEFAULT"` or `"DEVELOPER"` (SES) |
| `emailConfiguration.sourceArn` | `StringValueOrRef` | | SES identity ARN (required for DEVELOPER) |
| `emailConfiguration.fromEmailAddress` | `string` | | "From" address for DEVELOPER mode |
| `emailConfiguration.replyToEmailAddress` | `string` | | Reply-to address |
| `emailConfiguration.configurationSet` | `string` | | SES configuration set |

### Admin and Protection

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `allowAdminCreateUserOnly` | `bool` | `false` | Disable self-registration |
| `deletionProtection` | `bool` | `false` | Prevent accidental pool deletion |

### Custom Attributes

| Field | Type | Description |
|-------|------|-------------|
| `customAttributes[].name` | `string` | Attribute name (1-20 chars, auto-prefixed with `custom:`) |
| `customAttributes[].attributeDataType` | `string` | `"String"`, `"Number"`, `"DateTime"`, `"Boolean"` |
| `customAttributes[].mutable` | `bool` | Whether value can be changed after creation. ForceNew. |
| `customAttributes[].required` | `bool` | Whether required during registration. ForceNew. |
| `customAttributes[].stringMinLength` | `string` | Min string length (String type) |
| `customAttributes[].stringMaxLength` | `string` | Max string length (String type) |
| `customAttributes[].numberMinValue` | `string` | Min numeric value (Number type) |
| `customAttributes[].numberMaxValue` | `string` | Max numeric value (Number type) |

### Lambda Triggers

All fields in `lambdaConfig` accept a Lambda function ARN or `valueFrom` reference to an AwsLambda resource.

| Field | Description |
|-------|-------------|
| `lambdaConfig.preSignUp` | Custom validation before sign-up |
| `lambdaConfig.preAuthentication` | Custom validation before auth |
| `lambdaConfig.postAuthentication` | Logic after successful auth |
| `lambdaConfig.postConfirmation` | Logic after user confirmation |
| `lambdaConfig.preTokenGeneration` | Modify token claims |
| `lambdaConfig.customMessage` | Customize verification/invitation messages |
| `lambdaConfig.userMigration` | Migrate users from external provider |
| `lambdaConfig.defineAuthChallenge` | Define custom auth challenge |
| `lambdaConfig.createAuthChallenge` | Create custom auth challenge |
| `lambdaConfig.verifyAuthChallengeResponse` | Verify custom challenge response |

### App Clients

| Field | Type | Description |
|-------|------|-------------|
| `clients[].name` | `string` | Client name (map key for outputs). Required. |
| `clients[].generateSecret` | `bool` | Generate client secret. ForceNew. |
| `clients[].allowedOauthFlowsUserPoolClient` | `bool` | Enable OAuth flows |
| `clients[].allowedOauthFlows` | `string[]` | `"code"`, `"implicit"`, `"client_credentials"` |
| `clients[].allowedOauthScopes` | `string[]` | OAuth scopes (e.g., `"openid"`, `"email"`) |
| `clients[].callbackUrls` | `string[]` | OAuth redirect URLs |
| `clients[].logoutUrls` | `string[]` | Post-sign-out redirect URLs |
| `clients[].defaultRedirectUri` | `string` | Default callback URL |
| `clients[].supportedIdentityProviders` | `string[]` | Identity providers (default: `["COGNITO"]`) |
| `clients[].explicitAuthFlows` | `string[]` | Enabled auth APIs |
| `clients[].accessTokenValidityMinutes` | `int32` | Access token TTL in minutes (5-1440) |
| `clients[].idTokenValidityMinutes` | `int32` | ID token TTL in minutes (5-1440) |
| `clients[].refreshTokenValidityDays` | `int32` | Refresh token TTL in days (1-3650) |
| `clients[].enableTokenRevocation` | `bool` | Revoke tokens on sign-out |
| `clients[].preventUserExistenceErrors` | `string` | `"ENABLED"` or `"LEGACY"` |

### Domain

| Field | Type | Description |
|-------|------|-------------|
| `domain.domain` | `string` | Cognito prefix or custom FQDN. ForceNew. |
| `domain.certificateArn` | `StringValueOrRef` | ACM cert ARN (required for custom domains, must be in us-east-1) |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `user_pool_id` | `string` | User pool identifier (e.g., `us-east-1_Ab1Cd2EfG`) |
| `user_pool_arn` | `string` | User pool ARN |
| `user_pool_endpoint` | `string` | OIDC issuer URL for JWT validation |
| `user_pool_domain` | `string` | Full domain URL for hosted UI |
| `cloudfront_distribution_arn` | `string` | CloudFront ARN for custom domain DNS |
| `client_ids` | `map<string, string>` | Client name to client ID map |
| `client_secrets` | `map<string, string>` | Client name to client secret map (sensitive) |

## Deliberately Omitted (v1)

- **Identity providers** (social/OIDC/SAML) -- separate lifecycle, planned as AwsCognitoIdentityProvider
- **Resource servers** (custom OAuth scopes) -- niche M2M pattern
- **SMS configuration** -- requires separate IAM role and SNS; most deployments start email-only
- **Advanced security** (user pool add-ons) -- paid feature
- **WebAuthn, Email MFA, Sign-in policy** -- newer features with low adoption
- **Verification message templates** -- defaults work well for most use cases
- **Device configuration** -- defaults work well for most use cases
- **UI customization** (CSS/images) -- console task
