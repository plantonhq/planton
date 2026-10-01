# GcpSslCertificate Guide

The judgment this guide protects: a self-managed certificate is a frozen
artifact with an expiry date and no self-renewal. Everything about it is
immutable in GCP, so the operational skill is not configuring it — it is
rotating it without dropping TLS.

## Rotation is create-before-destroy, always

Every field is ForceNew: any change destroys and recreates the
certificate, and GCP refuses to delete a certificate a proxy still
references (`resourceInUseByAnotherResource`). The safe sequence is
therefore never "edit in place":

1. Create the replacement as a NEW resource with a versioned name
   (`payments-tls-2026-08` — the rotation preset shows the shape).
2. Repoint the proxy's `sslCertificates` list at the new self_link.
3. Destroy the old resource once no proxy references it.

The destroy failing while a proxy still points at the old certificate is
a feature — it fails rather than dropping TLS. `deletionPolicy: ABANDON`
is the mid-rotation escape hatch: it removes the old resource from
management without touching the certificate in GCP, letting cleanup
happen out of band.

## Watch expire_time like a pager, not a dashboard

Nothing here renews itself. The `expire_time` output is parsed from the
uploaded chain — alert on it at least two rotation-cycles early (a
rotation that needs a CA reissue is not a same-day operation). If
hands-off renewal fits the domain setup, prefer
`GcpManagedSslCertificate` (classic, simple) or `GcpCertManagerCert`
(wildcards, pre-issue validation) instead of building a renewal pipeline
around this kind.

## Key hygiene

The private key must be unencrypted PEM (no passphrase) — GCP rejects
encrypted keys. It is the only secret: marked sensitive, encrypted in
Pulumi state and in any OpenTofu state that has an encryption key,
write-only in GCP, never in outputs. The
certificate chain is public handshake material presented to every
client; treating it as a secret only obscures audits. Chain order
matters: leaf first, then intermediates, at most five certificates.

## When self-managed is the right call

Wildcard domains, EV/OV or private-CA issuance, internal load balancers
(no public DNS for managed validation), or serving TLS before a DNS
cutover. If none of those apply, a managed certificate deletes this
kind's entire operational burden.

## Conventions and gotchas

- Global and regional certificates are separate GCP collections with one
  shared name namespace per scope — a name used by a managed certificate
  is taken for this kind too.
- A regional certificate can only be referenced by proxies in its own
  region; the scope decision is permanent.
- `deletionPolicy: PREVENT` suits a certificate whose replacement is not
  yet serving — it blocks the IaC destroy outright.

## Pairs well with

- `GcpTargetHttpsProxy` — presents this certificate; its certificate
  list is where rotation repoints.
- `GcpSslPolicy` — hardens TLS versions and ciphers on the same proxy;
  the certificate is identity, the policy is negotiation.
- `GcpManagedSslCertificate` / `GcpCertManagerCert` — the hands-off
  alternatives when their validation models fit.
