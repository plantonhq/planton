package module

import (
	"fmt"
	"strconv"

	kuberneteskafkauiv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkaui/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteskafkauiv1alpha1.KubernetesKafkaUiSpec

	// Resource-identity labels stamped on the module-created satellites
	// (namespace, the console Secret — never injected into the chart's
	// own resources; Helm owns those).
	Labels map[string]string

	// Namespace the console installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// consoles coexist in one cluster (one per org, one per environment),
	// so each manifest gets its own release.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Whether login_form authentication was declared. Gates the console
	// Secret, the LOGIN_FORM values, and the KAFKA_UI_USER_PASSWORD
	// secret mapping.
	AuthEnabled bool

	// Deterministic name of the module-materialized Secret
	// ("<metadata.name>-secrets") holding every LITERAL credential the
	// spec declares (today: the console login password). Referenced
	// credentials (sasl / schema registry / Connect password_secret
	// entries) stay in their source Secrets — the console consumes both
	// kinds through envs.secretMappings, so no secret value ever lands
	// in rendered chart values.
	ConsoleSecretName string

	// Console replica count, resolved to the spec default (1) when
	// unset — the console is stateless, so this is purely availability.
	Replicas int32

	// Service exposure, resolved to the spec defaults (ClusterIP / 80)
	// when unset — always rendered so both engines emit the identical
	// service block and the endpoint output is deterministic.
	ServiceType string
	ServicePort int32

	// Name of the Service the chart renders — pinned to the resource
	// name through fullnameOverride (the catalog's Helm-kind
	// convention): outputs stay deterministic, multiple consoles per
	// cluster never collide on derived names, and verifiers/exposure
	// kinds address the Service by the resource name alone.
	ServiceName string

	// In-cluster console endpoint.
	Endpoint string

	// kubectl one-liner for reaching the console from a workstation
	// without any exposure. Local side pinned to 8080 (the container
	// port) since the Service port is often 80 — unprivileged locally.
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskafkauiv1alpha1.KubernetesKafkaUiIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKafkaUi.String(),
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

	replicas := int32(1)
	if spec.Replicas != nil {
		replicas = spec.GetReplicas()
	}

	serviceType := "ClusterIP"
	if spec.ServiceType != nil && spec.GetServiceType() != "" {
		serviceType = spec.GetServiceType()
	}

	servicePort := int32(80)
	if spec.ServicePort != nil {
		servicePort = spec.GetServicePort()
	}

	namespace := spec.Namespace.GetValue()
	releaseName := target.Metadata.Name
	// fullnameOverride pins the chart's object names to the release
	// name, so the Service IS the resource name.
	serviceName := releaseName

	return &Locals{
		Spec:              spec,
		Labels:            labels,
		Namespace:         namespace,
		ReleaseName:       releaseName,
		ChartVersion:      chartVersion,
		AuthEnabled:       spec.GetAuth() != nil,
		ConsoleSecretName: releaseName + "-secrets",
		Replicas:          replicas,
		ServiceType:       serviceType,
		ServicePort:       servicePort,
		ServiceName:       serviceName,
		Endpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
			serviceName, namespace, servicePort),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s 8080:%d",
			serviceName, namespace, servicePort),
	}
}
