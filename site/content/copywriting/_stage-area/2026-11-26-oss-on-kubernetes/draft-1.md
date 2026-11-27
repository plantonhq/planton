# Open source on your own cluster (draft 1)

The stage area for planton.ai's one page, which is built one or two sections per session. This session adds Home's section "run open source on Kubernetes instead of managed services". In the page's final order it is section 6 of 9. Today it sits between "ask your coding agent for infrastructure" and "run Planton yourself"; later sections go in above it. It was written 2026-11-26 under `../../_rules/update-planton-ai-copy-writing.mdc`, and every sentence sits beside the chapter and ledger row it rests on.

The story changed first. The positioning story is now v3.3 (2026-11-26): chapter 8's proof carries open source on your own cluster, and chapter 10 counts the catalog's open-source group. The claims ledger gained the set-deploy row and the open-source row.

## The media every candidate shares: one real ask, on a fresh cluster

A headless coding agent, with the released Planton CLI and skill and no account, got the ask below. It wrote six files from the catalog and ran one command. The session then checked the cluster itself. The record (the ask, the files, the full log, the checks, the transcript) is kept in the platform repository's website research.

**The ask** (verbatim): "We're a small team and nobody here does DevOps full time. Our dev environment pays for a managed Postgres, a managed Redis-style cache and a hosted sign-in service, and those bills add up. We already have a Kubernetes cluster for dev (kubectl context `dev`). Use Planton to run Postgres, Valkey and Keycloak on that cluster instead, with Keycloak keeping its data in that Postgres. Don't sign in to Planton; deploy it straight to the cluster. Then show me it's running."

**keycloak.yaml, one of six files the agent wrote** (verbatim, its `db:` block; the six: cnpg-operator.yaml, namespace.yaml, keycloak-operator.yaml, postgres.yaml, keycloak.yaml, valkey.yaml):

```yaml
  db:
    vendor: postgres
    host:
      valueFrom:
        kind: KubernetesPostgres
        name: postgres
        fieldPath: status.outputs.rw_service
    database: keycloak
    usernameSecret:
      name:
        valueFrom:
          kind: KubernetesPostgres
          name: postgres
          fieldPath: status.outputs.password_secret.name
      key: username
    passwordSecret:
      name:
        valueFrom:
          kind: KubernetesPostgres
          name: postgres
          fieldPath: status.outputs.password_secret.name
      key: password
```

**planton apply** (verbatim lines of the 20,571 it printed; `⋯` marks each cut):

```text
$ PLANTON_BACKEND=none planton apply -f infrastructure/ --kube-context dev --auto-approve
ℹ No backend configured -- running with the Planton open-source engine. Run `planton login` to connect to planton.ai.
⋯
✔ deploy order: KubernetesCloudNativePgOperator/cnpg -> KubernetesNamespace/dev-services -> KubernetesKeycloakOperator/keycloak-operator -> KubernetesPostgres/postgres -> KubernetesKeycloak/keycloak -> KubernetesValkey/valkey
⋯
✔ preflight passed: 25 facts verified, 1 warnings, 2 assumptions stated
⋯
✔ KubernetesCloudNativePgOperator/cnpg deployed
✔ KubernetesNamespace/dev-services deployed
✔ KubernetesKeycloakOperator/keycloak-operator deployed
✔ KubernetesPostgres/postgres deployed
✔ KubernetesKeycloak/keycloak deployed
✔ KubernetesValkey/valkey deployed
✔ All 6 resources deployed
```

**Your cluster, after** (each answer verbatim from the checks):

| Service | Asked | Answer |
|---|---|---|
| Postgres | `kubectl get cluster postgres` | Cluster in healthy state |
| Postgres | Keycloak's tables in the `keycloak` database | 100 |
| Valkey | `valkey-cli ping`, without its password | NOAUTH Authentication required. |
| Valkey | `valkey-cli ping`, with the password it generated | PONG |
| Keycloak | `kubectl get keycloak` | READY True |
| Keycloak | `GET /realms/master` | 200 |

