# AwsBedrockAgentCoreRuntime — Kind Guide

Authored operational judgment for the AgentCore runtime kind: the
design decisions behind the spec's shape, and what to know before
hosting agents in production.

## Design decisions

- **The runtime name is an explicit spec field.** AWS's charset (letter
  first, then letters/digits/underscore) rejects hyphens, which platform
  names use routinely — deriving from `metadata.name` would fail most
  manifests at apply. `spec.runtime_name` validates the exact charset at
  manifest time instead.
- **The artifact is a two-arm union, replace-on-switch.** Exactly one of
  `container`/`code` (CEL-guarded); the provider forces replacement when
  the arm switches and versions in place otherwise. The code arm's
  single-member `code.s3` wrapper flattens to one S3 message.
- **The resource policy folds in, scoped to the runtime's own ARN.**
  The provider resource accepts any AgentCore ARN; this kind applies it
  to the runtime it deploys (the secret-policy fold precedent). Sibling
  kinds gain their own policy arms on demand signals.
- **Never author a `Resource` member in policy statements.** AWS requires each statement's Resource to be exactly the runtime's own ARN — an AWS-suffixed value that does not exist until create (`PutResourcePolicy` rejects anything else, wildcards included; live-verified 2026-08-14). Both modules inject the deployed runtime's ARN into every statement, so authors write only Effect/Principal/Action/conditions.
- **Endpoints are name-keyed satellites with no separate ID.** An
  endpoint's AWS identity IS its name — the `endpoint_arns` output map
  and the import composite (`{agent_runtime_id},{name}`) both key on it.
- **The evaluations family deliberately lives elsewhere.** Evaluators,
  harnesses, and online evaluation configs are standalone AWS resources
  with no structural runtime edge — they are their own component, not
  runtime arms.

## Running agent runtimes in production

- **Ship changes through versions, promote through endpoints.** Edit the
  spec freely — sessions in flight finish on their version; live traffic
  moves only when a floating endpoint picks up the new version or you
  re-point a pinned one.
- **The role is the agent's blast radius.** The runtime role needs
  artifact access (ECR pull or S3 read) plus whatever AWS APIs your
  agent calls — scope it per agent, never share one broad role.
- **PUBLIC mode still gates data-plane calls** by IAM/JWT; it controls
  the session's OUTBOUND network only. Use VPC mode when the agent must
  reach private resources.
- **JWT and IAM auth compose.** Omitting `custom_jwt_authorizer` leaves
  SigV4 (IAM) callers only; adding it admits OIDC bearer tokens matching
  your audience/client/claim rules.
- **Session scratch is ephemeral by design.** `session_storage` mounts
  vanish when the session ends; use EFS access points for durable
  cross-session state.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
