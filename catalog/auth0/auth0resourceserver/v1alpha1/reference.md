# Auth0ResourceServer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0ResourceServerSpec defines an API in the Auth0 tenant the provider
connection's credential belongs to (Auth0 calls it a resource server): the
audience applications ask for tokens to, the scopes (permissions) those tokens
can carry, how the tokens are signed, shaped and protected, which applications
may get one at all, and the scopes every third-party application gets by
default.

Who can get a token for this API is decided in two places:
- subject_type_authorization is the API's access policy, one per kind of
  subject: user-delegated access (an application acting for a signed-in
  person) and client access (an application acting for itself, through the
  client-credentials flow).
- Client grants are the per-application ceilings the policies consult. A
  grant for one application lives on that application (Auth0Client's
  api_grants); the grants every third-party application gets without one of
  its own are declared here, in third_party_client_default_grants.

Unset means unmanaged: a field or block the spec leaves out is never sent, and the API keeps whatever
Auth0 holds -- Auth0's default on a new API, the live value on an API adopted
into this kind. A few settings have no value of their own in the provider
(verification_location, token_lifetime_for_anonymous_access_tokens,
access_token, and proof_of_possession.required_for once its block is
declared): on an adopted API, declare their live values, because leaving
them unset resets them. The comment on each says so.

Plans: defining APIs, scopes, access policies and default grants is on every
plan. Token encryption, mTLS proof of possession and transactional
authorization need the Enterprise plan with the Highly Regulated Identity
add-on; anonymous-session settings are an Enterprise Early Access feature;
Online Refresh Tokens are in Beta. Each field says which.

The credential needs create:resource_servers, read:resource_servers,
update:resource_servers and delete:resource_servers, and -- when default
grants are declared -- create:client_grants, read:client_grants,
update:client_grants and delete:client_grants on the tenant's Management API
(iac/permissions.yaml).

https://auth0.com/docs/get-started/apis
https://auth0.com/docs/get-started/apis/api-access-policies-for-applications
https://auth0.com/docs/get-started/applications/application-access-to-apis-client-grants
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/resource_server
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/client_grant

## Example

