# Auth0 Resource Server Guide

An Auth0ResourceServer is the API an access token is minted for. The judgment this guide protects is who gets such a token: the access policy (`subjectTypeAuthorization`), the per-application grants on Auth0Client, and the default grants third-party applications get from this kind are three halves of one decision, and choosing one without the others either locks your own applications out or lets every application in.

## When to use it (and when not)

Use one Auth0ResourceServer per audience your backends validate. Its `identifier` is permanent; a rename is a new API, and every grant, role permission and client configuration naming the old audience has to move with it.

**Scopes belong here, grants belong on the consumer.** `scopes` is authoritative once it has entries: removing an entry removes the scope from the API and from every grant and role built on it. The grant for one application lives on that application (Auth0Client `apiGrants`), so the application's node on the diagram carries the edge to this API. The one grant that lives here is the default grant for third-party applications, because it names no application.

**Default grants are for applications you do not register yourself.** Partners, AI agents and MCP clients register through Dynamic Client Registration or from a Client ID Metadata Document; nobody can declare an Auth0Client for each. `thirdPartyClientDefaultGrants` gives every one of them the same ceiling on this API, one grant per subject type. For a third-party application you do register, declare its own grant on its Auth0Client: a grant made for one application takes precedence over the default, to widen or narrow that one client.

## The access policy

Auth0 decides user-delegated access (an application acting for a signed-in person) and client access (an application acting for itself, through client credentials) separately:

