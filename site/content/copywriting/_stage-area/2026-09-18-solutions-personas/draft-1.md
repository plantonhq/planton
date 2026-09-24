# Draft 1: the five persona records

The story told to one person at a time. Each record below becomes one entry in `src/data/personas.ts`; the page at `/solutions/<slug>` and the deck at `/decks/<slug>` render from it. Every sentence is that person's framing of a chapter of `the-planton-story.md` (v1.2), never a new claim; the chapter is named beside each beat and objection. Proof sentences are quoted from `story.ts` by index, so they are not repeated here.

Conventions: headlines and labels in Title Case; everything else sentence case. A beat is `lead` (a full section on the page, a slide in the deck) or `supporting` (a door on the page, a slide in the deck). Objections are drawn from what people actually asked in demos (the question bank), answered in the chapter's words.

---

## Platform Engineer (`/solutions/platform-engineer`)

**Headline.** Self-Service That Cannot Break Your Rules

**Who.** You run the platform your developers and their coding agents build on, and you are the one who gets paged when something they created is wrong.

**Wall.** Agents can already create infrastructure in your account. You cannot see what they made, what it costs, or whether it follows the rules you wrote down last quarter.

**Doors.** Download Planton Desktop (primary), Start Free.

**What decides it.**
- Rules Written Once. Budgets, protected environments, and a curated catalog hold whether the request came from the console, the CLI, or an agent at two in the morning. (ch 4)
- A Record, Not a Transcript. Every deploy is one stack job you can query by resource, environment, time, and outcome, with the exact configuration embedded. (ch 5)
- Your Terraform Stays Yours. Every module is open-source Terraform and Pulumi under Apache 2.0. What you already run is adopted, not rewritten, and if you leave you keep deploying your manifests with the open-source CLI. (ch 8, 11)

**Beats.**
1. the-wall, lead. Your developers' agents are already creating infrastructure. The question is not whether to allow it; it is whether what they create is verified, recorded, and yours to repeat. Proof [0]. Proven at /product/coding-agents.
2. your-rules-hold, lead. You write the budget, the protected environments, and the allowed catalog once. Every door reads the same rules and refuses the same things, so an agent cannot do what a person could not. Proof [0, 1, 2, 3]. Proven at /trust/rules-and-approvals.
3. verified-before-it-exists, lead. Before an agent's design exists, you see its monthly cost with its coverage stated, the least-privilege policy it needs, and the controls each component enforces. Proof [0, 2, 3]. Proven at /trust/verified-before-deploy.
4. every-deployment-leaves-a-record, lead. The deploy nobody watched is a stack job with the configuration embedded, the verdicts stamped, and the approver named. You read it; you do not reconstruct it. Proof [0, 1, 3]. Proven at /trust/the-record.
5. what-planton-is, supporting. Two halves: Infra Hub, where a design becomes a template your team redeploys, and Service Hub, where a push becomes a deployment inside your environments' gates. Proof [0, 1]. Proven at /product/infra-hub.
6. runs-where-you-decide, supporting. Start on your laptop for free and move to hosted or your own cluster with the same manifests. Connections can be keyless, and every module is open source. Proof [0, 1, 3]. Proven at /distributions.
7. services-ship-from-git, supporting. Once the infrastructure exists, developers connect a repository and every push obeys the promotion order and the gates you declared. Proof [1, 2]. Proven at /product/service-hub.
8. bring-what-you-have, supporting. What already exists is adopted and its state imported in one verified step, so the record covers your estate, not only what Planton created. Proof [0, 1, 2]. Proven at /product/import.

