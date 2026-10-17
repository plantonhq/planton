# KubernetesLangfuse Guide

Langfuse looks like one application, but its durable state lives in four stores this kind never owns. Most mistakes come from treating the KubernetesLangfuse resource as the thing to protect, when the thing to protect is Postgres and the two secrets bound to it.

## When to use it (and when not)

Use KubernetesLangfuse when a team wants Langfuse (LLM tracing, prompt management, evaluations, datasets) running on its own cluster, with its data in stores it controls. It is the one kind for Langfuse in the catalog; there is no separate kind for the web server or the worker, because both run from one configuration and one release.

The stores are always separate resources. A KubernetesPostgres, KubernetesClickHouse, KubernetesValkey and KubernetesSeaweedFs is the in-cluster set, and every store field's composition default points at the matching kind, so a reference names only the resource. Each store can also be a managed service by literal host (RDS or Cloud SQL, ClickHouse Cloud, ElastiCache or Memorystore) and object storage can be S3, R2, GCS or Azure Blob. Mix freely, with one exception: Langfuse switches all upload types to Google Cloud Storage or Azure Blob Storage at once, so when any blob target is GCS (or Azure) every target must be, and the spec refuses a mix.

Do not reach for this kind to collect general application telemetry. Langfuse accepts OTLP traces at its `otel_endpoint`, but it is built around LLM calls, prompts and scores. A metrics, logs and traces stack for services is a different kind (KubernetesSignoz, for example), and the two can run side by side.

## Conventions and gotchas

**The salt and encryption key belong to the database, not to this resource.** The module generates `secrets.salt` and `secrets.encryptionKey` once and keeps them across applies, so day-to-day changes are safe. Deleting the resource and creating it again against a surviving Postgres is not: the new resource generates new values, every existing API key stops authenticating, and every stored LLM credential becomes unreadable, with no recovery. For an install you mean to keep, declare both as `$secret/` references from the first deploy. If an install started with generated values, copy them from the instance Secret (`instance_secret_name` output, keys `salt` and `encryption-key`) into managed secrets and declare them before anyone deletes anything. Changing `secrets.nextauthSecret` only signs everyone out.

**Everything Langfuse needs must live in its own namespace.** Each store password reaches the pods through a secretKeyRef, and a secretKeyRef cannot cross namespaces. Put the stores in Langfuse's namespace, or copy their Secrets there. A store in another namespace shows up as pods stuck in `CreateContainerConfigError`, not as a plan error.

**A ClickHouse user without `networks` looks like a wrong password.** A KubernetesClickHouse user declared without networks is fenced to ClickHouse's own pods, and Langfuse's first migration fails with an authentication error. Declare the Langfuse user's networks (for example `0.0.0.0/0` and `::/0`) and leave its password empty: the ClickHouse module generates it into the auth Secret under the user's name, which is exactly where `clickhouse.passwordSecret` looks by default.

**Cluster mode is a first-start decision.** `clickhouse.cluster.enabled` is on by default and matches a KubernetesClickHouse with replicas or a managed Keeper. On a single server without Keeper the migrations fail until it is off. Turning it on or off after data exists means migrating data, because the table engines differ.

**Valkey must run with `noeviction`.** The ingestion queues live in the Redis-protocol store. Under an eviction policy, Valkey drops queued jobs when memory fills, and traces vanish with no error anywhere. Set the policy on the KubernetesValkey, and size its memory for the backlog you can tolerate while workers catch up.

**Presigned URLs need a reachable object store.** Browsers and SDKs upload media and download exports straight from object storage, using URLs Langfuse signs for the target's endpoint. With an in-cluster SeaweedFS, set `blobStorage.mediaUpload.publicEndpoint` and `blobStorage.batchExport.publicEndpoint` to the address a route serves the store at; Langfuse then signs for the public address and keeps reading through the internal one. Media uploads cannot be turned off in this chart version, so plan the public endpoint before anyone traces images or audio. Event uploads stay inside the cluster.

**Headless initialization runs once per object.** The seeded organization, project, key pair and admin are created only if they do not exist. Renaming them, changing the admin password or the key pair later changes nothing that already exists, except that a new key pair is added beside the old one without revoking it. Treat `init` as the bootstrap, and manage everything after first start in the UI or through the API.

**Generated credentials are readable outputs, not copies.** `project_secret_key_secret` and `admin_password_secret` point at keys in the instance Secret. Wire `project_secret_key_secret` into an application's Deployment as a secretKeyRef rather than copying the value anywhere.

**Some Langfuse variables only work through their typed field.** The chart renders several settings itself and refuses a variable set both by a typed value and through `extraEnv`. Langfuse also ignores the chart's own experimental-features switch; `features.experimentalFeaturesEnabled` sets the variable Langfuse actually reads. When a variable has a field, use the field.

**Deleting leaves one empty Secret behind.** The chart always creates a `<name>-app` Secret with Helm's keep policy, so uninstall never removes it. Because the module supplies the salt, encryption key and session secret by reference, that Secret holds no keys. It is harmless, and a later install with the same name reuses it.

**Imports carry companion keys.** The instance Secret stores the encryption key's bytes in standard base64 (`encryption-key-std-b64`) and each generated project key without its prefix (`init-project-public-key-suffix`, `init-project-secret-key-suffix`), beside the forms Langfuse reads. Those are the values an import of an existing install needs; nothing else reads them.

## Configuration coverage

The spec is built for full parity with the official chart's values and Langfuse's documented self-hosting variables, pinned at chart 2.1.5 (Langfuse 4.56.0). Until a check measures that parity mechanically, treat it as a design goal, not a measured claim. What the spec deliberately leaves out, and why, is listed in the spec's own description in the [reference page](v1alpha1/reference.md): the bundled stores (superseded by their own kinds), the chart's ingress (compose a route instead), chart 1.x upgrade guards, build-time settings that need a custom image, OCI's native storage driver, generic pod plumbing, and Langfuse Cloud internals.

When you need something the spec does not type, use the narrowest escape hatch: `extraEnv` for a plain Langfuse variable, `extraEnvFromSecret` for a secret one, and `helmValues` for a chart key. The module merges `helmValues` last and then re-pins the release name and the bundled stores off, so the escape hatch can never bring a bundled store back.

## On the diagram

A Langfuse install renders as five nodes: Langfuse and its four stores, with edges from every reference. That is the point of composing the stores. Each store's backups, sizing and version show as its own node with its own spec, where a bundled store would render as nothing. Add the operators (CloudNativePG and Altinity) and a route, and the diagram shows the whole dependency chain in deployment order. A Langfuse wired to literal hosts renders without those edges, which is honest for managed services outside Planton and a missed opportunity for stores that Planton manages.

## Pairs well with

- **KubernetesPostgres** with a backup configured: it holds everything that cannot be rebuilt (users, projects, prompts, API keys), and the salt and encryption key are only useful beside it.
- **KubernetesClickHouse** with a managed Keeper: cluster mode on from the start, and room to add replicas later without migrating tables.
- **KubernetesValkey** with `noeviction`: the queue store.
- **KubernetesSeaweedFs** in-cluster, or a cloud bucket kind with workload identity: event, media and export storage. A lifecycle rule that expires the `events/` prefix after about 30 days is safe once events are in ClickHouse; never expire `media/`.
- **KubernetesHttpRoute** or **KubernetesIngress** over the `service` output, at the host `externalUrl` names.
- The `self-hosted-langfuse` InfraChart, which composes all of the above in one deployment.
