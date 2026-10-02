# KubernetesPlantonRunner Guide

The judgment this guide carries: the module wraps the official
`planton-runner` chart instead of re-modeling its resources -- the chart
is the one source of truth for the enrollment mechanics; the runner
token only gates JOINING and is only read at join; and the runner is
the last thing you tear down.

## Why a chart wrap, not raw resources

The enrollment mechanics are load-bearing and easy to get subtly wrong:
replicas pinned to 1 with a Recreate strategy (two live pods under one
runner name would revoke each other's keys), the ephemeral identity
volume the runner persists its minted identity into (container restarts
reuse it; pod recreation re-joins with the token), and the health
endpoints. The official chart already encodes all of it, so the module
installs the chart as a real Helm release -- a runner deployed through
this kind is byte-identical to a hand-installed one -- and confines
itself to the satellites: the optional namespace and the `<name>-token`
Secret. The corollary is the version floor: charts below 0.4.0 silently
ignore the enrollment values (the runner would deploy with no way to
join, and nothing downstream would name the cause), so both engines
refuse them with an explicit error instead.

One resource is ONE runner -- its own enrollment, its own release.
Scaling execution capacity means more resources of this kind, never
more replicas of this one; each renders as its own node on the
architecture diagram, which is exactly how the fleet should read.

## The escape hatch never carries the token

`helmValues` merges over the rendered values with Helm `-f` semantics
(maps deep-merge with your overrides winning, lists replace) -- use it
for chart knobs the spec does not model (nodeSelector, tolerations,
extra env). The one deliberate exception to its last-word contract: the
enrollment block is re-pinned AFTER the merge. Rendered values land in
Helm's release Secret, readable by anyone with release-history read
access, so the chart only ever receives the token Secret's NAME (its
existingSecret form) -- an override that moved the token inline would
break that discipline, and the module does not allow it.

## Rotating the token

The token is only read at join. Applying a new token value updates the
`<name>-token` Secret; running pods keep serving on their minted
identity, and the next pod recreation joins with the new value.
Revoking the old token never touches runners it already admitted --
identities are revoked individually, not through the token.

## The build worker needs Tekton first

`build.enabled: true` registers the runner as a build worker executing
container-image build pipelines through Tekton on this cluster. Tekton
Pipelines must already be installed -- this module does not install it,
and `build.tektonNamespace` only picks where the pipelines run
(defaulting to the runner's own namespace).

## Keep builds on their own nodes

A build is the heaviest thing most clusters run, and a burst of them can starve everything else on the nodes they share. On any cluster that also runs workloads you care about, give builds a node pool of their own -- tainted so nothing else lands there, scaling from zero -- and point `build.scheduling` at it:

```yaml
build:
  enabled: true
  scheduling:
    nodeSelector:
      planton.ai/workload: build
    tolerations:
      - key: planton.ai/workload
        operator: Equal
        value: build
        effect: NoSchedule
```

The runner puts both on every PipelineRun's pod template, so every task pod, and the helper pod Tekton uses to keep a run's pods together, lands only on that pool. Each image build step also declares what it uses, so the pool adds nodes as builds arrive and queues builds beyond its ceiling. Leave `build.scheduling` unset on a single-node cluster, where there is nowhere else to go.

## In-cluster runner vs the hosted fleet

Deploy this kind when the targets are reachable only from inside the
cluster's network -- private services, internal endpoints, the cluster
itself. If every target is reachable by Planton's hosted runner fleet,
an in-cluster appliance buys nothing and adds a workload you operate.

## Teardown is ordered, and the runner goes last

The appliance is standing infrastructure, not a bootstrap step. Destroy
in-cluster workloads THROUGH the runner first; destroy the runner
itself last -- destroying it first strands everything only it can
reach. The module supports this posture by never gating on pod liveness
(no Helm wait): the runner's readiness contract is its work queue
(operations wait there until the worker polls), so a runner whose
control plane is momentarily unreachable still deploys and destroys
cleanly.

## Pairs well with

- KubernetesNamespace -- own the runner's namespace as a first-class
  resource and wire `namespace` through ValueFromRef
  ([namespace-ownership pattern](../../_patterns/namespace-ownership.md)).
