# GcpCertManagerTrustConfig Guide

The judgment this guide protects: a trust config decides who gets in. Every CA you add as a trust anchor admits every certificate that CA has ever signed or will sign. Choosing what goes into the trust store is an access-control decision, not a certificate chore.

## Where a trust config sits

A trust config does nothing on its own. The chain for mutual TLS on an Application Load Balancer is:

1. This trust config — the CAs and allowlisted certificates to believe.
2. A server TLS policy whose mTLS client-validation settings name the trust config's `trust_config_id`.
3. The `GcpTargetHttpsProxy` whose `server_tls_policy` attaches that policy.

The same trust config can also back a backend authentication config, which validates the certificates your backends present. It never attaches to a `GcpCertManagerCert`: the certificate is what the load balancer shows clients, and the trust config is how it checks theirs.

## Anchor a CA, or allowlist a certificate

- **Trust anchor** — use when one CA signs a population of clients (your private CA, a partner's CA). New clients get in the moment the CA signs them, with no change here.
- **Intermediate CA** — add when clients send only their leaf certificate. The load balancer uses it to complete the chain up to an anchor.
- **Allowlisted certificate** — use for a few clients that chain to nothing, like self-signed device certificates. Each one is accepted individually, so a certificate you did not list never gets in.

Prefer the narrowest choice that works. Anchoring a public CA would admit anyone who can buy a certificate from it.

## Rotation is an overlap, not a swap

Every change to a trust config is an in-place update. Rotate a root in three deploys:

1. Add the new root beside the old one.
2. Move clients to certificates from the new root.
3. Remove the old root.

Removing the old root first locks out every client still holding its certificates.

## Location follows the load balancer

`global` (the default) serves global external and cross-region internal Application Load Balancers. A regional load balancer needs a trust config in the same region. Location is immutable, so pick it with the load balancer, not after.

## Conventions and gotchas

- Certificates are public material. Never put a private key in this resource.
- Google allows one trust store per config today, so `trustStores` holds at most one entry.
- The provider marks trust-store certificates sensitive, so both engines keep them out of plans and logs even though nothing in them is secret.
- Google refuses to delete a trust config while a TLS policy references it. Set `deletionPolicy: PREVENT` on one that serves live traffic.

## Pairs well with

- `GcpTargetHttpsProxy` — attaches the server TLS policy that names this trust config.
- `GcpCertManagerCert` — the server certificate on the same load balancer.
- `GcpPrivateCaPool` — a private CA whose root becomes the trust anchor for the client certificates it issues.
