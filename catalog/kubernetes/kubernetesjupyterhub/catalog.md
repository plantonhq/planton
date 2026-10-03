# JupyterHub

Deploys a multi-user JupyterHub from the official Zero to JupyterHub Helm chart (`4.4.0` ships JupyterHub 5.5.0), secured by default where the chart itself is not. The chart's own sign-in default accepts ANY username with NO password, and it never ships here -- an empty `authentication` block means a module-generated shared password (the `<name>-auth` Secret), or declare your identity provider (GitHub, Google, or any OIDC issuer -- a composed Keycloak realm's endpoints slot straight in). Every teammate who signs in gets their own JupyterLab server pod with a persistent per-user home volume, spawned on demand and culled when idle. Two truths shape the design: chart resource names are FIXED bare strings (`hub`, `proxy`, `proxy-public`) so exactly ONE JupyterHub fits per namespace, and user homes are DATA -- `claim-<username>` PVCs created at runtime that deliberately survive uninstall.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **The Helm release** -- the official `jupyterhub/jupyterhub` chart at the pinned `chartVersion`, rendering:
  - The hub (`hub`, port 8081) -- authentication, spawning, the admin panel and REST API; SINGLE-REPLICA BY DESIGN (no HA upstream; a Recreate strategy pinned), its state on a sqlite PVC or your declared PostgreSQL/MySQL
  - The configurable-http-proxy (`proxy`) and its front-door Service (`proxy-public`) -- EVERY byte of user traffic flows through it
  - Per-user JupyterLab server pods (`jupyter-<username>`) and their home PVCs (`claim-<username>`) -- created AT RUNTIME as users sign in, not by the deploy
  - The packing user-scheduler, warm placeholder pods, the image pre-puller hooks and the idle-culler service, per your scheduling/culling declarations
- **The authentication Secret** -- with shared-password sign-in and no BYO Secret, the module generates the password into `<name>-auth` (readable by cluster admins; sign-in tokens ride environment indirection, never rendered values)
- **NetworkPolicies** -- the chart's hub/proxy/single-user policies render by default (`networkPolicyEnabled` unset = on)
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A namespace with no other JupyterHub in it** -- chart resource names are fixed bare strings; a second install in the same namespace collides. One hub, one namespace.
- **A hub database that ALREADY EXISTS, when declared** -- the postgres/mysql arms connect to a database that must pre-exist (the hub creates tables, never the database). On a composed KubernetesPostgres, declare `jupyterhub` at bootstrap (`initdb`); the spec's foreign-key defaults then wire the host to its read-write Service and the password to its application-user Secret.
- **Credential Secrets in THIS namespace** -- the database password Secret and any OAuth client-secret Secret are read from the workload's own namespace. Co-locate the hub with its database, or replicate the credential Secret.
- **RWX storage for the shared-volume home mode** -- one PVC mounted by every user's server pod is a ReadWriteMany concern on multi-node clusters. The default dynamic per-user mode has no such requirement.
- **A plan for user homes** -- `claim-<username>` PVCs survive this resource's destroy, but DELETING THE NAMESPACE DELETES EVERY USER'S HOME. Back them up with your storage story.

## Deploy

### Console

Open the deployment store, find **JupyterHub**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team notebooks — shared-password sign-in** preset for the fastest real multi-user platform (generated shared password, per-user 10Gi homes, hour-idle culling), or **Production — OIDC sign-in on PostgreSQL** for identity-provider sign-in on a composed PostgreSQL, both in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesJupyterHub
metadata:
  name: team-notebooks
  org: acme-corp
  env: dev
spec:
  namespace:
    value: "notebooks"
  createNamespace: true
  authentication:
    sharedPassword: {}
    adminUsers:
      - ada
  singleUser:
    image:
      repository: quay.io/jupyter/scipy-notebook
      tag: "2026-07-28"
    memoryLimit: 2G
```

```shell
planton apply -f team-notebooks.yaml
```

This declares the secured team default: JupyterHub 5.5.0 (the chart pin's release), shared-password sign-in with the password module-generated into `team-notebooks-auth`, one admin, a real scientific-Python image in place of the chart's evaluation sample, a 2G ceiling per user, and everything else by absence -- sqlite hub state on a 1Gi PVC, dynamic 10Gi per-user homes, hour-idle culling. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the hub's database to a PostgreSQL managed by another Infra Component:

```yaml
spec:
  namespace:
    value: "jupyterhub"
  createNamespace: true
  hub:
    database:
      postgres:
        host:
          valueFrom:
            kind: KubernetesPostgres
            name: hub-db
            fieldPath: status.outputs.rw_service
        passwordSecret:
          secretName:
            valueFrom:
              kind: KubernetesPostgres
              name: hub-db
              fieldPath: status.outputs.password_secret.name
