# AwsLbListener

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsLbListenerSpec defines an ELBv2 listener: the port/protocol entry point
on a load balancer, with the default action taken when no listener rule
matches a request.

A listener is a first-class node in the routing graph. One load balancer
carries many listeners (80 and 443 at minimum on most ALBs), each listener
owns its own TLS material and default behavior, and listener rules attach to
a specific listener as services deploy -- so the listener is the anchor both
certificates and per-service routing hang off.

The same kind serves both load balancer families, exactly as AWS models it:
- ALB listeners ("HTTP", "HTTPS") take the full action set -- forward,
  redirect, fixed-response, authenticate-cognito, authenticate-oidc,
  jwt-validation -- plus mutual TLS and HTTP header controls.
- NLB listeners ("TCP", "UDP", "TCP_UDP", "TLS") only forward; AWS rejects
  every other action type at the API.
Field comments call out the scope of every family-specific field.

The load balancer is create-only: moving a listener replaces it. Port,
protocol, certificates, and actions all update in place.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLbListener
metadata:
  name: https-443-demo
spec:
  region: us-west-2
  loadBalancerArn:
    value: arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/demo/50dc6c495c0c9188
  port: 443
  protocol: HTTPS
  certificateArn:
    value: arn:aws:acm:us-west-2:123456789012:certificate/12345678-1234-1234-1234-123456789012
  sslPolicy: ELBSecurityPolicy-TLS13-1-2-2021-06
  # A two-step default-action chain (authenticate, then forward with weights)
  # deliberately exercises the discriminated-union action shape and the
  # weighted forward block -- the parts of the variable contract that would
  # break silently if the action objects were mistyped.
  defaultActions:
    - type: authenticate-oidc
      authenticateOidc:
        issuer: https://accounts.google.com
        authorizationEndpoint: https://accounts.google.com/o/oauth2/v2/auth
        tokenEndpoint: https://oauth2.googleapis.com/token
        userInfoEndpoint: https://openidconnect.googleapis.com/v1/userinfo
        clientId: demo-client-id
        clientSecret: demo-client-secret
    - type: forward
      forward:
        targetGroups:
          - arn:
              value: arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/blue/943f017f100becff
            weight: 90
          - arn:
              value: arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/green/50dc6c495c0c9188
            weight: 10
        stickiness:
          enabled: true
          durationSeconds: 3600
  httpHeaders:
    response:
      strictTransportSecurity: "max-age=31536000; includeSubDomains"
      xContentTypeOptions: nosniff
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.loadBalancerArn` | `string \| valueFrom` | yes |  | AwsAlb (`status.outputs.load_balancer_arn`), AwsNlb (`status.outputs.load_balancer_arn`) |
| `spec.port` | `int32` | yes |  |  |
| `spec.protocol` | `string` | yes |  |  |
| `spec.certificateArn` | `string \| valueFrom` |  |  | AwsCertManagerCert (`status.outputs.cert_arn`) |
| `spec.additionalCertificateArns` | `[]string \| valueFrom` |  |  | AwsCertManagerCert (`status.outputs.cert_arn`) |
| `spec.sslPolicy` | `string` |  |  |  |
| `spec.alpnPolicy` | `string` |  |  |  |
| `spec.mutualAuthentication` | `AwsLbListenerMutualAuthentication` |  |  |  |
| `spec.mutualAuthentication.mode` | `string` | yes |  |  |
| `spec.mutualAuthentication.trustStoreArn` | `string \| valueFrom` |  |  |  |
| `spec.mutualAuthentication.ignoreClientCertificateExpiry` | `bool` |  |  |  |
| `spec.mutualAuthentication.advertiseTrustStoreCaNames` | `string` |  |  |  |
| `spec.tcpIdleTimeoutSeconds` | `int32` |  |  |  |
| `spec.defaultActions` | `[]AwsLbListenerAction` | yes |  |  |
| `spec.defaultActions[].type` | `string` | yes |  |  |
| `spec.defaultActions[].order` | `int32` |  |  |  |
| `spec.defaultActions[].forward` | `AwsLbListenerActionForward` |  |  |  |
| `spec.defaultActions[].forward.targetGroups` | `[]AwsLbListenerActionForwardTargetGroup` | yes |  |  |
| `spec.defaultActions[].forward.targetGroups[].arn` | `string \| valueFrom` | yes |  | AwsLbTargetGroup (`status.outputs.target_group_arn`) |
| `spec.defaultActions[].forward.targetGroups[].weight` | `int32` |  |  |  |
| `spec.defaultActions[].forward.stickiness` | `AwsLbListenerActionForwardStickiness` |  |  |  |
| `spec.defaultActions[].forward.stickiness.enabled` | `bool` |  |  |  |
| `spec.defaultActions[].forward.stickiness.durationSeconds` | `int32` |  |  |  |
| `spec.defaultActions[].redirect` | `AwsLbListenerActionRedirect` |  |  |  |
| `spec.defaultActions[].redirect.statusCode` | `string` | yes |  |  |
| `spec.defaultActions[].redirect.protocol` | `string` |  |  |  |
| `spec.defaultActions[].redirect.port` | `string` |  |  |  |
| `spec.defaultActions[].redirect.host` | `string` |  |  |  |
| `spec.defaultActions[].redirect.path` | `string` |  |  |  |
| `spec.defaultActions[].redirect.query` | `string` |  |  |  |
| `spec.defaultActions[].fixedResponse` | `AwsLbListenerActionFixedResponse` |  |  |  |
| `spec.defaultActions[].fixedResponse.contentType` | `string` | yes |  |  |
| `spec.defaultActions[].fixedResponse.statusCode` | `string` |  |  |  |
| `spec.defaultActions[].fixedResponse.messageBody` | `string` |  |  |  |
| `spec.defaultActions[].authenticateCognito` | `AwsLbListenerActionAuthenticateCognito` |  |  |  |
| `spec.defaultActions[].authenticateCognito.userPoolArn` | `string \| valueFrom` | yes |  | AwsCognitoUserPool (`status.outputs.user_pool_arn`) |
| `spec.defaultActions[].authenticateCognito.userPoolClientId` | `string \| valueFrom` | yes |  | AwsCognitoUserPoolClient (`status.outputs.client_id`) |
| `spec.defaultActions[].authenticateCognito.userPoolDomain` | `string \| valueFrom` | yes |  | AwsCognitoUserPool (`status.outputs.user_pool_domain`) |
| `spec.defaultActions[].authenticateCognito.authenticationRequestExtraParams` | `map<string, string>` |  |  |  |
| `spec.defaultActions[].authenticateCognito.onUnauthenticatedRequest` | `string` |  |  |  |
| `spec.defaultActions[].authenticateCognito.scope` | `string` |  |  |  |
| `spec.defaultActions[].authenticateCognito.sessionCookieName` | `string` |  |  |  |
| `spec.defaultActions[].authenticateCognito.sessionTimeoutSeconds` | `int32` |  |  |  |
| `spec.defaultActions[].authenticateOidc` | `AwsLbListenerActionAuthenticateOidc` |  |  |  |
| `spec.defaultActions[].authenticateOidc.issuer` | `string` | yes |  |  |
| `spec.defaultActions[].authenticateOidc.authorizationEndpoint` | `string` | yes |  |  |
| `spec.defaultActions[].authenticateOidc.tokenEndpoint` | `string` | yes |  |  |
| `spec.defaultActions[].authenticateOidc.userInfoEndpoint` | `string` | yes |  |  |
| `spec.defaultActions[].authenticateOidc.clientId` | `string` | yes |  |  |
| `spec.defaultActions[].authenticateOidc.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.defaultActions[].authenticateOidc.authenticationRequestExtraParams` | `map<string, string>` |  |  |  |
| `spec.defaultActions[].authenticateOidc.onUnauthenticatedRequest` | `string` |  |  |  |
| `spec.defaultActions[].authenticateOidc.scope` | `string` |  |  |  |
| `spec.defaultActions[].authenticateOidc.sessionCookieName` | `string` |  |  |  |
| `spec.defaultActions[].authenticateOidc.sessionTimeoutSeconds` | `int32` |  |  |  |
| `spec.defaultActions[].jwtValidation` | `AwsLbListenerActionJwtValidation` |  |  |  |
| `spec.defaultActions[].jwtValidation.issuer` | `string` | yes |  |  |
| `spec.defaultActions[].jwtValidation.jwksEndpoint` | `string` | yes |  |  |
| `spec.defaultActions[].jwtValidation.additionalClaims` | `[]AwsLbListenerActionJwtClaim` |  |  |  |
| `spec.defaultActions[].jwtValidation.additionalClaims[].name` | `string` | yes |  |  |
| `spec.defaultActions[].jwtValidation.additionalClaims[].format` | `string` | yes |  |  |
| `spec.defaultActions[].jwtValidation.additionalClaims[].values` | `[]string` | yes |  |  |
| `spec.httpHeaders` | `AwsLbListenerHttpHeaders` |  |  |  |
| `spec.httpHeaders.request` | `AwsLbListenerHttpRequestHeaders` |  |  |  |
| `spec.httpHeaders.request.mtlsClientcertHeaderName` | `string` |  |  |  |
| `spec.httpHeaders.request.mtlsClientcertIssuerHeaderName` | `string` |  |  |  |
| `spec.httpHeaders.request.mtlsClientcertLeafHeaderName` | `string` |  |  |  |
| `spec.httpHeaders.request.mtlsClientcertSerialNumberHeaderName` | `string` |  |  |  |
| `spec.httpHeaders.request.mtlsClientcertSubjectHeaderName` | `string` |  |  |  |
| `spec.httpHeaders.request.mtlsClientcertValidityHeaderName` | `string` |  |  |  |
| `spec.httpHeaders.request.tlsCipherSuiteHeaderName` | `string` |  |  |  |
| `spec.httpHeaders.request.tlsVersionHeaderName` | `string` |  |  |  |
| `spec.httpHeaders.response` | `AwsLbListenerHttpResponseHeaders` |  |  |  |
| `spec.httpHeaders.response.accessControlAllowCredentials` | `string` |  |  |  |
| `spec.httpHeaders.response.accessControlAllowHeaders` | `string` |  |  |  |
| `spec.httpHeaders.response.accessControlAllowMethods` | `string` |  |  |  |
| `spec.httpHeaders.response.accessControlAllowOrigin` | `string` |  |  |  |
| `spec.httpHeaders.response.accessControlExposeHeaders` | `string` |  |  |  |
| `spec.httpHeaders.response.accessControlMaxAge` | `string` |  |  |  |
| `spec.httpHeaders.response.contentSecurityPolicy` | `string` |  |  |  |
| `spec.httpHeaders.response.serverEnabled` | `bool` |  |  |  |
| `spec.httpHeaders.response.strictTransportSecurity` | `string` |  |  |  |
| `spec.httpHeaders.response.xContentTypeOptions` | `string` |  |  |  |
| `spec.httpHeaders.response.xFrameOptions` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the listener is created. Must match the load
balancer's region. Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.loadBalancerArn

`string | valueFrom` · required

The load balancer this listener attaches to. Defaults to referencing an
AwsAlb's ARN; attach to an AwsNlb with a valueFrom naming that kind (both
are declared candidates, so the ARN output fills in), or to an external
load balancer with a literal ARN. Immutable: changing the load
balancer replaces the listener.

- references: AwsAlb (`status.outputs.load_balancer_arn`), AwsNlb (`status.outputs.load_balancer_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAlb, name: <that resource's name>, fieldPath: status.outputs.load_balancer_arn}} -- a bare string does not parse

