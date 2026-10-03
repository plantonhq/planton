# AWS Cognito User Pool

Deploy and manage an Amazon Cognito user pool using Planton -- the managed user directory and OpenID Connect token issuer for web and mobile applications, covering the full pool surface: identity model, password and passwordless sign-in policies, MFA (TOTP, email, SMS, passkeys), email/SMS delivery, verification and invitation messaging, custom schema attributes, Lambda triggers, threat protection, log delivery, and the hosted-UI domain.

## Overview

The user pool is the ROOT of the Cognito family. Resources with their own AWS lifecycle compose onto it as separate kinds:

- **`AwsCognitoUserPoolClient`** -- app clients (many per pool). Each carries its own OAuth contract and its own client ID that JWT authorizers and ALB authentication actions reference.
- **`AwsCognitoIdentityProvider`** -- federated sign-in (many per pool): Google, Facebook, Amazon, Apple, OIDC, SAML.
- **`AwsCognitoResourceServer`** -- custom OAuth scopes for APIs (many per pool), the scope vocabulary machine-to-machine clients request.

The hosted-UI **domain** stays folded here: AWS allows one domain per pool and the domain string is its identity, so it shares the pool's lifecycle.

## When to Use

- Managed sign-up/sign-in with JWTs for APIs -- without running an identity server.
- OAuth 2.0 / OIDC flows against the Cognito hosted UI or your own custom auth pages.
- Passwordless sign-in (email/SMS one-time codes, passkeys) via `allowedFirstAuthFactors`.
- The issuer behind an `AwsHttpApiGateway` JWT authorizer or an ALB `authenticate_cognito` action.

## Prerequisites

- (Optional) A verified SES identity for `emailConfiguration.emailSendingAccount: DEVELOPER` -- required for production email volumes and for email MFA.
- (Optional) An IAM role assumable by `cognito-idp.amazonaws.com` (with an external-ID condition) for any SMS feature (`smsConfiguration`).
- (Optional) An ACM certificate in **us-east-1** for a custom hosted-UI domain (Cognito fronts custom domains with CloudFront).
- (Optional) Lambda functions for triggers -- each must grant Cognito invoke permission (principal `cognito-idp.amazonaws.com`).

## ForceNew Fields (Cannot Change After Creation)

- `usernameAttributes` / `aliasAttributes` -- the identity model. Changing them REPLACES THE POOL and every user in it.
- `usernameCaseSensitive`.
- `domain.domain` -- replacing the domain string recreates the domain (the pool survives).
- Custom attributes are **append-only**: an attribute, once added, can never be modified or removed (its `mutable`/`required`/`developerOnlyAttribute` flags are fixed).

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoUserPool
metadata:
  name: my-auth
  org: my-org
  env: dev
  id: my-auth-dev
spec:
  region: us-east-1
  usernameAttributes:
    - email
  autoVerifiedAttributes:
    - email
  accountRecoveryMechanisms:
    - name: verified_email
      priority: 1
```

App clients are their own resources:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoUserPoolClient
metadata:
  name: web-app
  org: my-org
  env: dev
  id: web-app-dev
spec:
  region: us-east-1
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: my-auth
      fieldPath: status.outputs.user_pool_id
  explicitAuthFlows:
    - ALLOW_USER_SRP_AUTH
    - ALLOW_REFRESH_TOKEN_AUTH
```

## Spec Reference

### Identity model (ForceNew)

| Field | Type | Description |
|-------|------|-------------|
| `usernameAttributes` | `string[]` | Sign-in identifiers: `email`, `phone_number`. Mutually exclusive with `aliasAttributes`. |
| `aliasAttributes` | `string[]` | Alternative sign-in identifiers alongside a real username: `email`, `phone_number`, `preferred_username`. |
| `usernameCaseSensitive` | `bool` | Case-sensitive usernames. AWS default: false. |

### Pool posture

| Field | Type | Description |
|-------|------|-------------|
| `deletionProtection` | `bool` | Pool cannot be deleted while true. Recommended for production. |
| `userPoolTier` | `string` | Feature tier: `LITE`, `ESSENTIALS` (AWS default), `PLUS` (threat protection). |

### Password and sign-in policy

| Field | Type | Description |
|-------|------|-------------|
| `passwordPolicy.minimumLength` | `int32` | 6-99. AWS default: 8. |
| `passwordPolicy.requireLowercase` / `requireUppercase` / `requireNumbers` / `requireSymbols` | `bool` | Character-class requirements. |
| `passwordPolicy.passwordHistorySize` | `int32` | 0-24 previous passwords a new one must differ from (ESSENTIALS+). |
| `passwordPolicy.temporaryPasswordValidityDays` | `int32` | 0-365. AWS default: 7. |
| `allowedFirstAuthFactors` | `string[]` | The passwordless dial: `PASSWORD`, `EMAIL_OTP`, `SMS_OTP`, `WEB_AUTHN`. Empty keeps classic password-first sign-in. |

