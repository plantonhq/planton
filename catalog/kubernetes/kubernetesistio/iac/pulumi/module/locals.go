package module

import (
	"strconv"

	kubernetesistiov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesistio/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesistiov1alpha1.KubernetesIstioSpec

	// Resource-identity labels stamped on the namespace this module creates
	// (never injected into the charts' own resources — Helm owns those).
	Labels map[string]string

	// Namespace the control plane installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Control-plane revision: spec.revision when set, otherwise "default"
	// (the chart's own vocabulary for the unnamed revision).
	Revision string

	// istiod release name: "istiod" for the default revision,
	// "istiod-<revision>" for a named revision — matching how the chart
	// names the istiod Deployment/Service themselves.
	IstiodReleaseName string

	// The istiod Service name equals the release-derived resource name —
	// exported as the discovery-address handle.
	IstiodServiceName string

	// Chart version for every release, resolved to the pinned default when
	// unset, so both engines install the same charts whether or not the
	// platform's defaulting middleware ran.
	Version string

	// Ambient is true when spec.dataplane_mode == "ambient" — the cni and
	// ztunnel releases install only then (plus cni when spec.cni.enabled in
	// sidecar mode).
	Ambient bool

	// InstallCni: ambient mode always; sidecar mode when spec.cni.enabled.
	InstallCni bool

	// Resolved trust domain (spec value or the upstream default) — exported
	// for AuthorizationPolicy principal authoring.
	TrustDomain string

	// DataplaneMode as exported ("sidecar" or "ambient").
	DataplaneMode string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesistiov1alpha1.KubernetesIstioIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to
	// what the Terraform module stamps for the same manifest.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesIstio.String(),
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

	revision := spec.GetRevision()
	if revision == "" {
		revision = "default"
	}

	istiodReleaseName := vars.IstiodReleaseName
	if spec.GetRevision() != "" {
		istiodReleaseName = vars.IstiodReleaseName + "-" + spec.GetRevision()
	}

	version := spec.GetVersion()
	if version == "" {
		version = vars.DefaultVersion
	}

	ambient := spec.GetDataplaneMode() == "ambient"

	trustDomain := vars.DefaultTrustDomain
	if td := spec.GetMeshConfig().GetTrustDomain(); td != "" {
		trustDomain = td
	}

	dataplaneMode := spec.GetDataplaneMode()
	if dataplaneMode == "" {
		dataplaneMode = "sidecar"
	}

	return &Locals{
		Spec:              spec,
		Labels:            labels,
		Namespace:         spec.Namespace.GetValue(),
		Revision:          revision,
		IstiodReleaseName: istiodReleaseName,
		IstiodServiceName: istiodReleaseName,
		Version:           version,
		Ambient:           ambient,
		InstallCni:        ambient || spec.GetCni().GetEnabled(),
		TrustDomain:       trustDomain,
		DataplaneMode:     dataplaneMode,
	}
}
