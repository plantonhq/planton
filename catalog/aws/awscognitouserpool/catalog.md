# AWS Cognito User Pool

Deploys a Cognito User Pool — the user directory at the root of the Cognito family — with the identity model, password and passwordless sign-in factors, MFA and recovery, verification and invitation messaging, email/SMS delivery, custom schema attributes, Lambda lifecycle triggers, threat protection, log exports, and an optional hosted UI domain. App clients, federated identity providers, and OAuth resource servers are their own first-class kinds (AwsCognitoUserPoolClient, AwsCognitoIdentityProvider, AwsCognitoResourceServer) that reference the pool.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cognito User Pool** -- a user directory with the configured identity model (email/phone or alias sign-in), feature tier, password policy, passwordless first factors (email OTP, SMS OTP, passkeys), MFA settings, account recovery order, verification/invitation message templates, device tracking, and threat protection
- **User Pool Domain** -- created only when `domain` is configured; a Cognito-hosted prefix domain or a custom domain backed by an ACM certificate for the hosted sign-in UI and OAuth endpoints
- **Custom Schema Attributes** -- created only when `customAttributes` entries are provided; extends the pool's user schema beyond the standard set (append-only after creation)
- **User Groups** -- created only when `userGroups` entries are provided; one group per entry with its description, precedence, and optional IAM role (membership surfaces in the `cognito:groups` token claim; assigning users to groups is a runtime operation, not configuration)
- **Risk Configuration** -- created only when `riskConfiguration` is set (requires threat protection AUDIT/ENFORCED); the pool-wide automated-response policy for account-takeover and compromised-credential events, notification email templates, and IP allow/block exceptions
- **Log Export Configurations** -- created only when `logConfigurations` entries are provided; streams notification errors or auth events to CloudWatch Logs, Kinesis Firehose, or S3
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A verified SES identity** (optional) -- required only when using DEVELOPER email mode for production-volume sending. Provide the SES identity ARN directly or as a ValueFromRef.
- **An IAM role for SNS** (optional) -- required only when SMS delivery is configured (SMS OTP sign-in, SMS MFA, or phone auto-verification). The role's trust policy must allow `cognito-idp.amazonaws.com` with your external ID; reference an AwsIamRole Infra Component or provide the ARN directly.
- **Lambda functions** (optional) -- required when configuring authentication lifecycle triggers (pre-sign-up, post-confirmation, custom auth challenges, versioned token customization, custom senders). Each Lambda must grant Cognito permission to invoke it. Reference AwsLambda Infra Components via ValueFromRef or provide function ARNs directly.
- **A KMS key** (optional) -- required only when custom email/SMS sender Lambdas are configured; Cognito encrypts verification codes with it before invoking your functions.
- **An ACM certificate in us-east-1** (optional) -- required only when using a custom domain (FQDN) for the hosted UI. Cognito routes custom domains through CloudFront, so the certificate must be in us-east-1 regardless of the pool's region. Reference an AwsCertManagerCert Infra Component or provide the ARN directly.

## Deploy

### Console

Open the deployment store, find **AWS Cognito User Pool**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Email Auth Basic** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoUserPool
metadata:
  name: app-auth
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  usernameAttributes:
    - email
  autoVerifiedAttributes:
    - email
  accountRecoveryMechanisms:
    - name: verified_email
      priority: 1
  deletionProtection: true
```

```shell
planton apply -f cognito-user-pool.yaml
```

This creates a user pool with email as the sign-in identifier, email auto-verification, email-based recovery, and deletion protection. Add an AwsCognitoUserPoolClient resource so an application can authenticate — the pool alone has no app client. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire Lambda triggers and custom domain certificates:

```yaml
spec:
  lambdaConfig:
    preSignUp:
      valueFrom:
        kind: AwsLambda
        name: signup-validator
        fieldPath: status.outputs.function_arn
  domain:
    domain: auth.example.com
    certificateArn:
      valueFrom:
        kind: AwsCertManagerCert
        name: auth-cert
        fieldPath: status.outputs.cert_arn
