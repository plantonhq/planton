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
  internet create a Viewer. `hosted_domain` only narrows Google's account
  chooser; it is a hint, not a gate.
- **Roles come from a JMESPath.** `role_attribute_path` maps claims to
  Admin, Editor or Viewer (`email == 'lead@example.com' && 'Admin' ||
  'Viewer'`) and is re-read at every sign-in.
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
- **Keep the login form as break-glass.** Leave `disable_login_form`
  off and `auto_login` off until the identity provider has proven itself:
  the admin account is the way back in when it is down.

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
