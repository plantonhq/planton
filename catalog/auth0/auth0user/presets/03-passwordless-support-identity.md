# Passwordless Support Identity

## Pattern

A user in a passwordless connection: sign-in is a one-time code sent to the address, there is no password to mint, declare, or rotate, and the account still carries the roles its work needs.

## What It Does

- Creates the user in the referenced passwordless email connection and sends no password, because Auth0 refuses one on that connection kind.
- Marks the email verified so the first code is delivered without a confirmation step.
- Assigns the referenced support role authoritatively.
- Seeds a locale in `userMetadata`, the profile data the user may change themselves.

## When to Use

- Shared or rotating identities (an on-call mailbox, a demo account) where a password would be one more secret to manage.
- Tenants whose passwordless connection is the sign-in door for staff.
- Any user in an `email` or `sms` connection; for SMS, identify the user by `phoneNumber` instead of `email`.

## Customization

- Switch to an SMS connection by referencing it in `connectionName` and replacing `email` with `phoneNumber` in E.164 form (and `phoneVerified` for `emailVerified`).
- Add `permissions` for one-off scopes beside the role.
- Remove `emailVerified` to have Auth0 confirm the address at first sign-in.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `metadata.org` | Your Planton organization |
| `spec.connectionName.valueFrom.name` | The name of your passwordless `Auth0Connection` (strategy `email` or `sms`) |
| `spec.email` | The address the one-time codes are sent to |
| `spec.roles[].valueFrom.name` | The name of your `Auth0Role` |
