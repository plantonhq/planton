# Ask your coding agent for any infrastructure (draft 1)

Stage area for planton.ai's one page, built one or two sections per session. This session: Home's third section, between the first app and the close. Written 2026-11-25 under `../../_rules/update-planton-ai-copy-writing.mdc`. Every sentence sits beside what backs it: a story chapter and proof, a claims-ledger row (cited by claim and date), a constant in `src/data/platform-stats.ts`, or the real run. Nothing here is built until the founder approves the after.

The story changed first: the positioning story is v3.1 (2026-11-25). Chapters 2, 10 and 11 carry the agent's part and the modules' proving record in their proof; chapter 2's claim (the reopened sentence) is unchanged. The fact reference's catalog section was corrected the same day, and the claims ledger gained two rows: the agent's part, and the proving record.

## The media every candidate shares: what a real agent wrote

A headless Claude Code with the released Planton CLI (`v0.0.165-cli.20261124.0`) and skills (`plantonhq/skills` `v0.6.35`), signed in to nothing, in a copy of Slotwren's repository. The run's record (the ask, every command, the files, the check) is kept with the website project's research. Slotwren is a demo company; its API has no database today, so the ask states the new need.

**The ask, verbatim:**

> I'm Neha, co-founder and CTO of Slotwren. This is our booking API. We're adding stored bookings, so the API needs a Postgres database on Google Cloud for our staging environment, in our Planton organization slotwren. Use Planton to set up the infrastructure for it in this repository. Don't sign in and don't deploy anything yet; I'll review what you write first.

**What it wrote:** `infrastructure/booking-db/instance.yaml` (GcpCloudSql, 40 lines, shown whole), `database.yaml` (GcpCloudSqlDatabase), `app-user.yaml` (GcpCloudSqlUser). Passwords are `$secret/` references.

**The check, as the same CLI printed it:**

```text
$ planton validate infrastructure/booking-db/instance.yaml
✓ manifest is valid
$ planton validate infrastructure/booking-db/database.yaml
✓ manifest is valid
$ planton validate infrastructure/booking-db/app-user.yaml
✓ manifest is valid
```

| Rests on | |
|---|---|
| The file, the ask and the check | the run's record (2026-11-25); the ledger's "validated offline" row (re-checked 2026-11-25) |
| Written by Claude Code | the run's record; the ledger's coding-agents row (only Claude Code proven end to end; nothing here claims Codex or Cursor ran it) |
| GcpCloudSql proven 2026-11-20, GcpCloudSqlUser 2026-11-20, GcpCloudSqlDatabase 2026-08-13, each on both engines | each kind's `catalog/gcp/<kind>/e2e/profile.yaml` in the open-source repository; the ledger's proving-record row (2026-11-25) |

## Candidate 1: the file comes back

- **Heading:** Ask your coding agent for infrastructure. Read what it writes.
- **Line:** Your agent picks the kinds from Planton's catalog, writes a short manifest and checks it before anything exists. It never writes the Terraform or Pulumi: the module that deploys each kind is already written, in the open.
- **The ask** (a quiet quote, left): "Neha, co-founder and CTO of Slotwren, a demo company, to Claude Code:" and her words.
- **Artifact (right):** label "infrastructure/booking-db/instance.yaml · written by Claude Code from Neha's ask"; the file whole; under it "and database.yaml, app-user.yaml"; then the three validate lines.
- **Check:** `planton explain GcpCloudSql` lists every field the agent chose from, offline. Read the module that deploys it (link to `catalog/gcp/gcpcloudsql/iac/`).
- **Quiet line:** One of 850+ kinds across 9 providers.

## Candidate 2: two layers, one already written (the lead's recommendation going in)

- **Heading:** Your agent writes the manifest. The module is already written.
- **Line:** Ask your coding agent for the infrastructure you need next. It picks the kinds from Planton's catalog, writes a short manifest and checks it before anything exists. It never writes the Terraform or Pulumi that deploys it.
- **The ask** (quiet quote, left), as in candidate 1.
- **What deploys it** (a table, left, under the ask):

  | File | Kind | Module | Proven in a real Google Cloud account |
  |---|---|---|---|
  | instance.yaml | GcpCloudSql | Terraform and Pulumi, open source | 2026-11-20 |
  | database.yaml | GcpCloudSqlDatabase | Terraform and Pulumi, open source | 2026-08-13 |
  | app-user.yaml | GcpCloudSqlUser | Terraform and Pulumi, open source | 2026-11-20 |

  Under it: "Proven: deployed by its own module, checked with Google's own tools, destroyed and checked clean."
- **Artifact (right):** as in candidate 1.
- **Check:** `planton explain GcpCloudSql` lists every field, offline. Read the module on GitHub.
- **Quiet line:** One of 850+ kinds across 9 providers: AWS, Google Cloud, Azure, Kubernetes, Cloudflare, DigitalOcean, Auth0, Stripe and OpenFGA.

## Candidate 3: the founder's title

