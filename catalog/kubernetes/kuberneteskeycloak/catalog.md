# Keycloak

Declares a Keycloak server — the open-source identity and access management platform (OIDC, SAML, user federation) — as a `Keycloak` custom resource reconciled by the official Keycloak Operator. Keycloak ships no official Helm chart; the operator is the first-party Kubernetes distribution, and this component renders its CR faithfully: the operator turns the declaration into a running StatefulSet, its Services, network policy, and the one-time bootstrap admin credential.

The spec's purpose is converting **crash-loops into apply-time errors**: the CR itself requires almost nothing, but the server refuses to start without a database, a served-or-delegated TLS answer, and (in strict mode) a hostname. Every one of those server-startup rules is validated on this spec before anything deploys.

**Prerequisite**: a **Keycloak Operator** install watching this namespace — under its default namespaced watch, the operator and this resource live in the SAME namespace. Without it the declaration sits unreconciled.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **The `Keycloak` custom resource** (`k8s.keycloak.org/v2alpha1`) — the operator reconciles it into:
  - A **StatefulSet** named exactly after this resource, running the requested number of server instances
  - The main **Service** (`<name>-service`) — https 8443 and/or http 8080, plus the management port
  - The headless **discovery Service** (`<name>-discovery`) — JGroups cluster formation between instances; clustering needs nothing beyond it
  - A **NetworkPolicy** scoped to the server pods (on by default; a tri-state dial)
  - The **`<name>-initial-admin` Secret** — the operator-generated bootstrap admin credential (username `temp-admin`), created once and never rotated
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist
- **No Ingress** — the operator's own default Ingress is always disabled by this component; exposure composes from Gateway API kinds referencing the exported service handles

Realm imports and OIDC/SAML client CRs are deliberately not modeled — the import is a one-shot Job and the client CRs are alpha surfaces. Manage realms through Keycloak's admin console or Admin API, or apply those CRs via the Kubernetes Manifest escape hatch.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **A Keycloak Operator watching the namespace** — deploy **Keycloak Operator** first; its CRDs define the `Keycloak` kind this component declares.
- **A database** — required for anything real. A **PostgreSQL** resource composes naturally by reference; the embedded `dev-file`/`dev-mem` sandbox vendors exist only for evaluation.
- **A TLS answer** — a certificate Secret (a **Cert Manager Certificate** resource composes by reference), or a TLS-terminating proxy in front if you enable the plain-HTTP listener.
- **Name budget** — the operator derives every child name by suffixing this resource's name; keep `metadata.name` within 48 characters so the longest derived name stays DNS-legal.

## Deploy

### Console

Open the deployment store, find **Keycloak**, and click **Deploy**. The creation wizard walks you through namespace placement, the required database, the TLS-or-HTTP listener, the hostname contract, clustering, the image and its update strategy, sizing, scheduling, features, server options, admin access, health probes, the operator integrations, and tracing. Start from the **Standard preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeycloak
metadata:
  name: keycloak
  org: acme-corp
  env: prod
spec:
  namespace:
    value: keycloak
  instances: 2
  db:
    vendor: postgres
    host:
      value: keycloak-db-rw.keycloak.svc.cluster.local
    database: keycloak
    usernameSecret:
      name:
        value: keycloak-db-app
      key: username
    passwordSecret:
      name:
        value: keycloak-db-app
      key: password
  http:
    tlsSecretName:
      value: keycloak-tls
  hostname:
    hostname: https://auth.example.com
```

```shell
planton apply -f keycloak.yaml
```

This creates a two-instance clustered Keycloak backed by the named Postgres database, serving TLS from the `keycloak-tls` Secret and advertising `https://auth.example.com` as its identity. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the database by reference — the host resolves to the Postgres read-write Service (always the current primary) and the credentials ride Secret selectors against the operator-maintained app-credential Secret:

```yaml
spec:
  db:
    vendor: postgres
    host:
      valueFrom:
        kind: KubernetesPostgres
        name: keycloak-db
        fieldPath: status.outputs.rw_service
    database: keycloak
    usernameSecret:
      name:
        valueFrom:
          kind: KubernetesPostgres
          name: keycloak-db
          fieldPath: status.outputs.password_secret.name
      key: username
    passwordSecret:
      name:
        valueFrom:
          kind: KubernetesPostgres
          name: keycloak-db
          fieldPath: status.outputs.password_secret.name
      key: password
```

