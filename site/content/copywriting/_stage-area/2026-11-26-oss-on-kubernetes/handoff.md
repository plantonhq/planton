# Open source on your own Kubernetes cluster (handoff)

**Built:** round two of `draft-2.md` (A's wiring with C's bill framing), then the design review's fixes. The founder approved it on 2026-11-26 ("looks good proceed"), after asking whether to explain why or only what, and whether to say Kubernetes.

**The words:** `src/data/homepage.ts`, `HOMEPAGE.openSource`.
- **Heading:** "Your database, cache and sign-in on your own Kubernetes cluster, instead of three managed bills." The why is one clause, the bill, limited to what the walk showed: three services, three bills. The skeptic read a blanket heading ("run open source … instead of paying for managed services") as an over-claim, and two readers said it blurred with the self-hosting section below.
- **"Kubernetes" is said.** The non-technical reader recognises it, and "cluster" alone meant nothing to them. The section stays apart from "run it yourself" by its subject (the reader's own services, never Planton itself) and by "no Planton install on your cluster".
- **"The Planton CLI … with its open-source engine",** never "the open-source CLI". What ran was the installer's CLI, signed out, and its first output line says so. Signed in, the same command takes the platform's path, which is unproven for these kinds.
- **"Dev" lives in the ask's label,** so nobody reads the section as a production claim.

**The artifacts:** `OSS_WALK`, all verbatim from one walk on 2026-11-26, with lines cut only where a panel shows `⋯`.
- **The command:** split with the shell's `\` continuations.
- **The CLI's own deploy-order line:** kept, because it is the proof that the order was worked out, not typed.
- **What each service answered:** a table, one row per bill.
- **The keycloak.yaml excerpt:** it scrolls rather than wraps.
- **The catalog's list:** read from `platform-stats.ts`. Its count is the list's own length, since the whole list is shown.
- **A new walk replaces the record; never a hand edit.**

**Changed outside the section:**
- `Artifact.tsx`: a CLI status line (`✓`, `✔`, `ℹ`) hangs past its mark, and its hyphenated identifiers and arrows never break. The law's code-panel row says so.
- `check-page-weight.mjs`: HTML raised to 34 KB beside the measured 31.6 KB.

**The checks:** `check-homepage.mjs` covers:
- the section's place between infrastructure and run it yourself;
- the open-source-engine line;
- the operator first in the order;
- the services answering;
- the installer;
- no `$secret/`;
- no button.

**Must not claim:**
- a saving figure;
- a managed product by its vendor's name;
- "any open-source software";
- "production-ready";
- that signing in does the same;
- teardown (the CLI has no set destroy);
- a time figure;
- the host (a local cluster).

**Open:**
- **The next section's heading.** "Already on Kubernetes?" reads as a repeat directly under this one, until the context lake goes between them. It is the founder's call whether to reword it sooner.
- **No sentence on backups and upgrades.** That's the founder's approved after; the question stays in conversation.
- **The ask names no person.** No demo company is a small team on Kubernetes.
- **The phone length.**
