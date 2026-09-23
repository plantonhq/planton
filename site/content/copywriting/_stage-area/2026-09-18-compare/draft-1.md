# Draft 1: the Compare page

Chapter 11 of the story as one page at `/compare`. Three kinds of tool sit near Planton; each is described by what it does and never by name. For each: what it does, what Planton does at the same moment, when a team runs both, and the page that proves it. Every sentence is chapter 11's, another chapter's proof sentence quoted by index, or a fact a page of this site already states; the chapter is named beside each. Never: a vendor's name, "DevSecOps", "FinOps", a feature table with a competitor column, "policy as code", drift detection as shipped, rules over spec content as shipped.

Conventions: titles and labels in Title Case; everything else sentence case. The record becomes `src/data/compare.ts`.

---

## Hero

**Title.** How Planton Compares (registry)

**Lede.** Three kinds of tool sit near Planton, and none of them owns the moment infrastructure is created. Each is described here by what it does, never by name: what it does, what Planton does at the same moment, and when a team runs both. (ch 11; "owns the moment" is the story's own spine sentence)

**For whom.** For the platform engineer who already runs some of these, and the leader who signed for them.

**Doors.** Start Free, Download Planton Desktop (the user's pair); the umbrella sentence under them, as every persona hero carries it.

## Where the Difference Is

The story's spine, verbatim: Proof at creation, not observation after: every change Planton deploys is priced before it exists with its coverage stated, judged against a budget, held to your rules, stamped with the controls it enforces, and left behind as an immutable record.

## 1. Beside Governance and Posture Platforms

**What they do.** Cost, security, and operations dashboards over your whole estate. They read what exists, detect what is wrong, and tell you what to fix, after it was created. (ch 11)

**What Planton does at the same moment.**
- Priced Before It Exists. Every deployment-changing job is born with a verified monthly cost: an exact figure with line items when the pricing rules can derive one, a range otherwise, and plainly "unpriced" when neither is possible. A zero never stands in for unknown. (ch 3, proof 0)
- Controls Stated With Evidence. Every covered component states which of a fixed list of 17 technical controls it enforces, with evidence for each claim. (ch 3, proof 3)
- Stamped on the Record. The full configuration is embedded into the job when it is created, and the job is immutable: the resource may change later; the job never does. (ch 5, proof 0)

**When you run both.** Keep the posture tool for the estate you already have and for whatever Planton did not create. Planton hands it a smaller problem: everything that went through Planton arrives priced, within budget, inside your rules, and stamped with the controls it enforces. A complement, never a replacement. (ch 11, proof 0)

**Proven at.** /trust/security-posture

## 2. Beside Infrastructure-as-Code Tools and Their Guardrails

**What they do.** Orchestrators that run the Terraform your team writes, and guardrail tools that check it at plan time, field name by field name. Some estimate cost at plan time. The writing is still yours. (ch 11)

**What Planton does at the same moment.**
- Typed Self-Service, Not Authoring. Every component is a typed schema over an open-source module. A person or an agent writes a short manifest, the schema validates it before anything touches your cloud, and the module that runs is the same one every time. (ch 11 claim; the validation docs the Coding Agents page already links)
- One Vocabulary, Not a Hundred Field Names. Every component reports its controls against the same fixed list, so what you check is one vocabulary, not each kind's field names. (ch 11, proof 1)
- Every Door Obeys the Same Rules. A platform team writes the rules once and every request obeys them, whether it came from a person in the console, a script on the CLI, or a coding agent at two in the morning. (ch 4 claim)

**When you run both.** Keep the orchestrator for the Terraform you already run outside Planton. Your Terraform stays yours: the modules are open-source Terraform and Pulumi, what you already run is adopted, not rewritten, and if you leave, you take your manifests and keep deploying them with the open-source CLI. (ch 11, proof 1; ch 8, proof 1)

**Proven at.** /product/open-source

## 3. Beside Developer Portals and Service Catalogs

**What they do.** A catalog of what your organization has: services, owners, documentation, scorecards. A portal tells you what you have; someone still has to build and staff everything it points at. (ch 11)

**What Planton does at the same moment.**
- Deploys What You Need. The execution layer behind the catalog: it deploys, with the record attached. (ch 11, proof 2)
- A Design Becomes a Template. Describe what you need, watch it compose on a live canvas, see the monthly cost and the IAM policy before anything is created, deploy, and publish it as an Infra Chart, a template your team reuses. (ch 2, proof 0)
- Every Push Becomes a Deployment. Connect a Git repository and every push becomes a running deployment, no pipeline YAML, no Dockerfile required, with results written back into GitHub checks and deployments. (ch 2, proof 1)

**When you run both.** Keep the portal as the front page of your engineering organization and let it point at Planton for the part it cannot do: creating the infrastructure and shipping the service, with the record attached. (ch 11)

**Proven at.** /product/infra-hub

## You Will Ask

- **Why not just Terraform?** Planton runs Terraform; it does not compete with it. Every component in the catalog is a typed schema over a pre-written, tested open-source module, shipped for both OpenTofu and Pulumi, and you choose which engine runs without changing your manifest. What changes is who writes what: a person or an agent writes a short manifest, the schema validates it before anything touches your cloud, and the module that runs is the same one every time. (ch 2, 11; question bank entry 1) → /product/open-source
- **Is this another abstraction layer I will be fighting in six months?** Each provider keeps its full native configuration; Planton does not pretend one cloud's network is another's. What is consistent is the structure and the workflow: every component a typed schema over an open-source module, reporting its controls against one fixed list. What you already run is adopted, not rewritten, and if you leave, the modules and your manifests go with you. (ch 11; question bank entry 1) → /product/import
- **Is it a set of operators reconciling my cluster?** No. Nothing runs in your cluster watching your resources. Every change is one stack job: it plans, pauses at the gates you set, applies from a runner in your own network, and is kept and queryable with the exact configuration embedded. A control loop cannot naturally pause between the plan and the apply; a job can, and that pause is where your approvals live. The trade is that nothing self-heals: a job runs when a person, a push, or an agent asks. (ch 5; question bank entry 7) → /trust/the-record

## Read Next

/trust and /product/open-source, then the doors again.

## Read cold, as the platform engineer with a posture tool and a Terraform pipeline already running

Would I stay? The lede tells me in one line that this page will not tell me to throw anything away, which is the first thing I check. Section 1 says what my posture tool will keep doing; section 2 says my Terraform stays mine and names the exit; section 3 is not about me. The operators question is the one I would ask third, and the concession (nothing self-heals) is the sentence that makes me trust the rest. Changed on this read: the second category's first point said "the schema refuses" (a refusal record is roadmap); it now says "validates". "Kept forever" became "kept and queryable" (the Record page's own correction).

## Read cold, as a senior copywriter for developer-tools sales pages

First screen says what the page is and who it is for and offers the doors. One claim per section with its proof beside it; each section ends on a door. The three "When you run both" paragraphs are the page's argument and they each say "keep": that is the promise. Changed on this read: "the writing is still yours" in section 2 read as praise; it now reads as the cost it is. The third category's "what they do" had a second sentence that sneered; it now states the fact (someone still has to build what it points at). Every label is Title Case; every body sentence is sentence case.
