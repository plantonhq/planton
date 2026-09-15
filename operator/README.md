# Planton Operator

A Kubernetes operator that turns one `PlantonPlatform` resource into a complete, running Planton platform on any Kubernetes cluster, and keeps it converged.

## What It Does

The operator carries every deployment decision a self-hosted Planton needs -- which components exist, the order they come up in, their readiness gates, the credentials wired between them, and the first-boot seeding -- so an adopter declares a platform release and nothing else. It installs the prerequisite sub-operators (CloudNativePG, its Barman Cloud backup plugin when cert-manager is present, Tekton Pipelines), provisions the data services (PostgreSQL, Valkey, Temporal, the OpenFGA policy engine, the bundled OpenBAO secrets manager; Neo4j when opted in), renders the control plane, console, in-cluster runner, and identity server, and reports one phase and one plain-language message per platform in `kubectl get plantonplatform`.

Exactly one operator runs per cluster. A second installation refuses itself at startup, names the first, and says what to do; a cluster may run as many platforms as it likes under that one operator.

## Installing It

Adopters install the operator through its Helm chart, `helm/planton-operator` in this repository, published as an OCI chart. The chart and the operator image share one version line: chart `x.y.z` deploys operator image `vx.y.z`, both cut from the same tag. The chart also owns the two custom resource definitions the operator reconciles (`PlantonPlatform`, `PlantonIdentityProvider`): they are release resources, upgraded with the chart and kept on uninstall by default. The chart's README documents the values, the definition lifecycle, and the upgrade paths.

The same two steps exist as catalog kinds (`KubernetesPlantonOperator`, `KubernetesPlantonPlatform`) so an agent or a pipeline installs and upgrades a self-hosted Planton through OpenTofu or Pulumi from published modules. See `site/public/docs/self-hosting/` for the adopter-facing walk-through.

## Declaring a Platform

```yaml
apiVersion: planton.ai/v1
kind: PlantonPlatform
metadata:
  name: planton
  namespace: planton
spec:
  version: v0.0.45
```

`spec.version` names a Planton platform release as `vMAJOR.MINOR.PATCH`; the API server refuses any other shape. The operator runs releases from a floor upward: a version older than the oldest it supports is refused before anything is created, with the reason in the resource's `MESSAGE` column and a `VersionSupported` condition, and a platform already running is left untouched. The operator's first log line (`Platform version floor`) names the floor. To run a custom build, keep `spec.version` at a release and set `image.tag` on the component: the version names the contract, the tag names the bytes.

`spec.email` tells the platform how to send mail through the adopter's own provider, and one declaration powers both of the install's senders. It carries a sender identity (`from.address`, `from.name`, an optional `replyTo`) and exactly one provider: an SMTP relay (`host`, `port`, `security` as `starttls` | `tls` | `none`, and one way in -- a `kubernetes.io/basic-auth` Secret named by `credentialsSecretName`, an OAuth 2.0 client-credentials grant under `oauth2` for tenants that no longer accept passwords, or no credential for an allow-listed relay -- plus an optional private CA bundle by Secret reference) or a Resend account (`apiKeySecretRef`). Admission refuses a contradictory block in words: two providers or none, two ways in, credentials over a plaintext connection. The operator carries the declaration to the control plane as environment plus one projected volume of credential files that rotate in place (no secret value ever rides as an environment variable), preflights every referenced Secret and key -- a missing one renders the control plane as if no email were declared and reports the component not Ready with the Secret, its type, and the field to remove named -- and echoes what is declared in `status.email` and the `EMAIL` column (`NotConfigured` | `SMTP` | `Resend`; configuration, never a delivery verdict). The same declaration owns the bundled identity server's mail settings by key, with the credential rotated through a fingerprint and the relay CA in its truststore, so "Forgot password?" is on the sign-in page exactly when an email can be sent; with nothing declared the realm carries no relay and reset stays off, and a relay typed into the identity server's admin console is reverted on the next reconcile. With nothing declared the control plane is also handed two setup hints -- the Secret command and the `spec.email` fragment for this platform's own name and namespace -- which its Email settings page shows in place of a dead button.