```

The InfraPipeline resolves the dependency graph, deploys the Lambda function and ACM certificate first, then provisions the user pool with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Cognito User Pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Identity model** -- Choose between `usernameAttributes` (email/phone as the sign-in identifier) and `aliasAttributes` (separate username with email/phone/preferred_username as aliases). These fields and `usernameCaseSensitive` are ForceNew in AWS -- changing them destroys and recreates the pool with all users. Choose carefully before your first production deployment.

**Feature tier** -- `userPoolTier` selects the pricing plan: `LITE` (password-only basics), `ESSENTIALS` (the AWS default: passwordless factors, passkeys, managed login branding), or `PLUS` (adds threat protection). Several capabilities are tier-gated: passwordless first factors and email MFA need Essentials+; `userPoolAddOns` threat protection, `riskConfiguration`, and auth-event log exports need Plus.

**Threat protection responses** -- `userPoolAddOns.advancedSecurityMode` turns risk evaluation on (AUDIT observes, ENFORCED acts); `riskConfiguration` is the policy it acts on: per-risk-level account-takeover responses with user-notification emails, compromised-credential blocking, and IP allow/block exceptions. The pool-wide policy covers every app client; a client that needs its own posture overrides it with `riskConfiguration` on the AwsCognitoUserPoolClient spec.

**User groups** -- `userGroups` declares the pool's groups (name, precedence, optional IAM role for identity-pool federation). Membership lands in the `cognito:groups` token claim; assigning users to groups is a runtime admin operation, never configuration.

**Sign-in factors** -- `allowedFirstAuthFactors` opts users into passwordless sign-in: `EMAIL_OTP`, `SMS_OTP` (requires `smsConfiguration`), and `WEB_AUTHN` passkeys (configure the relying-party domain in `webAuthn`). An empty list keeps the AWS default of password-only.

**MFA enforcement** -- Set `mfaConfiguration` to `OFF` (default), `OPTIONAL` (users opt in), or `ON` (required for all users). With MFA active, enable `softwareTokenMfaEnabled` for authenticator apps, configure `emailMfa` for email codes, and customize `smsAuthenticationMessage` for SMS codes.

**Message delivery** -- Email defaults to Cognito's built-in service (limited to ~50 emails/day). Set `emailConfiguration.emailSendingAccount` to `DEVELOPER` with a verified SES `sourceArn` for production volumes. SMS always requires `smsConfiguration` with the SNS caller role and external ID.

**Deletion protection** -- Set `deletionProtection: true` for production pools to prevent accidental deletion. The pool is a data store: deleting it deletes every user account.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** (optional) | `smsConfiguration.snsCallerArn`, `userGroups[].roleArn` | `status.outputs.role_arn` |
| **AwsLambda** (optional) | `lambdaConfig.preSignUp`, `preAuthentication`, `postAuthentication`, `postConfirmation`, `preTokenGeneration`, `preTokenGenerationConfig.lambdaArn`, `customMessage`, `userMigration`, `defineAuthChallenge`, `createAuthChallenge`, `verifyAuthChallengeResponse`, `customEmailSender.lambdaArn`, `customSmsSender.lambdaArn` | `status.outputs.function_arn` |
| **AwsKmsKey** (optional) | `lambdaConfig.kmsKeyId` | `status.outputs.key_arn` |
| **AwsSesEmailIdentity** (optional) | `emailConfiguration.sourceArn`, `riskConfiguration.accountTakeover.notifyConfiguration.sourceArn` | `status.outputs.identity_arn` |
| **AwsSesConfigurationSet** (optional) | `emailConfiguration.configurationSet` | `status.outputs.configuration_set_name` |
| **AwsCloudwatchLogGroup** (optional) | `logConfigurations[].cloudwatchLogGroupArn` | `status.outputs.log_group_arn` |
| **AwsKinesisFirehose** (optional) | `logConfigurations[].firehoseStreamArn` | `status.outputs.delivery_stream_arn` |
| **AwsS3Bucket** (optional) | `logConfigurations[].s3BucketArn` | `status.outputs.bucket_arn` |
| **AwsCertManagerCert** (optional) | `domain.certificateArn` | `status.outputs.cert_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `user_pool_id` | User pool identifier ({region}_{poolId}) | AwsCognitoUserPoolClient, AwsCognitoIdentityProvider, and AwsCognitoResourceServer all bind it |
| `user_pool_arn` | Amazon Resource Name of the user pool | IAM policies, API Gateway authorizers, cross-service permissions |
| `user_pool_endpoint` | The pool's API endpoint | SDK configuration |
| `issuer` | OIDC issuer URL | JWT validation, OIDC client configuration |
| `user_pool_domain` | The configured hosted UI domain | Application redirect configuration, OAuth endpoint discovery |
| `hosted_ui_url` | Full URL of the hosted sign-in page | Application login links |
| `cloudfront_distribution` | CloudFront target for custom domains | Route53 alias record target |
| `cloudfront_distribution_arn` | CloudFront distribution ARN for custom domains | IAM/WAF policies on the distribution |
| `cloudfront_hosted_zone_id` | CloudFront hosted zone for alias records | Route53 alias record zone ID |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Email authentication** -- Email as sign-in identifier with auto-verification and email recovery. The standard starting point for development and simple applications. Start from the **Email Auth Basic** preset.

**OAuth with hosted UI** -- A Cognito-hosted domain for the sign-in pages plus an AwsCognitoUserPoolClient with the Authorization Code flow. Suitable for web applications needing OAuth/OIDC. Start from the **OAuth with Hosted UI** preset.

**Production hardened** -- Strong password policy with history, optional MFA with TOTP, SES-based email, verify-before-update, threat protection on the Plus tier, and deletion protection. Start from the **Production Hardened** preset.

**Threat protected** -- The hardened baseline plus the full automated-response policy: block high-risk sign-ins with user notification, MFA-challenge medium risk, block compromised credentials, skip evaluation for trusted CIDRs, and user groups with an admin role. Start from the **Threat Protected** preset.

## Works With

- [**AWS Cognito User Pool Client**](/infra-catalog/aws-cognito-user-pool-client) -- the app client that applications authenticate through; binds `user_pool_id`
- [**AWS Cognito Identity Provider**](/infra-catalog/aws-cognito-identity-provider) -- federates Google/Facebook/Apple/OIDC/SAML sign-in into this pool
- [**AWS Cognito Resource Server**](/infra-catalog/aws-cognito-resource-server) -- defines custom OAuth scopes for APIs protected by this pool
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the SNS caller role SMS delivery rides on
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- authentication lifecycle triggers, token customization, and custom message senders
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- encrypts codes handed to custom sender Lambdas
- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) -- the us-east-1 certificate for custom hosted UI domains
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group), [**AWS Kinesis Firehose**](/infra-catalog/aws-kinesis-firehose), [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- log export destinations
