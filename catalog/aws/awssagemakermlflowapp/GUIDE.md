# AwsSagemakerMlflowApp — Kind Guide

Authored operational judgment for the serverless MLflow app kind:
the design decisions behind the spec's shape, and what to know before
running serverless MLflow in production.

## Design decisions

- **The app and the tracking server are two AWS products, not one.**
  This kind is the SERVERLESS MLflow 3.x deployment — billed per use,
  no idle charge, nothing to size. `AwsSagemakerMlflowServer` is the classic
  dedicated-capacity server billed hourly with ~25-minute lifecycle
  operations. The app is NOT a tracking-server satellite: it does not
  attach to a server; it stands alone and associates with SageMaker
  DOMAINS. Two kinds keep that split visible instead of burying it in
  a mode flag.
- **The ARN is the identity.** All API operations key on `app_arn` —
  the name updates in place, so `metadata.name` can derive the AWS
  name without a rename ever forcing a replacement.
- **`role_arn` is the ONE replace-on-change field.** Everything else —
  name, artifact store, domains, registration mode, window — updates
  in place. The spec defaults the role to an `AwsIamRole` reference.
- **`default_domain_ids` are foreign keys to `AwsSagemakerDomain`.**
  Domain association is the app's distribution mechanism — Studio
  users in associated domains get it as their default MLflow — so the
  spec lets manifests compose domains by reference or pass literal
  IDs.
- **`model_registration_mode` uses AWS's own enum strings**
  (`AutoModelRegistrationEnabled` / `AutoModelRegistrationDisabled`) —
  unlike the tracking server's boolean, this surface toggles both ways.

## Running serverless MLflow in production

- **Default to the app for new deployments.** Per-use billing with no
  idle charge beats a dedicated server's always-on hourly meter for
  all but sustained heavy tracking — and there is no capacity to
  outgrow.
- **Get the role right the first time.** `role_arn` is the one change
  that replaces the app; everything else is an in-place edit.
- **One account default, ever.** `account_default_status: ENABLED` is
  account-global — only one app per account can hold it. Decide which
  app owns the default before two manifests fight over it.
- **Associate domains instead of circulating URIs.** Studio users in
  `default_domain_ids` domains track to the app automatically — that
  is the intended distribution path.
- **A deleted app may still be visible in AWS.** Deletion is a soft
  delete: an app in DELETED status reads as absent — seeing it in a
  raw API listing does not mean the delete failed.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