**You will ask.**
- Is this another abstraction layer I will be fighting in six months? Rules are written once over a fixed list of controls, not over a hundred field names, and your Terraform stays yours: the modules are open-source Terraform and Pulumi, and what you already run is adopted, not rewritten. (ch 11)
- What can the agent do that a person could not? Nothing. It comes through the same door as the console and the CLI, reads the same rules, and gets the same refusals. A protected environment pauses for a human, and nobody approves work they initiated, the assistant included. (ch 4)
- What happens to my environments if Planton goes away? Every infrastructure module is open source under Apache 2.0. You take your manifests and keep deploying them with the open-source CLI. In every shape it is your cloud account, your keys, your state, and your bill. (ch 8)

**Proof.** Sai Saketh; Balaji Borra.

**Deck opening.** Your developers' coding agents can already create cloud infrastructure in your account. Here is what happens when they do it through the rules you wrote.

**Siblings.** Engineering Leader; Security and Governance Leader.

---

## Engineering Leader (`/solutions/engineering-leader`)

**Headline.** What Your Team Deploys, Proven Before It Exists

**Who.** You sign for the cloud bill and answer for the outage. You do not want to open a console; you want to know what was deployed, what it costs, and that the rules held.

**Wall.** Your team says the agents are making them faster. Nobody can show you the cost of what was created before it was created, or who approved it.

**Doors.** Book a Demo (primary), Pricing.

**What decides it.**
- Cost Before Creation. Every deployment-changing job is born with a verified monthly cost, its coverage stated, and how much more or less it will cost than what runs today. (ch 3)
- No New Team. Your platform engineer stays the user and writes the rules once; developers and their agents self-serve inside them. What reaches you is the proof. (ch 9)
- A Record That Does Not Depend on Asking. Every change is one immutable stack job: configuration, cost, verdict, approver, outcome. Queryable whenever you want it, never reconstructed because you asked. (ch 5)

**Beats.**
1. the-wall, lead. The agents made your team faster and made the estate harder to answer for. Nobody priced what they created, nobody checked the permissions, and nothing remembers what was made. Proof [0]. Proven at /product/coding-agents.
2. verified-before-it-exists, lead. Cost, permissions, and controls become properties of what gets deployed, checked before it exists rather than found after. Proof [0, 1, 3]. Proven at /trust/verified-before-deploy.
3. your-rules-hold, lead. A budget on an environment pauses the deploy that would exceed it, and a protected environment refuses self-approval. The rule holds whoever asked. Proof [0, 1]. Proven at /trust/rules-and-approvals.
4. every-deployment-leaves-a-record, lead. The record is what you get to see: what was deployed, what it cost, who approved it and why. It is there whether or not anyone asks. Proof [0, 1, 2]. Proven at /trust/the-record.
5. who-it-is-for, supporting. The platform engineer is the user and you are who signs. The proof is what travels between you. Proof [0, 1]. Proven at /solutions/platform-engineer.
6. what-planton-is, supporting. One platform in your own cloud account: AI designs, the platform verifies, the design becomes a template, and services ship from Git onto it. Proof [0, 1]. Proven at /product.
7. runs-where-you-decide, supporting. Hosted, self-hosted, or on a laptop; in every shape it is your account, your keys, your state, and your bill, with keyless connections and open-source modules. Proof [0, 1]. Proven at /trust/your-cloud-your-keys.

**You will ask.**
- Do I have to open a console to get any of this? No. The proof reaches you as the record: the cost fact before creation, the budget verdict and who resolved it, the controls each component enforces, and the immutable job behind every change. Your platform engineer is the user; you read what they hand you. (ch 9)
- Is there a savings number I can put in a budget? No, and we will not invent one. What you get is the verified monthly cost of each change before it is committed, with its coverage stated, judged against the budget you set. A dollar-savings figure would be a claim nobody could check, so you will not find one here. (ch 3)
- What does this cost? Teams pay per seat, and below the self-serve ceiling nobody talks to sales. The hosted free tier is free for up to {FREE_TIER_SEATS} seats with no card, and Planton Desktop is free for individuals forever, commercial use included. (ch 12)

**Proof.** Rohit Reddy Gopu.

