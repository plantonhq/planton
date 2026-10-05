# CloudflareZeroTrustAccessPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustAccessPolicySpec defines a reusable Cloudflare Zero Trust
Access policy: a named, account-scoped decision (allow / deny / non-identity /
bypass) plus the include/exclude/require rules that determine who it applies
to. Policies are standalone and referenced by one or more Access applications
(by ID, with a precedence) — so the same policy can guard many applications and
its rules evolve in one place.

A policy is always account-scoped (account_id is required); zone-scoped access
is expressed by attaching the policy to a zone-scoped application.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessPolicy
metadata:
  name: test-access-policy
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: allow-staff
  decision: allow
  sessionDuration: 24h
  include:
    - emailDomain:
        domain: example.com
  require:
    - geo:
        countryCode: US
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.decision` | `enum` | yes |  |  |
| `spec.include` | `[]CloudflareAccessRule` | yes |  |  |
| `spec.include[].email` | `AccessRuleEmail` |  |  |  |
| `spec.include[].email.email` | `string` | yes |  |  |
| `spec.include[].emailDomain` | `AccessRuleEmailDomain` |  |  |  |
| `spec.include[].emailDomain.domain` | `string` | yes |  |  |
| `spec.include[].emailList` | `AccessRuleListRef` |  |  |  |
| `spec.include[].emailList.id` | `string \| valueFrom` | yes |  |  |
| `spec.include[].everyone` | `AccessRuleEmpty` |  |  |  |
| `spec.include[].ip` | `AccessRuleIp` |  |  |  |
| `spec.include[].ip.ip` | `string` | yes |  |  |
| `spec.include[].ipList` | `AccessRuleListRef` |  |  |  |
| `spec.include[].ipList.id` | `string \| valueFrom` | yes |  |  |
| `spec.include[].certificate` | `AccessRuleEmpty` |  |  |  |
| `spec.include[].group` | `AccessRuleGroupRef` |  |  |  |
| `spec.include[].group.id` | `string \| valueFrom` | yes |  | CloudflareZeroTrustAccessGroup (`status.outputs.group_id`) |
| `spec.include[].azureAd` | `AccessRuleAzureAd` |  |  |  |
| `spec.include[].azureAd.id` | `string` | yes |  |  |
| `spec.include[].azureAd.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.include[].githubOrganization` | `AccessRuleGithubOrganization` |  |  |  |
| `spec.include[].githubOrganization.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.include[].githubOrganization.name` | `string` | yes |  |  |
| `spec.include[].githubOrganization.team` | `string` |  |  |  |
| `spec.include[].gsuite` | `AccessRuleGsuite` |  |  |  |
| `spec.include[].gsuite.email` | `string` | yes |  |  |
| `spec.include[].gsuite.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.include[].okta` | `AccessRuleOkta` |  |  |  |
| `spec.include[].okta.name` | `string` | yes |  |  |
| `spec.include[].okta.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.include[].saml` | `AccessRuleSaml` |  |  |  |
| `spec.include[].saml.attributeName` | `string` | yes |  |  |
| `spec.include[].saml.attributeValue` | `string` | yes |  |  |
| `spec.include[].saml.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.include[].oidc` | `AccessRuleOidc` |  |  |  |
| `spec.include[].oidc.claimName` | `string` | yes |  |  |
| `spec.include[].oidc.claimValue` | `string` | yes |  |  |
| `spec.include[].oidc.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.include[].authContext` | `AccessRuleAuthContext` |  |  |  |
| `spec.include[].authContext.id` | `string` | yes |  |  |
| `spec.include[].authContext.acId` | `string` | yes |  |  |
| `spec.include[].authContext.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.include[].authMethod` | `AccessRuleAuthMethod` |  |  |  |
| `spec.include[].authMethod.authMethod` | `string` | yes |  |  |
| `spec.include[].commonName` | `AccessRuleCommonName` |  |  |  |
| `spec.include[].commonName.commonName` | `string` | yes |  |  |
| `spec.include[].geo` | `AccessRuleGeo` |  |  |  |
| `spec.include[].geo.countryCode` | `string` | yes |  |  |
| `spec.include[].devicePosture` | `AccessRuleDevicePosture` |  |  |  |
| `spec.include[].devicePosture.integrationUid` | `string \| valueFrom` | yes |  |  |
| `spec.include[].externalEvaluation` | `AccessRuleExternalEvaluation` |  |  |  |
| `spec.include[].externalEvaluation.evaluateUrl` | `string` | yes |  |  |
| `spec.include[].externalEvaluation.keysUrl` | `string` | yes |  |  |
| `spec.include[].loginMethod` | `AccessRuleLoginMethod` |  |  |  |
| `spec.include[].loginMethod.id` | `string \| valueFrom` | yes |  |  |
| `spec.include[].serviceToken` | `AccessRuleServiceToken` |  |  |  |
| `spec.include[].serviceToken.tokenId` | `string \| valueFrom` | yes |  |  |
| `spec.include[].anyValidServiceToken` | `AccessRuleEmpty` |  |  |  |
| `spec.include[].linkedAppToken` | `AccessRuleLinkedAppToken` |  |  |  |
| `spec.include[].linkedAppToken.appUid` | `string \| valueFrom` | yes |  |  |
| `spec.include[].userRiskScore` | `AccessRuleUserRiskScore` |  |  |  |
| `spec.include[].userRiskScore.userRiskScore` | `[]enum` | yes |  |  |
| `spec.include[].cloudflareAccountMember` | `AccessRuleCloudflareAccountMember` |  |  |  |
| `spec.include[].cloudflareAccountMember.accountId` | `string` |  |  |  |
| `spec.exclude` | `[]CloudflareAccessRule` |  |  |  |
| `spec.exclude[].email` | `AccessRuleEmail` |  |  |  |
| `spec.exclude[].email.email` | `string` | yes |  |  |
| `spec.exclude[].emailDomain` | `AccessRuleEmailDomain` |  |  |  |
| `spec.exclude[].emailDomain.domain` | `string` | yes |  |  |
| `spec.exclude[].emailList` | `AccessRuleListRef` |  |  |  |
| `spec.exclude[].emailList.id` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].everyone` | `AccessRuleEmpty` |  |  |  |
| `spec.exclude[].ip` | `AccessRuleIp` |  |  |  |
| `spec.exclude[].ip.ip` | `string` | yes |  |  |
| `spec.exclude[].ipList` | `AccessRuleListRef` |  |  |  |
| `spec.exclude[].ipList.id` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].certificate` | `AccessRuleEmpty` |  |  |  |
| `spec.exclude[].group` | `AccessRuleGroupRef` |  |  |  |
| `spec.exclude[].group.id` | `string \| valueFrom` | yes |  | CloudflareZeroTrustAccessGroup (`status.outputs.group_id`) |
| `spec.exclude[].azureAd` | `AccessRuleAzureAd` |  |  |  |
| `spec.exclude[].azureAd.id` | `string` | yes |  |  |
| `spec.exclude[].azureAd.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].githubOrganization` | `AccessRuleGithubOrganization` |  |  |  |
| `spec.exclude[].githubOrganization.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].githubOrganization.name` | `string` | yes |  |  |
| `spec.exclude[].githubOrganization.team` | `string` |  |  |  |
| `spec.exclude[].gsuite` | `AccessRuleGsuite` |  |  |  |
| `spec.exclude[].gsuite.email` | `string` | yes |  |  |
| `spec.exclude[].gsuite.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].okta` | `AccessRuleOkta` |  |  |  |
| `spec.exclude[].okta.name` | `string` | yes |  |  |
| `spec.exclude[].okta.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].saml` | `AccessRuleSaml` |  |  |  |
| `spec.exclude[].saml.attributeName` | `string` | yes |  |  |
| `spec.exclude[].saml.attributeValue` | `string` | yes |  |  |
| `spec.exclude[].saml.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].oidc` | `AccessRuleOidc` |  |  |  |
| `spec.exclude[].oidc.claimName` | `string` | yes |  |  |
| `spec.exclude[].oidc.claimValue` | `string` | yes |  |  |
| `spec.exclude[].oidc.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].authContext` | `AccessRuleAuthContext` |  |  |  |
| `spec.exclude[].authContext.id` | `string` | yes |  |  |
| `spec.exclude[].authContext.acId` | `string` | yes |  |  |
| `spec.exclude[].authContext.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].authMethod` | `AccessRuleAuthMethod` |  |  |  |
| `spec.exclude[].authMethod.authMethod` | `string` | yes |  |  |
| `spec.exclude[].commonName` | `AccessRuleCommonName` |  |  |  |
| `spec.exclude[].commonName.commonName` | `string` | yes |  |  |
| `spec.exclude[].geo` | `AccessRuleGeo` |  |  |  |
| `spec.exclude[].geo.countryCode` | `string` | yes |  |  |
| `spec.exclude[].devicePosture` | `AccessRuleDevicePosture` |  |  |  |
| `spec.exclude[].devicePosture.integrationUid` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].externalEvaluation` | `AccessRuleExternalEvaluation` |  |  |  |
| `spec.exclude[].externalEvaluation.evaluateUrl` | `string` | yes |  |  |
| `spec.exclude[].externalEvaluation.keysUrl` | `string` | yes |  |  |
| `spec.exclude[].loginMethod` | `AccessRuleLoginMethod` |  |  |  |
| `spec.exclude[].loginMethod.id` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].serviceToken` | `AccessRuleServiceToken` |  |  |  |
| `spec.exclude[].serviceToken.tokenId` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].anyValidServiceToken` | `AccessRuleEmpty` |  |  |  |
| `spec.exclude[].linkedAppToken` | `AccessRuleLinkedAppToken` |  |  |  |
| `spec.exclude[].linkedAppToken.appUid` | `string \| valueFrom` | yes |  |  |
| `spec.exclude[].userRiskScore` | `AccessRuleUserRiskScore` |  |  |  |
| `spec.exclude[].userRiskScore.userRiskScore` | `[]enum` | yes |  |  |
| `spec.exclude[].cloudflareAccountMember` | `AccessRuleCloudflareAccountMember` |  |  |  |
| `spec.exclude[].cloudflareAccountMember.accountId` | `string` |  |  |  |
| `spec.require` | `[]CloudflareAccessRule` |  |  |  |
| `spec.require[].email` | `AccessRuleEmail` |  |  |  |
| `spec.require[].email.email` | `string` | yes |  |  |
| `spec.require[].emailDomain` | `AccessRuleEmailDomain` |  |  |  |
| `spec.require[].emailDomain.domain` | `string` | yes |  |  |
| `spec.require[].emailList` | `AccessRuleListRef` |  |  |  |
| `spec.require[].emailList.id` | `string \| valueFrom` | yes |  |  |
| `spec.require[].everyone` | `AccessRuleEmpty` |  |  |  |
| `spec.require[].ip` | `AccessRuleIp` |  |  |  |
| `spec.require[].ip.ip` | `string` | yes |  |  |
| `spec.require[].ipList` | `AccessRuleListRef` |  |  |  |
| `spec.require[].ipList.id` | `string \| valueFrom` | yes |  |  |
| `spec.require[].certificate` | `AccessRuleEmpty` |  |  |  |
| `spec.require[].group` | `AccessRuleGroupRef` |  |  |  |
| `spec.require[].group.id` | `string \| valueFrom` | yes |  | CloudflareZeroTrustAccessGroup (`status.outputs.group_id`) |
| `spec.require[].azureAd` | `AccessRuleAzureAd` |  |  |  |
| `spec.require[].azureAd.id` | `string` | yes |  |  |
| `spec.require[].azureAd.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.require[].githubOrganization` | `AccessRuleGithubOrganization` |  |  |  |
| `spec.require[].githubOrganization.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.require[].githubOrganization.name` | `string` | yes |  |  |
| `spec.require[].githubOrganization.team` | `string` |  |  |  |
| `spec.require[].gsuite` | `AccessRuleGsuite` |  |  |  |
| `spec.require[].gsuite.email` | `string` | yes |  |  |
| `spec.require[].gsuite.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.require[].okta` | `AccessRuleOkta` |  |  |  |
| `spec.require[].okta.name` | `string` | yes |  |  |
| `spec.require[].okta.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.require[].saml` | `AccessRuleSaml` |  |  |  |
| `spec.require[].saml.attributeName` | `string` | yes |  |  |
| `spec.require[].saml.attributeValue` | `string` | yes |  |  |
| `spec.require[].saml.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.require[].oidc` | `AccessRuleOidc` |  |  |  |
| `spec.require[].oidc.claimName` | `string` | yes |  |  |
| `spec.require[].oidc.claimValue` | `string` | yes |  |  |
| `spec.require[].oidc.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.require[].authContext` | `AccessRuleAuthContext` |  |  |  |
| `spec.require[].authContext.id` | `string` | yes |  |  |
| `spec.require[].authContext.acId` | `string` | yes |  |  |
| `spec.require[].authContext.identityProviderId` | `string \| valueFrom` | yes |  |  |
| `spec.require[].authMethod` | `AccessRuleAuthMethod` |  |  |  |
| `spec.require[].authMethod.authMethod` | `string` | yes |  |  |
| `spec.require[].commonName` | `AccessRuleCommonName` |  |  |  |
| `spec.require[].commonName.commonName` | `string` | yes |  |  |
| `spec.require[].geo` | `AccessRuleGeo` |  |  |  |
| `spec.require[].geo.countryCode` | `string` | yes |  |  |
| `spec.require[].devicePosture` | `AccessRuleDevicePosture` |  |  |  |
| `spec.require[].devicePosture.integrationUid` | `string \| valueFrom` | yes |  |  |
| `spec.require[].externalEvaluation` | `AccessRuleExternalEvaluation` |  |  |  |
| `spec.require[].externalEvaluation.evaluateUrl` | `string` | yes |  |  |
| `spec.require[].externalEvaluation.keysUrl` | `string` | yes |  |  |
| `spec.require[].loginMethod` | `AccessRuleLoginMethod` |  |  |  |
| `spec.require[].loginMethod.id` | `string \| valueFrom` | yes |  |  |
| `spec.require[].serviceToken` | `AccessRuleServiceToken` |  |  |  |
| `spec.require[].serviceToken.tokenId` | `string \| valueFrom` | yes |  |  |
| `spec.require[].anyValidServiceToken` | `AccessRuleEmpty` |  |  |  |
| `spec.require[].linkedAppToken` | `AccessRuleLinkedAppToken` |  |  |  |
| `spec.require[].linkedAppToken.appUid` | `string \| valueFrom` | yes |  |  |
| `spec.require[].userRiskScore` | `AccessRuleUserRiskScore` |  |  |  |
| `spec.require[].userRiskScore.userRiskScore` | `[]enum` | yes |  |  |
| `spec.require[].cloudflareAccountMember` | `AccessRuleCloudflareAccountMember` |  |  |  |
| `spec.require[].cloudflareAccountMember.accountId` | `string` |  |  |  |
| `spec.sessionDuration` | `string` |  | `24h` |  |
| `spec.approvalRequired` | `bool` |  |  |  |
| `spec.approvalGroups` | `[]CloudflareZeroTrustAccessPolicyApprovalGroup` |  |  |  |
| `spec.approvalGroups[].approvalsNeeded` | `uint32` |  |  |  |
| `spec.approvalGroups[].emailAddresses` | `[]string` |  |  |  |
| `spec.approvalGroups[].emailListUuid` | `string \| valueFrom` |  |  |  |
| `spec.isolationRequired` | `bool` |  |  |  |
| `spec.purposeJustificationRequired` | `bool` |  |  |  |
| `spec.purposeJustificationPrompt` | `string` |  |  |  |
| `spec.connectionRules` | `CloudflareZeroTrustAccessPolicyConnectionRules` |  |  |  |
| `spec.connectionRules.rdp` | `CloudflareZeroTrustAccessPolicyRdpRules` |  |  |  |
| `spec.connectionRules.rdp.allowedClipboardLocalToRemoteFormats` | `[]string` |  |  |  |
| `spec.connectionRules.rdp.allowedClipboardRemoteToLocalFormats` | `[]string` |  |  |  |
| `spec.mfaConfig` | `CloudflareZeroTrustAccessPolicyMfaConfig` |  |  |  |
| `spec.mfaConfig.allowedAuthenticators` | `[]enum` |  |  |  |
| `spec.mfaConfig.mfaDisabled` | `bool` |  |  |  |
| `spec.mfaConfig.sessionDuration` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this policy.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.name

