# Restricted High-Security Policy

The strictest predefined posture: the RESTRICTED cipher profile with a TLS 1.2 floor. For frontends where security review outranks client reach.

## When to Use

- High-security applications (financial, healthcare, government-facing)
- Frontends whose clients are modern browsers or your own services
- When a security audit demands forward secrecy and AEAD ciphers only

## Key Configuration Choices

- **`profile: RESTRICTED`** — only cipher suites with modern security guarantees; drops some older clients that MODERN still accepts
- **`minTlsVersion: TLS_1_2`** — the natural pairing; RESTRICTED is also the profile GCP requires when the floor is raised beyond what other profiles allow

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the policy will live | GCP Console or `GcpProject` outputs |

## Remix Notes

- Check the `enabled_features` output after deploy — it lists the exact cipher suites GCP enabled, which is what an auditor asks for
- If a legacy client population must keep connecting, start from **01-modern-tls12** instead and tighten later (profile updates in place)

## Related Presets

- **01-modern-tls12** — Broader client compatibility with a TLS 1.2 floor
- **03-custom-cipher-list** — Hand-pick the exact cipher suites