**Deck opening.** You will not see a dashboard today. You will see what reaches you after your team's agent deploys something: the cost before it existed, the rule that held, and the record.

**Siblings.** Platform Engineer; Security and Governance Leader.

---

## IT Consultancy (`/solutions/it-consultancy`)

**Headline.** Repeatable Client Environments, Handed Back as Manifests

**Who.** You stand up environments for clients who mandate a cloud your team may not know, and you hand the work back when the engagement ends.

**Wall.** Every client starts from zero, and the Terraform you wrote for the last one does not fit the next.

**Doors.** Start Free (primary), Download Planton Desktop.

**What decides it.**
- One Organization per Client. Each client is its own organization, with its own cloud connection, environments, budgets, and record. (ch 9)
- Repeatability Is the Product. Publish the environment you built for one client as an Infra Chart and deploy it into the next. A prompt cannot be redeployed; a chart can. (ch 2)
- Hand Back Everything. Every module is open source under Apache 2.0. When the engagement ends, the client keeps their manifests and keeps deploying them with the open-source CLI. (ch 8)

**Beats.**
1. the-wall, lead. Every client starts from zero, on a cloud they chose and your team may not know, and the Terraform from the last engagement does not fit this one. Proof [0]. Proven at /product/catalog.
2. what-planton-is, lead. Describe the client's environment, verify it, deploy it, and publish it as an Infra Chart. The next client starts from the chart, not from a blank prompt. Proof [0, 1]. Proven at /product/infra-hub.
3. runs-where-you-decide, lead. The client's account, the client's keys, the client's bill. Connections can be keyless, so you never hold a long-lived credential for an account you do not own. Proof [0, 1, 3]. Proven at /trust/your-cloud-your-keys.
4. verified-before-it-exists, supporting. The monthly cost with its coverage stated, before the client's infrastructure exists. The estimate is the conversation with the client, not the invoice. Proof [0, 1]. Proven at /trust/verified-before-deploy.
5. every-deployment-leaves-a-record, supporting. Every change in every client organization is a stack job you can hand over: what was deployed, what it cost, who approved it. Proof [0, 1]. Proven at /trust/the-record.
6. services-ship-from-git, supporting. Connect the client's repositories and every push is built and deployed through the environments you declared. Proof [0, 1]. Proven at /product/service-hub.

**You will ask.**
- The client mandated a cloud my team has not worked in. Can we deliver? One of the people quoted below did exactly that: the client mandated GCP, the engineer had no GCP experience, and the consultancy delivered the whole environment. The catalog is {DEPLOYMENT_MODULE_COUNT} component kinds across {CLOUD_PROVIDER_COUNT} providers, each a typed schema over a tested module. (ch 10)
- What does the client keep when we leave? Everything. Their manifests, their state in their own backend, and the open-source CLI that deploys the same modules. It was their cloud account and their keys from the first day. (ch 8)
- Do we redo the work when a client wants a second environment? No. Publish the environment as an Infra Chart and deploy it into staging or the next region. The same manifests and the same model run on every shape, so nothing is redone. (ch 2, 8)

**Proof.** Rohit Reddy Gopu; Balaji Borra.

**Deck opening.** Your last client's Terraform does not fit this one. Here is how the environment you build today becomes the template you deploy for the next.

**Siblings.** Startup Founder; Platform Engineer.

---

## Startup Founder (`/solutions/startup-founder`)

**Headline.** Ship Without an Ops Hire, and Redo Nothing Later

**Who.** You are shipping a product with a small team and no ops hire, and every hour on infrastructure is an hour not on the product.

**Wall.** Your agent can build the infrastructure. You are not sure what it built, what it will cost next month, or how you will do it again for staging.

**Doors.** Start Free (primary), Download Planton Desktop.

