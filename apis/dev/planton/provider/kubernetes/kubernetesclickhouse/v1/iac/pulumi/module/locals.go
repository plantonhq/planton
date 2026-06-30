package module

import (
	"fmt"
	"strconv"

	kubernetesclickhousev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesclickhouse/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressExternalHostname     string
	KubePortForwardCommand      string
	KubeServiceFqdn             string
	KubeServiceName             string
	KubernetesClickHouse        *kubernetesclickhousev1.KubernetesClickHouse
	Namespace                   string
	ClickhousePodSelectorLabels map[string]string
	KubernetesLabels            map[string]string
	// Computed resource names to avoid conflicts when multiple instances share a namespace
	PasswordSecretName     string
	ExternalLbServiceName  string
	KeeperInstallationName string
	KeeperServiceName      string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesclickhousev1.KubernetesClickHouseStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesClickHouse = stackInput.Target

	target := stackInput.Target

	locals.KubernetesLabels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesClickHouse.String(),
	}

	if target.Metadata.Id != "" {
		locals.KubernetesLabels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.KubernetesLabels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.KubernetesLabels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// Get namespace from spec (required field)
	locals.Namespace = target.Spec.Namespace.GetValue()

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "clickhouse-my-db")
	locals.PasswordSecretName = fmt.Sprintf("%s-password", target.Metadata.Name)
	locals.ExternalLbServiceName = fmt.Sprintf("%s-external-lb", target.Metadata.Name)
	locals.KeeperInstallationName = fmt.Sprintf("%s-keeper", target.Metadata.Name)
	// Altinity operator creates keeper service with pattern: keeper-<chk-name>
	locals.KeeperServiceName = fmt.Sprintf("keeper-%s", locals.KeeperInstallationName)

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpUsername, pulumi.String(vars.DefaultUsername))
	ctx.Export(OpPasswordSecretName, pulumi.String(locals.PasswordSecretName))
	ctx.Export(OpPasswordSecretKey, pulumi.String(vars.ClickhousePasswordKey))

	locals.KubeServiceName = target.Metadata.Name

	// Determine cluster name - use spec.cluster_name if provided, otherwise use metadata.name
	clusterName := target.Spec.ClusterName
	if clusterName == "" {
		clusterName = target.Metadata.Name
	}

	// Altinity operator uses these labels for pod selection
	// These labels are automatically applied by the operator to ClickHouse pods
	locals.ClickhousePodSelectorLabels = map[string]string{
		"clickhouse.altinity.com/chi":     clusterName,
		"clickhouse.altinity.com/cluster": clusterName,
	}

	//export kubernetes service name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local:%d", locals.KubeServiceName, locals.Namespace, vars.ClickhouseHttpPort)

	//export kubernetes endpoint
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s %d:%d",
		locals.Namespace, target.Metadata.Name, vars.ClickhouseHttpPort, vars.ClickhouseHttpPort)

	//export kube-port-forward command
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals
	}

	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	//export ingress-external-hostname
	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

	return locals
}
