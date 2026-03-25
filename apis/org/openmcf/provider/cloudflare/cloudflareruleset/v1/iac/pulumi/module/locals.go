package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflarerulesetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareruleset/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareRuleset        *cloudflarerulesetv1.CloudflareRuleset
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarerulesetv1.CloudflareRulesetStackInput) *Locals {
	return &Locals{
		CloudflareRuleset:        stackInput.Target,
		CloudflareProviderConfig: stackInput.ProviderConfig,
	}
}
