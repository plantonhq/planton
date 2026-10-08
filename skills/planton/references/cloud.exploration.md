# Cloud Exploration and Consented Mutations

Exploration is part of the job — never send the user to another agent for a
lookup you can run. WHICH instrument runs it depends on the arm you resolved
at the start (SKILL.md, "Know your instruments"): on the CLI arm the cloud
provider CLIs (`aws`, `kubectl`, and their siblings) and the `planton` CLI
are the tools of the job; on the platform-tools arm the same questions ride
the platform's own read tools. Two regimes govern every command and every
tool call, identically on both arms:

## Read-only exploration — free, encouraged, proactive

This freedom is scoped to the CLOUD and the PLATFORM — lookups that travel
through the CLIs. It is never a license to explore the machine's filesystem
(see "The local filesystem" below).

Run read-only commands whenever they ground the chart or explain a failure.
No permission needed. Typical uses:

- **Ground a chart in reality**: `aws ec2 describe-vpcs`,
  `aws ec2 describe-subnets --filters …`, `aws eks describe-cluster`,
  `aws sts get-caller-identity` (which account am I looking at?).
- **Inspect what a deploy created**: `kubectl get pods -A`,
  `kubectl describe deployment …`, `kubectl get events --sort-by=…` — for
  clusters reachable from this machine's kubeconfig.
- **Planton lookups**: charts, Infra Stacks, pipelines, connections — see
  `craft.planton-cli.md`.

### The platform-tools arm: the same reads, through the platform

Without cloud CLIs, the same grounding rides the platform's read tools —
they answer from the organization's own connected cloud accounts, so no
login needs to exist where you run:

- **Cloud reads**: the VPC/subnet/security-group/cluster listing tools and
  the Kubernetes object reads (tools named `list_*`, `get_*`, `find_*`)
  cover the exploration above through the org's stored connections.
- **Platform lookups**: charts, Infra Stacks, pipelines, Infra Jobs, and
  connections each have list/get tools mirroring the `planton` commands in
  `craft.planton-cli.md` — the same four-step failed-deploy diagnosis works
  tool-for-command.
- **Identity adapts**: instead of an AWS profile or kubectl context, say
  which ORGANIZATION and which connection you are reading through — the org
  comes from your standing context, and exploring the wrong org
  produces confidently wrong charts just the same.

Rules that keep exploration honest:

- Read-only verbs only: `describe-*`, `list-*`, `get`, `logs`, `explain` —
  and on the platform-tools arm, only tools whose verb reads (`list_*`,
  `get_*`, `search_*`, `find_*`, `check_*`, `build_*`). A plan belongs
  here too: `planton tofu|terraform plan`, `planton pulumi preview` and
  `plan_infra_component` start a plan-only Infra Job that can never apply
  (`craft.planton-cli.md`). Anything that
  creates, modifies, or deletes is a mutation (below) — that includes
  `kubectl apply/delete/scale/rollout`, `aws … create-/delete-/
  modify-/put-/terminate-*`, `gh issue create`, AND every mutating platform
  tool (`apply_*`, `create_*`, `delete_*`, `destroy_*`, `run_*`, `rerun_*`,
  `resume_*`, `cancel_*`, `update_*`, `undeploy_*`, `purge_*`).
- Say which identity you are using when it matters: the active AWS
  profile/account and kubectl context — or, on the platform-tools arm, the
  org and connection — exploring the wrong account produces confidently
  wrong charts.
- Respect the user's context: never switch AWS profiles or kubectl contexts
  silently; propose the switch and say why.
- Missing CLI or credentials is a finding, not a dead end: adapt to the arm
  you are on and continue with what you have. Report a gap only when it
  blocks something the user asked for — never as ambient complaint.

### The local filesystem is NOT exploration territory

Read-only freedom is about cloud and platform state behind the CLIs — it
never extends to the machine. The attached workspace folder is the entire
filesystem you may touch on your own initiative: never `find`, `grep`,
`ls`, or read paths outside it — no scanning the home directory for other
charts, no browsing Documents or Desktop for examples. A path the user
pastes in chat, or a file your tools hand you, is invited — go there and
nowhere further. The CLIs read their own config (`~/.aws`, `~/.kube`) as
part of running them; that is their business, never a reason for YOU to
open those directories. Outside-the-workspace reads fire the operating
system's privacy prompts against the host app ("requesting data from
other apps"), which reads as spying regardless of your intent.

## Mutations — only on explicit ask, one confirmation each

You never mutate uninvited — not cloud state, not cluster contents, not
platform records. The rule is channel-blind: a mutating platform tool called
with the user's own identity changes the same real infrastructure a CLI
command would — it carries exactly the same discipline. When the user
explicitly asks you to change something:

1. **State the exact command or tool call** you intend to run and what it
   will change (blast radius in one plain sentence).
2. **Wait for a clear yes.** A vague "sure, go ahead" earlier in the
   conversation does not cover a new mutation — one confirmation per
   mutation, never a blanket approval for a chain of them.
3. **Run it, then report what actually happened** — including partial
   failures. Never summarize an error away.

A `planton` command that asks for a yes before it changes something never
answers for you: run without a terminal, it refuses with exit 1, says nothing
happened ("Nothing was revoked."), and names the flag that answers ahead of
time (`--force` or `--yes`) with the full command line. That refusal is not a
failure to route around. The person's yes from step 2 is what the flag stands
for, so rerun with it only for the mutation they confirmed, and report the
refusal as "not done" when you do not.

Prefer the platform-tracked path for anything Planton manages: a resource
deployed by a chart should change through the chart (edit + redeploy) or
through `planton` commands — not through raw `aws`/`kubectl` mutations. The
platform records the state of what it manages; mutating behind its back makes
that record lie, and the next deploy turns the lie into a confusing failure.
When the user insists on the raw path against a Planton-managed resource,
warn about the drift once, plainly, and proceed only on their confirmed yes.

Deploying a composed chart is never part of composition — it is offered,
and on the user's explicit ask you perform it under exactly this mutation
protocol (`infra.machine-deploy.md` has the command, the machine-login offer that
often precedes it, and the pipeline follow-through). Undeploying or purging
whole charts and Infra Stacks remains a workflow the user drives from the studio —
surface the need, let them drive it.
