# Inline PFX Certificate

Upload a certificate directly as a base64-encoded PKCS#12 bundle -- for certificates that live outside Key Vault (a CA-issued file, an org PKI export).

## When to Use

- CA-issued certificate files not managed in Key Vault
- Quick preproduction TLS (Container Apps accepts self-signed uploads; browsers still reject them)

## Key Configuration Choices

- Both fields are secret-bearing -- reference managed secrets (`$secret/...`); the blob bundles the private key
- Encode the PFX with `base64 -i certificate.pfx` and store the output as the secret value
- Rotation is manual and replaces the resource (briefly re-binding its domains) -- prefer the Key Vault preset for certificates that renew

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `app.example.com` | Replace with the certificate's name on the environment | The custom domain you are binding |
| `<your-environment>` | The AzureContainerAppEnvironment resource's metadata name | Your Planton resource inventory |
| `<your-pfx-secret>` / `<your-pfx-password-secret>` | Managed-secret slugs for the base64 PFX and its password | Your secrets manager |

## Related Presets

- `01-key-vault-certificate` -- the renewal-following production posture
