# The one sentence, and batch 1: what comes off (draft 1)

Stage area for planton.ai's rebuild around one sentence. Written 2026-11-24 under `../../_rules/update-planton-ai-copy-writing.mdc`. Every claim beside the ledger row it rests on; every ledger row cited by claim and date. Nothing here is built until the founder approves batch 1's after.

## The sentence (v1)

> **Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs.**

Under it, as the first action: **Start by pasting one prompt into your coding agent.**

| Claim in the sentence | Rests on |
|---|---|
| deploys every push to GitHub | ledger 2026-10-30, the delivery rows (containers from GitHub into your own Kubernetes, Google Cloud Run or Amazon ECS; a first build on Planton's runners), re-checked 2026-11-24 |
| into your own cloud | the same rows, and the keyless-connection row (AWS, Google Cloud, Azure; one-click on AWS), re-checked 2026-11-24 |
| shows the cost and permissions of every infrastructure change before it runs | ledger 2026-10-30, "every infrastructure change is a plan with its cost, permission and security impact; only the reviewed plan is applied", re-checked 2026-11-24 |
| start by pasting one prompt into your coding agent | ledger, the Setup Prompt row (added 2026-11-24, live on planton.ai since 2026-11-20) |

## The five candidates

| | Candidate | Rests on |
|---|---|---|
| A | Paste one prompt into your coding agent, and your app deploys from GitHub into your own cloud, then again on every push. | Setup Prompt row; delivery rows |
| B | Planton deploys every push to GitHub into your own cloud, and shows what each infrastructure change will cost and touch before it runs. | delivery rows; plan-with-impact row |
| C | Planton is how your coding agent ships to your own cloud: every push deployed, every change priced before it runs and recorded after. | delivery rows; plan-with-impact row; "every change records the person and the coding agent" row |
| D | Your coding agent sets up Planton from one prompt, and from then on every push to GitHub deploys into your own AWS, Google Cloud or Kubernetes. | Setup Prompt row; delivery rows (ECS from the CLI and YAML) |
| E | Deploy from GitHub into your own cloud without writing pipelines: your coding agent does the setup, and Planton checks the cost and access of every change. | delivery rows (no pipeline YAML, Buildpacks); Setup Prompt row; plan-with-impact row |

None says "production waits for a person" without a condition: that is true on the Team plan only (ledger, the approval rows).

## The cold reads

Six readers spanning the knowledge spectrum read the five cold, shuffled and unlabelled, each answering what it does, whether it is for them, what they would do next, and where they got lost. A senior developer-tools copywriter scored them.

| Reader | Kept reading | Closed the tab |
|---|---|---|
| A developer who inherited ops at a 5-person company | B | C |
| A working application developer | A | C |
| A platform engineer at a 300-person company | B | A |
| A head of engineering at a 25-person startup | D (E second) | C |
| A non-technical founder who pays the cloud bill | B | D |
| A visitor from a forum thread with a coding agent open | B | C |

The copywriter ranked A first (8/10), B and D next (7.5), C (7), E (6.5), and suggested B's second half as the line that says why to trust it.

**B wins:** the middle of the panel restated it correctly without help, and nobody closed the tab on it. Four readers found "touch" vague, so v1 says "the cost and permissions", which is what the plan shows. The developer who inherited ops, the application developer and the non-technical founder read v1 cold again; all three restated it correctly.

**The edges' gaps, accepted** (the page answers them, the sentence doesn't have to): "permissions", whose? (the founder, the application developer); "every push, straight to prod?" (the developer who inherited ops: pushes walk your environments in order, and on the Team plan production can wait for a person); "is it for me or for whoever runs deploys?" (the application developer: the buyer is whoever runs deploys); the clouds aren't named in the sentence (the line under it names them).

**C lost hardest:** "how your coding agent ships" assumes a workflow most readers don't have yet. The A-style sentence becomes the first action, under the sentence, because the platform engineer read it alone as "paste-and-pray".

## Batch 1: what comes off

134 changes across 11 surfaces. Each: where, before, after (or removed), why.

### `src/data/homepage.ts`

