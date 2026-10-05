# catalog/_compliance — The Control Catalog and Framework Crosswalks

This tree is the compliance vocabulary's one home. Everything the platform says about technical controls and compliance frameworks — on the catalog's Posture tab, in the assistant's answers, in the Security Review Pack — traces back to the files here plus each kind's own `controls.yaml`.

## What lives here

- **`controls-catalog.yaml`** — the single controlled vocabulary of technical controls (currently 17, across data protection, network exposure, identity and access, observability, resilience, and change safety). Each control has an immutable id, a name, a one-sentence statement, and a category. Catalog kind control profiles (`catalog/<provider>/<kind>/controls.yaml`) reference these ids; nothing anywhere invents its own control vocabulary. Ids are immutable once published.
- **`frameworks/`** — one crosswalk file per external framework (HIPAA Security Rule, SOC 2 Trust Services Criteria, FedRAMP Rev 5 Moderate, CIS AWS Foundations Benchmark), each mapping the framework's own requirements — quoted verbatim from the published source document — onto catalog control ids. A crosswalk may declare a provider scope (`spec.providers`, the shared provider enum): CIS AWS names `aws` and appears only on AWS kinds; an empty scope means provider-neutral. Each file opens with its own honesty preamble.

## The rules this tree holds

- **A mapping says catalog controls SERVE a requirement — never that anything is "compliant", "certified", or "authorized".** FedRAMP binds hardest: it is an authorization program, so its crosswalk is evidence an assessment consumes.
- **Dishonest stretches are omitted, not blurred.** Requirements only organizational process can serve are deliberately absent, with the omission reasoned in the file.
- **Referential integrity is gate-enforced in both directions** (`pkg/compliance` conformance tests): every control id a profile or crosswalk cites must exist, and unknown ids, duplicate requirements, and invalid provider scopes fail CI.

## Fixing a claim

A kind's posture is fixed in that kind's `controls.yaml` (citing ids from here, with evidence). A framework mapping is fixed in its `frameworks/` file, with requirement text verbatim from the published source. A new control enters `controls-catalog.yaml` deliberately — the vocabulary is closed on purpose, and growing it is a design decision, not a data entry.

The kind anatomy these files plug into is defined once, in [architecture/catalog-kind.md](../../architecture/catalog-kind.md).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
