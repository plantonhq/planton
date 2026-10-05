# AwsOrganizationAccount

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsOrganizationAccountSpec defines the desired configuration for one
MEMBER account of an AWS Organization: creating the account,
placing it in the OU tree, and managing its account-level settings
(alternate contacts, primary contact, opt-in region enablement).

The account's display name is an explicit spec field - AWS allows
names with spaces metadata.name cannot carry. AWS identifies the
account by its 12-digit account ID (the import ID; importing an
existing member account adopts it without recreating).

Everything here runs FROM the organization's management account.
The account-settings arms (contacts, regions) additionally require
trusted access for AWS Account Management on the organization
(service principal "account.amazonaws.com" in the AwsOrganization's
aws_service_access_principals).

THE DELETE CONTRACT deserves attention. close_on_deletion=false
(the default) merely REMOVES the account from the organization on
destroy - the account lives on as a standalone AWS account (it must
carry standalone billing information for the removal to succeed).
close_on_deletion=true CLOSES the account: AWS suspends it for ~90
days (PENDING_CLOSURE) before permanent deletion, and account
closures are quota-limited per rolling 30 days. Neither path is a
clean "delete" - treat member accounts as long-lived.

## Example

```yaml
# Canonical AwsOrganizationAccount example (hack/dev manifest and
# refgen Example source): a member account placed in an OU with a
# custom bootstrap role, root-only billing access, billing/security
# alternate contacts, primary contact information, and one opt-in
# region enabled. Literal values stand in for composed references so
# the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOrganizationAccount
metadata:
  name: workloads-prod
  id: workloads-prod
  org: test-org
  env: dev
spec:
  region: us-west-2
  accountName: Workloads Production
  email: aws-workloads-prod@example.com
  parentId:
    value: ou-e2e1-workload1
  roleName: OrgBootstrapRole
  iamUserAccessToBilling: DENY
  closeOnDeletion: true
  alternateContacts:
    billing:
      name: Jane Doe
      title: Finance Lead
      emailAddress: billing@example.com
      phoneNumber: "+1 555 0100"
    security:
      name: Security Team
      title: CISO
      emailAddress: security@example.com
      phoneNumber: "+1 555 0101"
  primaryContact:
    fullName: Jane Doe
    companyName: Acme Corp
    addressLine1: 1 Main St
    city: Seattle
    stateOrRegion: WA
    postalCode: "98101"
    countryCode: US
    phoneNumber: "+1 555 0100"
  regions:
    - regionName: ap-southeast-3
      enabled: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.accountName` | `string` | yes |  |  |
| `spec.email` | `string` | yes |  |  |
| `spec.parentId` | `string \| valueFrom` |  |  | AwsOrganizationalUnit (`status.outputs.ou_id`) |
| `spec.roleName` | `string` |  |  |  |
| `spec.iamUserAccessToBilling` | `string` |  |  |  |
| `spec.closeOnDeletion` | `bool` |  |  |  |
| `spec.createGovcloud` | `bool` |  |  |  |
| `spec.alternateContacts` | `AwsOrganizationAccountAlternateContacts` |  |  |  |
| `spec.alternateContacts.billing` | `AwsOrganizationAccountAlternateContact` |  |  |  |
| `spec.alternateContacts.billing.name` | `string` | yes |  |  |
| `spec.alternateContacts.billing.title` | `string` | yes |  |  |
| `spec.alternateContacts.billing.emailAddress` | `string` |  |  |  |
| `spec.alternateContacts.billing.phoneNumber` | `string` |  |  |  |
| `spec.alternateContacts.operations` | `AwsOrganizationAccountAlternateContact` |  |  |  |
| `spec.alternateContacts.operations.name` | `string` | yes |  |  |
| `spec.alternateContacts.operations.title` | `string` | yes |  |  |
| `spec.alternateContacts.operations.emailAddress` | `string` |  |  |  |
| `spec.alternateContacts.operations.phoneNumber` | `string` |  |  |  |
| `spec.alternateContacts.security` | `AwsOrganizationAccountAlternateContact` |  |  |  |
| `spec.alternateContacts.security.name` | `string` | yes |  |  |
| `spec.alternateContacts.security.title` | `string` | yes |  |  |
| `spec.alternateContacts.security.emailAddress` | `string` |  |  |  |
| `spec.alternateContacts.security.phoneNumber` | `string` |  |  |  |
| `spec.primaryContact` | `AwsOrganizationAccountPrimaryContact` |  |  |  |
| `spec.primaryContact.fullName` | `string` | yes |  |  |
| `spec.primaryContact.companyName` | `string` |  |  |  |
| `spec.primaryContact.addressLine1` | `string` | yes |  |  |
| `spec.primaryContact.addressLine2` | `string` |  |  |  |
| `spec.primaryContact.addressLine3` | `string` |  |  |  |
| `spec.primaryContact.city` | `string` | yes |  |  |
| `spec.primaryContact.districtOrCounty` | `string` |  |  |  |
| `spec.primaryContact.stateOrRegion` | `string` |  |  |  |
| `spec.primaryContact.postalCode` | `string` | yes |  |  |
| `spec.primaryContact.countryCode` | `string` |  |  |  |
| `spec.primaryContact.phoneNumber` | `string` |  |  |  |
| `spec.primaryContact.websiteUrl` | `string` |  |  |  |
| `spec.regions` | `[]AwsOrganizationAccountRegion` |  |  |  |
| `spec.regions[].regionName` | `string` | yes |  |  |
| `spec.regions[].enabled` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the account.
Organizations and Account Management are global services, but
every AWS API call is still made against a regional endpoint, so
a region is required.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.accountName

