# AWS Cognito Identity Provider

Deploys a federated identity provider into an existing Cognito User Pool, enabling sign-in through social providers (Google, Facebook, Amazon, Apple), enterprise OIDC (Okta, Azure AD, Auth0), or SAML 2.0 (ADFS, Salesforce). Attribute mapping translates provider claims into pool schema attributes, and the provider's registered name is what a User Pool Client lists in `supportedIdentityProviders` to offer the sign-in option. Provider name and type are ForceNew — settle them before the first production deployment.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cognito Identity Provider** -- a federation registration attached to the specified User Pool, configured with provider-specific details (OAuth client credentials, OIDC issuer, or SAML metadata) and optional attribute mapping from provider claims to pool schema attributes

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A Cognito User Pool** -- the identity provider attaches to an existing pool. Provide the `userPoolId` directly or reference an AwsCognitoUserPool Infra Component via ValueFromRef.
- **Provider credentials** -- OAuth client ID and secret (for Google, Facebook, Amazon, OIDC), Apple private key and team ID (for Sign In with Apple), or SAML metadata URL/file (for SAML 2.0). Obtain these from the external identity provider's developer console.

## Deploy

### Console

Open the deployment store, find **AWS Cognito Identity Provider**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Google OAuth** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoIdentityProvider
metadata:
  name: google-login
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  userPoolId:
    value: "us-west-2_Ab1Cd2EfG"
  providerName: Google
  providerType: Google
  google:
    clientId: "123456789.apps.googleusercontent.com"
    clientSecret: "$secret/google-oauth-client-secret"
    authorizeScopes: "email profile openid"
  attributeMapping:
    email: email
    username: sub
```

```shell
planton apply -f cognito-idp.yaml
```

This creates a Google identity provider attached to the specified User Pool with email and username attribute mapping. The client secret is a managed-secret reference (`$secret/<slug>`) — the platform rejects plaintext secrets and resolves the reference just-in-time at deploy. After creation, add `"Google"` to the `supportedIdentityProviders` list on your User Pool Client to enable federated sign-in. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the identity provider to a User Pool deployed in the same InfraPipeline:

```yaml
spec:
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: app-auth
      fieldPath: status.outputs.user_pool_id
```

The InfraPipeline resolves the dependency graph, deploys the User Pool first, then provisions the identity provider with the resolved pool ID.

## Key Configuration

These are the most important decisions when configuring a Cognito Identity Provider. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Provider type** -- Determines the authentication protocol and which configuration block to populate: `google`, `facebook`, `loginWithAmazon`, `signInWithApple` for social login; `oidc` for enterprise SSO with Okta, Azure AD, or any OIDC-compliant provider; `saml` for SAML 2.0 federation. The `providerType` field is ForceNew -- changing it requires replacing the identity provider.

**Provider name** -- A unique identifier within the User Pool (e.g., "Google", "CorpOkta", "AzureAD-SAML"). This name appears in the User Pool Client's `supportedIdentityProviders` list and in the hosted UI provider selection. ForceNew -- cannot be changed after creation.

**Attribute mapping** -- Maps provider-specific attribute names to Cognito pool attributes (e.g., Google's `sub` to `username`, SAML claim URIs to `email`). When omitted, AWS applies default mappings based on the provider type. Custom mappings are essential when your provider uses non-standard claim names.

**SAML metadata source** -- For SAML providers, choose between `metadataUrl` (Cognito fetches metadata from the IdP) and `metadataFile` (inline XML metadata for air-gapped environments). Exactly one must be set.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCognitoUserPool** | `userPoolId` | `status.outputs.user_pool_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `provider_name` | Identity provider name as registered in the User Pool | User Pool Client `supportedIdentityProviders` list |
| `provider_type` | Identity provider type (Google, OIDC, SAML, etc.) | Downstream tooling and display |
| `user_pool_id` | The pool this provider is attached to | Correlating the provider with its pool in composed environments |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Google social login** -- Google OAuth 2.0 with email, profile, and openid scopes. The quickest path to social sign-in for consumer-facing applications. Start from the **Google OAuth** preset.

**Enterprise OIDC** -- Generic OIDC configuration for Okta, Azure AD, Auth0, Keycloak, or any OIDC-compliant identity provider. Cognito auto-discovers endpoints from the issuer URL. Start from the **Enterprise OIDC** preset.

**SAML 2.0 federation** -- SAML federation with metadata URL, single logout enabled, and standard claim URI attribute mapping. Works with Azure AD, ADFS, Salesforce, and other SAML 2.0 IdPs. Start from the **SAML Federation** preset.

## Works With

- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) -- provides the parent user pool that this identity provider attaches to
- [**AWS Cognito User Pool Client**](/infra-catalog/aws-cognito-user-pool-client) -- lists this provider's name in `supportedIdentityProviders` to offer the sign-in option to an application