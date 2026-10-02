# Allowlisted Device Certificates

This preset creates a trust config that trusts no CA at all and instead accepts a short list of individual certificates — for self-signed device certificates or a single partner certificate you do not want to trust a whole CA for.

## When to Use

- A handful of devices or partners present self-signed certificates
- You want a certificate that is not on the list to fail, with no CA able to mint new ones that pass
- A stopgap while clients move to certificates from a private CA

## Key Configuration Choices

- **No trust store** — only the listed certificates pass validation
- **One entry per certificate** — adding or removing a device is an in-place update
- **Still verified** — a listed certificate must parse, prove possession of its private key, and satisfy its SAN constraints

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<device-certificate-pem-body>` | The base64 body of the device's certificate | The device's provisioning record |

## Related Presets

- **01-private-root-trust-store** — trust every certificate a private CA issues