### spec.port

`int32` · required

Port the listener accepts traffic on, 1-65535. The classic pairs are 80
(HTTP, usually a redirect listener) and 443 (HTTPS/TLS).

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.protocol

`string` · required

Protocol the listener speaks. Decides the load balancer family and the
allowed actions:
- ALB: "HTTP", "HTTPS" (full action set).
- NLB: "TCP", "UDP", "TCP_UDP", "TLS", "QUIC", "TCP_QUIC" (forward
  only). "QUIC" accepts QUIC connections natively; "TCP_QUIC" serves
  TCP and QUIC on one port (the HTTP/3 pattern with TCP fallback) and
  forwards to a matching QUIC-family target group.

- rule: {"required":true}

### spec.certificateArn

`string | valueFrom`

The default server certificate presented to clients that do not match an
SNI certificate. Required for "HTTPS" and "TLS" listeners; not valid
otherwise.

Requiredness is enforced by the IaC modules rather than a proto rule:
message-level CEL cannot inspect StringValueOrRef fields without breaking
protovalidate-java, so both engines fail fast with a clear error when a
TLS-protocol listener has no certificate.

- references: AwsCertManagerCert (`status.outputs.cert_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.cert_arn}} -- a bare string does not parse

### spec.additionalCertificateArns

`[]string | valueFrom`

Additional SNI certificates for serving several domains from one
listener: the client's SNI hostname picks the matching certificate, and
certificate_arn serves clients that match none. Folded into the listener
(AWS models each as a separate listener-certificate attachment) because
an attachment is pure glue with no referenceable identity of its own.

- references: AwsCertManagerCert (`status.outputs.cert_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.cert_arn}} -- a bare string does not parse

