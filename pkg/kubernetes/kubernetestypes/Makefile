go_build_dir=pkg

#https://github.com/kubernetes-sigs/gateway-api/tree/main/config/crd/standard
#https://github.com/kubernetes-sigs/gateway-api/tree/v1.6.1/config/crd/standard
#Note: TCPRoute and UDPRoute graduated to GA (v1) in the v1.6 release and are
#served from the standard channel; the experimental channel is no longer needed
#for any generated type below.
gateway_apis=v1.6.1

#https://github.com/istio/istio/tree/master/manifests/charts/base/crds
#https://github.com/istio/istio/tree/1.30.3
istio_release=1.30.3

#https://github.com/cert-manager/cert-manager/tree/master/deploy/crds
#https://github.com/cert-manager/cert-manager/tree/release-1.15
cert_manager_release=v1.20.3

#https://github.com/apache/solr-operator/tree/v0.9.1/helm/solr-operator/crds
#https://github.com/apache/solr-operator/releases/tag/v0.9.1
#Note: the served chart index (https://solr.apache.org/charts) pairs chart
#0.9.1 with app v0.9.1 — chart versions carry no `v` prefix, the operator/CRD
#artifacts do.
solr_operator_release=v0.9.1

#https://github.com/strimzi/strimzi-kafka-operator/tree/1.1.0/install/cluster-operator
#Strimzi 1.1.0 (KRaft-only; CRD API v1). No generated types: the Kafka and
#KafkaTopic CRDs type their `config` objects (and the listener/Cruise Control
#configuration blocks) with x-kubernetes-preserve-unknown-fields, which
#crd2pulumi cannot carry faithfully — the Kafka-family modules render their
#CRs through apiextensions.CustomResource with untyped args instead (see the
#kuberneteskafka Pulumi module). The pin is recorded for the CRD schemas the
#specs and modules are designed against.
strimzi_operator_release=1.1.0

#https://github.com/cloudnative-pg/cloudnative-pg/tree/v1.30.0/config/crd/bases
#Note: the served operator chart (https://cloudnative-pg.github.io/charts)
#pairs chart 0.29.0 with app 1.30.0 — the chart-index pin governs, never a
#vendored Chart.yaml.
cloudnative_pg_release=v1.30.0

#https://github.com/cloudnative-pg/plugin-barman-cloud/tree/v0.13.0/config/crd/bases
#The Barman Cloud CNPG-I plugin serves the ObjectStore CRD that
#KubernetesPostgres backup blocks render.
barman_cloud_plugin_release=v0.13.0

#https://github.com/external-secrets/external-secrets/blob/v2.8.0/deploy/crds/bundle.yaml
#https://github.com/external-secrets/external-secrets/tree/v2.8.0
#Note: the bundle carries the served external-secrets.io/v1 versions plus
#unserved v1beta1 blocks and the generators.external-secrets.io v1alpha1 CRDs.
external_secrets_release=v2.8.0

#https://github.com/opensearch-project/opensearch-k8s-operator/tree/v2.8.0/opensearch-operator/config/crd/bases
#The stable operator app release. The served chart pairing is 2.8.0 = app
#2.8.0 (later served charts default to a PRERELEASE manager image and bundle
#next-generation opensearch.org CRDs the stable operator does not serve —
#verified against the served charts; the chart index governs, never a
#vendored Chart.yaml). Only the OpenSearchCluster CRD is generated: it is
#the one CR the catalog declares (KubernetesOpenSearch); the operator's
#declarative satellite CRs (users, roles, ISM policies, templates) are not
#admitted kinds.
opensearch_operator_release=v2.8.0

# keycloak: no generated types by design — the Keycloak CR's truststores
# field is a map of nested objects (the crd2pulumi carriability gap), so
# the KubernetesKeycloak modules render the CR untyped in byte-lockstep
# with the Terraform twin. The operator bundle is pinned in the module.

