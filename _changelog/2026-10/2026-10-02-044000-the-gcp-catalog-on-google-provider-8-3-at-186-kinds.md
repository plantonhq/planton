# The GCP Catalog on Google Provider 8.3, at 186 Kinds

**Date**: October 2, 2026
**Type**: Feature (with breaking changes on `v1alpha1` fields)
**Components**: every GCP kind; 82 new GCP kinds; `shared/catalogkind`; `pkg/providerparity`; `pkg/iac/tofu/generators`; `catalog/gcp/aa_e2e`; `e2e/gcp`

## Summary

**The GCP catalog moves to Google's current provider and more than doubles in reach.** Every GCP module now pins `hashicorp/google ~> 8.3`, and Pulumi pins pulumi-gcp v9.37.0. 82 new kinds join the 104 already shipped, so a team can declare a whole Google Cloud organization from one catalog: the folder tree and its policies, shared networking, private certificate authorities, security controls, data pipelines, the AI platform, and the delivery path. Each kind is built on both engines and checked argument by argument against the pinned provider: **built for 100% Terraform parity**, with live proof still to run.

**Fields that a catalog kind can produce are now links.** Where another kind makes the value (a CA pool, an image, a fleet, a folder, a Cloud Build worker pool, a certificate issuance config), the field takes a typed reference, so a chart deploys things in the right order and a picker offers the right kinds. Sixteen of these fields shipped earlier as plain text. They change type in place, on the `v1alpha1` channel. See **Breaking changes**.

**Secrets have a home wherever a runtime can read one.** Cloud Run (services, jobs, and worker pools), Cloud Functions, Agent Engine, Workflows, and Cloud Composer keep a secret environment value in a Secret Manager secret the block owns, readable only by the runtime's identity. Plain text never reaches the deployed resource. Model Garden deployments and Colab runtime templates say plainly that their values are plain text, because no grant could make a stored secret readable there.

## What Changed

### New kinds (82)

- **Organization and governance:** `GcpFolder`, `GcpOrgPolicy`, `GcpOrgPolicyCustomConstraint`, `GcpTagKey`, `GcpTagValue`, `GcpTagBinding`, `GcpBillingBudget`, `GcpCloudIdentityGroup`.
- **Networking:** `GcpSharedVpcHost`, `GcpSharedVpcServiceProject`, `GcpVpcPeering`, `GcpHaVpnGateway`, `GcpHaVpnConnection` (generates its shared secret and BGP MD5 key unless you supply them), `GcpHierarchicalFirewallPolicy`, `GcpNetworkFirewallPolicy`, `GcpPscServiceAttachment`, `GcpNetworkEndpointGroup`.
- **Certificates:** `GcpPrivateCaPool`, `GcpPrivateCaCertificateAuthority`, `GcpPrivateCaCertificateTemplate`, `GcpPrivateCaCertificate`, `GcpCertManagerTrustConfig`, `GcpCertManagerIssuanceConfig`.
- **Security:** `GcpKmsAutokeyConfig`, `GcpKmsKeyHandle` (accepted by the 20 key fields whose resource types Autokey serves), `GcpSccNotificationConfig`, `GcpSccMuteConfig`, `GcpSccBigQueryExport`, `GcpBinaryAuthorizationPolicy`, `GcpBinaryAuthorizationAttestor`.
- **Access grants for grantees that depend on the resource:** `GcpPubSubTopicIamMember` and `GcpGcsBucketIamMember`. A logging sink's writer identity or a Security Command Center export's identity exists only after the sink or export is created, so granting it from the topic or bucket would be a dependency cycle. These grant kinds break it.
- **Data:** `GcpRedisCluster`, `GcpRedisClusterEndpointSet`, `GcpManagedKafkaCluster`, `GcpManagedKafkaTopic`, `GcpManagedKafkaAcl`, `GcpManagedKafkaConnectCluster`, `GcpManagedKafkaConnector`, `GcpBigQueryConnection`, `GcpBigQueryReservation`, `GcpBigQueryCapacityCommitment`, `GcpBigQueryReservationGroup`, `GcpDatastreamStream`, `GcpDatastreamConnectionProfile`, `GcpDatastreamPrivateConnection`.
- **Serverless:** `GcpCloudRunWorkerPool`.
- **AI:** `GcpVertexAiAgentEngine`, `GcpVertexAiModelGardenDeployment`, `GcpVertexAiRagEngineConfig`, `GcpVectorSearchCollection`, `GcpVertexAiSearchDataStore`, `GcpVertexAiSearchEngine`, `GcpVertexAiSearchDataConnector`, `GcpVertexAiFeatureGroup`, `GcpVertexAiFeatureOnlineStore`, `GcpVertexAiDataset`, `GcpVertexAiTensorboard`, `GcpVertexAiPersistentResource`, `GcpModelArmorTemplate`, `GcpModelArmorFloorSetting`, `GcpDocumentAiProcessor`, `GcpColabRuntimeTemplate`, `GcpColabRuntime`, `GcpColabSchedule`, `GcpTpuVm`, `GcpTpuQueuedResource`, `GcpDialogflowCxAgent`, `GcpDialogflowCxSecuritySettings`.
- **Platform engineering:** `GcpGkeFleet`, `GcpGkeFleetFeature`, `GcpGkeFleetScope`, `GcpGkeFleetMembership`, `GcpComputeImage`, `GcpCloudBuildConnection`, `GcpCloudBuildRepository`, `GcpCloudBuildTrigger`, `GcpCloudBuildWorkerPool`, `GcpDeliveryPipeline`, `GcpDeployTarget`, `GcpDeployPolicy`, `GcpDeployCustomTargetType`.

