# KubernetesGrafana Guide

The judgment this guide carries: standalone Grafana earns its node when
it reads MORE than one source — and its default state is a trap: every
hand-made dashboard lives in an ephemeral local database that vanishes on
pod restart unless persistence is declared.

## Standalone hub vs the stack's bundled Grafana

One kube-prometheus-stack and nothing else to look at? Its bundled
Grafana (on by default there) is the simpler path — skip this kind. The
moment dashboards must read Loki, Tempo, ClickHouse, Postgres, or a
second Prometheus, THIS kind is the composition hub: declare one
standalone Grafana with a `datasources` entry per source, each `url`
wired by `valueFrom` to the source's exported endpoint — the full wired
example lives in the
[observability-stack pattern](../../_patterns/observability-stack.md).
Never run both for the same audience.

## Declare state, or lose it

UI-authored dashboards, users, and preferences live in an embedded
SQLite on local disk, and the chart's default is EPHEMERAL — a pod
restart erases everything hand-made (the reference page states it).
Either declare persistence/database (the spec's own arms), or treat
dashboards as code: ship them as ConfigMaps labeled
`grafana_dashboard: "1"` (the sidecar loads them) or import them with
`community_dashboards`, present from first boot, immune to restarts. For anything
beyond a scratch environment, one of the two is part of the proposal.

## Dashboards as code

The dashboard sidecar (on by default) loads every ConfigMap labeled
`grafana_dashboard` from every namespace within a minute, so a dashboard
ships as a typed `KubernetesConfigMap` (its preset `04-grafana-dashboard`)
beside whatever it watches, never by editing this resource. What a live
install teaches:

- **Provisioned means read-only.** Grafana refuses to save over a
  dashboard loaded from a file ("Cannot save provisioned dashboard"),
  even for an Admin. An Editor can still save a copy, which is drift:
  compare `/api/search?type=dash-db` with the committed uids.
  `meta.provisioned` and `meta.provisionedExternalId` (the ConfigMap's
  data key) on `/api/dashboards/uid/<uid>` tell a hand-made copy from
  one another chart shipped.
- **Pin datasource uids, and name them in every panel.** A dashboard
  that reads `{"uid": "prometheus"}` survives the datasource's URL moving
  to another Prometheus; one that reads a name does not.
- **In an infra chart, keep the chart engine's delimiters out.** Every
  template is rendered, and the engine keeps a raw block's tags in its
  output, so a Prometheus legend format with double braces breaks the
  render. Write the JSON pretty-printed (closing braces then never sit
  side by side) and name series with a field override,
  `displayName: "${__field.labels.<label>}"`. Generating the JSON from
  short sources makes both rules a refusal rather than a review comment.
- **Match the running Grafana's `schemaVersion`** (42 on Grafana 13.1),
  so it is not migrated in the browser on every load. The stored copy
  then differs from the file only by `id` and `version`, which makes
  "identical to the committed file" a check.
- **A blank panel must mean broken.** Grafana answers a query with no
  data as status 200 with an empty frame, so a checker reads frames, not
  status, and a panel is written to return data on a healthy system:
  counts end in `or vector(0)`, lists are sorted rather than filtered.
- **One file for one cluster and many.** A `$cluster` variable whose
  "All" is `.*` also matches a series with no `cluster` label. Grafana's
  query API does not fill dashboard variables (it fills `$__range` and
  `$__rate_interval`), so a checker substitutes them itself.
- **Removing a dashboard is a purge.** An infra chart re-install never
  deletes a ConfigMap the chart stopped declaring; purge it by name, or
  the drift comparison above names it as shipped by a chart.

## Credentials

The chart generates the admin password once, into the `<name>` Secret —
consume by reference; or point `adminSecret` at a Secret you manage
(e.g. a KubernetesExternalSecret projection). Never inline.

## Who can open Grafana

Grafana is usually the one screen that shows every system at once, so
the question "who can sign in" belongs in the proposal, not after it.

- **Sign-in is typed.** `auth.google` (a Google Workspace) or
  `auth.generic_oauth` (Okta, Microsoft Entra ID, Keycloak, Auth0, ...)
  with the client secret as a `$secret/` reference. The modules write it
  into their own `<name>-sso` Secret; it never reaches grafana.ini or
  `helm_values`. Both need `server.root_url` — the provider sends people
  back to `<root_url>/login/google` (or `/login/generic_oauth`), which is
  also the redirect URI to register on the OAuth client.
- **Google admits whoever its consent screen admits.** An Internal
  consent screen admits only the Workspace that owns the Google project;
  an External one admits any Google account (in Testing, only listed
  test users). Grafana's own gate is `allowed_domains`, matched against
  the signed-in email; the kind refuses a Google sign-in that allows
  sign-up with no `allowed_domains`, because that would let anyone on the
  internet create a Viewer. `hosted_domain` sends Google's `hd` hint, and
  Google's sign-in screen then accepts only that domain's accounts (seen
  live: an outside account never reaches Grafana). Keep `allowed_domains`
  anyway; it is the gate Grafana itself enforces.
- **Roles come from a JMESPath.** `role_attribute_path` maps claims to
  Admin, Editor or Viewer (`email == 'lead@example.com' && 'Admin' ||
  'Viewer'`) and is re-read at every sign-in.
- **Reading logs and traces needs Editor in open-source Grafana.** Explore
  is the only place to search logs or open a trace until dashboards exist,
  and it is granted to Editors and Admins. The Viewer workaround
  (`viewers_can_edit`) is deprecated, and the finer "data sources
  explorer" role is assigned only in Grafana Enterprise. Staff who must
  investigate are Editors; keep dashboards in committed files and treat a
  saved hand-made one as drift (see "Dashboards as code"). Agents querying
  through the API need no more than Viewer.
- **With `hosted_domain`, Google's sign-in screen fixes the domain.** The
  email box carries `@<domain>` and an outside account cannot be typed
  in, so a browser test of the outside-account refusal stops at Google;
  `allowed_domains` remains Grafana's own gate behind it.
- **The manifest owns sign-in.** Once either provider is declared,
  Grafana's Administration > Authentication screen can no longer edit
  any OAuth provider: settings saved there would otherwise live in
  Grafana's database and silently override the manifest. (Grafana always
  leaves LDAP editable there, and it skips an empty
  `configurable_providers`, which is why the modules set a list naming no
  provider rather than an empty one.)
