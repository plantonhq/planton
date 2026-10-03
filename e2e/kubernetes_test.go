//go:build e2e

package e2e

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/plantonhq/planton/e2e/framework/discovery"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	profilepkg "github.com/plantonhq/planton/pkg/e2e/profile"
	kindv1 "github.com/plantonhq/planton/qa/catalogkinde2eprofile/v1"
)

// Kubernetes Tier 1 kinds: native K8s resources, zero dependencies.
var kubernetesTier1Kinds = []string{
	"kubernetesnamespace",
	"kubernetesconfigmap",
	"kubernetesdeployment",
	"kubernetesstatefulset",
	"kubernetessecret",
	"kubernetesserviceaccount",
	"kubernetesrbac",
	"kubernetesservice",
	"kubernetescronjob",
	"kubernetesjob",
	"kubernetesdaemonset",
	"kubernetesmanifest",
	"kuberneteshelmrelease",
	"kubernetesexternaldns",
	"kubernetesexternalsecretsoperator",
	"kubernetesclustersecretstore",
	"kubernetessecretstore",
	"kubernetesexternalsecret",
	"kubernetesingressnginx",
	"kubernetesmetricsserver",
	// Gateway API family. The CR kinds declare KubernetesGatewayApiCrds as a
	// registry prerequisite, which the harness installs (standard channel)
	// before applying the route/gateway scenario; verification is
	// controller-free (applies succeed once the CRDs are present).
	"kubernetesgatewayapicrds",
	"kubernetesgatewayclass",
	"kubernetesgateway",
	"kuberneteslistenerset",
	"kuberneteshttproute",
	"kubernetesgrpcroute",
	"kubernetestcproute",
	"kubernetesudproute",
	"kubernetestlsroute",
	"kubernetesreferencegrant",
	// Istio family. KubernetesIstio installs the control plane (istiod + the
	// Istio CRDs, plus cni + ztunnel in ambient mode); KubernetesIstioBaseCrds
	// installs the CRDs-only bundle the seven typed CR kinds declare as a
	// registry prerequisite. The httproute behavioral-routing scenario and the
	// authorizationpolicy behavioral-deny scenario chain KubernetesIstio as a
	// fixture — istiod is the catalog's first in-catalog Gateway API
	// implementation, so live routing and live L7 enforcement are proven here.
	"kubernetesistio",
	"kubernetesistiobasecrds",
	"kubernetespeerauthentication",
	"kubernetesrequestauthentication",
	"kubernetesauthorizationpolicy",
	"kubernetesserviceentry",
	"kubernetesdestinationrule",
	"kubernetesenvoyfilter",
	"kubernetestelemetry",
	// Postgres flagship. KubernetesPostgres declares
	// KubernetesCloudNativePgOperator as a registry prerequisite, which the
	// harness installs before applying the Cluster scenario; the backup and
	// recovery scenarios also declare KubernetesCnpgBarmanCloudPlugin (whose
	// own registry edges bring cert-manager and the operator); the
	// behavioral-failover scenario proves data durability live (write →
	// primary loss → promotion → read-back).
	"kubernetescloudnativepgoperator",
	"kubernetescnpgbarmancloudplugin",
	"kubernetespostgres",
}

