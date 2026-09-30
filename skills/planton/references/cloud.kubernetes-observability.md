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
   assembled wiring) is where signals are read. It never pages. Its logs
   and traces live in a bucket one environment above it, and its sign-in
   is declared, both before the first person opens it.

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

## The hub: where its data lives, who can open it

- **Logs and traces outlive the hub.** Put them in a bucket declared one
  environment above the hub (the pattern's "outside the cluster" section),
  so rebuilding or destroying the hub never destroys the evidence. On
  Cloudflare, use the `r2` arm: the bucket by reference, the key pair of a
  token scoped to that one bucket from secrets. Choose the bucket's
  location hint where the cluster runs, because R2 honours it only at
  creation; leave `jurisdiction` unset unless the person has a legal
  residency need, since it changes the bucket's address. Give the bucket a
  lifecycle expiry longer than Loki's and Tempo's retention, never shorter.
- **Ask who may open Grafana before composing it.** Usually "everyone at
  the company, nobody else". For a Google Workspace, the strongest answer
  is an OAuth client whose consent screen is Internal, which is only
  possible in a Google project inside the Workspace's own organization:
  Google then refuses every outside account itself. An External screen
  admits any Google account (in Testing, only listed test users), and
  `allowed_domains` on the kind becomes the only wall. Never compose Google
  sign-in with sign-up on and no `allowed_domains`; the kind refuses it.
- **The manifest owns sign-in.** Declaring `auth.google` or
  `auth.generic_oauth` locks Grafana's own authentication screen, because
  settings saved there live in Grafana's database and silently override
  the manifest. Tell the person that sign-in changes go through the
  manifest from then on.
- **Roles come from `role_attribute_path`,** re-read at every sign-in.
  Grafana ships Google with role sync off, which ignores a role path; the
  kind switches it on whenever one is declared, so the path in the manifest
  is what people get.
- **Rotation takes effect on apply.** Grafana, Loki and Tempo read their
  secrets only at start; the pods carry a checksum of the secret, so the
  apply that writes a new value rolls them onto it. Rotate by updating the
  secret, then re-applying.

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

5. For the hub, sign in with an account the person expects to get in and
   one that must not (another domain); the first lands with the role the
   manifest names, the second is refused. Then read a log line and a trace
   from Grafana and confirm objects are arriving in the bucket.

A rotated alerting secret is picked up on the next notification without a
restart, so rotation needs no drill of its own; the hub's secrets roll the
pods on the apply that writes them.
