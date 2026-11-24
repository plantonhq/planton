# KubernetesStigmer Guide

The judgment this guide carries: this kind is the one Stigmer on a cluster, its operator and its install in one block, because Stigmer runs one of each and the operator's version is Stigmer's. It keeps the one thing that cannot be recreated, the encryption keys, out of everything a destroy removes, and it refuses what Stigmer would refuse later, or never say: a contradictory declaration, a second install, a downgrade, and a restore without its keys.

## One operator, one Stigmer, one version

Stigmer's resource has no version field: the operator runs Stigmer at its own version, and the chart, the operator and the server share one version line. Two kinds (an operator and its Stigmer) would put the version on a resource other than the one people configure, for an operator that serves exactly one install. So `version` lives here, required and never defaulted: a default that moved with catalog updates would upgrade a running install on an ordinary re-apply.

Names are fixed where Stigmer derives everything from them. The Helm release is `stigmer-operator` (the chart names its cluster-wide role, and the definition records its owner, after the release), and the Stigmer is `stigmer` (every object the operator creates is named after it, and the outputs read those names). The operator's namespace is a field, `stigmer-system` by convention, created here only with `operator.create_namespace`, as the Planton operator's is.

## Two namespaces, two owners

The Stigmer's namespace holds the encryption keys (`stigmer-keys`), and this kind never creates or deletes it: a module-owned namespace would take the keys with it on destroy and make every backup unreadable. Give it a KubernetesNamespace of its own, and run `keys_export_command` before anyone deletes that namespace or the environment around it.

The operator's namespace is only where the controller runs. It may be created here; it may not be created here and also be the Stigmer's namespace (refused at validation and again at plan time, after references resolve). It is `containment_exempt` on the diagram: the Stigmer is placed in its own namespace, not in the operator's.

## Credentials: values or your own Secrets

Every credential is either a value (a Planton secret reference, which this kind writes into a Secret under the names Stigmer reads) or a Secret you keep in the Stigmer's namespace. Each Secret this kind writes holds only the keys declared: the runner maps all five model keys as optional, so a key present but empty would reach it as an empty setting.

Cross-kind credentials compose where the names agree. A KubernetesPostgres's application Secret holds the password under `password`, so `database.external.password.existing_secret` names it directly when the database runs in the Stigmer's namespace. A SeaweedFS's key pair does not (Stigmer reads `ACCESS_KEY_ID` and `SECRET_ACCESS_KEY`), so give it again as a value.

`server.extra_env` and `runner.extra_env` take the shared ContainerEnv: plain `variables` are written into the Stigmer's spec, where anyone who can read it sees them, and `secrets` reach the containers by reference. A setting Stigmer's operator sets itself is refused at validation, against the exact list the operator owns at the validated version, because a second value would silently replace the declaration. `env_from` is refused: Stigmer takes single settings.

## Sign-in follows the front door

Stigmer's rule, mirrored in validation: with no `ingress`, nobody signs in and every caller is the owner. With an `ingress` and no `identity`, the bundled sign-in runs and needs an `owner`, whose account is created first (`owner_password_command`). The bundled sign-in keeps its accounts on the bundled database, so it cannot run on your own Postgres: that combination needs `identity.oidc`. `allow_unauthenticated_exposure` is the deliberate way to expose an install with sign-in off, and it cannot be combined with an `identity`.

## Restore needs the keys

`database.bundled.recover_from` creates the database from an archive in the bucket, and only when the database is created; on a running install it changes nothing. With no `target_time` it restores the newest moment the archive holds, and the newest changes reach the archive only when the database moves on to its next WAL file, which can lag by minutes: take an on-demand backup before a destroy whose data you mean to restore. The archive is sealed with the source install's keys. If `stigmer-keys` is missing, the operator mints new ones and the restored records cannot be read, so the restore check stops the deploy before the Stigmer is declared and names the fix: apply the file `keys_export_command` saved. The check runs again whenever the namespace or the archive changes, and passes while a Stigmer already runs.

## Upgrades and the way back

A `version` move rolls the operator, which rolls the server. With backups on, the operator backs the database up before the server moves, and holds the upgrade while that backup fails; `upgrade_without_backup_to`, equal to `version`, releases that one upgrade (it is the `stigmer.ai/upgrade-without-backup` annotation, in the operator's own spelling of the version). A `version` below the one the server runs (`status.version`) is refused, because the server refuses a database a newer release migrated. Rolling back a chart whose upgrade is still held is allowed, because the server still runs the older version.

## Destroy

The Stigmer has no finalizer, so deleting it returns at once, and the operator's janitor cleans up only once no Stigmer is left. A destroy that removed the chart right away would kill the janitor mid-sweep and strand CloudNativePG and the sandbox namespace. So the destroy waits between the Stigmer and the chart, by the janitor's own rules: this Stigmer's database objects and pods first (their last step archives the final changes, which needs the bucket's credentials, still in place), then the sandbox namespace, then each prerequisite to be gone, someone else's, or still holding another install's objects. The wait is bounded and never fails a destroy. Under the `substrate` driver it runs out: the janitor deletes the sandbox namespace but never ends a live session's Substrate actor, whose worker then drains for up to an hour (its termination grace), so the destroy names the prerequisites that stay and the commands that remove them.

## Prerequisites from the catalog

