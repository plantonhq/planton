# KubernetesStigmer

Installs [Stigmer](https://github.com/stigmer/stigmer), the open-source AI agent platform, on a Kubernetes cluster: the Stigmer operator, from its `stigmer-operator` Helm chart (`oci://ghcr.io/stigmer/charts`), and the one `stigmer.ai/v1 Stigmer` resource the operator reconciles into a running install. One block declares the version, the model key, the database, the bucket, the front door, sign-in and sandboxes; one destroy removes it and everything the operator put on the cluster, and keeps the encryption keys.

## Why one kind

Stigmer runs one operator and one Stigmer per cluster, and the operator's version is Stigmer's version: the resource has no version of its own. So the version is declared, upgraded and removed in the one place the install is declared. The operator's Helm release is always `stigmer-operator` (its cluster-wide role and the definition it keeps are named after the release), and the Stigmer is always named `stigmer` (every object the operator creates is named after it).

## What it installs

In order, on both engines:

1. **The operator's namespace** (`operator.namespace`, `stigmer-system` by convention), created only with `operator.create_namespace`.
2. **The `stigmer-operator` chart** at `version`. The chart's values are the typed `operator` settings; the chart's image tag is never set, because the chart runs the operator at its own version. The release waits for the operator to be Available.
3. **The Secrets** the spec gives values for, in the Stigmer's namespace, under the key names the operator reads, each holding only the keys declared: the model key (`ANTHROPIC_API_KEY`, `ANTHROPIC_BASE_URL`, `OPENAI_API_KEY`, `ANTHROPIC_FOUNDRY_API_KEY`, `AWS_BEARER_TOKEN_BEDROCK`), the bucket's key pair (`ACCESS_KEY_ID`, `SECRET_ACCESS_KEY`), the database password, Temporal's API key and certificates, the SMTP login, the license, and secret extra settings. Every credential can instead name a Secret you keep.
4. **The restore check**: before a restore (`database.bundled.recover_from`), the encryption keys the archive is sealed with must be in the namespace. Without them the operator would mint new keys and the restored records could not be read, so the deploy stops and names the step that fixes it.
5. **The Stigmer**, rendered from only the keys the spec declares, so the operator's own defaults decide everything else.

The operator then installs what the declaration needs and is missing (CloudNativePG, its Barman Cloud backup plugin, agent-sandbox, Agent Substrate; `prerequisites` says which it may), and runs the server, the shared runner, Postgres, Temporal, the front door and sign-in. Deploy does not wait for Stigmer to be Ready; `status_command` shows its phase and the one sentence that explains it.

## Before anything is applied

The plan refuses, in a sentence that says what it saw, what it means and what to do:

- a Stigmer namespace that does not exist (this kind never creates or deletes it);
- an operator namespace that is missing, someone else's when asked to create it, or created here when no longer asked to;
- creating the operator's namespace in the Stigmer's own namespace (destroy would delete the keys);
- a Stigmer definition another Helm release owns: another operator install, or this one moved to another namespace in place;
- a Stigmer in the namespace that this resource did not declare;
- a `version` below the one the server runs: Stigmer's server refuses to start on a database a newer release migrated, so going back is a restore, not a downgrade;
- CloudNativePG's definitions, or its backup plugin's, left on the cluster by a Helm release that no longer runs a controller (the catalog's CloudNativePG kinds keep their definitions when destroyed). Stigmer's operator would trust them and wait forever for a database nothing creates; the refusal names the command that clears them.

`planton validate` refuses every contradiction Stigmer's own definition refuses, before any deploy, with the kind's own field paths: your own Postgres without your own Temporal, the bundled sign-in without a hostname or an owner, Enterprise without your own issuer, a license, or a per-session sandbox driver, backups or sandboxes without a bucket, runner sizes under a per-session sandbox driver (which runs no shared runner), and more.

## What a destroy leaves

Destroy deletes the Stigmer first. Then it waits, bounded at ten minutes and reporting every half minute, for the operator's own cleanup: this Stigmer's database and its pods, the sandbox namespace, and each prerequisite the operator installed. Only then are the Secrets and the chart removed. If the operator cannot finish (it has no ready replica, or the wait runs out), the destroy goes on and prints what stays and the commands that remove it; a destroy is never blocked.

Three things outlive a destroy on purpose:

- **`stigmer-keys`**, the encryption keys. Every stored secret and every backup is sealed with them. Save them with the `keys_export_command` output before you delete the namespace or the cluster; a restore into a new cluster applies that file first.
- **The bucket's archive.**
- **The `stigmers.stigmer.ai` definition**, which the chart keeps. A later install under the same release adopts it.

With the `substrate` sandbox driver, the operator's cleanup does not finish inside the wait. Stigmer's operator does not end a live session's Substrate actor when the Stigmer goes, so that actor's worker drains for up to an hour, the operator's cleanup waits behind it, and the destroy goes on and names what stays: CloudNativePG, the backup plugin, Agent Substrate and the sandbox namespace. Remove them with the commands it prints once the worker has stopped.

Deleting the Stigmer's namespace, including through an environment teardown that removes its KubernetesNamespace, deletes the keys. Run `keys_export_command` first.

## Upgrades

Move `version` to a newer release. With backups on, the operator backs the database up before it rolls the server, and holds the upgrade while that backup fails; `upgrade_without_backup_to`, set to the same value as `version`, releases that one upgrade without a backup. Going back below the version the server runs is refused; restore a backup taken under the older version into a new install instead. The operator cannot move to another namespace in place: destroy, delete the kept definition once no Stigmer is left, and deploy again.

## Composition

| Field | Composes with |
|---|---|
| `namespace`, `operator.namespace` | KubernetesNamespace |
| `database.external.host` | KubernetesPostgres (its read-write Service) |
| `database.external.password.existing_secret` | KubernetesPostgres (its application Secret, in the same namespace) |
| `temporal.external.host_port` | KubernetesTemporal (its frontend endpoint) |
| `object_store.s3.endpoint` | KubernetesSeaweedFs |
| `object_store.s3.bucket` | AwsS3Bucket |
| `ingress.tls.issuer.name` | KubernetesClusterIssuer |
| `ingress.gateway` | KubernetesGateway |

When the catalog's CloudNativePG, its backup plugin, agent-sandbox or Agent Substrate run on the cluster, order them first with a `depends_on` relationship (preset `07-composed-prerequisites`): Stigmer's operator leaves a copy someone else installed alone, but if it installed its own first, the catalog's resource would find it and refuse. A backup plugin someone else runs must be plugin v0.16.0 or newer, the release Stigmer bundles and KubernetesCnpgBarmanCloudPlugin's default (`chart_version` 0.9.0): with a named S3 endpoint, Stigmer writes object-store settings older releases do not declare, and the database's archive is refused in the Stigmer's status.

## Accepted limits

- **Your own Postgres with the bundled sign-in does not work.** The bundled sign-in keeps its accounts on the bundled database. With your own Postgres, sign people in through your own issuer (`identity.oidc`); its realm, the two clients, the audience, the owner's account and sign-up are yours to set up.
- **Some credentials are given again.** A KubernetesPostgres's password Secret can be named directly when the database runs in the Stigmer's namespace; a SeaweedFS's key pair cannot (Stigmer reads `ACCESS_KEY_ID` and `SECRET_ACCESS_KEY`, and SeaweedFS writes other names), so give it again from a Planton secret.
- **A restore into a brand-new cluster** needs the keys file saved with `keys_export_command`, applied before the deploy; the restore check names this step, it does not automate it.
- **Requests above limits** and memory below what a workload needs are refused by the operator in the Stigmer's status, not at validation.
- **A Stigmer in another namespace** is refused by Stigmer's own operator, which serves only the oldest Stigmer on a cluster, in that Stigmer's status.
- **A leftover prerequisite installed without Helm** (CloudNativePG or agent-sandbox applied from upstream's release files and later removed, leaving their definitions) names no controller the plan can look for, so it is trusted, as Stigmer's operator trusts it, and the install waits on it. Delete the leftover definitions first, once no object of theirs is left.
- **A destroy under the `substrate` driver leaves the operator's prerequisites behind** (see "What a destroy leaves"): Stigmer's operator does not end live sessions' Substrate actors, so their workers drain for up to an hour and its cleanup cannot finish inside the destroy's wait.
- **GKE 1.37 has no default StorageClass**: name one in the volume settings.
- **Stigmer's own open limits stay Stigmer's**: Agent Substrate's snapshots cannot use Cloud Storage's S3 endpoint (stigmer/stigmer#3415), a failed Substrate template is not retried (#3417), and switching the driver away from Substrate leaves it running (#3418).

## Outputs

| Output | What it is |
|---|---|
| `namespace`, `operator_namespace`, `operator_release` | Where the install and the operator live |
| `server_endpoint` | `http://stigmer.<namespace>.svc.cluster.local:7234`, for in-cluster clients |
| `artifacts_endpoint` | The files address, port 7235 |
| `url` | `https://<hostname>`, or `http://localhost:7234` through the port-forward |
| `port_forward_command`, `status_command` | Open it from a workstation; read its phase |
| `owner_password_command` | The owner's one-time password, when the bundled sign-in runs |
| `keys_secret`, `keys_export_command` | The keys Secret, and the command that saves it as a re-appliable file |

## Engines

Both engines apply the same documents, held together by a cross-engine test against a real `tofu plan` (`iac/pulumi/module/parity_test.go`). The one difference is where the destroy wait lives: OpenTofu keeps it on a state-only `terraform_data`, Pulumi on a marker ConfigMap whose delete hook runs the same script. Both need `kubectl` on the runner's PATH for the restore check and the wait. Built for 100% parity with Stigmer 0.3.2's resource and chart; the field-by-field account is in [GUIDE.md](GUIDE.md).

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