### MFA

| Field | Type | Description |
|-------|------|-------------|
| `mfaConfiguration` | `string` | `OFF` (default), `OPTIONAL`, `ON`. |
| `softwareTokenMfaEnabled` | `bool` | TOTP authenticator apps. Requires MFA not OFF. |
| `emailMfa` | `object` | Email OTP as second factor (`message` with `{####}`, `subject`). Requires MFA not OFF and SES DEVELOPER email. |
| `webAuthn.relyingPartyId` | `string` | The domain passkeys bind to -- set before users register passkeys. |
| `webAuthn.userVerification` | `string` | `required` or `preferred`. |

### SMS delivery

| Field | Type | Description |
|-------|------|-------------|
| `smsConfiguration.snsCallerArn` | `StringValueOrRef` | IAM role Cognito assumes to publish SMS via SNS (ref to `AwsIamRole`). Required for any SMS feature. |
| `smsConfiguration.externalId` | `string` | The `sts:ExternalId` the role's trust policy must match. |
| `smsConfiguration.snsRegion` | `string` | Where SMS originates when the pool's region cannot send. |
| `smsAuthenticationMessage` | `string` | Sign-in code SMS body; must contain `{####}`. |

### Verification and recovery

| Field | Type | Description |
|-------|------|-------------|
| `autoVerifiedAttributes` | `string[]` | Auto-verify on sign-up: `email`, `phone_number`. |
| `attributesRequireVerificationBeforeUpdate` | `string[]` | Keep the old value active until the new one is verified. |
| `accountRecoveryMechanisms[]` | `{name, priority}` | `verified_email`, `verified_phone_number`, `admin_only`; priority 1-2. |
| `verificationMessageTemplate` | `object` | Code vs link email verification (`defaultEmailOption`) plus message/subject templates with the `{####}` / `{##link##}` placeholders. |

### Email configuration

| Field | Type | Description |
|-------|------|-------------|
| `emailConfiguration.emailSendingAccount` | `string` | `COGNITO_DEFAULT` (50/day sandbox) or `DEVELOPER` (your SES identity). |
| `emailConfiguration.sourceArn` | `StringValueOrRef` | SES identity ARN; required for DEVELOPER. |
| `emailConfiguration.fromEmailAddress` / `replyToEmailAddress` / `configurationSet` | `string` | Sender identity details. |

### Admin and user management

| Field | Type | Description |
|-------|------|-------------|
| `allowAdminCreateUserOnly` | `bool` | Disable self-registration. |
| `inviteMessageTemplate` | `object` | Invitation email/SMS templates; must contain `{username}` and `{####}`. |
| `deviceConfiguration` | `object` | Remembered-device behavior (`challengeRequiredOnNewDevice`, `deviceOnlyRememberedOnUserPrompt`). |

### Custom attributes (append-only)

| Field | Type | Description |
|-------|------|-------------|
| `customAttributes[].name` | `string` | 1-20 chars; auto-prefixed `custom:`. |
| `customAttributes[].attributeDataType` | `string` | `String`, `Number`, `DateTime`, `Boolean`. |
| `customAttributes[].mutable` / `required` / `developerOnlyAttribute` | `bool` | Fixed at the moment the attribute is added. |
| `customAttributes[].stringMinLength` / `stringMaxLength` / `numberMinValue` / `numberMaxValue` | `string` | Type constraints. |

### Lambda triggers

All trigger fields accept a Lambda ARN or a reference to an `AwsLambda` resource.

| Field | Description |
|-------|-------------|
| `lambdaConfig.preSignUp` / `preAuthentication` / `postAuthentication` / `postConfirmation` | Lifecycle hooks. |
| `lambdaConfig.preTokenGeneration` | Claim customization (V1_0 event). |
| `lambdaConfig.preTokenGenerationConfig` | Versioned claim customization -- `lambdaVersion: V2_0`/`V3_0` also customizes ACCESS tokens. Set this or `preTokenGeneration`, not both. |
| `lambdaConfig.customMessage` / `userMigration` | Message customization; on-the-fly user import. |
| `lambdaConfig.defineAuthChallenge` / `createAuthChallenge` / `verifyAuthChallengeResponse` | Custom auth flow. |
| `lambdaConfig.customEmailSender` / `customSmsSender` | Deliver messages yourself; requires `lambdaConfig.kmsKeyId` (ref to `AwsKmsKey`) -- Cognito encrypts the payload with it. |

### Threat protection and log delivery

