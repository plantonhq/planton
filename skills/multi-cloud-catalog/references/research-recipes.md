# Research Recipes

The concrete command moves for catalog research. Every recipe runs from
`<pack-root>` (see `references/pack-layout.md`). They work because the pack
renders a fixed heading vocabulary on every page -- documented in
`reference-commons.md` and pinned by the catalog's own contract tests, so
these patterns hold across every component.

## Find components by name or capability

```
rg -il "kafka" -g 'reference*.md' -g 'GUIDE.md' -g '_patterns/*.md' .
```

The `-g` globs scope the search to pack files -- in a repo checkout the same
directories also hold protos, generated code, and IaC modules, and an
unscoped search drowns the answer in those (inside a skill mount's
`components/` the globs cost nothing and keep the recipe portable). Then
shortlist through the provider's `reference-index.md` -- each row is
`[Kind](path) | purpose | example? | guide?`. Full-text search matters more
than it looks: compatible alternatives document the well-known names they
substitute for in their own pages, so searching the name the user said finds
the alternative even when no component carries that name. When that happens,
follow the substitution workflow in `_docs/GUIDE.md` beside the root index --
propose openly, never silently.

## One component's required fields

Read `## Example` first -- it is a validated manifest, the fastest picture
of the component's real shape. Then the spec table:

```
rg "^## Spec Fields" -A 40 <page>
```

Columns: `Path | Type | Required | Default | References`. A `References`
cell names the kind (and field path) that field points at by default.

## One field, fully

```
rg "^### spec.replication" -A 12 <page>
```

Each field has its own `### <dotted path>` block under `## Field Details`:
docs, validation rules, and allowed enum values. Enum lists sit BELOW the
field's doc prose -- when a list looks truncated, widen `-A` rather than
concluding the values are missing.

## What a component exports

```
rg "^## Outputs" -A 20 <page>
```

Output paths are spelled snake_case (`status.outputs.vpc_id`) -- that is the
canonical `fieldPath` spelling, while spec YAML keys are camelCase; the
asymmetry is explained once in `reference-commons.md`.

## Wiring: both directions

