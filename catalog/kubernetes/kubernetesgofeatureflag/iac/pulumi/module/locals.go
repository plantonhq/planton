package module

import (
	"fmt"
	"strconv"

	kubernetesgofeatureflagv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgofeatureflag/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf - keep them in lockstep.
type Locals struct {
	Spec *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec

	// Resource-identity labels: stamped on the module-created objects
	// (namespace, env Secret, Roles, RoleBindings, ServiceMonitor) and on
	// the relay pods through the chart's podLabels.
	Labels map[string]string

	// Namespace the relay installs into.
	Namespace string

	// ReleaseName is metadata.name. fullnameOverride pins the chart
	// fullname to it, so the Deployment, Service, ConfigMap and
	// ServiceAccount all carry exactly this name.
	ReleaseName string

	// ChartVersion resolved to the pinned default when unset.
	ChartVersion string

	// ServiceName is the chart's Service (the fullname).
	ServiceName string

	// ServiceAccountName is the account the relay runs as: the chart's own
	// (the fullname) or an existing one.
	ServiceAccountName string

	// Port / MonitoringPort are the resolved listener ports.
	Port           int
	MonitoringPort int

	// Relay is the rendered configuration, secret environment and
	// ConfigMap grants.
	Relay *RelayConfig

	// EnvSecretName is the module-owned Secret carrying every secret value
	// as an environment variable; "" when the spec holds no secret.
	EnvSecretName string

	// FlagReaderName names the Role and RoleBinding in each namespace the
	// retrievers read ConfigMaps from.
	FlagReaderName string

	// ServiceMonitorName names the optional ServiceMonitor.
	ServiceMonitorName string

	// Composition handles (twin: outputs.tf).
	ApiEndpoint        string
	MonitoringEndpoint string
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagIacInput) (*Locals, error) {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesGoFeatureFlag.String(),
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

	chartVersion := spec.GetChartVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultChartVersion
	}

	namespace := spec.Namespace.GetValue()
	releaseName := target.Metadata.Name

	serviceAccountName := releaseName
	if spec.GetServiceAccount().GetExistingName() != "" {
		serviceAccountName = spec.GetServiceAccount().GetExistingName()
	}

	port := vars.DefaultPort
	if spec.GetServer() != nil && spec.GetServer().Port != nil {
		port = int(spec.GetServer().GetPort())
	}
	monitoringPort := vars.DefaultMonitoringPort
	if spec.GetServer() != nil && spec.GetServer().MonitoringPort != nil {
		monitoringPort = int(spec.GetServer().GetMonitoringPort())
	}

	relay, err := buildRelayConfig(spec, namespace)
	if err != nil {
		return nil, err
	}

	envSecretName := ""
	if len(relay.SecretEnv) > 0 {
		envSecretName = releaseName + vars.EnvSecretSuffix
	}

	return &Locals{
		Spec:               spec,
		Labels:             labels,
		Namespace:          namespace,
		ReleaseName:        releaseName,
		ChartVersion:       chartVersion,
		ServiceName:        releaseName,
		ServiceAccountName: serviceAccountName,
		Port:               port,
		MonitoringPort:     monitoringPort,
		Relay:              relay,
		EnvSecretName:      envSecretName,
		FlagReaderName:     releaseName + vars.FlagReaderSuffix,
		ServiceMonitorName: releaseName + vars.ServiceMonitorSuffix,
		ApiEndpoint:        fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", releaseName, namespace, port),
		MonitoringEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", releaseName, namespace, monitoringPort),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward -n %s svc/%s %d:%d", namespace, releaseName, port, port),
	}, nil
}