#https://github.com/Altinity/clickhouse-operator/tree/master/deploy/operator
#https://github.com/Altinity/clickhouse-operator/tree/release-0.27.2
#Note: upstream tags are `release-X.Y.Z` — the plain `X.Y.Z` tag form no
#longer exists, and a wrong ref fails SILENTLY through the curl pipeline
#(empty CRD file, stale generated types). Keep the full tag name here.
altinity_operator_release=release-0.27.2

#https://github.com/aws/karpenter-provider-aws/tree/v1.14.0/pkg/apis/crds
#Note: the karpenter/karpenter-crd Helm charts are OCI-served
#(public.ecr.aws/karpenter); the repo-tag vendored Chart.yaml lags the
#served chart, so the release pin below is verified against the ECR
#registry API, never the vendored chart version.
karpenter_release=v1.14.0

#https://github.com/rabbitmq/cluster-operator/tree/v2.22.3/config/crd/bases
#The operator has NO Helm chart; the official distribution is the released
#single-file manifest (releases/download/<tag>/cluster-operator.yml), which
#the KubernetesRabbitMqOperator module renders per document. The CRD pin
#below is the source the KubernetesRabbitMq spec and generated types are
#designed against. The released manifest requires cert-manager: it ships
#Mutating/Validating webhooks (failurePolicy: Fail) whose serving
#certificate is a cert-manager.io Certificate with CA injection.
rabbitmq_cluster_operator_release=v2.22.3

.PHONY: clean
clean:
	rm -rf ${go_build_dir}

.PHONY: gen-gateway-apis
gen-gateway-apis:
	crd2pulumi --force --goPath=gatewayapis https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_gatewayclasses.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_gateways.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_grpcroutes.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_httproutes.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_referencegrants.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_tlsroutes.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_tcproutes.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_udproutes.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_listenersets.yaml \
	https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${gateway_apis}/config/crd/standard/gateway.networking.k8s.io_backendtlspolicies.yaml

.PHONY: gen-karpenter
gen-karpenter:
	crd2pulumi --force --goPath=karpenter https://raw.githubusercontent.com/aws/karpenter-provider-aws/${karpenter_release}/pkg/apis/crds/karpenter.sh_nodepools.yaml \
		https://raw.githubusercontent.com/aws/karpenter-provider-aws/${karpenter_release}/pkg/apis/crds/karpenter.k8s.aws_ec2nodeclasses.yaml

.PHONY: gen-istio
gen-istio:
	crd2pulumi --force --goPath=istio https://raw.githubusercontent.com/istio/istio/${istio_release}/manifests/charts/base/files/crd-all.gen.yaml

# cert-manager renamed its CRD files from crd-<kind>.yaml to
# <group>_<plural>.yaml starting with the v1.16 release line — keep the file
# names below in sync with the pinned release's deploy/crds/ directory.
.PHONY: gen-cert-manager
gen-cert-manager:
	crd2pulumi --force --goPath=certmanager https://raw.githubusercontent.com/cert-manager/cert-manager/${cert_manager_release}/deploy/crds/cert-manager.io_certificates.yaml \
		https://raw.githubusercontent.com/cert-manager/cert-manager/${cert_manager_release}/deploy/crds/cert-manager.io_certificaterequests.yaml \
		https://raw.githubusercontent.com/cert-manager/cert-manager/${cert_manager_release}/deploy/crds/acme.cert-manager.io_challenges.yaml \
		https://raw.githubusercontent.com/cert-manager/cert-manager/${cert_manager_release}/deploy/crds/cert-manager.io_clusterissuers.yaml \
		https://raw.githubusercontent.com/cert-manager/cert-manager/${cert_manager_release}/deploy/crds/cert-manager.io_issuers.yaml \
		https://raw.githubusercontent.com/cert-manager/cert-manager/${cert_manager_release}/deploy/crds/acme.cert-manager.io_orders.yaml

.PHONY: gen-solr-operator
gen-solr-operator:
	crd2pulumi --force --goPath=solroperator https://raw.githubusercontent.com/apache/solr-operator/${solr_operator_release}/helm/solr-operator/crds/crds.yaml