`string` · required

The member account's display name (1-50 characters; spaces are
legal - which is why this is an explicit field rather than
metadata.name). Renames apply in place through the Account
Management API.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.email

`string` · required

The email address of the account's root user (6-64 characters,
unique across ALL of AWS). IMMUTABLE - changing it forces
replacement, which is a full member-account lifecycle event (see
the delete contract above).

- rule: {"string":{"minLen":"6","maxLen":"64","pattern":"^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$"}}

### spec.parentId

`string | valueFrom`

Where the account sits in the OU tree: an AwsOrganizationalUnit
reference (the default wiring), a literal "r-..."/"ou-..." ID, or
unset to leave the account under the organization root. Changing
it MOVES the account in place.

- references: AwsOrganizationalUnit (`status.outputs.ou_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsOrganizationalUnit, name: <that resource's name>, fieldPath: status.outputs.ou_id}} -- a bare string does not parse

### spec.roleName

`string`

The IAM role AWS pre-creates in the new account for the
management account to assume (1-64 characters of [\w+=,.@-]).
Unset = AWS's default "OrganizationAccountAccessRole". WRITE-ONCE:
it takes effect at account creation, AWS exposes NO API to read it
back, and both engines deliberately ignore later changes (without
that, importing an existing account would plan a destructive
replacement to "set" a value AWS can never echo). Config-only on
import.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[\\w+=,.@-]{1,64}$"}}

### spec.iamUserAccessToBilling

`string`

Whether IAM users in the member account may access billing
information (ALLOW) or only the root user may (DENY). Unset =
AWS's default (ALLOW). IMMUTABLE - changing it forces
replacement.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ALLOW","DENY"]}}

### spec.closeOnDeletion

`bool`

What destroy does to the account: false (default) removes it from
the organization (the account survives standalone); true CLOSES
it (~90-day PENDING_CLOSURE suspension, quota-limited, not
supported for GovCloud accounts). Engine behavior at destroy
time, not account state - imports never see it.

### spec.createGovcloud

`bool`

Create a companion GovCloud (US) account alongside the standard
account (the govcloud_id output carries its ID). Decided at
creation - changing it later is silently ignored by AWS
(config-only on import).

### spec.alternateContacts

`AwsOrganizationAccountAlternateContacts`

The account's alternate contacts (billing, operations, security)
- each one AWS routes that category's communications to. Removing
an arm deletes that contact.

### spec.alternateContacts.billing

`AwsOrganizationAccountAlternateContact`

The contact AWS routes billing communications to.

### spec.alternateContacts.billing.name

`string` · required

The contact's name (1-64 characters).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.alternateContacts.billing.title

`string` · required

The contact's title (1-50 characters).

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.alternateContacts.billing.emailAddress

`string`

The contact's email address.

- rule: {"string":{"pattern":"^[\\w+=,.-]+@[\\w.-]+\\.[\\w]+$"}}

### spec.alternateContacts.billing.phoneNumber

`string`

The contact's phone number (digits, spaces, parentheses, "+",
and "-").

- rule: {"string":{"pattern":"^[0-9\\s()+-]+$"}}

### spec.alternateContacts.operations

`AwsOrganizationAccountAlternateContact`

The contact AWS routes operational communications to.

### spec.alternateContacts.operations.name

`string` · required

The contact's name (1-64 characters).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.alternateContacts.operations.title

`string` · required

The contact's title (1-50 characters).

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.alternateContacts.operations.emailAddress

`string`

The contact's email address.

- rule: {"string":{"pattern":"^[\\w+=,.-]+@[\\w.-]+\\.[\\w]+$"}}

### spec.alternateContacts.operations.phoneNumber

`string`

The contact's phone number (digits, spaces, parentheses, "+",
and "-").

- rule: {"string":{"pattern":"^[0-9\\s()+-]+$"}}

### spec.alternateContacts.security

`AwsOrganizationAccountAlternateContact`

The contact AWS routes security communications to.

### spec.alternateContacts.security.name

`string` · required

The contact's name (1-64 characters).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.alternateContacts.security.title

`string` · required

The contact's title (1-50 characters).

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.alternateContacts.security.emailAddress

`string`

The contact's email address.

- rule: {"string":{"pattern":"^[\\w+=,.-]+@[\\w.-]+\\.[\\w]+$"}}

### spec.alternateContacts.security.phoneNumber

`string`

The contact's phone number (digits, spaces, parentheses, "+",
and "-").

- rule: {"string":{"pattern":"^[0-9\\s()+-]+$"}}

### spec.primaryContact

`AwsOrganizationAccountPrimaryContact`

The account's primary contact information (the postal address AWS
keeps on file). AWS provides no delete API for primary contact
information: removing this arm (or destroying the account
resource) leaves the last-written contact in place, and clearing
an optional leaf leaves that leaf's last value on file.

### spec.primaryContact.fullName

`string` · required

The full name of the primary contact (1-64 characters).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.primaryContact.companyName

`string`

The company name, if the account belongs to one.

### spec.primaryContact.addressLine1

`string` · required

The first line of the primary contact's address.

- rule: {"string":{"minLen":"1"}}

### spec.primaryContact.addressLine2

`string`

The second line of the address, if needed.

### spec.primaryContact.addressLine3

`string`

The third line of the address, if needed.

### spec.primaryContact.city

`string` · required

The city of the primary contact's address.

- rule: {"string":{"minLen":"1"}}

### spec.primaryContact.districtOrCounty

`string`

The district or county, if needed.

### spec.primaryContact.stateOrRegion

`string`

The state or region. AWS requires it for some countries (e.g.
US, CA) and rejects it for others - AWS validates per country at
apply time.

### spec.primaryContact.postalCode

`string` · required

The postal code of the address.

- rule: {"string":{"minLen":"1"}}

### spec.primaryContact.countryCode

`string`

The ISO-3166 two-letter country code (e.g. "US", "DE").

- rule: {"string":{"pattern":"^[A-Z]{2}$"}}

### spec.primaryContact.phoneNumber

`string`

The phone number of the primary contact. AWS requires the
leading "+" country code here (unlike alternate contacts).

- rule: {"string":{"pattern":"^[+][0-9\\s()-]+$"}}

### spec.primaryContact.websiteUrl

`string`

The website of the company or account, if any.

### spec.regions

`[]AwsOrganizationAccountRegion`

Opt-in region enablement for the member account (e.g.
"ap-southeast-3"). Only OPT-IN regions can be managed - regions
enabled by default reject disable attempts. Enabling or disabling
a region is a long operation (up to ~60 minutes each way).
Removing an entry does NOT opt the region back out - the region
keeps its last state (the provider's destroy is a no-op).

### spec.regions[].regionName

`string` · required

The opt-in region to manage (e.g. "ap-southeast-3",
"eu-central-2").

- rule: {"string":{"minLen":"1"}}

### spec.regions[].enabled

`bool`

Whether the region is enabled for the account. An entry with
enabled=false actively DISABLES the region (also up to ~60
minutes); removing the entry instead leaves the region as-is.

## Validation Rules

- `spec.parent_id_literal_format`: a literal parent_id must be an organization root (r-...) or an organizational unit (ou-...)
- `spec.region_names_unique`: regions entries must have unique region names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsOrganizationAccount, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.account_id` | `string` | The member account's 12-digit AWS account ID (also the provider's import ID; the folded contact and region settings import as composites of it). |
| `status.outputs.arn` | `string` | The member account's ARN. |
| `status.outputs.state` | `string` | The account's lifecycle state as reported by Organizations (ACTIVE, SUSPENDED, PENDING_CLOSURE). |
| `status.outputs.govcloud_id` | `string` | The companion GovCloud (US) account's ID, when create_govcloud was set. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.parentId` | AwsOrganizationalUnit | `status.outputs.ou_id` |

## See Also

- [Overview](../README.md)
