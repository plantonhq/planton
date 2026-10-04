# AwsCognitoIdentityProvider

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsCognitoIdentityProviderSpec defines the desired configuration for an
external identity provider federated into an Amazon Cognito User Pool.

This resource manages social (Google, Facebook, Amazon, Apple), enterprise
OIDC (Okta, Azure AD, Auth0), and SAML 2.0 (Azure AD, Salesforce, ADFS)
identity providers. Once configured, users can sign in to the User Pool
through the federated provider, and Cognito maps the provider's user
attributes to the pool's schema.

This is a child resource of AwsCognitoUserPool. The User Pool must exist
before creating identity providers. After creating an identity provider,
add its provider_name to the User Pool Client's
`supported_identity_providers` list to enable federated sign-in for that
client.

Key constraints:
- provider_name and provider_type are ForceNew: changing either destroys
  and recreates the identity provider.
- user_pool_id is ForceNew: the provider cannot be moved between pools.
- provider_name must be unique within a User Pool.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoIdentityProvider
metadata:
  name: test-google-idp
  org: test-org
  env: dev
  id: awscogidp-test-google-idp
spec:
  region: us-west-2
  userPoolId:
    value: us-west-2_TestPool123
  providerName: Google
  providerType: Google
  google:
    clientId: "123456789-test.apps.googleusercontent.com"
    clientSecret: "GOCSPX-test-secret-value"
    authorizeScopes: "email profile openid"
  attributeMapping:
    email: email
    username: sub
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.userPoolId` | `string \| valueFrom` | yes |  | AwsCognitoUserPool (`status.outputs.user_pool_id`) |
| `spec.providerName` | `string` | yes |  |  |
| `spec.providerType` | `string` | yes |  |  |
| `spec.google` | `AwsCognitoIdpGoogleConfig` |  |  |  |
| `spec.google.clientId` | `string` | yes |  |  |
| `spec.google.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.google.authorizeScopes` | `string` | yes |  |  |
| `spec.facebook` | `AwsCognitoIdpFacebookConfig` |  |  |  |
| `spec.facebook.clientId` | `string` | yes |  |  |
| `spec.facebook.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.facebook.authorizeScopes` | `string` | yes |  |  |
| `spec.facebook.apiVersion` | `string` |  |  |  |
| `spec.loginWithAmazon` | `AwsCognitoIdpLoginWithAmazonConfig` |  |  |  |
| `spec.loginWithAmazon.clientId` | `string` | yes |  |  |
| `spec.loginWithAmazon.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.loginWithAmazon.authorizeScopes` | `string` | yes |  |  |
| `spec.signInWithApple` | `AwsCognitoIdpSignInWithAppleConfig` |  |  |  |
| `spec.signInWithApple.clientId` | `string` | yes |  |  |
| `spec.signInWithApple.teamId` | `string` | yes |  |  |
| `spec.signInWithApple.keyId` | `string` | yes |  |  |
| `spec.signInWithApple.privateKey` | `string` (sensitive) | yes |  |  |
| `spec.signInWithApple.authorizeScopes` | `string` | yes |  |  |
| `spec.oidc` | `AwsCognitoIdpOidcConfig` |  |  |  |
| `spec.oidc.clientId` | `string` | yes |  |  |
| `spec.oidc.oidcIssuer` | `string` | yes |  |  |
| `spec.oidc.authorizeScopes` | `string` |  |  |  |
| `spec.oidc.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.oidc.attributesRequestMethod` | `string` |  |  |  |
| `spec.oidc.authorizeUrl` | `string` |  |  |  |
| `spec.oidc.tokenUrl` | `string` |  |  |  |
| `spec.oidc.attributesUrl` | `string` |  |  |  |
| `spec.oidc.jwksUri` | `string` |  |  |  |
| `spec.oidc.attributesUrlAddAttributes` | `bool` |  |  |  |
| `spec.saml` | `AwsCognitoIdpSamlConfig` |  |  |  |
| `spec.saml.metadataFile` | `string` |  |  |  |
| `spec.saml.metadataUrl` | `string` |  |  |  |
| `spec.saml.idpSignOut` | `bool` |  |  |  |
| `spec.saml.idpInit` | `bool` |  |  |  |
| `spec.saml.encryptedResponses` | `bool` |  |  |  |
| `spec.saml.requestSigningAlgorithm` | `string` |  |  |  |
| `spec.attributeMapping` | `map<string, string>` |  |  |  |
| `spec.idpIdentifiers` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.userPoolId

`string | valueFrom` · required

The ID of the Cognito User Pool to attach this identity provider to.
Format: "{region}_{poolId}" (e.g., "us-east-1_Ab1Cd2EfG").

This field is ForceNew: changing it requires replacing the identity provider.

- references: AwsCognitoUserPool (`status.outputs.user_pool_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_id}} -- a bare string does not parse

