# The hero and the close (draft 1)

Stage area for planton.ai's one page, built one or two sections per session. This session: the hero (a brand-new cloud account to a running app) and the close (start). Written 2026-11-25 under `../../_rules/update-planton-ai-copy-writing.mdc`. Every sentence sits beside what backs it: a story chapter, a claims-ledger row (cited by claim and date), or a constant in `src/data/pricing.ts`. Nothing here is built until the founder approves the after.

The story changed first: the positioning story is v3.0 (2026-11-25), whose spine is a company with no DevOps team running its own cloud, and whose chapter 1 is "A new cloud account, and no DevOps team".

## The media every candidate shares: the Setup Prompt, as the server writes it

The prompt a new organization's Home hands over, shown as words, never as a working prompt (the real one carries a personal one-time sign-in code). The example organization is Slotwren, a demo company. There is no copy button.

```text
Set up Planton for this repository and take it to its first deployment, in my organization slotwren on planton.ai. I'm neha@slotwren.com.

1. Install the Planton CLI and sign it in as me. The code works once, within 15 minutes:
   curl -fsSL https://planton.ai/install.sh -o /tmp/planton-install.sh && sh /tmp/planton-install.sh
   ~/.planton/bin/planton login --grant <your one-time code> --instance planton.ai
   Call the CLI by that full path in every command.
2. Fetch https://planton.ai/agents.md with curl, read it as written, and follow it. Skip what you already did in step 1.
3. Follow the planton skill's first-deployment guide for this repository in slotwren. Ask me before anything that changes my cloud or costs money; connecting a cloud account counts. Tick each Getting Started step as you finish it with ~/.planton/bin/planton setup done <step> (run it with no step to list them).
4. When you finish, tell me what you set up, with a link to each thing.
```

Caption: "Your Home on Planton gives you your own prompt, with a sign-in code that works once."

| Rests on | |
|---|---|
| The prompt's words | `SetupAgentPrompt.java` (rendered by the server; the console never retypes it), the ledger's Setup Prompt row (2026-11-24, re-checked 2026-11-25) |
| The code works once, within 15 minutes | `SetupAgentPrompt.java` (the lifetime is passed in); the console's own words, `agent-setup-words.ts`: "The code in the prompt works once, for 15 minutes." |

## Candidate 1: the stand, proven by the prompt

- **Headline:** Run your cloud without a DevOps team.
- **Line:** Start from a brand-new cloud account. Paste one prompt into Claude Code, Codex or Cursor, and your agent takes your app to a running deployment in your own cloud. It asks you before anything changes your cloud or costs money.
- **Media:** the Setup Prompt, above.
- **Doors:** Start Free · Book a Demo. Under them: "Free for up to 3 seats. Never asks for a card."

## Candidate 2: the founder's measure, proven by the prompt

- **Headline:** From a brand-new cloud account to a running app.
- **Line:** Paste one prompt into Claude Code, Codex or Cursor. Your agent installs Planton, connects your cloud without storing a key, and ships your first deployment. You say yes to every change to your cloud or your bill, and install one GitHub App.
- **Media:** the Setup Prompt, above.
- **Doors:** Start Free · Book a Demo. Under them: "Free for up to 3 seats. Never asks for a card."

## Candidate 3: who does what (typography only)

- **Headline:** One prompt to a running app. No DevOps team.
- **Line:** From a brand-new cloud account, in Claude Code, Codex or Cursor.
- **Two columns:**
  - **You:** paste the Setup Prompt from your app's repository. Say yes before anything changes your cloud or costs money. Install Planton's GitHub App.
  - **Your agent:** installs the Planton CLI and signs in as you. Connects your cloud without storing a key. Ships your first deployment into your own cloud account. Tells you what it set up, with a link to each thing.
- **Doors:** Start Free · Book a Demo. Under them: "Free for up to 3 seats. Never asks for a card."

## The close (shared by every candidate)

- **Headline:** Start from a brand-new cloud account.
- **Line:** Sign up, copy the Setup Prompt from your Home, and paste it into your coding agent from your app's repository. Your app and its infrastructure run in your own cloud, on your own bill.
- **Doors:** Start Free · Book a Demo. Under them: "Free for up to 3 seats. Never asks for a card."

## Every sentence and what backs it

