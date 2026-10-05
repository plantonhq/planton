# Kubernetes analytics, vector and messaging tiers live-proven; the import framework learns secret-material IDs

## What changed

- **Five kinds proven against a live cluster, both engines** —
  KubernetesAltinityOperator, KubernetesClickHouse, KubernetesQdrant,
  KubernetesRabbitMqOperator, KubernetesRabbitMq. Every behavioral
  promise ran with verifier-output evidence: ClickHouse served
  replica-synced rows DURING a live replica loss and after recovery;
  Qdrant answered similarity searches correctly after a pod loss;
  RabbitMQ served a quorum-queue message during a broker outage as the
  operator-generated credentials. Blind import round-trips proved the
  two-GVK composed-identity map (ClickHouse + its managed Keeper), the
  16-document multi-GVK release-manifest bundle (the RabbitMQ operator),
  and the Helm-release maps. All five entered the green E2E CI matrix.
  KubernetesSeaweedFs is Terraform-proven (including the volume-loss
  byte-identical durability proof); its Pulumi lanes await a
  provider-pin decision — the chart's templates require Helm >= 3.16
  functions and the pinned Pulumi provider embeds an older Helm.

- **ClickHouse lessons landed where their readers live** (all verified
  live): `ON CLUSTER` DDL requires `GRANT CLUSTER ON *.*` on top of
  database-scoped rights; a user declared WITHOUT grants receives
  ClickHouse's unrestricted config-user default access; a
  grants-constrained user needs `GRANT SELECT ON system.clusters`; and
  the in-server cluster definition can LAG the installation reaching
  Completed — distributed DDL initiated in that window silently executes
  on the visible subset (the install verifier now gates on the
  initiator's converged cluster view, and the docs carry the operational
  recipe). Spec field comments, docs, and the production presets teach
  all four; the sharded-analytics preset's ETL user gained the CLUSTER
  grant its own documented workflow needs.

- **Two Terraform-only rendering defects fixed to engine parity**
  (the operator copies CR templates VERBATIM into core-API objects, so
  template validity is core-API validity — a class plan/preview cannot
  catch): the managed Keeper's pod template rendered a container without
  an image (an explicit container entry suppresses the operator's
  default-image injection; the API rejects the StatefulSet), and the
  client service template rendered without ports (a port-less ClusterIP
  Service is rejected, so the cluster's client Service never appeared).
  Both converged on the Pulumi modules' correct-by-construction shapes,
  with the lesson in both engines' module comments and the kind
  update workflow.

- **The import framework learned secret-material import IDs** — a new
  `from_cluster_secret_key` derivation arm: an import ID that IS a
  secret (the canonical case is `random_password`, imported by the
  password value itself) derives by reading the module-materialized
  Kubernetes Secret, exactly the recipe the map gives a human;
  disconnected contexts fall back to asking. The E2E round-trip binds
  the reader through its cluster credentials and REDACTS the resolved
  ID from progress lines, command echoes, and error wraps. Modules
  declaring a `random_password` now carry `ignore_changes` /
  `IgnoreChanges` on every generation-shape argument (both engines):
  the provider's importer assumes its own generation defaults and every
  argument forces replacement, so without the guard the first plan after
  any import proposes silently rotating the live credential.

- **Verifier robustness** (both live-caught): a deleted StatefulSet
  pod's recovery is only real once a pod with a NEW uid reports Ready
  (the StatefulSet's status lags the deletion), and an S3 GetObject that
  succeeds on headers can still die mid-stream while the serving pod
  goes down — the body read now retries with the request.

- **Output conformance cases** added for the Altinity operator,
  ClickHouse, SeaweedFS and Qdrant naming contracts.

## Validation

22 scenario-engine lanes green on the live cluster (2 blocked on the
provider-pin decision); 13 blind import round-trips green; four
behavioral durability/persistence proofs green with verifier-output
evidence; all touched packages built and vetted; import-map resolver
unit tests green; the green-only E2E CI matrix regenerated with the five
proven kinds; zero orphaned resources (kept CRDs per each kind's
documented keep posture).
