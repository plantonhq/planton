# KubernetesHarbor rebuilt: composition-first registry on the official chart, secret-native end to end

## What changed

- **KubernetesHarbor rebuilt on the official `harbor` chart 1.19.1
  (Harbor 2.15.1)** from a years-stale design that bundled plaintext
  database/Redis credentials, embedded a hardcoded ingress, and modeled
  surfaces Harbor itself removed upstream (Notary left at v2.6; Clair
  was replaced by Trivy). The spec now models the chart's real surface:
  per-component topology (core, portal, registry, jobservice, Trivy,
  exporter, the nginx front door), internal TLS with a cert-manager
  seam, metrics/ServiceMonitor, image mirroring, the outbound proxy,
  and an escape hatch whose `fullnameOverride` is re-pinned after the
  merge.

- **The data plane composes**: database internal-XOR-external (the
  external arm takes host/username as references and a password Secret
  under the chart's contract key `password` — exactly what a
  KubernetesPostgres application Secret exports), cache
  internal-XOR-external (a KubernetesValkey composes; the
  username-keyed-vs-REDIS_PASSWORD bridge is taught on the field),
  artifact storage filesystem XOR S3-compatible (a KubernetesSeaweedFs
  endpoint composes; keyless/IRSA supported by omitting credentials;
  `disable_redirect` taught for in-cluster stores) XOR GCS XOR Azure.
  Registry replicas on filesystem-RWO are fenced at validation time.

- **Credentials are secret-native and never the chart's public
  defaults** (`Harbor12345`, `changeit`, `not-a-secure-key`,
  `harbor_registry_password`): unset admin auth generates a random
  password into `<name>-admin-auth` (key `HARBOR_ADMIN_PASSWORD`,
  exported as the credential handle); six inter-component credentials
  land in `<name>-internal-auth` under each chart site's contract key,
  with a STABLE bcrypt htpasswd line (the chart's own helper re-salts
  every render, which would rotate the registry credential on every
  apply). Only Secret NAMES render into Helm values; generation-shape
  arguments carry `ignore_changes` in both engines so imports never
  plan a credential rotation.

- **Exposure composes**: the chart's ClusterIP/NodePort/LoadBalancer
  Service surface plus the REQUIRED `externalUrl` — taught as
  load-bearing (Harbor embeds it in the token-service URL returned to
  every OCI client, so `docker login/push/pull` fail auth when the
  dialed address disagrees). The chart's ingress and Gateway API route
  types are deliberately not modeled; north-south exposure rides the
  catalog's exposure kinds against the exported front-door Service.

- **Trivy is the first-class scanner** with honest air-gap knobs:
  `skip_update`/`offline_scan` taught with the fail-closed truth (no
  pre-loaded vulnerability DB means every scan fails).

- **A dedicated OCI-native E2E verifier** (go-containerregistry): login
  as the generated admin (fails loud if the chart's public default ever
  ships), create a project, push and pull a digest-verified artifact,
  and assert the unauthenticated 401 auth gate first; the durability
  arm proves blobs survive a UID-verified registry pod replacement
  through the PersistentVolume. Scenarios, a CloudNativePG-composed
  full-surface lane, a conformance-green import map (including the
  password-by-value recipes for every module-generated credential and
  the chart-created internal-database Secret), presets, docs, and the
  catalog page all rebuilt to match.

- The `metadata.name` budget is enforced fail-loud in both engines at
  39 characters (the chart truncates its fullname at 63 and appends up
  to 24 characters of component suffix).
