# Observability on Kubernetes

A person asking to "set up monitoring" on a cluster is asking one question:
will someone know when this breaks, before a user says so? This reference is
the craft for answering it with the catalog's assembled stack. Kind
facts (every field, default and validation) live in the catalog pack, on
`KubernetesKubePrometheusStack`'s reference page and guide and in the
observability-stack pattern; read them there, never from memory.
`cloud.kubernetes-architecture.md` covers what else runs on the cluster, and
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
   - **What the cluster installed before the stack is watched from the
     stack's side.** The gateway, istiod, cert-manager, external-dns and
     the database operator come with the cluster, before the monitor CRDs
     exist, so their own `serviceMonitor` switches stay off (on a fresh
     cluster they fail the install). Declare one class monitor per kind of
     component beside the stack instead: a `KubernetesPodMonitor` or
     `KubernetesServiceMonitor` with `namespace_selector: {any: true}` and
     a selector every instance carries (the PodMonitor's `istio-gateways`
     preset; `cnpg.io/podRole: instance` for every CloudNativePG
     instance). Components installed after the stack (an environment's
     vault, caches, workflow engine) turn their own switches on.
   - **Every namespace with a NetworkPolicy admits the stack's Prometheus**
     on the metrics ports, in one peer (namespace
     `kubernetes.io/metadata.name: <stack namespace>` with pods
     `app.kubernetes.io/name: prometheus`), applied before the monitors. A
     fenced target reads down with a dial timeout and posts `TargetDown`
     ten minutes later.
   - **Read the serving code for the port.** A Planton runner and runner
     tunnel (embedded konnectivity) serve `/metrics` on their health port
     (8093); their admin port listens on the pod's loopback only. Neo4j
     community serves no metrics at all.
   - **Read what a new component's series carry, at its first scrape.**
     A series that already has `cluster` keeps it over the stack's
     external label (OpenBao's own cluster id: drop it with a
     `labeldrop`); a component's own monitor switch keeps every series
     (Temporal's per-task-queue histograms, about 150,000 series per idle
     environment: declare the monitor with a keep list instead); a metric
     naming its subject's namespace needs `honor_labels` (cert-manager's
     certificates). Compare `scrape_samples_post_metric_relabeling` with
     the hub's storage budget before you leave it running.
   - **On GKE, turn the managed collection off** on the cluster:
     `monitoring.managed_prometheus_enabled: false` with
     `monitoring.components: [SYSTEM_COMPONENTS]` in one update (an empty
     list keeps the billed packages).
2. **Alert delivery in the same change**: `alertmanager.notifications`, not
   a follow-up. Out of the box Alertmanager notifies nobody.
3. **The outside heartbeat**: `notifications.heartbeat` to a monitor that
   runs outside every cluster and pages when the heartbeat stops. Apply
   the stack and see its first heartbeat arrive, then tell the monitor to
   expect the cluster; the other order opens with a false "cluster silent"
   alert.
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
- Give Prometheus, Alertmanager and the operator a `KubernetesPriorityClass`
  just below the workloads (value -1, preemption `never`) through their
  `scheduling` blocks, so monitoring is evicted first and evicts nothing.
  Never go below -10: the cluster autoscaler treats those pods as
  expendable and will not add a node for them.
