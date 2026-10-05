# Apache Superset

Declares one Apache Superset install -- the open-source BI platform: dashboards, charts, and SQL Lab over any SQLAlchemy-speaking database. The official Helm chart renders the web tier (`supersetNode`, gunicorn on port 8088) against ONE required input: an external PostgreSQL metadata database, where dashboards, charts, users, saved queries and the ENCRYPTED datasource credentials live -- the chart's bundled PostgreSQL/Redis subcharts ride frozen legacy image lines and NEVER ship from this kind. SECURED BY DEFAULT: the session-signing `SECRET_KEY` (Superset refuses to start on its insecure default) and the bootstrap admin password are module-generated Secrets -- the chart's documented admin/admin default never deploys -- and every credential rides environment references composed into a module-owned Secret at apply time, never rendered literals. Declaring a cache/broker forks the install async: Celery workers execute SQL Lab queries, thumbnails and reports; beat fires the schedules; optional websocket and MCP arms add live result push and an AI-agent surface.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **The Helm release** -- the official `superset/superset` chart, rendering:
  - The web Deployment (`<name>-superset`, port 8088) -- the UI and REST API; 1 replica by default, optionally HPA-scaled
  - The init Job -- migrates the metadata schema (`superset db upgrade`), creates the bootstrap admin, optionally loads example dashboards
  - With a cache declared: the Celery worker Deployment, the beat scheduler (a SINGLETON by design), and optionally flower (port 5555), the websocket server, and the MCP server (port 5008)
  - Config rendered with checksum annotations, so `configOverrides` / `featureFlags` changes roll the pods automatically