- **Heading:** Ask your coding agent for any infrastructure.
- **Under it, first:** 850+ kinds across 9 providers.
- Everything else as candidate 2.

## Every sentence and what backs it

| Sentence | Backed by |
|---|---|
| Your agent picks the kinds from Planton's catalog, writes a short manifest and checks it before anything exists. | ch 2, proof 3 (the agent line); ledger "agent writes manifests" (2026-11-25), "validated offline" (re-checked 2026-11-25) |
| It never writes the Terraform or Pulumi (that deploys it). | ch 2, proof 3; ledger "agent writes manifests" |
| The module that deploys each kind is already written, in the open. | ch 2, proof 3; ch 8 (Apache-2.0 modules in the open-source repository); every kind's `iac/` |
| Your agent writes the manifest. The module is already written. | ch 11, proof (the agent writes a manifest; the module was written and proven before anyone asked) |
| Proven: deployed by its own module, checked with Google's own tools, destroyed and checked clean. | ch 10, proof (the proving record); ledger "proving record" (2026-11-25); `e2e/README.md` (the lane) |
| Proven dates per kind | each kind's `e2e/profile.yaml` |
| 850+ kinds across 9 providers | `PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT`, `CLOUD_PROVIDER_COUNT`, `CLOUD_PROVIDERS` (platform-stats.ts); ledger L28 (re-counted 2026-11-25) |
| `planton explain GcpCloudSql` lists every field, offline. | `planton explain --help` (released CLI); `domain.infra-hub.planton-cli.md` |

## The claim check

- Held: every sentence above, against its source.
- Reworded to the source: "tested" became "proven", with the lane's own words (deployed, checked, destroyed, checked clean), because "tested" invites "continuously" and the lanes are manual.
- Cut: "any cloud" (the walk is Google Cloud only); "in minutes" (no timed walk); "guaranteed"; "every module"; the founder's comparison with an agent writing Terraform (chapter 11's never-say: the agent is not the problem).
- Engines: "Terraform and Pulumi" holds for these three kinds; it is not said of the catalog as a whole (Stripe and OpenFGA are Terraform only).

## The cold reads

Seven fresh readers read the three candidates in place between the first app and the close: a developer who inherited ops, a working application developer, a platform engineer, an engineering leader, a non-technical founder, a visitor from a forum thread, and a developer-tools copywriter.

- **Candidate 2 won.** Five of seven kept reading after it (inherited ops, app developer, platform engineer, engineering leader, forum visitor): the proven table is "the first thing that earns my trust". The copywriter wanted candidate 2's body under a plain heading; the non-technical founder preferred candidate 3.
- **"Any infrastructure" lost.** Five distrusted "any" ("the word HN shows up to disprove"; "850 kinds is not any"). The heading became the mechanism in plain words: "Ask your coding agent for infrastructure. The code that deploys it is already written."
- **Neha's self-introduction cut** (four readers): the ask is excerpted with marked ellipses; the record keeps it whole.
- **What happens next was missing** (four readers): one line added, "When you say so, Planton deploys it into your own cloud account with those modules."
- **"Kind" undefined** (four readers): shown by example, "GcpCloudSql is one of 850+ kinds in the catalog".
- **The table** became Kind and Last proven, two columns, so it fits a 320 px phone; the files are named in the manifest's label.
- **Kept on purpose:** the file verbatim, including its "public IP" comment (it explains itself, and now wraps whole on a phone); the August date on GcpCloudSqlDatabase (the truth, under "Last proven"); the full 40-line file (the engineering leader asked for 15 lines; the developer, the skeptic and the forum visitor judged the work by reading it whole).
- **Answered by later sections, not this one:** the deploy and what happens when it fails, the cost, where `$secret/` lives, plan or diff before apply, state, other clouds and agents.

## What the page says (after the cold reads)

- **Heading:** Ask your coding agent for infrastructure. The code that deploys it is already written.
- **Intro:** Your agent picks from Planton's catalog and writes a short manifest that says what you want, then checks it before anything exists. It never writes the Terraform or Pulumi: the module that deploys each part is already written, in the open, and proven in a real cloud account.
- **Next:** When you say so, Planton deploys it into your own cloud account with those modules.
- **The ask:** Neha, co-founder and CTO of Slotwren, a demo company, to Claude Code: "…We're adding stored bookings, so the API needs a Postgres database on Google Cloud for our staging environment… Don't sign in and don't deploy anything yet; I'll review what you write first."
- **Artifacts:** `instance.yaml` whole; the three validate lines; the Kind and Last proven table.
- **Proven note:** Proven means deployed by its own module into a real Google Cloud account, checked with Google's own tools, destroyed and checked clean, on Terraform and on Pulumi.
- **Check:** Read the module that deploys GcpCloudSql, or list every field the agent chose from, offline: `planton explain GcpCloudSql`.
- **Reach:** GcpCloudSql is one of 850+ kinds in the catalog, across 9 providers, from AWS, Google Cloud and Azure to Cloudflare and Kubernetes.