- **12 HOMEPAGE.eyebrow**: “THE SELF-SERVICE CLOUD PLATFORM” → removed (retired line)
- **16 HOMEPAGE.caption**: “AI-assisted infrastructure. Git-based delivery. Your team’s controls.” → “Your coding agent. Your cloud. A plan before every change.” (no AI inside the product)
- **10 HOMEPAGE.description (search snippet)**: “Set up cloud infrastructure, deploy applications from Git, and give your team a repeatable way to ship. See how Planton works in your cloud.” → “Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs. Start by pasting one prompt into your coding agent.” (the one sentence)
- **23 overview.intro**: “Platform engineers establish the foundation. Developers use it to get their applications running.” → “Your coding agent sets it up from one prompt. After that, every push deploys.” (retired persona)
- **49 agents.eyebrow**: “SELF-SERVICE DEVOPS / FROM YOUR EDITOR” → “FROM YOUR CODING AGENT” (retired framing)
- **54 agents.setup**: “Use Planton skills with the CLI in Cursor, Claude Code, or Codex. Supported agents can also connect through MCP. Connect your account and cloud before deploying; your team’s access and approval rules still apply.” → “On planton.ai, paste the Setup Prompt into Claude Code, Codex or Cursor from your app’s repository. The agent installs the CLI, connects your cloud and ships the first deployment, asking before anything costs money. Agents can also connect through MCP.” (lead with the Setup Prompt)
- **70 infrastructure.paragraphs[0]**: “Use the assistant to help compose the architecture, or work directly with its manifests. Review the resulting configuration before you deploy. The template stays readable and editable as your requirements change.” → “Write the manifests yourself, or have your coding agent write them. Review the resulting configuration before you deploy. The template stays readable and editable as your requirements change.” (no in-product assistant)
- **87 delivery.paragraphs[1]**: “Follow builds and logs, promote an artifact through your environments, and pause for approval where required. Each delivery records what shipped. When something fails, open the run and ask the assistant for help with the actual configuration and repository context.” → “Follow builds and logs, promote an artifact through your environments, and pause for approval where required. Each delivery records what shipped. When something fails, open the run and read its log and configuration.” (no in-product assistant)
- **98-125 HOMEPAGE.controls**: “Give developers freedom. Keep changes accountable. … The assistant works within the access of the person asking. Repository changes go through a reviewed pull request.” → removed (removed (unused; nothing renders it))
- **132-149 HOMEPAGE.adoption**: “Fits the cloud—and the way—you work. …” → removed (removed (unused; nothing renders it))
- **152 faq.title**: “Before we talk.” → “Before you start.” (start free, not a call)
- **162 faq: Can we use infrastructure we already have?**: “… Tracking a resource is different from deploying it; we can use the demo to examine the right adoption path for your stack.” → “… Tracking a resource is different from deploying it; the docs list the supported kinds and import paths.” (demo-first)
- **175-177 faq: What does the AI assistant do?**: “It helps compose infrastructure and investigate deployments using the context of the screen and the access of the person asking. …” → removed (the Assistant is not offered)
- **180-182 faq: What will we see in the demo?**: “A walkthrough of infrastructure setup, application delivery, and deployment controls, guided by your current stack and team’s needs. …” → “Q: How do I start? A: Start free on planton.ai: free for up to three seats, no card. Paste the Setup Prompt into your coding agent from your app’s repository.” (hosted is the front door)
- **186-191 close**: “See how Planton would work for your team. Bring your current stack. We’ll walk through infrastructure setup, application delivery, and the controls your team needs to ship with confidence. A technical walkthrough, shaped around your team.” → “Start with one repository. Start free on planton.ai: paste one prompt into your coding agent, and your first deployment lands in your own cloud. Free for up to three seats. No card.” (hosted is the front door)

### `src/components/landing-page/v6-2026-09-23/index.tsx`

- **35 hero doors**: “Book a Demo (the only button)” → “Start Free first; Book a Demo beside it as the second door” (hosted is the front door)
- **137, 186 controls and close doors**: “Book a Demo, then Start Free” → “Start Free, then Book a Demo” (hosted is the front door)

### `src/data/homepage-experience.ts`

- **8 HERO.description**: “Platform engineers define reusable environments and controls. Developers work in Planton or through their coding agent. GitHub supplies source changes. Planton deploys infrastructure and applications into your cloud and returns execution details.” → “You and your coding agent push to GitHub. Planton deploys infrastructure and applications into your cloud and returns execution details.” (retired persona)
- **9-10 HERO.team, compactTeam**: “Platform Engineers · Developers” → “Developers · Coding Agents” (retired persona)

### `src/data/homepage-video.ts`

