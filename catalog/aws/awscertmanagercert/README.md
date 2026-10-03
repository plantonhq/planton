# AwsCertManagerCert

The **AwsCertManagerCert** resource provisions an AWS Certificate Manager (ACM) certificate in any of ACM's three creation modes — requested (Amazon-issued), imported (bring-your-own material), or private (issued by an ACM Private CA) — and, for DNS-validated certificates whose zone lives in Route53, automates the validation records and the issuance wait.

## Creation Modes

- **Requested (the default)** — ACM issues a public certificate for domains you prove ownership of, via `DNS` (recommended) or `EMAIL` validation. Set `primary_domain_name` (plus any `alternate_domain_names`). Amazon-issued certificates renew automatically as long as the validation records stay in place.
- **Imported** — bring certificate material issued by an external CA and let ACM distribute it to integrated services. Set the `imported` block (`certificate_body`, `private_key`, optional `certificate_chain`). Imported certificates never auto-renew; re-importing new material before expiry updates in place and keeps the same ARN.
- **Private (ACM-PCA)** — your AWS Private Certificate Authority issues the certificate directly, with no public validation. Set `primary_domain_name` together with `certificate_authority_arn`.

Exactly one of `primary_domain_name` / `imported` drives the mode; validation rules keep the arms honest so a manifest cannot mix them.

## DNS Validation: Managed or External

- **Managed (Route53)** — set `route53_hosted_zone_id` (a literal zone ID or a reference to an `AwsRoute53Zone`). The module creates the validation CNAMEs in the zone and, unless `wait_for_validation` is `false`, waits for the certificate to reach `ISSUED`.
- **External DNS** — leave `route53_hosted_zone_id` unset. The certificate is created in `PENDING_VALIDATION` and the exact records to create are exported as the `domain_validation_records` output. Once you create them, ACM issues — and keeps renewing — automatically.

## Key Spec Fields

- **`region`** — where the certificate lives. Regional consumers (ALB, API Gateway, OpenSearch) need it in their own region; **CloudFront only accepts certificates from `us-east-1`**.
- **`primary_domain_name`** / **`alternate_domain_names`** — the domains covered; wildcards (`*.example.com`) cover one label level. A common pairing is the apex plus its wildcard.
- **`validation_method`** — `DNS` (recommended; renewals are automatic), `EMAIL` (re-approval on every renewal), or `HTTP` (serve a validation token — for domains whose DNS you cannot touch). Empty keeps DNS.
- **`key_algorithm`** — `RSA_2048` (default), `RSA_3072`, `RSA_4096`, `EC_prime256v1`, `EC_secp384r1`, `EC_secp521r1`. Create-time immutable.
- **`options`** — Certificate Transparency logging preference (`ENABLED`/`DISABLED`) and exportability (`export: ENABLED` allows exporting the private key for use off-AWS, at an additional AWS charge).
- **`validation_options`** — per-domain overrides of where the validation request is sent (e.g. EMAIL-validating a subdomain at its parent domain).
- **`early_renewal_duration`** — for PRIVATE (ACM-PCA) certificates only: start ACM's managed renewal this long before expiry (`P90D` or `2160h`; durations under 60 days have no effect). The certificate ARN stays stable across renewals. Publicly validated certificates renew on ACM's own schedule, and imported certificates never renew.

## Outputs

- **`cert_arn`** — the join key every TLS consumer references (load-balancer listeners, CloudFront, Cognito, OpenSearch, Client VPN).
- **`status`** — `PENDING_VALIDATION` until ownership is proven, then `ISSUED`.
- **`domain_validation_records`** — the DNS records proving ownership (create these in external DNS when the module does not manage them; keep them in place so renewals stay automatic).
- **`not_before`** / **`not_after`** — the validity window; `not_after` is the re-import deadline for imported certificates.
- **`certificate_type`** — `AMAZON_ISSUED`, `IMPORTED`, or `PRIVATE`.

## Deliberately Not Modeled

- **Write-only private-key arms** (`private_key_wo`) — the `imported.private_key` field is marked sensitive and handled by the platform's secret machinery; a second write-only arm would be redundant surface.

## Important Notes

- Wildcard certificates require DNS validation; EMAIL validation does not support them.
- The Route53 zone must be authoritative for **every** domain on the certificate.
- A certificate swap is create-before-destroy: the replacement is issued before the old ARN is released, so consumers holding the reference are never left dangling.

## References

- [AWS Certificate Manager Documentation](https://docs.aws.amazon.com/acm/)
- [DNS Validation for ACM Certificates](https://docs.aws.amazon.com/acm/latest/userguide/dns-validation.html)
- [Importing Certificates into ACM](https://docs.aws.amazon.com/acm/latest/userguide/import-certificate.html)
- [AWS Private Certificate Authority](https://docs.aws.amazon.com/privateca/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
