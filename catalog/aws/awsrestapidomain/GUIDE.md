# AwsRestApiDomain — Kind Guide

Authored operational judgment for the REST API custom-domain kind:
the design decisions behind the spec's shape, and what to know before
fronting APIs with your own hostname.

## Design decisions

- **A domain is not an API field.** One hostname maps many APIs and
  outlives any of them; putting it on AwsRestApiGateway would force
  every mapping to share an API lifecycle.
- **`routing_mode` is modeled; routing rules are not.** Rules are an
  API Gateway v2 surface that also attaches to v1 domains, and they
  already live on AwsHttpApiDomain. This component exposes the v1
  knob that chooses between base-path mappings and those rules.
- **DNS stays outside.** The domain exists independently of any
  record pointing at it. Alias targets and zone IDs are outputs so
  AwsRoute53DnsRecord can compose them.
- **Certificate source is exactly one.** ACM in-region for REGIONAL /
  PRIVATE, ACM in us-east-1 for EDGE — the modules wire the matching
  provider argument (EDGE and PRIVATE use `certificate_arn`, REGIONAL
  uses `regional_certificate_arn` — AWS's own argument split).
- **Uploaded certificate material is a legacy path.** AWS's docs are
  ambiguous about which endpoint types accept direct uploads (the SDK
  says edge-or-private, the provider shows a regional example); the
  component permits it everywhere and lets AWS arbitrate. Prefer ACM.

## Running custom domains in production

- **REGIONAL is the default for new work.** EDGE still exists for
  global CloudFront distribution, but it requires the certificate in
  us-east-1 and a longer create.
- **Keep the root mapping.** An empty `base_path` (the `(none)`
  output key — AWS's own empty-path sentinel) is how
  `https://api.example.com/` reaches a stage; forgetting it 404s the
  hostname itself.
- **`endpoint_access_mode` pairs with modern security policies.** AWS
  supports it only with the `SecurityPolicy_*` family — the spec
  enforces the pairing so a legacy `TLS_1_2` domain never carries a
  dead access-mode setting.
- **PRIVATE domains need access associations.** Without them the
  hostname exists and nobody inside the VPC can call it.
- **Do not rotate the hostname in place.** Changing `domain_name`
  replaces the domain; stand up the new name, move DNS, then destroy
  the old one.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
