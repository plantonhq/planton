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
rg -n "^## (Backups|Disaster recovery|Restore)" -A 40 <kind-dir>/GUIDE.md        # the judgment
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
  cross-kind judgment; the kind's `GUIDE.md` carries its resource-set tables
  and runbooks; the kind's `presets/` carry the validated starting manifests.
- Never claim a keyless posture the client does not support: the pack
  states it per arm in the field's own doc block (`### spec.backup...keyless`).

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
