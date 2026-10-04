# AwsOrganization — Kind Guide

Authored operational judgment for the organization kind: the
design decisions behind the spec's shape, and what to know before
operating an organization in production.

## Design decisions

- **This kind IS the management-account act.** Deploying it from an
  account performs CreateOrganization there — the account becomes the
  management account. There is exactly one organization per account;
  the kind models that singleton honestly (no name, no multiples).
- **Service access lives HERE, nowhere else.** The provider ships both
  an `aws_service_access_principals` argument and a standalone
  service-access resource, and its own docs warn that using both
  produces a perpetual diff. The catalog gives the config one home —
  this spec — and records the standalone resource as composed.
- **Delegated administrators fold as immutable pairs.** A registration
  is `{account, service}` and nothing else; both leaves ForceNew, so a
  change re-registers (deregister + register). The registered account
  must already be a member.
- **The resource policy folds as a singleton arm.** AWS keeps ONE
  resource policy per organization (PutResourcePolicy upserts), so a
  standalone kind would be an instances-fight-over-one-object defect —
  the settings-singleton lesson applied inside a fold.
- **Root-access management folds here, not into the IAM settings
  kind.** IAM's organization features (RootCredentialsManagement /
  RootSessions) are a management-account act requiring
  iam.amazonaws.com trusted access — a field THIS spec models, so the
  CEL wires the dependency; folding it into an account-local settings
  kind would strand a permanently management-account-only arm there.
  Destroying the arm disables every enabled feature (member-account
  root credentials become locally manageable again).
- **All-features gates are validation, not surprises.** Trusted
  access, policy types, delegated admins, and the resource policy all
  require `featureSet: ALL` at AWS — four CELs front-load what would
  otherwise fail at apply.

## Operating an organization in production

- **The feature-set downgrade is an organization rebuild.** ALL →
  CONSOLIDATED_BILLING replaces the resource, which means deleting the
  organization — only legal once every member, OU, and policy is gone.
  Treat it as a migration project, never a settings change.
- **Destroy ordering is structural**: DeleteOrganization fails while
  members/OUs/policies exist. Tear down policies and OUs (their own
  components) and remove accounts first.
- **Enabling a policy type here is the gate** for every
  [policy component](../awsorganizationpolicy) attachment of that
  type — enable SERVICE_CONTROL_POLICY before shipping SCPs.
- **Trusted access lets services create service-linked roles in member
  accounts** — enable principals deliberately; AWS recommends the
  owning service's own console/API for services with extra setup
  (the provider carries the same caution).
- **Contacts/regions on member accounts need
  `account.amazonaws.com`** in the service-access list (the
  [member-account component](../awsorganizationaccount)'s settings
  arms depend on it).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
