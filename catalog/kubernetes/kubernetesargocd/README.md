# Kubernetes Argo CD

## When NOT to Use This

**One resource is ONE Argo CD control plane** — the declarative GitOps
engine that keeps a cluster converged on what Git says.

Not the right kind when:

- **You want to declare the apps it delivers** — Applications,
  AppProjects and ApplicationSets are plain custom resources once the
  control plane runs; declare them via `KubernetesManifest`, a chart, or
  Argo CD's own UI/CLI. This kind installs the engine, never your apps.
- **You need workflow/pipeline execution** — that is
  `KubernetesArgoWorkflows`. Argo CD converges state; it does not run
  jobs.
- **You expect a public endpoint out of the box** — everything is
  ClusterIP; exposure composes from `KubernetesIngress` or Gateway kinds
  over the exported `server_service` handle, with `server.insecure` set
  when the edge terminates TLS.

## First login

Argo CD itself generates the admin password at first start into the
fixed-name Secret `argocd-initial-admin-secret` (key `password`) —
exported as an output handle. The name is fixed by the application, so
ONE generated-password instance runs per namespace. Disable the local
admin (`admin_enabled: false`) only AFTER SSO works.

## Credentials never ride the manifest

The OIDC client secret and dex connector secrets use Argo CD's own
`$<secret-name>:<key>` runtime indirection against Secrets labeled
`app.kubernetes.io/part-of: argocd`; private-repo credentials are
Secrets labeled `argocd.argoproj.io/secret-type: repository` composed
from `KubernetesSecret`/`KubernetesExternalSecret`. The spec's
`repositories` list registers PUBLIC repositories only.

## The Redis arm

Argo CD's Redis is a disposable cache — losing it costs a re-sync,
never state. Empty = the bundled single pod; `ha` = the three-node
Sentinel subchart (needs three schedulable workers — its anti-affinity
is required); `external` = a managed endpoint or a `KubernetesValkey`
resource by reference.

## CRDs

The chart templates the Application/AppProject/ApplicationSet CRDs with
`crds.keep` defaulting true: destroying this resource leaves the CRDs
(and every Application in the cluster) intact. Turning `keep` off
cascades the delete to all of them — throwaway clusters only.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