- **Module-owned Secrets** -- the environment Secret (`<name>-env`, the chart's runtime-credential contract composed from YOUR referenced database/cache Secrets), the session-signing key (`<name>-secret-key`, STABLE across applies), the admin credential (`<name>-admin-auth`), and with websockets a shared JWT (`<name>-ws-jwt`)
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A PostgreSQL metadata database that PRE-EXISTS** -- the one required input. On a composed KubernetesPostgres, declare the `superset` database at its bootstrap (`initdb.database`); the init Job creates and migrates the SCHEMA but never the database itself.
- **Credential Secrets in THIS namespace** -- the module composes the chart's environment Secret by READING the database (and cache) credential Secrets at apply time, and a Secret can only be read from the workload's own namespace. Co-locate Superset with its database (the default composition), or replicate the Secrets.
- **A name within budget** -- keep `metadata.name` at 52 characters or fewer: the chart renders a `<name>-superset-celerybeat` Deployment whose suffix caps the name against the Kubernetes 63-character limit. The module fails loudly over budget.
- **A Redis-protocol store, for the async posture** -- a composed KubernetesValkey is the natural cache/broker; any Redis-compatible endpoint works as a literal.

## Deploy

### Console

Open the deployment store, find **Apache Superset**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team BI** preset for the smallest honest install against a composed database, or **Production BI** for the full async posture in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSuperset
metadata:
  name: team-bi
  org: acme-corp
  env: prod
spec:
  namespace:
    value: superset
  createNamespace: true
  metadataDatabase:
    host:
      value: superset-pg-rw.superset.svc
    passwordSecret:
      secretName:
        value: superset-pg-app
```

```shell
planton apply -f team-bi.yaml
```

This declares the smallest honest install: the web application against the named PostgreSQL, the session-signing key and admin password module-generated (sign in as `admin` with the password from the `team-bi-admin-auth` Secret), the official image at its pin. Without a cache this is the WEB-ONLY shape -- every query runs synchronously and the Celery family stays off. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the metadata database (and cache) to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    value: analytics
  metadataDatabase:
    host:
      valueFrom:
        kind: KubernetesPostgres
        name: superset-pg
        fieldPath: status.outputs.rw_service
    passwordSecret:
      secretName:
        valueFrom:
          kind: KubernetesPostgres
          name: superset-pg
          fieldPath: status.outputs.password_secret.name
  cache:
    host:
      valueFrom:
        kind: KubernetesValkey
        name: superset-cache
        fieldPath: status.outputs.service
    passwordSecret:
      secretName:
        valueFrom:
          kind: KubernetesValkey
          name: superset-cache
          fieldPath: status.outputs.password_secret.name
```

The InfraPipeline deploys the database and cache first, then declares Superset against them -- the SAME parent for each pair: the host resolves the resource's Service and the password its own credential Secret, composed into the environment Secret at apply time.

## Key Configuration

These are the most important decisions when configuring Superset. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The metadata database is the one required input, read this first** -- dashboards, charts, users, saved queries and the ENCRYPTED datasource credentials all live there; the Superset pods are otherwise stateless. It is ALWAYS external (the bundled subchart never ships). The database itself must PRE-EXIST -- on a KubernetesPostgres declare it at bootstrap (`initdb.database`); the init Job creates and migrates the schema under the configured user ("superset" by default). The password rides a Secret reference, never a literal.

**The cache is THE FORK** -- absent, this is the web-only Superset: every query runs synchronously and the whole Celery family (workers, beat, flower, websockets, MCP) must stay off -- the spec's own rules enforce it. Declared (a composed KubernetesValkey, or any Redis-protocol store), it becomes the query cache, the Celery broker and the async-results backend in one. The credential is OPTIONAL (absence = an auth-less store); `cacheDb`/`celeryDb` are Redis database numbers 0-15 where an explicit 0 is a real database (blank = the chart defaults 1/0).

**The worker dial is DEFAULT-TRUE** -- with the cache declared, an UNSET `worker.enabled` means the fleet RUNS (async SQL Lab queries, thumbnails, and alerts & reports execute there); the explicit FALSE is the deliberate web+cache-only opt-out (query caching without async execution). Beat is the SINGLETON clock that makes schedules FIRE -- off means alert schedules exist in the UI but nothing runs them; it requires the cache and the worker not explicitly disabled. Flower ships NO authentication of its own -- anyone reaching its Service sees task payloads INCLUDING EXECUTED SQL; keep it off or fence it with a NetworkPolicy.

**THE ROTATION STORY** -- the session-signing `SECRET_KEY` signs sessions AND encrypts the stored datasource credentials. The module generates it into `<name>-secret-key` and keeps it STABLE across applies. Changing it logs out every session and ORPHANS every stored datasource credential -- rotate only via `superset re-encrypt-secrets` with the old key in `PREVIOUS_SECRET_KEY`.

**THE DRIVER STORY, verified live** -- the official image is the driver-less "lean" build; even the PostgreSQL metadata driver ships only in dev/ci variants. The module's default `bootstrapScript` installs the exact psycopg2 pin -- and DECLARING a script REPLACES that default ENTIRELY: keep a driver install in it or the pods crash-loop with "No module named 'psycopg2'". Installs must target the app's venv (`uv pip install --python /app/.venv/bin/python <driver>`) -- the image's plain `pip` installs invisibly into the system interpreter. pip at container start needs internet and re-runs on every restart; for production bake a custom image and set the script to a no-op.

**Configuration is four surfaces** -- `featureFlags` (name → bool: `DASHBOARD_RBAC`, `ALERT_REPORTS`, `GLOBAL_ASYNC_QUERIES`), `configOverrides` (name → raw python appended to `superset_config.py` -- OAuth providers, row limits, timeouts), `extraEnv` (plain values), and `extraEnvFromSecret` (name → Secret reference). The pairing that keeps credentials out of every rendered surface: read secret values in overrides via `os.environ.get(...)` and declare them as secret-backed variables -- literals in overrides also sit in rendered Helm values.

**Realtime is two opt-in arms** -- `websockets` switches the `GLOBAL_ASYNC_QUERIES` transport to `ws` automatically (results push instead of polling; the shared JWT is module-generated), but its chart-default image is a COMMUNITY build with an UNPINNED tag -- pin deliberately for production. `mcp` (the Model Context Protocol server, port 5008) exposes dashboards and datasets to AI agents but requires the `fastmcp` python extra the official image does NOT include -- a custom image or a bootstrap-script install, or its pods crash-loop.

**The escape hatch, merged LAST** -- `helmValues` is raw chart values with Helm `-f` semantics for what the typed fields don't model. The module RE-PINS security-critical values AFTER the merge: the bundled postgresql/redis subcharts stay OFF and the chart's own env Secret stays OFF -- those cannot be silently re-enabled from here. NEVER secret material.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| PostgreSQL | `spec.metadataDatabase.host` | `status.outputs.rw_service` |
| PostgreSQL | `spec.metadataDatabase.passwordSecret.secretName` | `status.outputs.password_secret.name` |
| Valkey | `spec.cache.host` | `status.outputs.service` |
| Valkey | `spec.cache.passwordSecret.secretName` | `status.outputs.password_secret.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace Superset runs in | Co-locating exposure kinds and credential Secrets |
| `service` | The web Service (port 8088) | The handle exposure kinds route to |
| `endpoint` | In-cluster endpoint (`http://<service>.<namespace>.svc.cluster.local:8088`) | What browsers and API clients reach behind composed exposure |
| `admin_username` | The bootstrap admin username ("admin" by default) | Pairing with the credential below to sign in |
| `admin_password_secret` | Secret + key holding the bootstrap admin password | Signing in; wiring automation against the REST API |
| `env_secret_name` | The module-owned environment Secret (`<name>-env`) | Audit and advanced composition |
| `secret_key_secret_name` | The Secret holding the session-signing `SECRET_KEY` | The `superset re-encrypt-secrets` rotation procedure |
| `port_forward_command` | kubectl port-forward one-liner for the Superset UI | Reaching the UI from a workstation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team BI** -- The smallest honest install: the web application against a composed PostgreSQL metadata database, secured by the module-generated session key and admin credential. Every query synchronous -- right for small teams and evaluation. Start from the **Team BI** preset.

**Production BI** -- The full analytics posture: two web replicas, a Celery worker pair for async SQL Lab queries and thumbnails, beat firing scheduled alerts & reports, dashboard-level RBAC -- over a composed PostgreSQL and a composed Valkey cache/broker. Start from the **Production BI** preset.

**The AI-connected warehouse** -- Superset over a composed Trino coordinator endpoint (one datasource that JOINs across every catalog), with the MCP arm exposing dashboards and datasets to AI agents -- a custom image carrying the `fastmcp` extra and the Trino driver.

## Works With

- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) -- the metadata database; the `metadataDatabase` foreign-key defaults point at it
- [**Valkey**](/infra-catalog/kubernetes-valkey) -- the natural cache/broker; the `cache` foreign-key defaults point at it
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the deployment
- [**Trino**](/infra-catalog/kubernetes-trino) -- the federated SQL engine Superset naturally fronts; connect its exported coordinator endpoint as a datasource
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- cluster monitoring for the web and worker tiers
