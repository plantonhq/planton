package module

import (
	"fmt"
	"strconv"

	kubernetesrayclusterv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesraycluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep: same rendered CR body, same
// outputs.
type Locals struct {
	KubernetesRayCluster *kubernetesrayclusterv1alpha1.KubernetesRayCluster
	Spec                 *kubernetesrayclusterv1alpha1.KubernetesRayClusterSpec

	// ResourceName is metadata.name — the CR name and the operator's
	// naming root: the head Service is `<name>-head-svc`, pod names
	// prefix it, and in token auth mode the generated bearer-token
	// Secret is named EXACTLY the cluster name.
	ResourceName string

	// Namespace the cluster deploys into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Labels tie the module-created objects back to the Planton
	// resource.
	Labels map[string]string

	// Image every Ray node runs: spec.image when set, else
	// `rayproject/ray:<ray_version>` — the VERSION/IMAGE LOCKSTEP the
	// operator relies on (it reads rayVersion to shape its commands but
	// runs the image as given; a mismatch fails at runtime, not at
	// apply).
	Image string

	// TokenAuthEnabled is this catalog's secure-by-default posture:
	// true unless spec.auth.mode is explicitly "disabled".
	TokenAuthEnabled bool

	// AuthTokenSecretName is the bearer-token credential handle (token
	// mode only): the bring-your-own Secret when named, else the
	// operator-generated Secret named EXACTLY the cluster name
	// (reconcileAuthSecret → utils.CheckName(instance.Name);
	// CheckName only rewrites names past 50 characters, which the
	// 40-char name budget rules out). Empty when auth is disabled.
	AuthTokenSecretName string

	// Output handles per the operator naming contract.
	HeadService        string
	ClientEndpoint     string
	DashboardEndpoint  string
	GcsEndpoint        string
	PortForwardCommand string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesrayclusterv1alpha1.KubernetesRayClusterIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesRayCluster.String(),
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

	resourceName := target.Metadata.Name
	namespace := spec.Namespace.GetValue()

	image := spec.GetImage()
	if image == "" {
		image = fmt.Sprintf("%s:%s", vars.DefaultImageRepository, spec.GetRayVersion())
	}

	// Empty auth or empty mode means token (the spec's option default);
	// only an explicit "disabled" opts out.
	tokenAuthEnabled := spec.GetAuth().GetMode() != "disabled"

	authTokenSecretName := ""
	if tokenAuthEnabled {
		authTokenSecretName = spec.GetAuth().GetExistingTokenSecretName()
		if authTokenSecretName == "" {
			authTokenSecretName = resourceName
		}
	}

	headService := resourceName + vars.HeadServiceSuffix

	return &Locals{
		KubernetesRayCluster: target,
		Spec:                 spec,
		ResourceName:         resourceName,
		Namespace:            namespace,
		Labels:               labels,
		Image:                image,
		TokenAuthEnabled:     tokenAuthEnabled,
		AuthTokenSecretName:  authTokenSecretName,
		HeadService:          headService,
		ClientEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			headService, namespace, vars.ClientPort),
		DashboardEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			headService, namespace, vars.DashboardPort),
		GcsEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			headService, namespace, vars.GcsPort),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward -n %s service/%s %d:%d",
			namespace, headService, vars.DashboardPort, vars.DashboardPort),
	}
}
