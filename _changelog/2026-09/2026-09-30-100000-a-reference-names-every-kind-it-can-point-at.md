# A Reference Names Every Kind It Can Point At, and Its Grain Is Its Type

**Date**: September 30, 2026
**Type**: Feature
**Components**: the foreign-key options (`shared/foreignkey/v1`), `pkg/refannotations`, `pkg/manifestgraph`, `pkg/refcheck`, `pkg/explain`, `pkg/catalogschema`, `pkg/yamldiag`, `pkg/iac/mappingeval`, the OpenTofu generators, `pkg/deferrules`, `pkg/infrachart`; KubernetesExternalSecret, KubernetesKeda, KubernetesMetricsServer, KubernetesNetworkPolicy, KubernetesSecret, AzureFrontDoorOrigin, CloudflareDnsZone, the five AWS database kinds' subnet groups, and nine fields that now declare their candidate kinds

## Summary

**A reference field can now say every kind it accepts, each with the output it composes from.** Before, a field could name one `default_kind`; a field whose value comes from several kinds (a log destination that is a bucket, a log group or a Firehose stream; an address range that is a VPC's, a subnet's or a cluster's) named none, and the kinds it meant lived in comments or in each console's hand-made list. The new `candidate` option declares them, and every reader in the catalog applies one set of rules to them.

Three add-on kinds chose their store or issuer grain with a string or enum beside a one-kind reference, so a cluster-scoped store named by its name was read as the namespaced kind. Their grain is now the arm of a `oneof`, the way the Certificate kind's issuer already was.

## What Changed

- **`(dev.planton.shared.foreignkey.v1.candidate)`** (field option 200004, a repeated `ReferenceCandidate {kind, field_path}`). With `default_kind`, the entries are a field's composition keys:
  - a valueFrom on a kind with exactly one key defaults its `fieldPath` to it;
  - a kind that appears twice (a subnetwork's primary range and its secondary ranges) must name its output;
  - an explicit path must equal one of the kind's keys or extend it with a list index or map key;
  - a kind outside the list is still accepted with an explicit `fieldPath`;
  - when both are declared, the default kind is one of the candidates.
- **One Go reader of reference annotations**, `pkg/refannotations`. The manifest graph's rules and edges, the registry gate, `explain`, the catalog-schema artifact (`foreignKeyCandidates`), the YAML suggestions, the mapping evaluator and the OpenTofu generator read through it. `CheckRef`'s rules carry a case table in `pkg/manifestgraph/rules_test.go` that other implementations copy.
- **The registry gate resolves every declared key**, and refuses a default kind missing from its candidates and reference options on a field that is not a reference.
- **Reference pages list every candidate** (`references: ...`), and each candidate kind's page lists the field under "Referenced by".
- **Candidates declared on:** DigitalOceanProject `resources`, AwsCloudwatchLogDelivery `destination_resource_arn`, AwsWafWebAcl logging `destination_arn`, AwsEventBridgeScheduler target `arn`, GcpTargetHttpsProxy `ssl_certificates`, AwsLbListener `load_balancer_arn`, AzureMonitorScheduledQueryAlert `scope`, AzureBackupProtectedFileShare `source_storage_account_id`, and the fields below.
- **KubernetesNetworkPolicy's `ipBlock.cidr` and `except` take a literal or a reference** to the resource that owns the range: a VPC's or subnet's, a GKE subnetwork's primary or secondary range, a GKE, EKS or AKS cluster's range. The CIDR rule reads the literal. An egress allowlist that carves out a cluster's own ranges no longer repeats addresses the network chart declares.
- **AzureFrontDoorOrigin's `private_link.private_link_target_id`** is a literal ARM ID or a reference to a Linux web app, function app, storage account, Container Apps environment, Application Gateway or Private Link Service. A Private Link Service reference needs no `target_type`, like a literal `/privateLinkServices/` ID.
- **The grain is the arm (breaking, v1alpha1):**
  - KubernetesExternalSecret `storeRef` is `secretStore.name` or `clusterSecretStore.name`;
  - KubernetesKeda `certificates.certManagerIssuer` and KubernetesMetricsServer `tls.certManagerIssuer` use the shared `CertManagerIssuerRef`, `issuer.name` or `clusterIssuer.name`.

  Each arm references its own kind and output. The string and enum selectors are gone.
- **The five subnet-group names are plain strings** (AwsRedshiftCluster, AwsDocumentDb, AwsNeptuneCluster, AwsRdsCluster, AwsRdsInstance): no catalog kind produces a subnet group, so a reference that named no kind was never resolvable.
- **Rules about a literal stay about the literal.** `pkg/deferrules` sets aside a spec rule violation on a value a host resolves later, given the host's own token classifier; `infrachart.Options.IsDeferredToken` wires it into chart validation. KubernetesSecret's `binaryData` rule is plain base64 again, with no reference arm in its pattern.
- **CloudflareDnsZone's inline NS, CNAME, MX and PTR records drop a trailing root dot on OpenTofu**, as the record kind already did, so a nameserver another zone reports settles instead of re-planning.
- **Teaching:** the forge rule states that a reference names its kinds or is plain text, that grain goes in typed arms, and that shape rules are about the literal. The skill's references cover candidate kinds, the arms, rules checked on resolved values, and connection variables that name `<group>/<entry>`. The site's variables page and credentials tutorial describe what connections accept today.

## Verification

- **Live, on kind, both engines:** KubernetesNetworkPolicy (five scenarios), KubernetesExternalSecret (namespaced and cluster store arms), KubernetesMetricsServer and KubernetesKeda, 8 of 8 green.
- **Offline:** `go test` for every touched package and kind; the registry gate, containment registry, preset validity and reference drift; `buf lint`; `defspack`.
- **Not run live:** the CloudflareDnsZone scenario's new root-dot NS record.
