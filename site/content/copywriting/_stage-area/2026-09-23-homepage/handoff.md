# Homepage implementation handoff

Routes: `/` and the existing `/book-demo`. Homepage version: v6-2026-09-23.

Engineering leaders should understand infrastructure setup, application delivery, and team controls before encountering internal product names. A single diagram vocabulary explains those jobs. The existing contact form and calendar complete the conversion.

Copy lives in src/data/homepage.ts, sourced by chapter and wiki documents; approved quotes stay in testimonials.ts. Full document inventory: source-ledger.md and source-manifest.json. The existing story.ts is retained for other pages; the new homepage and homepage discovery export do not reproduce its broader verification claims.

Reusable primitives: ArchitecturePlanes, WorkflowVisuals (InfrastructureVisual, DeliveryVisual, ReviewVisual, OwnershipVisual), HomepageAppearance, DemoLink. Layout CSS is scoped. Default dark palette comes from website-shell tokens; the scoped light palette applies to homepage chrome and demo; other routes retain dark defaults.

Do not claim universal import, universal keyless availability, certified deployments, universal proven costs/controls, automatic approvals, guaranteed savings, or a booked meeting from a form submission. Diagrams are explicitly illustrative and carry no invented dollar figures.

Validation: full site build and guards, responsive captures, keyboard and reduced-motion checks, mocked form and SDK events, analytics privacy and deduplication checks. A production booking is never submitted during validation.

Rollback: restore the previous page composition and landing-page index from the pre-change commit. v5 remains intact. Shared navigation and other page records are preserved.

Review revisions and validation results are recorded in verification.md after the completed build. The founder authorized PR creation, merge after checks, and production publication on 2026-09-23.


## Final launch revision

Light-only homepage and booking flow, matching header/footer and portal menus. The appearance switch is removed. Add an agent-driven DevOps section after the product explanation: ask, prepare, review, deploy. Setup claims are grounded in public/docs/coding-agents.md and the current shipped skills/planton/SKILL.md, plus the service engine and catalog curation sources already inventoried. Approval requirements remain explicit. The product theme and docs presentation are unchanged.