`string` · required

The display name of the Access policy.

- rule: {"string":{"minLen":"1"}}

### spec.decision

`enum` · required

The decision applied when this policy matches a request.

- rule: decision must be allow, deny, non_identity, or bypass
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `decision_unspecified` -- Unspecified (invalid).
- `allow` -- Allow the request when the rules match.
- `deny` -- Block the request when the rules match.
- `non_identity` -- Allow without requiring an identity (service-token / mTLS only).
- `bypass` -- Skip Access entirely for matching requests (e.g. health checks).

### spec.include

`[]CloudflareAccessRule` · required

Rules evaluated with an OR operator: a request matches the policy if it
satisfies ANY include rule. At least one include rule is required.

- rule: {"repeated":{"minItems":"1"}}

### spec.include[].email

`AccessRuleEmail`

Match a single email address (e.g. "jane@example.com").

### spec.include[].email.email

`string` · required

The email address to match.

- rule: {"required":true,"string":{"email":true}}

### spec.include[].emailDomain

`AccessRuleEmailDomain`

Match any email at a domain (e.g. "example.com").

### spec.include[].emailDomain.domain

`string` · required

The email domain to match (e.g. "example.com").

- rule: {"required":true}

### spec.include[].emailList

`AccessRuleListRef`

Match emails in a Cloudflare email list (by list ID).