### Existing kinds

- **Provider 8.3:** every module moved from `~> 7.43` to `~> 8.3`. Where 8.0 changed a default, the modules keep the old behavior: `GcpBackendService` and `GcpGlobalForwardingRule` send `load_balancing_scheme = EXTERNAL` when the spec leaves it unset, because the provider's default moved to `EXTERNAL_MANAGED` and the scheme is immutable. The 8.x arguments are modeled on `GcpCloudRun`, `GcpCloudRunJob`, `GcpCloudSql`, `GcpComputeInstance`, `GcpComputeMig`, `GcpDataprocCluster`, `GcpGkeCluster`, `GcpGkeNodePool`, `GcpMemorystoreInstance`, and `GcpVertexAiNotebook`.
- **Regional load balancing:** `GcpBackendService`, `GcpUrlMap`, `GcpTargetHttpProxy`, `GcpTargetHttpsProxy`, `GcpGlobalForwardingRule`, and `GcpCloudArmorPolicy` build Google's regional resources when `region` is set. Global-only levers are refused on the regional arm. `GcpCloudArmorPolicy.rules[].priority` now accepts 0, the highest priority Google allows.
- **New outputs and fields:** `GcpCloudSqlUser.service_account` (an IAM database user named by its service account), `GcpProject.folder_id`, `GcpGkeCluster.fleet_membership`, `GcpKmsKey.initial_version_name`, `GcpComputeImage.image_id`, and the secret-value arms on the runtimes above.
- **References name what they accept:** `GcpUrlMap`'s backend fields accept `GcpBackendService` or `GcpBackendBucket`; `GcpGlobalForwardingRule.target` accepts either target proxy or a `GcpPscServiceAttachment`, and its `ip_address` accepts either address kind.

### Tooling

- **Every GCP `variables.tf` is generated** from the kind's proto (`planton tofu generate-variables`), formatted, and documented, and the drift test holds it there.
- **google-beta enters only through the admission list** (`pkg/providerparity/admissions/google-beta.yaml`). The TPU VM and the TPU queued resource are the GCP kinds admitted beside the Firebase family, and a guard refuses any unlisted `provider = google-beta`.
- **The parity report is honest about IAM.** Google's per-resource IAM triplets are reported as covered only where a kind's `iam_members` field reaches them (24 covered, 390 not offered per resource).
- **Permission manifests** name only permissions Google defines (checked against the IAM inventory snapshot), and state when a permission is needed as data (`condition.spec_field_set`) wherever "this field is set" can say it.

### Waiting for pulumi-gcp v10

A few 8.x arguments are missing from pulumi-gcp v9.37.0. Each is left out of the spec on both engines, never offered on one engine only, and named in the kind's README, guide, and both module READMEs: `GcpApiKey`'s `check_existing_usage`, `GcpCloudRunWorkerPool`'s `sandbox_launcher` and multi-header probes, `GcpVertexAiAgentEngine`'s build service account and audio-transcription example parts, `GcpManagedKafkaCluster`'s internet access and bootstrap address, and `GcpGkeFleet`'s labels and compliance posture. `google_vertex_ai_rag_corpus` is absent from v9 entirely and joins as a kind with v10.

