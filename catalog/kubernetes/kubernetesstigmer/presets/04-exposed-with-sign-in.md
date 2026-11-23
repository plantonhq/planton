# Exposed with Sign-in

Stigmer at a hostname of your own, behind an Ingress with a certificate cert-manager issues, with the bundled sign-in at `<hostname>/idp`. The owner signs in first with a one-time password (the `owner_password_command` output), then people with an address in your domains may sign up, verified by email.

## When to Use

- Your team reaches Stigmer from browsers, the desktop app or the CLI outside the cluster
- You want sign-in without running an identity provider of your own

## Key Configuration Choices

- **`ingress.tls.issuer`** — a ClusterIssuer by reference; cert-manager must run on the cluster. A Gateway works instead (`ingress.gateway`), with its HTTPS listener holding the certificate
- **The backend-protocol annotation** — the server speaks gRPC on the same port as the console, so the controller needs its hint (ingress-nginx's annotation here, Traefik's on `service_annotations`)
- **`identity.bundled`** — needs the bundled database: with your own Postgres, use `identity.oidc` instead
- **`sign_up_email_domains` needs `smtp`** — without a verification email anyone could claim an address in the domain

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `hostname` | Where people reach Stigmer | A DNS name pointing at your ingress controller |
| `class_name` | Your IngressClass | `kubectl get ingressclass` |
| `letsencrypt` | Your KubernetesClusterIssuer | The issuer resource in this environment |
| SMTP fields | Your mail server | Your mail provider |

## Related Presets

- **06-enterprise** — your own issuer and Stigmer Enterprise