| Sentence | Rests on |
|---|---|
| Run your cloud without a DevOps team. / No DevOps team. | Story v3.0, the spine and chapter 1 (founder, 2026-11-25: "Planton's purpose is to help companies operate without a DevOps team") |
| Start from a brand-new cloud account. / From a brand-new cloud account to a running app. | Story chapter 1; the ledger's Setup Prompt row: paste to running app proven end to end on Google Cloud Run |
| Paste one prompt into Claude Code, Codex or Cursor | The ledger's Setup Prompt row; the console's words ("Works with Claude Code, Codex and Cursor", `agent-setup-words.ts`) |
| your agent installs Planton / installs the Planton CLI and signs in as you | The prompt's step 1 (`SetupAgentPrompt.java`) |
| connects your cloud without storing a key | The ledger's keyless-connection rows (Google Cloud, AWS, Azure, proven on prod) |
| ships your first deployment / takes your app to a running deployment in your own cloud | The ledger's Setup Prompt row; the delivery row (containers into the customer's own cloud) |
| It asks you before anything changes your cloud or costs money. / You say yes to every change… | The prompt's step 3, verbatim in meaning |
| install one GitHub App | The ledger's Setup Prompt row: one browser moment remains, installing Planton's GitHub App |
| Tells you what it set up, with a link to each thing. | The prompt's step 4 |
| Your app and its infrastructure run in your own cloud, on your own bill. | Story chapter 8; the delivery row |
| Free for up to 3 seats. | `FREE_TIER_SEATS` in `src/data/pricing.ts`; the ledger's free-tier row |
| Never asks for a card. | The ledger's no-card row (added 2026-11-25) |

Nothing here names a cloud other than Google Cloud beside "running app"; nothing claims it is hands-off; no figure, no analogy, no competitor.

## The claim check

Every row above was re-checked against the claims ledger on 2026-11-25, and the ledger corrected in place:

- **The Setup Prompt row** holds, with its proof narrowed: the prod checks prove the single-use sign-in, the Getting Started ticks and the playbook's commands; the full paste-to-running-service path is proven once, on dev, with Claude Code on Google Cloud Run. So no cloud but Google Cloud stands beside "running app", and nothing says Codex or Cursor are proven end to end.
- **"Works with Claude Code, Codex and Cursor"** is a new row: true today, built; only Claude Code is proven end to end.
- **"The infrastructure and the app run in your own cloud, on your own bill"** is a new row: true for the resources and the running app; builds run on Planton's runners, which the page does not contradict.
- **"Never asks for a card"** is added to the price row: the free offer carries no price and creating an organization has no payment step.
- **Keyless connect** holds on prod for Google Cloud, AWS and Azure (Azure's consent app was replaced on 2026-11-24 and is noted).

Held as written: every sentence. Reworded to the source: "ships your first deployment" became "gets your app running", one name for the result everywhere. Cut: nothing.

## The cold reads

Three rounds, every reader a fresh agent reading only the sheet.

1. **Three candidates, seven readers** (a developer who inherited ops, an application developer, a platform engineer, an engineering leader, a non-technical founder, a forum visitor with a coding agent open, a developer-tools copywriter). The stand ("Run your cloud without a DevOps team.") led with the developer who inherited ops, the engineering leader and the copywriter (7.5 of 10). The prompt-led measure led with the platform engineer and the forum visitor; who-does-what led with the application developer and the non-technical founder. Changes: "a brand-new cloud account" read as a precondition, so the line says "even a brand-new one"; the prompt read as someone else's, so it is labelled as Neha's example and the code as hers; "Setup Prompt" and "Home" were unexplained, so the caption drops "Home" and the words use "setup prompt" in lowercase.
2. **The revision, three readers** (the developer who inherited ops, the engineering leader, the application developer): all three restated it correctly.
3. **The founder kept today's team diagram** and asked for the story to fit it. The hero became the stand beside the diagram, and the prompt moved to the first section. Three readers (the developer who inherited ops, the engineering leader, the copywriter) named what fought: a "set up" headline beside an every-day diagram, the result called by two names, and the GitHub App step with no reason. All three are fixed: the headline says "Run", the result is "a running app" everywhere, and the GitHub App line says "so every push after this ships the same way".

Gaps the edges named, accepted: the platform engineer reads "without a DevOps team" as aimed at their job (the founder's stand, said to whoever asks); day two (changes, the 2 a.m. page) is the next sections' to show; working with an existing account lives in the docs and in conversation; `curl | sh` and the agent following `agents.md` are the product as it ships; the non-technical founder's door is Book a Demo.
