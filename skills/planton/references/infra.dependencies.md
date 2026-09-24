# Wiring Resources Together

Charts rarely stand alone — a VPC feeds subnets, subnets feed a cluster, a
cluster feeds add-ons. Planton resolves these links at deploy time using a
dependency graph built from your references.

This file covers values that RESOURCES produce (`valueFrom`). Values that
OPERATORS manage — credentials and org/env config — are the other reference
family, `$var`/`$secret` (`config-references.md`). One test tells them
apart: if deploying something creates the value, wire `valueFrom`; if a
person or team owns the value, reference it from the config manager.

References are not fenced by the chart: at deploy time a reference resolves
against everything deployed in the organization and environment, so a chart
can wire to resources another chart created. The "References cross chart
boundaries" section below is the judgment that prevents the single worst
wiring failure — asking the user to hand-copy a value the platform already
knows.

## valueFrom — the primary wiring mechanism

When resource B needs an ID or ARN that resource A creates after deployment,
never paste a literal. Reference A's stack output:

```yaml
spec:
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: "{{ values.env }}-vpc"
      fieldPath: status.outputs.vpc_id
```

Rules:

- **kind** — PascalCase cloud resource kind of the producer.
- **name** — must match the producer's `metadata.name` exactly (including
  template expressions — both sides must render to the same string).
- **fieldPath** — `status.outputs.<outputName>`. Both snake_case
  (`status.outputs.vpc_id`) and camelCase (`status.outputs.vpcId`) are
  accepted: the validator tries the exact proto field name first, then
  converts camelCase to snake_case. The schema report's `outputs[]` lists
  the camelCase names; the fleet mostly writes snake_case. Either resolves
  to the same output — pick one style per chart and stay consistent.

Fields typed `string | valueFrom` in the schema report accept either a plain
string or the `valueFrom` block above.

### Nested valueFrom

Some fields nest the reference one level deeper (e.g. route targets):

```yaml
routes:
  - destinationCidrBlock: 0.0.0.0/0
    targetType: internet_gateway
    targetId:
      valueFrom:
        kind: AwsInternetGateway
        name: "{{ values.env }}-igw"
        fieldPath: status.outputs.internet_gateway_id
```

## Finding output field paths

The multi-cloud-catalog skill answers this class fastest: the producer
kind's reference page lists every output leaf in its `## Outputs` table,
and its `## References` / `## Referenced By` tables answer the wiring
question in BOTH directions — what this kind can point at, and what can
point at it — a class no schema dump carries (`reference-graph.yaml` for
catalog-wide edges). One or two file reads usually settle both the leaf
and the wiring direction.

When no catalog pack is reachable, or to drill one leaf's exact contract:

1. Run `planton explain <ProducerKind> -o json`.
2. Read the `outputs` array — each entry's `name` is the leaf under
   `status.outputs.`.
3. When a charts fleet is reachable inside your filesystem boundary (a
   checkout or workspace you were given carries a `charts/` tree),
   cross-check by grepping its `valueFrom` blocks for that kind. Never go
   looking for a fleet beyond the boundary — the explain report alone is
   sufficient.

Build errors for bad references are explicit: `Invalid valueFrom references:
Field 'no_such_output' not found in …StackOutputs for kind: …` — fix the
`fieldPath` leaf to match the schema report, not the provider's API docs.

## References cross chart boundaries — look before you expose a param

Deploy-time resolution is a lookup over the organization and environment's
whole deployed estate — kind + resource name (+ environment), never "is the
producer in this chart". A resource in one chart can therefore reference a
VPC, a cluster, a gateway, or a zone that a DIFFERENT chart deploys — in
this workspace, or one deployed weeks ago.

**The rule: a value some other resource produces is wired by reference,
never collected from the user.** Exposing a param whose value the user must
copy out of another chart or the console ("paste the gateway id here") makes
the user do the platform's job — it is the failure class this section
exists to prevent. Before ANY param whose value is an id, ARN, endpoint, or
name that infrastructure produces, run this check in order:

1. **This chart** — the producer is here: plain `valueFrom`, same as always.
2. **The workspace's other charts** — the producer lives in a sibling chart
   of this workspace: write the same `valueFrom`, with `name` matching the
   producer's `metadata.name` expression (see the naming nuance below).
3. **The org's existing estate** — the producer was deployed by an earlier
   chart or by hand: ground it with the CLI (`planton search
   by-resource-kind <Kind>`, `planton get <kind> <name>`, `planton infra
   project list`) and reference the real deployed name.
4. **Only when all three come up empty** is a param honest — and even then,
   prefer a param that names the RESOURCE (`vpc_name`) feeding a `valueFrom`
   expression over a param that carries a raw id the user must go find.

### Cross-chart mechanics (what changes, what does not)

