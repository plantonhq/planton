# CloudflareTurnstileWidget

Provision a Cloudflare Turnstile widget — a privacy-preserving CAPTCHA
alternative. The widget yields a public **site key** (embedded in your page) and
a **secret key** (used server-side to verify tokens). The secret is exported as a
sensitive stack output so a Worker or backend can reference it.

## When to use

- Protecting forms (login, signup, contact) from bots without classic CAPTCHAs.
- Pairing with a Worker that validates the Turnstile token via `/siteverify`.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareTurnstileWidget
metadata:
  name: login-form-widget
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: login-form
  domains:
    - example.com
    - localhost
  mode: managed
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `name` | yes | Human-readable widget name |
| `domains` | yes | Domains the widget may run on (≥1) |
| `mode` | yes | `non-interactive`, `invisible`, or `managed` |
| `clearanceLevel` | no | `no_clearance`, `jschallenge`, `managed`, or `interactive` |
| `botFightMode` | no | Enterprise: costly challenges for malicious bots |
| `ephemeralId` | no | Enterprise: return Ephemeral ID in `/siteverify` |
| `offlabel` | no | Enterprise: hide Cloudflare branding |
| `region` | no | `world` (default) or `china` (immutable) |

## Outputs

| Output | Description |
|---|---|
| `sitekey` | Public site key (embed in the page frontend) |
| `secret` | Secret key for `/siteverify` (sensitive) |
| `created_on` | Creation timestamp |
| `modified_on` | Last-modified timestamp |

## A note on secrets

The `secret` output is exported as a sensitive value. Downstream, resolve it as a
managed-secret reference (e.g. into a `CloudflareWorker` `secret_text` binding)
rather than embedding it in plaintext.

## Related components

- `CloudflareWorker` — a Worker that validates Turnstile tokens server-side using
  the `secret` output.
