package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud"
	hetznercloudsshkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud/hetznercloudsshkey/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudSshKey         *hetznercloudsshkeyv1.HetznerCloudSshKey
	Labels                     map[string]string
}

// initializeLocals copies stack-input fields into the Locals struct and builds
// a reusable label map. Hetzner Cloud labels are key-value pairs (map), unlike
// Scaleway's flat string tags.
func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudsshkeyv1.HetznerCloudSshKeyStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudSshKey = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudSshKey.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudSshKey.String(),
	}

	if locals.HetznerCloudSshKey.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudSshKey.Metadata.Org
	}

	if locals.HetznerCloudSshKey.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudSshKey.Metadata.Env
	}

	if locals.HetznerCloudSshKey.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudSshKey.Metadata.Id
	}

	// Merge user-specified metadata labels; standard labels take precedence.
	for k, v := range locals.HetznerCloudSshKey.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