- Outbound (what K's fields can point at): K's `## References` table.
- Inbound (who can point at K): K's `## Referenced By` table --
  `Kind | Field | Reads`.
- Catalog-wide, without opening pages:

```
rg 'to: "KubernetesValkey"' _docs/reference-graph.yaml      # every field that can target it
rg -A 3 'from: "AwsEcsService"' _docs/reference-graph.yaml  # everything it can reference
```

## What a component costs, enforces, and needs (the fact-sheets)

Covered components carry three sidecars plus a generated estimate document;
all are small YAML files meant to be read whole:

```
cat aws/awsalb/cost.yaml                    # billing model, baseline charges,
                                            # the spec fields that move the bill
cat _pricing/estimates/awsalb.yaml          # per-preset dollar estimates:
                                            # exact quantities, unit prices,
                                            # source URL + verification date
cat aws/awsalb/controls.yaml                # posture on every catalog control
cat aws/awsalb/iac/permissions.yaml         # least-privilege runner manifest
```

- A missing file means the release publishes no verified data for that kind
  yet -- say "not yet published", never $0 and never a memory-recalled rate.
- Estimate money fields (`list_unit_price`, `list_cost`, totals) are exact
  decimal strings -- quote them verbatim; each line's `price_source` +
  `retrieved_on` is the citation to hand the user.
- Resolve `control_id` values against `_compliance/controls-catalog.yaml`
  (names + statements); framework questions read
  `_compliance/frameworks/<framework>.yaml` on top -- check the crosswalk's
  `spec.providers` first, and never apply a provider-scoped framework to
  another provider's component. Posture is never "compliant" -- see the
  honesty grammar in SKILL.md.
- Which kinds are covered at a glance:

```
rg -l 'kind: ComponentCostProfile' -g 'cost.yaml' .
```

## "Can I back this up, and get it back?" (stateful kinds)

A stateful kind's backup story is a COMPOSITION, never one field, and the
pack answers it in three reads:

```
rg -n "^## Spec Fields" -A 80 <page> | rg -i "backup|restore|recover|bootstrap"   # the kind's own blocks
rg -n "^## References" -A 30 <page>                                              # the store, identity, and token kinds it wires to
rg -n "^## (Storage engine|Backups|Disaster recovery|Restore)" -A 40 <kind-dir>/GUIDE.md   # the judgment (a kind with more than one engine decides the story there first)
```

- The kind's backup block names the store in that store's OWN vocabulary
  (S3, GCS, Azure Blob, or Cloudflare R2), each arm a `StringValueOrRef`
  onto the catalog's bucket, identity, and token kinds -- read the
  `References` column, never invent an endpoint or a region for R2.
- The credential posture is per arm and exactly one: keyless where the
  cluster's cloud allows it (an identity kind, referenced) or declared keys
  (a key exported by a catalog kind, referenced). R2 has NO keyless posture
  anywhere; its credential is a `CloudflareAccountApiToken`, referenced as
  the S3 key pair.
- The restore is a SECOND declared instance against the same store, and each
  kind names the one step that stays with the operator (a seal key held on
  both sides and an init token for the vault; the source's credential Secret
  for the databases). `_patterns/stateful-kind-disaster-recovery.md` is the
  cross-kind judgment and embeds validated manifests; the kind's `GUIDE.md`
  carries its resource-set tables, runbooks, and day-2 operations. Presets
  are named there by slug but do not travel in the pack
  (`pack-layout.md`, "What the pack does not carry").
- Never claim a keyless posture the client does not support: the pack
  states it per arm in the field's own doc block (`### spec.backup...keyless`).
- Read the operational truths the module enforces or prints from the
  reference page before promising anything: the `KubernetesOpenBao` page
  states the name budget that applies once `backup` is declared (in the
  spec's own header), the pod state a restore Job shows while it waits for
  the operator's token (`### spec.restore.rootToken`), what happens when an
  install fails part-way and how to recover (`### spec.restore`), the TLS
  name the jobs need (`### spec.tls.certSecretName`), and the two KMS roles a
  Cloud KMS seal identity needs (`### spec.autoUnseal.gcpKms`).

Before writing a `restore` for the vault, start from the complete restore
target the guide embeds under "Restore on the bad day" and run this
checklist against the manifest -- every line is a rule the pack states, and
a restore that violates one fails on the bad day, not at validation:

0. The source is a Raft vault. Snapshots exist only for integrated Raft
   storage; a vault on PostgreSQL storage has none, refuses the `backup`
   block, and comes back when its database is restored — its checklist is
   the database kind's, and the guide's engine section says which vault a
   customer has before any of the steps below apply.
1. Same seal key on source and target (`autoUnseal` points at the same KMS
   key or the same transit key on the same key holder); a Shamir vault
   restores only by the guide's manual runbook.
2. The target declares the SOURCE's `backup` block -- same store, same
   `objectStore.prefix` -- so it can read the snapshots; `restore` without
   `backup` is refused.
3. `latest: true` only when the source is gone; beside a live source (a
   clone, a rehearsal) name the `snapshotKey`, and give the clone its own
   prefix in the same apply that removes `restore`.
4. Backups are suspended while `restore` is declared; the closing step is to
   remove the block and apply again, then delete the token Secret.
5. The restored state carries the source's login role bound to the source's
   ServiceAccount name and namespace; a target under a different name or
   namespace re-runs the four-command recipe.
6. Never delete the finished restore Job to tidy up -- a changed
   declaration is how a restore runs again.

### Self-hosted Planton: one archive for records and secrets

A self-hosted Planton (`KubernetesPlantonPlatform`) bundles its own vault,
and that vault stores INSIDE the platform's database -- so "can we get our
secrets back?" is answered by the platform's one backup block, never by a
second archive. Three reads:

```
rg -n "^### spec.database.postgresql.(backup|recoverFrom)" -A 12 <platform-page>   # the archive and the restore
rg -n "^### spec.vault" -A 14 <platform-page>                                      # what opens the restored vault
rg -n "^## (Disaster recovery|Restore)" -A 60 kubernetes/kubernetesplantonplatform/GUIDE.md   # the resource sets and the runbook
```

The checklist -- every line is a rule the pack states, and the platform's
guide names the sentence the operator prints when a line is broken:

1. A `backup` with the vault enabled needs `vault.autoUnseal` or
   `vault.initSecretName`; the kind refuses the declaration with neither
   (the archive would carry every secret and no way to open them).
2. A cloud seal is the standalone vault's four arms, byte for byte; on
   Google Cloud the seal identity needs `roles/cloudkms.cryptoKeyEncrypterDecrypter`
   AND `roles/cloudkms.viewer`, and the key and grants exist BEFORE the
   platform because the seal is checked when the server starts.
3. A keyless platform on GKE carries TWO Workload Identity bindings: the
   vault's on `<platform>-openbao`, the database's on `<platform>-postgres`;
   the vault's identity travels by reference on the arm, the database's is a
   literal email in `backup.serviceAccountAnnotations`.
4. `vault.initSecretName` is the one object no archive carries: the
   operator writes it once, never deletes it, and a namespace the
   declaration owns takes it along -- the copy outside the cluster is the
   runbook's first step.
5. On the bad day, under the built-in seal recreate that Secret before (or
   within minutes of) declaring `recoverFrom`; under a cloud seal the
   restored vault opens itself and the Secret is break-glass you recreate
   after. A restore that finds no Secret is refused in one sentence naming
   the archive, the Secret, its keys, and the step.
6. The seal is decided at creation: a declaration whose seal differs from
   the archive's is refused before anything renders.
7. Read `status.backup.vault` (`covered`, `seal`, `initSecretName`, a
   sentence) before promising anything about what comes back.

## Where judgment has been written

A page that has authored wisdom links it in its head:

```
rg -l '^\*\*Guide\*\*:' kubernetes/            # every guided kind in a provider
```

The per-provider index carries the same signal as its Guide column, and
`_patterns/` (its `README.md` is the list) holds the multi-component recipes
-- each pattern declares the kinds it composes and embeds validated
manifests.

## Craft notes, learned the hard way

- Prefer anchored heading greps (`rg "^## Outputs" -A 20`) and then reading
  the section over regexing markdown table cells -- cell patterns are
  whitespace-sensitive and silently miss rows.
- `-l` before content: when a search may hit many pages, list files first
  (`rg -il <term> .`), pick from the index, then read one page deeply.
- The `## Example` manifest is validated against the schema at generation
  time -- trust it as a starting shape, then adjust against `## Spec Fields`
  and `## Validation Rules`.