// Kubernetes Tier 3 kinds: operator-dependent. Each declares its operator
// as a registry prerequisite (CatalogKindMeta.prerequisites) AND ships an
// explicit e2e/fixtures/ override that pins the operator's exact config; the
// override wins, so the fixture is what actually deploys here. Either way the
// harness installs the operator before the test and tears it down after
// (see e2e/framework/runner/dependencies.go -- ResolveDependencies).
var kubernetesTier3Kinds = []string{
	"kuberneteskafka",
	"kubernetesopensearch",
	"kubernetesmongodb",
	"kubernetesmysql",
	"kubernetessolr",
	"kubernetesclickhouse",
	// SigNoz composes KubernetesClickHouse (registry prerequisite;
	// consumer-scoped fixtures pin the operator's watch scope and the
	// keeper-backed telemetry store) — the data-plane-dependency class,
	// same as the Kafka ecosystem kinds.
	"kubernetessignoz",
	// The TektonConfig declaration KubernetesTektonOperator (its
	// registry prerequisite) reconciles into running kinds.
	"kubernetestekton",
	// A runner fleet KubernetesGhaRunnerScaleSetController (its
	// registry prerequisite) reconciles into a listener and ephemeral
	// runner pods.
	"kubernetesgharunnerscaleset",
	// The Keycloak CR declaration KubernetesKeycloakOperator (registry
	// prerequisite; the consumer-scoped fixture pins the namespaced
	// watch) reconciles into a running server against the composed
	// CloudNativePG database fixture.
	"kuberneteskeycloak",
	// The OpenTelemetryCollector CR declaration KubernetesOtelOperator
	// (its registry prerequisite; cert-manager chains transitively)
	// reconciles into a collector workload per mode.
	"kubernetesotelcollector",
	// Airflow composes KubernetesPostgres (registry prerequisite;
	// transitively the CloudNativePG operator) as its metadata
	// database — the composed-database fixture class, same as
	// Keycloak.
	"kubernetesairflow",
	// The RayCluster CR declaration KubernetesKubeRayOperator (its
	// registry prerequisite) reconciles into head and worker pods;
	// the behavioral lane chains a consumer-scoped Valkey fixture for
	// GCS fault tolerance.
	"kubernetesraycluster",
	// The FlinkDeployment CR declaration KubernetesFlinkOperator (its
	// registry prerequisite; cert-manager chains transitively for the
	// operator's webhook) reconciles into a JobManager and its
	// TaskManagers; the behavioral lane chains a consumer-scoped
	// SeaweedFS fixture for checkpoint/HA storage.
	"kubernetesflinkdeployment",
}

// Kubernetes Tier 4 kinds: operators, addons, and cluster-level
// infrastructure, including operators that are also exercised as Tier 3
// fixtures.
var kubernetesTier4Kinds = []string{
	"kubernetesstrimzikafkaoperator",
	"kubernetesopensearchoperator",
	"kubernetesaltinityoperator",
	"kubernetesgharunnerscalesetcontroller",
	"kubernetestektonoperator",
	"kuberneteskeycloakoperator",
	"kubernetesoteloperator",
	"kuberneteskuberayoperator",
	"kubernetesflinkoperator",
}

// Kubernetes Tier 2 kinds: Helm-based, self-contained chart installs.
var kubernetesTier2Kinds = []string{
	"kubernetesvalkey",
	"kubernetesgrafana",
	"kubernetesopenbao",
	"kubernetesargocd",
	"kubernetesargoworkflows",
	"kuberneteslocust",
	"kubernetesnats",
	"kubernetesneo4j",
	"kubernetesjenkins",
	"kubernetessolroperator",
	"kubernetesperconamongooperator",
	"kubernetesperconamysqloperator",
	"kubernetestemporal",
	"kubernetesseaweedfs",
	"kubernetesqdrant",
	"kubernetesopenfga",
	// An operator whose verifier proves a real workload CR with no
	// fixture prerequisites (the Kyverno/Gatekeeper class).
	"kubernetessparkoperator",
}

// ─── Tier 1 Pulumi ──────────────────────────────────────────────────────────

func TestKubernetesNamespace_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesnamespace", "pulumi")
}
func TestKubernetesConfigMap_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesconfigmap", "pulumi")
}
func TestKubernetesServiceAccount_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesserviceaccount", "pulumi")
}
func TestKubernetesRbac_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesrbac", "pulumi")
}
func TestKubernetesDeployment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesdeployment", "pulumi")
}
func TestKubernetesStatefulSet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesstatefulset", "pulumi")
}
func TestKubernetesSecret_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessecret", "pulumi")
}
func TestKubernetesService_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesservice", "pulumi")
}
func TestKubernetesIngress_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesingress", "pulumi")
}
func TestKubernetesNetworkPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesnetworkpolicy", "pulumi")
}
func TestKubernetesCronJob_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescronjob", "pulumi")
}
func TestKubernetesJob_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesjob", "pulumi")
}
func TestKubernetesDaemonSet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesdaemonset", "pulumi")
}
func TestKubernetesManifest_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmanifest", "pulumi")
}
func TestKubernetesHelmRelease_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteshelmrelease", "pulumi")
}
func TestKubernetesPersistentVolumeClaim_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespersistentvolumeclaim", "pulumi")
}
func TestKubernetesStorageClass_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesstorageclass", "pulumi")
}
func TestKubernetesResourceQuota_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesresourcequota", "pulumi")
}
func TestKubernetesPriorityClass_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespriorityclass", "pulumi")
}
func TestKubernetesPodDisruptionBudget_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespoddisruptionbudget", "pulumi")
}
func TestKubernetesHorizontalPodAutoscaler_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteshorizontalpodautoscaler", "pulumi")
}
func TestKubernetesCertManager_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescertmanager", "pulumi")
}
func TestKubernetesClusterIssuer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesclusterissuer", "pulumi")
}
func TestKubernetesIssuer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesissuer", "pulumi")
}
func TestKubernetesCertificate_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescertificate", "pulumi")
}
func TestKubernetesExternalDns_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesexternaldns", "pulumi")
}
func TestKubernetesExternalSecretsOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesexternalsecretsoperator", "pulumi")
}
func TestKubernetesClusterSecretStore_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesclustersecretstore", "pulumi")
}
func TestKubernetesSecretStore_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessecretstore", "pulumi")
}
func TestKubernetesExternalSecret_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesexternalsecret", "pulumi")
}
func TestKubernetesIngressNginx_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesingressnginx", "pulumi")
}
func TestKubernetesMetricsServer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmetricsserver", "pulumi")
}