```yaml
# Auth0 Resource Server Test Manifest
# The component's full surface, for the offline plan and preview proofs: it
# declares settings that need the Enterprise plan with the Highly Regulated
# Identity add-on (token encryption, consent_policy) and Early Access
# features (anonymous sessions, Online Refresh Tokens, authorization
# policies), so a live apply needs a tenant that has them. The live lanes
# run e2e/scenarios/.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: Your Auth0 tenant domain (e.g., your-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - create:resource_servers, read:resource_servers,
#      update:resource_servers, delete:resource_servers
#    - create:client_grants, read:client_grants,
#      update:client_grants, delete:client_grants

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ResourceServer
metadata:
  name: test-api
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # Required: API identifier (audience)
  identifier: https://api.test.planton.dev/

  # Optional: Friendly display name
  name: Test API

  # Token configuration
  signingAlg: RS256
  tokenLifetime: 86400         # 24 hours
  tokenLifetimeForWeb: 7200    # 2 hours
  tokenLifetimeForAnonymousAccessTokens: 86400

  # Access control
  allowOfflineAccess: true
  allowOnlineAccess: false
  allowOnlineAccessWithEphemeralSessions: false
  skipConsentForVerifiableFirstPartyClients: true
  enforcePolicies: true
  tokenDialect: access_token_authz
  consentPolicy: transactional-authorization-with-mfa

  # API Scopes/Permissions
  scopes:
    - name: read:items
      description: Read access to items
    - name: write:items
      description: Create and update items
    - name: delete:items
      description: Delete items
    - name: admin:all
      description: Full administrative access

  # Anonymous-session claims
  accessToken:
    claimsMapping:
      customClaims:
        - name: country
          expression: anonymous_session.metadata.country

  # Rich Authorization Request types
  authorizationDetails:
    - type: payment
    - type: money_transfer

  authorizationPolicy:
    policyId: pol_e2e_policy

  # Sender-constrained tokens (DPoP), required for public applications
  proofOfPossession:
    mechanism: dpop
    required: true
    requiredFor: public_clients

  # The access policy
  subjectTypeAuthorization:
    user:
      policy: require_client_grant
    client:
      policy: require_client_grant
    anonymousUser:
      policy: deny_all

  # Encrypted tokens, to the API's public key
  tokenEncryption:
    format: compact-nested-jwe
    encryptionKey:
      name: e2e-encryption-key
      algorithm: RSA-OAEP-256
      pem: |
        -----BEGIN PUBLIC KEY-----
        MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4DWiwKitrj8kcow8SnCl
        UaHCLbmvHNh82rL4Et597I83qObJOVU8WhiZL5NnOVGoLxVE4db49xZpjQZMqhre
        kEVsnKdRA858/XV9etkP6mbB5KP/ALB292iKHzQnF/cMtjctmhI310lJFWkBa7nW
        pz6V4TFbWJ7lATodIZJoHAV8fzuoaz07e8nzDYvnRYfbroi5gD724lBrDWzDxv0W
        BYc+KS0rNglZU9AsrHFDicaoPDsxsep3yd+nWefMf7ww+ZFpYi7pTJyxxMFVPuQV
        rAY2kGbZqF8EFRdW+pUs3aLUGlKrOtkuGuv96+h3VqGZ8TjnGzjwZm4BlaoWutPP
        NwIDAQAB
        -----END PUBLIC KEY-----

  # Default grants every third-party application gets
  thirdPartyClientDefaultGrants:
    - subjectType: user
      scopes:
        - read:items
        - write:items
      authorizationDetailsTypes:
        - payment
    - subjectType: client
      allowAllScopes: true
      organizationUsage: allow
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.identifier` | `string` | yes |  |  |
| `spec.name` | `string` |  |  |  |
| `spec.signingAlg` | `string` |  |  |  |
| `spec.allowOfflineAccess` | `bool` |  |  |  |
| `spec.tokenLifetime` | `int32` |  |  |  |
| `spec.tokenLifetimeForWeb` | `int32` |  |  |  |
| `spec.skipConsentForVerifiableFirstPartyClients` | `bool` |  |  |  |
| `spec.enforcePolicies` | `bool` |  |  |  |
| `spec.tokenDialect` | `string` |  |  |  |
| `spec.scopes` | `[]Auth0ResourceServerScope` |  |  |  |
| `spec.scopes[].name` | `string` | yes |  |  |
| `spec.scopes[].description` | `string` |  |  |  |
| `spec.allowOnlineAccess` | `bool` |  |  |  |
| `spec.allowOnlineAccessWithEphemeralSessions` | `bool` |  |  |  |
| `spec.consentPolicy` | `string` |  |  |  |
| `spec.tokenLifetimeForAnonymousAccessTokens` | `int32` |  |  |  |
| `spec.verificationLocation` | `string` |  |  |  |
| `spec.signingSecret` | `string` (sensitive) | yes |  |  |
| `spec.accessToken` | `Auth0ResourceServerAccessToken` |  |  |  |
| `spec.accessToken.claimsMapping` | `Auth0ResourceServerClaimsMapping` |  |  |  |
| `spec.accessToken.claimsMapping.customClaims` | `[]Auth0ResourceServerCustomClaim` |  |  |  |
| `spec.accessToken.claimsMapping.customClaims[].name` | `string` | yes |  |  |
| `spec.accessToken.claimsMapping.customClaims[].expression` | `string` | yes |  |  |
| `spec.authorizationDetails` | `[]Auth0ResourceServerAuthorizationDetail` |  |  |  |
| `spec.authorizationDetails[].type` | `string` |  |  |  |
| `spec.authorizationDetails[].disable` | `bool` |  |  |  |
| `spec.authorizationPolicy` | `Auth0ResourceServerAuthorizationPolicy` |  |  |  |
| `spec.authorizationPolicy.policyId` | `string` |  |  |  |
| `spec.proofOfPossession` | `Auth0ResourceServerProofOfPossession` |  |  |  |
| `spec.proofOfPossession.disable` | `bool` |  |  |  |
| `spec.proofOfPossession.mechanism` | `string` |  |  |  |
| `spec.proofOfPossession.required` | `bool` |  |  |  |
| `spec.proofOfPossession.requiredFor` | `string` |  |  |  |
| `spec.subjectTypeAuthorization` | `Auth0ResourceServerSubjectTypeAuthorization` |  |  |  |
| `spec.subjectTypeAuthorization.user` | `Auth0ResourceServerUserAuthorization` |  |  |  |
| `spec.subjectTypeAuthorization.user.policy` | `string` |  |  |  |
| `spec.subjectTypeAuthorization.client` | `Auth0ResourceServerClientAuthorization` |  |  |  |
| `spec.subjectTypeAuthorization.client.policy` | `string` |  |  |  |
| `spec.subjectTypeAuthorization.anonymousUser` | `Auth0ResourceServerAnonymousUserAuthorization` |  |  |  |
| `spec.subjectTypeAuthorization.anonymousUser.policy` | `string` |  |  |  |
| `spec.tokenEncryption` | `Auth0ResourceServerTokenEncryption` |  |  |  |
| `spec.tokenEncryption.disable` | `bool` |  |  |  |
| `spec.tokenEncryption.format` | `string` |  |  |  |
| `spec.tokenEncryption.encryptionKey` | `Auth0ResourceServerTokenEncryptionKey` |  |  |  |
| `spec.tokenEncryption.encryptionKey.algorithm` | `string` | yes |  |  |
| `spec.tokenEncryption.encryptionKey.pem` | `string` | yes |  |  |
| `spec.tokenEncryption.encryptionKey.kid` | `string` |  |  |  |
| `spec.tokenEncryption.encryptionKey.name` | `string` |  |  |  |
| `spec.thirdPartyClientDefaultGrants` | `[]Auth0ResourceServerThirdPartyClientDefaultGrant` |  |  |  |
| `spec.thirdPartyClientDefaultGrants[].subjectType` | `string` | yes |  |  |
| `spec.thirdPartyClientDefaultGrants[].scopes` | `[]string` |  |  |  |
| `spec.thirdPartyClientDefaultGrants[].authorizationDetailsTypes` | `[]string` |  |  |  |
| `spec.thirdPartyClientDefaultGrants[].allowAllScopes` | `bool` |  |  |  |
| `spec.thirdPartyClientDefaultGrants[].organizationUsage` | `string` |  |  |  |
| `spec.thirdPartyClientDefaultGrants[].allowAnyOrganization` | `bool` |  |  |  |

