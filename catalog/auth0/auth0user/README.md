# Auth0User

Manages an [Auth0 User](https://auth0.com/docs/manage-users/user-accounts/create-users) -- an identity in one of the tenant's database or passwordless connections, with its profile, its authoritative roles and direct API permissions, and, on a database connection, a password the modules mint when none is declared. The user's identity-provider subject (the `auth0|...` id, exactly the `sub` claim in every token) is an output other declarations reference.

## When to Use

- **Identities you own**: a staff root, a break-glass administrator, a support or service account -- born from the same files as the rest of the environment, never created by signing up.
- **Standing as code**: the roles a user holds and the one-off scopes beside them, every one by reference, reviewed by reading one file.
- **Seeded test identities**: end-to-end suites that need a known user, recreated with the environment.
- **Subjects by reference**: a system that grants standing by identity-provider subject reads `status.outputs.user_id` instead of a value copied from a dashboard.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0User
metadata:
  name: staff-root
  org: acme-corp
  env: production
spec:
  connection_name:
    value_from:
      kind: Auth0Connection
      name: users
      field_path: status.outputs.name
  email: platform-root@acme.example
  name: Platform root
  email_verified: true
  verify_email: false
  roles:
    - value_from:
        kind: Auth0Role
        name: administrator
        field_path: status.outputs.id
```

No `password` is declared, so the modules mint one and report it in `status.outputs.password`.

## Key Behaviors

- **connection_name**: The database (`auth0`) or passwordless (`email`, `sms`) connection the user is created in, by reference to an [Auth0Connection](../auth0connection/README.md). Users of social and enterprise connections are created by the identity provider at sign-in and cannot be declared. Immutable: a change replaces the user.
- **password**: Three honest states. Empty on a database connection means the modules generate a 24-character password (letters and digits) and report it once as an output. Declared -- always as a managed-secret reference, never plaintext -- means used as given and never echoed back. `passwordless: true` means the connection takes no password at all.
- **verify_email**: Left unset, Auth0 decides (a confirmation is sent unless the address is already verified). State `false` only to deliberately suppress the message for an unverified address; `email_verified: true` is the other half for an identity the operator vouches for.
- **roles** and **permissions**: Authoritative sets. The deployment manages the complete role list (`auth0_user_roles`) and the complete direct-permission list (`auth0_user_permissions`); an entry removed from the manifest is removed from the user on the next apply. Roles reference [Auth0Role](../auth0role/README.md) ids; permissions name a scope on an [Auth0ResourceServer](../auth0resourceserver/README.md) identifier.
- **user_metadata** and **app_metadata**: Two JSON documents on the user -- the first the person may edit through a profile screen, the second only the application writes and reads at sign-in.
- **blocked**: A reversible suspension: every sign-in is refused while the record and its subject remain.
- **user_id**: Empty lets Auth0 assign the identifier; declared pins the unprefixed half (`root-dev` becomes `auth0|root-dev`). Immutable.

## Outputs

| Output | Description |
|---|---|
| `user_id` | The full identity-provider subject, connection prefix included (e.g. `auth0\|66f1c2d3...`) |
| `email` | The email address as stored |
| `username` | The login name, when the connection requires one |
| `name` | The display name as stored |
| `nickname` | The short name as stored |
| `picture` | The avatar URL as stored |
| `connection_name` | The connection the user belongs to |
| `password` | The minted initial password -- set only when the modules generated it |

## Auth0 Documentation

- [Create users](https://auth0.com/docs/manage-users/user-accounts/create-users)
- [User metadata](https://auth0.com/docs/manage-users/user-accounts/metadata)
- [Assign roles to users](https://auth0.com/docs/manage-users/access-control/configure-core-rbac/rbac-users/assign-roles-to-users)
- [Passwordless connections](https://auth0.com/docs/authenticate/passwordless)
- [Terraform auth0_user](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/user)
- [Terraform auth0_user_roles](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/user_roles)
- [Terraform auth0_user_permissions](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/user_permissions)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