func TestKubernetesCilium_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescilium", "pulumi")
}

func TestKubernetesKeda_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskeda", "pulumi")
}

func TestKubernetesClusterAutoscaler_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesclusterautoscaler", "pulumi")
}

func TestKubernetesVelero_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesvelero", "pulumi")
}

// Karpenter's three kinds carry deferred profiles (the controller cannot
// start off AWS), so these entrypoints skip on kind and activate when the
// batched EKS real-cluster lane flips the profiles green.
func TestKubernetesKarpenter_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskarpenter", "pulumi")
}

func TestKubernetesKarpenterNodePool_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskarpenternodepool", "pulumi")
}

func TestKubernetesKarpenterEc2NodeClass_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskarpenterec2nodeclass", "pulumi")
}

// ─── Tier 1 Terraform ───────────────────────────────────────────────────────

func TestKubernetesNamespace_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesnamespace", "terraform")
}
func TestKubernetesConfigMap_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesconfigmap", "terraform")
}
func TestKubernetesServiceAccount_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesserviceaccount", "terraform")
}
func TestKubernetesRbac_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesrbac", "terraform")
}
func TestKubernetesDeployment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesdeployment", "terraform")
}
func TestKubernetesStatefulSet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesstatefulset", "terraform")
}
func TestKubernetesSecret_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessecret", "terraform")
}
func TestKubernetesService_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesservice", "terraform")
}
func TestKubernetesIngress_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesingress", "terraform")
}
func TestKubernetesNetworkPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesnetworkpolicy", "terraform")
}
func TestKubernetesCronJob_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescronjob", "terraform")
}
func TestKubernetesJob_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesjob", "terraform")
}
func TestKubernetesDaemonSet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesdaemonset", "terraform")
}
func TestKubernetesManifest_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmanifest", "terraform")
}
func TestKubernetesHelmRelease_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteshelmrelease", "terraform")
}
func TestKubernetesPersistentVolumeClaim_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespersistentvolumeclaim", "terraform")
}
func TestKubernetesStorageClass_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesstorageclass", "terraform")
}
func TestKubernetesResourceQuota_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesresourcequota", "terraform")
}
func TestKubernetesPriorityClass_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespriorityclass", "terraform")
}
func TestKubernetesPodDisruptionBudget_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespoddisruptionbudget", "terraform")
}
func TestKubernetesHorizontalPodAutoscaler_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteshorizontalpodautoscaler", "terraform")
}
func TestKubernetesCertManager_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescertmanager", "terraform")
}
func TestKubernetesClusterIssuer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesclusterissuer", "terraform")
}
func TestKubernetesIssuer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesissuer", "terraform")
}
func TestKubernetesCertificate_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescertificate", "terraform")
}
func TestKubernetesExternalDns_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesexternaldns", "terraform")
}
func TestKubernetesExternalSecretsOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesexternalsecretsoperator", "terraform")
}
func TestKubernetesClusterSecretStore_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesclustersecretstore", "terraform")
}
func TestKubernetesSecretStore_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessecretstore", "terraform")
}
func TestKubernetesExternalSecret_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesexternalsecret", "terraform")
}
func TestKubernetesIngressNginx_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesingressnginx", "terraform")
}
func TestKubernetesMetricsServer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmetricsserver", "terraform")
}

func TestKubernetesCilium_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescilium", "terraform")
}

func TestKubernetesKeda_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskeda", "terraform")
}

func TestKubernetesClusterAutoscaler_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesclusterautoscaler", "terraform")
}

func TestKubernetesVelero_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesvelero", "terraform")
}

