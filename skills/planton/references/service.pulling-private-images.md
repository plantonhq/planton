# Pulling Private Images — How a Workload Signs In to Its Registry, and What the Deploy Fills for You

Read this when a deploy's pod sits in `ImagePullBackOff` or `ErrImagePull`, when a person asks "how does my private image get pulled" or "do I need a pull secret", when a run's environment row carries a pull sentence, when a GHCR registry is connected through a GitHub sign-in and the service will deploy to Kubernetes, or when the deploy target is Cloud Run, App Runner, or Lambda and the image lives anywhere but that cloud's own registry.

## The doctrine in one sentence

A pull login is a field on the workload's own manifest, or a Secret declared beside it, or nothing at all when the cluster's own identity reaches the registry — never something the platform derives out of sight. The service's registry connection is how builds push; where it holds a login a cluster can KEEP, the deploy writes that login onto the workload for you, as a reference, in the open.

Pushes and pulls have different lifetimes. A build signs in once and is done, so a token minted from a connection the organization trusts (a GitHub sign-in, an AWS/GCP/Azure session) serves it. A cluster signs in every time a pod starts, for as long as the workload runs — it needs a login that outlives the deploy, and Planton never copies a person's sign-in or a minted token into a cluster. Never propose otherwise.

## The three ways a Kubernetes workload pulls — pick one

| Situation | What to declare | Why |
|---|---|---|
| The registry is the cluster's own cloud's registry and the cluster's identity is granted on it: EKS→ECR (node role with `AmazonEC2ContainerRegistryReadOnly`, or IRSA), GKE→Artifact Registry (node service account, `devstorage.read_only`), AKS→ACR (kubelet identity with `AcrPull`), DOKS→DOCR (registry integration on) | **Nothing.** No Secret, no field | The identity already exists; a credential beside it is a liability with no benefit |
| The login belongs to this workload alone (the common case; the deploy fills it from the registry connection) | `spec.pod.imageRegistries: [{server, username, password: $secret/<slug>, email?}]` | The workload module builds ONE `kubernetes.io/dockerconfigjson` Secret named `<workload>-image-pull` in the workload's namespace, created, rolled, and destroyed with the workload; one entry per server (two for the same server are refused) |
| Many workloads share one login; the login is attached on the ServiceAccount the pods run as; or the deploy runs OFFLINE | `spec.pod.imagePullSecrets` naming a `KubernetesSecret` (its `dockerConfigJson` arm, default kind, `spec.name`) or a `KubernetesExternalSecret` (explicit `kind: KubernetesExternalSecret`, `fieldPath: status.outputs.secret_name`; start from its **Docker Registry Pull Secret (Template)** preset) declared beside the workload | The Secret is a resource in the environment; the resource graph deploys it first; the ExternalSecret route works with no Planton backend because the cluster reads the credential from the organization's own secrets manager |

Two laws that never bend: `password` (on `imageRegistries[]` and on the `KubernetesSecret` docker arm) accepts ONLY a `$secret/<slug>` reference — a literal is refused at apply naming the field; and the offline lane (`planton service deploy --env <env>` with no backend) refuses every `$secret/` reference, so `imageRegistries` is a connected-lane arm and the ExternalSecret route is the offline one (`references/service.offline-deploy.md`).

## What the deploy fills, and when it fills nothing

On a Kubernetes target, the deploy reads the service's registry connection (`spec.build.registry`) and decides by the connection's credential ARM alone — never by reading the cluster's cloud:

| The registry connection | The deploy fills on `pod.imageRegistries` |
|---|---|
| GHCR with `pullToken` (either arm) | the pull token's username + `$secret/<its secret>` — preferred over a stored PAT even when both exist |
| GHCR with a stored personal access token, no pull token | the PAT's username + its `$secret/` reference |
| GHCR through a GitHub connection, no pull token | **nothing** — *add a read-only pull token to the registry connection, or declare the login on the workload's imageRegistries* |
| Artifact Registry with a stored service-account key | `_json_key` + the key's `$secret/` reference |
| Artifact Registry through a GCP connection | **nothing** — a GKE cluster pulls with its node service account when granted on the repository; any other cluster needs a stored key or a declared login |
| ACR with a stored service principal | the client id + the client secret's `$secret/` reference |
| ACR through an Azure connection | **nothing** — an AKS cluster pulls with its kubelet identity when it holds `AcrPull`; any other cluster needs a stored principal or a declared login |
| ECR, on ANY arm | **nothing, ever** — *ECR issues only twelve-hour tokens — the cluster pulls with its own AWS identity (the node role or IRSA) when granted on the registry, or through a pull-through cache* |
| JFrog Artifactory | the username + the access token's `$secret/` reference |

The fill yields to exactly one thing: an `imageRegistries` entry the workload already declares for the SAME server. A Secret named in `imagePullSecrets` never suppresses it (the lane cannot see which registry a Secret covers; a redundant login still pulls, a withheld one does not). Entries for other servers are kept and the registry's is appended beside them. The username and the server resolve on the control plane; the password stays a reference the runner resolves inside the cluster's own account. Tell users they do NOT need to author a pull login for their own registry when the connection holds one.