### spec.include[].emailList.id

`string | valueFrom` · required

The Cloudflare list ID, as a literal or a reference to another resource's
output (future-proofed for a first-class list kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].everyone

`AccessRuleEmpty`

Match everyone (an unconditional allow building block).

### spec.include[].ip

`AccessRuleIp`

Match a client IP or CIDR range (e.g. "203.0.113.0/24").

### spec.include[].ip.ip

`string` · required

An IPv4/IPv6 address or CIDR range (e.g. "203.0.113.4" or "203.0.113.0/24").

- rule: {"required":true}

### spec.include[].ipList

`AccessRuleListRef`

Match client IPs in a Cloudflare IP list (by list ID).

### spec.include[].ipList.id

`string | valueFrom` · required

The Cloudflare list ID, as a literal or a reference to another resource's
output (future-proofed for a first-class list kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].certificate

`AccessRuleEmpty`

Match a valid mutual-TLS client certificate (presence only).

### spec.include[].group

`AccessRuleGroupRef`

Match members of an Access group (by group ID).

### spec.include[].group.id

`string | valueFrom` · required

The Access group to include, as a literal group ID or a reference to a
CloudflareZeroTrustAccessGroup — composing reusable membership criteria.

- references: CloudflareZeroTrustAccessGroup (`status.outputs.group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustAccessGroup, name: <that resource's name>, fieldPath: status.outputs.group_id}} -- a bare string does not parse

### spec.include[].azureAd

`AccessRuleAzureAd`

Match an Azure AD / Entra ID group.

### spec.include[].azureAd.id

`string` · required

The Azure AD group ID.

- rule: {"required":true}

### spec.include[].azureAd.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Azure AD connection, as a literal
or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].githubOrganization

`AccessRuleGithubOrganization`

Match a GitHub organization (optionally a specific team).

### spec.include[].githubOrganization.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the GitHub connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].githubOrganization.name

`string` · required

The GitHub organization name.

- rule: {"required":true}

### spec.include[].githubOrganization.team

`string`

Optional GitHub team slug within the organization; omit to match the whole org.

### spec.include[].gsuite

`AccessRuleGsuite`

Match a Google Workspace group.

### spec.include[].gsuite.email

`string` · required

The Google Workspace group email.

- rule: {"required":true}

### spec.include[].gsuite.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Google Workspace connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].okta

`AccessRuleOkta`

Match an Okta group.

### spec.include[].okta.name

`string` · required

The Okta group name.

- rule: {"required":true}

### spec.include[].okta.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Okta connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].saml

`AccessRuleSaml`

Match a SAML attribute name/value pair.

### spec.include[].saml.attributeName

`string` · required

The SAML attribute name (e.g. "groups").

- rule: {"required":true}

### spec.include[].saml.attributeValue

`string` · required

The SAML attribute value to match.

- rule: {"required":true}

### spec.include[].saml.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the SAML connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].oidc

`AccessRuleOidc`

Match an OIDC claim name/value pair.

### spec.include[].oidc.claimName

`string` · required

The OIDC claim name (e.g. "groups").

- rule: {"required":true}

### spec.include[].oidc.claimValue

`string` · required

The OIDC claim value to match.

- rule: {"required":true}

### spec.include[].oidc.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the OIDC connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].authContext

`AccessRuleAuthContext`

Match an authentication context from an identity provider.

### spec.include[].authContext.id

`string` · required

The authentication-context ID.

- rule: {"required":true}

### spec.include[].authContext.acId

`string` · required

The authentication-context "ac_id" as configured at the identity provider.

- rule: {"required":true}

### spec.include[].authContext.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID this auth context belongs to.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].authMethod

`AccessRuleAuthMethod`

Match a specific authentication method (AMR value, e.g. "mfa", "pwd").

### spec.include[].authMethod.authMethod

`string` · required

The authentication method reference value (e.g. "mfa", "pwd", "swk").

- rule: {"required":true}

### spec.include[].commonName

`AccessRuleCommonName`

Match an mTLS client certificate common name.

### spec.include[].commonName.commonName

`string` · required

The certificate common name to match.

- rule: {"required":true}

### spec.include[].geo

`AccessRuleGeo`

Match a country by ISO 3166-1 alpha-2 code (e.g. "US").

### spec.include[].geo.countryCode

`string` · required

The ISO 3166-1 alpha-2 country code (e.g. "US", "DE").

- rule: {"required":true,"string":{"len":"2"}}

### spec.include[].devicePosture

`AccessRuleDevicePosture`

Match a device-posture check result (by posture integration UID).

### spec.include[].devicePosture.integrationUid

`string | valueFrom` · required

The device-posture integration UID whose result must pass, as a literal or a
reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].externalEvaluation

`AccessRuleExternalEvaluation`

Delegate the decision to an external evaluation service.

### spec.include[].externalEvaluation.evaluateUrl

`string` · required

The URL Cloudflare calls to evaluate the request.

- rule: {"required":true}

### spec.include[].externalEvaluation.keysUrl

`string` · required

The URL serving the public keys used to verify the evaluation response.

- rule: {"required":true}

### spec.include[].loginMethod

`AccessRuleLoginMethod`

Match users who authenticated through a specific login method / IdP.

### spec.include[].loginMethod.id

`string | valueFrom` · required

The Cloudflare identity-provider ID of the required login method, as a literal
or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].serviceToken

`AccessRuleServiceToken`

Match a specific Access service token (by token ID).

### spec.include[].serviceToken.tokenId

`string | valueFrom` · required

The Access service-token ID, as a literal or a reference to another resource's
output (future-proofed for a first-class service-token kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].anyValidServiceToken

`AccessRuleEmpty`

Match any valid Access service token (machine-to-machine).

### spec.include[].linkedAppToken

`AccessRuleLinkedAppToken`

Match a token minted by a linked Access application (by app UID).

### spec.include[].linkedAppToken.appUid

`string | valueFrom` · required

The UID of the linked Access application, as a literal or a reference to
another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.include[].userRiskScore

`AccessRuleUserRiskScore`

Match users whose Cloudflare user-risk score is in the given set.

### spec.include[].userRiskScore.userRiskScore

`[]enum` · required

The set of risk levels that match. At least one is required. (The field name
mirrors the Cloudflare provider's nested attribute for a 1:1 mapping.)

- rule: user-risk levels must be low, medium, high, or unscored
- rule: {"repeated":{"minItems":"1","items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `level_unspecified` -- Unspecified (invalid).
- `low` -- Low risk.
- `medium` -- Medium risk.
- `high` -- High risk.
- `unscored` -- No score computed yet.

### spec.include[].cloudflareAccountMember

`AccessRuleCloudflareAccountMember`

Match members of the Cloudflare account (by account ID).

tofu↔pulumi parity: the Pulumi Cloudflare SDK (v6.17.0) does not yet expose
this rule variant; the Terraform module implements it and the Pulumi module
omits it. See the Pulumi module README.

### spec.include[].cloudflareAccountMember.accountId

`string`

The Cloudflare account ID whose members match. Omit to match the account that
owns this policy.

- rule: account_id must be a 32-character hex string

### spec.exclude

`[]CloudflareAccessRule`

Rules evaluated with a NOT operator: a request is rejected by the policy if it
satisfies ANY exclude rule (exclude wins over include). Optional.

### spec.exclude[].email

`AccessRuleEmail`

Match a single email address (e.g. "jane@example.com").

### spec.exclude[].email.email

`string` · required

The email address to match.

- rule: {"required":true,"string":{"email":true}}

### spec.exclude[].emailDomain

`AccessRuleEmailDomain`

Match any email at a domain (e.g. "example.com").

### spec.exclude[].emailDomain.domain

`string` · required

The email domain to match (e.g. "example.com").

- rule: {"required":true}

### spec.exclude[].emailList

`AccessRuleListRef`

Match emails in a Cloudflare email list (by list ID).

### spec.exclude[].emailList.id

`string | valueFrom` · required

The Cloudflare list ID, as a literal or a reference to another resource's
output (future-proofed for a first-class list kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].everyone

`AccessRuleEmpty`

Match everyone (an unconditional allow building block).

### spec.exclude[].ip

`AccessRuleIp`

Match a client IP or CIDR range (e.g. "203.0.113.0/24").

### spec.exclude[].ip.ip

`string` · required

An IPv4/IPv6 address or CIDR range (e.g. "203.0.113.4" or "203.0.113.0/24").

- rule: {"required":true}

### spec.exclude[].ipList

`AccessRuleListRef`

Match client IPs in a Cloudflare IP list (by list ID).

### spec.exclude[].ipList.id

`string | valueFrom` · required

The Cloudflare list ID, as a literal or a reference to another resource's
output (future-proofed for a first-class list kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].certificate

`AccessRuleEmpty`

Match a valid mutual-TLS client certificate (presence only).

### spec.exclude[].group

`AccessRuleGroupRef`

Match members of an Access group (by group ID).

### spec.exclude[].group.id

`string | valueFrom` · required

The Access group to include, as a literal group ID or a reference to a
CloudflareZeroTrustAccessGroup — composing reusable membership criteria.

- references: CloudflareZeroTrustAccessGroup (`status.outputs.group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustAccessGroup, name: <that resource's name>, fieldPath: status.outputs.group_id}} -- a bare string does not parse

### spec.exclude[].azureAd

`AccessRuleAzureAd`

Match an Azure AD / Entra ID group.

### spec.exclude[].azureAd.id

`string` · required

The Azure AD group ID.

- rule: {"required":true}

### spec.exclude[].azureAd.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Azure AD connection, as a literal
or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].githubOrganization

