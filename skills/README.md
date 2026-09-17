# Planton Skills

This tree is the source of truth for the skill definitions that power
Planton's AI assistant surfaces. A skill packages the judgment, workflow,
and working craft an agent loads at runtime — what to do, in what order,
and how to decide — as a `SKILL.md` instruction file plus on-demand
reference documents.

Agent identity definitions (the instructions an agent carries independent
of any one skill) live in the sibling `agents/` tree, one directory per
agent slug.

## The one-skill doctrine

The Planton Assistant has ONE skill: `planton`. It carries both product
domains — infrastructure (charts, manifest sets, deployed projects) and
service delivery (registration, deploys, CI/CD, the offline lane) — as
domains of a single craft with a single doctrine: one "Know your
instruments" ladder, one set of invariants, one prohibitions list. Do not
spin a domain off into its own skill. Two skill bodies mount side by side
in the same agent context, so a split buys no context savings — it buys two
doctrine documents that WILL drift apart, and the agent reconciling them
mid-conversation is the failure mode.

`multi-cloud-catalog` is the one deliberate exception, and the reasons are
its admission test: its content is GENERATED at package time (the component
reference pack assembled from `catalog/`, far too large to author or review
as prose) and it is a research layer other skills read from rather than a
craft of its own. A new skill is justified only by BOTH properties —
machine-assembled bulk plus consumed-by-reference — never by a domain
feeling big enough to deserve its own directory.

The same doctrine holds across HOSTS. One skill serves every host that
loads it — the Planton Assistant on the desktop and web surfaces, and
coding agents such as Cursor, Claude Code, or Codex working inside a
developer's own repository. Where a host changes the choreography (the
folder it hands the agent, where the shell starts, whether a canvas is
watching), the skill resolves that inside its own instructions — the
`planton` skill's workspace-posture check is the pattern — and never forks
into a per-host variant. A per-host copy is the same drift hazard as a
per-domain split, with the added cost that the two hosts would then
disagree about what the platform does.

## The specification

The [Agent Skills specification](https://agentskills.io/specification) is
this tree's binding standard, and the lint gate enforces its limits so
skills can never drift out of conformance:

- **`name`**: lowercase alphanumerics and single hyphens, at most 64
  characters, equal to the directory slug exactly.
- **`description`**: non-empty, at most 1,024 characters. This is the one
  surface loaded for every agent at startup — it must say what the skill
  does, when to use it, and carry the trigger keywords, and nothing else.
  Capability detail belongs in the body or the references.
- **Frontmatter fields**: only the spec's set (`name`, `description`,
  `license`, `compatibility`, `metadata`, `allowed-tools`). Custom
  properties go under the spec's `metadata` map.
- **`SKILL.md` body**: at most 500 lines. The body loads whole on every
  activation; references load on demand — when the body grows, move detail
  into a reference and leave the routing sentence behind.

## Structure contract

```
skills/
  <skill-slug>/
    SKILL.md            # frontmatter (per the specification above);
                        # then the instruction body
    references/         # on-demand reference documents cited by SKILL.md
      <domain>.<topic>.md
  compat.yaml           # compatibility floor carried by every release manifest
  README.md
agents/
  <agent-slug>/
    instructions.md     # the agent's system instructions
```

Rules the lint gate enforces on every pull request (the same validator the
release lane packages with, so a tree that passes here is a tree that
ships):

- Every limit in "The specification" above.
- Reference integrity is bidirectional: every `references/<file>.md` that
  `SKILL.md` cites must exist and be non-empty, and every file present in
  `references/` must be cited by `SKILL.md`. Orphaned or empty reference
  files fail the gate.
- **References are FLAT files — a directory under `references/` is
  refused.** This is load-bearing, not taste: packaging and the citation
  gate both operate on the flat file list, so a nested file would be
  silently invisible to releases while appearing to exist in git. The
  specification's own guidance agrees ("keep file references one level
  deep").
- Reference filenames follow the dot taxonomy grammar: lowercase
  alphanumeric segments (hyphens within a segment) separated by dots,
  ending in `.md`.

## The reference taxonomy

Dots organize a large skill's flat references by domain — the segments are
the folder structure the spec advises against nesting. The `planton`
skill's domains:

| Domain | Carries |
|---|---|
| `infra.*` | Chart craft and lifecycle: format, templating, dependencies, config references, environments, the build contract, deployed projects, machine deploys, state import |
| `cloud.*` | Provider judgment: AWS architecture, Kubernetes architecture, read-only cloud exploration |
| `service.*` | Service delivery: registration doors, briefing a service from its page, reading a service's repository and writing a fix back as a pull request, delivery verbs, runs and build failures, fixing a failed run from its page (including the repository's own CI run, read through Planton), managed pipelines and their authoring, organization publishing of pipelines and tasks, serving domains, previews, local env vars, kustomize authoring, the offline/GitHub-Actions lane, the delete cascade |
| `self-hosted.*` | Administering a Planton the person runs themselves: reading the platform the operator converges, front doors and the CLI, upgrading, the first admin and seats, connecting a company directory, primary sign-in and the break-glass, mapping groups to roles, offboarding and sync |
| `catalog.*` | The research layer's doors: component grounding, catalog availability |
| `craft.*` | Working method and the person: discovery, personalization, the profile vocabulary, the CLI command map, cost transparency, gap filing |

Choosing a name: two segments (`<domain>.<topic>.md`) is the default;
add a third only when a topic genuinely subdivides (prefer
`service.serving-domains-custom.md` over inventing a deeper hierarchy). A
small skill (like the catalog skill's handful of references) may use
single-segment names — the grammar allows it; the taxonomy earns its dots
when the estate is large enough to need domains.

## Authoring a reference

- **Frontmatter**: a `title` and a `description` whose second half is the
  "Read when" — the exact moments an agent should open this file. The
  SKILL.md reference table's "Read when" column and this description tell
  the same story.
- **One topic per file.** References load on demand; a focused file costs
  the agent one read at the right moment, a grab-bag costs every moment.
- **Cite it or it fails the gate**: add the file's row to SKILL.md's
  reference table in the same change that creates it.
- **Compose, never duplicate.** Facts that have an authoritative home
  elsewhere — component schemas, generated reference pages, the secret
  snippets surface, another reference's topic — are pointed at, never
  copied. A copied fact is a fact that rots.

## How releases carry this content

Every semver release of this repository packages each skill directory into
a deterministic zip and publishes them to the downloads CDN together with a
`definitions-manifest.json` carrying the release version, per-archive
SHA-256 checksums, and the compatibility floor from `compat.yaml`
(the minimum daemon and CLI versions the content assumes). Consumers verify
checksums before installing; the Planton desktop's local daemon carries no
embedded copy — it resolves a stable release-channel pointer at boot and
seeds the pointed-at release into its engine, verifying each artifact's
checksum, the compatibility floor, and content shape before adopting it.

When content starts depending on a newer CLI or daemon capability, raise
the matching floor in `compat.yaml` in the same pull request — a consumer
whose binaries are older than the floor refuses the release and keeps
serving the previous one, which is exactly the graceful degradation the
floor exists to buy.

Each release also ships in a second, browsable shape for reading surfaces:
every skill's files individually fetchable under
`releases/{tag}/definitions/exploded/{slug}/`, described by a
`definitions-browse.json` file tree with per-file checksums, and a releases
index at `definitions/releases/index.json` naming every complete release.
The exploded files are the archive entries byte for byte (the archives are
deterministic stored zips), verified against the browse manifest before
upload — so browsing a release reads exactly what installers install.

The third consumer is the coding-agent lane:
[github.com/plantonhq/skills](https://github.com/plantonhq/skills) is a
distribution repository whose `skills/` tree the release lane rewrites from
the verified exploded trees after the stable pointer has moved, one commit
per release tagged with the release's tag. Coding agents install from it
(`npx skills add plantonhq/skills`, Claude Code's marketplace), so they run
exactly the bytes the stable channel points to. Nothing under its `skills/`
is authored by hand — a guard workflow there refuses pull requests that
try — which keeps this tree the only place skill content is written.

## The authoring bar

Every skill in this tree is crafted individually, grounded in thorough
research of the platform's actual behavior — real command output, real
schemas, real failure modes. There is no bulk-import path: a skill earns
its place here by being verifiably correct, and behavioral claims about an
agent's judgment are proven against a live engine before they ship.

Skills carry judgment and workflow. Component facts (what a deployment
component is, its fields, its examples) live in the generated reference
pages beside each component's protos (`catalog/...`), and the verified
data layer (cost fact-sheets with generated estimates, control posture,
provisioning-permission manifests) lives in the catalog's committed
sidecars and central `_pricing/`/`_compliance/` trees — both are composed
into agents by reference (the multi-cloud-catalog skill's reference pack
teaches reading them as files), never duplicated into a skill's prose,
where they would rot.
