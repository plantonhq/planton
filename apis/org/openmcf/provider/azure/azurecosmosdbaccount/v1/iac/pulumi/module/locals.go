package module

import (
	"strings"

	azurecosmosdbaccountv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurecosmosdbaccount/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureCosmosdbAccount *azurecosmosdbaccountv1.AzureCosmosdbAccount
	ResourceGroupName    string
	AzureTags            map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurecosmosdbaccountv1.AzureCosmosdbAccountStackInput) *Locals {
	locals := &Locals{}
	locals.AzureCosmosdbAccount = stackInput.Target
	target := stackInput.Target
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureCosmosdbAccount.String()),
	}
	if target.Metadata.Id != "" {
		locals.AzureTags["resource_id"] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.AzureTags["organization"] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.AzureTags["environment"] = target.Metadata.Env
	}
	return locals
}