`AccessRuleGithubOrganization`

Match a GitHub organization (optionally a specific team).

### spec.exclude[].githubOrganization.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the GitHub connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].githubOrganization.name

`string` · required

The GitHub organization name.

- rule: {"required":true}

### spec.exclude[].githubOrganization.team

`string`

Optional GitHub team slug within the organization; omit to match the whole org.

### spec.exclude[].gsuite

`AccessRuleGsuite`

Match a Google Workspace group.

### spec.exclude[].gsuite.email

`string` · required

The Google Workspace group email.

- rule: {"required":true}

### spec.exclude[].gsuite.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Google Workspace connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].okta

`AccessRuleOkta`

Match an Okta group.

### spec.exclude[].okta.name

`string` · required

The Okta group name.

- rule: {"required":true}

### spec.exclude[].okta.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Okta connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].saml

`AccessRuleSaml`

Match a SAML attribute name/value pair.

### spec.exclude[].saml.attributeName

`string` · required

The SAML attribute name (e.g. "groups").

- rule: {"required":true}

### spec.exclude[].saml.attributeValue

`string` · required

The SAML attribute value to match.

- rule: {"required":true}

### spec.exclude[].saml.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the SAML connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].oidc

`AccessRuleOidc`

Match an OIDC claim name/value pair.

