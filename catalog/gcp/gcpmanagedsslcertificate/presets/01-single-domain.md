# Single-Domain Load Balancer Certificate

The most common Google-managed SSL certificate: one fully-qualified domain name for an application served through a global external HTTPS load balancer.

## When to Use

- A single hostname (e.g. `app.example.com`) behind a global HTTPS load balancer
- Standard production TLS where Google handles issuance and renewal
- Attaching a managed cert to a target HTTPS proxy without managing key material

## Key Configuration Choices

- **One domain** in `domains` — Google-managed certificates do not support `*.` wildcards; list each hostname explicitly
- **Optional `certificateName`** — defaults to `metadata.name` when omitted; set explicitly when the GCP name must differ from the Planton resource name
- **`description`** — documents what the cert secures; helpful when debugging PROVISIONING status in the console

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the certificate will live | GCP Console or `GcpProject` outputs |
| `app.example.com` | The hostname the certificate should secure | Your DNS / load balancer hostname |

## Remix Notes

- Point DNS for the domain at the load balancer's global IP (`GcpGlobalAddress`) before expecting provisioning to complete
- Reference this certificate's `self_link` from a target HTTPS proxy's `sslCertificates` list
- To rotate domains, create a new certificate first and repoint the proxy before destroying the old one (create-before-destroy)

## Related Presets

- **02-multi-domain** — Cover apex and `www` (or several subdomains) on one certificate
- **03-explicit-name** — Separate Planton resource name from the GCP certificate name
