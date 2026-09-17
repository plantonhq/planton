# Full Terraform Parity: Built for the Agent Era

## The Design Principle

The Deployment Component Store is built for **100% parity with the canonical Terraform provider** of every cloud it supports — in breadth (which resources exist in the catalog) and in depth (which configuration knobs each component exposes). Every knob the provider exposes has a home in the manifest.

## Where We Started: Curation Made Sense

The catalog began as a curated library. After 13+ years working as a cloud engineer across multiple companies and cloud providers, a consistent pattern was clear: most teams deploy the same core services, and most deployments touch a small fraction of each service's configuration surface.

When humans read configuration forms, curation was the right call:

- A form with 200 fields overwhelms; a form with 12 well-chosen fields empowers
- Surfacing the essentials and defaulting the rest made infrastructure approachable
- The modeling cost of every extra field was paid by hand, so every field had to earn its place

That was the correct trade-off — for that reader.

## What Changed: The Reader Is Now an Agent

AI agents are now the catalog's primary readers. Agents compose real architectures from these components, and agents are a fundamentally different audience:

- **Agents already know Terraform deeply.** LLMs are trained extensively on the Terraform providers' surfaces. A catalog that mirrors the full provider surface lets that training transfer directly to Planton manifests.
- **Agents don't get overwhelmed by breadth.** The usability argument for hiding fields disappears when the reader can hold the entire surface at once.
- **Agents hit hard stops where humans find workarounds.** A human presented with a missing knob improvises; an agent either fails the request or silently mis-models it.

## Why Full Parity, Not a Curated Subset

For an agent-first catalog, a partial surface is worse than a large one:

- **An incomplete configuration surface is a dead end.** If the provider exposes a setting and the manifest cannot express it, the agent's task fails — regardless of how rarely that setting is used.
- **A partial-coverage claim is ambiguous in every direction.** "Most settings are covered" leaves unanswerable questions: which ones are missing, why, and will they stay missing? Agents cannot reason reliably over an ambiguous surface.
- **Full parity is a single, checkable sentence.** A catalog that targets full parity with the provider at a pinned version is something an agent can be told once and trust everywhere — no dumbed-down subset, no dead ends.

The economics changed too: component authoring is now agent-executed, so the hand-modeling cost that justified curation is gone, while the ambiguity cost of partial coverage compounds.

## What Parity Means

### Depth: The Full Configuration Surface

Every component models the full configurable argument surface of the Terraform provider resources it maps to, at a pinned provider version. Restructuring is encouraged — idiomatic shapes, cross-component references, names that read better — as long as every provider argument remains representable.

### Breadth: Every Resource Accounted For

Every resource in the provider's catalog is accounted for in exactly one of three ways:

1. **Modeled** — it is a catalog component
2. **Composed** — deliberately folded into another component, with the mapping documented (e.g., a router and its NAT managed together)
3. **Excluded with a recorded reason** — reserved for deprecated or superseded resources

Omission is a decision, and every decision is recorded — never a silent gap.

### The Provider Block: Parity Below the Resources

Parity covers the provider's own configuration block, not just its resources. Enterprise-mandatory settings live there — role-assumption chains, organization-wide default tags, private endpoint overrides, retry tuning — and they are held to the same rule: every provider-block argument is expressible through the connection's provider configuration, deliberately excluded with a recorded reason, or owned by the modules by recorded judgment. Even an argument a module sets inside its own provider block is a census fact that must carry a decision — never an invisible leak.

### Pinned Versions Keep the Promise Honest

Parity is always declared against a named provider version, never against "latest". Providers ship weekly; the pinned version is kept fresh so the catalog never silently decays.

## Opinionated Composition on Top

Full parity does not mean dumping raw provider schemas on users. The complete surface is the foundation; the experience is built on top:

- **Essentials surface first**: forms and manifests organize so the settings teams actually change are front and center, with sensible defaults covering the rest
- **Opinionated composition**: components combine related resources the way practitioners actually use them
- **Production-ready defaults**: the common case stays simple — parity raises the ceiling without raising the floor

## The Value Proposition

### For Developers
- The common case is still a short form with sensible defaults
- When you need an advanced setting, it exists — no waiting on a platform team to add a field

### For AI Agents
- The full provider surface means every request is expressible — no dead ends
- Terraform knowledge transfers directly to Planton manifests

### For Platform Teams
- No more "is this knob supported?" tickets
- One clear coverage story to communicate, instead of an ever-shifting curated list

## The Key Insight

Curation optimized for the reader who existed when the catalog was born: a human scanning a form. The reader changed. Agents thrive on complete, unambiguous surfaces — so the catalog is built for 100% Terraform-provider parity, with opinionated composition and essentials-first ergonomics layered on top.

The result is a catalog where the common case stays simple and the complete case is always possible.