- Ask where the cluster runs. On a managed control plane (GKE, EKS, AKS)
  turn off the controller-manager, scheduler, etcd and kube-proxy
  scrapers together with their rule groups (the kind's guide names them),
  and on GKE the CoreDNS scraper too, because GKE runs kube-dns. Left on,
  each is a target that is down forever and an alert that never clears,
  which teaches the person to ignore the channel on day one.
- Ask whether the node pools autoscale and whether the cluster runs CI
  builds or batch machines. On autoscaled pools put `KubeCPUOvercommit`
  and `KubeMemoryOvercommit` in `default_rules.disabled_alerts`: they
  count today's nodes and fire on every burst. A curated alert that
  misreads work done on purpose (Tekton build pods read not-ready once a
  step ends; a build machine sits at full CPU) is replaced, not muted:
  disable it and declare the same alert name in a `KubernetesPrometheusRule`
  that leaves the work out (`kube_pod_owner{owner_kind!~"Job|TaskRun"}`,
  and every pod on a build machine, whose pods all go not ready when it
  goes dark; `unless` the build taint in `kube_node_spec_taint`), keeping upstream's
  `for`, severity and `namespace` label (Alertmanager's info inhibition
  matches on it). Then the work's real failure needs its own alert: a
  build machine that runs out of memory goes dark and is replaced before
  `KubeNodeNotReady`'s 15 minutes or the node-memory alert's 15 minutes
  pass, so nothing fires for it unless you add a short-hold rule on the
  build machines' free memory. `alert_overrides` changes a kept alert's
  `for` or `severity` instead. Read `/api/v1/rules?type=alert` after the
  change: a name that matches no rule is skipped silently.
- Pushover's emergency priority repeats every minute until someone
  acknowledges it in the app, even after the alert resolves. Say so before
  the person's phone starts ringing.
- The person's own alert and recording rules are `KubernetesPrometheusRule`
  objects beside each cluster's agent stack, one object per owner, never
  rules pasted into the stack's `helm_values`. Each alerting rule carries
  `severity` and `component` labels and a `runbook_url` annotation, with
  static annotation text (a `{{ $labels.x }}` can carry a customer's name
  into a message, and a templating chart engine mangles it). A rule
  object loads into every stack on the default `all_monitors` discovery;
  a stack on `release_managed_only` (a hub receiver) loads it only when
  the rule's own `labels` carry `release: <that stack's release_name>`.
  `labels` on the rule object are the object's, not the alert's: a
  `severity` there reaches no alert.
- After applying a rule object, prove it evaluates: the Prometheus rules
  API (`/api/v1/rules`) lists its group, and `/api/v1/alerts` or a query
  for a recorded series answers. One malformed rule makes the operator drop
  the whole object while the apply still succeeds, so a green deploy is not
  the proof.
- What the agent scrapes beyond the cluster itself is a
  `KubernetesServiceMonitor` (the workload's Service names its metrics
  port) or a `KubernetesPodMonitor` (pods no Service exposes: a database
  operator's instances, a DaemonSet's exporters, or replicas that must be
  seen while unready). Put it beside the workload. Its `selector` matches
  the Service's or the pods' labels, so read them off the cluster before
  writing it. Point every credential (a token, basic auth, a CA) at a
  `KubernetesSecret` or `KubernetesConfigMap` by reference, in the
  monitor's namespace. Set `job_label` to a label whose value is the same
  in every environment, and a `sample_limit`.
- After applying a monitor, prove it scrapes: the Prometheus targets API
  (`/api/v1/targets`) lists a target of the job
  `serviceMonitor/<namespace>/<name>/<n>` (or `podMonitor/...`) with health
  `up`. The operator skips a whole monitor whose Secret is missing or whose
  endpoint breaks one of its rules, and nothing but its own log says so,
  so here too a green deploy is not the proof. A `target_port` written as
  a number reaches the object as a number; a port name must be declared on
  the Service or the pod, or no target appears.
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
- **Ask what teammates must do in Grafana.** In open-source Grafana only
  Editors and Admins can open Explore, the one place to read logs and
  traces before dashboards exist. People who investigate are Editors;
  tell the person the cost (an Editor can save a hand-made dashboard) and
  keep dashboards in committed files.
- **Put the log collector in each cluster's agent, the stores in the
  hub,** and give the hub its own listener set on the cluster's Gateway
  (the pattern's "A hub beside a cluster's agent"), with external-dns's
  `gateway_listener_sets` on, or the hub's route is Accepted and its
  hostname never resolves. Copy the collector preset whole: its
  `include_file_path`, `file_storage` and self-exclude are each the
  difference between logs arriving and silence, and its sending queue
  (batching there and never in a `batch` processor, capped in bytes
  below Loki's burst, blocking when full) is the difference between a
  restart or a long Loki outage losing lines and losing none. Turn on
  the collector's `service_monitor_enabled` so its queue is watched. Keep
  its toleration of every `NoSchedule` taint: a cluster with a tainted
  pool (builds, GPUs) otherwise never ships those nodes' logs, and the
  daemonset still reads complete. To see coverage, compare the cluster's
  nodes (`kube_node_info`) with the collector's ready pods; the
  daemonset's own desired count never includes a node it can't tolerate.
- **A `$var/` reference works in a plain-string field** (a Grafana
  `client_id`, for one): it resolves at deploy like any other.
- **Dashboards are files, not clicks.** Ship each as a
  `KubernetesConfigMap` labeled `grafana_dashboard` (the kind's preset
  `04-grafana-dashboard`), in its own composition beside the hub so a
  panel change never re-plans Grafana. Ask the person which questions
  they ask mid-incident and give each its own dashboard, titled by the
  question, with each panel's description the question it answers;
  offer that before any generic community dashboard. Grafana refuses to
  save over a provisioned dashboard, so tell the person screens change
  only through the files. In an Infra Chart, keep double braces out of
  the dashboard JSON (the chart engine renders it): pretty-print it and
  name series with `${__field.labels.<label>}` display names.
- **When several clusters report to one hub, answer per cluster.** Join
  and group by `cluster` everywhere (node addresses and pod names repeat
  across clusters, and a join on them alone fails), name clusters by a
  short label, and lead a capacity screen with one row per cluster:
  reserved, used now and the busiest node's peak in the range for
  memory and CPU, memory used a week ahead at the range's trend, the
  fullest disk, OOM kills. Name every column by its noun and window.
  Withhold the forecast until the range holds three days of history,
  and say so in the cell ("Needs 3 Days"): a line through a few hours
  projected a week out is noise.
- **Show alert notifications as a table per channel,** sent and failed
  totals over the selected range, never a chart of per-interval counts:
  a heartbeat on a two-minute rhythm aliases to zero or a saw-tooth in
  short windows. List every channel the estate uses, so a pager that
  sent nothing reads zero.
- **Give the operator a first screen that belongs to no cluster:**
  each environment's console and API as the outside watcher reaches
  them, how long both were up, workloads with nothing ready, every
  firing alert with its environment (a cluster-wide alert keeps its
  cluster's own environment label), and each cluster's heartbeat age;
  up and down over time as a state timeline. Scrape the watcher with
  `honor_labels: true` so its probes keep the environment they probed.
- **A data-tier screen is one row per component per environment:**
  workloads down (nothing ready, so a component scaled to zero reads
  red), restarts and out-of-memory kills in the range, fullest volume,
  and error lines in the last hour from Loki joined into the same row.
  Match each component's real log format for errors (a JSON level,
  Postgres's `error_severity`, plain `panic:`), never the bare word
  "error", which Postgres prints in every record. A cell nothing
  measured reads a dash, never a zero nobody counted; an environment no
  agent reports on yet gets one row that says so.
- **Read "how full is the node" from the node exporter,** not the
  containers' working set: the kubelet stops reporting container memory
  first when a node starves. Put it beside what pods reserve, because
  the autoscaler sees only reservations: a node at 99% with 45%
  reserved is a build or a workload with no honest memory request.

## When a second cluster reports to the hub

Each cluster keeps its own agent and Alertmanager; only copies travel.
The pattern's "Several clusters, one hub" has the exact switches. What
to decide with the person, and what to watch for:

- **The hub gets its own receiving Prometheus,** a second stack with
  everything the agent already runs turned off. Never make the hub
  cluster's agent the receiver: its rules would run over the other
  clusters' samples and every one of their alerts would post twice.
- **The other clusters write through a door of their own: a dedicated
  Gateway, so a second load balancer.** Tell the person it costs about
  one forwarding rule a month, and why it can't share the platform's
  front door: a JWT check there would refuse every other bearer token
  that Gateway serves. Put the door's pods at the monitoring priority
  through the Gateway's `parameters_ref` ConfigMap.
- **Mint one token per cluster yourself,** RS256 with a fresh key: the
  token into the person's vault, never onto a screen, and only the public
  key into the door's inline key set, with the cluster's name as key id
  and subject. Destroy the signing key; revocation is deleting the key
  from the door. Ask the person to name the issuer (the door's hostname
  reads best) before the first token, because every token carries it.
- **Label every signal with where it came from.** Metrics already carry
  each agent's `external_labels` (`cluster`, `environment`) on the
  remote write. Logs need the collector to insert `k8s.cluster.name` and
  `deployment.environment.name`, which Loki indexes by default. Where
  several environments share one cluster, the namespace tells them apart
  until each pod states its own environment; say so.
- **Give the receiver an out-of-order window** about as long as a sender
  can resend (two hours), or an outage of the door longer than about an
  hour leaves a gap at the hub, while each cluster still keeps its own.
- **Size the hub's Loki for every cluster catching up at once.** Raise
  `limits.ingestion_rate_mb` (12 serves a few clusters) and keep
  `ingestion_burst_size_mb` above the collectors' batch cap (24 against
  4 MiB): Loki refuses a push larger than its burst every time, and a
  collector retrying forever then stalls that node's logs for good.
- **Add a cluster to a running hub in order:**
  1. Mint its token.
  2. Add its key and principal to the door and re-apply the hub.
  3. Write the token where its agent reads it.
  4. Install the agent.
  5. Only after the agent's first heartbeat, add the cluster to whatever
     watches heartbeats.

  Take a cluster away (or rebuild it) in reverse: out of the heartbeat
  list, then its agent, then the cluster. Write that order into the
  cluster's own rebuild runbook, so a rebuild brings monitoring back
  instead of dropping it. A rebuilt cluster with the same name keeps
  its token.
- **Expect real alerts in the first hour** of a cluster that never had
  in-cluster alerting. Read them with the person and list their causes;
  never silence one by hand.

## The alerts that page, and the ones that post

Write the user's page-class rules once the components they read are
scraped (`KubernetesPrometheusRule`, beside each cluster's agent; the
pattern's alert section is the reference). The ones a platform owes its
customers, and how each misfires if written naively:

- **A front door burning its error budget** (14.4x over 1h and 5m, 6x
  over 6h and 30m of a 99.5% target). Record the gateway's traffic per
  window once and alert on the recorded ratio; give the long window a
  floor of failed requests, or two failures at night page at low traffic;
  say that a gRPC call failing inside HTTP 200 is not counted. If the
  user has a public status page that colours a part from an alert's
  `component`, naming a burn rule after that part is a decision about
  what customers read: ask, never assume.
- **A database without a recent backup or archive.** Hold a database to a
  backup only once it is older than the threshold (its volume's creation
  time), or every new database pages before its first nightly; require
  WAL segments waiting before calling an old archive stale.
- **A certificate within seven days.** cert-manager renews 30 days
  ahead, so also post "renewal overdue" to the channel three weeks
  earlier.
- **Channel alerts** for what fails quietly: the outside prober not
  running (scraped by one agent, read without `absent()`), a Ready node
  with no log collector (per node, by uid), a sealed vault (read from its
  StatefulSet's ready pods: a sealed vault's own metrics vanish), a
  workflow backlog aging, a runner tunnel holding no runner, and metrics
  not reaching the hub.

Each rule sets `environment` from its namespace when one cluster serves
several environments (external labels never reach a rule's result),
carries a runbook whose first line is a command, and has a promtool test.
Prove each with a deliberately fired alert: a real failure where one can
be declared safely (a podless Service behind one route answers 503 for a
burn), a synthetic alert with the production labels on production's own
Alertmanager otherwise, and `amtool config routes test` for every name
the pager route lists.

Silences are part of the design, not a click: one drops the alert before
routing (the status page's webhook included), and one matching only an
environment silences the heartbeat too. Give the user a command that
silences one exact alert for at most a day, tied to a record.

## Proving it

Do these with the person, and report what arrived and when:

0. Minutes after install, confirm every active scrape target reads `up`
   (Prometheus's targets page or API) and every monitor has at least one
   target (`/api/v1/scrape_pools` against the active targets: a monitor
   whose selector matches nothing has no target and no error anywhere).
   A target that is down now is a wrong scraper posture or a fence, not an
   incident; a target that reads `unknown` was found since the last scrape,
   so read again after one interval.
1. Fire a channel alert from inside the Alertmanager pod:
   `amtool alert add alertname=Drill severity=warning environment=<env> --alertmanager.url=http://localhost:9093`.
   It must arrive in the channel, titled with the environment. Double-quote
   any value with spaces inside its argument
   (`--annotation='summary="Drill: no action needed."'`). Alertmanager's
   `alertmanager_notifications_total` and `_failed_total` for the
   integration prove delivery even when the channel can't be read back;
   the person confirms what the post says.
2. Test the paging route without paging anyone:
   `amtool config routes test --config.file=/etc/alertmanager/config_out/alertmanager.env.yaml severity=page environment=<env>`
   must name the pager receiver. Pass each label as its own argument: in
   zsh, a label list held in one variable arrives as one malformed label
   and every test falsely names the root receiver.
3. With the person's go, fire one `severity=page` alert and confirm the
   phone rang. Resolve it (`amtool alert add ... --end=<now>`) before they
   acknowledge: it keeps ringing, and the resolved notice arrives one
   `group_interval` (5 minutes by default) later. That is the behaviour to
   warn them about.
4. With the person's go, scale the Alertmanager resource to zero replicas
   and confirm the outside monitor pages within its window (a 5-minute
   window pages about five and a half minutes after the stop); restore the
   declared replica count and confirm the heartbeat recovers within a
   minute. A drill never breaks a real service to test the pager.

5. For the hub, sign in with an account the person expects to get in and
   one that must not (another domain); the first lands with the role the
   manifest names, the second is refused. With `hosted_domain` set, Google's
   own screen fixes the domain, so the outside account stops there; say so
   rather than claiming Grafana refused it. Unauthenticated, `/` must
   redirect to sign-in and `/api/datasources` answer 401.
6. Prove every datasource from the server: each one's
   `/api/datasources/uid/<uid>/health` reads OK, a query returns data (a
   log line from a known namespace, `up` from Prometheus), and one
   synthetic span plus one log record with the same `traceId` come back
   from Tempo by id and from Loki by the trace-to-logs query. Objects
   appear in the buckets once Loki and Tempo flush (minutes for Tempo,
   longer for Loki's chunks); an index file in the logs bucket proves the
   key writes.
7. Prove every dashboard from the server: `/api/dashboards/uid/<uid>`
   reads `meta.provisioned: true` and matches the committed JSON except
   `id` and `version`; every panel query returns at least one frame over
   the dashboard's default range through `/api/ds/query` (a 200 with an
   empty frame is no data; fill `$cluster`-style variables yourself, the
   API fills only `$__range` and `$__rate_interval`); and
   `/api/search?type=dash-db` lists nothing the files do not declare.
8. For a telemetry door, four requests: no token is refused (403), a
   token with the right claims signed by a key the door never saw is
   refused (401), the real token with an empty body gets the store's own
   error (Prometheus's 400, Loki's 422), which proves the door let it
   pass, and the real token on any other path or method is refused (403).
   Then confirm the sending cluster's series and log lines at the hub by
   their cluster label, and with the person's go stop the receiver for ten
   minutes and confirm no gap in the sender's series after it returns.
9. For the log path, two drills. Load: post gzipped OTLP log batches
   through the door at the summed catch-up rate for three minutes, retry
   an unanswered batch once as a collector would, and confirm every batch
   was accepted, `loki_discarded_samples_total` did not move, and the
   platform's own front page is no slower. Durability, with the person's
   go: a test pod writes numbered lines, the hub's Loki is scaled to zero
   for ten minutes and the writer's node's collector is deleted halfway,
   and after Loki returns every number must come back from Loki (a
   duplicate is fine; the queue delivers at least once). Ask Loki for at
   most 5,000 lines per query, its default per-query limit.

A rotated alerting secret is picked up on the next notification without a
restart, so rotation needs no drill of its own; the hub's secrets roll the
pods on the apply that writes them.
