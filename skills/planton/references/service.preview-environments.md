# Preview Environments: A Real Environment per Pull Request

A pull request on an opted-in service gets its own short-lived **environment** — a real Environment record, not a special kind. It is always a preview OF a base environment ("a preview of dev"): the record carries the base environment's slug, the service, the pull request number, and an expiry instant. Deployed-resource identity on Planton is environment-scoped, so a per-PR environment isolates everything the preview deploys with no renaming anywhere.

## The opt-in

```yaml
spec:
  build:
    triggers:
      pullRequests:
        deploy: true          # implies build; previews per pull request
        previewTtlHours: 168  # optional; how long an untouched preview lives (default 72, max 720)
```

A service that never sets this sees zero new records, zero new surface, zero cost. `build: true` alone builds PR commits without ever minting an environment. `previewTtlHours` tunes the abandoned-PR window per service — every push refreshes the clock, so only a PR nobody touches ages toward teardown; values outside 1–720 hours are refused at save.

## What happens on a pull request

- **Open or push** (opened, synchronize, reopened, ready for review): the build runs, and the platform ensures the PR's preview environment — named `{service}-pr-{number}` (e.g. `storefront-pr-123`) — minting it on the first push and refreshing its expiry on every later one.
- **The base environment** is where the PR's target branch deploys: a branch mapped to an environment previews that environment; a PR against a trigger branch previews the first environment of the promotion walk.
- **The run DEPLOYS into the preview** when the service's kustomize tree authors a `previews/<base-env>` directory (see the authoring section below): the tree renders at the PR's own commit — a PR that changes deploy configuration previews that change — the platform re-stamps the manifests onto the preview environment, fills the blank slots with per-PR values, and applies. Rollout verification then watches the workload come online and stamps the deployment's URLs, exactly as for any environment — so an agent that authored the PR can verify its own preview before asking for review.
- **The preview's address**: when the base environment declares a serving domain, the preview serves at `{preview-env-slug}.{base-domain}` (e.g. `storefront-pr-123.dev.acme.com`), filled into the same blank-hostname slots normal deploys use. Without a serving domain, the deployment record carries the target's native endpoint (a run.app URL, an ALB hostname) discovered by rollout verification.
- **Configuration stays live**: config references name their environment inside the reference itself, so the base environment's variables and secrets serve the preview for its whole life — nothing is copied, and a rotated value is what the preview reads next. The previewed service runs against the base environment's stable neighbors; only the changed service deploys into the preview.
- **The pull login follows the workload**: the preview deploy fills `spec.pod.imageRegistries` from the registry connection exactly as a normal deploy does, so the workload's own image-pull Secret lands in the preview namespace with it; a `KubernetesSecret` or `KubernetesExternalSecret` declared beside the workload with a blank namespace is filled into the preview namespace too. Teardown removes them with the preview (`references/service.pulling-private-images.md`).
- **Close tears it down** (merged or not): the preview's infra components are destroyed first, then its records — its deployment receipts and the environment record itself — while the pull request's builds stay browsable in the run history. A preview whose expiry passes with no close (a laptop closed on a Friday) is torn down by a scheduled sweep, so an abandoned pull request never leaks cloud spend.
- **A push racing the teardown builds without a preview**, with the reason on the run; the next push after the teardown completes mints a fresh preview.

## Authoring the previews tree (kustomize services)

Preview topology is the TEAM's, never the platform's: a `previews/<env>` directory in the `_kustomize` tree — the exact sibling of the local-dev flavors under `dev/<flavor>` — stacks on the base environment's overlay and patches exactly what the team's topology needs:

```
_kustomize/
├── base/
├── overlays/
│   └── dev/                  # env: dev, the base environment's overlay
└── previews/
    └── dev/
        ├── kustomization.yaml   # resources: [../../overlays/dev], patches: [service.yaml]
        └── service.yaml         # the deltas previews of dev deploy with
```

Two teams, two topologies, zero platform opinion:

- **Isolated previews**: patch the workload's `spec.namespace` to blank (and drop `createNamespace`) in the previews tree — the platform fills each preview's namespace with the preview environment's own name and creates it, so every PR is fully isolated and destroyed wholesale.
- **Shared namespace**: leave the namespace as the overlay authored it — all previews land beside the base environment's workloads, coexisting through the workload kind's own version track (the deploy pipeline stamps `spec.version` from the PR branch, so multiple tracks of one app share a namespace with disjoint workloads).

The platform's contribution stays mechanical: manifests rendered from the base overlay legitimately carry the BASE environment's name, and the platform re-stamps exactly that onto the preview (any OTHER environment name still refuses — a manifest naming a third environment is a lie); blank namespaces and blank hostnames fill with per-PR values; authored values are never touched.

