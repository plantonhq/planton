# Observability on Kubernetes

A person asking to "set up monitoring" on a cluster is asking one question:
will someone know when this breaks, before a user says so? This reference is
the craft for answering it with the catalog's assembled stack. Component
facts (every field, default and validation) live in the catalog pack, on
`KubernetesKubePrometheusStack`'s reference page and guide and in the
observability-stack pattern; read them there, never from memory.
`kubernetes-architecture.md` covers what else runs on the cluster, and
`infra.config-references.md` covers the `$secret/` grammar.

## What "monitored" means

A cluster is monitored when a deliberately fired alert has reached a person,
and when stopping the alerting itself has paged someone from outside. Until
both have happened, it is not monitored, whatever dashboards exist. Hold
every proposal to that bar and say so plainly when a plan stops short of it.

## The order

1. **One `KubernetesKubePrometheusStack` per cluster, right after the
   cluster**, in the shared-cluster chart: its CRDs are what every other
   kind's `serviceMonitor` toggle needs. Each cluster keeps its OWN
   Alertmanager, so a cluster pages on its own and a central hub going down
   never silences it.
2. **Alert delivery in the same change**: `alertmanager.notifications`, not
   a follow-up. Out of the box Alertmanager notifies nobody.
3. **The outside heartbeat**: `notifications.heartbeat` to a monitor that
   runs outside every cluster and pages when the heartbeat stops.
4. **Only then the hub**: Grafana with Loki and Tempo (the pattern's
   assembled wiring) is where signals are read. It never pages.

## The questions to ask

Ask these before composing, in the person's words, not the chart's:

- **Who gets woken, and for what?** A page means a person must act now.
  Only those rules carry `severity=page`; everything else posts to a
  channel read in the morning. Ask which environments page at all
  (production usually; preview and development usually never).
- **Where do alerts go?** A Discord channel webhook, a Pushover recipient
  for the phone, or a generic webhook. Each credential is a secret: look
  it up or create it as an organization secret and reference it as
  `$secret/<slug>`; never paste it into a manifest.
- **What watches the watcher?** A dead-man's-switch monitor outside the
  cluster (a hosted heartbeat service or the person's own) and its bearer
  token.

## Composing it

- Put `environment` (and `cluster`) on every alert through
  `prometheus.external_labels`; put `component` on the person's own rules.
  Every message title is `[<environment>] <component>: <alertname>`, so
  these labels are what the person reads at 3 a.m.
- Route `severity=page` to a pager receiver that carries BOTH the phone and
  the channel. The root receiver takes only alerts no child route claims:
  a page route with `continue_matching` and nothing after it never reaches
  the channel.
- Messages never render a namespace, pod or description, because on a
  shared cluster a namespace can name a customer. Tell the person when it
  matters to them: their alert channel can be read by people who must not
  see customer names.
- Give Prometheus and Alertmanager a `KubernetesPriorityClass` just below
  the workloads (value -1, preemption `never`) through their `scheduling`
  blocks, so monitoring is evicted first and evicts nothing. Never go below
  -10: the cluster autoscaler treats those pods as expendable and will not
  add a node for them.
- Pushover's emergency priority repeats every minute until someone
  acknowledges it in the app, even after the alert resolves. Say so before
  the person's phone starts ringing.
- Typed Discord delivery needs the kind's default chart (88 or later); if
  the person pins an older `chart_version`, Discord refuses to load and
  Alertmanager never starts. When Alertmanager is missing, read the
  Alertmanager resource's `Reconciled` condition before anything else.

## Proving it

Do these with the person, and report what arrived and when:

1. Fire a channel alert from inside the Alertmanager pod:
   `amtool alert add alertname=Drill severity=warning environment=<env> --alertmanager.url=http://localhost:9093`.
   It must arrive in the channel, titled with the environment.
2. Test the paging route without paging anyone:
   `amtool config routes test --config.file=/etc/alertmanager/config_out/alertmanager.env.yaml severity=page environment=<env>`
   must name the pager receiver.
3. With the person's go, fire one `severity=page` alert and confirm the
   phone rang; they acknowledge it.
4. With the person's go, scale the Alertmanager resource to zero replicas
   and confirm the outside monitor pages within its window; restore the
   declared replica count and confirm the heartbeat recovers. A drill never
   breaks a real service to test the pager.

A rotated secret is picked up on the next notification without a restart,
so rotation needs no drill of its own.
