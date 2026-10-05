# AwsSesEmailIdentity

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsSesEmailIdentitySpec defines an Amazon SES (SESv2) email identity:
the verified domain or email address an application is allowed to send
mail FROM.

An identity is the trust anchor of the SES graph. A DOMAIN identity is
the production shape -- it verifies through DNS (the dkim_tokens
output composes directly into AwsRoute53DnsRecord CNAMEs), signs mail
with DKIM, covers every address at the domain, and unlocks a custom
MAIL FROM domain for aligned SPF. An EMAIL_ADDRESS identity verifies
through a confirmation link sent to the mailbox -- quick for testing,
but it cannot carry its own DKIM configuration.

The identity composes with AwsSesConfigurationSet (its default sending
rules and event destinations) and with AwsRoute53DnsRecord (DKIM CNAMEs
and the MAIL FROM domain's MX/SPF records).

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSesEmailIdentity
metadata:
  name: test-ses-identity
  org: test-org
  env: dev
  id: test-ses-identity-dev
spec:
  region: us-west-2
  emailIdentity: mail.example.com
  dkimSigning:
    nextSigningKeyLength: RSA_2048_BIT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.emailIdentity` | `string` | yes |  |  |
| `spec.configurationSet` | `string \| valueFrom` |  |  | AwsSesConfigurationSet (`status.outputs.configuration_set_name`) |
| `spec.dkimSigning` | `AwsSesEmailIdentityDkimSigning` |  |  |  |
| `spec.dkimSigning.nextSigningKeyLength` | `string` |  |  |  |
| `spec.dkimSigning.domainSigningPrivateKey` | `string` (sensitive) |  |  |  |
| `spec.dkimSigning.domainSigningSelector` | `string` |  |  |  |
| `spec.mailFrom` | `AwsSesEmailIdentityMailFrom` |  |  |  |
| `spec.mailFrom.mailFromDomain` | `string` | yes |  |  |
| `spec.mailFrom.behaviorOnMxFailure` | `string` |  | `USE_DEFAULT_VALUE` |  |
| `spec.emailForwardingEnabled` | `bool` |  |  |  |
| `spec.policies` | `[]AwsSesEmailIdentityPolicy` |  |  |  |
| `spec.policies[].name` | `string` | yes |  |  |
| `spec.policies[].policy` | `object` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the identity is created. SES identities are
regional: verify the same domain in each region you send from.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.emailIdentity

`string` · required

The identity to verify: a domain ("example.com") for the production
shape, or a single email address ("sender@example.com") for
mailbox-verified sending. IMMUTABLE: changing it replaces the
identity (the new identity re-verifies from scratch).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"253"}}

### spec.configurationSet

`string | valueFrom`

The configuration set applied by default to every message sent from
this identity -- the delivery/tracking/event-publishing rules defined
once and inherited here. Reference an AwsSesConfigurationSet's
configuration_set_name output or pass a literal set name. Can be
attached, swapped, or removed in place.

- references: AwsSesConfigurationSet (`status.outputs.configuration_set_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSesConfigurationSet, name: <that resource's name>, fieldPath: status.outputs.configuration_set_name}} -- a bare string does not parse

### spec.dkimSigning

`AwsSesEmailIdentityDkimSigning`

DKIM signing configuration -- Easy DKIM (AWS-managed keys) or BYODKIM
(bring your own key pair). Only DOMAIN identities carry DKIM
configuration; leave unset for email-address identities (they inherit
the domain's DKIM when the domain is also verified) and to accept
Easy DKIM with a 2048-bit key, the AWS default, on domains.

- rule: dkim_signing must configure an arm -- next_signing_key_length (Easy DKIM) or the BYODKIM key/selector pair; omit the block to accept AWS's Easy DKIM default
- rule: domain_signing_private_key and domain_signing_selector must be set together (the BYODKIM pair)
- rule: next_signing_key_length (Easy DKIM) cannot be combined with the BYODKIM key/selector pair

### spec.dkimSigning.nextSigningKeyLength

`string`

Easy DKIM: the RSA key length AWS generates for the NEXT signing key
-- "RSA_1024_BIT" or "RSA_2048_BIT" (prefer 2048, the AWS default;
1024 exists for DNS providers with 255-character TXT limits).
Setting it on a live identity rotates the key. Mutually exclusive
with the BYODKIM pair (empty when BYODKIM is used).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["RSA_1024_BIT","RSA_2048_BIT"]}}

### spec.dkimSigning.domainSigningPrivateKey

`string` · sensitive

BYODKIM: the base64-encoded RSA PRIVATE key SES signs with (PKCS #8,
headers stripped). A signing secret -- never logged, never exported.
Requires domain_signing_selector.

- rule: {"string":{"maxLen":"20480"}}

### spec.dkimSigning.domainSigningSelector

`string`

BYODKIM: the DKIM selector under which YOU publish the public key
("<selector>._domainkey.<domain>" TXT record). Requires
domain_signing_private_key.

