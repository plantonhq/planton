# Auth0User

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0UserSpec defines the configuration for an Auth0 User.
In Auth0, a user is an identity that signs in through exactly one connection.
The Management API can create users only in connections Auth0 itself holds
the credential for: a database connection (email and password, optionally a
username) or a passwordless connection (email or SMS one-time codes). Users
of social and enterprise connections are created by the identity provider
at first sign-in and cannot be declared here.

Declared users are the identities an operator owns rather than a person who
signed up: a staff root account, a support or service account, a seeded test
identity, a break-glass administrator. The user's stable identity-provider
subject (the `auth0|...` user id, exactly the `sub` claim in every token
issued for the user) is an output, so other declarations reference the
identity by output instead of copying a value someone read from a dashboard.

Roles and API permissions are folded into this component: the IaC modules
create the user AND set its complete role list and permission list in one
deployment, and both sets are authoritative -- a role or permission removed
from the manifest is removed from the user on the next apply.

https://auth0.com/docs/manage-users/user-accounts/create-users
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/user
https://www.pulumi.com/registry/packages/auth0/api-docs/user/

## Example

```yaml
# Auth0 User Test Manifest
# This file is used for testing the Auth0User component
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: Your Auth0 tenant domain (e.g., your-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - create:users
#    - read:users
#    - update:users
#    - delete:users
#    - read:roles (when roles are assigned)
#
# 3. The connection must be a database (auth0) or passwordless (email, sms)
#    connection that already exists in the tenant (e.g., created via the
#    Auth0Connection component); the referenced role and resource server
#    must exist too.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0User
metadata:
  name: test-staff-root
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The connection the user is created in, by reference to an Auth0Connection
  connection_name:
    value_from:
      kind: Auth0Connection
      name: users
      field_path: status.outputs.name

  # Who the person is
  email: test-staff-root@test.planton.dev
  name: Test Staff Root
  given_name: Test
  family_name: Root

  # An identity the operator vouches for: verified, no confirmation email
  email_verified: true
  verify_email: false

  # No password: the modules mint one and report it in status.outputs.password

  # Data the application reads at sign-in; the user cannot change it
  app_metadata:
    plan: internal

  # Authoritative roles and a direct permission, by reference
  roles:
    - value_from:
        kind: Auth0Role
        name: administrator
        field_path: status.outputs.id
  permissions:
    - name: read:items
      resource_server_identifier:
        value_from:
          kind: Auth0ResourceServer
          name: api
          field_path: status.outputs.identifier
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.connectionName` | `string \| valueFrom` | yes |  | Auth0Connection (`status.outputs.name`) |
| `spec.email` | `string` |  |  |  |
| `spec.emailVerified` | `bool` |  |  |  |
| `spec.verifyEmail` | `bool` |  |  |  |
| `spec.username` | `string` |  |  |  |
| `spec.name` | `string` |  |  |  |
| `spec.givenName` | `string` |  |  |  |
| `spec.familyName` | `string` |  |  |  |
| `spec.nickname` | `string` |  |  |  |
| `spec.picture` | `string` |  |  |  |
| `spec.phoneNumber` | `string` |  |  |  |
| `spec.phoneVerified` | `bool` |  |  |  |
| `spec.blocked` | `bool` |  |  |  |
| `spec.userId` | `string` |  |  |  |
| `spec.password` | `string` (sensitive) |  |  |  |
| `spec.passwordless` | `bool` |  |  |  |
| `spec.userMetadata` | `object` |  |  |  |
| `spec.appMetadata` | `object` |  |  |  |
| `spec.customDomainHeader` | `string` |  |  |  |
| `spec.roles` | `[]string \| valueFrom` |  |  | Auth0Role (`status.outputs.id`) |
| `spec.permissions` | `[]Auth0UserPermission` |  |  |  |
| `spec.permissions[].name` | `string` | yes |  |  |
| `spec.permissions[].resourceServerIdentifier` | `string \| valueFrom` | yes |  | Auth0ResourceServer (`status.outputs.identifier`) |

## Field Details

### spec.connectionName

`string | valueFrom` · required

connection_name is the connection the user belongs to, by name.
Auth0 creates users only in connections it holds the credential for: a
database connection (strategy auth0) or a passwordless connection (email
or sms). Reference an Auth0Connection's status.outputs.name so the connection
is created first and its name is never retyped. Immutable after creation: a
user cannot be moved between connections, and the modules replace the user
when this changes.

https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/user#connection_name

