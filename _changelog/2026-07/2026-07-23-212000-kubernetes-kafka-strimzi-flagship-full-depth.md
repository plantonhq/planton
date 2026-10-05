# Kubernetes Kafka family rebuilt at full depth on Strimzi 1.1.0 (KRaft), with first-class topics and users

## What changed

- **KubernetesStrimziKafkaOperator** rebuilt on the served
  `strimzi-kafka-operator` chart 1.1.0 (the KRaft-only Strimzi line):
  typed watch scoping, reconciliation timing, log level, feature gates,
  operand policy-generation toggles, second-install RBAC posture
  (`create_global_resources`), air-gap image source, scheduling, and the
  `helm_values` escape hatch. The `crds/`-directory CRD posture is
  documented on every surface that reads it: uninstall never deletes
  CRDs or Kafka clusters, and chart upgrades never upgrade CRDs — apply
  a new release's CRDs per its release notes. Watched namespaces must
  already exist (verified live: the chart templates RoleBindings into
  them and the install fails otherwise) — now on the field comment.
- **KubernetesKafka** rebuilt from the retired ZooKeeper-era container
  design onto the `kafka.strimzi.io/v1` `Kafka` + `KafkaNodePool`
  resources: KRaft node pools (controller/broker/dual roles;
  persistent/ephemeral/jbod storage with StorageClass references and
  the KRaft metadata marker; per-pool sizing and scheduling), the full
  listener surface (internal, cluster-ip, nodeport, loadbalancer,
  ingress, route, tlsroute; TLS/SCRAM-SHA-512/custom authentication;
  per-listener configuration including the cert-manager server-cert
  seam and the bootstrap/broker annotation surfaces where cloud LB and
  external-dns recipes ride), broker configuration with the
  operator-owned prefixes documented, simple/custom authorization with
  the super-user lockout warning, default-on entity operators, Cruise
  Control with auto-rebalance modes, Kafka Exporter, a module-owned JMX
  Prometheus rules ConfigMap (byte-identical across engines), CA
  policy, rack awareness, JVM heap, and maintenance windows. The
  embedded ingress block and the bundled Confluent Schema Registry and
  Kowl UI containers are gone — the schema-registry and console roles
  move to permissively-licensed first-class kinds.
- **KubernetesKafkaTopic (new)** — declarative topics reconciled by the
  cluster's topic operator, with the placement contract (the cluster's
  own namespace, the `strimzi.io/cluster` binding) on the exact fields
  manifests are written with, partition-growth and
  replication-vs-broker-count semantics documented.
- **KubernetesKafkaUser (new)** — declarative client identities: SCRAM/
  mTLS/external-TLS authentication, the full ACL vocabulary, client
  quotas, and the operator-generated credentials-Secret handle exported
  for workload composition.
- Enum band: the Kafka family holds 907–915 (911–915 reserved for the
  Kafka ecosystem kinds); the Elastic/Altinity/Solr/Neo4j/RookCeph
  entries shifted to 916–924 (zero-adoption mechanical renumber).
- The stale crd2pulumi `strimzioperator` types package (0.42-era) is
  deleted and its generation target retired: the Strimzi CRDs'
  free-typed config objects cannot be carried by generated types, so
  the family's Pulumi modules render untyped custom resources in
  byte-lockstep with the Terraform locals.

## Why

Kafka is the highest-demand streaming workload in the catalog and the
shipped kinds predated the program bar and the KRaft era. The rebuild
also removes source-available-licensed bundled components (Confluent
Community License / BSL families) in line with the licensing
invariants; the streaming role is served end-to-end by Apache-2.0
software (Kafka, Strimzi).

## Verification

- Spec suites for all four kinds (every validation rule locked with
  accepting and rejecting cases); per-kind and entrypoint builds;
  repo-wide Bazel gate; secret-coverage; FK reference guard; outputs
  conformance (+4 cases); import-map conformance; offline tofu plan
  proofs (full-surface and optionals-absent shapes ×4) with
  type-fidelity spot-checks; 18 scenario/fixture manifests + 13 presets
  validated with the working-tree CLI.
- Live kind-cluster E2E, BOTH engines, full six-phase runner: operator
  2×2; kafka 3×2 including the stream-durability proof (acks=all
  markers produced through the bootstrap Service onto an RF-3/min-ISR-2
  topic, a broker deleted, every marker consumed during the outage,
  full-strength re-read after recovery); topic 2×2 including the
  reconcile proof (declared partition count matched via
  kafka-topics --describe); user 2×2 including the SCRAM wire-auth
  proof (produce+consume as the declared user through the scram
  listener under its ACLs). Blind import round-trips for all four kinds
  (nine lanes, including a live three-node cluster re-import). Zero
  orphaned resources.