`spec.database.postgresql.backup` tells the platform where its own database is backed up: an object store the adopter owns (`objectStore` with `destinationPath` and exactly one of `s3`, `gcs`, `azureBlob`, or `r2`; the cloud's keyless identity through `serviceAccountAnnotations` where the cloud has one, or a credentials Secret named by `credentialsSecretName` -- R2 always names one), a six-field cron `schedule` for base backups (`0 0 2 * * *` by default; the first base backup runs the moment backups are declared), and a `retentionPolicy` the store enforces (`30d` by default). Declaring it installs CloudNativePG's backup engine, the Barman Cloud plugin, when the cluster lacks it (`spec.prerequisites.postgresBackupPlugin`, `auto` by default, needs cert-manager), turns on continuous WAL archiving, and files the archive under a server name unique to this platform (`status.backup.serverName`). On a fresh install the database is created only once the plugin serves and the credentials Secret exists, so it is born archiving rather than restarted later to attach the archive -- and if the plugin itself cannot be installed (an apply the cluster refuses, a webhook down), the database is held rather than created empty, because with a restore declared an empty database would stand where the archive's copy was expected; on a running platform a missing Secret is `Failing` in words, an uninstallable plugin is `Unavailable` in words, and the database is left as it is. The base-backup schedule is created only once the database operator reports the plugin loaded on the Cluster, so its immediate first run meets a live archiving sidecar instead of failing against one that is still attaching. `status.backup` and the `BACKUP` column say `NotConfigured`, `Deploying`, `Unavailable`, `Healthy`, or `Failing` in the plugin's own words, with the first recoverability point and the last successful base backup read from the archive's own record (the plugin's `ObjectStore` status, per server name); a failing backup never takes a working platform out of `Ready`. `spec.database.postgresql.recoverFrom` declares a new platform restored from such an archive (the same `objectStore`, the source's `serverName`, an optional RFC 3339 `targetTime`), honored only when the database is first created and explained on a running one; the restored platform archives under a new server name and never writes over its source, and `status.backup.restoredFrom` names the archive it came from for as long as the database lives. A restored realm carries its source's identity users and passwords, so existing people sign in -- and its master admin credential, which this install's freshly generated bootstrap admin Secret does not match; the operator re-establishes it exactly once through Keycloak's own recovery command (a one-shot Job before the identity server starts, then the real admin reset and the recovery admin removed once it answers), recorded in a ConfigMap, each step named in the identity component's status. The identity component's Ready sentence on such a realm says so -- existing users sign in as before, no setup code applies -- instead of the first-run journey. The plugin install is re-applied on every pass its controller stands unready, so an apply refused on one object (a missing verb, a webhook down) heals the moment the cause is removed and is reported in `status.backup` until then. The archive carries the bundled secrets manager too: the vault stores its data in the platform's PostgreSQL -- its own database, its own least-privilege role, both declared to CloudNativePG and reconciled by it on a fresh cluster and a restored one alike -- so every connection credential, managed secret, and signing key rides the same WAL stream as the records and comes back to the same instant. A restore is whole only if the restored vault can be opened, so a backup requires the vault's keys to outlive the platform (`spec.vault.autoUnseal`, a key in the adopter's cloud, or `spec.vault.initSecretName`, a Secret the adopter owns); the definition refuses a backup with neither, and `status.backup.vault` says whether the archive covers the vault, which seal opens it, and which Secret to keep a copy of. A vault that comes back from an archive initialized and sealed while its keys Secret does not exist is refused in words that name the Secret, its keys, and the two ways out.

`spec.vault` is the bundled secrets manager (OpenBAO), deployed by default and integral the way the database is: it holds every pasted connection credential, every managed secret, the license signing key, and the OIDC issuer's signing key, and it keeps all of it in the platform's PostgreSQL, so the archive above carries the vault with the records. What opens the vault is decided when it is created and never changes after. `autoUnseal` names exactly one seal in the adopter's cloud -- `gcpKms`, `awsKms`, `azureKeyVault`, or a central OpenBao's `transit` engine -- keyless through the vault's own ServiceAccount identity where the cloud has one (`serviceAccountAnnotations`: GKE Workload Identity, IRSA, Azure Workload Identity) and otherwise through a Secret the declaration names (`credentialsSecretName`, keyed by the environment variable the seal's wrapper reads); nothing credential-bearing ever enters the vault's configuration, which the chart renders into a ConfigMap. The seal is checked when the server starts, before the vault can be initialized: the identity must be able to describe the key as well as encrypt and decrypt with it (on Google Cloud that is `roles/cloudkms.viewer` beside `roles/cloudkms.cryptoKeyEncrypterDecrypter`), and a key it cannot reach crash-loops the vault with `Error configuring seal` in its log -- the component's status names that check when it sees the crash-loop. Under a cloud seal the vault opens itself on every start, including the start after a restore on a cluster that has never seen it; the operator never unseals it, and a vault that stays sealed is reported with the decrypt grant its identity lacks. Unset, the vault uses the built-in key shares and the operator unseals it with the shares it wrote at initialization. `initSecretName` names a Secret the adopter owns where the operator writes the vault's keys at initialization -- unseal keys under the built-in seal, recovery keys under a cloud seal (they never unseal anything; they authorize the break-glass quorum) -- and the root token. The operator writes into it once, without an owner reference, and never deletes it, so deleting the `PlantonPlatform` leaves it standing; a namespace the declaring resource owns is deleted with that resource and takes every Secret in it, so the copy kept outside the cluster is what a restore into a new cluster needs. A Secret that already holds a vault's keys is never written over: a fresh vault meeting one is refused in words, because those keys open some other vault's data. Unset, the operator keeps the keys in a Secret it owns (`<platform>-openbao-init`), deleted with the platform. A declaration that names a different seal than the one the vault was initialized under is refused before anything renders: there is no seal migration, and the server would refuse to start against its own storage.

`config/samples/` holds a minimal declaration, a lite profile for Kind, and a full profile that exercises every optional arm.

## How It Reconciles

The controller in `internal/controller/` is a thin orchestrator. Each reconcile fetches the resource, initializes status when the spec's shape changed, judges `spec.version` against the operator's floor (`internal/platformversion/`; a refused version ends the reconcile without a requeue and touches no component), then walks every registered component in `internal/component/`: a component whose dependencies are not yet Ready is marked Pending with the dependency named, the rest reconcile their own resources and report Ready, Deploying, or Error. One status write per reconcile; a 30-second requeue for ongoing convergence.

A component backed by a Deployment is Ready only when its rollout has finished -- the spec observed, every desired replica on the current template, no stale replica remaining, every replica available -- so a version change never reads as Ready while the previous release still serves.

A bundled component's data belongs in the platform's database when the platform already archives it. The vault is the case: a store of its own on a volume of its own would need a second archive, a second schedule, and a second restore whose point in time could never match the records'; stored in the one PostgreSQL, it rides the one archive and the one `recoverFrom`, and its only need beyond that is the least-privilege role and database the database component declares for it. A component that cannot create its own database is born with one the same way (OpenFGA's through `postInitSQL`; the vault's through a CloudNativePG `Database` object, the shape the rest converge on).

What a declaration cannot honor is refused before it renders, never discovered as a pod that cannot start: a seal whose credentials Secret is missing its key, a seal that differs from the one recorded on the vault's init Secret at initialization, a keys Secret that already belongs to another vault. Each is a sentence with the object named and the way out.

A component that is not Ready explains itself. Every component's status carries a phase, a one-word `reason`, the `object` the reason is about, a sentence, and when that condition began; the platform's `Ready` condition -- the `MESSAGE` column of `kubectl get plantonplatform` -- speaks for the worst-off component in its own words, prefixed by its key. The not-ready branch of every component goes through one helper (`Base.NotReady` in `internal/component/workload_pending.go`) that reads the workload's own pods, the volume claims those pods reference, the workload's conditions, and the namespace's Events, and classifies them in order of certainty: a claim that will never provision, then what the kubelet states outright (a pull failure, a missing Secret key, an out-of-memory kill, a crash loop), then the scheduler's and the Deployment controller's verdicts, and last the calm "running, not yet answering its health check". The reasons are documented constants (`api/v1/component_reason.go`), not an enum, because an operator is upgraded before the platform and a new reason must never be refused by an older definition; failure reasons become a Warning Event on the platform when a component enters them, once, and a Normal one when it recovers. A component's `Refused(...)` result is the third shape: the declaration cannot be honored as written, and nothing deploys until the named field changes.

`PlantonIdentityProvider` has no controller of its own: a change to one re-enqueues the platforms in its namespace, and the identity component resolves the binding inside the same loop.

## The Front Door

A platform is reached through one public origin, rendered from one route table (`internal/resources/front_door_routes.go`) onto whichever door the declaration chooses: an Ingress object, a Gateway API HTTPRoute attached to a Gateway the cluster already runs, or the built-in nginx gateway served over `kubectl port-forward`. The browser API (gRPC-Web) lives under `/rpc`, the storage relay under `/storage`, the identity server under `/idp`, the keyless identity issuer's two discovery documents (`/.well-known/openid-configuration`, `/.well-known/jwks.json`) and inbound webhooks (`/webhooks`) reach the control plane's webhook port, and the console answers everything else.

The front door is the platform's keyless identity issuer, and it declares whether the public internet reaches it (`spec.ingress.reachability`: `auto` | `public` | `private`; `auto` reads a hostname served over HTTPS as public). The operator resolves that once, publishes the answer in `status.reachability` beside the URL, and renders every internet-facing posture of the control plane from it: the issuer URL is the door; keyless cloud connections are offered when the door is public, HTTPS, and the platform vault runs, and otherwise the connection wizard's keyless card names the one fact that closed the door; GitHub webhooks are delivered to `<door>/webhooks/github` when the door is public; and the door is the browser console, so the control plane hands out the door's own pages as the addresses a customer's GitHub App is configured with (its Setup URL and Callback URL) instead of any hosted address. No self-hosted install carries Planton's own GitHub App or its Google and Microsoft OAuth apps, so those doors are declared closed with a reason, never inferred from placeholder credentials.

The install's GitHub declaration (`spec.github`: the hosts a company uses, an install-wide GitHub App per host, whether each host can deliver webhooks) reaches the control plane the way the identity-federation facts do -- one facts file in a ConfigMap the control-plane component rewrites every pass and the Deployment mounts whole (`internal/resources/control_plane_github.go`), so a corrected declaration is live without a pod roll -- and the App keys ride the email credentials' shape, projected PEM files under one directory. Every App Secret is preflighted (`internal/component/github.go`); a missing one is `Refused` in words and that host is advertised without its App. The facts JSON is a cross-repo contract pinned by `testdata/github-facts.json` on both sides. Nothing declared is the github.com adopter's posture, stated out loud in the same file.

Native gRPC clients -- the `planton` CLI, the in-cluster runner, any grpc-go or grpc-java program -- reach the control plane's raw gRPC port at the same origin on the Gateway API door: a request whose `content-type` is `application/grpc` is routed there by an exact header match, a core Gateway API rule. The Ingress object and the nginx gateway have no portable header match, so they do not offer that row; on those doors a native client port-forwards the control plane Service's `grpc` port. The console publishes the address a native client should dial (`GRPC_ENDPOINT`, set only when the door routes it) and the deployment shape (`PLANTON_DEPLOYMENT_KIND`) in its device discovery document, so a client learns what an instance is from the instance itself.

## Quick Start (development)

```bash
# Create a Kind cluster for local development
make kind-create

# Install the CRDs into the cluster
make install

# Run the operator from your host against Kind's API server
make run

# In another terminal, declare a platform and watch it converge
kubectl apply -f config/samples/minimal.yaml
kubectl get plantonplatform -w
```

`make kind-deploy` builds the image, loads it into Kind, and deploys the operator in-cluster instead; `make kind-e2e-lite` does the whole lite journey in one command; `make kind-status` and `make kind-logs` read the result.

## Development

| Command | Description |
| --- | --- |
| `make manifests` | Regenerate the CRDs and the manager ClusterRole into `config/` AND into the Helm chart (the chart derives both; CI fails a stale chart) |
| `make generate` | Regenerate DeepCopy methods |
| `make test` | Unit tests, envtest, and the chart render test |
| `make test-e2e` | The Kind e2e suite on a dedicated cluster (`setup-test-e2e` / `cleanup-test-e2e` manage it) |
| `make test-chart-lifecycle` | The chart lifecycle suite on its own Kind cluster: fresh install, keep, reinstall, keep off, both upgrade paths from the last published charts |
| `make test-realm-convergence` | The Keycloak realm-convergence and federation suite (needs Docker) |
| `make lint` / `make lint-fix` | golangci-lint |
| `make docker-buildx` | Multi-platform image build |
| `make generate-manifests` | Refresh the embedded third-party manifests and chart archives the operator renders at runtime |

Run `make help` for every target with its description.

## Package Map

```
cmd/main.go                      Manager entry: floor logged, singleton guard, controller registration
api/v1/                          The PlantonPlatform and PlantonIdentityProvider types (README)
internal/controller/             The reconcile loop (README)
internal/component/              One file per platform component; the readiness rules
internal/resources/              Renders the Kubernetes objects each component owns (README)
internal/status/                 Status, conditions, and the version refusal (README)
internal/platformversion/        The platform-version floor and its verdicts
internal/singleton/              The one-operator-per-cluster guard
internal/bootstrap/              First-boot bootstrapping the operator performs for the platform
internal/keycloak/               Realm and client convergence for the bundled identity server; re-establishing its master admin on a restored realm (recovery.go)
internal/keycloaklogintheme/     The sign-in and email themes served by the identity server (README)
config/                          Kubebuilder scaffolding: generated CRDs and RBAC, samples, kustomize
hack/                            Generators (the chart's CRD templates) and the lab directory fixture
test/chart                       The chart render test (inside make test)
test/chartlifecycle              The chart lifecycle suite (Kind)
test/e2e                         The kubebuilder e2e suite (Kind)
```

Packages marked (README) carry their own design notes.

## Related

- [The operator Helm chart](../helm/planton-operator/README.md): values, CRD lifecycle, upgrade paths.
- [The platform Helm chart](../helm/planton/README.md): declares a platform against a running operator.
- [Self-hosting Planton](../site/public/docs/self-hosting/index.md): the adopter-facing install and upgrade guide.
- [Repository architecture](../architecture/README.md): where the operator sits in the open-source repository.
