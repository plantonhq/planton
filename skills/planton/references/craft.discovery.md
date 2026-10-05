# Discovery — Know the Person, the System, and the Motive

Composition quality is decided by what you know about the person, what
already exists in their Planton, and what the chart is FOR. But knowing is
not interviewing: **discovery never stands between the user and their first
architecture.** Everything here that is a LOOKUP runs silently before you
build (SKILL.md Phase 0); everything that is a QUESTION fires after the
user has something real to react to (Phase 4a) — refinement instruments,
never an entry toll. Skip what the conversation has already answered —
repeating a settled question is worse than not asking.

## 1. Ground in their Planton first (look, don't ask — before building)

Before writing anything, look up what the platform already knows — these
are read-only and fast (see `craft.planton-cli.md` for exact commands):

- The active org and environment (`planton context get`).
- Existing Infra Charts and their descriptions.
- Existing Infra Stacks and their deploy status — what has actually been
  built, and did it succeed?
- Available provider connections (which clouds, which Kubernetes clusters).
- Environments in the org.
- The organization's catalog availability — which kinds its catalog policy
  disables (`catalog-availability.md`: check here, design complete anyway,
  disclose before deploy).

What you find shapes the BUILD, not a question: an existing green cluster is
something to build on, not duplicate; a connection to exactly one cloud
settles the provider assumption. Fold the findings into the explain-after —
"I saw your `eks-cluster` in `dev` is green, so I wired the workloads to it"
— and never ask the user to describe infrastructure the platform already
knows.

Grounding is platform lookups through the CLI — never filesystem searches.
Other charts that may exist on the user's machine are not context: the
attached workspace is your entire filesystem (SKILL.md, Hard boundaries),
and `planton` commands answer everything the platform knows.

## 2. Read the person from their words (signals, not a questionnaire)

Their words JOIN the profile fact sheet, never replace it: the sheet
(standing session context — companion mode, experience numbers, goal,
Always lines) is the person's own deliberate answer to "how should you work
with me", so on any conflict ABOUT THE PERSON the sheet wins — a request
worded like an expert's from a `cloud 0` profile still gets the teaching
register (`craft.personalization.md`). Words stay authoritative about the TASK.

The user's first message usually answers most of what an interview would
have asked. Read it for:

1. **The goal, in product terms.** "A REST API with a Postgres backend"
   tells you the architecture; "a VPC and an ECS service" tells you their
   guess at one. Build to the goal either way — and translate for
   developers (below).
2. **Specify or delegate?** A user who names CIDRs, subnets, and instance
   classes is an expert — honor every specific they gave and assume only in
   the gaps. A user who says "I just need somewhere to run my app" has
   delegated — design confidently, explain as you go.
3. **Purpose.** Dev, production, or test bed — when stated, it drives the
   defaults (section 3). When unstated, take the dev-shaped assumption and
   NAME it in the register; do not stop to ask.
4. **Review posture.** Build-then-review is the default — the user reacts
   to rendered resources on a live canvas. An explicit "walk me through
   the plan first" flips it for this conversation; honor it without
   negotiation.

**The developer persona is the common case.** Developers speak application
and environment language — "my app", "staging", "route it on a hostname" —
not infrastructure language. Translating that into VPCs, gateways, and
connections is YOUR responsibility, never theirs. Mirror their vocabulary
back when you explain what you built, and never require them to learn a
resource kind to get what they want.

## 2a. Advertise what you can do (users don't know)

Users do not know you can explore their cloud, inspect their cluster, or dig
into their Planton records yourself. At the first moment where grounding
would help, SAY what you can do and what you need:

> "If the AWS CLI on this machine is configured for your account, I can look
> at your current ECS setup myself instead of asking you to describe it —
> want me to?"

Same pattern for `kubectl` (inspecting a cluster) and `gh` (filing gaps).
When the tool or credentials are missing, name exactly what to install or
authorize and offer to help set it up. Proactively asking for the tools that
make you self-sufficient IS the collaborative experience — silence about
your own capabilities forces the user to do machine work by hand.

## 3. The motive drives the defaults — assumed first, refined after

The same request ("an EKS cluster") composes differently depending on what
it is for:

- **Dev sandbox / learning** — minimize cost: single NAT gateway, small
  nodes, public endpoint, no multi-AZ redundancy. Say what you cheapened and
  why that is fine here.
- **Production** — resilience first: multi-AZ, private endpoint (with its
  runner implications), right-sized nodes, deletion protection where the
  schema offers it.
- **Cost-conscious test bed for a real workload** — production shape at
  minimum viable scale.

**When the motive is unstated, assume the dev shape and build** — it is the
cheapest to run, the cheapest to reshape, and the honest default for a first
draft. The assumption goes into the register with its consequence attached:
"I assumed dev-scale (~$X/month); production hardening would add multi-AZ
and roughly $Y." That one sentence does what the old motive question did —
without making the user answer a form before seeing anything.

Recommend with reasons — "for a dev environment I'd use a single NAT gateway:
each extra gateway bills around the clock (the exact monthly figure is in its
cost fact-sheet) and you don't need the redundancy" — never a neutral menu of
options. You are the experienced partner, not a form.

## 4. The refinement conversation (after delivery — where questions belong)

With the architecture on the canvas, ask the two or three questions that
would most change it — calibrated to the person (section 2's signals), never
a questionnaire. Good refinement questions attack the biggest assumptions
first: purpose (if unstated), the domain they own, what already exists that
the chart should connect to, compliance or cost ceilings. Each answer loops
through edit → rebuild → re-explain what changed.

Discovery is not a phase that ends: when the conversation reveals a new fact
(they mention a domain they own, an existing database, a compliance need),
fold it in — look it up if the platform knows it, adjust the chart, and say
what changed.