- **The block is identical** — `kind`, `name`, `fieldPath`; nothing marks a
  reference as cross-chart. The DAG treats an out-of-chart producer as an
  external reference node: it deploys nothing and creates no ordering edge.
- **Names must match the DEPLOYED name.** When both charts deploy into the
  same environment and both write env-prefixed names, the SAME expression on
  both sides renders identically — the shared chart's
  `"{{ values.env }}-cluster"` is exactly what the app chart's reference
  renders to. Reference names are slug-normalized the same way resource
  names are (a zone named `example.com` is matched as `example-com`), so
  reference the name as authored and let the platform normalize.
- **`env` reaches across environments.** A reference resolves in the
  deploying environment by default; set `env` explicitly to consume a
  producer that lives in another environment (a shared cluster in `shared`
  consumed by an app in `dev`).
- **Ordering is YOUR duty across charts.** Inside one chart the reference
  creates the deploy-order edge; across charts there is no edge — the
  producer must already be deployed when the consumer deploys. Compose
  producers-first, finish and deploy the shared chart before the app chart,
  and say the order out loud when handing off.
- **Only `string | valueFrom` fields carry references** (the explain report
  and the component's reference page mark them). A plain string field cannot
  hold a reference — for those, ground the literal with the CLI rather than
  asking the user.
- **The build validates the reference's SHAPE, not its target's existence**:
  kind known, field path real on that kind. A green build does not prove the
  producer is deployed — the deploy resolves it, so name the assumption when
  the producer lives outside the workspace.

### Worked example — the shared gateway

A platform chart deploys the shared cluster and its gateway; an app chart
routes `api.example.com` through that gateway. The app chart never asks the
user for the gateway's identity — its route references it:

```yaml
# app chart, templates/route.yaml — the Gateway lives in the PLATFORM chart
spec:
  parentRefs:
    - name:
        valueFrom:
          kind: KubernetesGateway
          name: "{{ values.env }}-gateway"   # the platform chart's exact name expression
          fieldPath: status.outputs.gateway_name
```

The same shape wires an app's security group to a shared VPC
(`kind: AwsVpc … fieldPath: status.outputs.vpc_id`), a workload's DNS to a
shared zone, an app database to a shared subnet group. If you catch yourself
writing a param called `vpc_id`, `cluster_arn`, or `gateway_host` — stop;
that is a reference.

## DAG semantics

The platform builds a directed acyclic graph from two edge sources — every
`valueFrom` link AND every `metadata.relationships` entry — and deploys
resources in topological order.

Implications for chart authors:

- A producer inside the chart gets a real ordering edge. A producer OUTSIDE
  the chart renders as an external reference node — no manifest, no edge —
  and resolves at deploy time from the org's deployed estate, so it must
  already exist by then (see "References cross chart boundaries" above).
- Circular references fail at build time.
- Resources with no incoming references deploy in parallel in the earliest
  layer.

## metadata.relationships — real ordering edges

Relationships create REAL dependency edges: a resource with a `runs_on` or
`depends_on` relationship does not start until its target has deployed
successfully.

```yaml
metadata:
  name: "{{ values.env }}-istio"
  relationships:
    - kind: AwsEksCluster            # PascalCase kind of the target
      name: "{{ values.env }}-cluster"  # target's exact metadata.name
      type: runs_on                  # runs_on | depends_on | uses | managed_by
```

The `name` must match the target's `metadata.name` exactly (same template
expression on both sides). Choosing between the two mechanisms:

- **`valueFrom`** when a spec field needs a value the producer creates — it
  carries the data AND the edge.
- **`relationships`** when the dependency is real but no spec field carries a
  value — the canonical case is Kubernetes workloads that must wait for their
  cluster (see `kubernetes-on-cluster.md`), or an operator that must install
  before the instances it serves.

A relationship never substitutes for `valueFrom` when a spec field needs the
actual value.

## metadata.group — layout hint

`metadata.group` (e.g. `network`, `compute`, `kubernetes`) groups resources
in the UI and in build reports. It does not affect deploy order. The fleet
uses consistent group names per concern — follow the same convention in new
charts.

## Common wiring mistakes

| Mistake | Fix |
|---------|-----|
| Hardcoded AWS account ID or VPC ID | valueFrom to the creating resource |
| A param collecting another resource's output (`vpc_id`, `gateway_host`) | The references-before-params check above — wire it, in-chart or cross-chart |
| Mismatched name between producer and consumer | Same template expression on both sides |
| Wrong output leaf (`vpc_id` vs `vpcId`) | Schema report (fleet grep only when a fleet is in your boundary) |
| Referencing a conditionally omitted resource | Wrap the consumer in the same `{% if %}` or ensure the producer always renders |