**What decides it.**
- Push to Deploy. Connect a repository; every push is built, containerized, and deployed. No pipeline YAML and no Dockerfile required. (ch 6)
- The Monthly Cost, Before It Exists. Before your agent's design is created, you see what it will cost each month with its coverage stated honestly, never a zero that means unknown. (ch 3)
- Free to Start, Nothing Redone. The hosted free tier is free for up to {FREE_TIER_SEATS} seats with no card. The same manifests run when you are a team, so nothing is redone. (ch 12, 8)

**Beats.**
1. the-wall, lead. Your agent built the infrastructure in an afternoon. What it built is unpriced, unrecorded, and cannot be redeployed for staging from anything but a new prompt. Proof [0]. Proven at /product/coding-agents.
2. what-planton-is, lead. Your agent designs through Planton instead of around it: the design is verified, deployed, and published as a template you redeploy for staging. Proof [0, 1, 2]. Proven at /product/infra-hub.
3. services-ship-from-git, lead. Once the infrastructure exists, push. Every push is built and deployed, promotion follows the order you declared, and the result lands in GitHub, where you already are. Proof [0, 1]. Proven at /product/service-hub.
4. verified-before-it-exists, supporting. The monthly cost of what your agent designed, before it exists, so next month's bill is not a surprise. Proof [0, 1]. Proven at /trust/verified-before-deploy.
5. runs-where-you-decide, supporting. Your cloud account, your keys, your bill. Start hosted with no card, or on your laptop for free; the same manifests move with you. Proof [1, 3]. Proven at /distributions/hosted.
6. every-deployment-leaves-a-record, supporting. When you hire the first engineer who asks what is running and why, the answer is a record, not a memory. Proof [0, 1]. Proven at /trust/the-record.

**You will ask.**
- I do not know what a multi-AZ database is. Is this for me? The prompt can be one sentence of intent: say what you built and where you want it to run. The depth is there when you have it, and every component is a typed schema, so a wrong field fails before it touches your cloud. (ch 2)
- Why not let my agent write the Terraform? It can, and you get different Terraform every time, with nobody pricing it and nothing remembering it. Through Planton the agent writes a small validated manifest; the module that runs is pre-written, tested, and open source; and the design becomes a template you redeploy. (ch 1, 2)
- What happens when there are five of us? Nothing is redone: the same manifests and the same model run on every shape. The hosted free tier covers {FREE_TIER_SEATS} seats with no card; after that, teams pay per seat, and below the self-serve ceiling nobody talks to sales. (ch 8, 12)

**Proof.** Rakesh Kandhi.

**Deck opening.** Your agent can build your infrastructure tonight. Here is how to make sure you can afford it, repeat it, and hand it to your first hire.

**Siblings.** IT Consultancy; Platform Engineer.

---

## Security and Governance Leader (`/solutions/security-and-governance-leader`)

**Headline.** Rules That Hold Before Anything Exists

**Who.** You own the posture of an estate you did not build, and the tools you have tell you what went wrong after it did.

**Wall.** Agents create infrastructure faster than your scanners find the problems. Prevention has to happen where creation happens.

**Doors.** Book a Demo (primary), Pricing.

**What decides it.**
- Prevention at the Write Boundary. Budgets, protected environments, a curated catalog, and managed secrets hold before anything exists, whether the request came from a person, a script, or an agent. (ch 4)
- Controls Stated, Never Asserted. Every covered component states which of {CONTROL_COUNT} technical controls it enforces, with evidence for each claim, and never calls itself compliant. (ch 3)
- A Record of Every Change. Every change is one immutable stack job, queryable by resource, environment, time, and outcome, with identity tags on every resource it created. (ch 5)