### spec.sslPolicy

`string`

The TLS security policy naming the protocol versions and cipher suites
the listener negotiates. Only for "HTTPS" and "TLS" listeners; AWS
defaults it when omitted.
Recommended: "ELBSecurityPolicy-TLS13-1-2-2021-06" (TLS 1.3 + 1.2).

### spec.alpnPolicy

`string`

ALPN policy for NLB "TLS" listeners: which application protocols are
advertised during the TLS handshake. Valid values: "HTTP1Only",
"HTTP2Only", "HTTP2Optional", "HTTP2Preferred", "None".

### spec.mutualAuthentication

`AwsLbListenerMutualAuthentication`

Mutual TLS: require or accept client certificates on an ALB "HTTPS"
listener. When omitted, mTLS is off.

- rule: mode must be 'off', 'verify', or 'passthrough'
- rule: ignore_client_certificate_expiry only applies when mode is 'verify'
- rule: advertise_trust_store_ca_names must be 'on' or 'off' when set
- rule: advertise_trust_store_ca_names only applies when mode is 'verify'

### spec.mutualAuthentication.mode

`string` · required

The mTLS mode. Required.
- "off": no client certificates (the AWS default without this block).
- "passthrough": the ALB forwards the whole client certificate to the
  target in the X-Amzn-Mtls-Clientcert header and does no verification
  itself.
