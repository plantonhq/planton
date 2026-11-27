# The run (draft 1)

This is the stage area for planton.ai's one page, built one or two sections per session. This section follows "ask your coding agent for infrastructure" and shows those files deployed: in the page's final order it is the third section, "watch it deploy together". Written 2026-11-27 under `../../_rules/update-planton-ai-copy-writing.mdc`; every sentence sits beside what it rests on.

## The media: a real stack run on demo.planton.ai

Neha's coding agent (a headless Claude Code with the released CLI and skills, signed in as Neha) was asked to deploy the section above's three files to Slotwren's staging as one stack. The record is kept in the platform repository's website research.

- Planton's plan showed three resources, a monthly cost of $100 to $428 across the catalog's presets, and 16 added Google Cloud permissions. The cost was over Neha's $100 limit, so the agent stopped and asked her. She said yes.
- The first run failed on a missing Google project (filed as a product bug); the agent named the project, and the second run applied all three resources in 6m 25s, credited to "neha and claude".
- The stack was then undeployed, keeping its runs, so nothing runs in Slotwren's cloud. The screen is a crop of the demo universe scene `slotwren.staging-database.run.applied`, below the run's title.

## Candidates

**A.** "Watch it deploy. No pipeline to write." Neha told her coding agent to deploy those files as one stack. Planton showed the plan and its monthly cost first, then ran the deploy itself, in Slotwren's own Google Cloud. *(screen)* The run stays in your organization: who started it, which coding agent, and each resource's outcome and time. *(folded: the plan)*

**B.** "Your agent writes it. Planton deploys it and keeps the record." Same screen and record line.

| Rests on | |
|---|---|
| The plan with its cost and permissions before anything ran | the ledger's plan row (true today, the impact as built, checked live 2026-11-27) |
| Planton runs the deploy itself; no pipeline to write | the run on demo (`infpipe_01m8cgms5pv20b49tdt62q3nb3`); an infra run starts from an applied plan, never from a push, so no sentence says "every push deploys infrastructure" |
| Who started it and which agent | the ledger's person-and-agent row; the run's own header ("neha and claude") |
| "Together" | not said: a developer scoped to staging is refused on the run page today (filed); Ben's view joins when that is fixed |