**What else the catalog runs on your cluster** (36 projects, each green in its own test; grouped):

| Kind of work | Projects |
|---|---|
| Databases | Postgres, MySQL, MongoDB, ClickHouse, Neo4j, Qdrant |
| Caches and messaging | Valkey, Kafka, Karapace, Kafka UI, RabbitMQ, NATS |
| Search and analytics | OpenSearch, Solr, Trino, Superset |
| Data and AI | Airflow, Ray, Flink, JupyterHub, MLflow, Langfuse |
| Sign-in and secrets | Keycloak, OpenBao, OpenFGA |
| Observability | Grafana, Loki, Tempo, SigNoz |
| Delivery and storage | Temporal, Harbor, GitHub Actions runners, SeaweedFS, Locust, flagd, GO Feature Flag |

| Rests on | |
|---|---|
| The ask, the file, the command, the output, the answers | the walk's record (2026-11-26); the ledger's set-deploy row |
| No account; the open-source CLI | the log's first line; the ledger's set-deploy row (its condition: signed out) |
| Deployed in order, operators first | the log's deploy-order line; ch 8 proof |
| No file holds a password | the six files; Valkey's NOAUTH then PONG |
| On the cluster you already pay for | ch 8 claim ("your bill"); the ask |
| The 36 names and "35+" | the ledger's open-source row; `platform-stats.ts` (its rule); each kind's `e2e/profile.yaml` |

## Candidate A: the proof chain (the lead's recommendation going in)

- **Heading:** Run Postgres, Valkey and Keycloak on the cluster you already pay for.
- **Line 1:** Your coding agent writes them from Planton's catalog, and the open-source CLI deploys them as one ordered set, operators first. No account needed. *(ch 8 proof; ledger set-deploy row)*
- **The ask (left, quoted):** the ask above.
- **Artifacts (right):** keycloak.yaml's `db:` block, labelled "keycloak.yaml · Keycloak reads its database from Postgres"; then planton apply; then Your cluster, after.
- **Below (left):** "No file holds a password: each service generates its own, into your cluster." Then "The catalog runs 35+ open-source projects on your cluster the same way, each proven by its own test:" and the grouped table. *(ch 8 proof, ch 10; ledger open-source row)*

## Candidate B: the catalog first

- **Heading:** Postgres, Kafka, Keycloak, Grafana and 30+ more, on the cluster you already pay for.
- **Line 1:** Run the open-source project itself where your apps already run, instead of paying a managed service for each. The open-source CLI deploys a folder of them in order, with no account. *(ch 8 proof; ledger rows)*
- **Artifacts (right):** the grouped table first, then "Three of them, from one ask": the ask's middle sentence, planton apply, Your cluster, after.
- **Below:** "No file holds a password."

## Candidate C: one row per bill

- **Heading:** Postgres, Valkey and Keycloak on your own cluster, not three managed bills.
- **Line 1:** The open-source CLI deploys a folder of the catalog's open-source projects in order, on any cluster your kubeconfig names, with no account. Everything stays on the cluster you already pay for.
- **The ask (left, quoted).**
- **Artifacts (right):** planton apply (the command, the first line, `All 6 resources deployed`); then a table "What replaced each bill": Postgres (postgres.yaml, deployed, healthy with Keycloak's 100 tables), Valkey (valkey.yaml, deployed, NOAUTH then PONG), Keycloak (keycloak.yaml, reads Postgres by reference, deployed, Ready, realm 200).
- **Below:** the grouped table.

## Never said, checked
- No dollar or percentage saving, no cost figure, no time figure.
- No managed product by its vendor's name.
- No "any open-source software", "production-ready", "highly available" or "zero lock-in".
- Never that the signed-in product does this; never the host (laptop, kind).
