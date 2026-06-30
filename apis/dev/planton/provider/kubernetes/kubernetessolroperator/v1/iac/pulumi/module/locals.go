package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetessolroperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetessolroperator/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed configuration values from the stack input
type Locals struct {
	// KubernetesSolrOperator is the target resource
	KubernetesSolrOperator *kubernetessolroperatorv1.KubernetesSolrOperator

	// Namespace is the Kubernetes namespace to deploy to
	Namespace string

	// Labels are common labels applied to all resources
	Labels map[string]string

	// HelmReleaseName is the name of the Helm release (prefixed with metadata.name for uniqueness)
	HelmReleaseName string

	// CrdsResourceName is the name of the CRDs ConfigFile resource (prefixed for uniqueness)
	CrdsResourceName string

	// ChartVersion is the Helm chart version to install (without 'v' prefix)
	ChartVersion string

	// CrdManifestURL is the URL to download CRDs from
	CrdManifestURL string
}

// initializeLocals creates computed values from stack input
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetessolroperatorv1.KubernetesSolrOperatorStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesSolrOperator = stackInput.Target

	target := stackInput.Target

	// Build common labels
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesSolrOperator.String(),
	}

	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// Get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// Export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Users can prefix metadata.name with component type if needed (e.g., "solr-operator-prod")
	locals.HelmReleaseName = target.Metadata.Name
	locals.CrdsResourceName = fmt.Sprintf("%s-crds", target.Metadata.Name)

	// Helm chart version without 'v' prefix
	// The default version (v0.9.1) is set in spec.proto via options.default
	locals.ChartVersion = strings.TrimPrefix(target.Spec.OperatorVersion, "v")

	// CRD manifest URL uses version without 'v' prefix
	locals.CrdManifestURL = fmt.Sprintf(vars.CrdManifestURLFormat, locals.ChartVersion)

	return locals
}