### spec.exclude[].oidc.claimName

`string` · required

The OIDC claim name (e.g. "groups").

- rule: {"required":true}

### spec.exclude[].oidc.claimValue

`string` · required

The OIDC claim value to match.

- rule: {"required":true}

### spec.exclude[].oidc.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the OIDC connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].authContext

`AccessRuleAuthContext`

Match an authentication context from an identity provider.

### spec.exclude[].authContext.id

`string` · required

The authentication-context ID.

- rule: {"required":true}

### spec.exclude[].authContext.acId

`string` · required

The authentication-context "ac_id" as configured at the identity provider.

- rule: {"required":true}

### spec.exclude[].authContext.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID this auth context belongs to.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].authMethod

`AccessRuleAuthMethod`

Match a specific authentication method (AMR value, e.g. "mfa", "pwd").

### spec.exclude[].authMethod.authMethod

`string` · required

The authentication method reference value (e.g. "mfa", "pwd", "swk").

- rule: {"required":true}

### spec.exclude[].commonName

`AccessRuleCommonName`

Match an mTLS client certificate common name.

### spec.exclude[].commonName.commonName

`string` · required

The certificate common name to match.

- rule: {"required":true}

### spec.exclude[].geo

`AccessRuleGeo`

Match a country by ISO 3166-1 alpha-2 code (e.g. "US").