**Where to read what happened**: the run's environment row (`image_pull_posture` on `get_service_pipeline`; the run page's Environments section) — *Pull login for ghcr.io filled from registry connection 'ghcr-acme' (the connection's pull token) — the workload's imageRegistries carries the username and a secret reference the runner resolves inside the cluster's account.* — or the reason nothing was filled. The applied infra component and its IaC input carry the filled `imageRegistries`; the service's own configuration shows the AUTHORED manifests, so the fill is not there. A promote or rollback re-states the posture from the captured bytes: *The applied manifests carry a login for ghcr.io in imageRegistries.*

## The GHCR pull token

A GHCR connection that pushes through a GitHub connection has no long-lived login to give a cluster (App installation tokens last an hour; a person's sign-in is theirs). The connection carries an optional read-only pull token: `spec.githubContainerRegistry.pullToken: {username: {value: <bot account>}, token: {secret: <org secret holding a read:packages PAT>}}` — allowed with either credential arm, preferred over a stored write-capable PAT when both exist. Author it in the connection wizard's **Pull Token** step, the connection page's **Pull Token** section (the page's **How Deploys Pull With This** panel points there when it is missing), the **You Already Trust These** GHCR card's **Add Pull Token**, or `planton apply` on the record. A bot account is the usual owner. Without it, builds still push; a Kubernetes deploy fills no login and the run says so.

## Targets that pull only from their own cloud's registry

These are the providers' rules, not the catalog's, and no field on the target changes them:

- **Cloud Run / Cloud Run Jobs**: private images only from Artifact Registry (public Docker Hub and GHCR images deploy directly). The service wizard warns at authoring time when the registry is anything else; the deploy says it again.
- **App Runner**: ECR or ECR Public only (private ECR needs the access role on the spec).
- **Lambda** container images: ECR in the same account and Region only.
- **ECS**: ECR with the task execution role; any other registry with the Secrets Manager credential the task definition itself declares (`repositoryCredentialsSecretArn`).

The remedy is upstream: push to the provider's registry, or declare a pull-through repository that proxies the registry the image lives in — a `GcpArtifactRegistryRepo` in `REMOTE_REPOSITORY` mode, or `AwsEcrRegistrySettings.pullThroughCacheRules` — and point the image at it.

## The service wizard's dial

The wizard's Environments step, for a Kubernetes target, shows **How It Pulls**: the target's posture from the catalog, the registry connection's arm hint (fills / fills nothing and why), the Cloud Run warning when it applies, and the **Private image?** dial — for an image in a registry OTHER than the service's registry (a sidecar from Quay, say): server, username, and a password picked from organization secrets, never typed. The same login rides every environment. The dial is never needed for the service's own registry when the connection holds a login.

## The sentences to quote back, and the fix for each

- **Refused before any infra job, on `update` / `update_preview` / `state_import` only**: *`<Kind> '<slug>'` references `<TargetKind> '<name>'` at `<fieldPath>` (from `<consumer field>`), which has no value yet — deploy it first or fix the reference.* One line per unresolved reference. `refresh` and state reads never refuse; destroying a workload whose Secret is gone replays the last-applied snapshot and succeeds. Fix: deploy the referenced Secret first (the resource graph does this when both are in the environment), or correct the reference's `name`/`kind`/`fieldPath`.
- **A literal password**: the sensitive-field refusal at apply naming `spec.pod.imageRegistries[].password` (or `spec.dockerConfigJson.password`) and the `$secret/` grammar. Fix: `planton secret set` (or the console's secret picker) and write `$secret/<slug>` (`references/infra.config-references.md`).
- **`ImagePullBackOff` / `ErrImagePull` in the rollout verdict**: the kubelet's own line, then *the cluster cannot pull this image: declare the registry login on the workload (pod.imageRegistries) or a pull secret beside it (pod.imagePullSecrets), or pull from a registry the cluster's own identity reaches.* Diagnose in this order: (1) the run's pull sentence — did the deploy fill a login, and if not, why; (2) the registry connection's arm — a trusted arm without a pull token, or ECR, fills nothing by design; (3) the cluster's identity grant — an EKS node role without the ECR read policy, a GKE node account not granted on the repository, an AKS kubelet without `AcrPull`; (4) the image path itself (`spec.build.imageRepositoryPath`). The fix is the one the sentence names for the case found; a rerun changes nothing until it is made.
- **Offline preflight refuses `$secret/` at `spec.pod.imageRegistries[0].password`**: the connected-lane arm in an offline tree. Fix: move the login to a `KubernetesExternalSecret` (docker-registry template) named in `imagePullSecrets`.
- **Two `imageRegistries` entries name the same server**: *a workload holds one login per registry; merge them into one entry* — refused by the spec before any plan.

## Previews and rotation

A pull-request preview is a real environment: the workload's own `<workload>-image-pull` Secret follows it into the preview namespace with the workload, and a `KubernetesSecret` or `KubernetesExternalSecret` declared beside the workload lands in the preview namespace too (the deploy fills their blank namespace like the workload's). Teardown removes them with the preview. Rotation needs no ceremony: the manifest holds a reference, so the next deploy reads the secret's current value; an ExternalSecret-backed Secret refreshes on its own interval with no deploy at all.

## Never suggest

- A `kubectl create secret docker-registry` outside the manifests, or a hand-edited Secret — the deploy shows everything it creates, and a Secret nobody declared is the failure class this design ends.
- A plaintext password anywhere, including "just for now" — the field refuses it.
- Reading the cluster's cloud to decide whether the deploy should fill — the registry connection's arm decides, and a redundant login still pulls.
- Refreshing GitHub App or sign-in tokens inside the cluster (an operator, a generator, a cron) — Planton does not carry a person's or the App's identity into a cluster; the pull token is the answer.
- Storing a docker-config file on a machine for a module to read — nothing outside the manifest can put a pull credential into a workload.
