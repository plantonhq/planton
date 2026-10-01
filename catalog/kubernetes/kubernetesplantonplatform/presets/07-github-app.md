# One GitHub App for the Whole Install

A platform at a real hostname with a GitHub App registered for the whole install on github.com, so every organization connects its repositories in one click: no team registers an App of its own, pushes are verified as your App's deliveries, and builds post check runs back to GitHub.

## When to Use

- A shared install where many teams deploy from GitHub and the platform team wants one App to govern instead of one per team
- Any install whose front door GitHub can reach, so pushes trigger runs the moment they land

## Prerequisites

- A GitHub App registered on github.com (or on your GitHub Enterprise Server, declared as its own host), its webhook pointed at `https://<hostname>/webhooks/github`, and its webhook secret set
- The Secret holding the App's PEM private key and webhook secret in the platform's namespace, created before the platform is applied. The operator preflights it, and a missing Secret shows in the platform's status while the host is offered without an App
- Platform release `v0.0.113` or newer (the first whose control plane signs with an App the install declares) and operator chart 0.14.1 or newer (the first whose definition knows `github`)

## Key Configuration Choices

- **Register the App on the host it signs for.** An App on github.com cannot sign for an enterprise server; declare each host with its own App
- **The key is never encoded or inline.** Keep the PEM exactly as GitHub generated it and name it by Secret key reference; it reaches the control plane as a mounted file, so a rotated key is live on the next token
- **Hosts are offered in the order you list them.** Put a company's enterprise server first to make it the wizard's first choice
- **Say when the front door cannot tell.** `webhooks: reachable` for an enterprise server on the install's private network, `unreachable` for a public host that cannot reach a door public only inside a perimeter; Planton then checks GitHub for pushes and tells every team so

## Placeholders to Replace

- `planton.example.com` — the install's hostname, also the App's webhook host
- `letsencrypt` — your cert-manager issuer
- `Iv23liExampleClientId` — the App's Client ID from its settings page
- `planton-github-app` — the name of the Secret holding the key and webhook secret

## Related Presets

- **02-ingress-tls** — the same hostname and TLS without a GitHub App; teams bring their own
- **04-gateway-api** — the same URL through a Gateway API Gateway
- **05-email-smtp** — outbound email, the other declaration a team-ready install usually adds
