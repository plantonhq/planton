# Kubernetes Harbor

## When NOT to Use This

**One resource is ONE Harbor registry** — the CNCF-graduated container
registry that stores, signs, and scans OCI artifacts, from the official
`harbor` chart (1.19.x = Harbor 2.15+), on a data plane you pick.

Not the right kind when:

- **You want a managed registry** — point your clusters at it and
  deploy nothing here (ECR, GCR, Artifact Registry, Docker Hub).
- **You only need a single-node scratch registry** — Harbor's
  cooperating components (core, portal, registry, jobservice, Trivy)
  are the wrong shape for a throwaway `registry:2` Deployment.
- **You want north-south exposure declared inside this resource** —
  exposure composes. This kind always deploys the chart's nginx front
  door behind a ClusterIP, NodePort, or LoadBalancer Service; point a
  catalog exposure kind at the exported `expose_service`.

## The data-plane contract

Exactly one arm per concern:

- **Database** — `internal` (chart PostgreSQL StatefulSet, evaluation
  grade by upstream's own position) XOR `external` (host/username as
  references; password through a Secret whose key is the chart's
  contract key `password` — exactly what a `KubernetesPostgres`
  application Secret exports).
- **Cache** — `internal` XOR `external` Redis (a `KubernetesValkey`
  composes naturally; bridge its username-keyed auth Secret to the
  chart's `REDIS_PASSWORD` contract, or declare the password and let
  the module materialize it).
- **Artifact storage** — `filesystem` XOR S3-compatible (a
  `KubernetesSeaweedFs` endpoint composes; credentials Secret-native
  or keyless/IRSA) XOR GCS XOR Azure.

Credentials never ride rendered Helm values except where the chart
forces it (the internal-database superuser password — the chart's
only intake). Every other site is an `existingSecret` reference.

## Credentials that never ship as defaults

The chart's publicly documented defaults (`Harbor12345` admin
password, `changeit` database password, `not-a-secure-key` encryption
key, `harbor_registry_password`) are public knowledge and never ship.
Unset admin auth = the module generates a random password into
`<name>-admin-auth` (key `HARBOR_ADMIN_PASSWORD`) and exports it as
`admin_password_secret`. Inter-component secrets land in
`<name>-internal-auth` under each chart site's contract key.

## Anonymous visibility is per-project, not per-registry

Harbor's guard is its project visibility model, verified live: a
PUBLIC project's metadata is anonymously listable and its artifacts
anonymously pullable BY DESIGN (the projects API serves unauthenticated
reads of the public subset), while a PRIVATE project rejects anonymous
pulls at the registry surface. Generating a strong admin password does
not make a public project private — audit project visibility, not just
credentials, when hardening a deployment.

## Exposure and `externalUrl`

`externalUrl` is load-bearing for pushes and pulls — Harbor embeds it
in the token-service URL returned to every OCI client. Dial a
different address than `externalUrl` and auth fails. Set it to the
address the composed exposure (or the port-forward for evaluation)
actually serves.

## Trivy and air-gapped clusters

Trivy is on by default. Its first boot downloads the vulnerability DB
from the public internet; in an air-gapped cluster set `skip_update`
and `offline_scan` and preload the DB onto the cache volume yourself —
with `skip_update` and no pre-loaded DB, every scan fails closed.

## Name budget

The chart truncates its fullname at 63 and then APPENDS component
suffixes. The longest (`-jobservice-internal-tls`, 24 characters)
renders whenever internal TLS runs in auto mode, so `metadata.name`
caps at 39 characters — enforced fail-loud in both engines.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
