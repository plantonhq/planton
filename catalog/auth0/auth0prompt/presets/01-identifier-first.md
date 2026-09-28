# Identifier-First Login

This preset runs the tenant on Universal Login and asks for the email or username on a first screen and the password on a second. A person whose email domain belongs to an enterprise connection is routed to their company's identity provider after the first screen, never asked for a password your tenant does not hold. It is the flow a product that sells to companies wants from its first enterprise customer on.

## When to Use

- Any tenant that has, or will have, enterprise connections (Okta, Entra ID, Google Workspace, SAML) whose people sign in with their company account
- A tenant moving off the Classic, Lock-based pages onto Universal Login

## Key Configuration Choices

- **Universal Login** (`universalLoginExperience: new`) -- the experience every branding, theme, custom text and partial styles; the Classic pages reach none of them
- **Identifier first** (`identifierFirst: true`) -- the email decides the next step: the company's identity provider for an enterprise connection's domains, the password screen for everyone else
- **Unset is unmanaged** -- `webauthnPlatformFirstFactor` is left out, so the tenant keeps its current value

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

Routing by email domain also needs each enterprise connection's identity provider domains (Auth0 dashboard: Authentication, Enterprise, the connection's Login Experience tab); without them everyone reaches the password screen.

## Related Presets

- **02-device-biometrics-first** -- the same flow, with the device's own authenticator offered before the password