Stigmer's operator installs CloudNativePG, its backup plugin, agent-sandbox and Agent Substrate when `prerequisites` says `auto` (the default) and the cluster lacks them, and leaves a copy someone else installed alone. The only conflict is order: if the operator installs its copy first, the catalog's kind later finds it and refuses it. When the environment holds those kinds, give this one a `depends_on` relationship to each. A copy someone else runs is used as it is, so it must be at least as new as what Stigmer bundles where Stigmer relies on newer fields: the backup plugin from v0.16.0 (KubernetesCnpgBarmanCloudPlugin's default, `chart_version` 0.9.0), because with a named S3 endpoint Stigmer writes the object store's addressing style, which older releases do not declare.

A bucket with the bundled database turns the database's archive on from its first minute, whatever the bucket was declared for, and the backup plugin that writes the archive needs cert-manager. The plugin is installed into CloudNativePG's namespace and the database is created only once it serves, so without cert-manager the Stigmer stays `Pending` with `WaitingForBackupPlugin` ("installing the Barman Cloud backup plugin") and never says why. Give the Stigmer a `depends_on` to a KubernetesCertManager whenever it declares `object_store` and runs the bundled database.

The operator recognizes someone else's copy by its definition alone, and trusts it without asking whether anything runs it. The catalog's CloudNativePG kinds keep their definitions when destroyed, so a cluster where one was removed looks to the operator like a CloudNativePG it must leave alone, and the bundled database waits forever. In `auto` mode the plan therefore reads each definition the install needs; one a Helm release put there whose release has no Deployment left is refused, with the command that clears it. A definition installed without Helm names no controller to look for, and is trusted.

## Sizes and what they reserve

Every size is optional: the operator fills what you leave out from its own defaults for the declared version, sized for one team, key by key (declare only a memory limit and the requests stay the operator's). Planton's capacity estimate counts only the sizes you declare, because the defaults move with the version; preset `08-sized-for-one-team` writes the 0.3.2 defaults out, so its estimate shows what one install reserves. Temporal takes one size for each of its four services. The shared runner exists only without a per-session sandbox driver, so its sizes are refused with one.

## On the diagram

KubernetesStigmer draws inside its namespace. The operator's namespace is drawn as a reference, not a container. Its references to a KubernetesPostgres, a KubernetesTemporal, a SeaweedFS, an issuer and a Gateway draw as edges.

## Parity accounting

Built for 100% parity with the `stigmer.ai/v1` Stigmer resource and the `stigmer-operator` chart at Stigmer 0.3.2.

### Chart values

| Chart value | Spec field |
|-------------|------------|
| `image.repository`, `image.pullPolicy`, `image.pullSecrets` | `operator.image_repository`, `.image_pull_policy`, `.image_pull_secrets` |
| `replicas`, `resources`, `logLevel`, `nodeSelector`, `tolerations` | `operator.replicas`, `.resources`, `.log_level`, `.node_selector`, `.tolerations` |
| `image.tag` | not modeled: the chart runs the operator at its own version, and a different tag would run a different Stigmer from `version` |
| `localRegistry` | not modeled: the chart documents it as a development-only setting |

The chart's values schema is closed, so there is no free-form escape hatch: the only keys the typed fields leave out are the two above.

### The Stigmer resource

| Resource field | Spec field |
|----------------|------------|
| `image` | `image` |
| `owner` | `owner` |
| `server.resources`, `.storage`, `.artifacts`, `.help`, `.internalAccounts`, `.allowUnauthenticatedExposure`, `.egress`, `.extraEnv` | `server.*`, with `extra_env` as ContainerEnv |
| `runner.llm.secretName` | `runner.model` (`keys`, or `existing_secret_name`) |
| `runner.resources`, `.storage`, `.extraEnv` | `runner.*`; `resources` and `storage` are refused under a per-session sandbox driver, which runs no shared runner for them to size |
| `database.postgresql.bundled` | `database.bundled` |
| `database.postgresql.external` | `database.external`, with `password` as a value or a Secret |
| `temporal.bundled`, `temporal.external` | `temporal.*`; `external.tls` as PEM values or a Secret, `external.apiKey` as `api_key` |
| `objectStore.s3` | `object_store.s3`, with `credentials` as a key pair or a Secret |
| `ingress` (`tls` or `gatewayRef`) | `ingress` (`tls` or `gateway`) |
| `identity.bundled` (`signUp.emailDomains`, `smtp`) | `identity.bundled` (`sign_up_email_domains`, `smtp` with `credentials`) |
| `identity.oidc` | `identity.oidc` |
| `enterprise` | `enterprise`, with `license` as a value or a Secret |
| `sandboxes` (`driver`, `idleSeconds`, `agentSandbox`, `substrate`) | `sandboxes`; the substrate worker and sandbox settings flattened to `worker_*` and `sandbox_*` |
| `prerequisites` | `prerequisites` |
| annotation `stigmer.ai/upgrade-without-backup` | `upgrade_without_backup_to` |

### Refused by the operator, not at validation

Two of the operator's checks need numbers validation cannot compare: a request above its limit, and memory below what a workload needs. They reach the Stigmer's status in the operator's words. A Stigmer in another namespace is refused there too, because the operator serves only the oldest Stigmer on a cluster.

## Pairs well with

- **KubernetesNamespace** for the Stigmer's namespace, which must outlive the install.
- **KubernetesSeaweedFs** (or an AwsS3Bucket) for the bucket, and **KubernetesCertManager** with it whenever the database is bundled: the bucket turns the archive on.
- **KubernetesPostgres** and **KubernetesTemporal** for your own databases, with your own issuer when exposed.
- **KubernetesClusterIssuer** or **KubernetesGateway** for the front door.
- **KubernetesCloudNativePgOperator**, **KubernetesCnpgBarmanCloudPlugin**, **KubernetesAgentSandboxController** and **KubernetesAgentSubstrate**, ordered first with `depends_on`, on clusters that share them.