- "verify": the ALB verifies client certificates against the trust
  store before forwarding.

- rule: {"required":true}

### spec.mutualAuthentication.trustStoreArn

`string | valueFrom`

The ELBv2 trust store holding the CA bundle used for verification.
Required when mode is "verify"; not valid otherwise. Accepts a literal
trust store ARN or a reference (no default kind -- trust stores are not
modeled as a Planton kind).

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.mutualAuthentication.ignoreClientCertificateExpiry

`bool`

"verify" mode only: accept client certificates that are past their
expiry date. For controlled rollouts of certificate rotation, not
steady-state use.

### spec.mutualAuthentication.advertiseTrustStoreCaNames

`string`

"verify" mode only: whether the ALB advertises the trust store's CA
subject names during the TLS handshake, helping clients pick the right
certificate. Valid values: "on", "off".

### spec.tcpIdleTimeoutSeconds

`int32`

Seconds an idle TCP connection stays open on an NLB "TCP" listener
before the NLB closes it. Range 60-6000. AWS default: 350. Raise it for
protocols with long quiet periods (database wire protocols, MQTT).

### spec.defaultActions

`[]AwsLbListenerAction` · required

What happens to a request no listener rule matches. Required, at least
one action. With several actions (an authenticate action chained before
its forward), order follows the list; every chain ends in exactly one of
forward, redirect, or fixed-response.

NLB-protocol listeners must use exactly one "forward" action -- AWS
rejects everything else at Layer 4.

- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: type must be one of: forward, redirect, fixed-response, authenticate-cognito, authenticate-oidc, jwt-validation
- rule: order must be between 1 and 50000 when set
- rule: forward configuration must be set when (and only when) type is 'forward'
- rule: redirect configuration must be set when (and only when) type is 'redirect'
- rule: fixed_response configuration must be set when (and only when) type is 'fixed-response'
- rule: authenticate_cognito configuration must be set when (and only when) type is 'authenticate-cognito'
- rule: authenticate_oidc configuration must be set when (and only when) type is 'authenticate-oidc'
- rule: jwt_validation configuration must be set when (and only when) type is 'jwt-validation'

### spec.defaultActions[].type

`string` · required

The action type. Valid values: "forward", "redirect", "fixed-response",
"authenticate-cognito", "authenticate-oidc", "jwt-validation".
NLB-protocol listeners accept only "forward".

- rule: {"required":true}

### spec.defaultActions[].order

`int32`

Explicit evaluation order within the chain, 1-50000 (lower runs first).
When omitted, list position decides -- which is almost always enough.

### spec.defaultActions[].forward

`AwsLbListenerActionForward`

Forward the request to one or more weighted target groups.

### spec.defaultActions[].forward.targetGroups

`[]AwsLbListenerActionForwardTargetGroup` · required

The destination target groups, 1-5, each with an optional weight.

- rule: {"required":true,"repeated":{"minItems":"1","maxItems":"5"}}
- rule: weight must be between 0 and 999

### spec.defaultActions[].forward.targetGroups[].arn

`string | valueFrom` · required

The target group receiving the traffic.

- references: AwsLbTargetGroup (`status.outputs.target_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLbTargetGroup, name: <that resource's name>, fieldPath: status.outputs.target_group_arn}} -- a bare string does not parse

### spec.defaultActions[].forward.targetGroups[].weight

`int32`

Relative traffic share, 0-999 (proportional to the sum across groups;
0 drains the group). AWS default: 1, so equal weights need no
configuration.

### spec.defaultActions[].forward.stickiness

`AwsLbListenerActionForwardStickiness`

Target-group stickiness for multi-group forwards: pins a client to the
group (not the individual target) that served its first request, so a
session does not flap between blue and green mid-canary.

- rule: duration_seconds is required when stickiness is enabled
- rule: duration_seconds must be between 1 and 604800 when set

### spec.defaultActions[].forward.stickiness.enabled

`bool`

Whether group-level stickiness is active.

### spec.defaultActions[].forward.stickiness.durationSeconds

`int32`

Seconds the group association lasts, 1-604800 (7 days). Required when
enabled.

### spec.defaultActions[].redirect

`AwsLbListenerActionRedirect`

Reply with an HTTP redirect. ALB only.

- rule: status_code must be 'HTTP_301' or 'HTTP_302'
- rule: protocol must be 'HTTP', 'HTTPS', or '#{protocol}' when set

### spec.defaultActions[].redirect.statusCode

`string` · required

The redirect status code. Required. Valid values: "HTTP_301" (permanent
-- browsers cache it) or "HTTP_302" (temporary).

- rule: {"required":true}

### spec.defaultActions[].redirect.protocol

`string`

Target protocol: "HTTPS", "HTTP", or "#{protocol}" (keep the original).

### spec.defaultActions[].redirect.port

`string`

Target port: a port number as a string, or "#{port}" (keep the
original).

### spec.defaultActions[].redirect.host

`string`

Target hostname, 1-128 characters; supports placeholders (e.g.
"#{host}").

### spec.defaultActions[].redirect.path

`string`

Target absolute path, starting with "/"; supports placeholders (e.g.
"/#{path}" or "/new/#{path}").

### spec.defaultActions[].redirect.query

`string`

Target query string, without the leading "?"; supports placeholders.

### spec.defaultActions[].fixedResponse

`AwsLbListenerActionFixedResponse`

Reply with a canned HTTP response, never touching a target. ALB only.
The classic default action for a listener whose real traffic all flows
through listener rules: unmatched requests get an explicit 404.

- rule: content_type must be one of: text/plain, text/css, text/html, application/javascript, application/json
- rule: status_code must be a 2xx, 4xx, or 5xx code when set

### spec.defaultActions[].fixedResponse.contentType

`string` · required

The Content-Type of the response. Required. Valid values: "text/plain",
"text/css", "text/html", "application/javascript", "application/json".

- rule: {"required":true}

### spec.defaultActions[].fixedResponse.statusCode

`string`

The HTTP status code, as a string: 2xx, 4xx, or 5xx (e.g. "404").
AWS default: "503".

### spec.defaultActions[].fixedResponse.messageBody

`string`

The response body, up to 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.defaultActions[].authenticateCognito

`AwsLbListenerActionAuthenticateCognito`

Authenticate the client against an Amazon Cognito user pool before the
terminal action runs. ALB HTTPS only.

- rule: on_unauthenticated_request must be 'deny', 'allow', or 'authenticate' when set

### spec.defaultActions[].authenticateCognito.userPoolArn

`string | valueFrom` · required

The Cognito user pool performing authentication.

- references: AwsCognitoUserPool (`status.outputs.user_pool_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_arn}} -- a bare string does not parse

