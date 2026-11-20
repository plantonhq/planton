# GcpPrivateCaCertificateTemplate Guide

The judgment this guide protects: a template is where a certificate's shape is decided once, so requests only say who the certificate is for -- never what it may do.

## Template or pool policy

Put what is true of every certificate from a pool in the pool's baseline values; put what differs by use (server, client, code signing) in templates. The two must agree: a template value that conflicts with a pool baseline value fails the request rather than winning, and if the pool stamps an extension the template does not list in `passthroughExtensions`, issuance fails too. When a template is meant for several pools, keep the pools' baselines minimal.

## Identities

`identityConstraints` decides what a request may say about itself. Turn off `allowSubjectPassthrough` when a requested subject should be dropped: the template sets no subject, so the certificate then carries none and its subject alternative names are its whole identity. Keep `allowSubjectAltNamesPassthrough` on with a CEL expression that pins the names -- for example `subject_alt_names.all(san, san.type == URI && san.value.startsWith("spiffe://example.org/"))` for workload identities.

## CA options, carefully

`isCa` is presence-based here as everywhere in the catalog: unset leaves the CA flag out of issued certificates, `false` states CA:FALSE. Leaf templates should set `isCa: false` explicitly, so no request can mint a CA certificate through them.

## Access

Using a template needs `privateca.templateUser` on it -- grant it to the teams or workloads that issue with it, and keep the template in the project that owns the PKI.

## IDs are used once

A deleted template's ID is never usable again in its project and location ("Previously used CertificateTemplate ids may not be reused"), so a template recreated by a pipeline needs a new ID each time.