func TestKubernetesKarpenter_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskarpenter", "terraform")
}

func TestKubernetesKarpenterNodePool_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskarpenternodepool", "terraform")
}

func TestKubernetesKarpenterEc2NodeClass_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskarpenterec2nodeclass", "terraform")
}

// ─── Tier 2 Pulumi (Helm-based) ─────────────────────────────────────────────

func TestKubernetesValkey_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesvalkey", "pulumi")
}
func TestKubernetesGrafana_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgrafana", "pulumi")
}
func TestKubernetesKubePrometheusStack_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskubeprometheusstack", "pulumi")
}

func TestKubernetesPrometheusRule_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesprometheusrule", "pulumi")
}

func TestKubernetesServiceMonitor_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesservicemonitor", "pulumi")
}

func TestKubernetesPodMonitor_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespodmonitor", "pulumi")
}

func TestKubernetesOpenBao_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesopenbao", "pulumi")
}
func TestKubernetesOpenBao_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesopenbao", "terraform")
}
func TestKubernetesOpenFga_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesopenfga", "pulumi")
}
func TestKubernetesOpenFga_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesopenfga", "terraform")
}
func TestKubernetesHarbor_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesharbor", "pulumi")
}
func TestKubernetesHarbor_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesharbor", "terraform")
}
func TestKubernetesArgoCD_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesargocd", "pulumi")
}
func TestKubernetesArgoWorkflows_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesargoworkflows", "pulumi")
}
func TestKubernetesLocust_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteslocust", "pulumi")
}
func TestKubernetesNats_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesnats", "pulumi")
}
func TestKubernetesSeaweedFs_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesseaweedfs", "pulumi")
}
func TestKubernetesSeaweedFs_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesseaweedfs", "terraform")
}
func TestKubernetesLoki_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesloki", "pulumi")
}
func TestKubernetesLoki_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesloki", "terraform")
}
func TestKubernetesTempo_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestempo", "pulumi")
}
func TestKubernetesTempo_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestempo", "terraform")
}
func TestKubernetesQdrant_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesqdrant", "pulumi")
}
func TestKubernetesQdrant_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesqdrant", "terraform")
}
func TestKubernetesKyverno_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskyverno", "pulumi")
}
func TestKubernetesKyverno_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskyverno", "terraform")
}
func TestKubernetesGatekeeper_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgatekeeper", "pulumi")
}
func TestKubernetesGatekeeper_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgatekeeper", "terraform")
}
func TestKubernetesSparkOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessparkoperator", "pulumi")
}
func TestKubernetesSparkOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessparkoperator", "terraform")
}
func TestKubernetesRabbitMqOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesrabbitmqoperator", "pulumi")
}
func TestKubernetesRabbitMqOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesrabbitmqoperator", "terraform")
}
func TestKubernetesRabbitMq_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesrabbitmq", "pulumi")
}
func TestKubernetesRabbitMq_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesrabbitmq", "terraform")
}
func TestKubernetesNeo4j_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesneo4j", "pulumi")
}
func TestKubernetesNeo4j_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesneo4j", "terraform")
}
func TestKubernetesJenkins_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesjenkins", "pulumi")
}
func TestKubernetesSolrOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessolroperator", "pulumi")
}
func TestKubernetesPerconaMongoOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesperconamongooperator", "pulumi")
}
func TestKubernetesPerconaMysqlOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesperconamysqloperator", "pulumi")
}
func TestKubernetesTemporal_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestemporal", "pulumi")
}

// ─── Tier 2 Terraform (Helm-based) ──────────────────────────────────────────

func TestKubernetesValkey_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesvalkey", "terraform")
}
func TestKubernetesGrafana_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgrafana", "terraform")
}
func TestKubernetesKubePrometheusStack_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskubeprometheusstack", "terraform")
}

func TestKubernetesPrometheusRule_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesprometheusrule", "terraform")
}

func TestKubernetesServiceMonitor_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesservicemonitor", "terraform")
}

func TestKubernetesPodMonitor_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespodmonitor", "terraform")
}

func TestKubernetesArgoCD_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesargocd", "terraform")
}
func TestKubernetesArgoWorkflows_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesargoworkflows", "terraform")
}
func TestKubernetesLocust_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteslocust", "terraform")
}
func TestKubernetesNats_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesnats", "terraform")
}
func TestKubernetesSolrOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessolroperator", "terraform")
}
func TestKubernetesPerconaMongoOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesperconamongooperator", "terraform")
}
func TestKubernetesPerconaMysqlOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesperconamysqloperator", "terraform")
}
func TestKubernetesTemporal_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestemporal", "terraform")
}