Check a tree renders before pushing: `planton service env check --preview dev` (or `pull`/`run`) renders `previews/dev` locally. Per-PR fills happen only on a real preview run, so blank slots render blank locally — that is expected.

A service with the opt-in but NO previews tree still builds every PR; its deploy skips naming the authoring path. Inline-configured services (no kustomize tree) do not have preview deploys yet — the run says so plainly.

## Previews on ECS: the shared-ALB recipe

ECS has no namespace concept and needs none: deployed-resource identity is environment-scoped, so the service's own manifests deployed into `checkout-api-pr-88` are NEW resources beside dev's — same names, different environment, zero renaming. The recipe is the service's overlay following two authoring laws, plus a one-file previews tree.

**The durable arrangement** (authored once, outside the service's tree): the environment's shared ALB, its HTTPS listener carrying a wildcard certificate (`*.dev.acme.com`), one wildcard DNS record in the zone pointing at the ALB, and the ECS cluster. The service's `overlays/dev` declares its own four resources — task definition, ECS service, target group, and listener rule (trimmed here to the wiring the recipe teaches; each kind's reference documents the full spec):

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsEcsTaskDefinition
metadata: {name: checkout-api, env: dev}
spec:
  region: us-west-2
  containers:
    - name: app
      image: ""                                  # blank — the artifact slot every deploy fills
      portMappings: [{containerPort: 8080}]
---
apiVersion: aws.planton.dev/v1
kind: AwsEcsService
metadata: {name: checkout-api, env: dev}
spec:
  region: us-west-2
  clusterArn:
    valueFrom: {name: dev-cluster, env: dev}     # SHARED infrastructure: environment named
  taskDefinition:
    valueFrom: {name: checkout-api}              # OWN resource: env absent — rebinds per environment
  loadBalancers:
    - targetGroupArn:
        valueFrom: {name: checkout-api}          # OWN resource
      containerName: app
      containerPort: 8080
---
apiVersion: aws.planton.dev/v1
kind: AwsLbTargetGroup
metadata: {name: checkout-api, env: dev}
spec:
  region: us-west-2
  targetType: ip                                 # awsvpc tasks register by IP
  port: 8080
  protocol: HTTP
  vpcId:
    valueFrom: {name: dev-vpc, env: dev}         # SHARED infrastructure
---
apiVersion: aws.planton.dev/v1
kind: AwsLbListenerRule
metadata: {name: checkout-api, env: dev}
spec:
  region: us-west-2
  listenerArn:
    valueFrom: {name: dev-alb-https, env: dev}   # SHARED infrastructure
  conditions:
    - hostHeader: {}                             # blank — the hostname slot every deploy fills
  actions:
    - type: forward
      forward:
        targetGroups:
          - arn:
              valueFrom: {name: checkout-api}    # OWN resource
```

**The two authoring laws** that make this overlay preview-ready as it stands:

1. **Blank slots serve both lanes.** The blank container image and the blank host-header condition are filled at every deploy — the durable lane fills dev's artifact and `checkout-api.dev.acme.com`, a preview run fills the PR's artifact and `checkout-api-pr-88.dev.acme.com`. Nothing is ever stored filled in git, and an uninjected blank refuses at apply naming the field — blankness is authoring intent, never an accident that ships.
2. **Own resources reference each other with env absent; shared infrastructure names its environment.** A reference with no `env` binds to the environment the manifest deploys into — exactly right for the service's own target group and task definition, which must rebind inside each preview. The cluster, the listener, and the VPC live in the durable environment, so their references say `env: dev` (or use literal ARNs) and keep pointing there from every preview.

**The previews tree is then one file** — nothing to patch:

```
previews/
└── dev/
    └── kustomization.yaml    # resources: [../../overlays/dev]