- **A rotation takes effect on apply.** Grafana reads the client secret
  only at start; the pods carry a checksum of the `<name>-sso` Secret, so
  the apply that writes a new secret also rolls Grafana onto it.
- **Keep the login form as break-glass until sign-in is proven.** Leave
  `disable_login_form` and `auto_login` off until the first person has
  signed in through the provider. After that, turning both on sends
  everyone straight to the provider; the way back in when the provider is
  down is the admin account from `admin_secret_name` over a port-forward
  (the HTTP API with basic auth), or a re-apply with the form on.

## Trace to logs, and back

The kind has no typed field for Grafana's trace links; they live in each
datasource's `json_data`, and each side names the other by `uid`, so pin
the uids. OTLP logs keep `trace_id` as structured metadata, which a
`label` matcher reads:

```yaml
datasources:
  - name: Loki
    type: loki
    uid: loki
    json_data: |
      derivedFields:
        - name: TraceID
          matcherType: label
          matcherRegex: trace_id
          datasourceUid: tempo
          url: "${__value.raw}"
  - name: Tempo
    type: tempo
    uid: tempo
    json_data: |
      tracesToLogsV2:
        datasourceUid: loki
        spanStartTimeShift: -5m
        spanEndTimeShift: 5m
        customQuery: true
        query: '{service_name=~".+"} | trace_id="${__trace.traceId}"'
```

Prove it with one synthetic span and one log record carrying the same
`traceId` (OTLP HTTP to Tempo's `otlp_http_endpoint` and to Loki's
`otlp_push_endpoint`): Tempo returns the trace by id, and the query above
returns the line. Loki indexes OTLP's `service.name` as `service_name`,
which is what the query above matches on.

## On the diagram

Grafana renders as the hub with a labeled datasource edge into every
source — the observability topology is readable from the graph alone.
Provisioned dashboards live inside the node; hand-made ones are invisible
AND ephemeral, a double reason to provision.

## Pairs well with

- KubernetesKubePrometheusStack / KubernetesLoki / KubernetesTempo — the
  three standard datasources (pattern above).
- KubernetesExternalSecret — managed admin credentials.
- KubernetesIngress / route kinds — exposing the UI, composed as always.
