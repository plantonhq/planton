# Enterprise

Stigmer Enterprise under a license: every caller signs in through your own OpenID Connect issuer, each session runs in a sandbox of its own, and the front door is routes on a Gateway you run. The license is a Planton secret reference, written into a Secret the server reads.

## When to Use

- You hold a Stigmer Enterprise license
- Your organization signs people in through its own identity provider

## Key Configuration Choices

- **`identity.oidc`** — required by Enterprise; the realm or tenant, the console and native clients, the audience, the owner's account and sign-up are set up in your issuer
- **`sandboxes.driver`** — a licensed install refuses to run sessions side by side on the shared runner, so a per-session driver is required (and with it, `object_store`, which with the bundled database turns the archive on: cert-manager must run on the cluster)
- **`ingress.gateway`** — the Gateway's HTTPS listener for the hostname holds the certificate; Stigmer never edits the Gateway
- **Temporal** — bundled here; with `temporal.external`, Enterprise requires TLS or an API key to it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `hostname` | Where people reach Stigmer | A DNS name your Gateway serves |
| `edge` | Your KubernetesGateway | The Gateway resource in this environment |
| `issuer`, client ids | Your OIDC issuer and clients | Your identity provider |
| `$secret/stigmer-license-key` | Your license key | The key you were sent, as a Planton secret |

## Related Presets

- **04-exposed-with-sign-in** — the bundled sign-in, without a license
