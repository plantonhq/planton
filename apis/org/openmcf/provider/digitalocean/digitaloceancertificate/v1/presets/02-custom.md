# Custom Certificate

This preset creates an SSL certificate from user-provided PEM content. Use when you have a certificate from an enterprise CA, a purchased certificate, or a certificate issued outside of Let's Encrypt. The leaf certificate, private key, and optional intermediate chain are supplied directly.

## When to Use

- Enterprise or purchased SSL certificates
- Certificates from private CAs or internal PKI
- Wildcard or EV certificates not available via Let's Encrypt
- Migrating existing certificates to DigitalOcean

## Key Configuration Choices

- **Custom type** (`type: custom`, `custom`) -- you supply the PEM content; no auto-renewal.
- **Leaf certificate** (`leafCertificate`) -- PEM-encoded server certificate; required.
- **Private key** (`privateKey`) -- PEM-encoded private key; must match the leaf certificate; keep secure.
- **Certificate chain** (`certificateChain`) -- optional; include intermediate(s) if your CA provides them (e.g., DigiCert, Sectigo).
- **No auto-renewal** -- you must manually replace the certificate before expiry; consider external secret management (e.g., Vault, sealed secrets).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<paste-your-leaf-certificate-pem>` | PEM content of the leaf/server certificate | Your CA or certificate issuer |
| `<paste-your-private-key-pem>` | PEM content of the private key | Generated with the cert or from your CA |
| `<paste-intermediate-chain-pem-optional>` | PEM content of intermediate certificates | Your CA; often included in the cert package |
| `my-custom-cert` | Human-readable certificate identifier | Choose a name; used when referencing in load balancers |

## Related Presets

- **01-lets-encrypt** -- Use when you can use Let's Encrypt for free auto-renewal
