package module

import (
	"fmt"
	"strconv"

	kubernetesgrafanav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrafana/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesgrafanav1alpha1.KubernetesGrafanaSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace Grafana installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// Grafana instances coexist in one Kubernetes cluster (per-team
	// dashboards, a mesh-specific instance, ...).
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Name of the Grafana Service — grafana.fullname, pinned to the
	// resource name via fullnameOverride (the 63-char child-name
	// discipline), so the Service and the chart-owned admin Secret derive
	// from the resource name deterministically.
	ServiceName string

	// Name of the Secret carrying the admin credentials: the referenced
	// existing Secret when spec.admin_secret is declared, else the
	// chart-owned `<name>` Secret the chart generates ONCE at first
	// install (stable across upgrades via its lookup).
	AdminSecretName string

	// In-cluster endpoint (the chart serves plain HTTP on the Service;
	// TLS terminates at the composed exposure layer — ingress or
	// gateway — never inside this module).
	Endpoint string

	// kubectl one-liner for reaching the UI from a workstation.
	PortForwardCommand string

	// Name of the module-owned Secret carrying the sign-in client
	// secrets (`<name>-sso`), and the sign-in rendering itself; nil when
	// no provider is declared.
	SsoSecretName string
	SignIn        *signIn
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesgrafanav1alpha1.KubernetesGrafanaIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesGrafana.String(),
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

	// fullnameOverride is pinned to metadata.name (values.go), so the
	// chart's Service is exactly the resource name.
	serviceName := releaseName

	// The chart-generated admin Secret is named grafana.fullname (= the
	// resource name); an existing Secret's own name wins when declared.
	adminSecretName := releaseName
	if existing := spec.GetAdminSecret(); existing != nil {
		adminSecretName = existing.GetName()
	}

	ssoSecretName := releaseName + vars.SsoSecretSuffix

	return &Locals{
		SsoSecretName:   ssoSecretName,
		SignIn:          buildSignIn(spec.GetAuth(), ssoSecretName),
		Spec:            spec,
		Labels:          labels,
		Namespace:       namespace,
		ReleaseName:     releaseName,
		ChartVersion:    chartVersion,
		ServiceName:     serviceName,
		AdminSecretName: adminSecretName,
		Endpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local",
			serviceName, namespace),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s 3000:%d",
			serviceName, namespace, vars.ServicePort),
	}
}
