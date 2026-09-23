# Draft 1: the desktop landing and download pages as one record

The two pages under `/desktop` were written on 2026-09-11 and reviewed then; their words are good and mostly stay. What changes is where they live: every sentence moves from a component into `src/data/desktop.ts` beside the chapter or document that backs it, and the pages render the record. This draft is the human-readable record. Sentences already in data (`positioning.ts` desktop block, `desktop-download.ts` platforms and commands, `pricing.ts` seats, `platform-stats.ts` counts) are named, not repeated.

Conventions: titles and labels in Title Case; everything else sentence case. Numbers carry their provenance on the page.

## Claim check (before any code)

Every sentence on the two pages was read against the story (chapters 1, 2, 8, 12) and the desktop record: `public/docs/ci-cd/on-your-laptop.md`, `public/docs/coding-agents.md`, `public/docs/connections/index.md`, and the company repository's knowledge articles on the local runtime, runtime-artifact distribution, desktop identity, and local cloud-credential autodetect, plus the desktop wiki. Findings:

- "The Planton Assistant converses with zero keys and zero accounts; Planton funds it": backed. The desktop's free device lane bills Planton's own engine account; a signed-in wallet-holder's conversations draw their own wallet. Kept as written; this is the shipped anonymous lane, not the monthly allowance (designed, not live), which the page never mentions.
- "Bring your own Anthropic or Cursor key ... the conversation and the model traffic never leave the laptop": backed by the demo record's precise statement (bring-your-own-key is the privacy choice; the hosted option is never called fully private). Kept.
- "Cloudflare and DigitalOcean tokens are detected from your environment and verified against the provider": backed by the autodetect article (detection-first treatment for both). Kept.
- "Postgres, Temporal, and a cache, as native processes ... downloaded once on first launch, verified": backed (the daemon supervises Postgres, Temporal, Redis, and the control-plane jar; every artifact is installed only after a SHA-256 verification). Kept.
- "Planton honors the tofu or pulumi already on your PATH; only when nothing usable exists does it install one": backed (tools already on the PATH are used and never removed). Kept.
- "Updates are signed by Planton and verified before they are applied": backed (the desktop's updater signs releases). Kept.
- The measured figures (35 s, 30 s, 34 s, about 800 MiB): a measurement recorded in the component's comment, consistent with the CI/CD guide's ranges (thirty to forty-five seconds to build; about 800 MB idle). They move into the record with their provenance (the machine, the timestamps) as a field the page prints as the strip's caption.
- The "Read Next" block ("Git push to production with built-in CI/CD", "portable infrastructure definitions"): 2025's sentences, retyped beside the registry's own. Replaced by the registry's descriptions through the shared page card.
- "Download Desktop App" (the hero's button): the door vocabulary says "Download Planton Desktop". Replaced by the shared doors.

## The landing (`/desktop`)

**Hero.** Title: the desktop line (positioning). Lede: The whole Planton platform runs on your laptop and deploys to your cloud with the logins already on your machine. No account. Free for individuals, including commercial use. (ch 8, ch 12) For whom: the ways line, derived from the platform list (which platforms today, one command where live, signed and notarized on macOS). Doors: Download Planton Desktop, Set Up Your Coding Agent. Screenshot slot: the local home (a real capture or nothing).

**You Already Have Two Ways to Do This.** (ch 1) A Platform That Hides the Cloud: convenience, in exchange for control; your app runs in someone else's account, under IAM you did not write, on a bill you cannot itemize; when you outgrow it, you start over. Your Agent and the Cloud CLI: control, in exchange for a record; the agent writes the Terraform from memory and you learn at apply time what it got wrong; permissions nobody derived; a cost you find out about next month; a database password that went through the chat to reach the shell; no pipeline; the second environment is a second conversation. The turn: Planton is a platform too. It just runs on your machine, in your account, and gives the agent rails you can inspect. The concession (positioning).

**Two Ways to Ask. One Craft.** (ch 2 proof 2) Ask from inside the tool you already use, or ask the assistant built in. Both run the same skills. Your Coding Agent: install the Planton skills once; from then on Cursor, Claude Code, or Codex composes infrastructure grounded in the real schema, validates it offline, and applies it through the platform on your laptop, while you stay in your editor; the install command (desktop-download); what Planton adds (positioning, seven points). The Built-In Assistant: open the app and ask; the Planton Assistant converses with zero keys and zero accounts; Planton funds it. Prefer to keep everything on your machine? Bring your own Anthropic or Cursor key in the engine chooser, and the conversation and the model traffic never leave the laptop. Screenshot slot: the chooser. Closing line: your agent and the platform's own assistant share one craft: the same skills, the same catalog, the same rules about what never happens without your say-so.

**What Actually Runs on Your Machine.** A full platform as ordinary processes. No containers to babysit. Four tiles: The Control Plane and One Runner; Postgres, Temporal, and a Cache, as Native Processes; OpenTofu and Pulumi as Your Tools; Your Data Stays Here (ch 8 proof 2).

**Infra Hub on Your Laptop.** Kicker: the hub's analogy (positioning). Lede: the hub's line. Screenshot slot: the studio. Three points: Your Clouds, Found, Not Pasted; N Resource Kinds Across M Providers (platform-stats); A Map of Everything You Run.

**Service Hub on Your Laptop.** Kicker: the hub's analogy. Lede: Push to GitHub. Your laptop notices, builds the commit in a pod, deploys it to the cluster or cloud you connected, and puts a status on the commit. (ch 6; the CI/CD guide) The measured strip: four figures over their provenance. Three points: No Public URL, No GitHub App; Builds in a Pod on Your Machine; Deploys Where You Point It. The Docker note.

**The Same Planton Everywhere.** (ch 8 proof 3) This laptop, planton.ai, or your company's cluster: pick where in the instance switcher. One data model, one set of code paths; only which capabilities run differs. Your laptop can also deploy into your team's Planton, with your cloud credentials still never leaving your machine. The umbrella sentence.

**Why It Is Free.** (ch 12) Planton Desktop is free for individuals, including commercial use. Not a trial, not a tier with the good parts removed: the full catalog, the wizards, the IaC pipelines, service CI/CD, and secrets management are the core product on every edition. Here is the business model, stated plainly. Planton makes its money when a team adopts it: hosted seats beyond the free N, self-hosted licenses beyond the community edition's M seats, and prepaid AI credits. The solo developer is not a funnel stage. You are the person this was built to delight, and if you one day bring a team, nothing you built gets redone. Door: Pricing.

**Read Next.** CLI, Infra Hub, Service Hub (registry cards).

**Close.** Install It. Ask It Something. Download, open, pick Run on This Computer, and ask for what you need in your own words. Or install the skills and ask from your editor. Doors: Download Planton Desktop, Set Up Your Coding Agent; the CI/CD guide as the third link.

## The download page (`/desktop/download`)

**Hero.** Title: the registry's. Lede: Free for individuals, including commercial use. No account, no sign-up. Pick your platform; the rest is one launch. The live version and checksums. The platform tabs; the platform card (desktop-download) or the unavailable card: The {platform} Installer Returns with the Next Release; the build currently published for {platform} is not one we would ask you to run, so the link is off until the next release replaces it; two ways forward today (the CLI on WSL; switch to an available platform); the CLI runs the same schema lookups, validation, and deploys from a terminal; the desktop app adds the assistant, the canvas, and the local instance.

**After You Install.** One launch does the rest. Three beats: Pick Where Planton Runs; Watch It Set Itself Up; Your Clouds Are Already There. Then, From Wherever You Work: ask from your editor, from the terminal, or build your services on this machine; each is a minute. Three tabs: Coding Agent (the skills install command and its teaching), CLI (three commands and the platform's CLI note), Build Locally (Optional) (the build-cluster status command and its teaching).

**Verify the Download on {platform}.** Hash the file you downloaded and compare it with the release's checksums; on macOS, the next two lines ask Gatekeeper and the notarization ticket directly. The platform's verify commands (desktop-download). **Updates.** The app checks for updates and installs them with one click; on Homebrew the upgrade command does the same; updates are signed by Planton and verified before they are applied. Footer: why run Planton on your laptop (the landing); the CLI on its own; the self-hosted edition.

## Read cold, as the developer who wants the app tonight

The landing tells me in the first screen what it is, that it is free, that it is for my machine, and how to get it; the two-ways section names my exact situation (I have an agent and the cloud CLI) without sneering; the concession makes me trust the rest. The download page gives me my platform's button first with the size, and tells me what happens after I install, which download pages never do. Changed on this read: nothing in the words; the hero's button now says what every other door on the site says.

## Read cold, as a senior copywriter for developer-tools sales pages

One claim per screen with its proof beside it, except the screenshot slots, which are honest nulls until real captures exist. The measured strip is the strongest thing on the page and now carries its provenance on the page rather than in a code comment. The "Read Next" block was the one place the page said something the rest of the site no longer says; it now reads the registry. Every label Title Case, every body sentence case.
