# GcpPrivateCaCertificate Guide

The judgment this guide protects: a certificate declared as code is a long-lived, tracked, revocable object -- use this kind for the certificates you would otherwise hand-issue and forget, not for the thousands a mesh mints per hour.

## When this kind, and when the API

Declare a certificate here when a person would otherwise issue it by hand: a load balancer's or server's certificate, a partner's client certificate, a device bootstrap certificate. Workloads that mint many short-lived certificates (a service mesh, SPIFFE) should call the CA Service API directly against a DevOps pool; a manifest per certificate would be the wrong grain, and DevOps certificates cannot be tracked as resources anyway. That is why this kind needs an Enterprise pool.

## CSR or config

- **`pemCsr`** -- the workload generates a key and a CSR; the CSR says who the certificate is for. The pool's and template's policies decide what survives.
- **`config`** -- the manifest states the subject, SANs, and X.509 fields, plus the public key (base64 of the PEM file, e.g. `filebase64("key.pub.pem")`). Use it when the certificate's contents should be reviewed in the manifest.

Either way the private key stays with its owner.

## Renewal and revocation

Everything is immutable, so renewing means issuing a successor: declare a new certificate (a new ID) while the old one is still valid, move the workload, then remove the old one -- which revokes it. Destroying a certificate always revokes it; Google keeps the revoked record for the CRL, and the ID cannot be reused in the pool.

## Lifetimes

The effective lifetime is the shortest of `lifetime`, the pool's and template's maximums, and what remains of the signing authority's own certificate. Plan subordinate rotation so leaves are never cut short unexpectedly.
