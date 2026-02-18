# AWS Cognito User Pool Examples

## 1. Minimal Email Authentication

The simplest user pool: email-based sign-in with a single public app client.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: my-auth
  org: acme
  env: dev
  id: awscog-dev-001
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack: dev.AwsCognitoUserPool.my-auth
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

## 2. OAuth with Hosted UI

Email sign-in with OAuth Authorization Code flow, a Cognito-hosted domain for the sign-in UI, and callback URLs for a web application.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: webapp-auth
  org: acme
  env: staging
  id: awscog-stg-001
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack: staging.AwsCognitoUserPool.webapp-auth
spec:
  region: us-east-1
  usernameAttributes:
    - email
  passwordPolicy:
    minimumLength: 10
    requireLowercase: true
    requireUppercase: true
    requireNumbers: true
    requireSymbols: false
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
      defaultRedirectUri: https://staging.example.com/callback
      supportedIdentityProviders:
        - COGNITO
      explicitAuthFlows:
        - ALLOW_USER_SRP_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
      enableTokenRevocation: true
      preventUserExistenceErrors: ENABLED
  domain:
    domain: acme-staging-auth
```

## 3. Production Multi-Client Setup

Strong password policy, optional MFA, SES email, two clients (public SPA + confidential server), pre-sign-up Lambda trigger, and deletion protection.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: prod-auth
  org: acme
  env: prod
  id: awscog-prd-001
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack: prod.AwsCognitoUserPool.prod-auth
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
    fromEmailAddress: "Acme Auth <noreply@example.com>"
    replyToEmailAddress: support@example.com
  deletionProtection: true
  lambdaConfig:
    preSignUp:
      valueFrom:
        kind: AwsLambda
        name: pre-signup-validator
        fieldPath: status.outputs.function_arn
  customAttributes:
    - name: tenant_id
      attributeDataType: String
      mutable: true
      stringMaxLength: "64"
    - name: role
      attributeDataType: String
      mutable: true
      stringMaxLength: "32"
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
        - https://app.example.com/silent-renew
      logoutUrls:
        - https://app.example.com/logout
      defaultRedirectUri: https://app.example.com/callback
      supportedIdentityProviders:
        - COGNITO
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
      explicitAuthFlows:
        - ALLOW_USER_SRP_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
      accessTokenValidityMinutes: 30
      enableTokenRevocation: true
  domain:
    domain: acme-prod-auth
```

## 4. API Gateway JWT Authorizer Integration (valueFrom pattern)

Shows how an HTTP API Gateway references the Cognito User Pool endpoint and client ID for JWT authorization.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsHttpApiGateway
metadata:
  name: my-api
  org: acme
  env: prod
  id: awshttpapigw-prd-001
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack: prod.AwsHttpApiGateway.my-api
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

## 5. Custom Attributes for Multi-Tenant SaaS

Demonstrates custom attributes for a B2B SaaS user pool with tenant isolation.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: saas-auth
  org: acme
  env: prod
  id: awscog-prd-002
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack: prod.AwsCognitoUserPool.saas-auth
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
  allowAdminCreateUserOnly: true
  customAttributes:
    - name: tenant_id
      attributeDataType: String
      mutable: false
      required: true
      stringMinLength: "1"
      stringMaxLength: "64"
    - name: tenant_role
      attributeDataType: String
      mutable: true
      stringMaxLength: "32"
    - name: employee_id
      attributeDataType: Number
      mutable: false
      numberMinValue: "1"
      numberMaxValue: "9999999"
  lambdaConfig:
    preSignUp:
      value: "arn:aws:lambda:us-east-1:123456789012:function:validate-tenant"
    preTokenGeneration:
      value: "arn:aws:lambda:us-east-1:123456789012:function:add-tenant-claims"
  clients:
    - name: admin-portal
      generateSecret: true
      allowedOauthFlowsUserPoolClient: true
      allowedOauthFlows:
        - code
      allowedOauthScopes:
        - openid
        - email
        - profile
      callbackUrls:
        - https://admin.example.com/callback
      explicitAuthFlows:
        - ALLOW_USER_SRP_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
        - ALLOW_ADMIN_USER_PASSWORD_AUTH
      enableTokenRevocation: true
  domain:
    domain: acme-saas-auth
```

## 6. Custom Domain with ACM Certificate

Uses a custom domain (auth.example.com) backed by an ACM certificate for branded login URLs.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: branded-auth
  org: acme
  env: prod
  id: awscog-prd-003
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack: prod.AwsCognitoUserPool.branded-auth
spec:
  region: us-east-1
  usernameAttributes:
    - email
  autoVerifiedAttributes:
    - email
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
        - https://app.example.com/callback
      explicitAuthFlows:
        - ALLOW_USER_SRP_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
  domain:
    domain: auth.example.com
    certificateArn:
      valueFrom:
        kind: AwsCertManagerCert
        name: auth-cert
        fieldPath: status.outputs.certificate_arn
```

## 7. Alias-Based Identity Model

Uses alias attributes instead of username attributes, allowing users to have a separate username plus email/preferred_username as aliases.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoUserPool
metadata:
  name: alias-auth
  org: acme
  env: dev
  id: awscog-dev-002
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack: dev.AwsCognitoUserPool.alias-auth
spec:
  region: us-east-1
  aliasAttributes:
    - email
    - preferred_username
  autoVerifiedAttributes:
    - email
  clients:
    - name: mobile-app
      explicitAuthFlows:
        - ALLOW_USER_SRP_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
        - ALLOW_CUSTOM_AUTH
```