**Beats.**
1. the-wall, lead. Agents create infrastructure faster than scanners find the problems in it. Detection after creation cannot keep pace with creation. Proof [0]. Proven at /product/coding-agents.
2. your-rules-hold, lead. The rules are enforced at the one place infrastructure is created, so a coding agent cannot do what a person could not, and the refusal is identical at every door. Proof [0, 1, 2, 3]. Proven at /trust/rules-and-approvals.
3. verified-before-it-exists, lead. Least-privilege permissions derived from exactly what is composed, and the technical controls each component enforces, stated with evidence before it exists. Proof [2, 3, 4]. Proven at /trust/security-posture.
4. every-deployment-leaves-a-record, lead. The immutable stack job: configuration, verdicts, approver, outcome, and identity tags on every resource created. Evidence that exists whether or not anyone asked. Proof [0, 1, 3]. Proven at /trust/the-record.
5. how-it-compares, supporting. Posture platforms observe after the fact; Planton prevents at creation. They complement each other, and neither replaces the other. Proof [0, 1]. Proven at /trust/security-posture.
6. runs-where-you-decide, supporting. Keyless connections mean no long-lived cloud credential is stored anywhere. Self-hosted runs on your cluster with a license that verifies offline. Proof [0, 1, 2]. Proven at /trust/your-cloud-your-keys.
7. what-planton-is, supporting. One platform in the customer's own account, with one door every request goes through. That door is where the rules live. Proof [0, 1]. Proven at /product.

**You will ask.**
- Is this compliant with SOC 2 or HIPAA? No component is ever called compliant, and no deployment gets a framework verdict. What you get is each component's stated controls with evidence, and {FRAMEWORK_CROSSWALK_COUNT} framework crosswalks that map those controls to HIPAA, SOC 2, FedRAMP Moderate, and CIS AWS Foundations so your assessor can read them. (ch 3, 5)
- Which rules hold today? Deployment budgets that pause a deploy exceeding them; protected environments that refuse self-approval; a catalog curated to the kinds you allow, refused identically at every door; sensitive fields that take only a managed secret. Every one is stated here as it works today. (ch 4)
- Does this cover what already exists in the account? Infrastructure that already exists can be adopted and its live state imported and verified in one step; from then on it carries the same record as everything Planton created. Import recipes exist for S3 buckets, VPCs, security groups, and container registries, proven in a live round trip. (ch 7)

**Proof.** None on record for this persona; the page shows the counts and no quote. A paraphrase is not a testimonial.

**Deck opening.** Your scanners tell you what went wrong after it did. Here is what it looks like when the rule holds before the resource exists.

**Siblings.** Engineering Leader; Platform Engineer.

---

## The index (`/solutions`)

Chapter 9 as the hero (title, claim); the two people in every sale as two cards (the platform engineer is the user; the engineering leader is who signs, quoted from the chapter's claim); then five doors, one per persona, each showing the persona's name and wall; then the user's doors.

## Registry titles and descriptions

- `/solutions`: Solutions. Planton for the platform engineer who runs the platform, the leader who signs for it, the consultancy that delivers it, the founder who ships on it, and the security leader who governs it.
- `/solutions/platform-engineer`: Planton for Platform Engineers. Self-service your developers and their coding agents cannot break: rules written once, cost and permissions verified before anything exists, a record of every deploy.
- `/solutions/engineering-leader`: Planton for Engineering Leaders. What your team deploys, proven before it exists: the cost, the rule that held, and the record, without a new team and without opening a console.
- `/solutions/it-consultancy`: Planton for IT Consultancies. One organization per client, a client environment from a published template, and everything handed back as manifests when the engagement ends.
- `/solutions/startup-founder`: Planton for Startup Founders. Ship without an ops hire: push to deploy, the monthly cost before it exists, free to start, and nothing redone when you become a team.
- `/solutions/security-and-governance-leader`: Planton for Security and Governance Leaders. Rules that hold at the moment of creation, controls stated with evidence and never called compliant, and a record of every change; a complement to your posture tools.
- `/decks/<slug>`: The Planton Story for <Persona, plural>. The story told for <persona, plural>, with presenter notes under every slide. Unindexed.