```

One optional patch is worth knowing: `AwsEcsService.spec.forceDelete: true` skips ECS's scale-to-zero dance on destroy — appropriate for ephemeral environments, so previews tear down faster while the durable overlay keeps the safer default.

**What a pull request then does**: the platform re-stamps `env: dev → checkout-api-pr-88` on the four manifests, injects the PR's image into the task definition, and fills the rule's blank host with `checkout-api-pr-88.dev.acme.com`. Four fresh resources deploy beside dev's, riding the SHARED ALB — the preview hostname is exactly one label under the base domain, so the listener's wildcard certificate and the zone's wildcard record cover every preview with zero per-PR TLS or DNS work. Listener-rule priority can stay unset: AWS appends after the highest, and host-match rules on disjoint hostnames never shadow each other. Rollout verification watches the ECS service's own deployment state (the workload check), and the domain check probes the filled hostname. Closing the PR destroys the four preview resources; the shared ALB never notices.

One honest limit: ECS exports no native endpoint of any kind — a preview whose base environment declares NO serving domain still deploys and verifies through the workload check, but discovers no URL. Give the base environment a serving domain and every preview gets an address for free.

## Reading a preview run's deploy state

The run record's deployment stage carries exactly one of these explanations:

| The record says | It means | The working path |
|---|---|---|
| "enable build.triggers.pullRequests.deploy…" | The service never opted in | Set the flag if the user wants previews |
| "no preview environment was born for this run…" | The flag is on, but the mint was refused — the per-service concurrency cap (previews of other open PRs), a missing base environment, or the environment name is already taken by a non-preview environment | Close a PR to free a preview slot, declare deploy environments, or rename the colliding environment |
| "…no preview configuration for base environment…" | The preview exists but the repository authors no `previews/<base-env>` tree | Author the tree (the skip names the exact path) |
| "The previews tree for base environment '…' failed to render…" | The tree exists and kustomize refused it — carried with the kustomize error, failing only the PR's own run (a broken previews tree never fails a main-branch deploy) | Fix the tree; `planton service env check --preview <env>` reproduces locally |
| "…this service's configuration is inline…" | Preview deploys ride the kustomize previews tree; the inline arm does not exist yet | Eject to kustomize to author preview configuration |
| "Preview environment '…' no longer exists (it was torn down…)" | A close or the TTL sweep raced this run | Push again for a fresh preview |

A preview outcome never fails the BUILD: a capped-out or refused preview degrades the run to build-only, with the reason in the delivery log.

## Inspecting previews: one call answers everything

`list_service_previews` (CLI: `planton service previews <service>`) is THE read for a pull request's preview — one composed answer per pull request: the preview environment's life facts (base environment, expiry), the latest pull-request run's outcome with skips and failures explained verbatim, and the standing deployment's URLs and rollout verdict. Never join environment and deployment reads yourself; this read exists so the answer is computed once, honestly.

The `phase` field says where the preview stands, in the user's words:

| Phase | It means |
|---|---|
| `preview_not_enabled` | The service never opted in — the reason names the flag to set |
| `preview_not_born` | The mint was refused (cap, missing base environment, name collision) or nothing has happened yet — the reason explains |
| `preview_building` / `preview_deploying` | The pull request's latest run is in flight |
| `preview_build_failed` / `preview_deploy_skipped` / `preview_deploy_failed` | The latest run's honest outcome, reason verbatim (the deploy-state table above) |
| `preview_live` | Deployed — read `rollout_status` (verified / failed / unverifiable) and `urls` for the addresses |
| `preview_tearing_down` | Close or expiry started the teardown; the record disappears shortly |
| `preview_torn_down` | It lived and died; a new push mints a fresh preview |

Ask without a pull request number to list a service's live previews; ask with one to get that pull request's whole story — including `preview_not_born` and `preview_torn_down`, which have no environment record to show and therefore never appear in the plain list. `deployed_commit_sha` beside `latest_commit_sha` tells you when the URLs still answer with an older commit while a newer push builds.

**The agent journey this enables**: an agent that authored a pull request verifies its own preview before asking for review — push, poll `list_service_previews` with the PR number until the phase settles, confirm `rollout_status` is `verified`, then hand reviewers the working URL (paste it into the PR description). No baseline platform gives the authoring agent this loop.

**The pull request page tells the same story by itself.** A "Preview" check on the PR's head commit carries the composed answer — in progress while building or deploying, success with the live address when the preview is up, failure with the run's own reason, neutral with the refusal when a preview could not be born — and a GitHub Deployment named by the preview's slug puts the native "View deployment" button beside it. When the preview tears down (close or expiry), its GitHub deployment goes `inactive`. Nothing posts PR comments, ever. A PR page showing neither surface on an opted-in service usually means the GitHub App installation lacks the 'Checks'/'Deployments' write permissions.

`list_environments` and `get_environment` (CLI: the environment read commands) still show previews beside durable environments; the `spec.preview` block is the tell — base environment, service, pull request number, expiry. The block is server-managed: create, update, and apply refuse it, so no manifest can disguise a durable environment as a preview or convert one. Previews never appear in promotion order — promotion walks are derived from each service's own declared environments.

## A preview has exactly one writer: its pull request

Every door except the pull request's own pushes refuses a preview, each naming the working path:

- **Delete**: `delete_environment` (CLI: `planton env delete`) REFUSES a preview, marked or not — the direct delete removes records only, which would orphan the preview's infra components. A preview dies only through the platform's own teardown: close the pull request, or let the expiry pass. If a user asks to remove a preview, close its pull request — that IS the delete button.
- **The delivery verbs**: promote, rollback, and deploy all refuse a preview target — a preview deploys only from its own pull request's pushes. Promoting FROM a preview's deployment refuses too: merge the pull request to take the change to the durable environments, or deploy its image directly.
- **Ordinary pushes**: a declared deploy entry, overlay directory, or branch mapping that happens to name a live preview's slug never deploys there — the run's environment entry fails carrying this same truth.