### spec.providerName

`string` · required

Display name for this identity provider within the User Pool. Must be
unique within the pool. Referenced by User Pool Clients in their
`supported_identity_providers` list.

For the SOCIAL provider types AWS requires the name to EQUAL the type
("Google", "Facebook", "LoginWithAmazon", "SignInWithApple") -- custom
names are an OIDC/SAML feature.

Examples: "Google", "CorpOkta", "AzureAD-SAML"

1-32 UTF-8 characters. This field is ForceNew.

- rule: {"string":{"minLen":"1","maxLen":"32"}}

### spec.providerType

`string` · required

The type of identity provider, exactly as the AWS API spells it. Valid
values: "Google", "Facebook", "LoginWithAmazon", "SignInWithApple",
"OIDC", "SAML". Determines which provider configuration field to populate
(google, facebook, login_with_amazon, sign_in_with_apple, oidc, or saml).

This field is ForceNew: changing it requires replacing the identity provider.

- rule: {"required":true}

### spec.google

`AwsCognitoIdpGoogleConfig`

Google OAuth 2.0 configuration. Set when provider_type is Google.

### spec.google.clientId

`string` · required

Google OAuth 2.0 client ID from the Google Cloud Console.

- rule: {"string":{"minLen":"1"}}

### spec.google.clientSecret

`string` · required · sensitive

Google OAuth 2.0 client secret from the Google Cloud Console.

- rule: {"string":{"minLen":"1"}}

### spec.google.authorizeScopes

`string` · required

Space-separated OAuth scopes (e.g., "email profile openid").

- rule: {"string":{"minLen":"1"}}

### spec.facebook

`AwsCognitoIdpFacebookConfig`

Facebook Login configuration. Set when provider_type is Facebook.

### spec.facebook.clientId

`string` · required

Facebook App ID from the Facebook Developer Portal.

- rule: {"string":{"minLen":"1"}}

### spec.facebook.clientSecret

`string` · required · sensitive

Facebook App Secret from the Facebook Developer Portal.

- rule: {"string":{"minLen":"1"}}

### spec.facebook.authorizeScopes

`string` · required

Comma-separated OAuth scopes (e.g., "email,public_profile").
Note: Facebook uses comma-separated scopes, unlike other providers
which use space-separated.

- rule: {"string":{"minLen":"1"}}

### spec.facebook.apiVersion

`string`

Facebook Graph API version (e.g., "v17.0"). When omitted, Cognito uses
the latest version it supports.

### spec.loginWithAmazon

`AwsCognitoIdpLoginWithAmazonConfig`

Login with Amazon configuration. Set when provider_type is LoginWithAmazon.

### spec.loginWithAmazon.clientId

`string` · required

Login with Amazon client ID from the Amazon Developer Console.

- rule: {"string":{"minLen":"1"}}

### spec.loginWithAmazon.clientSecret

`string` · required · sensitive

Login with Amazon client secret from the Amazon Developer Console.

- rule: {"string":{"minLen":"1"}}

### spec.loginWithAmazon.authorizeScopes

`string` · required

Space-separated OAuth scopes (e.g., "profile postal_code").

- rule: {"string":{"minLen":"1"}}

### spec.signInWithApple

`AwsCognitoIdpSignInWithAppleConfig`

Sign in with Apple configuration. Set when provider_type is SignInWithApple.

### spec.signInWithApple.clientId

`string` · required

Apple Services ID configured in the Apple Developer Portal.

- rule: {"string":{"minLen":"1"}}

### spec.signInWithApple.teamId

`string` · required

Apple Developer Team ID (10-character alphanumeric string).

- rule: {"string":{"minLen":"1"}}

### spec.signInWithApple.keyId

`string` · required

Key ID for the Apple private key.

- rule: {"string":{"minLen":"1"}}

### spec.signInWithApple.privateKey

`string` · required · sensitive

Apple private key in PEM format. Used to generate the client_secret JWT.

- rule: {"string":{"minLen":"1"}}

### spec.signInWithApple.authorizeScopes

`string` · required

Space-separated OAuth scopes (e.g., "email name").

- rule: {"string":{"minLen":"1"}}

### spec.oidc

`AwsCognitoIdpOidcConfig`

Generic OIDC configuration. Set when provider_type is OIDC.

### spec.oidc.clientId

`string` · required

OIDC client ID registered with the identity provider.

- rule: {"string":{"minLen":"1"}}

### spec.oidc.oidcIssuer

`string` · required

