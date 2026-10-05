# Kubernetes Keycloak

## When NOT to Use This

**One resource is ONE Keycloak server** — the declaration of the
Keycloak CR (`keycloaks.k8s.keycloak.org/v2beta1`) that the official
Keycloak Operator reconciles into a StatefulSet, its Services, and
the one-time admin credential Secret.

Not the right kind when:

- **The operator is missing** — a `KubernetesKeycloakOperator`
  watching this namespace is the PREREQUISITE (under its default
  namespaced watch, the operator and this resource live in the SAME
  namespace). Nothing reconciles this declaration without it.
- **You want a managed IdP** — the catalog's Auth0 kinds exist for
  exactly that; this resource is for running Keycloak yourself.
- **You want realms and clients as declarations** — deliberately not
  modeled: the operator's KeycloakRealmImport CR is a ONE-SHOT import
  Job (edits after a successful import are silently ignored upstream)
  and the OIDC/SAML client CRs are alpha, experimental-gated. Manage
  realms and clients through Keycloak's admin API/console, or declare
  the CRs via `KubernetesManifest` at your own risk.

## TLS-or-HTTP and hostname, decided up front

Keycloak REFUSES TO START with neither a TLS certificate nor plain
HTTP enabled — upstream surfaces the mistake only as a
CrashLoopBackOff; this spec makes it a validation error at apply
time. Set `http.tls_secret_name` (a kubernetes.io/tls Secret or a
`KubernetesCertificate` reference — the recommended posture) or opt
into `http.http_enabled` behind a TLS-terminating proxy. Hostname is
the same story: with strict resolution (the server default) the
public hostname is mandatory; set `hostname.strict: false` only
behind a trusted proxy that rewrites Host headers — and then set
`proxy_headers` too, or the server computes wrong origins and
browsers fail CORS (the classic "login randomly breaks").

## The database is required

Keycloak without an explicit database silently runs embedded H2 on
ephemeral pod storage and loses everything on restart — so `db` is
required, with a typed vendor enum. `postgres` is the recommended
production path (a `KubernetesPostgres` composes naturally: host
references its read-write Service, credentials ride Secret selectors
against the operator-maintained `<cluster>-app` Secret — nothing
credential-bearing ever rides this CR). The `dev-file`/`dev-mem`
sandbox vendors are embedded-H2 ephemeral storage, capped at a single
instance, never production.

## First login

Unless you bring your own bootstrap-admin Secret, the operator
generates the one-time `<name>-initial-admin` Secret (username
`temp-admin`, created once and never rotated) — exported as this
resource's credential handle. It seeds the FIRST admin only; create
durable admins inside Keycloak and treat the Secret as spent.

## Image changes take an outage

With the default update strategy, changing the image triggers a full
scale-to-zero recreate — two Keycloak versions cannot share one cache
cluster/schema, so the outage window is by design. `Auto` trades up
to ~5 minutes per spec change for a compatibility-check Job;
`Explicit` puts the recreate decision on your `update.revision` bump.

## Exposure and the operator defaults

The operator's own Ingress is ALWAYS disabled by this component —
exposure composes from Gateway API kinds referencing the exported
service handles. The operator's NetworkPolicy and ServiceMonitor
default ON; both are explicit spec fields (`network_policy_enabled`,
`service_monitor_enabled`) rather than hidden behavior.

## The 48-character name budget

Keep this resource's name at 48 characters or fewer: the operator
derives child names by suffixing (`-network-policy` is the longest at
15) and StatefulSet pod hostnames must stay DNS-legal. Both modules
fail loudly past the budget.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
