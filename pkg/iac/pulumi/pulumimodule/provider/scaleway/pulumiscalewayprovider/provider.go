package pulumiscalewayprovider

import (
	"fmt"
	"reflect"

	scalewayprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
)

// Get builds a pulumi-scaleway Provider using the supplied credential.
// If the credential is nil or any individual field is blank, Pulumi's provider
// will fall back to environment variables (SCW_ACCESS_KEY, SCW_SECRET_KEY, etc.),
// matching Terraform's behavior.
func Get(
	ctx *pulumi.Context,
	scalewayProviderConfig *scalewayprovider.ScalewayProviderConfig,
	nameSuffixes ...string,
) (*scaleway.Provider, error) {

	providerArgs := &scaleway.ProviderArgs{}

	// Map credential fields when present; leave them nil to defer to env-vars.
	if scalewayProviderConfig != nil {
		if scalewayProviderConfig.AccessKey != "" {
			providerArgs.AccessKey = pulumi.StringPtr(scalewayProviderConfig.AccessKey)
		}
		if scalewayProviderConfig.SecretKey != "" {
			providerArgs.SecretKey = pulumi.StringPtr(scalewayProviderConfig.SecretKey)
		}
		if scalewayProviderConfig.ProjectId != "" {
			providerArgs.ProjectId = pulumi.StringPtr(scalewayProviderConfig.ProjectId)
		}
		if scalewayProviderConfig.OrganizationId != "" {
			providerArgs.OrganizationId = pulumi.StringPtr(scalewayProviderConfig.OrganizationId)
		}
		if scalewayProviderConfig.Region != "" {
			providerArgs.Region = pulumi.StringPtr(scalewayProviderConfig.Region)
		}
		if scalewayProviderConfig.Zone != "" {
			providerArgs.Zone = pulumi.StringPtr(scalewayProviderConfig.Zone)
		}
	}

	provider, err := scaleway.NewProvider(
		ctx,
		ProviderResourceName(nameSuffixes),
		providerArgs,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scaleway provider")
	}

	return provider, nil
}

// ProviderResourceName builds a deterministic Pulumi resource name such as
// "scaleway-primary". Mirrors the digitalocean/civo helpers for naming consistency.
func ProviderResourceName(suffixes []string) string {
	name := "scaleway"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}

// PulumiOutputName produces canonical output names (e.g. "scw_vpc_id") to keep
// stack outputs predictable across modules.
func PulumiOutputName(r interface{}, name string, suffixes ...string) string {
	output := fmt.Sprintf("scw_%s", pulumioutput.Name(reflect.TypeOf(r), name))
	for _, s := range suffixes {
		output = fmt.Sprintf("%s_%s", output, s)
	}
	return output
}
