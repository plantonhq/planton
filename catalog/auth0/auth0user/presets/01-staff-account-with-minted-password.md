# Staff Account with Minted Password

## Pattern

An identity the operator owns rather than a person who signed up: a staff root, a break-glass administrator, a service account. The connection is referenced, the email is vouched for, and no password appears anywhere -- the module mints one and reports it once in the outputs.

## What It Does

- Creates the user in the referenced database connection, ordering the connection first.
- Marks the email verified and sends no confirmation message, because the operator vouches for the address.
- Generates a 24-character initial password (letters and digits, three classes) and reports it in `status.outputs.password`; nothing is typed or escrowed up front.
- Reports the user's identity-provider subject in `status.outputs.user_id`, the value a system that grants standing by subject reads.

## When to Use

- You need an account nobody can create by signing up, born from the same files as the rest of the environment.
- The password's first home is a credential store you read the output into once, and every later rotation happens in Auth0.
- The account's subject must be referenceable by other declarations without anyone copying it from a dashboard.

## Customization

- Change `spec.email` and `spec.name` to the identity you own. Auth0 requires a unique email per connection.
- Add `spec.roles` (by reference to `Auth0Role`) when the account should hold standing from the start; the set is authoritative.
- Set `spec.userId` only when the subject must be known before the user exists.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `metadata.org` | Your Planton organization |
| `spec.connectionName.valueFrom.name` | The name of your `Auth0Connection` database connection |
| `spec.email` | The address of the identity you own |
| `spec.name` | The display name shown in the Auth0 dashboard and the profile's name claim |
