# KubernetesHarbor live-proven: the registry flagship enters the green E2E matrix

## What changed

- **KubernetesHarbor proven against a live cluster, both engines.**
  All six scenario-engine lanes green with verifier-output evidence on
  every lane: the module-generated admin credential logged in (failing
  loud if the chart's public default ever shipped), a PRIVATE project
  was created, an OCI artifact was pushed and pulled back
  digest-verified through the front-door Service, the API rejected
  unauthenticated calls on an authenticated-only route, and the
  registry surface itself rejected an ANONYMOUS pull of the private
  artifact. The behavioral lane deleted the registry pod after the
  push, waited for a UID-verified replacement, and pulled the artifact
  back byte-identical — blobs survive pod loss through the filesystem
  PVC. The composed-database lane ran the same proof through
  two-replica core/portal/jobservice/nginx against a CloudNativePG
  PostgreSQL wired by reference. Three blind Terraform import
  round-trips re-imported all 12 resources — including the seven
  generated credentials imported by value and the internal-database
  password read from the chart-created Secret — with the follow-up
  plan proposing no real change. Harbor entered both Tier-2 engine
  tiers of the green E2E CI matrix.

- **The Harbor verifier's auth gate now matches Harbor's real security
  model.** Harbor deliberately serves anonymous reads of
  public-project metadata (the projects API carries no authentication
  requirement — verified in the server source at the pin), so
  "any unauthenticated call answers 401" was never a valid gate. The
  gate now probes `GET /api/v2.0/users/current` (RequireAuthenticated
  at the pin) and additionally proves the product boundary: the proof
  project is private, and an anonymous pull of its artifact must be
  rejected by the token service/registry. The docs and README teach
  the operational consequence: project visibility — not credential
  strength — is the anonymous-access boundary; hardening means
  auditing visibility, not only rotating passwords.

- **Port-forward helpers must not report success before the tunnel
  listens.** The Harbor verifier's port-forward now waits (bounded)
  until the local port accepts a TCP connection before returning —
  caught live when a fresh post-replacement tunnel was dialed before
  kubectl bound the port and the durability pull read
  connection-refused. A "started" tunnel is not a listening tunnel;
  retry loops downstream had been absorbing the race invisibly.

- **The import framework gained the `import_normalized` vocabulary for
  values that cannot round-trip by provider construction.** A
  kind's import map may now declare, per module resource, dotted
  attribute sub-paths whose post-import plan update is the documented
  adoption shape — each with a mandatory reason. The round-trip oracle
  tolerates exactly those sub-paths and still fails on any sibling
  drift; the conformance guard enforces real resource names, dotted
  paths, and non-empty reasons. The canonical (and first, live-proven)
  case: a Secret data key wiring `random_password`'s `bcrypt_hash` —
  the random provider recomputes that hash with a fresh salt on
  import, so the first post-adoption apply rewrites the key to an
  equivalent hash of the same password, a functional no-op. Harbor's
  registry htpasswd line declares it; the update rule teaches the
  salted-hash class and why neither a provider-wide tolerance nor
  ignore_changes on Secret data is the answer.

- **The full-surface scenario joins its fixture's namespace.** The
  composed-database scenario had claimed namespace creation in the
  namespace its PostgreSQL fixture owns and creates first — every
  Terraform lane would have failed with AlreadyExists. The scenario
  now joins the fixture-owned namespace, matching the convention every
  fixture-sharing scenario follows.

## Why

A container registry is trusted with the artifacts a team ships;
"install succeeded" is not evidence anyone can push, pull, or survive
a crash. These lanes prove the product loop the way a customer uses it
— authenticated push/pull with digest verification, anonymous access
rejected at both the API and the registry surface, and artifact
durability through pod loss — on both engines, with import recipes
proven correct by blind re-import rather than review.
