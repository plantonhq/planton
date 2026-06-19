# Generic OIDC Provider with Explicit Thumbprint

This preset registers any standards-compliant OIDC issuer as a trusted identity provider, with an explicit root-CA thumbprint. Use it for self-hosted or partner issuers whose certificate authority is not in AWS's trusted store.

## When to Use

- A self-hosted OIDC issuer (e.g. an internal Dex/Keycloak) fronted by a private CA
- A partner or vendor issuer whose root CA is not publicly trusted
- Any case where AWS cannot derive the thumbprint automatically

## Key Configuration Choices

- **Explicit issuer URL** (`url.value`) -- the HTTPS issuer endpoint (the `iss` claim)
- **Audience** (`clientIdList`) -- the audience the issuer's tokens carry in the `aud` claim
- **Explicit thumbprint** (`thumbprintList`) -- the 40-character SHA-1 fingerprint of the issuer's root CA certificate

## Placeholders to Replace

- `<aws-region>` -- the AWS region used to configure the provider
- `<https-issuer-url>` -- the issuer URL (must be HTTPS, no query/fragment)
- `<audience>` -- the client ID / audience the issuer's tokens carry
- `<40-char-sha1-thumbprint>` -- the SHA-1 thumbprint of the issuer's root CA (see the AWS guide on obtaining it)

## Note

If your issuer is backed by a well-known public CA, drop the `thumbprintList` field entirely -- AWS validates TLS against its trusted store and derives the thumbprint for you. See presets **01-eks-irsa** and **02-github-actions**.

## Related Presets

- **01-eks-irsa** -- Use instead to enable IRSA on an EKS cluster
- **02-github-actions** -- Use instead for keyless GitHub Actions deployments