// ─── Tier 1 Pulumi/Terraform (Postgres flagship) ────────────────────────────
// KubernetesPostgres declares KubernetesCloudNativePgOperator as a registry
// prerequisite; the harness installs the operator before every Cluster
// scenario. Backup and object-store recovery scenarios additionally declare
// KubernetesCnpgBarmanCloudPlugin, which chains cert-manager and the
// operator through its own registry edges.

func TestKubernetesCloudNativePgOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescloudnativepgoperator", "pulumi")
}
func TestKubernetesCloudNativePgOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescloudnativepgoperator", "terraform")
}
func TestKubernetesCnpgBarmanCloudPlugin_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescnpgbarmancloudplugin", "pulumi")
}
func TestKubernetesCnpgBarmanCloudPlugin_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetescnpgbarmancloudplugin", "terraform")
}
func TestKubernetesPostgres_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespostgres", "pulumi")
}
func TestKubernetesPostgres_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespostgres", "terraform")
}

// ─── Tier 3 Pulumi (operator-dependent) ─────────────────────────────────────

func TestKubernetesSignoz_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessignoz", "pulumi")
}
func TestKubernetesKafka_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafka", "pulumi")
}
func TestKubernetesKafkaTopic_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkatopic", "pulumi")
}
func TestKubernetesKafkaUser_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkauser", "pulumi")
}
func TestKubernetesKafkaConnect_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkaconnect", "pulumi")
}
func TestKubernetesKafkaConnector_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkaconnector", "pulumi")
}
func TestKubernetesKafkaMirrorMaker2_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkamirrormaker2", "pulumi")
}
func TestKubernetesKarapace_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskarapace", "pulumi")
}
func TestKubernetesKafkaUi_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkaui", "pulumi")
}
func TestKubernetesOpenSearch_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesopensearch", "pulumi")
}
func TestKubernetesMongodb_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmongodb", "pulumi")
}
func TestKubernetesMysql_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmysql", "pulumi")
}
func TestKubernetesSolr_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessolr", "pulumi")
}
func TestKubernetesClickHouse_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesclickhouse", "pulumi")
}
func TestKubernetesKeycloak_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskeycloak", "pulumi")
}
func TestKubernetesAirflow_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesairflow", "pulumi")
}
func TestKubernetesRayCluster_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesraycluster", "pulumi")
}
func TestKubernetesFlinkDeployment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesflinkdeployment", "pulumi")
}
func TestKubernetesJupyterHub_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesjupyterhub", "pulumi")
}
func TestKubernetesMlflow_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmlflow", "pulumi")
}
func TestKubernetesTrino_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestrino", "pulumi")
}
func TestKubernetesSuperset_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessuperset", "pulumi")
}
func TestKubernetesGhaRunnerScaleSet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgharunnerscaleset", "pulumi")
}
func TestKubernetesPlantonRunner_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesplantonrunner", "pulumi")
}
func TestKubernetesPlantonPlatform_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesplantonplatform", "pulumi")
}

// ─── Tier 3 Terraform (operator-dependent) ──────────────────────────────────

func TestKubernetesSignoz_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessignoz", "terraform")
}
func TestKubernetesKafka_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafka", "terraform")
}
func TestKubernetesKafkaTopic_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkatopic", "terraform")
}
func TestKubernetesKafkaUser_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkauser", "terraform")
}
func TestKubernetesKafkaConnect_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkaconnect", "terraform")
}
func TestKubernetesKafkaConnector_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkaconnector", "terraform")
}
func TestKubernetesKafkaMirrorMaker2_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkamirrormaker2", "terraform")
}
func TestKubernetesKarapace_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskarapace", "terraform")
}
func TestKubernetesKafkaUi_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskafkaui", "terraform")
}
func TestKubernetesOpenSearch_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesopensearch", "terraform")
}
func TestKubernetesMongodb_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmongodb", "terraform")
}
func TestKubernetesMysql_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmysql", "terraform")
}
func TestKubernetesSolr_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessolr", "terraform")
}
func TestKubernetesClickHouse_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesclickhouse", "terraform")
}
func TestKubernetesKeycloak_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskeycloak", "terraform")
}
func TestKubernetesAirflow_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesairflow", "terraform")
}
func TestKubernetesRayCluster_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesraycluster", "terraform")
}
func TestKubernetesFlinkDeployment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesflinkdeployment", "terraform")
}
func TestKubernetesJupyterHub_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesjupyterhub", "terraform")
}
func TestKubernetesMlflow_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesmlflow", "terraform")
}
func TestKubernetesTrino_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestrino", "terraform")
}
func TestKubernetesSuperset_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetessuperset", "terraform")
}
func TestKubernetesGhaRunnerScaleSet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgharunnerscaleset", "terraform")
}
func TestKubernetesPlantonRunner_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesplantonrunner", "terraform")
}
func TestKubernetesPlantonPlatform_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesplantonplatform", "terraform")
}