## Breaking changes

All GCP kinds are on the `v1alpha1` channel, where a field may change type in place. These changes are deliberate, and validation names the fix.

**Sixteen fields became references.** Write the literal as `{value: ...}`, or link the producing kind with `valueFrom`:

| Kind | Field | Default kind it links to |
|---|---|---|
| `GcpComputeDisk` | `spec.image` | `GcpComputeImage` (`status.outputs.self_link`) |
| `GcpComputeInstance` | `spec.bootDisk.image` | `GcpComputeImage` (`status.outputs.self_link`) |
| `GcpComputeMig` | `spec.template.disks[].sourceImage` | `GcpComputeImage` (`status.outputs.self_link`) |
| `GcpGkeNodePool` | `spec.nodeConfig.secondaryBootDisks[].diskImage` | `GcpComputeImage` (`status.outputs.image_id`) |
| `GcpGkeCluster` | `spec.fleetProject` | `GcpGkeFleet` (`status.outputs.project_id`); also accepts `GcpProject` |
| `GcpGkeCluster` | `spec.userManagedKeys.clusterCa`, `.etcdApiCa`, `.etcdPeerCa`, `.aggregationCa` | `GcpPrivateCaPool` (`status.outputs.name`) |
| `GcpCloudSql` | `spec.network.serverCaPool` | `GcpPrivateCaPool` (`status.outputs.name`) |
| `GcpMemorystoreInstance` | `spec.serverCaPool` | `GcpPrivateCaPool` (`status.outputs.name`) |
| `GcpWorkloadIdentityPool` | `spec.inlineCertificateIssuanceConfig.caPools` (each map value) | `GcpPrivateCaPool` (`status.outputs.name`) |
| `GcpLoggingSink` | `spec.scope.folderId` | `GcpFolder` (`status.outputs.folder_id`) |
| `GcpCloudRun` | `spec.buildConfig.workerPool` | `GcpCloudBuildWorkerPool` (`status.outputs.name`) |
| `GcpCloudFunction` | `spec.buildConfig.workerPool` | `GcpCloudBuildWorkerPool` (`status.outputs.name`) |
| `GcpCertManagerCert` | `spec.managed.issuanceConfig` | `GcpCertManagerIssuanceConfig` (`status.outputs.issuance_config_id`) |

```yaml
# Before
spec:
  image: debian-cloud/debian-12
# After: a literal
spec:
  image:
    value: debian-cloud/debian-12
# After: a link
spec:
  image:
    valueFrom:
      kind: GcpComputeImage
      name: golden-base
      fieldPath: status.outputs.self_link
```

A manifest still in the old form is refused with: `got a bare str "debian-cloud/debian-12" -- write as {value: <literal>} or {valueFrom: {kind: GcpComputeImage, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse`.

**Three references that no catalog kind can feed became plain strings:** `GcpDataprocCluster`'s `dataprocMetastoreService` (under `spec.clusterConfig.metastoreConfig` and `spec.virtualClusterConfig.auxiliaryServicesConfig.metastoreConfig`), `GcpTargetHttpsProxy.spec.serverTlsPolicy`, and `GcpKmsKey.spec.cryptoKeyBackend`.

```yaml
# Before
spec:
  serverTlsPolicy:
    value: projects/my-project/locations/global/serverTlsPolicies/mtls
# After
spec:
  serverTlsPolicy: projects/my-project/locations/global/serverTlsPolicies/mtls
```

The old form is refused with `expects a string; got an object`.

## Verification

- Protos through `make protos`, including the Java stub compile and the protovalidate-java CEL conformance gate; every spec test.
- Every module: `tofu fmt`, `init`, and `validate`; Pulumi builds; offline plans with a rendered sweep and refused negatives for every new kind's manifest and scenarios.
- The parity report at 186 of 186 kinds at total accounting; the registry snapshot, containment golden, outputs conformance, secret coverage, reference, permission, cost, control, preset, catalog-page, and logo gates; `make e2e-build` and `make e2e-vet`.
- Live proof: not yet. Every new kind's profile is `pending_proof` (or `deferred`, with the reason, where a lane needs an organization, a billing account, or another owner-arranged fixture, listed in `e2e/README.md`).
