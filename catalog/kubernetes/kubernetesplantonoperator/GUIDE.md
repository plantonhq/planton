# KubernetesPlantonOperator Guide

The judgment this guide carries: one Planton operator serves every
Planton platform on the cluster — the manager is a cluster-owner install,
the platforms are per-team declarations, and the coupling that actually
bites is DESTROY ORDERING, not installation.

## Once per cluster, platforms by the many

Install the operator once; teams declare KubernetesPlantonPlatform
resources — each platform gets its own namespace, its own URL, its own
identity server and databases, all reconciled by this one manager (it
watches every namespace). The operator polices its own singleton-ness at
startup: a second install refuses to start and says why, so the failure
mode of a mistake is a crash-looping pod with a remedy in its log, never
two managers fighting.

Two honest shared facts ride the topology: the cluster has ONE
`plantonplatforms.planton.ai` CRD schema (so one operator version defines
the spec surface for every platform, while each platform still pins its
own application version), and Tekton's cluster-wide build-events sink
means builds can feed only one build-enabled platform per cluster.

## The chart owns its definitions

The `PlantonPlatform` and `PlantonIdentityProvider` CRDs are ordinary
resources of the release, rendered from the operator's own source behind
two chart values the spec's `crds` dials map onto: `install` (the
definitions ship with the release; false only when another owner already
has them on the cluster, and with them absent the operator cannot start)
and `keep_on_uninstall` (Helm stamps the definitions to survive an
uninstall; false deletes them and every platform behind them). Because the
definitions travel with the chart, `chart_version` is the one lever:
upgrading it upgrades the operator and its schema together, and the
modules carry no copy of the schema anywhere. Kept definitions carry the
release's identity, and the release name is fixed, so every later install
of the operator on the cluster adopts them.

## Destroy is safe by construction

The operator's own destroy strands nothing: the definitions survive (kept
by default), declarations survive, running platforms keep serving —
merely unmanaged until the next install adopts them. Even platform
DELETION completes without the operator, because platform teardown is
Kubernetes garbage collection of the declaration's owner-referenced
objects, not operator work. Prefer platforms-then-operator as hygiene
(the operator publishes status while platforms drain), never as a
requirement.

## On the diagram

The operator draws one explicit `depends_on` edge FROM each
KubernetesPlantonPlatform — no spec field consumes an operator output
(the coupling is the cluster-global schema contract), so composed charts
declare the edge in metadata. An Infra Chart that installs Planton
carries namespace + operator + platform as one deployable arm.