The InfraPipeline deploys the operator and the database first, then provisions Keycloak against them. Nothing credential-bearing rides this declaration — the selectors name a Secret and its keys, and the server reads them in its own namespace.

## Key Configuration

These are the most important decisions when configuring a Keycloak server. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The database is required** — `db.vendor` takes one of eight values: `postgres` (recommended — a PostgreSQL composes by reference), `mysql`, `mariadb`, `tidb`, `mssql`, `oracle` for databases you operate, and the never-production sandbox vendors `dev-file`/`dev-mem` (embedded H2 on ephemeral pod storage or memory — data dies with the pod, and the spec caps them at a single instance). Real vendors need host + database + both credential selectors, or a full `jdbcUrl`.

**TLS or HTTP, decided up front** — the server refuses to start with neither. Recommended: `http.tlsSecretName` referencing a **Cert Manager Certificate**. Enable the plain-HTTP listener only behind a TLS-terminating proxy, paired with `proxyHeaders` (`xforwarded` or `forwarded`) so the server trusts the forwarded scheme and host.

**The hostname is identity, not routing** — `hostname.hostname` (a full URL, e.g. `https://auth.example.com`) is what tokens, redirects, and the OIDC discovery document advertise. The working full-surface posture is a fixed public URL plus `backchannelDynamic: true`, letting in-cluster clients reach the server on its Service address. `strict` (default true) makes the server refuse to answer without a declared hostname — and is ignored once a hostname is set. Both full-URL rules mirror the server's own startup errors.

**Image changes take an outage window** — under the default update strategy, changing `image` triggers a full scale-to-zero recreate (two Keycloak versions cannot share one cache cluster and schema). `update.strategy: Auto` lets the operator decide; `Explicit` recreates only when `update.revision` changes.

**Two default-true dials** — `networkPolicyEnabled` and `serviceMonitorEnabled` are tri-state: unset means enabled. Disable the NetworkPolicy only when a cluster-level policy engine owns isolation; the ServiceMonitor engages only where the Prometheus Operator CRDs exist.

**Sizing and probes** — Keycloak is a JVM; the module's default sizing (requests `250m`/`768Mi`, limits `1`/`1Gi`) is lab-grade, and production typically runs 1–2Gi. The startup probe's default budget is roughly 10 minutes because first boots run schema migrations — don't tighten it reflexively.

**Bootstrap admin is literal bootstrap** — the `<name>-initial-admin` Secret (or your own via `bootstrapAdminSecretName`) seeds the FIRST admin login only, is never rotated, and never reflects password changes made inside Keycloak. Create durable admin users in Keycloak and treat the bootstrap Secret as break-glass material. Realm data itself survives pod replacement — state lives in the database, not the pods.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesPostgres** | `db.host` | `status.outputs.rw_service` |
| **KubernetesPostgres** | `db.usernameSecret.name` / `db.passwordSecret.name` | `status.outputs.password_secret.name` |
| **KubernetesCertificate** | `http.tlsSecretName` | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the server runs in | Composition |
| `stateful_set` | The StatefulSet name (exactly this resource's name) | Debugging, monitoring |
| `service` | The main Service (`<name>-service`) | Gateway API exposure, in-cluster clients |
| `discovery_service` | The headless JGroups Service (`<name>-discovery`) | Cluster-formation diagnostics |
| `api_endpoint` | In-cluster API endpoint, scheme included | OIDC/SAML issuer for in-cluster clients |
| `management_endpoint` | The management port endpoint | Health probes, metrics scrape configuration |
| `initial_admin_secret_name` | The bootstrap admin credential Secret | First login; break-glass access |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command | Workstation access without exposure |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard** — the production shape: two instances clustering through the discovery Service, a PostgreSQL referenced for the database, TLS from a Cert Manager Certificate reference, and a declared public hostname. Start from the **Standard preset**.

**Dev sandbox** — the smallest Keycloak that starts: `dev-mem`, the plain-HTTP listener, strict resolution off. Deliberately disposable — data dies with the pod. Start from the **Dev-sandbox preset**.

## Works With

- [**Keycloak Operator**](/infra-catalog/kubernetes-keycloak-operator) — the manager that reconciles this declaration; deploy it FIRST, in this same namespace under its default watch, and destroy declarations BEFORE the operator on the way out.
- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) — the recommended production database; its read-write Service and app-credential Secret compose by reference.
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — issues the TLS Secret the listener serves.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — provides the namespace by reference; co-locate the operator, the database, and Keycloak so the credential secretKeyRefs resolve.