## Field Details

### spec.identifier

`string` · required

identifier is the unique identifier for the resource server.
This value is used as the "audience" parameter for authorization calls.
Typically a URI representing your API (e.g., "https://api.example.com/").
Cannot be changed once set: changing it replaces the API.

Example: "https://api.mycompany.com/", "api.planton.live"

- rule: {"required":true}

### spec.name

`string`

name is a friendly display name for the resource server.
This is shown in the Auth0 dashboard and consent screens.
Cannot include `<` or `>` characters. Unset, the module uses metadata.name.

### spec.signingAlg

`string`

signing_alg is the algorithm used to sign access tokens for this API.
Options:
- "RS256": RSA using SHA-256 (asymmetric, recommended): Auth0 holds the
  private key, and the API verifies tokens against the tenant's public
  keys (JWKS)
- "HS256": HMAC using SHA-256 (symmetric): the API verifies tokens with a
  shared secret (signing_secret) that can also mint them
- "PS256": RSA-PSS using SHA-256, which Auth0 offers through an add-on
Set it explicitly: the Management API reference names HS256 as the default
for an API created without one, while the dashboard preselects RS256.

- rule: {"string":{"in":["","RS256","HS256","PS256"]}}

### spec.allowOfflineAccess

`bool` · optional (explicit presence)

allow_offline_access indicates whether refresh tokens can be issued for this API.
When true, applications can request refresh tokens using the "offline_access" scope.
This allows applications to obtain new access tokens without user interaction.
Unset: unmanaged -- Auth0's default (false) on a new API, the live value
on an adopted one.

### spec.tokenLifetime

`int32`

token_lifetime is the duration (in seconds) that access tokens remain valid
when issued from the token endpoint.
Range: 0 to 2592000 (30 days)
Default: 86400 (24 hours)

- rule: {"int32":{"lte":2592000,"gte":0}}

### spec.tokenLifetimeForWeb

`int32`

token_lifetime_for_web is the duration (in seconds) that access tokens remain valid
when issued via implicit or hybrid flows.
This should typically be shorter than token_lifetime for security.
Cannot be greater than token_lifetime.
Range: 0 to 2592000 (30 days)
Default: 7200 (2 hours)

- rule: {"int32":{"lte":2592000,"gte":0}}

### spec.skipConsentForVerifiableFirstPartyClients

`bool` · optional (explicit presence)

skip_consent_for_verifiable_first_party_clients indicates whether to skip
the consent prompt for applications flagged as first-party.
When true, first-party applications don't show the consent screen to users.
Third-party applications always show it.
Unset: unmanaged -- Auth0's default (true) on a new API, the live value
on an adopted one.