- **10 OVERVIEW_VIDEO.description**: “Proven infrastructure charts. Agents that work within your policies. Production approvals stay with your team.” → “Every change priced and planned before it runs. Agents that work within your policies. Production approvals stay with your team.” (Planton ships no charts (the video itself still says it; a new render is the founder's))

### `src/data/workflow-explainers.ts`

- **65 Wait for a Person**: “Protected production waits for an authorized human decision. The assistant cannot approve this gate.” → “Protected production waits for an authorized human decision. A coding agent cannot approve this gate.” (no in-product assistant)

### `src/data/architecture-stories.ts`

- **6 scope (every architecture tab)**: “Illustrative architecture · deployment time compressed. Application code, data, and access configuration are supplied separately. …” → “Illustrative architecture · deployment time compressed. Planton creates these resources; your application code, data, and access configuration are supplied separately, and pushes from GitHub deploy to Kubernetes, Google Cloud Run or Amazon ECS. …” (no app delivery to Lambda, Container Apps or Workers)

### `src/data/positioning.ts`

- **26 umbrella.tagline**: “The Self-Service Cloud Platform” → removed (retired line)
- **28 umbrella.sentence**: “Planton turns your own cloud account into a self-service platform. AI designs the infrastructure, verifies the cost and permissions before anything is created, and publishes it as templates your whole team can deploy. Your services then ship onto that infrastructure straight from Git.” → “Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs.” (the one sentence)
- **new umbrella.firstAction**: “(none)” → “Start by pasting one prompt into your coding agent.” (the first action, under the sentence)
- **68 infraHub.analogy**: “Cursor for Cloud Infrastructure” → removed (no analogies; implied in-product AI)
- **69 infraHub.line**: “Describe what you need, watch it compose on a live canvas, see the monthly cost and the IAM policy before anything is created, deploy, and publish it as an Infra Chart — a template your team reuses.” → “Every infrastructure change is a plan that shows its monthly cost and the exact permissions it needs before anything is created. Only the plan you reviewed is applied, and what you built can be saved as an Infra Chart your team redeploys.” (implied in-product AI)
- **73 serviceHub.analogy**: “Vercel for Backend, In Your Own Cloud” → removed (no analogies; front-end hosting is not offered)
- **74 serviceHub.line**: “Connect a Git repository and every push becomes a running deployment — no pipeline YAML, no Dockerfile required, with results written back into GitHub checks and deployments.” → “Connect a GitHub repository and every push is built on Planton’s runners and deployed into your own Kubernetes, Google Cloud Run or Amazon ECS, with results written back to GitHub. No pipeline YAML, no Dockerfile required.” (GitHub deployments permission is missing)

### `src/data/story.ts`

- **98-104 chapter 2 What Planton Is**: “title: The Self-Service Cloud Platform · claim: Planton turns your own cloud account into a self-service platform. AI designs the infrastructure, …” → “title: What Planton Is · claim: Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs. Start by pasting one prompt into your coding agent.” (mirrors the story v2.0)
- **106 chapter 2 neverSay**: “an analogy for the umbrella · either hub analogy about the whole product · "Template" capitalized as a product name” → “"The Self-Service Cloud Platform" · "AI designs your infrastructure" · any in-product Assistant · an analogy for Planton · "production waits for a person" without "Team plan" · "Template" capitalized” (mirrors the story v2.0)
- **119 chapter 3 proof**: “The console, the CLI, and the assistant render the same server-stamped statement word for word.” → “The console and the CLI render the same server-stamped statement word for word.” (no in-product assistant)
- **129 chapter 4 claim**: “A platform team writes the rules once and every request obeys them, …” → “You write the rules once and every request obeys them, …” (one buyer)
- **132 chapter 4 proof**: “A change to a protected environment is a request: an approver reads the exact plan the engines previewed, and only that plan applies. Nobody approves work they initiated, the assistant included.” → “A change to a protected environment is a request: an approver reads the exact plan the engines previewed, and only that plan applies. A coding agent never approves its own person’s run, and on the Team plan production can require a second person.” (a sole approver may approve their own change)
- **151 chapter 5 neverSay**: “drift detection (a future consideration)” → “scheduled drift checks (drift is checked on demand only)” (drift on demand exists)
- **159 chapter 6 claim**: “… every push is built, containerized, and deployed, with the result written back into GitHub as checks and deployments.” → “… every push is built, containerized, and deployed, with the result written back to GitHub.” (GitHub deployments permission is missing)
- **166 chapter 6 neverSay**: “"Vercel for backend" about Planton” → “"Vercel for backend", about Planton or any hub · GitHub "deployments" written back” (mirrors the story v2.0)
- **203 chapter 9 claim**: “The platform engineer is the user: they write the rules, curate the catalog, publish the Infra Charts, and hand developers and agents a self-service platform that cannot break those rules. The engineering leader is who signs: what reaches them is the proof.” → “A developer who inherited ops, on a team of 2 to 10 with no platform team: you own the cloud account and the deploys because nobody else would, you already use a coding agent, and you want every push to land in your own cloud without a surprise bill or a surprise permission. If you already have CI, a production approval gate and an audit trail you trust, you may not need Planton.” (one buyer)
- **206 chapter 9 proof[1]**: “A consultancy delivered GCP infrastructure with no GCP expertise on the team, and runs one organization per client.” → removed (retired persona (the quote stays in chapter 10))
- **209 chapter 9 neverSay**: “built for developers as the headline · replaces your DevOps team · a persona we have not sold to described as a customer” → “replaces your DevOps team · a persona we have not sold to described as a customer · a page per persona” (the true headline is for developers)
- **223 chapter 10 neverSay**: “any savings figure · 50+ charts · …” → “any savings figure · any count of charts Planton ships · a framework name on a public page · …” (Planton ships no charts)
- **231 chapter 11 claim**: “Governance platforms observe your estate after the fact and tell you what to fix; Planton prevents at the moment of creation … Developer portals catalog what you have; Planton deploys what you need.” → “The question is "why not my coding agent plus Terraform and a GitHub Action?" The agent writes the change and stops when the chat ends. Planton is the loop that keeps going: it builds and deploys every push through your environments in order, prices and permission-checks every infrastructure change before it runs, holds production for a person on the Team plan, and records who asked and which agent made each change, with no stored cloud keys.” (retired comparison framing)
- **232-236 chapter 11 proof**: “Prevent at creation versus observe after … The execution layer behind the catalog: it deploys, with the record attached.” → “The agent writes the change; Planton keeps deploying after the chat ends. · Every infrastructure change is a plan with its cost and permissions, applied only as reviewed. · Every change records who asked and which coding agent made it.” (retired comparison framing)
- **244 chapter 12 claim**: “Start free today. Planton Desktop is free for individuals forever, commercial use included. The hosted free tier is free for up to 3 seats with no card. …” → “Start free today on planton.ai: free for up to 3 seats with no card, and the first thing you do is paste one prompt into your coding agent. Planton Desktop is free for individuals forever, commercial use included. …” (hosted is the front door)
- **249 chapter 12 neverSay**: “a monthly free AI allowance as live” → “AI credits or an AI allowance” (the Assistant is not offered)

### `src/components/product/ProductIndex.tsx`

- **20 forWhom (typed into the component; moves to its record)**: “For the platform engineer deciding what to hand their team, and the developer who wants to see what their agent gets.” → “For the developer who inherited ops and wants to see what Planton does in their cloud, and what their coding agent gets.” (retired persona; a claim typed in a component)
- **20 title (from the umbrella)**: “The Self-Service Cloud Platform” → “What Planton Is” (retired line)

### `src/data/site-pages.ts`

- **177 /product description**: “Infra Hub, Service Hub, your coding agent, the CLI, the catalog, import, and the open source underneath: what Planton is, as a platform engineer meets it.” → “Infra Hub, Service Hub, your coding agent, the CLI, the catalog, import, and the open source underneath: what Planton does in your cloud.” (retired persona)
- **185 /product/infra-hub description**: “Describe what you need, see the monthly cost and the IAM policy before anything is created, deploy, and publish it as an Infra Chart your team reuses.” → “Every infrastructure change as a plan with its monthly cost and the permissions it needs, applied only after you review it, and saved as an Infra Chart your team redeploys.” (implied in-product AI)
- **167 /trust/your-cloud-your-keys description**: “Planton runs in your account with your keys, your state, and your bill. …” → “Planton deploys into your account with your keys, your state, and your bill. …” (the control plane is at planton.ai)

### `src/data/product.ts`

- **70, 105 analogy kickers (Infra Hub, Service Hub)**: “Cursor for Cloud Infrastructure · Vercel for Backend, In Your Own Cloud” → removed (no analogies)
- **72 Infra Hub forWhom**: “For the platform engineer who publishes the templates, and the developer or coding agent who deploys them.” → “For the developer who inherited ops and wants to see what a change costs, and what access it needs, before it touches the cloud.” (retired persona)
- **74 Infra Hub step 1**: “Describe · Say what you need: in the console, from the CLI, or through your coding agent. Every kind is a typed schema, so a wrong field fails before it touches your cloud.” → “Write · You or your coding agent write the change as a manifest, from the CLI or the console. Every kind is a typed schema, so a wrong field fails before it touches your cloud.” (implied in-product AI)
- **75 Infra Hub step 2**: “Compose · Watch it compose on a live canvas: the components, how they connect, what each needs from the others.” → “See It Connect · The canvas shows the components, how they connect, and what each needs from the others.” (implied in-product AI)
- **107 Service Hub forWhom**: “For the developer who wants a push to mean deployed, and the platform engineer who wants that push to obey the environment’s gate.” → “For the developer who wants a push to mean deployed, into their own cloud, with production behind a gate.” (retired persona)
- **111 Service Hub step Promote**: “Promotion follows your Promotion Paths, one per product line; a protected environment carries a manual gate and refuses self-approval.” → “Promotion follows your Promotion Paths, one per product line; a protected environment carries a manual gate, and on the Team plan production can require a second person.” (a sole approver may approve)
- **112 Service Hub step Written Back**: “The result lands in GitHub as checks and deployments, so the person who pushed learns what happened where they already are.” → “The result is written back to GitHub, so the person who pushed learns what happened where they already are.” (GitHub deployments permission is missing)
- **127-128 Service Hub record rows**: “approval: prod gate resolved by a reviewer · never the initiator · github: check passed · deployment active” → “approval: prod gate resolved by a second person · Team plan · github: check passed” (precise approval; no deployments write-back)
- **140 Coding Agents forWhom**: “For the developer with Cursor, Claude Code, or Codex already open, and the platform engineer who wants that agent inside the same rules as everyone else.” → “For the developer with Cursor, Claude Code, or Codex already open who wants the agent to deploy without handing it the keys to production.” (retired persona)
- **142 Coding Agents steps (new first step)**: “Install the Skills (first step)” → “Paste the Setup Prompt · On planton.ai, a new organization gets one Setup Prompt. Paste it into Claude Code, Codex or Cursor from your app’s repository: the agent installs the CLI and skills, connects your cloud, creates an environment and ships the first deployment, asking before anything changes your cloud or costs money. (Install the Skills follows as the manual way.)” (lead with what already works)
- **151 Coding Agents point**: “One Assistant, Every Surface · The same skills power the Planton Assistant in Planton Desktop and the web console, so your agent and the platform’s own assistant work from the same instructions.” → removed (the Assistant is not offered)
- **207 Catalog forWhom**: “For the platform engineer choosing what their organization may deploy, and the reviewer who wants the fact sheet before the deploy.” → “For the developer choosing what to deploy, and the reviewer who wants the fact sheet before the deploy.” (retired persona)

### `src/data/trust.ts`

- **114 Rules and Approvals record row**: “who may approve: anyone with approve access on prod · never the requester” → “who may approve: anyone with approve access on prod · not the requester when someone else can approve” (a sole approver may approve their own change)
- **137 The Record honesty**: “Drift detection is not shipped. The record tells you what was deployed; it does not yet tell you what changed behind Planton’s back.” → “Drift is checked when you ask; there are no scheduled checks.” (false: drift is checked on demand)
- **168 Security Posture point**: “Four Framework Crosswalks · The HIPAA Security Rule, SOC 2 Trust Services Criteria, FedRAMP Moderate, and the CIS AWS Foundations Benchmark are mapped onto the control vocabulary, requirement by requirement, quoted from the published source.” → “Mapped Requirement by Requirement · Published security frameworks are mapped onto the control vocabulary, requirement by requirement, quoted from each source.” (a framework name reads as an attestation)
- **180-181 Security Posture record rows**: “HIPAA 164.312(a)(2)(iv): served by encryption at rest, customer-managed keys · SOC 2 CC6.1: served by no public exposure, role-based access control” → removed (a framework name reads as an attestation)
- **184 Security Posture record footer**: “Two crosswalk rows are shown; evidence is quoted from the kind’s profile. There is no verdict that says compliant, …” → “Evidence is quoted from the kind’s profile. There is no verdict that says compliant, …” (rows removed)
- **187 Security Posture honesty**: “Posture never becomes “compliant.” The correct sentence is: this kind enforces these controls, which serve these requirements. SOC 2 reports on an organization’s controls, not a product’s; HIPAA compliance is a property of an entity’s program.” → “Posture never becomes “compliant.” The correct sentence is: this kind enforces these controls, which serve these requirements. A framework judges an organization’s program, never a product.” (a framework name reads as an attestation)
- **188 Security Posture honesty**: ““FedRAMP authorized” is never said of any kind, chart, or deployment. The crosswalk is evidence an assessment consumes.” → removed (a framework name reads as an attestation)

### `src/data/doors.ts`

- **64 BUYER_DOORS (the Trust pages' buttons)**: “Book a Demo, then Pricing” → “Start Free, then Pricing” (hosted is the front door)
- **40 DOORS.hosted (every Start Free button)**: “Start Free → https://planton.ai/login” → “Start Free → https://planton.ai/signup” (Start Free opens sign-up, not sign-in)

### `src/data/distributions.ts`

- **48 Hosted: Where the Work Runs**: “A deploy executes on a runner: one Planton operates, your own laptop with the cloud sign-in already on it, or a standing appliance inside your network. Credentials are resolved at the moment of use: keyless where the cloud supports it, otherwise from the secrets manager just in time.” → “A deploy executes on a runner Planton operates. Credentials are resolved at the moment of use: keyless where the cloud supports it, otherwise from the secrets manager just in time. A runner of your own, on your laptop or inside your network, is coming to hosted organizations.” (your own runner on hosted is not released)
- **59 Hosted record row: runner**: “Planton’s, your laptop, or an appliance in your network” → “Planton’s · a runner of your own is coming soon” (your own runner on hosted is not released)
- **51 Hosted: Per Seat, Self-Serve**: “Teams pay per seat. Up to 25 seats, nobody talks to sales: the plan is a card away.” → “Teams pay per seat, by card. Nobody has to talk to sales.” (25 is the self-hosted license ceiling, not a hosted one)
- **62 Hosted record row: seats**: “3 free · then per seat · no sales call up to 25” → “3 free · then per seat, by card” (25 is the self-hosted license ceiling)
- **85 Self-hosted: Guided From the Desktop**: “Planton Desktop offers the same install as a guided experience: pick a cluster from your kubeconfig, preflight it, choose the front door, and watch the install converge.” → “Planton Desktop offers the same install as a guided experience: pick a cluster, preflight it, choose the front door, and watch the install converge.” (kubeconfig is never said)

### `src/data/desktop.ts`

- **136-137 ask.title, ask.lede**: “Two Ways to Ask. The Same Skills. · Ask from inside the tool you already use, or ask the assistant built in. Both run the same skills.” → “Ask From Your Editor. · Install the Planton skills once, and the coding agent you already use works from the real schema.” (the Assistant is not offered)
- **143-149 ask.assistant**: “The Built-In Assistant · Open the app and ask. The Planton Assistant converses with zero keys and zero accounts; Planton funds it. · Prefer to keep everything on your machine? Bring your own Anthropic or Cursor key …” → removed (the Assistant is not offered)
- **152 ask.close**: “Your agent and the platform’s own assistant share the same skills, the same catalog, and the same rules about what never happens without your say-so.” → “Your agent works from the same skills, the same catalog, and the same rules about what never happens without your say-so.” (the Assistant is not offered)
- **164, 280 runtime**: “Postgres, Temporal, and a Cache, as Native Processes · … a Java runtime, Postgres, Temporal, a cache, and the control plane.” → “Postgres, a Workflow Engine, and a Cache, as Native Processes · … a Java runtime, Postgres, a workflow engine, a cache, and the control plane.” (runner technologies never appear in public)
- **182 Infra Hub on your laptop kicker**: “Cursor for Cloud Infrastructure” → removed (no analogies)
- **186 Your Clouds, Found, Not Pasted**: “Planton reads the cloud logins already on your machine (AWS profiles, gcloud configurations, az subscriptions, kubeconfig contexts) and turns each into a ready connection. …” → “Planton reads the cloud logins already on your machine (AWS profiles, gcloud configurations, az subscriptions) and turns each into a ready connection. …” (Kubernetes from a kubeconfig is never said)
- **243 business model**: “… hosted seats beyond the 3 free ones, self-hosted licenses beyond the community edition’s 5 seats, and prepaid AI credits. …” → “… hosted seats beyond the 3 free ones, and self-hosted licenses beyond the community edition’s 5 seats. …” (no AI credits)
- **250 start body**: “Download, open, pick Run on This Computer, and ask for what you need in your own words. Or install the skills and ask from your editor.” → “Download, open, pick Run on This Computer, then install the skills and ask from your editor.” (the Assistant is not offered)
- **267 CLI note**: “The CLI runs the same schema lookups, validation, and deploys from a terminal; the desktop app adds the assistant, the canvas, and the local instance.” → “The CLI runs the same schema lookups, validation, and deploys from a terminal; the desktop app adds the canvas and the local instance.” (the Assistant is not offered)
- **284 Your Clouds Are Already There**: “Planton finds the AWS, Google Cloud, Azure, and Kubernetes sign-ins on your machine and offers each as a ready connection. Then ask for what you need.” → “Planton finds the AWS, Google Cloud and Azure sign-ins on your machine and offers each as a ready connection.” (kubeconfig never said; no assistant)

### `src/components/pricing/plan-grid.tsx`

- **173-180 every plan card's AI line**: “AI Assistant Included · Prepaid credits — spend protection on by default / AI Assistant — Coming Soon · Runs in your cluster with your own LLM provider key” → removed (the Assistant is not offered)

### `src/components/pricing/enterprise-band.tsx`

- **the Enterprise band**: “Enterprise starts at $60K/year — published rate card · Up to 250 Seats · SSO/SCIM · Air-Gap · Compliance Reporting · 24×7 SLA · See Enterprise Plans” → removed (no enterprise plan exists to sell)

### `src/app/(site)/pricing/enterprise/`

- **/pricing/enterprise (the whole page)**: “Enterprise at Planton: a published rate card in your market, enterprise identity, air-gap, compliance reporting, and real SLAs. Self-serve below 25 seats.” → removed (removed; the address answers the site's 404)

### `src/components/pricing/runners-note.tsx`

- **24-26**: “Self-hosted runners are included in every plan — they run inside your private network and connect outbound-only, so you never open an inbound firewall port.” → “Deploys run on Planton’s runners on every hosted plan, and runner minutes are never billed. A runner of your own, inside your network, is included on self-hosted and coming soon on hosted.” (your own runner on hosted is not released)

### `src/data/value-matrix.ts`

- **35 the Enterprise column**: “Enterprise (a fifth column)” → removed (no enterprise plan exists)
- **171-184 AI Assistant rows**: “Built-in AI Assistant · Prepaid AI Credits with Spend Protection” → removed (the Assistant is not offered)
- **211 Multi-Cloud Catalog**: “700+ catalog kinds across AWS, GCP, Azure, Kubernetes, and more” → “850+ catalog kinds across AWS, Google Cloud, Azure, Kubernetes, and more” (re-counted 2026-11-24)
- **228-238 Runners in Your Own Network; Credentials Stay in Your Environment**: “Free: Unlimited · Team: Unlimited / Free: included · Team: included” → “Free: Coming Soon · Team: Coming Soon (self-hosted columns unchanged)” (your own runner on hosted is not released)
- **281-291 SCIM, Access Transparency, Compliance Reporting**: “rows whose only cell is the Enterprise column” → removed (no enterprise plan exists)

### `src/components/pricing/faqs.tsx`

- **126-146 How do AI credits work?**: “AI usage is prepaid: you load a credit pack, usage draws it down, …” → removed (the Assistant is not offered)
- **148-163 Does the AI Assistant work on my own infrastructure?**: “Not yet. Today the assistant runs on planton.ai organizations and in the Planton desktop app; …” → removed (the Assistant is not offered)
- **171-172 refund answer**: “… Licenses carry a 14-day money-back guarantee. Unused AI credit packs are refundable within 14 days of purchase.” → “… Licenses carry a 14-day money-back guarantee.” (no AI credits)

### `src/data/pricing.ts`

- **54-55, 84, 111 creditPackStart; 89-102, 116-129 enterprise tiers**: “$10 / ₹499 smallest credit pack · Enterprise Standard $60K / ₹30L, Enterprise Plus $120K / ₹60L” → removed (nothing in the catalog sells them)

### `src/data/platform-stats.ts`

- **19-20 printed**: “700+ catalog kinds · 8 providers” → “850+ catalog kinds · 9 providers” (re-counted 2026-11-24)
- **34-43 CLOUD_PROVIDERS**: “AWS, Azure, GCP, Kubernetes, Cloudflare, DigitalOcean, Auth0, OpenFGA” → “AWS, Azure, GCP, Kubernetes, Cloudflare, DigitalOcean, Auth0, Stripe, OpenFGA” (Stripe joined the catalog)
- **47-53 exact counts**: “719 kinds · 8 providers · 718 control profiles · counted 2026-09-17” → “852 kinds · 9 providers · 851 control profiles · counted 2026-11-24” (re-counted 2026-11-24)

### `public/docs/index.md`

- **14 Overview**: “Planton turns your own cloud account into a self-service platform. AI designs the infrastructure, verifies the cost and permissions before anything is created, and publishes it as templates your whole team can deploy. Your services then ship onto that infrastructure straight from Git. It deploys to your cloud accounts …” → “Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs. Start by pasting one prompt into your coding agent. It deploys to your cloud accounts …” (the one sentence)
- **45 Overview: CI/CD card**: “… deploy to Kubernetes, ECS, Cloud Run, or Cloudflare Workers, and manage secrets and ingress configuration.” → “… deploy to Kubernetes, ECS, or Cloud Run, and manage secrets and ingress configuration.” (Workers is not a delivery target)

### `public/docs/platform/index.md`

- **14 Platform overview**: “Planton turns your own cloud account into a self-service platform. AI designs the infrastructure, … It deploys to your own cloud accounts …” → “Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs. Start by pasting one prompt into your coding agent. It deploys to your own cloud accounts …” (the one sentence)
- **68 pricing line**: “Seat-based pricing — automation is never metered for billing, and on planton.ai the AI Assistant runs on prepaid credits.” → “Seat-based pricing — automation is never metered for billing.” (the Assistant is not offered)

### `public/docs/security/index.md`

- **16 Security overview**: “Planton is a self-service cloud platform that touches some of the most sensitive parts of your infrastructure: …” → “Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs. To do that, it touches some of the most sensitive parts of your infrastructure: …” (retired line)
- **40 secrets**: “Secrets stored through Planton's Secrets Manager live in the secret backend you choose — …” → “Secrets stored through Planton's Secrets Manager can live in the secret backend you choose — …” (the default is Planton's vault)
- **85 roles**: “… Five standard roles apply across most resources: owner, admin, IAM admin, viewer, and member.” → “… Seven built-in roles apply in an organization: Organization Owner, Access Admin, Infrastructure Admin, Developer, Release Approver, Billing Manager and Auditor.” (the platform ships seven organization roles)
- **122 summary table: Secrets**: “Provider-native storage in your own vault; read-audited; never stored in the platform's database” → “Provider-native storage, in your own vault if you choose; read-audited; never stored in the platform's database” (the default is Planton's vault)

### `public/docs/infrastructure/infra-components.md`

- **20**: “… across 8 cloud providers and 700+ resource types.” → “… across 9 cloud providers and 850+ resource types.” (re-counted 2026-11-24)

### `public/docs/coding-agents.md`

- **16**: “Planton ships its assistant's knowledge as agent skills: …” → “Planton ships its knowledge as agent skills: …” (the Assistant is not offered)
- **31 last sentence**: “The same skills power the Planton Assistant in Planton Desktop and the web console, so your agent and the platform's own assistant share one craft.” → removed (the Assistant is not offered)
- **133**: “… The skill files are byte-identical to the checksummed archives that release published, and they are what Planton Desktop and the hosted assistant run at that version.” → “… The skill files are byte-identical to the checksummed archives that release published.” (the Assistant is not offered)

### `public/docs/teams-and-access/billing.md`

- **16**: “… A self-hosted deployment has no billing surface — its capabilities come from its license, and the AI Assistant is not yet available on self-hosted installs.” → “… A self-hosted deployment has no billing surface — its capabilities come from its license.” (the Assistant is not offered)

### `public/docs/ci-cd/self-managed-pipelines.md`

- **221**: “The publish check runs before the record is written — at the CLI, through the assistant's tools, and at the API: …” → “The publish check runs before the record is written — at the CLI and at the API: …” (the Assistant is not offered)
- **227**: “…; the assistant reads the same through its get_tekton_pipeline / get_tekton_task tools and search.” → removed (the Assistant is not offered)

### `public/docs/ci-cd/trusted-workflows.md`

- **118**: “through the Planton Assistant, which reads the same records and can walk you to the fix.” → removed (the Assistant is not offered)

### `public/docs/ci-cd/on-your-laptop.md`

- **82**: “The desktop assistant runs on the same commands, so everything the health dashboard shows is a verb:” → “Everything the health dashboard shows is a CLI verb:” (the Assistant is not offered)

### `public/docs/ci-cd/monorepo-support.md`

- **190**: “…; or ask the Planton Assistant ("also rebuild payment-api when proto/ changes") and it makes the same edit.” → removed (the Assistant is not offered)

### `README.md`

- **7 tagline**: “Deploy real infrastructure to your own cloud — without writing Terraform.” → “Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs.” (the one sentence)
- **9-13 pitch**: “Planton is a free desktop app and CLI you download and open. It finds the cloud you're already signed into; pick a stack and fill a short form, or `planton apply -f` a manifest from your terminal — and watch it deploy, with clean, auditable infrastructure-as-code running underneath. No account. No connections. No ceremony.” → “Start by pasting one prompt into your coding agent. On planton.ai, a new organization gets a Setup Prompt: paste it into Claude Code, Codex or Cursor from your app's repository, and the agent installs the CLI, connects the cloud you're already signed into, and ships your first deployment, asking before anything changes your cloud or costs money. Underneath is clean, auditable infrastructure-as-code from the open-source modules in this repository.” (hosted is the front door)
- **15-19 links**: “Download Planton · Free forever, including for commercial use · Runs on macOS · Linux · Windows” → “Start free on planton.ai · Free for up to 3 seats, no card · Or Planton Desktop, free on macOS · Linux · Windows” (hosted first, Desktop second)
- **37-38**: “On your own cloud. Nothing to lock you in.” → “On your own cloud. If you leave, you keep the modules and your manifests, and the open-source CLI keeps deploying them.” (say the exit, not a badge)
- **50-52 catalog**: “700+ catalog kinds across 8 cloud providers (AWS, GCP, Azure, Kubernetes, DigitalOcean, Cloudflare, Auth0, and OpenFGA)” → “850+ catalog kinds across 9 cloud providers (AWS, GCP, Azure, Kubernetes, DigitalOcean, Cloudflare, Auth0, Stripe, and OpenFGA)” (re-counted 2026-11-24)

### `plantonhq/.github profile/README.md`

- **1-5 comment**: “The printed counts (700+, 8, 18, 17, 4) come from site/src/data/platform-stats.ts …” → “The printed counts (850+, 9, 17, 4) come from site/src/data/platform-stats.ts …” (Planton ships no charts)
- **11 headline**: “The Self-Service Cloud Platform” → “Planton” (retired line)
- **15**: “Planton turns your own cloud account into a self-service platform. AI designs the infrastructure, the platform verifies the cost and permissions before anything is created, and the design is published as an Infra Chart, a template your whole team can deploy. Your services then ship onto that infrastructure straight from Git. It sits beside the coding agent and the cloud CLI you already use; nothing about how you work changes.” → “Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs. Start by pasting one prompt into your coding agent. It sits beside the coding agent and the cloud CLI you already use; nothing about how you work changes.” (the one sentence)
- **17 Infra Hub**: “Infra Hub is Cursor for cloud infrastructure: describe what you need, watch it compose, see the monthly cost and the IAM policy before anything exists, deploy, and publish it as an Infra Chart.” → “Infra Hub: every infrastructure change is a plan that shows its monthly cost and the exact permissions it needs before anything exists; only the plan you reviewed is applied, and what you built can be saved as an Infra Chart your team redeploys.” (no analogy; implied in-product AI)
- **18 Service Hub**: “Service Hub is Vercel for backend, in your own cloud: connect a repository and every push becomes a running deployment, with the result written back into GitHub checks and deployments. No pipeline YAML, no Dockerfile required.” → “Service Hub: connect a GitHub repository and every push is built and deployed into your own Kubernetes, Google Cloud Run or Amazon ECS, with the result written back to GitHub. No pipeline YAML, no Dockerfile required.” (no analogy; GitHub deployments permission is missing)
- **25**: “Protected environments refuse self-approval.” → “Protected environments wait for a person, and a coding agent never approves its own person's run.” (a sole approver may approve)
- **38 planton row**: “The catalog: 700+ component kinds across 8 cloud providers, … 18 Infra Charts. The planton CLI and the IaC engine. …” → “The catalog: 850+ component kinds across 9 cloud providers, … The planton CLI and the IaC engine. …” (re-counted; Planton ships no charts)
- **39 skills row**: “… so your agent composes, validates, prices, and deploys from the same craft the platform's own assistant uses.” → “… so your agent composes, validates, prices, and deploys from the same craft.” (the Assistant is not offered)
- **44-49 Start Tonight**: “Start Tonight · Planton Desktop is free for individuals forever … · From the terminal: brew install plantonhq/tap/planton …” → “Start Today · Start free on planton.ai: free for up to 3 seats, no card; paste the Setup Prompt into your coding agent. · Teach your coding agent … · From the terminal: the one install line (curl -fsSL https://planton.ai/install.sh …) · Planton Desktop is free for individuals forever … · For a team …” (hosted first; Homebrew 6+ refuses the untrusted tap)

## The claim check

Every after above was read against the re-checked ledger rows (2026-11-24) and the never-say list:

- **Held as written:** the one sentence and the first action; every "written back to GitHub"; every hub line (each claim is a ledger row: plan with cost and permissions, delivery into Kubernetes, Cloud Run or ECS); the catalog counts (852 kinds with a `spec.proto` across 9 providers, 851 with a control profile, counted from this repository on 2026-11-24); seven built-in organization roles (counted from the platform's built-in role definitions).
- **Reworded to the source's words:** every approval sentence now says "on the Team plan" or "not the requester when someone else can approve" (a sole approver may approve their own change); "drift" now says on demand, not "not shipped"; "runners in your own network" on hosted plans reads Coming Soon (a runner of your own waits for a release on hosted).
- **Cut:** every in-product Assistant and AI-credit sentence (the Assistant is not offered); the enterprise tiers and `/pricing/enterprise` (no plan in the catalog sells them); every framework name on the trust pages (a reader takes "SOC 2" on a vendor page for an attestation); the hub analogies; Cloudflare Workers as a delivery target; kubeconfig as a connection; "Temporal" on the desktop page (runner technologies never appear in public).
- **Numbers in prose** ("three seats", "850+") are rendered from `pricing.ts` and `platform-stats.ts`, never typed.

## Left for a later batch, on purpose

- **Home's headline and intro, and its layout:** batch 2 rebuilds Home around the sentence, with real demo-company clips.
- **The homepage film** says "Proven charts. A short form." in its baked frames; only the site-side label changes now. A new render is the founder's.
- **Legal:** the refund policy's "Prepaid AI Credits" section and the terms' AI Assistant (Beta) clauses (and the refund page's registry description) change with the trust and legal pages in batch 3.
- **Docs getting-started and the overview's Day 1–3 journey,** and the security overview's runner-first framing (`public/docs/index.md:57`, `security/index.md:22-26`): batch 4, after the stranger trials.
- **Every `mcp.planton.ai` mention** stays as it is (product.ts's MCP sample, `coding-agents.md`): its own issue, by the founder's standing default.
- **The README's How it works** ("pick a stack, fill a short form", `planton chart install`) is left: both are true of Planton Desktop and an organization's own charts.
