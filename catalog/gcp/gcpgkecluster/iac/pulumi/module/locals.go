package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpgkeclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkecluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpGkeCluster     *gcpgkeclusterv1alpha1.GcpGkeCluster
	GcpLabels         map[string]string
	// ClusterName is the cloud-side cluster name: spec.cluster_name when
	// set, otherwise metadata.name (the spec-level contract).
	ClusterName string
	// ReleaseChannel is the provider's literal channel string. The spec's
	// NONE (opt out of channel-based upgrades) is spelled UNSPECIFIED on the
	// provider — the API has no literal NONE value.
	ReleaseChannel string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpgkeclusterv1alpha1.GcpGkeClusterIacInput) *Locals {
	locals := &Locals{}
	locals.GcpGkeCluster = iacInput.Target

	locals.ClusterName = locals.GcpGkeCluster.Spec.ClusterName
	if locals.ClusterName == "" {
		locals.ClusterName = locals.GcpGkeCluster.Metadata.Name
	}

	switch locals.GcpGkeCluster.Spec.GetReleaseChannel() {
	case gcpgkeclusterv1alpha1.GkeReleaseChannel_RAPID:
		locals.ReleaseChannel = "RAPID"
	case gcpgkeclusterv1alpha1.GkeReleaseChannel_STABLE:
		locals.ReleaseChannel = "STABLE"
	case gcpgkeclusterv1alpha1.GkeReleaseChannel_EXTENDED:
		locals.ReleaseChannel = "EXTENDED"
	case gcpgkeclusterv1alpha1.GkeReleaseChannel_NONE:
		locals.ReleaseChannel = "UNSPECIFIED"
	default:
		locals.ReleaseChannel = "REGULAR"
	}

	// User labels merge in first so the platform attribution labels can never
	// be clobbered by a spec label with the same key. The cluster name (not
	// metadata.name) keys the name label so the label matches what is
	// visible in the GCP console.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpGkeCluster.Spec.ResourceLabels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.ClusterName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpGkeCluster.String())

	if locals.GcpGkeCluster.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpGkeCluster.Metadata.Org
	}
	if locals.GcpGkeCluster.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpGkeCluster.Metadata.Env
	}
	if locals.GcpGkeCluster.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpGkeCluster.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