```

The InfraPipeline deploys the PostgreSQL cluster first, then declares JupyterHub against it -- the SAME parent for both fields: the host resolves the cluster's read-write Service (always the current primary) and the password its application-user Secret. Declare the `jupyterhub` database at the Postgres kind's bootstrap (`initdb`).

## Key Configuration

These are the most important decisions when configuring JupyterHub. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Sign-in, read this first** -- the chart's own default accepts ANY username with NO password; it NEVER ships. An empty `authentication` block is the secured default: shared-password sign-in with the password module-generated into `<name>-auth`. Five explicit arms: `sharedPassword` (optionally BYO via `passwordSecret`) | `native` (per-user accounts inside the hub; `openSignup` means anyone reaching the login page can create a WORKING account) | `github` (client ID + client-secret Secret + callback URL; set `allowedOrganizations` in production or ANY GitHub identity signs in) | `google` (set `hostedDomains` or any Google account signs in) | `oidc` (six endpoints/fields; Keycloak, Okta, Auth0, Dex all compose). `adminUsers` and `allowedUsers` sit OUTSIDE the arm choice and apply to every method -- an empty allowed-users roster admits any authenticated identity.

**Hub state, sqlite is right for most installs** -- the hub is SINGLE-REPLICA BY DESIGN, so no `database` block (sqlite on a 1Gi PVC) carries hundreds of users. Declare `postgres` or `mysql` when hub state must survive volume loss or snapshot with your database fleet -- the database itself must PRE-EXIST (declare it at the Postgres kind's `initdb`). Named servers pair with their limit: `namedServerLimitPerUser` requires `allowNamedServers` -- the spec holds the pair.

**Per-user sizing has two dials that differ** -- `memoryGuarantee` (default 1G) is what the scheduler RESERVES; `memoryLimit` (empty = UNLIMITED) is where the kernel kills the kernel. Set a limit in production -- one runaway `pd.read_csv` on an unlimited pod evicts its node neighbors. The chart's sample image is evaluation-grade: point `singleUser.image` at a real notebook image (`quay.io/jupyter/scipy-notebook` or your own) for real work.

**Home storage has THREE postures** -- no `storage` block = a dynamic per-user PVC (10Gi default) -- the right default; `static` mounts ONE shared PVC with per-user sub-paths (RWX on multi-node clusters, quota is shared); `none` is the VANISHING WORKSPACE -- everything a user creates disappears when their server stops, and the idle culler stops servers by default. Ephemeral is legal and taught, never accidental.

**The spawn menu** -- `profiles` puts a machine-size menu at login ("Standard 2G", "Big-memory ETL", "GPU workstation"), each row overriding image or sizing with blank overrides falling back to the `singleUser` baseline. At most one row is `default: true`. No profiles = no menu -- every user gets the baseline directly.

**Capacity that scales DOWN** -- `userSchedulerEnabled` (default on) packs user pods onto the fewest nodes so autoscalers can reclaim the rest; `userPlaceholderReplicas` keeps N warm decoys a real user evicts INSTANTLY instead of waiting for a node boot. Culling is the cost model: idle servers stop after `timeoutSeconds` (default 3600) -- homes survive culls; never enable `cullUsers` with real user rosters (it deletes hub-database user records).

**The proxy fronts everything, privately by default** -- the chart's own `proxy-public` default is `LoadBalancer` (a public IP on install); this kind DELIBERATELY inverts it to ClusterIP. Reach the hub by port-forward, or compose exposure over the exported Service handle. HTTPS is deliberately unmodeled here -- TLS terminates at your composed ingress/gateway kind.

**The escape hatch, merged LAST** -- `helmValues` is raw chart values with Helm `-f` semantics for what the typed fields don't model (per-pod tolerations, custom hub config snippets, registry credentials). NEVER secret material -- the chart-owned hub Secret embeds the ENTIRE rendered values document, readable by anyone who can read Secrets; NEVER `proxy.https` or ingress keys -- exposure composes from first-class kinds.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** (optional) | `namespace` | `spec.name` |
| **KubernetesPostgres** | `hub.database.postgres.host` | `status.outputs.rw_service` |
| **KubernetesPostgres** | `hub.database.postgres.passwordSecret.secretName` | `status.outputs.password_secret.name` |
| **KubernetesMysql** | `hub.database.mysql.host` | `status.outputs.primary_service` |
| **KubernetesMysql** | `hub.database.mysql.passwordSecret.secretName` | `status.outputs.root_password_secret.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace JupyterHub runs in | Co-locating the database, exposure kinds and diagnostics |
| `proxy_public_service` | The front-door Service (`proxy-public`) -- every user request enters here | The handle exposure kinds route to |
| `endpoint` | In-cluster endpoint of the front door | What in-cluster clients use |
| `hub_service` | The hub's internal Service (`hub`, port 8081) -- the REST API | Automation driving the hub API (spawning, user admin) |
| `shared_password_secret` | Secret + key holding the generated shared password (EMPTY with GitHub/Google/OIDC/native sign-in) | Distributing the team password; rotating it |
| `port_forward_command` | kubectl port-forward one-liner for the hub UI | Reaching the UI from a workstation on the ClusterIP default |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team notebooks in minutes** -- The secured default: shared-password sign-in (generated into `<name>-auth`), a real scientific-Python image, per-user 10Gi homes that survive restarts, idle servers culled after an hour. Start from the **Team notebooks — shared-password sign-in** preset.

**Production on your identity provider** -- OIDC sign-in against a composed Keycloak realm (Okta/Auth0/Dex identical), hub state in a composed KubernetesPostgres, a spawn menu of machine sizes, warm placeholder pods keeping "start my server" instant. Start from the **Production — OIDC sign-in on PostgreSQL** preset.

**Classroom / ephemeral workshops** -- `storage.none` (no home PVCs to clean up), `native` accounts with `openSignup` inside a private network, aggressive culling (`timeoutSeconds: 1800`, `maxAgeSeconds: 28800`) -- the vanishing workspace as a feature, chosen deliberately.

## Works With

- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) -- the durable hub-state database; the `hub.database.postgres` foreign-key defaults point at it
- [**MySQL**](/infra-catalog/kubernetes-mysql) -- the MySQL alternative for hub state
- [**Keycloak**](/infra-catalog/kubernetes-keycloak) -- the natural OIDC issuer; a realm's endpoints slot into `authentication.oidc`
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the deployment
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) -- Gateway API exposure over the exported `proxy_public_service` handle, where TLS terminates
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) -- the Ingress alternative for exposing the front door
