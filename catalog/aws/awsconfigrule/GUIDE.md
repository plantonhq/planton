# AwsConfigRule — Kind Guide

Authored operational judgment for the Config rule kind: the
design decisions behind the spec's shape, and what to know before
operating compliance rules in production.

## Design decisions

- **One kind, four provider resources.** Account rules and the three
  organization rule variants are the same noun (a compliance rule) at
  two scopes — the spec discriminates by source arm + `organization`
  presence, and the modules pick the provider resource. The name cap
  drops from 128 to 64 characters when `organization` is set.
- **Exactly one source arm, enforced early.** Managed / custom-lambda
  / custom-policy are mutually exclusive in AWS; the spec's CELs fail
  the manifest before the cloud sees it, including the org-scope
  rules AWS only rejects at apply (ScheduledNotification on Guard
  rules, debug-log accounts on non-Guard rules, remediation on org
  rules).
- **Remediation folds here** because it attaches by RULE NAME (a real
  structural edge) — unlike conformance packs, which create their own
  rules from templates and ship as their own kind.
- **`rule_name` is an output on purpose** — remediation attachments,
  aggregator queries, and the import recipe all address rules by name.

## Operating compliance rules in production

- **No recorder, no rule.** AWS rejects PutConfigRule in a region
  without a configuration recorder
  (NoAvailableConfigurationRecorderException), and a rule under a
  STOPPED recorder evaluates nothing silently. Deploy
  AwsConfigRecorder first and keep its scope covering the types your
  rules evaluate.
- **Custom Lambda rules need the invoke grant BEFORE create** — a
  `lambda:InvokeFunction` permission for `config.amazonaws.com` on the
  function (AWS validates it at PutConfigRule).
- **Guard policy rules are the cheap custom path.** No compute to
  operate; the engine evaluates on configuration changes only (AWS
  rejects scheduled triggers for org Guard rules — the spec encodes
  it).
- **Automatic remediation is a loop; bound it.** AWS requires the
  retry contract (attempts + window); set the error-percentage
  circuit breaker so a broken SSM document stops instead of
  remediating the fleet into an outage. Start with `automatic: false`
  and graduate.
- **Organization rules deploy from the management or delegated-admin
  account** and land in member accounts asynchronously (the provider
  waits with 20-minute timeouts). Exclusions are per-account-id.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
