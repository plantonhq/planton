# Catalog Availability — the Organization's Governed Catalog

Platform teams can curate which cloud providers and catalog kinds their
organization's members may self-service: disable whole providers, disable
specific kinds, or narrow a provider to an approved list. The control plane
refuses creating a disabled kind at every door — direct creates, chart
deploys, service deploys — and the same server-side answer that enforces
the refusals is what the availability instruments below report, so what you
learn here can never disagree with what a deploy attempt would hit.

**The law: check availability, design complete, disclose before deploy.**
Three duties, in that order, whenever you work in an organization's context:

1. **Check** the organization's availability during silent grounding
   (Phase 0) — a lookup, never a question.
2. **Design the complete architecture regardless.** Availability never
   truncates or distorts a design: the right architecture for the user's
   need is the same whether or not their organization has enabled every
   piece of it yet. A design silently missing its best kind because a
   policy disables it today is a worse lie than a refused deploy.
3. **Disclose before any deploy.** Before the deploy offer or act, tell the
   user which kinds of the design the policy disables and what that
   means — the rest can deploy now, and someone who manages the
   organization's catalog policy (an Infrastructure Admin, under Org
   Settings → Catalog) can enable the disabled kinds. Never frame it as a
   product limitation or an upsell: the organization chose this.

## The instruments, per arm

- **Platform-tools arm**: `get_catalog_availability` (takes the org from
  your standing context) returns the disabled providers and kinds; empty
  sets mean the full catalog. The discovery tools
  (`search_catalog_kinds`, `get_catalog_kind`) and the kind
  catalog resource deliberately keep serving the FULL release catalog —
  they answer "what exists", this tool answers "what is enabled here".
- **CLI arm**: the browse commands are offline-first and serve the full
  embedded catalog by default; `planton catalog search --server` subtracts
  what the current organization's policy disables (with a note counting the
  hidden kinds), and `planton catalog get <kind> --server` adds an
  Availability line to the entry. Use `--server` when the question is "what
  can THIS org create", and the default when the question is "what exists".

Availability is a fact you READ AT ANSWER TIME, never recall: operators
edit the policy whenever they need to, so what you learned last
conversation is stale by construction — the same law that governs schema
and cost facts.

## The disclosure shape

Short, at the deploy moment, in the user's language:

> Two kinds of this design — the DigitalOcean droplet and the OpenFGA
> store — are disabled by your organization's catalog policy, so deploying
> them would be refused. Everything else can deploy now. An Infrastructure
> Admin can enable those two under Org Settings → Catalog if you need them.

When the policy disables nothing that the design uses, say nothing — a
disclosure with nothing to disclose is noise.

## What the policy never touches

**Disabling never strands existing infrastructure.** Updates, redeploys,
destroys, and deletes of resources that already exist are never blocked by
the policy — only NEW creations of disabled kinds are refused. Two
consequences for your behavior:

- A working copy of a deployed project saves and redeploys normally even
  when its kinds are now disabled — never warn about availability on a
  working-copy save (`deployed-projects.md`).
- References to existing resources of disabled kinds keep resolving; wiring
  a new resource TO one is a read the platform permits.

## Degradation and refusals

- **Instrument unreachable** (an older instance, no read access on the org,
  no backend): treat the catalog as unrestricted and move on without
  commentary — the control plane enforces regardless of what you could not
  learn, and its refusal is honest and names the way out.
- **A deploy is refused naming the catalog policy**: that is the
  advertisement you skipped arriving late, not an error to work around.
  Relay it faithfully — the disabled kind, that existing resources are
  untouched, and that an Infrastructure Admin can enable it — and offer to
  deploy the rest.
- **Chart listings are already curated**: the platform's chart search
  subtracts charts composing any disabled kind server-side, so a chart you
  can list is a chart the org can deploy.
