package pulumihcloudprovider

import (
	"fmt"

	hetznercloudprovider "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds a pulumi-hcloud Provider using the supplied credential.
// If the credential is nil or any individual field is blank, Pulumi's provider
// will fall back to the HCLOUD_TOKEN environment variable, matching Terraform's
// behavior.
func Get(
	ctx *pulumi.Context,
	hetznercloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig,
	nameSuffixes ...string,
) (*hcloud.Provider, error) {

	providerArgs := &hcloud.ProviderArgs{}

	if hetznercloudProviderConfig != nil {
		if hetznercloudProviderConfig.Token != "" {
			providerArgs.Token = pulumi.StringPtr(hetznercloudProviderConfig.Token)
		}
		if hetznercloudProviderConfig.Endpoint != "" {
			providerArgs.Endpoint = pulumi.StringPtr(hetznercloudProviderConfig.Endpoint)
		}
		if hetznercloudProviderConfig.EndpointHetzner != "" {
			providerArgs.EndpointHetzner = pulumi.StringPtr(hetznercloudProviderConfig.EndpointHetzner)
		}
		if hetznercloudProviderConfig.PollInterval != "" {
			providerArgs.PollInterval = pulumi.StringPtr(hetznercloudProviderConfig.PollInterval)
		}
		if hetznercloudProviderConfig.PollFunction != "" {
			providerArgs.PollFunction = pulumi.StringPtr(hetznercloudProviderConfig.PollFunction)
		}
	}

	provider, err := hcloud.NewProvider(
		ctx,
		ProviderResourceName(nameSuffixes),
		providerArgs,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create hetzner cloud provider")
	}

	return provider, nil
}

// ProviderResourceName builds a deterministic Pulumi resource name such as
// "hetznercloud" or "hetznercloud-secondary".
func ProviderResourceName(suffixes []string) string {
	name := "hetznercloud"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}