### spec.exclude[].geo.countryCode

`string` · required

The ISO 3166-1 alpha-2 country code (e.g. "US", "DE").

- rule: {"required":true,"string":{"len":"2"}}

### spec.exclude[].devicePosture

`AccessRuleDevicePosture`

Match a device-posture check result (by posture integration UID).

### spec.exclude[].devicePosture.integrationUid

`string | valueFrom` · required

The device-posture integration UID whose result must pass, as a literal or a
reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].externalEvaluation

`AccessRuleExternalEvaluation`

Delegate the decision to an external evaluation service.

### spec.exclude[].externalEvaluation.evaluateUrl

`string` · required

The URL Cloudflare calls to evaluate the request.

- rule: {"required":true}

### spec.exclude[].externalEvaluation.keysUrl

`string` · required

The URL serving the public keys used to verify the evaluation response.

- rule: {"required":true}

### spec.exclude[].loginMethod

`AccessRuleLoginMethod`

Match users who authenticated through a specific login method / IdP.

### spec.exclude[].loginMethod.id

`string | valueFrom` · required

The Cloudflare identity-provider ID of the required login method, as a literal
or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].serviceToken

`AccessRuleServiceToken`

Match a specific Access service token (by token ID).

### spec.exclude[].serviceToken.tokenId

`string | valueFrom` · required

The Access service-token ID, as a literal or a reference to another resource's
output (future-proofed for a first-class service-token kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].anyValidServiceToken

`AccessRuleEmpty`

Match any valid Access service token (machine-to-machine).

### spec.exclude[].linkedAppToken

`AccessRuleLinkedAppToken`

Match a token minted by a linked Access application (by app UID).

### spec.exclude[].linkedAppToken.appUid

`string | valueFrom` · required

The UID of the linked Access application, as a literal or a reference to
another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.exclude[].userRiskScore

`AccessRuleUserRiskScore`

Match users whose Cloudflare user-risk score is in the given set.

### spec.exclude[].userRiskScore.userRiskScore

`[]enum` · required

