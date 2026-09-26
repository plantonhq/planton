# AwsCognitoUserPoolClient

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsCognitoUserPoolClientSpec defines the desired configuration for a Cognito
User Pool app client -- the OAuth 2.0 / OIDC contract between ONE application
and a user pool.

An app client is deliberately its own resource rather than a field on the
pool: a pool serves many applications (a web frontend, a mobile app, an M2M
service), each with its own grant types, redirect URLs, token lifetimes, and
client ID -- and that client ID is what downstream systems reference (an API
Gateway JWT authorizer's audience, an ALB authenticate-cognito action).

Key design notes:
- `generate_secret` is **ForceNew**: whether a client is confidential
  (server-side, holds a secret) or public (SPA/mobile, no secret) is decided
  at creation. Public clients rely on PKCE instead of a secret.
- Token lifetimes pair a value with a unit (`token_validity_units`). AWS
  bounds them: access/ID tokens 5 minutes to 24 hours, refresh tokens
  60 minutes to 10 years.
- Federated sign-in is enabled per client: list the identity providers
  (by reference or literal name, plus "COGNITO" for the pool's own
  directory) in `supported_identity_providers`.

Credentials, region, and deployment workflow live outside this spec in
stack inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoUserPoolClient
metadata:
  name: test-web-app
  org: test-org
  env: dev
  id: awscogclient-test-001
spec:
  region: us-west-2
  userPoolId:
    value: us-west-2_TestPool123
  allowedOauthFlowsUserPoolClient: true
  allowedOauthFlows:
    - code
  allowedOauthScopes:
    - openid
    - email
    - profile
  callbackUrls:
    - http://localhost:3000/callback
  logoutUrls:
    - http://localhost:3000/logout
  explicitAuthFlows:
    - ALLOW_USER_SRP_AUTH
    - ALLOW_REFRESH_TOKEN_AUTH
  preventUserExistenceErrors: ENABLED
  riskConfiguration:
    accountTakeover:
      highAction:
        eventAction: MFA_REQUIRED
        notify: false
    riskException:
      skippedIpRanges:
        - 10.0.0.0/8
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.userPoolId` | `string \| valueFrom` | yes |  | AwsCognitoUserPool (`status.outputs.user_pool_id`) |
| `spec.generateSecret` | `bool` |  |  |  |
| `spec.allowedOauthFlowsUserPoolClient` | `bool` |  |  |  |
| `spec.allowedOauthFlows` | `[]string` |  |  |  |
| `spec.allowedOauthScopes` | `[]string` |  |  |  |
| `spec.callbackUrls` | `[]string` |  |  |  |
| `spec.logoutUrls` | `[]string` |  |  |  |
| `spec.defaultRedirectUri` | `string` |  |  |  |
| `spec.supportedIdentityProviders` | `[]string \| valueFrom` |  |  | AwsCognitoIdentityProvider (`status.outputs.provider_name`) |
| `spec.explicitAuthFlows` | `[]string` |  |  |  |
| `spec.authSessionValidity` | `int32` |  |  |  |
| `spec.accessTokenValidity` | `int32` |  |  |  |
| `spec.idTokenValidity` | `int32` |  |  |  |
| `spec.refreshTokenValidity` | `int32` |  |  |  |
| `spec.tokenValidityUnits` | `AwsCognitoUserPoolClientTokenValidityUnits` |  |  |  |
| `spec.tokenValidityUnits.accessToken` | `string` |  |  |  |
| `spec.tokenValidityUnits.idToken` | `string` |  |  |  |
| `spec.tokenValidityUnits.refreshToken` | `string` |  |  |  |
| `spec.refreshTokenRotation` | `AwsCognitoUserPoolClientRefreshTokenRotation` |  |  |  |
| `spec.refreshTokenRotation.feature` | `string` | yes |  |  |
| `spec.refreshTokenRotation.retryGracePeriodSeconds` | `int32` |  |  |  |
| `spec.enableTokenRevocation` | `bool` |  |  |  |
| `spec.enablePropagateAdditionalUserContextData` | `bool` |  |  |  |
| `spec.preventUserExistenceErrors` | `string` |  |  |  |
| `spec.readAttributes` | `[]string` |  |  |  |
| `spec.writeAttributes` | `[]string` |  |  |  |
| `spec.analyticsConfiguration` | `AwsCognitoUserPoolClientAnalyticsConfig` |  |  |  |
| `spec.analyticsConfiguration.applicationArn` | `string` |  |  |  |
| `spec.analyticsConfiguration.applicationId` | `string` |  |  |  |
| `spec.analyticsConfiguration.externalId` | `string` |  |  |  |
| `spec.analyticsConfiguration.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.analyticsConfiguration.userDataShared` | `bool` |  |  |  |
| `spec.riskConfiguration` | `AwsCognitoUserPoolClientRiskConfiguration` |  |  |  |
| `spec.riskConfiguration.accountTakeover` | `AwsCognitoUserPoolClientAccountTakeoverConfig` |  |  |  |
| `spec.riskConfiguration.accountTakeover.lowAction` | `AwsCognitoUserPoolClientAccountTakeoverAction` |  |  |  |
| `spec.riskConfiguration.accountTakeover.lowAction.eventAction` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.lowAction.notify` | `bool` |  |  |  |
| `spec.riskConfiguration.accountTakeover.mediumAction` | `AwsCognitoUserPoolClientAccountTakeoverAction` |  |  |  |
| `spec.riskConfiguration.accountTakeover.mediumAction.eventAction` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.mediumAction.notify` | `bool` |  |  |  |
| `spec.riskConfiguration.accountTakeover.highAction` | `AwsCognitoUserPoolClientAccountTakeoverAction` |  |  |  |
| `spec.riskConfiguration.accountTakeover.highAction.eventAction` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.highAction.notify` | `bool` |  |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration` | `AwsCognitoUserPoolClientRiskNotifyConfig` |  |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.sourceArn` | `string \| valueFrom` | yes |  | AwsSesEmailIdentity (`status.outputs.identity_arn`) |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.from` | `string` |  |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.replyTo` | `string` |  |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.blockEmail` | `AwsCognitoUserPoolClientRiskNotifyEmail` |  |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.blockEmail.subject` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.blockEmail.htmlBody` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.blockEmail.textBody` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.mfaEmail` | `AwsCognitoUserPoolClientRiskNotifyEmail` |  |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.mfaEmail.subject` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.mfaEmail.htmlBody` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.mfaEmail.textBody` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.noActionEmail` | `AwsCognitoUserPoolClientRiskNotifyEmail` |  |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.noActionEmail.subject` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.noActionEmail.htmlBody` | `string` | yes |  |  |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.noActionEmail.textBody` | `string` | yes |  |  |
| `spec.riskConfiguration.compromisedCredentials` | `AwsCognitoUserPoolClientCompromisedCredentialsConfig` |  |  |  |
| `spec.riskConfiguration.compromisedCredentials.eventAction` | `string` | yes |  |  |
| `spec.riskConfiguration.compromisedCredentials.eventFilter` | `[]string` |  |  |  |
| `spec.riskConfiguration.riskException` | `AwsCognitoUserPoolClientRiskExceptionConfig` |  |  |  |
| `spec.riskConfiguration.riskException.blockedIpRanges` | `[]string` |  |  |  |
| `spec.riskConfiguration.riskException.skippedIpRanges` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.userPoolId

`string | valueFrom` · required

The Cognito User Pool this client authenticates against.
Format: "{region}_{poolId}" (e.g., "us-east-1_Ab1Cd2EfG").
ForceNew -- a client cannot be moved between pools.
Accepts a direct pool ID or a reference to an AwsCognitoUserPool resource.

- references: AwsCognitoUserPool (`status.outputs.user_pool_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_id}} -- a bare string does not parse

### spec.generateSecret

`bool`

Whether AWS generates a client secret. ForceNew -- cannot be changed after
creation. Set true for confidential clients (server-side applications that
can protect the secret, machine-to-machine clients); leave false for
public clients (SPAs, mobile apps), which authenticate with PKCE instead.

### spec.allowedOauthFlowsUserPoolClient

`bool`

Enable OAuth 2.0 flows for this client. Must be true for
`allowed_oauth_flows` and `allowed_oauth_scopes` to take effect.

### spec.allowedOauthFlows

`[]string`

OAuth 2.0 grant types this client can use. Valid values:
- "code": Authorization Code grant (recommended for most apps; pair with
  PKCE for public clients).
- "implicit": Implicit grant (legacy; tokens leak into browser history --
  avoid for new applications).
- "client_credentials": Client Credentials grant (machine-to-machine;
  requires a client secret and custom scopes from a resource server, and
  cannot be combined with the user-facing grants on the same client).

- rule: {"repeated":{"maxItems":"3"}}

### spec.allowedOauthScopes

`[]string`

OAuth 2.0 scopes this client can request. Built-in values: "openid",
"email", "profile", "phone", "aws.cognito.signin.user.admin". Custom
scopes minted by an AwsCognitoResourceServer use the form
"{resource-server-identifier}/{scope_name}" -- reference the resource
server's scope_identifiers output for the exact strings.

- rule: {"repeated":{"maxItems":"50"}}

### spec.callbackUrls

`[]string`

Callback URLs for OAuth redirects after authentication. Required for
Authorization Code and Implicit grants. Maximum 100 URLs.

- rule: {"repeated":{"maxItems":"100"}}

### spec.logoutUrls

`[]string`

URLs where Cognito redirects after sign-out. Maximum 100 URLs.

- rule: {"repeated":{"maxItems":"100"}}

### spec.defaultRedirectUri

`string`

Default redirect URI. Must be one of the `callback_urls`. Used when no
redirect_uri is specified in the authorization request.

- rule: {"string":{"maxLen":"1024"}}

### spec.supportedIdentityProviders

`[]string | valueFrom`

Identity providers this client offers at sign-in. Use the literal
"COGNITO" for the pool's own user directory, and add federated providers
by name -- either as literals ("Google", "CorpOkta") or as references to
AwsCognitoIdentityProvider resources (which also gives the deployment
graph the right ordering: the provider exists before the client lists it).
When omitted, AWS enables all of the pool's providers for this client.

- references: AwsCognitoIdentityProvider (`status.outputs.provider_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoIdentityProvider, name: <that resource's name>, fieldPath: status.outputs.provider_name}} -- a bare string does not parse

### spec.explicitAuthFlows

`[]string`

Explicit authentication flows enabled for this client. Controls which
authentication APIs the client can use. Valid values:
- "ALLOW_USER_SRP_AUTH": Secure Remote Password (recommended)
- "ALLOW_REFRESH_TOKEN_AUTH": Enable token refresh (always recommended)
- "ALLOW_USER_PASSWORD_AUTH": Direct username/password (less secure)
- "ALLOW_ADMIN_USER_PASSWORD_AUTH": Admin-initiated auth
- "ALLOW_CUSTOM_AUTH": Custom auth flow via Lambda triggers
- "ALLOW_USER_AUTH": Choice-based sign-in (passwordless first factors --
  pairs with the pool's allowed_first_auth_factors)
The legacy pre-ALLOW spellings ("ADMIN_NO_SRP_AUTH",
"CUSTOM_AUTH_FLOW_ONLY", "USER_PASSWORD_AUTH") are accepted for pools that
still use them, but cannot be mixed with ALLOW_* values on one client.

### spec.authSessionValidity

`int32` · optional (explicit presence)

How long the session token of one sign-in attempt (the challenge
handshake) stays valid, in minutes. Range: 3-15. AWS default: 3.

- rule: {"int32":{"lte":15,"gte":3}}

### spec.accessTokenValidity

`int32` · optional (explicit presence)

Access token lifetime, in `token_validity_units.access_token` units
(default: hours). AWS bounds the result to 5 minutes - 24 hours.
AWS default: 1 hour.

### spec.idTokenValidity

`int32` · optional (explicit presence)

ID token lifetime, in `token_validity_units.id_token` units (default:
hours). AWS bounds the result to 5 minutes - 24 hours. AWS default: 1 hour.

### spec.refreshTokenValidity

`int32` · optional (explicit presence)

Refresh token lifetime, in `token_validity_units.refresh_token` units
(default: days). AWS bounds the result to 60 minutes - 10 years.
AWS default: 30 days.

### spec.tokenValidityUnits

`AwsCognitoUserPoolClientTokenValidityUnits`

The units the three token lifetimes are expressed in.

- rule: token validity units must be 'seconds', 'minutes', 'hours', or 'days'

### spec.tokenValidityUnits.accessToken

`string`

Unit for access_token_validity: "seconds", "minutes", "hours" (AWS
default), or "days".

### spec.tokenValidityUnits.idToken

`string`

Unit for id_token_validity: "seconds", "minutes", "hours" (AWS default),
or "days".

### spec.tokenValidityUnits.refreshToken

`string`

Unit for refresh_token_validity: "seconds", "minutes", "hours", or "days"
(AWS default).

### spec.refreshTokenRotation

`AwsCognitoUserPoolClientRefreshTokenRotation`

Refresh-token rotation: each refresh issues a NEW refresh token and
retires the old one, shrinking the blast radius of a stolen token.
When ENABLED, do not also list ALLOW_REFRESH_TOKEN_AUTH in
explicit_auth_flows -- rotation owns the refresh behavior and AWS
rejects the combination.

- rule: refresh_token_rotation feature must be 'ENABLED' or 'DISABLED'

### spec.refreshTokenRotation.feature

`string` · required

"ENABLED" or "DISABLED".

- rule: {"required":true}

### spec.refreshTokenRotation.retryGracePeriodSeconds

`int32` · optional (explicit presence)

How long (seconds) the RETIRED refresh token keeps working after a
rotation, absorbing clients that lose the response carrying the new token.
Range: 0-60, where an EXPLICIT 0 disables the grace period (immediate
retirement -- AWS's default posture). Tri-state: unset leaves the choice
to AWS, explicit 0 pins immediate retirement, 1-60 grants that many
seconds of grace.

- rule: {"int32":{"lte":60,"gte":0}}

### spec.enableTokenRevocation

`bool` · optional (explicit presence)

Whether tokens can be revoked (sign-out revokes the refresh token and the
access/ID tokens minted from it). AWS default: true. Only set false when
the marginal token-validation latency matters more than revocability.

### spec.enablePropagateAdditionalUserContextData

`bool`

Propagate client IP and user-agent to Cognito threat protection for
server-side flows (where Cognito otherwise sees only the server's IP).
Requires a client secret and the pool's threat protection to be active.

### spec.preventUserExistenceErrors

`string`

How to handle requests for non-existent users. Valid values:
- "ENABLED": Return the same error for non-existent and incorrect-password
  users to prevent user enumeration attacks (recommended).
- "LEGACY": Return different errors (reveals whether a user exists).

### spec.readAttributes

`[]string`

User attributes this client can READ (e.g. "email", "email_verified",
"custom:tenant_id"). When omitted, AWS grants read access to all
attributes.

### spec.writeAttributes

`[]string`

User attributes this client can WRITE. When omitted, AWS grants write
access to all mutable attributes.

### spec.analyticsConfiguration

`AwsCognitoUserPoolClientAnalyticsConfig`

Amazon Pinpoint analytics wiring: Cognito publishes sign-in/sign-up events
to a Pinpoint project for user-journey analytics.

- rule: set exactly one of application_arn or application_id
- rule: application_id requires both external_id and role_arn (application_arn derives them automatically)

### spec.analyticsConfiguration.applicationArn

`string`

The Pinpoint project (application) ARN. Mutually exclusive with
`application_id`; when set, AWS derives the publish role itself.

### spec.analyticsConfiguration.applicationId

`string`

The Pinpoint application ID. Requires `external_id` and `role_arn`.

### spec.analyticsConfiguration.externalId

`string`

The external ID for the role assumption (confused-deputy guard).

### spec.analyticsConfiguration.roleArn

`string | valueFrom`

The IAM role Cognito assumes to publish events to Pinpoint. Accepts a
direct role ARN or a reference to an AwsIamRole resource.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.analyticsConfiguration.userDataShared

`bool`

Whether Cognito includes user data (endpoint attributes) in the events
it publishes.

### spec.riskConfiguration

`AwsCognitoUserPoolClientRiskConfiguration`

Client-scoped risk configuration for threat protection. When set, this
client stops following the pool-wide risk configuration (set on the
AwsCognitoUserPool spec) and uses these responses instead. Requires the
POOL to have threat protection active (`user_pool_add_ons.
advanced_security_mode` of "AUDIT" or "ENFORCED") -- a cross-resource
requirement this spec cannot validate; AWS rejects the configuration
when threat protection is off.

- rule: set at least one of account_takeover, compromised_credentials, or risk_exception

### spec.riskConfiguration.accountTakeover

`AwsCognitoUserPoolClientAccountTakeoverConfig`

Automated responses to sign-in attempts Cognito assesses as possible
account takeover, by risk level, plus the notification templates for
"notify the user" actions.

- rule: set at least one of low_action, medium_action, or high_action
- rule: notify_configuration requires at least one action with notify enabled

### spec.riskConfiguration.accountTakeover.lowAction

`AwsCognitoUserPoolClientAccountTakeoverAction`

Response to a LOW-risk assessment.

- rule: event_action must be 'BLOCK', 'MFA_IF_CONFIGURED', 'MFA_REQUIRED', or 'NO_ACTION'

### spec.riskConfiguration.accountTakeover.lowAction.eventAction

`string` · required

The response Cognito takes (acts only in ENFORCED mode; AUDIT gathers
metrics without acting):
- "BLOCK": reject the sign-in attempt.
- "MFA_IF_CONFIGURED": require MFA when the user has a factor set up,
  otherwise allow the attempt.
- "MFA_REQUIRED": require MFA; users without a factor are blocked.
- "NO_ACTION": allow the attempt (pair with notify to warn the user).

- rule: {"required":true}

### spec.riskConfiguration.accountTakeover.lowAction.notify

`bool`

Whether Cognito emails the user about the event using the templates in
notify_configuration.

### spec.riskConfiguration.accountTakeover.mediumAction

`AwsCognitoUserPoolClientAccountTakeoverAction`

Response to a MEDIUM-risk assessment.

- rule: event_action must be 'BLOCK', 'MFA_IF_CONFIGURED', 'MFA_REQUIRED', or 'NO_ACTION'

### spec.riskConfiguration.accountTakeover.mediumAction.eventAction

`string` · required

The response Cognito takes (acts only in ENFORCED mode; AUDIT gathers
metrics without acting):
- "BLOCK": reject the sign-in attempt.
- "MFA_IF_CONFIGURED": require MFA when the user has a factor set up,
  otherwise allow the attempt.
- "MFA_REQUIRED": require MFA; users without a factor are blocked.
- "NO_ACTION": allow the attempt (pair with notify to warn the user).

- rule: {"required":true}

### spec.riskConfiguration.accountTakeover.mediumAction.notify

`bool`

Whether Cognito emails the user about the event using the templates in
notify_configuration.

### spec.riskConfiguration.accountTakeover.highAction

`AwsCognitoUserPoolClientAccountTakeoverAction`

Response to a HIGH-risk assessment.

- rule: event_action must be 'BLOCK', 'MFA_IF_CONFIGURED', 'MFA_REQUIRED', or 'NO_ACTION'

### spec.riskConfiguration.accountTakeover.highAction.eventAction

`string` · required

The response Cognito takes (acts only in ENFORCED mode; AUDIT gathers
metrics without acting):
- "BLOCK": reject the sign-in attempt.
- "MFA_IF_CONFIGURED": require MFA when the user has a factor set up,
  otherwise allow the attempt.
- "MFA_REQUIRED": require MFA; users without a factor are blocked.
- "NO_ACTION": allow the attempt (pair with notify to warn the user).

- rule: {"required":true}

### spec.riskConfiguration.accountTakeover.highAction.notify

`bool`

Whether Cognito emails the user about the event using the templates in
notify_configuration.

### spec.riskConfiguration.accountTakeover.notifyConfiguration

`AwsCognitoUserPoolClientRiskNotifyConfig`

SES sending configuration and message templates for the notification
emails sent when an action has notify enabled. Required by AWS when any
action notifies the user.

### spec.riskConfiguration.accountTakeover.notifyConfiguration.sourceArn

`string | valueFrom` · required

The SES identity Cognito sends notification emails from. Accepts a
direct identity ARN or a reference to an AwsSesEmailIdentity resource.

- references: AwsSesEmailIdentity (`status.outputs.identity_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSesEmailIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_arn}} -- a bare string does not parse

### spec.riskConfiguration.accountTakeover.notifyConfiguration.from

`string`

"From" address shown on notification emails. When omitted, AWS derives
it from the source identity.

### spec.riskConfiguration.accountTakeover.notifyConfiguration.replyTo

`string`

Reply-to address on notification emails.

### spec.riskConfiguration.accountTakeover.notifyConfiguration.blockEmail

`AwsCognitoUserPoolClientRiskNotifyEmail`

Template for the "we blocked a sign-in attempt" notification.

### spec.riskConfiguration.accountTakeover.notifyConfiguration.blockEmail.subject

`string` · required

Email subject. 1-140 characters.

- rule: {"string":{"minLen":"1","maxLen":"140"}}

### spec.riskConfiguration.accountTakeover.notifyConfiguration.blockEmail.htmlBody

`string` · required

HTML email body. 6-20000 characters.

- rule: {"string":{"minLen":"6","maxLen":"20000"}}

### spec.riskConfiguration.accountTakeover.notifyConfiguration.blockEmail.textBody

`string` · required

Plain-text email body. 6-20000 characters.

- rule: {"string":{"minLen":"6","maxLen":"20000"}}

### spec.riskConfiguration.accountTakeover.notifyConfiguration.mfaEmail

`AwsCognitoUserPoolClientRiskNotifyEmail`

Template for the "we required MFA on a sign-in attempt" notification.

### spec.riskConfiguration.accountTakeover.notifyConfiguration.mfaEmail.subject

`string` · required

Email subject. 1-140 characters.

- rule: {"string":{"minLen":"1","maxLen":"140"}}

### spec.riskConfiguration.accountTakeover.notifyConfiguration.mfaEmail.htmlBody

`string` · required

HTML email body. 6-20000 characters.

- rule: {"string":{"minLen":"6","maxLen":"20000"}}

### spec.riskConfiguration.accountTakeover.notifyConfiguration.mfaEmail.textBody

`string` · required

Plain-text email body. 6-20000 characters.

- rule: {"string":{"minLen":"6","maxLen":"20000"}}

### spec.riskConfiguration.accountTakeover.notifyConfiguration.noActionEmail

`AwsCognitoUserPoolClientRiskNotifyEmail`

Template for the "we observed a risky sign-in attempt" (no action taken)
notification.

### spec.riskConfiguration.accountTakeover.notifyConfiguration.noActionEmail.subject

`string` · required

Email subject. 1-140 characters.

- rule: {"string":{"minLen":"1","maxLen":"140"}}

### spec.riskConfiguration.accountTakeover.notifyConfiguration.noActionEmail.htmlBody

`string` · required

HTML email body. 6-20000 characters.

- rule: {"string":{"minLen":"6","maxLen":"20000"}}

### spec.riskConfiguration.accountTakeover.notifyConfiguration.noActionEmail.textBody

`string` · required

Plain-text email body. 6-20000 characters.

- rule: {"string":{"minLen":"6","maxLen":"20000"}}

### spec.riskConfiguration.compromisedCredentials

`AwsCognitoUserPoolClientCompromisedCredentialsConfig`

Automated response when Cognito detects the presented credentials in a
known-compromised set.

- rule: event_action must be 'BLOCK' or 'NO_ACTION'
- rule: event_filter must contain only 'SIGN_IN', 'PASSWORD_CHANGE', and/or 'SIGN_UP'

### spec.riskConfiguration.compromisedCredentials.eventAction

`string` · required

The response Cognito takes (acts only in ENFORCED mode):
- "BLOCK": reject the attempt.
- "NO_ACTION": allow the attempt (still logged/audited).

- rule: {"required":true}

### spec.riskConfiguration.compromisedCredentials.eventFilter

`[]string`

Which authentication events are checked against the compromised set.
When empty, AWS checks all supported events. Valid values: "SIGN_IN",
"PASSWORD_CHANGE", "SIGN_UP".

### spec.riskConfiguration.riskException

`AwsCognitoUserPoolClientRiskExceptionConfig`

IP-range exceptions that bypass or force the risk decision.

- rule: set at least one of blocked_ip_ranges or skipped_ip_ranges
- rule: blocked_ip_ranges and skipped_ip_ranges entries must be CIDR notation (e.g. '192.0.2.0/24' or '2001:db8::/32')

### spec.riskConfiguration.riskException.blockedIpRanges

`[]string`

Always-BLOCK list: authentication requests from these CIDR ranges are
rejected regardless of risk assessment. Up to 200 entries.

- rule: {"repeated":{"maxItems":"200","items":{"string":{"minLen":"1"}}}}

### spec.riskConfiguration.riskException.skippedIpRanges

`[]string`

Always-ALLOW list: risk detection is skipped for these CIDR ranges.
Up to 200 entries.

- rule: {"repeated":{"maxItems":"200","items":{"string":{"minLen":"1"}}}}

## Validation Rules

- `allowed_oauth_flows_valid`: allowed_oauth_flows must contain only 'code', 'implicit', or 'client_credentials'
- `client_credentials_not_mixed`: 'client_credentials' cannot be combined with 'code' or 'implicit' on the same client -- create a separate machine-to-machine client
- `oauth_flows_require_oauth_enabled`: allowed_oauth_flows requires allowed_oauth_flows_user_pool_client to be true
- `redirect_grants_require_callback_urls`: the 'code' and 'implicit' grants require at least one callback_url
- `default_redirect_uri_in_callback_urls`: default_redirect_uri must be one of the callback_urls
- `explicit_auth_flows_valid`: explicit_auth_flows must contain only ALLOW_USER_SRP_AUTH, ALLOW_REFRESH_TOKEN_AUTH, ALLOW_USER_PASSWORD_AUTH, ALLOW_ADMIN_USER_PASSWORD_AUTH, ALLOW_CUSTOM_AUTH, ALLOW_USER_AUTH, or the legacy ADMIN_NO_SRP_AUTH / CUSTOM_AUTH_FLOW_ONLY / USER_PASSWORD_AUTH values
- `explicit_auth_flows_no_mixing`: legacy auth flow values (ADMIN_NO_SRP_AUTH, CUSTOM_AUTH_FLOW_ONLY, USER_PASSWORD_AUTH) cannot be mixed with ALLOW_* values on one client
- `refresh_rotation_excludes_refresh_auth_flow`: ALLOW_REFRESH_TOKEN_AUTH cannot be listed in explicit_auth_flows when refresh_token_rotation is ENABLED -- rotation owns the refresh behavior
- `access_token_validity_range`: access_token_validity must be between 5 minutes and 24 hours (in the configured token_validity_units)
- `id_token_validity_range`: id_token_validity must be between 5 minutes and 24 hours (in the configured token_validity_units)
- `refresh_token_validity_range`: refresh_token_validity must be between 60 minutes and 10 years (in the configured token_validity_units)
- `prevent_user_existence_errors_valid`: prevent_user_existence_errors must be 'ENABLED' or 'LEGACY' when set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCognitoUserPoolClient, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.client_id` | `string` | The app client ID (e.g. "1a2b3c4d5e6f7g8h9i0j"). The public identifier applications present at sign-in and token endpoints -- and the "aud" claim JWT authorizers validate. |
| `status.outputs.client_secret` | `string` (sensitive) | The app client secret. Only populated when `generate_secret` is true. Sensitive -- treat as a credential and handle securely; confidential clients present it at the token endpoint. |
| `status.outputs.user_pool_id` | `string` | The user pool this client belongs to, resolved from the spec reference. Application configs typically need the (pool id, client id) pair together, and a consumer holding only this resource gets both from its outputs. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.userPoolId` | AwsCognitoUserPool | `status.outputs.user_pool_id` |
| `spec.supportedIdentityProviders` | AwsCognitoIdentityProvider | `status.outputs.provider_name` |
| `spec.analyticsConfiguration.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.riskConfiguration.accountTakeover.notifyConfiguration.sourceArn` | AwsSesEmailIdentity | `status.outputs.identity_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsHttpApiGateway | `spec.authorizers[].jwtConfiguration.audiences` | `status.outputs.client_id` |
| AwsLbListener | `spec.defaultActions[].authenticateCognito.userPoolClientId` | `status.outputs.client_id` |
| AwsLbListenerRule | `spec.actions[].authenticateCognito.userPoolClientId` | `status.outputs.client_id` |

## See Also

- [Overview](../README.md)