- references: Auth0Connection (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0Connection, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.email

`string`

email is the user's email address. The sign-in identifier for database
connections that do not require a username, and for passwordless email
connections. Auth0 stores it lowercase.
Example: "platform-root@example.com".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"email":true}}

### spec.emailVerified

`bool`

email_verified marks the email address as already verified, so the user
is never asked to confirm it. Set it for identities an operator owns and
vouches for (a staff account, a service identity); leave it false for a
person who should confirm their own address. Independent of verify_email,
which controls whether a verification message is sent at creation.

### spec.verifyEmail

`bool` · optional (explicit presence)

verify_email controls whether Auth0 sends a verification email when the
user is created. Auth0's own behavior when this is not set is to send the
message unless the address is already marked verified, so leave it unset
to keep that behavior and state false only to deliberately suppress the
message for an address that is not yet verified. Overrides the implicit
behavior of email_verified.

### spec.username

`string`

username is the user's login name. Valid only when the connection is a
database connection configured to require usernames
(database_options.requires_username on Auth0Connection); Auth0 refuses it
on every other connection. 1 to 15 characters unless the connection's
username policy widens the range.

### spec.name

`string`

name is the user's full display name, shown in the Auth0 dashboard and
carried in the profile's "name" claim. Defaults to the email address when
omitted on a database connection. Example: "Planton dev root".

### spec.givenName

`string`

given_name is the user's first name, carried in the profile's "given_name"
claim. Purely descriptive; Auth0 never derives it from name.

### spec.familyName

`string`

family_name is the user's last name, carried in the profile's
"family_name" claim. Purely descriptive; Auth0 never derives it from name.

### spec.nickname

`string`

nickname is the user's preferred short name, carried in the profile's
"nickname" claim. Defaults to the local part of the email address when
omitted on a database connection.

### spec.picture

`string`

picture is the URL of the user's avatar, carried in the profile's
"picture" claim. Defaults to a Gravatar-style placeholder when omitted.
Example: "https://www.example.com/avatar.png".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.phoneNumber

`string`

phone_number is the user's phone number in E.164 form (a leading "+", the
country code, then the number with no spaces). The sign-in identifier for
passwordless SMS connections; descriptive on every other connection.
Example: "+14155550123".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^\\+[1-9][0-9]{6,14}$"}}

### spec.phoneVerified

`bool`

phone_verified marks the phone number as already verified. Requires
phone_number.

### spec.blocked

`bool`

blocked disables the user: every sign-in attempt is refused until the flag
is cleared. The account and its profile remain; this is a suspension, not
a deletion. Independent of the brute-force blocks Auth0 applies on its own.

### spec.userId

`string`

user_id is the identifier Auth0 stores for the user, WITHOUT the
connection prefix: a user declared with user_id "root-dev" on a database
connection is "auth0|root-dev" in every token. Leave it empty and Auth0
assigns a random identifier, which is the right choice for almost every
user; declare it only when the subject must be known before the user
exists (a fixture another system is configured against). Immutable after
creation: the modules replace the user when this changes.

### spec.password

`string` · sensitive

password is the user's initial password, as a managed-secret reference
(`$secret/<slug>`) -- never plaintext. Leave it empty and the modules
generate one that satisfies the connection's password policy (24
characters, upper- and lower-case letters and digits) and report it in
status.outputs.password, so a declared user is usable without anyone
typing or escrowing a value up front. A declared password is used as
given and is never echoed back in the outputs. Meaningless on a
passwordless connection: set passwordless instead. Changing this field
resets the user's password on the next apply.

https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/user#password

### spec.passwordless

`bool`

passwordless declares that the connection is a passwordless one (strategy
email or sms), where users sign in with a one-time code and Auth0 refuses
a password at creation. When set, the modules send no password and mint
none. Leave it false for a database connection.

### spec.userMetadata

`object`

user_metadata is free-form profile data the user themselves may read and
change through a profile screen: display preferences, a work address, a
locale. Stored as a JSON object on the user; never influences what the
user can do. Keys and nesting are the application's to define.

Example:
  user_metadata:
    locale: en-US
    theme: dark

https://auth0.com/docs/manage-users/user-accounts/metadata

### spec.appMetadata

`object`

app_metadata is data that affects what the user can do or how the
application behaves for them -- a support plan, an external account id, a
tenant assignment -- and that the user cannot change themselves. Stored
as a JSON object on the user and available to Actions and rules at
sign-in. Keys and nesting are the application's to define. Roles and
permissions have first-class fields below; do not encode them here.