### spec.enforcePolicies

`bool` · optional (explicit presence)

enforce_policies enables RBAC authorization policies for this API.
When true, role and permission assignments are evaluated during login.
This allows you to use Auth0's built-in RBAC to control API access.
Requires token_dialect to be set to a value that includes permissions.
Unset: unmanaged -- Auth0's default (false) on a new API, the live value
on an adopted one.

https://auth0.com/docs/manage-users/access-control/rbac

### spec.tokenDialect

`string`

token_dialect determines the format of access tokens issued for this API.
Options:
- "access_token": Standard Auth0 JWT with claims
- "access_token_authz": Standard Auth0 JWT including RBAC permissions claims
- "rfc9068_profile": IETF JWT Access Token Profile compliant
- "rfc9068_profile_authz": IETF profile with RBAC permissions claims

Use "_authz" variants when enforce_policies is true to include permissions in tokens.
Default: access_token

https://auth0.com/docs/secure/tokens/access-tokens/access-token-profiles

- rule: {"string":{"in":["","access_token","access_token_authz","rfc9068_profile","rfc9068_profile_authz"]}}

### spec.scopes

`[]Auth0ResourceServerScope`

scopes defines the permissions that can be granted for this API.
Each scope represents a specific permission that applications can request.
Applications request scopes during authorization, and granted scopes
appear in the access token's "scope" claim.

The list is authoritative once it has entries: each apply makes it the
API's complete scope list, and a scope removed here is removed from the API
(along with its place in every role and grant).

Example scopes:
- read:users (permission to read user data)
- write:orders (permission to create/update orders)
- delete:products (permission to delete products)

https://auth0.com/docs/get-started/apis/api-settings#scopes

### spec.scopes[].name

`string` · required

name is the scope identifier used in OAuth flows.
Should follow the pattern: action:resource (e.g., "read:users", "write:orders")
This is what applications request and what appears in access tokens.

- rule: {"required":true}

### spec.scopes[].description

`string`

description is a human-readable explanation of what this scope grants.
Shown on consent screens and in the Auth0 dashboard.
Example: "Read access to user profiles"

### spec.allowOnlineAccess

`bool` · optional (explicit presence)

allow_online_access lets applications ask for Online Refresh Tokens for
this API (the "online_access" scope): refresh tokens bound to the person's
Auth0 session, which stop working when the session ends and extend it each
time they are used. They suit single-page applications whose browsers block
the cookies silent authentication relies on; public clients must use DPoP
with them. Unset, the API keeps what Auth0 holds (off on a new API).
Online Refresh Tokens are in Beta: Auth0 enables them on request.

https://auth0.com/docs/secure/tokens/refresh-tokens/online-refresh-tokens

### spec.allowOnlineAccessWithEphemeralSessions

`bool` · optional (explicit presence)

allow_online_access_with_ephemeral_sessions lets Online Refresh Tokens be
issued even when the tenant's sessions are ephemeral (they end when the
browser closes). Unset, the API keeps what Auth0 holds. Part of the Online
Refresh Tokens Beta.

### spec.consentPolicy

`string` · optional (explicit presence)

consent_policy is how Auth0 handles consent for rich authorization requests
to this API:
- "transactional-authorization-with-mfa": transactional authorization --
  Auth0 shows no consent prompt of its own when a push notification is
  sent; your own interface shows the authorization_details, and a
  post-login Action receives the request's linking id to step the person
  up with MFA. Not supported with Client-Initiated Backchannel
  Authentication. Needs the Enterprise plan with the Highly Regulated
  Identity add-on.
- "null": Auth0's standard consent behavior (a customized consent prompt,
  or the Guardian app showing the authorization_details).
Unset, the API keeps what Auth0 holds (standard on a new API).

https://auth0.com/docs/get-started/apis/configure-rich-authorization-requests

- rule: {"string":{"in":["transactional-authorization-with-mfa","null"]}}

### spec.tokenLifetimeForAnonymousAccessTokens

`int32` · optional (explicit presence)

token_lifetime_for_anonymous_access_tokens is how long, in seconds, an
access token this API issues in an anonymous session stays valid (a
session Auth0 opens before anyone signs in): 86400 (one day) to 2592000
(30 days). It has no value of its own in the provider: declare the live
value when adopting an existing API; leaving it unset on an adopted API
resets it. Anonymous sessions are an Enterprise Early Access feature.

https://auth0.com/docs/manage-users/sessions/anonymous-sessions/configure-anonymous-sessions