- rule: {"string":{"maxLen":"63"}}

### spec.mailFrom

`AwsSesEmailIdentityMailFrom`

Custom MAIL FROM domain configuration. By default SES uses its own
bounce domain (amazonses.com) as the envelope MAIL FROM, which fails
strict DMARC alignment on SPF; a custom MAIL FROM subdomain (e.g.
"mail.example.com" with its MX + SPF records, composable via
AwsRoute53DnsRecord) aligns the envelope with the sending domain.

### spec.mailFrom.mailFromDomain

`string` · required

The custom MAIL FROM domain. Must be a subdomain of the identity's
domain (e.g. "mail.example.com" for "example.com"), must not be used
to receive mail, and needs an MX record pointing at the regional SES
feedback endpoint plus an SPF TXT record -- both composable with
AwsRoute53DnsRecord.

- rule: {"required":true,"string":{"maxLen":"253"}}

### spec.mailFrom.behaviorOnMxFailure

`string` · optional (explicit presence)

What SES does when the MAIL FROM domain's MX record is missing or
broken:
  USE_DEFAULT_VALUE -- fall back to amazonses.com and keep sending
                       (the AWS default; mail flows but loses DMARC
                       SPF alignment).
  REJECT_MESSAGE    -- fail the send with MailFromDomainNotVerified
                       (strict: nothing leaves unaligned).

- default: `USE_DEFAULT_VALUE`
- rule: {"string":{"in":["USE_DEFAULT_VALUE","REJECT_MESSAGE"]}}

### spec.emailForwardingEnabled

`bool` · optional (explicit presence)

Whether bounce and complaint notifications are forwarded by email to
the identity's mailbox. Tri-state: leave UNSET to accept AWS's own
default (forwarding on) with no managed setting at all; set true or
false to pin the position explicitly -- the modules materialize the
feedback sub-resource only when a position is taken. Turn it off once
event destinations or SNS feedback handle bounces -- forwarding is the
fallback channel, not the production one. One retention caveat
(live-verified 2026-08-12): SES remembers the LAST-WRITTEN forwarding
value per identity name, surviving even DeleteEmailIdentity and
re-creation -- so on an identity that was ever explicitly managed,
clearing this field stops managing the setting but does NOT restore
AWS's default; set true explicitly to turn forwarding back on.

### spec.policies

`[]AwsSesEmailIdentityPolicy`

Named authorization policies on this identity -- the cross-account
sending grants that let another AWS account or role send mail AS this
identity (SendEmail with this identity as the source). Each policy is
its own AWS sub-resource keyed by name, materialized per-name by the
modules. Names must be unique.

### spec.policies[].name

`string` · required

The policy's name, unique within the identity (1-64 characters;
changing it replaces the policy).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64","pattern":"^[a-zA-Z0-9_-]+$"}}

### spec.policies[].policy

`object` · required

The policy document as a structured JSON object (IAM policy syntax;
resource must be this identity's ARN, principals are the accounts or
roles being granted ses:SendEmail / ses:SendRawEmail).

- rule: {"required":true}

## Validation Rules

- `dkim_requires_domain_identity`: dkim_signing can only be set on a domain identity (email-address identities inherit the domain's DKIM)
- `policy_names_unique`: each identity policy must have a unique name within the identity

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSesEmailIdentity, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.identity_arn` | `string` | The Amazon Resource Name (ARN) of the identity -- the resource for identity-policy grants and IAM statements that scope sending. |
| `status.outputs.email_identity` | `string` | The identity string itself (the domain or email address) -- the join key downstream automation uses to compose DNS names. |
| `status.outputs.identity_type` | `string` | The identity type AWS classified this as: "DOMAIN" or "EMAIL_ADDRESS". |
| `status.outputs.verification_status` | `string` | The identity's verification status at deploy time: "PENDING" until the DKIM CNAMEs (domain) or the confirmation link (email address) have been acted on, then "SUCCESS". A PENDING identity exists but cannot send yet -- publish the dkim_tokens records to complete domain verification. |
| `status.outputs.dkim_tokens` | `[]string` | Easy DKIM's three CNAME tokens. Publish each as "<token>._domainkey.<domain>" CNAME "<token>.dkim.amazonses.com" (composable with AwsRoute53DnsRecord); SES flips the identity to verified once it sees them. Empty for BYODKIM (you publish your own selector record) and for email-address identities. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.configurationSet` | AwsSesConfigurationSet | `status.outputs.configuration_set_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsCognitoUserPool | `spec.emailConfiguration.sourceArn` | `status.outputs.identity_arn` |
| AwsCognitoUserPool | `spec.riskConfiguration.accountTakeover.notifyConfiguration.sourceArn` | `status.outputs.identity_arn` |
| AwsCognitoUserPoolClient | `spec.riskConfiguration.accountTakeover.notifyConfiguration.sourceArn` | `status.outputs.identity_arn` |

## See Also

- [Overview](../README.md)
