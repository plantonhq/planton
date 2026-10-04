# The Deployment Model — What Happens After Compose

You compose the chart; the platform runs it. Knowing the machinery lets you
set expectations and diagnose failures — but it is background knowledge, not
conversation material. **Educate only when it serves the user's next step**
(they clicked deploy and wonder what is happening; a pipeline failed and you
are explaining where; they ask). Never lecture the model at someone who just
wants their infrastructure up.

## The path from chart to real infrastructure

```
InfraChart (the template you compose)
  → deploy: InfraStack (the chart rendered with the user's values — a
    versioned record of exactly what was requested)
  → InfraPipeline (executes the stack's dependency graph in order)
  → each pipeline node = one InfraComponent (one manifest from the chart)
  → each node's deploy = an Infra Job (one IaC engine run for that component)
  → real cloud infrastructure
```

Practical implications worth sharing at the right moment:

- Deploying a chart does not "run the chart" — it creates an Infra Stack, whose
  pipeline deploys component by component in dependency order. Independent
  components run in parallel; a failed node stops its dependents only.
- A failed pipeline is diagnosed node by node: find the failed node, read its
  Infra Job's error and logs (`craft.planton-cli.md` has the exact commands).
- Redeploying after a chart fix creates a new pipeline run; already-green
  resources converge (no duplicate infrastructure).

## Every resource deploys through an open-source module

Each catalog kind is deployed by its IaC module in the open-source
repository `github.com/plantonhq/planton`, at:

```
catalog/<provider>/<kind-lowercase>/iac/
  ├── tf/       # the Terraform/OpenTofu module
  └── pulumi/   # the Pulumi (Go) module
```

Which engine runs is the deploying organization's setting — **the default is
OpenTofu (`tf/`) unless the org or manifest explicitly says otherwise**. Never
set a provisioner in a chart.

The module source is the ground truth for what a spec field actually DOES in
the cloud — one level deeper than the schema. Use it when:

- A spec field's cloud-side effect is ambiguous and the choice matters.
- A deploy failed inside the engine and the error names cloud-provider
  concepts the manifest never mentions — map the failing provider resource back
  to the module code that creates it, then back to the spec field feeding it.
- You need to know a default the schema does not state (what the module does
  when a field is empty).

To inspect: ask the user's permission ONCE per session to clone the
open-source repo read-only (`git clone --depth 1
https://github.com/plantonhq/planton /tmp/planton-oss` — say why), then read
the kind's `iac/tf/` (or `iac/pulumi/` when that engine is in play). Cloning
writes only to the local temp directory and touches no infrastructure; it
still gets one honest ask because it downloads a large repository onto their
machine. Combine what the module does with what the user wants to produce the
most accurate chart — this is what separates a partner from a form-filler.