- rule: {"int32":{"lte":2592000,"gte":86400}}

### spec.verificationLocation

`string` · optional (explicit presence)

verification_location is the URL Auth0 retrieves this API's public keys
(a JWKS) from, to verify JWTs the API sends to Auth0 for token
introspection. Most APIs leave it unset. It has no value of its own in the
provider: declare the live value when adopting an existing API; leaving it
unset on an adopted API resets it.

### spec.signingSecret

`string` · required · optional (explicit presence) · sensitive

signing_secret is the shared secret an HS256 API's tokens are signed with,
at least 16 characters. Whoever holds it can both verify and mint tokens
for this API, so store it like any credential. Unset, Auth0 generates one
for an HS256 API (reported in the signing_secret output) and keeps it
across applies; set it to rotate to a secret of your own. RS256 and PS256
APIs do not use it.

- rule: {"string":{"minLen":"16"}}

### spec.accessToken

`Auth0ResourceServerAccessToken`

access_token configures what Auth0 puts into the access tokens this API
issues. It has no value of its own in the provider: declare the live
configuration when adopting an existing API; leaving it unset on an adopted
API clears it.

### spec.accessToken.claimsMapping

`Auth0ResourceServerClaimsMapping`

claims_mapping maps values of an anonymous session onto claims of the
access tokens issued in it. Anonymous sessions are an Enterprise Early
Access feature.

https://auth0.com/docs/manage-users/sessions/anonymous-sessions/configure-custom-claims-for-anonymous-sessions

### spec.accessToken.claimsMapping.customClaims

`[]Auth0ResourceServerCustomClaim`

custom_claims are the claims to add, at most 20. The list is sent whole:
it replaces every claim the API had, and declaring claims_mapping with no
claims clears them.

- rule: {"repeated":{"maxItems":"20"}}

### spec.accessToken.claimsMapping.customClaims[].name

`string` · required

name is the claim to emit in the access token (for example "country"),
stored with the casing given. Reserved OIDC and JWT claim names (sub, aud,
exp, ...) are refused by Auth0.

- rule: {"required":true,"string":{"maxLen":"255"}}

### spec.accessToken.claimsMapping.customClaims[].expression

`string` · required

expression is the dot path the claim's value is read from in the
anonymous session (for example "anonymous_session.metadata.country").

- rule: expression is a dot path of at least two names, such as anonymous_session.metadata.country (letters, digits, underscores, and hyphens inside a name)
- rule: {"required":true,"string":{"maxLen":"255"}}

### spec.authorizationDetails

`[]Auth0ResourceServerAuthorizationDetail`

authorization_details are the Rich Authorization Request types this API
accepts (for example "payment", "money_transfer"): structured requests a
client sends in an authorization_details parameter, which the person
approves and which land in the access token. A client grant's
authorization_details_types names which of them an application may
request. Clients send them through Pushed Authorization Requests (which
need the Enterprise plan with the Highly Regulated Identity add-on) or
Client-Initiated Backchannel Authentication (which needs the Enterprise
plan or an add-on). Empty is unmanaged: the API keeps the types it has. A
single {disable: true} entry removes every type.

https://auth0.com/docs/get-started/apis/configure-rich-authorization-requests

- rule: each authorization_details entry names a type (for example payment), or is the single {disable: true} entry that removes every type

### spec.authorizationDetails[].type

`string` · optional (explicit presence)

type is the authorization_details type (for example "payment"), the value
of the "type" field of each object a client sends.

### spec.authorizationDetails[].disable

`bool` · optional (explicit presence)

disable, true, removes every authorization_details type from the API. It
is the only entry of the list when set.

### spec.authorizationPolicy

`Auth0ResourceServerAuthorizationPolicy`

authorization_policy attaches an authorization policy to this API by its
identifier. Auth0 offers authorization policies to Early Access tenants.
Unset, the API keeps the policy it has; removing it after it was applied
detaches the policy.

### spec.authorizationPolicy.policyId

`string` · optional (explicit presence)

policy_id is the identifier of the authorization policy to apply. Unset,
the policy Auth0 holds is kept.

- rule: {"string":{"maxLen":"1024"}}

### spec.proofOfPossession

`Auth0ResourceServerProofOfPossession`

proof_of_possession sender-constrains this API's access tokens: a token is
bound to the application that obtained it (by its mTLS certificate or its
DPoP key), so a stolen token is useless to anyone else. Applications that
require it must use it with an API that accepts it. Unset, the API keeps
what Auth0 holds (none on a new API).

