# AwsConfigConformancePack — Kind Guide

Authored operational judgment for the conformance pack component: the
design decisions behind the spec's shape, and what to know before
operating packs in production.

## Design decisions

- **Its own kind, not a rule arm.** A pack references no standalone
  Config rules — its template CREATES its own (pack-name-prefixed,
  service-linked-role-managed), so it never belonged on
  AwsConfigRule.
- **One spec, two scopes.** `organization_scope` picks which provider
  resource renders (account pack or organization pack) — the
  AwsConfigRule precedent of one spec branching across scope
  variants. `excluded_accounts` only means something at organization
  scope (CEL-enforced).
- **The template asymmetry is modeled honestly.** Account packs
  accept BOTH template forms at once (AWS prefers the S3 one);
  organization packs accept exactly one (the provider's
  ConflictsWith). Both scopes require at least one — a rule AWS
  enforces server-side even where the provider under-checks it.
- **Template drift is undetectable by AWS design.** Neither provider
  resource reads the template back; imports re-assert it on the first
  apply as a server-side no-op (declared `config_only_attributes` in
  the import catalog).
- **No tags.** Neither provider resource carries a tags argument —
  the one untaggable surface in the Config family.

## Operating conformance packs in production

- **The recorder is a hard prerequisite.** No running recorder in the
  region, no pack — pair with AwsConfigRecorder and create it first.
- **Org packs have naming contracts.** The delivery bucket must begin
  with `awsconfigconforms`, and the pack name caps at 128 characters
  (account packs: 256). AWS enforces both at deploy.
- **Pack rules are pack-owned.** They appear in the rules list
  prefixed with the pack name, but editing or deleting them outside
  the pack fights the service-linked role — change the template
  instead.
- **Deletion is slow at org scope.** Member-account unwinding takes
  minutes; the provider waits up to 20 (its delete timeout).
- **Evaluations bill like ordinary rules.** A pack of 20 rules
  evaluates like 20 rules — scope the recorder's recording group to
  what the rules actually inspect.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