### spec.defaultActions[].authenticateCognito.userPoolClientId

`string | valueFrom` · required

The app client ID within the user pool. Accepts a direct client ID or a
reference to an AwsCognitoUserPoolClient resource.

- references: AwsCognitoUserPoolClient (`status.outputs.client_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPoolClient, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.defaultActions[].authenticateCognito.userPoolDomain

`string | valueFrom` · required

The user pool's hosted-UI domain: the prefix of a Cognito domain (e.g.
"my-app" for my-app.auth.us-west-2.amazoncognito.com) or a full custom
domain. Accepts a direct domain string or a reference to the pool's
user_pool_domain output.

- references: AwsCognitoUserPool (`status.outputs.user_pool_domain`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_domain}} -- a bare string does not parse

### spec.defaultActions[].authenticateCognito.authenticationRequestExtraParams

`map<string, string>`

Extra query parameters appended to the authorization-endpoint request.

### spec.defaultActions[].authenticateCognito.onUnauthenticatedRequest

`string`

What happens to an unauthenticated request: "authenticate" (AWS default
-- redirect to login), "deny" (401), or "allow" (pass through; the app
sees the request without identity headers).

### spec.defaultActions[].authenticateCognito.scope

`string`

Space-separated OAuth scopes to request. AWS default: "openid".

### spec.defaultActions[].authenticateCognito.sessionCookieName

`string`

The session cookie name. AWS default: "AWSELBAuthSessionCookie".

### spec.defaultActions[].authenticateCognito.sessionTimeoutSeconds

`int32`

Seconds the authentication session stays valid. AWS default: 604800
(7 days).

### spec.defaultActions[].authenticateOidc

`AwsLbListenerActionAuthenticateOidc`

Authenticate the client against any OpenID Connect provider before the
terminal action runs. ALB HTTPS only.

- rule: on_unauthenticated_request must be 'deny', 'allow', or 'authenticate' when set

### spec.defaultActions[].authenticateOidc.issuer

`string` · required

The OIDC issuer identifier (e.g. "https://accounts.google.com").

- rule: {"required":true}

### spec.defaultActions[].authenticateOidc.authorizationEndpoint

`string` · required

The provider's authorization endpoint URL.

- rule: {"required":true}

### spec.defaultActions[].authenticateOidc.tokenEndpoint

`string` · required

The provider's token endpoint URL.

- rule: {"required":true}

### spec.defaultActions[].authenticateOidc.userInfoEndpoint

`string` · required

The provider's UserInfo endpoint URL.

- rule: {"required":true}

### spec.defaultActions[].authenticateOidc.clientId

`string` · required

The OAuth client ID registered with the provider.

- rule: {"required":true}

### spec.defaultActions[].authenticateOidc.clientSecret

`string` · required · sensitive

The OAuth client secret. Handled as a secret end to end -- never exposed
in plans, logs, or the console.

- rule: {"required":true}

### spec.defaultActions[].authenticateOidc.authenticationRequestExtraParams

`map<string, string>`

Extra query parameters appended to the authorization-endpoint request.

### spec.defaultActions[].authenticateOidc.onUnauthenticatedRequest

`string`

What happens to an unauthenticated request: "authenticate" (AWS default
-- redirect to login), "deny" (401), or "allow" (pass through).

### spec.defaultActions[].authenticateOidc.scope

`string`

Space-separated OAuth scopes to request. AWS default: "openid".

### spec.defaultActions[].authenticateOidc.sessionCookieName

`string`

The session cookie name. AWS default: "AWSELBAuthSessionCookie".

### spec.defaultActions[].authenticateOidc.sessionTimeoutSeconds

`int32`

Seconds the authentication session stays valid. AWS default: 604800
(7 days).

### spec.defaultActions[].jwtValidation

`AwsLbListenerActionJwtValidation`

Validate a JWT bearer token on the request before the terminal action
runs -- API-style auth with no session cookies or login redirects.
ALB HTTPS only.

### spec.defaultActions[].jwtValidation.issuer

`string` · required

The expected token issuer (the "iss" claim), 1-256 characters.

- rule: {"required":true,"string":{"maxLen":"256"}}

### spec.defaultActions[].jwtValidation.jwksEndpoint

`string` · required

The JWKS endpoint URL serving the signing keys tokens are verified
against, 1-256 characters.

- rule: {"required":true,"string":{"maxLen":"256"}}

### spec.defaultActions[].jwtValidation.additionalClaims

`[]AwsLbListenerActionJwtClaim`

Extra claims the token must carry, beyond issuer and signature. Up to 10.

- rule: {"repeated":{"maxItems":"10"}}
- rule: format must be 'single-string', 'string-array', or 'space-separated-values'

### spec.defaultActions[].jwtValidation.additionalClaims[].name

`string` · required

The claim name in the token payload (e.g. "aud", "scope").

- rule: {"required":true}

### spec.defaultActions[].jwtValidation.additionalClaims[].format

`string` · required

How the claim value is encoded in the token. Valid values:
"single-string", "string-array", "space-separated-values" (how OAuth
"scope" claims are typically encoded).

- rule: {"required":true}

### spec.defaultActions[].jwtValidation.additionalClaims[].values

`[]string` · required

Accepted values for the claim, 1-10 entries of up to 256 characters.
The claim satisfies validation when any listed value matches.

- rule: {"required":true,"repeated":{"minItems":"1","maxItems":"10","items":{"string":{"maxLen":"256"}}}}

### spec.httpHeaders

`AwsLbListenerHttpHeaders`

HTTP header handling on ALB listeners: inject TLS/mTLS details as
request headers, and set response headers (CORS, security headers).
When omitted, no headers are injected or overridden.

### spec.httpHeaders.request

`AwsLbListenerHttpRequestHeaders`

Request headers the ALB injects toward targets (HTTPS listeners only).

### spec.httpHeaders.request.mtlsClientcertHeaderName

`string`

Header carrying the full client certificate (URL-encoded PEM).

### spec.httpHeaders.request.mtlsClientcertIssuerHeaderName

`string`

Header carrying the client certificate's issuer DN.

### spec.httpHeaders.request.mtlsClientcertLeafHeaderName

`string`

Header carrying the leaf client certificate (URL-encoded PEM).

### spec.httpHeaders.request.mtlsClientcertSerialNumberHeaderName

`string`

Header carrying the client certificate's serial number.

### spec.httpHeaders.request.mtlsClientcertSubjectHeaderName

`string`

Header carrying the client certificate's subject DN.

### spec.httpHeaders.request.mtlsClientcertValidityHeaderName

`string`

Header carrying the client certificate's validity period.

### spec.httpHeaders.request.tlsCipherSuiteHeaderName

`string`

Header carrying the negotiated TLS cipher suite.

### spec.httpHeaders.request.tlsVersionHeaderName

`string`

Header carrying the negotiated TLS protocol version.

### spec.httpHeaders.response

`AwsLbListenerHttpResponseHeaders`

Response headers the ALB sets toward clients (HTTP and HTTPS).

### spec.httpHeaders.response.accessControlAllowCredentials

`string`

Access-Control-Allow-Credentials header value.

### spec.httpHeaders.response.accessControlAllowHeaders

`string`

Access-Control-Allow-Headers header value.

### spec.httpHeaders.response.accessControlAllowMethods

`string`

Access-Control-Allow-Methods header value.

### spec.httpHeaders.response.accessControlAllowOrigin

`string`

Access-Control-Allow-Origin header value.

### spec.httpHeaders.response.accessControlExposeHeaders

`string`

Access-Control-Expose-Headers header value.

### spec.httpHeaders.response.accessControlMaxAge

`string`

Access-Control-Max-Age header value (seconds, as a string).

### spec.httpHeaders.response.contentSecurityPolicy

`string`

Content-Security-Policy header value.

### spec.httpHeaders.response.serverEnabled

`bool` · optional (explicit presence)

Whether the ALB includes its "Server" header in responses. AWS default:
true. Optional rather than plain bool so that false ("strip the
header") is distinguishable from unset ("keep the AWS default").

### spec.httpHeaders.response.strictTransportSecurity

`string`

Strict-Transport-Security header value (e.g. "max-age=31536000;
includeSubDomains").

### spec.httpHeaders.response.xContentTypeOptions

`string`

X-Content-Type-Options header value (typically "nosniff").

### spec.httpHeaders.response.xFrameOptions

`string`

X-Frame-Options header value (e.g. "DENY", "SAMEORIGIN").

## Validation Rules

- `protocol_valid`: protocol must be one of: HTTP, HTTPS, TCP, UDP, TCP_UDP, TLS, QUIC, TCP_QUIC
- `ssl_policy_only_for_tls_protocols`: ssl_policy only applies when protocol is 'HTTPS' or 'TLS'
- `alpn_policy_only_for_nlb_tls`: alpn_policy only applies when protocol is 'TLS'
- `alpn_policy_valid`: alpn_policy must be one of: HTTP1Only, HTTP2Only, HTTP2Optional, HTTP2Preferred, None
- `mutual_authentication_only_for_https`: mutual_authentication only applies when protocol is 'HTTPS'
- `tcp_idle_timeout_only_for_tcp`: tcp_idle_timeout_seconds only applies when protocol is 'TCP'
- `tcp_idle_timeout_range`: tcp_idle_timeout_seconds must be between 60 and 6000 when set
- `http_headers_only_for_alb_protocols`: http_headers only applies when protocol is 'HTTP' or 'HTTPS'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsLbListener, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.listener_arn` | `string` | The ARN of the listener (e.g. "arn:aws:elasticloadbalancing:us-west-2: 123456789012:listener/app/api/50dc6c495c0c9188/f2f7dc8efc522ab2"). The primary handle other resources reference via status.outputs.listener_arn -- listener rules attach through this value. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.loadBalancerArn` | AwsAlb | `status.outputs.load_balancer_arn` |
| `spec.loadBalancerArn` | AwsNlb | `status.outputs.load_balancer_arn` |
| `spec.certificateArn` | AwsCertManagerCert | `status.outputs.cert_arn` |
| `spec.additionalCertificateArns` | AwsCertManagerCert | `status.outputs.cert_arn` |
| `spec.defaultActions[].forward.targetGroups[].arn` | AwsLbTargetGroup | `status.outputs.target_group_arn` |
| `spec.defaultActions[].authenticateCognito.userPoolArn` | AwsCognitoUserPool | `status.outputs.user_pool_arn` |
| `spec.defaultActions[].authenticateCognito.userPoolClientId` | AwsCognitoUserPoolClient | `status.outputs.client_id` |
| `spec.defaultActions[].authenticateCognito.userPoolDomain` | AwsCognitoUserPool | `status.outputs.user_pool_domain` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsLbListenerRule | `spec.listenerArn` | `status.outputs.listener_arn` |

## See Also

- [Overview](../README.md)