| Field | Type | Description |
|-------|------|-------------|
| `userPoolAddOns.advancedSecurityMode` | `string` | `OFF`, `AUDIT`, `ENFORCED`. AUDIT/ENFORCED require the PLUS tier. |
| `userPoolAddOns.customAuthMode` | `string` | Extends threat protection to custom auth flows. |
| `riskConfiguration` | `object` | The POOL-WIDE automated-response policy threat protection acts on (requires `advancedSecurityMode` AUDIT/ENFORCED; responses fire only in ENFORCED). Applies to every app client without a client-scoped policy of its own -- a client overrides on its `AwsCognitoUserPoolClient` spec. |
| `riskConfiguration.accountTakeover` | `object` | Per-risk-level responses (`lowAction`/`mediumAction`/`highAction`, each `BLOCK`, `MFA_IF_CONFIGURED`, `MFA_REQUIRED`, or `NO_ACTION` with a `notify` flag) plus `notifyConfiguration` (SES identity ref + block/MFA/no-action email templates). |
| `riskConfiguration.compromisedCredentials` | `object` | `BLOCK` or `NO_ACTION` when credentials appear in known breach data; `eventFilter` narrows which events are checked (default: all). |
| `riskConfiguration.riskException` | `object` | `blockedIpRanges` (always reject) / `skippedIpRanges` (skip risk evaluation), CIDR notation, up to 200 each. |
| `logConfigurations[]` | `object[]` | Route `userNotification` (ERROR) or `userAuthEvents` (INFO, PLUS tier) to exactly one destination: CloudWatch log group, Firehose stream, or S3 bucket (all refs). |

### User groups

| Field | Type | Description |
|-------|------|-------------|
| `userGroups[]` | `object[]` | Groups in this pool: `name` (unique; renaming replaces the group), `description`, `precedence` (LOWEST wins for the `cognito:preferred_role` claim; AWS accepts 0 but the provider cannot send it -- use 1 as the strongest expressible priority), `roleArn` (ref to `AwsIamRole` -- lands in members' tokens and identity-pool federation). Membership itself is data-plane content managed at runtime, never here. |

### Hosted-UI domain (folded; one per pool)

| Field | Type | Description |
|-------|------|-------------|
| `domain.domain` | `string` | Cognito prefix (`myapp-auth`) or custom FQDN (`auth.example.com`). ForceNew. |
| `domain.certificateArn` | `StringValueOrRef` | ACM cert (us-east-1) -- required for custom domains; ref to `AwsCertManagerCert`. |
| `domain.managedLoginVersion` | `int32` | 1 = classic hosted UI, 2 = managed login (AWS default for new domains). |

## Outputs

| Output | Description |
|--------|-------------|
| `user_pool_id` | Pool identifier (`{region}_{poolId}`) -- the join key for clients, IdPs, and resource servers. |
| `user_pool_arn` | Pool ARN -- IAM policies and ALB `authenticate_cognito` actions. |
| `user_pool_endpoint` | The endpoint as AWS reports it, WITHOUT a scheme (`cognito-idp.{region}.amazonaws.com/{id}`). |
| `issuer` | The full OIDC issuer URL (`https://` + endpoint) -- what JWT authorizers validate the `iss` claim against. |
| `user_pool_domain` | The domain exactly as configured (prefix or FQDN, no scheme) -- what ALB actions take as `user_pool_domain`. |
| `hosted_ui_url` | The full `https://` hosted sign-in URL. |
| `cloudfront_distribution` | CloudFront domain name fronting a custom domain -- the DNS alias target. |
| `cloudfront_distribution_arn` | ARN of that distribution. |
| `cloudfront_hosted_zone_id` | The CloudFront alias-target zone ID for Route53 alias records. |

## Composing the JWT Story

```yaml
# In an AwsHttpApiGateway spec:
authorizers:
  - name: cognito
    authorizerType: JWT
    jwtConfiguration:
      issuer:
        valueFrom:
          kind: AwsCognitoUserPool
          name: my-auth
          fieldPath: status.outputs.issuer
      audiences:
        - valueFrom:
            kind: AwsCognitoUserPoolClient
            name: web-app
            fieldPath: status.outputs.client_id
```

## Deliberately Omitted

- **Users and group memberships** (`aws_cognito_user`, `aws_cognito_user_in_group`): directory CONTENT, not infrastructure shape -- and user passwords would land in state. (Groups THEMSELVES are modeled -- `userGroups` above.)
- **UI customization** (classic hosted UI CSS/images): superseded by managed login's branding designer.
- **Managed-login branding**: client-scoped design-asset payloads with their own artifact lifecycle -- a future kind.
- **Cognito Identity Pools** (`cognitoidentity` service): a separate product surface (AWS-credential federation); composes later as its own kind.
- **Legacy top-level verification message fields**: the `verificationMessageTemplate` block is the one honest spelling (the provider keeps both in sync).
- **Per-kind tags**: identity tags derive from metadata; custom user tags are a platform-wide concern.

## Worth Knowing

- `passwordPolicy.temporaryPasswordValidityDays: 0` is NOT "no expiration" -- AWS treats a submitted 0 as null and applies its 7-day default. There is no way to make temporary passwords permanent.
- Group `precedence` 0 (AWS's strongest priority) is unsendable through the Terraform provider on either engine -- a recorded provider asymmetry; 1 is the strongest expressible value.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