https://auth0.com/docs/secure/sender-constraining/configure-sender-constraining

- rule: proof_of_possession.disable turns sender constraining off, so it cannot be combined with a mechanism or required: true
- rule: proof_of_possession names its mechanism (mtls or dpop) and whether it is required -- Auth0 needs both -- or sets disable: true to turn it off
- rule: mTLS sender constraining applies to every application (public applications cannot present a client certificate), so required_for cannot be public_clients with mechanism mtls

### spec.proofOfPossession.disable

`bool` · optional (explicit presence)

disable, true, removes the API's proof-of-possession configuration: its
tokens are plain bearer tokens again.

### spec.proofOfPossession.mechanism

`string` · optional (explicit presence)

mechanism is how tokens are bound to the application:
- "dpop": Demonstrating Proof-of-Possession -- the application signs a
  proof with a key pair of its own on every request; works for public
  applications (single-page and native) as well as confidential ones.
- "mtls": the application's mutual-TLS client certificate -- confidential
  applications only. Needs the Enterprise plan with the Highly Regulated
  Identity add-on.

- rule: {"string":{"in":["mtls","dpop"]}}

### spec.proofOfPossession.required

`bool` · optional (explicit presence)

required, true, refuses to issue this API an access token that is not
sender-constrained (for the applications required_for names). False
accepts sender-constrained tokens without demanding them.

### spec.proofOfPossession.requiredFor

`string` · optional (explicit presence)

required_for is which applications must sender-constrain their tokens
when required is true: "all_clients", or "public_clients" (single-page and
native applications; DPoP only). It has no value of its own in the
provider: once proof_of_possession is declared, declare the live value
when adopting an existing API; leaving it unset on an adopted API resets
it.

- rule: {"string":{"in":["all_clients","public_clients"]}}

### spec.subjectTypeAuthorization

`Auth0ResourceServerSubjectTypeAuthorization`

subject_type_authorization is the API's access policy: which applications
can get an access token for it, decided separately for tokens that act for
a person (user) and tokens an application gets for itself (client).
Third-party applications always need a client grant, whatever the policy
(third_party_client_default_grants gives every one of them the same
grant). Unset, and for each policy left unset, the API keeps what Auth0
holds: on a new API, every application may get a user-delegated token and
a client token needs a grant.

https://auth0.com/docs/get-started/apis/api-access-policies-for-applications

### spec.subjectTypeAuthorization.user

`Auth0ResourceServerUserAuthorization`

user is the policy for user-delegated access: tokens an application gets
to call the API on a signed-in person's behalf (every flow but client
credentials).

### spec.subjectTypeAuthorization.user.policy

`string` · optional (explicit presence)

policy is who can get a token to call this API on a person's behalf:
- "allow_all": every first-party application in the tenant, with no grant
  of its own ("All apps allowed"). Third-party applications still need a
  grant -- theirs, or a default grant (third_party_client_default_grants
  with subject_type user).
- "require_client_grant": only applications holding a user grant for this
  API, first-party ones included; the grant caps the scopes they can
  request ("Per-app authorization", the least-privilege choice Auth0
  recommends). Applications must then name the scopes they want in each
  token request.
- "deny_all": no application, whatever its grants ("No apps allowed").
Unset, the API keeps the policy Auth0 holds (allow_all on a new API): the
block is sent only with its policy.

- rule: {"string":{"in":["allow_all","deny_all","require_client_grant"]}}

### spec.subjectTypeAuthorization.client

`Auth0ResourceServerClientAuthorization`

client is the policy for client access: tokens an application gets for
itself through the client-credentials flow (machine to machine).

### spec.subjectTypeAuthorization.client.policy

`string` · optional (explicit presence)