OIDC issuer URL (e.g., "https://login.microsoftonline.com/{tenant}/v2.0").
Cognito auto-discovers authorize, token, userinfo, and JWKS endpoints from
the issuer's .well-known/openid-configuration document.

- rule: {"string":{"minLen":"1"}}

### spec.oidc.authorizeScopes

`string`

Space-separated OAuth/OIDC scopes (e.g., "openid email profile").

### spec.oidc.clientSecret

`string` · sensitive

OIDC client secret. Optional for public clients that use PKCE.

### spec.oidc.attributesRequestMethod

`string`

HTTP method for the userinfo endpoint: "GET" or "POST".
When omitted, Cognito defaults to "GET".

### spec.oidc.authorizeUrl

`string`

Override the auto-discovered authorization endpoint URL.

### spec.oidc.tokenUrl

`string`

Override the auto-discovered token endpoint URL.

### spec.oidc.attributesUrl

`string`

Override the auto-discovered userinfo endpoint URL.

### spec.oidc.jwksUri

`string`

Override the auto-discovered JWKS endpoint URL.

### spec.oidc.attributesUrlAddAttributes

`bool`

When true, Cognito appends the requested attributes as query parameters
to the userinfo (attributes_url) request instead of relying on the
provider returning them by default. Only some providers need this.

### spec.saml

`AwsCognitoIdpSamlConfig`

SAML 2.0 configuration. Set when provider_type is SAML.

### spec.saml.metadataFile

`string`

SAML metadata XML content as a string. Use this for inline metadata.
Mutually exclusive with metadata_url.

### spec.saml.metadataUrl

`string`

URL pointing to the IdP's SAML metadata document. Cognito fetches the
metadata from this URL. Mutually exclusive with metadata_file.

### spec.saml.idpSignOut

`bool`

Enable single logout (SLO). When true, Cognito signs the user out of
the SAML IdP when they sign out of the User Pool.

### spec.saml.idpInit

`bool`

Enable IdP-initiated SSO. When true, the IdP can start the sign-in
flow directly without the user first visiting the Cognito hosted UI.

### spec.saml.encryptedResponses

`bool`

Require encrypted SAML assertions from the IdP.

### spec.saml.requestSigningAlgorithm

`string`

SAML request signing algorithm (e.g., "rsa-sha256"). When omitted,
Cognito uses the default algorithm.

### spec.attributeMapping

`map<string, string>`

Maps identity provider attributes to Cognito User Pool attributes.
Keys are Cognito user pool attribute names (e.g., "email", "username",
"given_name"). Values are provider-specific attribute names or paths
(e.g., "sub", "email").

When omitted, AWS applies default mappings based on the provider type.

### spec.idpIdentifiers

`[]string`

Alternative identifiers for this identity provider. These can be used
in the login endpoint's `idp_identifier` parameter to redirect to this
provider without exposing the provider_name.

Maximum 50 identifiers, each 1-40 characters.

- rule: {"repeated":{"maxItems":"50"}}

## Validation Rules

- `provider_type_valid`: provider_type must be 'Google', 'Facebook', 'LoginWithAmazon', 'SignInWithApple', 'OIDC', or 'SAML'
- `provider_type_config_match`: provider_type must match the provider configuration field: Google requires 'google', Facebook requires 'facebook', LoginWithAmazon requires 'loginWithAmazon', SignInWithApple requires 'signInWithApple', OIDC requires 'oidc', SAML requires 'saml'
- `saml_metadata_required`: SAML providers must set metadata_file or metadata_url
- `saml_metadata_exclusive`: SAML providers must set only one of metadata_file or metadata_url, not both
- `oidc_attributes_request_method_valid`: OIDC attributes_request_method must be 'GET' or 'POST'
- `social_provider_name_matches_type`: for social provider types the provider_name must equal the provider_type (e.g. type 'Google' requires name 'Google'); custom names are only supported for OIDC and SAML providers

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCognitoIdentityProvider, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.provider_name` | `string` | The name of the identity provider as registered in the User Pool. This is the value that must appear in a User Pool Client's `supported_identity_providers` list to enable federated sign-in through this provider. |
| `status.outputs.provider_type` | `string` | The type of the identity provider (e.g., "Google", "OIDC", "SAML"). Informational — useful for downstream tooling and display. |
| `status.outputs.user_pool_id` | `string` | The user pool this identity provider is attached to, resolved from the spec reference. Providers are keyed by (pool id, provider name) in AWS, and a consumer holding only this resource gets both halves of that key from its outputs. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.userPoolId` | AwsCognitoUserPool | `status.outputs.user_pool_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsCognitoUserPoolClient | `spec.supportedIdentityProviders` | `status.outputs.provider_name` |

## See Also

- [Overview](../README.md)