The set of risk levels that match. At least one is required. (The field name
mirrors the Cloudflare provider's nested attribute for a 1:1 mapping.)

- rule: user-risk levels must be low, medium, high, or unscored
- rule: {"repeated":{"minItems":"1","items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `level_unspecified` -- Unspecified (invalid).
- `low` -- Low risk.
- `medium` -- Medium risk.
- `high` -- High risk.
- `unscored` -- No score computed yet.

### spec.exclude[].cloudflareAccountMember

`AccessRuleCloudflareAccountMember`

Match members of the Cloudflare account (by account ID).

tofu↔pulumi parity: the Pulumi Cloudflare SDK (v6.17.0) does not yet expose
this rule variant; the Terraform module implements it and the Pulumi module
omits it. See the Pulumi module README.

### spec.exclude[].cloudflareAccountMember.accountId

`string`

The Cloudflare account ID whose members match. Omit to match the account that
owns this policy.

- rule: account_id must be a 32-character hex string

### spec.require

`[]CloudflareAccessRule`

Rules evaluated with an AND operator: a request must satisfy EVERY require
rule (e.g. require a country AND a device-posture check). Optional.

### spec.require[].email

`AccessRuleEmail`

Match a single email address (e.g. "jane@example.com").

### spec.require[].email.email

`string` · required

The email address to match.

- rule: {"required":true,"string":{"email":true}}

### spec.require[].emailDomain

`AccessRuleEmailDomain`

Match any email at a domain (e.g. "example.com").

### spec.require[].emailDomain.domain

`string` · required

The email domain to match (e.g. "example.com").

- rule: {"required":true}

### spec.require[].emailList

`AccessRuleListRef`

Match emails in a Cloudflare email list (by list ID).

### spec.require[].emailList.id

`string | valueFrom` · required

The Cloudflare list ID, as a literal or a reference to another resource's
output (future-proofed for a first-class list kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].everyone

`AccessRuleEmpty`

Match everyone (an unconditional allow building block).

### spec.require[].ip

`AccessRuleIp`

Match a client IP or CIDR range (e.g. "203.0.113.0/24").

### spec.require[].ip.ip

`string` · required

An IPv4/IPv6 address or CIDR range (e.g. "203.0.113.4" or "203.0.113.0/24").

- rule: {"required":true}

### spec.require[].ipList

`AccessRuleListRef`

Match client IPs in a Cloudflare IP list (by list ID).

### spec.require[].ipList.id

`string | valueFrom` · required

The Cloudflare list ID, as a literal or a reference to another resource's
output (future-proofed for a first-class list kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].certificate

`AccessRuleEmpty`

Match a valid mutual-TLS client certificate (presence only).

### spec.require[].group

`AccessRuleGroupRef`

Match members of an Access group (by group ID).

### spec.require[].group.id

`string | valueFrom` · required

The Access group to include, as a literal group ID or a reference to a
CloudflareZeroTrustAccessGroup — composing reusable membership criteria.

- references: CloudflareZeroTrustAccessGroup (`status.outputs.group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustAccessGroup, name: <that resource's name>, fieldPath: status.outputs.group_id}} -- a bare string does not parse

### spec.require[].azureAd

`AccessRuleAzureAd`

Match an Azure AD / Entra ID group.

### spec.require[].azureAd.id

`string` · required

The Azure AD group ID.

- rule: {"required":true}

### spec.require[].azureAd.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Azure AD connection, as a literal
or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].githubOrganization

`AccessRuleGithubOrganization`

Match a GitHub organization (optionally a specific team).

### spec.require[].githubOrganization.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the GitHub connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].githubOrganization.name

`string` · required

The GitHub organization name.

- rule: {"required":true}

### spec.require[].githubOrganization.team

`string`

Optional GitHub team slug within the organization; omit to match the whole org.

### spec.require[].gsuite

`AccessRuleGsuite`

Match a Google Workspace group.

### spec.require[].gsuite.email

`string` · required

The Google Workspace group email.

- rule: {"required":true}

### spec.require[].gsuite.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Google Workspace connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].okta

`AccessRuleOkta`

Match an Okta group.

### spec.require[].okta.name

`string` · required

The Okta group name.

- rule: {"required":true}

### spec.require[].okta.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the Okta connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].saml

`AccessRuleSaml`

Match a SAML attribute name/value pair.

### spec.require[].saml.attributeName

`string` · required

The SAML attribute name (e.g. "groups").

- rule: {"required":true}

### spec.require[].saml.attributeValue

`string` · required

The SAML attribute value to match.

- rule: {"required":true}

### spec.require[].saml.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the SAML connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].oidc

`AccessRuleOidc`

Match an OIDC claim name/value pair.

### spec.require[].oidc.claimName

`string` · required

The OIDC claim name (e.g. "groups").

- rule: {"required":true}

### spec.require[].oidc.claimValue

`string` · required

The OIDC claim value to match.

- rule: {"required":true}

### spec.require[].oidc.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID for the OIDC connection.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].authContext

`AccessRuleAuthContext`

Match an authentication context from an identity provider.

### spec.require[].authContext.id

`string` · required

The authentication-context ID.

- rule: {"required":true}

### spec.require[].authContext.acId

`string` · required

The authentication-context "ac_id" as configured at the identity provider.

- rule: {"required":true}

### spec.require[].authContext.identityProviderId

`string | valueFrom` · required

The Cloudflare identity-provider ID this auth context belongs to.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].authMethod

`AccessRuleAuthMethod`

Match a specific authentication method (AMR value, e.g. "mfa", "pwd").

### spec.require[].authMethod.authMethod

`string` · required

The authentication method reference value (e.g. "mfa", "pwd", "swk").

- rule: {"required":true}

### spec.require[].commonName

`AccessRuleCommonName`

Match an mTLS client certificate common name.

### spec.require[].commonName.commonName

`string` · required

The certificate common name to match.

- rule: {"required":true}

### spec.require[].geo

`AccessRuleGeo`

Match a country by ISO 3166-1 alpha-2 code (e.g. "US").

### spec.require[].geo.countryCode

`string` · required

The ISO 3166-1 alpha-2 country code (e.g. "US", "DE").

- rule: {"required":true,"string":{"len":"2"}}

### spec.require[].devicePosture

`AccessRuleDevicePosture`

Match a device-posture check result (by posture integration UID).

### spec.require[].devicePosture.integrationUid

`string | valueFrom` · required

The device-posture integration UID whose result must pass, as a literal or a
reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].externalEvaluation

`AccessRuleExternalEvaluation`

Delegate the decision to an external evaluation service.

### spec.require[].externalEvaluation.evaluateUrl

`string` · required

The URL Cloudflare calls to evaluate the request.

- rule: {"required":true}

### spec.require[].externalEvaluation.keysUrl

`string` · required

The URL serving the public keys used to verify the evaluation response.

- rule: {"required":true}

### spec.require[].loginMethod

`AccessRuleLoginMethod`

Match users who authenticated through a specific login method / IdP.

### spec.require[].loginMethod.id

`string | valueFrom` · required

The Cloudflare identity-provider ID of the required login method, as a literal
or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].serviceToken

`AccessRuleServiceToken`

Match a specific Access service token (by token ID).

### spec.require[].serviceToken.tokenId

`string | valueFrom` · required

The Access service-token ID, as a literal or a reference to another resource's
output (future-proofed for a first-class service-token kind).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].anyValidServiceToken

`AccessRuleEmpty`

Match any valid Access service token (machine-to-machine).

### spec.require[].linkedAppToken

`AccessRuleLinkedAppToken`

Match a token minted by a linked Access application (by app UID).

### spec.require[].linkedAppToken.appUid

`string | valueFrom` · required

The UID of the linked Access application, as a literal or a reference to
another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.require[].userRiskScore

`AccessRuleUserRiskScore`

Match users whose Cloudflare user-risk score is in the given set.

### spec.require[].userRiskScore.userRiskScore

`[]enum` · required

The set of risk levels that match. At least one is required. (The field name
mirrors the Cloudflare provider's nested attribute for a 1:1 mapping.)

- rule: user-risk levels must be low, medium, high, or unscored
- rule: {"repeated":{"minItems":"1","items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `level_unspecified` -- Unspecified (invalid).
- `low` -- Low risk.
- `medium` -- Medium risk.
- `high` -- High risk.
- `unscored` -- No score computed yet.

### spec.require[].cloudflareAccountMember

`AccessRuleCloudflareAccountMember`

Match members of the Cloudflare account (by account ID).

tofu↔pulumi parity: the Pulumi Cloudflare SDK (v6.17.0) does not yet expose
this rule variant; the Terraform module implements it and the Pulumi module
omits it. See the Pulumi module README.

### spec.require[].cloudflareAccountMember.accountId

`string`

The Cloudflare account ID whose members match. Omit to match the account that
owns this policy.

- rule: account_id must be a 32-character hex string

### spec.sessionDuration

`string` · optional (explicit presence)

How long a user session is valid before re-authentication, as a duration
string (e.g. "30m", "24h", "768h"); "0s" forces authentication on every
request. Defaults to 24h.

- default: `24h`
- rule: session_duration must be a duration like "30m", "24h", or "0s"

### spec.approvalRequired

`bool`

Whether a request for access must be approved before the user is granted
entry (purpose-justification / approval workflow). Defaults to false.

### spec.approvalGroups

`[]CloudflareZeroTrustAccessPolicyApprovalGroup`

Approval groups consulted when approval_required is true. A request is
approved once each listed group has supplied its required number of approvals.

### spec.approvalGroups[].approvalsNeeded

`uint32`

The number of approvals required from this group.

- rule: {"uint32":{"gte":1}}

### spec.approvalGroups[].emailAddresses

`[]string`

Email addresses of the approvers in this group.

### spec.approvalGroups[].emailListUuid

`string | valueFrom`

The Cloudflare email-list UUID supplying the approvers (alternative to
email_addresses), as a literal or a reference to another resource's output.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.isolationRequired

`bool`

Whether the user must complete a browser-isolation session to use the
application this policy guards. Defaults to false.

### spec.purposeJustificationRequired

`bool`

Whether the user must enter a justification when requesting access. Defaults
to false.

### spec.purposeJustificationPrompt

`string`

The prompt shown to the user when a purpose justification is required.

### spec.connectionRules

`CloudflareZeroTrustAccessPolicyConnectionRules`

Connection rules for infrastructure / RDP targets this policy guards.

### spec.connectionRules.rdp

`CloudflareZeroTrustAccessPolicyRdpRules`

RDP connection constraints.

### spec.connectionRules.rdp.allowedClipboardLocalToRemoteFormats

`[]string`

Clipboard formats allowed from the local machine to the remote host.
Currently only "text" is supported.

- rule: {"repeated":{"items":{"string":{"in":["text"]}}}}

### spec.connectionRules.rdp.allowedClipboardRemoteToLocalFormats

`[]string`

Clipboard formats allowed from the remote host to the local machine.
Currently only "text" is supported.

- rule: {"repeated":{"items":{"string":{"in":["text"]}}}}

### spec.mfaConfig

`CloudflareZeroTrustAccessPolicyMfaConfig`

Per-policy multi-factor requirements layered on top of the matched identity.

### spec.mfaConfig.allowedAuthenticators

`[]enum`

The set of authenticators a user may use to satisfy MFA for this policy.

- rule: allowed_authenticators must be totp, biometrics, or security_key
- rule: {"repeated":{"items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `authenticator_unspecified` -- Unspecified (invalid).
- `totp` -- Time-based one-time password.
- `biometrics` -- Platform biometrics.
- `security_key` -- Hardware/passkey security key.

### spec.mfaConfig.mfaDisabled

`bool`

Disable MFA enforcement for this policy (overrides allowed_authenticators).

### spec.mfaConfig.sessionDuration

`string`

Re-prompt for MFA after this duration (e.g. "8h"). Empty uses the session
default.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustAccessPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | The Cloudflare-assigned identifier of the policy. Reference this from a CloudflareZeroTrustAccessApplication's policies list to attach the policy. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.include[].group.id` | CloudflareZeroTrustAccessGroup | `status.outputs.group_id` |
| `spec.exclude[].group.id` | CloudflareZeroTrustAccessGroup | `status.outputs.group_id` |
| `spec.require[].group.id` | CloudflareZeroTrustAccessGroup | `status.outputs.group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareZeroTrustAccessApplication | `spec.policies[].policy` | `status.outputs.policy_id` |

## See Also

- [Overview](../README.md)