// ─── Tier 4 Pulumi (operators, addons) ──────────────────────────────────────

func TestKubernetesStrimziKafkaOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesstrimzikafkaoperator", "pulumi")
}
func TestKubernetesOpenSearchOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesopensearchoperator", "pulumi")
}
func TestKubernetesAltinityOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesaltinityoperator", "pulumi")
}
func TestKubernetesGatewayApiCrds_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgatewayapicrds", "pulumi")
}
func TestKubernetesGhaRunnerScaleSetController_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgharunnerscalesetcontroller", "pulumi")
}
func TestKubernetesTekton_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestekton", "pulumi")
}
func TestKubernetesKeycloakOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskeycloakoperator", "pulumi")
}
func TestKubernetesOtelOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesoteloperator", "pulumi")
}
func TestKubernetesKubeRayOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskuberayoperator", "pulumi")
}
func TestKubernetesFlinkOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesflinkoperator", "pulumi")
}
func TestKubernetesOtelCollector_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesotelcollector", "pulumi")
}
func TestKubernetesTektonOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestektonoperator", "pulumi")
}
func TestKubernetesPlantonOperator_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesplantonoperator", "pulumi")
}
func TestKubernetesIstio_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesistio", "pulumi")
}
func TestKubernetesIstioBaseCrds_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesistiobasecrds", "pulumi")
}

// ─── Tier 4 Terraform (operators, addons) ───────────────────────────────────

func TestKubernetesStrimziKafkaOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesstrimzikafkaoperator", "terraform")
}
func TestKubernetesOpenSearchOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesopensearchoperator", "terraform")
}
func TestKubernetesAltinityOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesaltinityoperator", "terraform")
}
func TestKubernetesGatewayApiCrds_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgatewayapicrds", "terraform")
}
func TestKubernetesGhaRunnerScaleSetController_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgharunnerscalesetcontroller", "terraform")
}
func TestKubernetesTekton_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestekton", "terraform")
}
func TestKubernetesKeycloakOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskeycloakoperator", "terraform")
}
func TestKubernetesOtelOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesoteloperator", "terraform")
}
func TestKubernetesKubeRayOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteskuberayoperator", "terraform")
}
func TestKubernetesFlinkOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesflinkoperator", "terraform")
}
func TestKubernetesOtelCollector_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesotelcollector", "terraform")
}
func TestKubernetesTektonOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestektonoperator", "terraform")
}
func TestKubernetesPlantonOperator_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesplantonoperator", "terraform")
}
func TestKubernetesIstioBaseCrds_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesistiobasecrds", "terraform")
}
func TestKubernetesIstio_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesistio", "terraform")
}

// ─── Gateway API Pulumi ─────────────────────────────────────────────────────
// Each kind declares KubernetesGatewayApiCrds as a registry prerequisite, which
// the harness installs before the scenario applies. Verification asserts the CR
// exists (controller-free: applies succeed once the CRDs are present).

func TestKubernetesGatewayClass_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgatewayclass", "pulumi")
}
func TestKubernetesGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgateway", "pulumi")
}
func TestKubernetesHttpRoute_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteshttproute", "pulumi")
}
func TestKubernetesGrpcRoute_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgrpcroute", "pulumi")
}
func TestKubernetesTcpRoute_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestcproute", "pulumi")
}
func TestKubernetesTlsRoute_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestlsroute", "pulumi")
}
func TestKubernetesReferenceGrant_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesreferencegrant", "pulumi")
}

func TestKubernetesBackendTlsPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesbackendtlspolicy", "pulumi")
}
func TestKubernetesUdpRoute_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesudproute", "pulumi")
}
func TestKubernetesListenerSet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteslistenerset", "pulumi")
}

// ─── Gateway API Terraform ──────────────────────────────────────────────────

