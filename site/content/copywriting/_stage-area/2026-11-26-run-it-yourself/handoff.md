# Run Planton yourself (handoff)

**Built:** candidate A from `draft-1.md`, with the copywriter's edits. The founder approved the after on 2026-11-26.

**The words:** `src/data/homepage.ts`, in `HOMEPAGE.selfHosted`.
- **Heading:** "Already on Kubernetes? Run all of Planton there." Self-hosting is offered as a choice for teams that already run a cluster, so the hero's stand (no DevOps team) is never contradicted.
- **The order of the lines:** first what installs, with the payoff first (infrastructure management and CI/CD). Then that nothing needs setting first, with the setup code before "administrator": three readers read "the first person in is the admin" as a security hole until the code explained it. Then the bridge, that the manifests from the section above deploy the same way. Then the Community seat line, the docs link, and the requirements in small print.

**The artifacts:** `SELF_HOSTED_WALK`.
- Both are verbatim from a real install on a fresh cluster on 2026-11-26.
- Both chart and release are pinned, so the commands stay valid past the operator's release floor.
- The install keeps the shell's own `\` continuations, and the status command scrolls inside its panel. Neither wraps mid-flag.
- When a release moves the floor past the pinned pair, the install is run again and the record replaced.

**Changed outside the section:** the close's note now names planton.ai ("Free on planton.ai for up to N seats"). All seven cold readers took the section's Community count and the close's hosted count, one scroll apart, as a contradiction.

**The checks:**
- `check-homepage.mjs`: the section sits between infrastructure and the close; it has the pinned commands, the Ready status, the docs link and no button; it never shows what an install left unconfigured.
- `check-page-weight.mjs`: HTML raised to 30 KB beside the measured 27.6 KB.

**Must not claim:**
- a time figure (the install ran emulated on a laptop);
- "air-gapped";
- "zero lock-in";
- that the platform is open source;
- any internal component's name;
- what isn't configured;
- a typed seat count.

**Open (the design review's minors, round 2 PASS):** the status command scrolls inside its panel; on a phone the operator chart's address wraps after its hyphen; the lede and body keep two measures (the shared `.words` rule); the phone heading breaks after "Run"; past the Community seats the header's Pricing answers.
