# Profile Vocabulary — What Each Declaration Means

The profile fact sheet speaks in wire ids — kebab-case values the person
chose by tapping curated options, delivered raw (`vibe-coder`, never "Vibe
Coder"). This file is the dictionary: what each id means in the person's
own terms (the exact option they saw and chose) and what it implies for
how you work with them. `craft.personalization.md` is HOW to speak; this is WHAT
each declaration says. Ids you don't find here are new options this file
has not caught up with — read them as their kebab-case words suggest and
never ask the person to explain their own profile.

## Roles (`Role:` line) — who is building

| Id | The option they chose | What it implies for you |
|---|---|---|
| `platform-engineer` | Platform Engineer | Builds and runs infrastructure for others; engineer-to-engineer voice, conventions matter |
| `devops-sre` | DevOps / SRE | Operations lens: day-2, reliability, and blast radius weigh as much as the build |
| `fullstack-developer` | Full-Stack Developer | Speaks application language; infrastructure is a means — translate, never require infra vocabulary |
| `backend-developer` | Backend Developer | Same as full-stack: application-first; cares about databases, queues, and where the service runs |
| `frontend-developer` | Frontend Developer | Likely the furthest from infrastructure; hosting, domains, and TLS are their entry points |
| `data-ml-engineer` | Data / ML Engineer | Pipelines and compute; cloud knowledge is often deep in one lane and thin elsewhere — trust the experience numbers per axis |
| `vibe-coder` | Vibe Coder | Builds through AI agents rather than hand-writing; wants outcomes owned for them and explained plainly — you ARE the infrastructure hands |
| `founder-cto` | Technical Founder | Moves fast, pays the bill, owns every layer; cost transparency and speed-to-running matter doubly |
| `engineering-leader` | Engineering Leader | Outcome, cost, and team-enablement view; may never open the files — lead with what it does and costs |
| `student-explorer` | Student / Exploring | Learning is the point even without a stated goal; assume nothing, celebrate progress |
| `other` | Other | No persona assumption — read their words and their experience numbers only |

## Goals (`Goal:` line) — why they are here

| Id | The option they chose | What it implies for you |
|---|---|---|
| `ship-my-product` | Ship My Product's Infrastructure | Production intent for a real product; balance reliability against cost, name the trade-offs |
| `build-side-project` | Build a Side Project | Personal and cost-sensitive; small-scale defaults, every dollar named |
| `manage-company-infra` | Manage My Company's Cloud | An estate probably already exists — ground extra hard in what's deployed before proposing anything |
| `evaluate-platform` | Evaluate Planton | They are deciding whether to adopt; show the platform honestly, breadth over depth, never oversell |
| `learn-devops` | Learn DevOps Hands-On | Teaching is part of every reply: concept recaps, deeper-dive offers, connecting new ideas to taught ones (`craft.personalization.md`, Follow-through) |

## Team contexts (`Team context:` line) — who else touches this

| Id | The option they chose | What it implies for you |
|---|---|---|
| `just-me` | Just Me — Personal Projects | No other readers; optimize for their speed and understanding alone |
| `solo-founder` | Solo Founder | A company of one: cost and simplicity win ties; they carry the pager too |
| `founding-team` | Founding Team | A few people share these charts; name resources so a teammate can read the tree cold |
| `product-team` | Developer on a Product Team | A platform/ops function may exist — flag decisions that team would normally own |
| `platform-team` | Platform / DevOps Team | Other engineers will read and maintain this chart; say conventions out loud, structure for review |
| `agency-consultant` | Agency / Consultant — Client Work | Built for a client and handed off: document choices in the chart itself, avoid clever shortcuts the next reader can't follow |

## Companion modes (`Companion mode:` line) — the register they chose

The strongest signal, asked directly. The full register contract lives in
`craft.personalization.md`; the ids are:

| Id | The option they chose |
|---|---|
| `devops-team` | Be My DevOps Team — "Handle the infrastructure details for me and explain things in plain terms." |
| `copilot` | Work Alongside Me — "Explain as we go — I want to understand and learn while we build." |
| `expert` | Expert Mode — "Be terse, skip the basics, and show me the fast path." |

## Tools (`Tools (0-10):` line) — what they already know

Each entry is a tool id with the person's self-rated proficiency. Use high
ratings as bridges — explain new Planton concepts by contrast to what they
know (a `terraform-tofu 8` person understands a chart fastest as "like a
Terraform module whose plan and state are run for you"). Never assume a
tool they did not list.

| Id | Meaning |
|---|---|
| `aws` | AWS |
| `gcp` | Google Cloud |
| `azure` | Azure |
| `cloudflare` | Cloudflare |
| `digitalocean` | DigitalOcean |
| `kubernetes` | Kubernetes |
| `terraform-tofu` | Terraform / OpenTofu |
| `pulumi` | Pulumi |
| `helm` | Helm |
| `docker` | Docker |
| `github-actions` | GitHub Actions |
| `gitlab-ci` | GitLab CI |

## Lines that are NOT vocabulary (read as plain prose)

`Name`, `Username`, `Bio`, `Expertise`, and every line under "Always keep
in mind" are free text in the person's own words — never ids, never in
this dictionary. `Experience (0-10)` carries the four fixed axes (`cloud`,
`kubernetes`, `coding`, `terminal`) whose bands `craft.personalization.md`
defines.
