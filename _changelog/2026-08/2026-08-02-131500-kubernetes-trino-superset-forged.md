# Kubernetes analytics query + BI pair forged: Trino and Superset complete the analytics band

## What changed

- **Two new kinds built to full depth on fresh official pins** —
  KubernetesTrino (the distributed SQL query engine, from the official
  trinodb chart 1.42.2 = Trino 480) and KubernetesSuperset (the
  business-intelligence platform, from the official ASF chart 0.22.4 =
  Superset 6.1.0). Both ship the complete anatomy: four protos with
  dense field comments and CEL rules (79 spec tests), both engine
  modules at full parity, presets, docs/catalog pages, conformance-green
  import maps, and a compiled E2E surface (six scenarios, three
  consumer-scoped fixtures, two product-grade verifiers) awaiting live
  proof.

- **Trino is secured by default where upstream ships nothing**: an
  empty `auth` block means PASSWORD (file) authentication with a
  module-generated admin credential — one random feeds both the
  plaintext key and the bcrypt htpasswd line the chart mounts, a
  verified pairing by construction — plus the internal-communication
  shared secret Trino requires once authentication is on
  (`internal-communication.md` at the pin) and the
  `allow-insecure-over-http` pairing for ClusterIP traffic with TLS
  terminating at composed exposure.

- **Trino catalogs are secret-native through Trino's own mechanism**:
  every properties surface in the chart renders into ConfigMaps, so
  catalog passwords ride `${ENV:VAR}` references (the server's secrets
  substitution, verified to work in ALL properties files including
  catalogs) delivered as secretKeyRef env vars. Typed postgres/mysql
  catalog arms FK-compose the catalog's database kinds; the in-image
  tpch/tpcds samples answer SQL on a fresh install; worker autoscaling
  (HPA XOR KEDA), graceful shutdown (termination budget auto-set to
  2× the drain window), fault-tolerant execution with spooled
  exchanges, access-control/resource-group/session documents and JMX
  metrics complete the surface.

- **Superset's design spine is the chart's own environment-Secret
  contract, taken over end to end**: `secretEnv.create=false` turns off
  the chart's credential-bearing Secret and the module composes
  `<name>-env` itself — non-secret connection facts plus generated
  material (the SECRET_KEY Superset refuses to start without, the
  admin password, the websocket JWT), while REFERENCED credentials
  (the composed PostgreSQL metadata-database password, the composed
  Valkey cache password) arrive as `extraEnvRaw` secretKeyRef entries
  (explicit env beats envFrom — the chart's own bring-your-own
  mechanism). No apply-time secret reads exist in either engine.

- **Superset's literal-credential chart paths are never used**: the
  admin bootstrap rides an init-command override reading
  ADMIN_PASSWORD from environment (the chart's `createAdmin` renders
  the password literally into its config Secret and hard-fails its
  template on an empty one — both paths bypassed); the two config
  blocks that render `cache.password` from values (the results backend
  and the async-queries backends) are replaced by module
  configOverrides snippets reading environment. The bundled
  postgresql/redis subcharts (frozen `bitnamilegacy` image lines)
  never ship — the metadata database is external-required and the
  cache external-or-absent, with the Celery components CEL-fenced on
  the cache's presence.

- **The Terraform `variables.tf` files are generator-produced**
  (`planton tofu generate-variables`), eliminating the hand-mirror
  drift class where a silently dropped field renders nothing; the
  forge rule now mandates it. The update rule gained a sharpening of
  the HCL type-unification class proven in this session's plan lanes:
  the jsonencode/jsondecode seam does NOT rescue ternaries (plan-time
  constant folding re-derives concrete types through jsondecode) —
  branch-shaped conditionals decompose into merges of complementary
  single-attribute ternaries.

- **Verifiers are product-grade and ride the strong shared helpers**:
  Trino proves THE AUTH GATE (anonymous statement submission
  rejected), THE QUERY PROOF (`tpch.tiny.nation` = 25 through the
  statement API), THE FEDERATION PROOF (the composed PostgreSQL
  catalog answers AND a cross-catalog join runs in one statement) and
  THE RECOVERY PROOF (a UID-verified coordinator replacement answers
  the same query); Superset proves /health, THE AUTH GATE (anonymous
  read rejected + the documented admin/admin default dead), a real
  security-API sign-in with the JWT+CSRF pairing, THE DASHBOARD PROOF
  (created through the REST API, read back, swept) and THE STATE
  PROOF (the dashboard survives a UID-verified web-pod replacement in
  the composed database).

## Verification

Offline bar complete: 79 spec tests green; 4 OpenTofu plan + 4 Pulumi
preview lanes (full + minimal per kind) with rendered-value
spot-checks and clean leak greps; all seven structural guards;
secret-coverage, validate-refs, importmap conformance
(kubernetes-clean), outputs conformance, containment and kind-map
tests green; `make build-go`; `e2e-build`/`e2e-vet`; 15 manifests
validated; proto-docs, site catalog pages and the E2E matrix
regenerated (matrix byte-identical — both kinds ship `pending_proof`
and enter CI only when proven).
