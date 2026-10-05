package module

import (
	"fmt"
	"strconv"

	kubernetesargocdv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesargocd/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesargocdv1alpha1.KubernetesArgocdSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace Argo CD installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// Argo CD instances can coexist in one cluster (one per namespace
	// when using the generated admin password — the app fixes that
	// Secret's name).
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Name of the API/UI server Service — `<fullname>-server`, with the
	// fullname pinned to the resource name via fullnameOverride.
	ServerServiceName string

	// Whether the server serves plain HTTP (spec.server.insecure — the
	// posture behind composed TLS-terminating exposure). The exported
	// endpoint and port-forward command follow it.
	ServerInsecure bool

	// In-cluster endpoint of the server (scheme follows ServerInsecure).
	ServerKubeEndpoint string

	// Whether the local admin user is enabled (spec default true) — the
	// initial-admin-secret handle is exported only while it is.
	AdminEnabled bool

	// Name of the generated initial-admin-password Secret ("" when the
	// admin user is disabled). Fixed by the application, never by the
	// chart or this module.
	InitialAdminSecretName string

	// kubectl one-liner for reaching the UI from a workstation.
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesargocdv1alpha1.KubernetesArgocdIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesArgocd.String(),
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
	// server Service is exactly `<name>-server`.
	serverServiceName := fmt.Sprintf("%s-server", releaseName)

	serverInsecure := spec.GetServer().GetInsecure()
	scheme := "https"
	forwardPort := 443
	if serverInsecure {
		scheme = "http"
		forwardPort = 80
	}

	// The admin user defaults ON (proto default true); the generated
	// initial-admin Secret exists only while it is enabled.
	adminEnabled := true
	if spec.AdminEnabled != nil {
		adminEnabled = spec.GetAdminEnabled()
	}
	initialAdminSecretName := ""
	if adminEnabled {
		initialAdminSecretName = vars.InitialAdminSecretName
	}

	return &Locals{
		Spec:                   spec,
		Labels:                 labels,
		Namespace:              namespace,
		ReleaseName:            releaseName,
		ChartVersion:           chartVersion,
		ServerServiceName:      serverServiceName,
		ServerInsecure:         serverInsecure,
		AdminEnabled:           adminEnabled,
		InitialAdminSecretName: initialAdminSecretName,
		ServerKubeEndpoint: fmt.Sprintf("%s://%s.%s.svc.cluster.local",
			scheme, serverServiceName, namespace),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s 8080:%d",
			serverServiceName, namespace, forwardPort),
	}
}