Example:
  app_metadata:
    plan: enterprise
    crm_account_id: acct_01H...

https://auth0.com/docs/manage-users/user-accounts/metadata

### spec.customDomainHeader

`string`

custom_domain_header sets the Auth0-Custom-Domain header on every
Management API request the modules make for this user, so the request is
handled under that custom domain (which matters for emails Auth0 sends on
the user's behalf, such as the verification message, whose links then
carry the custom domain). A custom domain configured on the provider
takes precedence when both are set. Example: "login.example.com".

### spec.roles

`[]string | valueFrom`

roles is the authoritative set of Auth0 roles assigned to the user, each
by role id. Reference an Auth0Role's status.outputs.id so the role is
created first and its id is never retyped. The set is authoritative: the
deployment manages the complete role list, and a role omitted from the
manifest is removed from the user on the next apply. Leave it empty for a
user with no roles.

https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/user_roles

- references: Auth0Role (`status.outputs.id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0Role, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.permissions

`[]Auth0UserPermission`

permissions is the authoritative set of API permissions (scopes) granted
directly to the user, beside whatever the user's roles grant. Each entry
names a scope and the resource server (API) that defines it. The set is
authoritative: a permission omitted from the manifest is removed from the
user on the next apply. Prefer roles for anything more than a one-off
grant; direct permissions exist for the exceptions.

Example:
  permissions:
    - name: read:reports
      resource_server_identifier: https://api.example.com/

https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/user_permissions

- rule: Each permission needs a name -- the scope as defined on the resource server, e.g. 'read:reports'.

### spec.permissions[].name

`string` · required

name is the permission (scope) name as defined on the resource server --
the same value that appears in an access token's "scope" or "permissions"
claim. Example: "read:reports", "manage:billing".

- rule: {"required":true}

### spec.permissions[].resourceServerIdentifier

`string | valueFrom` · required

resource_server_identifier is the identifier (audience) of the Auth0
Resource Server that defines this permission -- the API's unique
identifier, typically a URI, the same value used as the "audience" in
authorization requests. Reference an Auth0ResourceServer's
status.outputs.identifier so the API is created first and its identifier
is never retyped. Example: "https://api.example.com/".

- references: Auth0ResourceServer (`status.outputs.identifier`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0ResourceServer, name: <that resource's name>, fieldPath: status.outputs.identifier}} -- a bare string does not parse

## Validation Rules

- `identifier_required`: A user needs at least one sign-in identifier: an email address, a username (on a connection that requires usernames), or a phone number (on an SMS connection).
- `phone_verified_requires_phone_number`: phone_verified can only be set when phone_number is provided.
- `password_and_passwordless_exclusive`: A passwordless connection (email or SMS one-time codes) has no password to set -- remove the password, or clear passwordless for a database connection.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0User, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.user_id` | `string` | user_id is the user's full identity-provider subject, connection prefix included (e.g. "auth0\|66f1c2d3e4a5b6c7d8e9f0a1" for a database user, "email\|..." or "sms\|..." for a passwordless one). This is exactly the "sub" claim in every token Auth0 issues for the user, and the value a system that grants standing by identity-provider subject reads. Stable for the life of the user. |
| `status.outputs.email` | `string` | email is the user's email address as Auth0 stored it (lowercase). |
| `status.outputs.username` | `string` | username is the user's login name, when the connection requires one. |
| `status.outputs.name` | `string` | name is the user's full display name as stored, including the default Auth0 derived when none was declared. |
| `status.outputs.nickname` | `string` | nickname is the user's short name as stored, including the default Auth0 derived when none was declared. |
| `status.outputs.picture` | `string` | picture is the URL of the user's avatar as stored, including the placeholder Auth0 assigned when none was declared. |
| `status.outputs.connection_name` | `string` | connection_name is the name of the connection the user belongs to. |
| `status.outputs.password` | `string` | password is the initial password the modules generated, set ONLY when spec.password was left empty on a database connection. A declared password is never echoed back here, and a passwordless user has none. Read it once into the credential store that owns it; every later rotation happens in Auth0, and this output keeps the value the modules minted, not the current one. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.connectionName` | Auth0Connection | `status.outputs.name` |
| `spec.roles` | Auth0Role | `status.outputs.id` |
| `spec.permissions[].resourceServerIdentifier` | Auth0ResourceServer | `status.outputs.identifier` |

## See Also

- [Overview](../README.md)