func TestKubernetesGatewayClass_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgatewayclass", "terraform")
}
func TestKubernetesGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgateway", "terraform")
}
func TestKubernetesHttpRoute_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteshttproute", "terraform")
}
func TestKubernetesGrpcRoute_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesgrpcroute", "terraform")
}
func TestKubernetesTcpRoute_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestcproute", "terraform")
}
func TestKubernetesTlsRoute_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestlsroute", "terraform")
}
func TestKubernetesReferenceGrant_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesreferencegrant", "terraform")
}

func TestKubernetesBackendTlsPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesbackendtlspolicy", "terraform")
}
func TestKubernetesUdpRoute_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesudproute", "terraform")
}
func TestKubernetesListenerSet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kuberneteslistenerset", "terraform")
}

// ─── Istio API Pulumi (853-859) ─────────────────────────────────────────────
// Each kind declares KubernetesIstioBaseCrds as a registry prerequisite, which
// the harness installs (istio/base CRDs, no istiod) before the scenario applies.
// Verification asserts the typed Istio CR exists (object-grade); the
// authorizationpolicy behavioral-deny scenario additionally chains a real mesh.

func TestKubernetesPeerAuthentication_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespeerauthentication", "pulumi")
}

func TestKubernetesRequestAuthentication_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesrequestauthentication", "pulumi")
}

func TestKubernetesAuthorizationPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesauthorizationpolicy", "pulumi")
}

func TestKubernetesServiceEntry_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesserviceentry", "pulumi")
}

func TestKubernetesDestinationRule_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesdestinationrule", "pulumi")
}

func TestKubernetesEnvoyFilter_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesenvoyfilter", "pulumi")
}

func TestKubernetesTelemetry_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestelemetry", "pulumi")
}

// ─── Istio API Terraform (853-859) ──────────────────────────────────────────

func TestKubernetesPeerAuthentication_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetespeerauthentication", "terraform")
}

func TestKubernetesRequestAuthentication_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesrequestauthentication", "terraform")
}

func TestKubernetesAuthorizationPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesauthorizationpolicy", "terraform")
}

func TestKubernetesServiceEntry_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesserviceentry", "terraform")
}

func TestKubernetesDestinationRule_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesdestinationrule", "terraform")
}

func TestKubernetesEnvoyFilter_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetesenvoyfilter", "terraform")
}

func TestKubernetesTelemetry_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "kubernetestelemetry", "terraform")
}

// runAllScenariosForKind discovers and runs all E2E scenarios for a kind
// using the specified IaC engine ("pulumi" or "terraform").
func runAllScenariosForKind(t *testing.T, kindDir, engine string) {
	t.Helper()

	if cp, err := profilepkg.LoadKindProfile(repoRoot, "kubernetes", kindDir); err == nil && cp.Spec != nil {
		switch cp.Spec.Status {
		case kindv1.CatalogKindE2EProfileSpec_deferred,
			kindv1.CatalogKindE2EProfileSpec_skip,
			kindv1.CatalogKindE2EProfileSpec_stub,
			// pending_proof: fully authored, offline-validated, awaiting its
			// first live proof. The proving session flips the profile to green
			// immediately before executing the lanes; until then a sweep must
			// never run it.
			kindv1.CatalogKindE2EProfileSpec_pending_proof:
			reason := cp.Spec.DeferredReason
			if reason == "" {
				reason = cp.Spec.Status.String()
			}
			t.Skipf("kind %s E2E profile status is %s: %s", kindDir, cp.Spec.Status, reason)
		case kindv1.CatalogKindE2EProfileSpec_real_cluster:
			// Runs only against an externally provided real cluster; the
			// scenarios' own cluster-profile annotations then gate WHICH
			// real cluster satisfies each of them.
			if !testHarness.External() {
				reason := cp.Spec.DeferredReason
				if reason == "" {
					reason = "every lane requires an externally provided real cluster"
				}
				t.Skipf("kind %s E2E profile status is %s: %s", kindDir, cp.Spec.Status, reason)
			}
		}
	}

	moduleDir, err := discovery.ModuleDir(repoRoot, "kubernetes", kindDir, engine)
	if err != nil {
		t.Fatalf("failed to locate %s %s module: %v", kindDir, engine, err)
	}

	if !fileExists(moduleDir) {
		t.Skipf("kind %s %s module not found at %s", kindDir, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "kubernetes", kindDir)
	if err != nil {
		t.Fatalf("failed to discover test scenarios for %s: %v", kindDir, err)
	}

	if len(scenarios) == 0 {
		t.Skipf("no test scenarios found for %s under its e2e/scenarios directory", kindDir)
	}

	t.Logf("Discovered %d scenarios for %s [%s]", len(scenarios), kindDir, engine)

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			runSingleScenario(t, kindDir, moduleDir, engine, scenario)
		})
	}
}