policy is which applications can get a token for themselves, through the
client-credentials flow, to call this API:
- "require_client_grant": only applications holding a client grant for
  this API; the grant names the scopes they receive ("Per-app
  authorization").
- "deny_all": none, whatever their grants ("No apps allowed") -- for an
  API only people use, through applications acting on their behalf.
Unset, the API keeps the policy Auth0 holds (require_client_grant on a new
API): the block is sent only with its policy.

- rule: {"string":{"in":["deny_all","require_client_grant"]}}

### spec.subjectTypeAuthorization.anonymousUser

`Auth0ResourceServerAnonymousUserAuthorization`

anonymous_user is the policy for tokens issued in an anonymous session,
before anyone signs in. Anonymous sessions are an Enterprise Early Access
feature.

### spec.subjectTypeAuthorization.anonymousUser.policy

`string` · optional (explicit presence)

policy is which applications can get a token for this API in an anonymous
session: "require_client_grant" (applications holding a grant with subject
type anonymous_user) or "deny_all" (none). Unset, the API keeps the policy
Auth0 holds (deny_all on a new API): the block is sent only with its
policy.

- rule: {"string":{"in":["deny_all","require_client_grant"]}}

### spec.tokenEncryption

`Auth0ResourceServerTokenEncryption`

token_encryption encrypts this API's access tokens (a signed JWT nested in
a JWE) with the API's public key, so only the API -- holding the private
key -- can read what they carry; applications and anything in between see
an opaque token. Needs the Enterprise plan with the Highly Regulated
Identity add-on. Unset, the API keeps what Auth0 holds (unencrypted on a
new API).

https://auth0.com/docs/get-started/apis/configure-json-web-encryption

- rule: token_encryption.disable turns encryption off, so it cannot be combined with a format or an encryption_key
- rule: token_encryption names both its format (compact-nested-jwe) and the API's encryption_key -- Auth0 needs both -- or sets disable: true to turn it off

### spec.tokenEncryption.disable

`bool` · optional (explicit presence)

disable, true, removes the API's token encryption: its tokens are signed
JWTs anyone holding them can read again.

### spec.tokenEncryption.format

`string` · optional (explicit presence)

format is the encrypted token's format. "compact-nested-jwe" (a signed JWT
inside a compact JWE) is the only one Auth0 offers.

- rule: {"string":{"in":["compact-nested-jwe"]}}

### spec.tokenEncryption.encryptionKey

`Auth0ResourceServerTokenEncryptionKey`

encryption_key is the API's public key, which Auth0 encrypts tokens to.

### spec.tokenEncryption.encryptionKey.algorithm

`string` · required

algorithm is the key-management algorithm tokens are encrypted with:
"RSA-OAEP-256", "RSA-OAEP-384" or "RSA-OAEP-512".

- rule: {"required":true,"string":{"in":["RSA-OAEP-256","RSA-OAEP-384","RSA-OAEP-512"]}}

### spec.tokenEncryption.encryptionKey.pem

`string` · required

pem is the API's RSA public key in PEM format ("-----BEGIN PUBLIC
KEY-----..."), at most 4096 characters. It is public by design: only the
private key, which stays with the API, can decrypt.

- rule: {"required":true,"string":{"maxLen":"4096"}}

### spec.tokenEncryption.encryptionKey.kid

`string` · optional (explicit presence)

kid is the key's identifier, carried in each token's header so the API
picks the right private key (useful while rotating). Unset, Auth0 assigns
one.

- rule: {"string":{"maxLen":"128"}}

### spec.tokenEncryption.encryptionKey.name

`string` · optional (explicit presence)

name is the key's name in the Auth0 dashboard. Unset, Auth0 names it.

- rule: {"string":{"maxLen":"128"}}

### spec.thirdPartyClientDefaultGrants

`[]Auth0ResourceServerThirdPartyClientDefaultGrant`

third_party_client_default_grants are the grants every third-party
application in the tenant gets on this API without a grant of its own --
the applications external developers, partners and AI agents register,
including every application registered through Dynamic Client Registration
or from a Client ID Metadata Document, which no one can grant access one by
one. One entry per subject type: "user" for tokens an application gets on a
person's behalf, "client" for tokens it gets for itself. A grant made for
one application (Auth0Client's api_grants) takes precedence over the
default. Empty is unmanaged: default grants made elsewhere are left alone.
System APIs (the Management API, My Account API) do not accept default
grants.

https://auth0.com/docs/get-started/applications/application-access-to-apis-client-grants#default-permissions-for-third-party-applications

- rule: a default grant names its scopes, or sets allow_all_scopes: true to grant every scope the API defines (now and later) -- one or the other, not both
- rule: authorization_details_types applies only to user-delegated access -- set it on the grant with subject_type user
- rule: third-party applications cannot use allow_any_organization: Auth0 requires an organization client grant per organization for them

### spec.thirdPartyClientDefaultGrants[].subjectType

`string` · required

subject_type is which tokens the grant is for:
- "user": tokens a third-party application gets to call the API on a
  signed-in person's behalf. The token carries the scopes the application
  asked for, this grant allows, the person's roles permit (with
  enforce_policies) and the person consented to.
