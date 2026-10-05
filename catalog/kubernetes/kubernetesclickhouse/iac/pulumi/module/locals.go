package module

import (
	"fmt"
	"strconv"

	kubernetesclickhousev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesclickhouse/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	KubernetesClickHouse *kubernetesclickhousev1alpha1.KubernetesClickHouse
	Spec                 *kubernetesclickhousev1alpha1.KubernetesClickHouseSpec

	// Resource-identity labels stamped on the module-created objects
	// (namespace, auth Secret, CHI, CHK). The operator derives ITS
	// objects' identity from the CHI/CHK names; these labels tie the
	// family back to the Planton resource.
	Labels map[string]string

	// Namespace the cluster lives in (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// ChiName is metadata.name — the naming root the operator derives
	// every object from: the cluster-wide client Service
	// `clickhouse-<name>`, the per-cluster Service
	// `cluster-<name>-<cluster>`, and per-host StatefulSets/Services
	// `chi-<name>-<cluster>-<shard>-<replica>`.
	ChiName string

	// ClusterName is the logical ClickHouse cluster (remote_servers
	// entry, `ON CLUSTER` target). Spec default: "main".
	ClusterName string

	// Shards / Replicas are the resolved topology (spec defaults: 1/1).
	Shards   int
	Replicas int

	// Image is the resolved server image reference `repo:tag`
	// (spec.image.repo default clickhouse/clickhouse-server, tag default
	// spec.version).
	Image string

	// AuthSecretName is the module-managed Secret holding one key per
	// provisioned user — empty when spec.users is empty (nothing to
	// hold, output contract exports empty).
	AuthSecretName string

	// DeployKeeper resolves the coordination contract: an explicit
	// managed_keeper always deploys; UNSET coordination deploys exactly
	// when the topology needs coordination (replicas > 1 or shards > 1);
	// external/none never deploy.
	DeployKeeper bool

	// KeeperName / KeeperServiceName are the managed CHK handles
	// (`<name>-keeper`, operator Service naming contract
	// `keeper-<chk-name>`) — empty when no managed Keeper is deployed.
	KeeperName        string
	KeeperServiceName string

	// ServiceName is the operator's cluster-wide client Service
	// `clickhouse-<name>` covering all hosts.
	ServiceName string

	// TcpEndpoint / HttpEndpoint are the in-cluster endpoints on the
	// fixed native (9000) and HTTP (8123) interface ports.
	TcpEndpoint  string
	HttpEndpoint string

	PortForwardCommand string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesclickhousev1alpha1.KubernetesClickHouseIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesClickHouse.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	namespace := spec.Namespace.GetValue()
	chiName := target.Metadata.Name

	clusterName := spec.GetClusterName()
	if clusterName == "" {
		clusterName = vars.DefaultClusterName
	}

	shards := 1
	if spec.Shards != nil && spec.GetShards() > 0 {
		shards = int(spec.GetShards())
	}
	replicas := 1
	if spec.Replicas != nil && spec.GetReplicas() > 0 {
		replicas = int(spec.GetReplicas())
	}

	imageRepo := spec.GetImage().GetRepo()
	if imageRepo == "" {
		imageRepo = vars.DefaultImageRepo
	}
	imageTag := spec.GetImage().GetTag()
	if imageTag == "" {
		imageTag = spec.GetVersion()
	}

	authSecretName := ""
	if len(spec.GetUsers()) > 0 {
		authSecretName = chiName + vars.AuthSecretSuffix
	}

	coordinationType := kubernetesclickhousev1alpha1.KubernetesClickHouseCoordination_unspecified
	if spec.GetCoordination() != nil {
		coordinationType = spec.GetCoordination().GetType()
	}
	deployKeeper := coordinationType == kubernetesclickhousev1alpha1.KubernetesClickHouseCoordination_managed_keeper ||
		(coordinationType == kubernetesclickhousev1alpha1.KubernetesClickHouseCoordination_unspecified &&
			(shards > 1 || replicas > 1))

	keeperName := ""
	keeperServiceName := ""
	if deployKeeper {
		keeperName = chiName + "-keeper"
		keeperServiceName = "keeper-" + keeperName
	}

	serviceName := "clickhouse-" + chiName

	return &Locals{
		KubernetesClickHouse: target,
		Spec:                 spec,
		Labels:               labels,
		Namespace:            namespace,
		ChiName:              chiName,
		ClusterName:          clusterName,
		Shards:               shards,
		Replicas:             replicas,
		Image:                imageRepo + ":" + imageTag,
		AuthSecretName:       authSecretName,
		DeployKeeper:         deployKeeper,
		KeeperName:           keeperName,
		KeeperServiceName:    keeperServiceName,
		ServiceName:          serviceName,
		TcpEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			serviceName, namespace, vars.TcpPort),
		HttpEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
			serviceName, namespace, vars.HttpPort),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
			serviceName, namespace, vars.HttpPort, vars.HttpPort),
	}
}
