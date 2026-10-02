# Check Senders' Tokens at a Gateway You Own

A door other systems write through (clusters shipping telemetry, services calling an ingest endpoint), where you mint each sender's token yourself. The Gateway checks every token against public keys it holds inline, one key per sender under one issuer, so istiod never fetches keys and nothing outside the cluster can change who gets in.

## When to Use

- You run a dedicated Gateway for machine-to-machine writes and mint one long-lived token per sender.
- You want to revoke one sender without touching the others.

## Key Configuration Choices

- **`target_refs`** -- the Gateway, by reference. The policy must live in the Gateway's own namespace (cross-namespace target references are not supported).
- **`jwt_rules[].issuer`** -- required beside an inline key set; every sender's token names it as `iss`. Using the door's hostname keeps it self-describing.
- **`jwt_rules[].audiences`** -- set explicitly; the door's hostname again.
- **`jwt_rules[].jwks`** -- a JSON Web Key Set with one RSA key per sender, each key id the sender's name. Public keys only. Revoking a sender is deleting its key; replacing its token is minting a new key and token and swapping that one key.

## Prerequisites

- Istio with Gateway API support (`KubernetesIstio`), and the Gateway (`KubernetesGateway`), which should be the door's own: a token check on a Gateway refuses every bearer token that is not one of its JWTs, so a shared front door would refuse other clients' tokens.
- An ALLOW `KubernetesAuthorizationPolicy` on the same Gateway listing each sender's principal, `<issuer>/<sender>`. Without it a request with no token passes untouched.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | The Gateway's namespace. |
| `<gateway>` | The `KubernetesGateway` resource's name. |
| `<issuer>` | The issuer and audience every token names, e.g. the door's hostname. |
| `<sender>` | A sender's name: its key id and its tokens' subject. |
| `<modulus>` | The sender's RSA public modulus, base64url. |

## What a correct door answers

No token: 403 (the ALLOW policy). A token signed by a key the set does not hold: 401. A valid token on an allowed path: the backend's own answer. The token is removed before the backend sees it unless `forward_original_token` is set.