- "client": tokens a third-party application gets for itself through the
  client-credentials flow, carrying the scopes this grant names.
It is the grant's identity: changing it replaces the grant.

- rule: {"required":true,"string":{"in":["user","client"]}}

### spec.thirdPartyClientDefaultGrants[].scopes

`[]string`

scopes are the API's scopes (from spec.scopes) the grant allows. Required
unless allow_all_scopes is true.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"280"}}}}

### spec.thirdPartyClientDefaultGrants[].authorizationDetailsTypes

`[]string`

authorization_details_types are the API's authorization_details types a
third-party application may request on a person's behalf. User grants
only.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"255"}}}}

### spec.thirdPartyClientDefaultGrants[].allowAllScopes

`bool` · optional (explicit presence)

allow_all_scopes, true, grants every scope the API defines, including
scopes added later, in place of a scopes list.

### spec.thirdPartyClientDefaultGrants[].organizationUsage

`string` · optional (explicit presence)

organization_usage is how a third-party application may use Organizations
when it gets a token for itself (subject_type client): "deny" (never; the
default), "allow" (with or without an organization) or "require" (always
for an organization). Each organization still needs its own organization
client grant for third-party applications.

- rule: {"string":{"in":["deny","allow","require"]}}

### spec.thirdPartyClientDefaultGrants[].allowAnyOrganization

`bool` · optional (explicit presence)

allow_any_organization would let the grant be used with any organization
without an organization client grant. Auth0 does not offer it to
third-party applications, so on a default grant it can only be false.

## Validation Rules

- `spec.third_party_client_default_grants.one_per_subject_type`: third_party_client_default_grants holds at most one grant per subject_type (one for user, one for client): Auth0 keeps a single default grant per API and subject type
- `spec.authorization_details.disable_alone`: authorization_details clears every registered type with a single {disable: true} entry, so it cannot also list types -- keep the types, or keep only the disable entry

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0ResourceServer, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the internal Auth0 identifier for this resource server. This is a unique string assigned by Auth0 when the resource is created. Used for API calls to manage the resource server. |
| `status.outputs.identifier` | `string` | identifier is the API identifier (audience) for this resource server. This is the value used in authorization requests as the "audience" parameter. Same as the identifier specified in the spec. |
| `status.outputs.name` | `string` | name is the friendly display name of the resource server. Derived from spec.name or metadata.name. |
| `status.outputs.signing_alg` | `string` | signing_alg is the algorithm used to sign tokens for this API. One of: RS256, HS256, PS256 |
| `status.outputs.signing_secret` | `string` (sensitive) | signing_secret is the secret used for signing tokens (HS256 only). This is only populated when signing_alg is HS256. IMPORTANT: Keep this secret secure and never expose in client-side code. |
| `status.outputs.token_lifetime` | `string` | token_lifetime is the configured token validity duration in seconds. |
| `status.outputs.token_lifetime_for_web` | `string` | token_lifetime_for_web is the token validity for implicit/hybrid flows. |
| `status.outputs.allow_offline_access` | `string` | allow_offline_access indicates if refresh tokens can be issued. |
| `status.outputs.skip_consent_for_verifiable_first_party_clients` | `string` | skip_consent_for_verifiable_first_party_clients indicates consent skip setting. |
| `status.outputs.enforce_policies` | `string` | enforce_policies indicates if RBAC is enabled for this API. |
| `status.outputs.token_dialect` | `string` | token_dialect is the access token format configured for this API. |
| `status.outputs.is_system` | `string` | is_system indicates if this is a system-managed resource server. System resource servers (like the Auth0 Management API) cannot be modified. |
| `status.outputs.client_id` | `string` | client_id is the associated client ID if one has been linked. Some resource servers may have an associated client for certain features. |
| `status.outputs.third_party_client_default_grant_ids` | `map<string, string>` | third_party_client_default_grant_ids are the identifiers Auth0 assigned the default grants for third-party applications (cgr_...), keyed by subject type ("user", "client") -- one entry per spec.third_party_client_default_grants entry. Each grant imports by its id. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| Auth0Client | `spec.apiGrants[].audience` | `status.outputs.identifier` |
| Auth0TenantSettings | `spec.defaultAudience` | `status.outputs.identifier` |
| Auth0User | `spec.permissions[].resourceServerIdentifier` | `status.outputs.identifier` |

## See Also

- [Overview](../README.md)