#gen-strimzi-operator deliberately retired: the Strimzi CRDs' free-typed
#config objects (x-kubernetes-preserve-unknown-fields) cannot be carried by
#crd2pulumi's generated types, so the Kafka-family Pulumi modules render
#untyped CRs and no generated package exists to regenerate. The pin above
#records which CRD schemas the specs were designed against.

# Clusters, ScheduledBackups and Backups from the operator; the ObjectStore
# CRD from the Barman Cloud plugin (backups are plugin-based — the operator's
# in-tree object-store support is deprecated upstream). One goPath: the four
# types are consumed together by the KubernetesPostgres Pulumi module (the
# on-demand Backup is the base backup every archive series starts from).
.PHONY: gen-cloudnative-pg
gen-cloudnative-pg:
	crd2pulumi --force --goPath=cloudnativepg https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/${cloudnative_pg_release}/config/crd/bases/postgresql.cnpg.io_clusters.yaml \
		https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/${cloudnative_pg_release}/config/crd/bases/postgresql.cnpg.io_scheduledbackups.yaml \
		https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/${cloudnative_pg_release}/config/crd/bases/postgresql.cnpg.io_backups.yaml \
		https://raw.githubusercontent.com/cloudnative-pg/plugin-barman-cloud/${barman_cloud_plugin_release}/config/crd/bases/barmancloud.cnpg.io_objectstores.yaml

.PHONY: gen-external-secrets
gen-external-secrets:
	crd2pulumi --force --goPath=externalsecrets https://raw.githubusercontent.com/external-secrets/external-secrets/${external_secrets_release}/deploy/crds/bundle.yaml

.PHONY: gen-opensearch-operator
gen-opensearch-operator:
	crd2pulumi --force --goPath=opensearchoperator https://raw.githubusercontent.com/opensearch-project/opensearch-k8s-operator/${opensearch_operator_release}/opensearch-operator/config/crd/bases/opensearch.opster.io_opensearchclusters.yaml

.PHONY: gen-altinity-operator
gen-altinity-operator:
	@echo "Extracting CRDs from Altinity bundle..."
	@curl -s https://raw.githubusercontent.com/Altinity/clickhouse-operator/${altinity_operator_release}/deploy/operator/clickhouse-operator-install-bundle.yaml | \
		awk '/^kind: CustomResourceDefinition$$/,/^---$$/' | \
		sed '/name: clickhouseoperatorconfigurations.clickhouse.altinity.com/,/^---$$/d' > /tmp/altinity-crds-only.yaml
	crd2pulumi --force --goPath=altinityoperator /tmp/altinity-crds-only.yaml
	@echo "Generated Altinity operator types (excluding ClickHouseOperatorConfiguration)"

# No generated types for the Percona PXC / PSMDB CRDs, deliberately: both
# schemas type their backup `storages` as a map of NESTED objects, which
# crd2pulumi flattens to map[string]map[string]string — a generated type
# that structurally cannot carry the nested s3/gcs/azure storage bodies.
# The KubernetesMysql and KubernetesMongodb Pulumi modules render those CRs
# through apiextensions.CustomResource with untyped args instead (the same
# posture as the cert-manager issuer kinds), so the generated packages
# would have no consumer.

.PHONY: gen-rabbitmq-cluster-operator
gen-rabbitmq-cluster-operator:
	crd2pulumi --force --goPath=rabbitmqoperator https://raw.githubusercontent.com/rabbitmq/cluster-operator/${rabbitmq_cluster_operator_release}/config/crd/bases/rabbitmq.com_rabbitmqclusters.yaml

.PHONY: gen
gen: clean gen-gateway-apis gen-karpenter gen-istio gen-cert-manager gen-cloudnative-pg gen-solr-operator gen-external-secrets gen-opensearch-operator gen-altinity-operator gen-rabbitmq-cluster-operator

.PHONY: build
build: clean gen
