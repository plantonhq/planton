# Auth0 User Guide

Declare a user only for an identity you own -- a root, a service account, a seeded test identity -- and never for a person who should sign up. The one judgment this kind exists to protect: the subject is the durable thing, so read it by reference and pin it only when something outside Planton is configured against it.

## When to use it (and when not)

**Use it** when an identity must exist before anyone signs in and must be recreatable from files: a staff root a control plane trusts by subject, an on-call mailbox in a passwordless connection, the fixture an end-to-end suite signs in as. The user, its password posture, and its complete standing are one manifest, and destroying the environment destroys the user with it.

**Do not use it** for the tenant's ordinary members. A self-serve product's users arrive through signup and social connections; declaring them here would fight the connection's own lifecycle, and users of social and enterprise connections cannot be created through the Management API at all -- the provider refuses with a connection error at create. Only `auth0` (database), `email`, and `sms` connections accept a declared user.

**Roles vs direct permissions.** `roles` is the normal shape; `permissions` exists for the single scope a role would be too heavy to mint for. Both are authoritative, so the manifest is the complete statement of standing -- which also means a role assigned in the dashboard is revoked on the next apply. If someone needs to grant roles by hand, that user is not this kind's to declare.

## Conventions and gotchas

- **The mint law, SaaS form.** Leave `password` empty on a database connection and the module mints one (24 characters, letters and digits, three classes -- above every built-in Auth0 policy including `excellent`) and reports it in `status.outputs.password`. There is no Kubernetes Secret to land in here, so the output is the credential's first home: read it once into the store that owns it. Every later rotation happens in Auth0 and the output keeps the minted value, not the current one. A declared password is never echoed back.
- **Three password states, one connection kind decides.** `passwordless: true` sends no password and mints none; Auth0 refuses a password on an `email` or `sms` connection, so a database-shaped manifest pointed at a passwordless connection fails at create with a `password` error, not a connection error. The validation rule refuses `password` beside `passwordless`, but it cannot know the connection's strategy -- that is the author's to match.
- **`verify_email` is three-state on purpose.** Unset means Auth0 decides (a confirmation goes out unless the address is already verified); an explicit `false` suppresses it. For an identity you own, the pair is `emailVerified: true` and `verifyEmail: false`; setting only the second on an unverified address leaves the person unable to prove the address later.
- **Email is unique per connection, and the subject is not the email.** A second user with the same email in the same connection fails at create with a conflict. The subject (`status.outputs.user_id`, the `auth0|...` id) is what other systems should hold; a destroyed and recreated user gets a NEW subject unless `userId` pins the unprefixed half, and everything that granted standing to the old subject must be re-pointed. Pin `userId` only for the fixture case; let Auth0 assign it for real identities.
- **`connectionName` and `userId` are immutable.** Changing either replaces the user (a new subject, a new minted password). The reference to `Auth0Connection` orders the connection first and keeps the name from being retyped; a literal `value:` is for connections managed outside Planton.
- **Metadata is two JSON documents, not one.** `userMetadata` is the person's (editable through a profile screen); `appMetadata` is the application's (read at sign-in by Actions, never user-editable). Both reach the provider as JSON strings; an untouched document is left unset rather than sent as `{}`.
- **`blocked` is the reversible verb.** It refuses every sign-in and keeps the record and subject; a delete removes the subject for good. Suspend first.
- **Management API scopes.** The provider connection's Machine-to-Machine application needs `create:users read:users update:users delete:users`, and `read:roles` when `roles` is set (the provider reads role names back). A tenant credential minted for clients and connections alone lacks all five, and the first apply fails with a 403 naming the missing scope.
- **MFA is not a field.** Auth0 exposes no Management API surface to enroll a factor for a user; enrollment is a recorded operational step at the user's first sign-in, and `controls.yaml` says so.

## Built for parity, hand-derived

The Auth0 provider carries no parity schema artifact, so this table is the accounting the parity tool would otherwise measure. Every `auth0_user` argument at provider 1.57.0 and both companion resources map to a spec field:

| Provider argument | Spec field | Note |
|---|---|---|
| `connection_name` | `connectionName` | reference to `Auth0Connection` |
| `email`, `email_verified`, `verify_email` | same, camelCase | `verifyEmail` presence-tracked |
| `username`, `name`, `given_name`, `family_name`, `nickname`, `picture` | same | |
| `phone_number`, `phone_verified` | same | E.164 enforced |
| `blocked` | `blocked` | |
| `user_id` | `userId` | the unprefixed half |
| `password` | `password` | sensitive; empty means minted; `passwordless` is the third state |
| `user_metadata`, `app_metadata` | `userMetadata`, `appMetadata` | Struct in, JSON string out |
| `custom_domain_header` | `customDomainHeader` | |
| `auth0_user_roles.roles` | `roles` | references to `Auth0Role`; authoritative |
| `auth0_user_permissions.permissions[]` | `permissions[]` | reference to `Auth0ResourceServer`; authoritative |

Not modeled, with reasons: `auth0_user_role` and `auth0_user_permission` (the singular companions) are the non-authoritative form of the two folded sets and would let dashboard edits survive an apply, the opposite of what a declared identity wants.

## On the diagram

A user renders as its own node with three kinds of edges: into the connection it lives in, and out to each role and each resource server it holds standing on. Declaring standing by reference is what draws those edges; a literal role id or audience draws nothing, so an architecture whose access model should be visible declares its roles and APIs as kinds first.

## Pairs well with

- **Auth0Connection** (strategy `auth0`, `email`, or `sms`) -- the connection the user is created in; the user references its name.
- **Auth0Role** -- the standing the user holds; referenced by id, authoritative.
- **Auth0ResourceServer** -- the API whose scopes a direct permission names; referenced by identifier.
- **Auth0Client** -- the application the user signs in through; not referenced by the user, but the connection must be enabled for that client or the sign-in is refused before the password is ever checked.
