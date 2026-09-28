# Device Biometrics First

This preset runs the tenant on Universal Login with identifier-first login, and offers the device's own authenticator -- Touch ID, Face ID, Windows Hello, a platform passkey -- as the first factor. After the first screen, a person who has enrolled their device signs in with it instead of typing a password; everyone else is offered to enroll after their next password sign-in. It is Auth0's "Identifier First + Biometrics" authentication profile.

## When to Use

- A product whose people sign in from their own laptops and phones, where a password is the slowest and weakest step
- A tenant with a database connection (Auth0's own user store): device biometrics apply to database users

## Key Configuration Choices

- **Universal Login** (`universalLoginExperience: new`) -- device biometrics are offered only on Universal Login
- **Identifier first** (`identifierFirst: true`) -- the email comes first, so Auth0 knows whether this person enrolled this device before asking for anything else
- **Device biometrics first** (`webauthnPlatformFirstFactor: true`) -- Auth0 accepts it only when WebAuthn with device biometrics is enabled as a multi-factor method on the tenant (Dashboard: Security, Multi-factor Auth); enable the factor first, or the apply is refused
- **Not a requirement** -- a person without an enrolled device still signs in with their password; enforce multi-factor authentication in the tenant's multi-factor policy

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-identifier-first** -- identifier-first login without the device's authenticator
