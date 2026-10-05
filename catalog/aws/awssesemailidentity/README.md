# AwsSesEmailIdentity

An **Amazon SES (SESv2) email identity** is the verified domain or email address an application is allowed to send mail FROM. A DOMAIN identity is the production shape — it verifies through DNS, signs mail with DKIM, covers every address at the domain, and unlocks a custom MAIL FROM domain for aligned SPF. An EMAIL_ADDRESS identity verifies through a confirmation link — quick for testing, but it cannot carry its own DKIM configuration.

The identity's satellites are folded: the custom MAIL FROM domain, bounce/complaint email forwarding, and named cross-account authorization policies are each their own AWS sub-resource, materialized by the modules so they update independently of the identity itself.

## When to Use

- **Production sending domain** — Verify `example.com` once; every `*@example.com` sender inherits it.
- **DMARC alignment** — Easy DKIM plus a custom MAIL FROM domain give aligned DKIM and SPF.
- **Cross-account sending** — Identity policies grant another account `ses:SendEmail` AS this identity.
- **Shared sending posture** — Attach a configuration set as the identity's default rules.

## When NOT to Use

- To receive email — SES inbound (receipt rules) is a separate surface not modeled by this kind.
- For a domain you cannot publish DNS records for — the identity will stay PENDING and cannot send.

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | — | AWS region (identities are regional). |
| `emailIdentity` | string | Yes | — | Domain or email address to verify. IMMUTABLE. |
| `configurationSet` | ref | No | — | Default configuration set (ref to `AwsSesConfigurationSet` or literal name). |
| `dkimSigning.nextSigningKeyLength` | string | No | RSA_2048_BIT | Easy DKIM key length (`RSA_1024_BIT`/`RSA_2048_BIT`). |
| `dkimSigning.domainSigningPrivateKey` | string (secret) | No | — | BYODKIM private key (base64 PKCS #8). Pairs with the selector. |
| `dkimSigning.domainSigningSelector` | string | No | — | BYODKIM selector you publish the public key under. |
| `mailFrom.mailFromDomain` | string | No | amazonses.com | Custom MAIL FROM subdomain (needs MX + SPF records). |
| `mailFrom.behaviorOnMxFailure` | string | No | USE_DEFAULT_VALUE | `USE_DEFAULT_VALUE` or `REJECT_MESSAGE`. |
| `emailForwardingEnabled` | bool | No | unset | Tri-state: unset leaves the setting unmanaged (a fresh identity gets AWS's default, forwarding on); an explicit true/false pins the position (materializes the feedback sub-resource). SES retains the last-written value per identity name even across identity deletion, so unsetting the field on a previously-managed identity does not restore the default — set true explicitly to turn forwarding back on. |
| `policies[]` | list | No | — | Named cross-account authorization policies (IAM policy JSON). |

## Outputs

| Output | Description |
|--------|-------------|
| `identity_arn` | ARN for identity-policy grants and IAM statements. |
| `email_identity` | The identity string — the join key for composing DNS record names. |
| `identity_type` | `DOMAIN` or `EMAIL_ADDRESS`. |
| `verification_status` | `PENDING` until DNS/mailbox verification completes, then `SUCCESS`. |
| `dkim_tokens` | Easy DKIM's three CNAME tokens — publish each as `<token>._domainkey.<domain>` CNAME `<token>.dkim.amazonses.com` (composable with `AwsRoute53DnsRecord`). |

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSesEmailIdentity
metadata:
  name: prod-sender
spec:
  region: us-west-2
  emailIdentity: example.com
  dkimSigning:
    nextSigningKeyLength: RSA_2048_BIT
```

## Related Kinds

- [AwsSesConfigurationSet](../awssesconfigurationset/README.md) — The default sending rules and event destinations this identity opts into.
- `AwsRoute53DnsRecord` — Publishes the DKIM CNAMEs and the MAIL FROM domain's MX/SPF records.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