| Policy | User-delegated (`user`) | Client (`client`) |
|---|---|---|
| `allow_all` | Every first-party application, no grant needed (Auth0's default for a new API) | Not offered |
| `require_client_grant` | Only applications holding a `user` grant; the grant caps their scopes | Only applications holding a `client` grant (Auth0's default for a new API) |
| `deny_all` | No application | No application |

Third-party applications need a grant under every policy, and `deny_all` denies them too. Two consequences worth designing for:

- **Moving `user` to `require_client_grant` locks out your own applications** until each has a `user` grant (Auth0Client `apiGrants` with subject type `user`). Deploy the grants with, or before, the policy. Under this policy an application must also name the scopes it wants in every token request (a refresh-token request excepted, which keeps the originally granted scopes).
- **`deny_all` on `client` is how an API says "people only".** An MCP server, or any API whose tools act for a person, should refuse client-credentials tokens outright rather than rely on nobody holding a client grant.

## Adopting an existing API

Import the API (and its scopes and default grants) with the ids the import map names, then declare the spec. Unset means unmanaged: a field or block the spec leaves out is never sent, and the live value stays -- with these exceptions to check against the live API before the first apply:

- **Kept when unset**: `allowOfflineAccess`, `skipConsentForVerifiableFirstPartyClients` and `enforcePolicies` are sent only when declared, and Auth0 computes them otherwise, so an adopted API that leaves them out keeps its live values. The same holds for `tokenLifetime` and `tokenLifetimeForWeb`, whose zero means unset.
- **No value of their own in the provider**: `verificationLocation`, `tokenLifetimeForAnonymousAccessTokens`, `accessToken`, and `proofOfPossession.requiredFor` once its block is declared. Leaving one unset on an API that has it resets it on the first apply.
- **Authoritative lists**: declared `scopes` replace the API's scope list, and a declared `accessToken.claimsMapping` replaces its claims.
- **Access policies are safe to leave out.** Each policy block is sent only with its policy, and a policy never declared keeps its live value -- but a policy you do declare overwrites the live one, so read it first (`GET /api/v2/resource-servers/{id}`, `subject_type_authorization`).
- **Default grants adopt by import, not by create.** Auth0 keeps one default grant per API and subject type and refuses a second, so a live default grant must be imported (its `cgr_...` id, keyed by subject type) before an entry for that subject type is applied.

## Plan boundaries

Defining an API, its scopes, its access policy and its default grants works on every plan. These settings need more, per Auth0's documentation:

| Setting | Needs |
|---|---|
| `tokenEncryption` | The Enterprise plan with the Highly Regulated Identity add-on |
| `proofOfPossession.mechanism: mtls` | The Enterprise plan with the Highly Regulated Identity add-on |
| `consentPolicy: transactional-authorization-with-mfa` | The Enterprise plan with the Highly Regulated Identity add-on |
| `authorizationDetails` | Registering the types has no plan statement; clients use them through Pushed Authorization Requests (Enterprise plan with the Highly Regulated Identity add-on) or Client-Initiated Backchannel Authentication (Enterprise plan or an add-on) |
| `tokenLifetimeForAnonymousAccessTokens`, `accessToken.claimsMapping`, `subjectTypeAuthorization.anonymousUser` | Anonymous sessions, an Enterprise Early Access feature |
| `allowOnlineAccess`, `allowOnlineAccessWithEphemeralSessions` | Online Refresh Tokens, in Beta |
| `authorizationPolicy` | Early Access |

DPoP (`proofOfPossession.mechanism: dpop`) carries no plan statement in Auth0's documentation.

## Conventions and gotchas

- **Sender constraining is two-sided.** An application that requires proof of possession gets a token only for an API that accepts it, and an API with `required: true` refuses tokens that are not bound. Turn the API on as "allowed" (`required: false`) first, move the applications, then require it.
- **Encryption keys are public.** `tokenEncryption.encryptionKey.pem` is the API's public key; the private key never enters a manifest. The signing secret of an HS256 API (`signingSecret`) is the opposite: whoever holds it can mint tokens.
- **System APIs take no default grants.** The Management API and My Account API refuse `thirdPartyClientDefaultGrants`, and third-party applications can never be granted them.

## On the diagram

The API renders as one node carrying its scopes, policy and default grants. Each application's grant renders as an edge from that Auth0Client to this node; the default grants render on this node alone, since they name no application -- which is exactly what they mean.

## Pairs well with

- **Auth0Client** -- the applications that call the API; each `apiGrants` entry references this API's `identifier` output.
- **Auth0Role** -- groups this API's scopes into assignable permissions for `enforcePolicies`.
- **Auth0TenantSettings** and **Auth0Connection** -- the tenant-level half of opening an API to third-party applications: dynamic client registration and connections promoted to the domain level.

Presets (`api-with-scopes`, `rbac-api`, `mcp-server-api`) ship in the release's `presets.zip` and the repository, not in the reference pack.

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Auth0's published certifications and security standards:

- SOC 2 Type II (annual audit)
- ISO 27001, ISO 27018 (privacy controls)
- HIPAA BAA available on enterprise plans
- PCI DSS Level 1 Service Provider
- FedRAMP Authorized (moderate baseline)
- CSA STAR Level 2
- GDPR compliant with Data Processing Agreement

## Resource Server-Specific Security Notes

### Signing Algorithms

- **RS256 (recommended)**: the API validates tokens against the tenant's public keys; nothing secret is shared, and keys rotate without coordination.
- **HS256**: the API holds the shared `signing_secret`, which can also mint tokens. Treat it as a credential, and set `signingAlg` explicitly: Auth0's API reference and its dashboard disagree on the default for an API created without one.

### Token Validation

APIs validate every access token: the signature, the issuer (`iss`, the tenant's domain or its custom domain), the audience (`aud`, this API's identifier) and the expiry. An encrypted token is decrypted with the API's private key first; a sender-constrained one is checked against the proof the caller presents.

## Permissions
## Management API Scopes

| Operation | Scope | When |
|-----------|-------|------|
| Read | `read:resource_servers` | Always |
| Create | `create:resource_servers` | Always |
| Update | `update:resource_servers` | Always (the scopes list rides it too) |
| Delete | `delete:resource_servers` | Always |
| Default grants | `create:client_grants`, `read:client_grants`, `update:client_grants`, `delete:client_grants` | When `thirdPartyClientDefaultGrants` is declared |

## Compliance
## Resource Server-Specific Compliance Notes

### Access Policy as Auditable Configuration

Which applications may get a token for the API, and what every third-party application may do by default, are version-controlled with the API itself, so opening an API to third parties is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

### Access Token Content

Access tokens may carry user identifiers, granted scopes, permissions and mapped claims. Include only the claims the API needs; `tokenEncryption` keeps them unreadable to everyone but the API where regulation requires it.

### API Identifier Stability

The identifier (audience) is immutable after creation. Changing it requires a new API, which matters for compliance documentation that references specific audiences.

## Cost
## Pricing Model

Auth0 bills by plan and monthly active users, not by API: defining APIs, scopes, access policies and default grants carries no per-object charge. Machine-to-machine tokens issued for the API count against the tenant's monthly machine-to-machine token allotment. The settings in the plan-boundaries table above need a plan or add-on, not a charge per API.