func runSingleScenario(t *testing.T, kindDir, moduleDir, engine string, scenario discovery.TestScenario) {
	t.Helper()

	// Scenarios restricted to one engine (the e2e-engines annotation — spec
	// arms the other engine rejects by documented PARITY-EXCEPTION design)
	// skip the excluded engine's lane with the reason instead of failing on
	// their own designed rejection.
	if ok, err := runner.ScenarioSupportsEngine(scenario.ManifestPath, engine); err != nil {
		t.Fatalf("reading engine restriction for scenario %s/%s: %v", kindDir, scenario.Name, err)
	} else if !ok {
		t.Skipf("scenario %s/%s does not run on engine %s (per %s)",
			kindDir, scenario.Name, engine, runner.ScenarioEnginesAnnotation)
	}

	// Scenarios needing owner-arranged external credentials (the
	// e2e-required-env annotation) skip honestly where the environment
	// does not carry the arrangement — unset ${E2E_ENV:...} tokens would
	// otherwise fail expansion loudly, turning a deferral into a false
	// failure on every lane without the tokens (CI included).
	if missing, err := runner.ScenarioMissingRequiredEnv(scenario.ManifestPath); err != nil {
		t.Fatalf("reading required-env declaration for scenario %s/%s: %v", kindDir, scenario.Name, err)
	} else if len(missing) > 0 {
		t.Skipf("scenario %s/%s needs owner-arranged environment variables that are unset: %s (per %s)",
			kindDir, scenario.Name, strings.Join(missing, ", "), runner.ScenarioRequiredEnvAnnotation)
	}

	// Route the scenario to the cluster its manifest asks for (the
	// e2e-cluster-profile annotation; default = the shared cluster) and point
	// the process KUBECONFIG at it. Both engines read cluster credentials
	// through the environment and scenarios run serially within a process, so
	// activating per scenario is what keeps multi-cluster runs race-free.
	// A skip reason means the scenario's profile cannot be satisfied in this
	// lane by design (real-cluster profiles locally; unmatched profiles on an
	// external cluster) — honest skip, never a wrong-cluster run.
	scenarioHarness, skipReason, err := harnessForScenario(scenario.ManifestPath)
	if err != nil {
		t.Fatalf("failed to resolve cluster for scenario %s/%s: %v", kindDir, scenario.Name, err)
	}
	if skipReason != "" {
		t.Skipf("scenario %s/%s: %s", kindDir, scenario.Name, skipReason)
	}
	scenarioHarness.ActivateKubeconfig()

	tc := &provider.KindTestContext{
		Kind:         kindDir,
		Provider:     "kubernetes",
		Engine:       engine,
		ModuleDir:    moduleDir,
		ManifestPath: scenario.ManifestPath,
		RepoRoot:     repoRoot,
		RunID:        runID,
		T:            t,
		// A prerequisite deploys on Pulumi whenever its kind has a Pulumi
		// module, as these kinds do — even for Terraform scenarios — so the
		// backend URL must be set unconditionally.
		// Leaving it empty makes the dependency stacks fall back to the
		// machine's ambient `pulumi login` backend, coupling the run to
		// stale developer state.
		BackendURL: pulumiBackendURL,
	}

	if engine == "pulumi" {
		// GenerateStackName enforces the length cap uniqueness-preservingly
		// (blind truncation here would collide long kind names' scenarios).
		tc.StackName = runner.GenerateStackName(kindDir+"-"+scenario.Name, runID)
	}

	ctx := context.Background()
	result := runner.RunKindTest(ctx, tc, scenarioHarness)

	for _, phase := range result.Phases {
		status := "PASS"
		if !phase.Passed {
			status = "FAIL"
		}
		t.Logf("  %s: %s (%s)", phase.Phase, status, phase.Duration)
		if phase.Error != nil {
			t.Logf("    Error: %v", phase.Error)
		}
	}

	if !result.Passed {
		t.Fatalf("scenario %s/%s [%s] failed (total: %s)", kindDir, scenario.Name, engine, result.Duration)
	}

	t.Logf("scenario %s/%s [%s] passed (total: %s)", kindDir, scenario.Name, engine, result.Duration)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
