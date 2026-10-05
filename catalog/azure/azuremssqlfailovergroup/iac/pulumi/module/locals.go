package module

import (
	"strings"

	azuremssqlfailovergroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremssqlfailovergroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMssqlFailoverGroup *azuremssqlfailovergroupv1alpha1.AzureMssqlFailoverGroup

	// ServerId is the resolved ARM ID of the primary logical server;
	// PartnerServerIds and DatabaseIds are the resolved ARM IDs of the
	// partner servers and the databases to replicate.
	ServerId         string
	PartnerServerIds []string
	DatabaseIds      []string

	// FailoverMode is the ARM string for the read-write policy mode.
	FailoverMode string

	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremssqlfailovergroupv1alpha1.AzureMssqlFailoverGroupIacInput) *Locals {
	locals := &Locals{}

	locals.AzureMssqlFailoverGroup = iacInput.Target
	target := iacInput.Target
	spec := target.Spec

	locals.ServerId = spec.ServerId.GetValue()

	for _, p := range spec.PartnerServers {
		locals.PartnerServerIds = append(locals.PartnerServerIds, p.ServerId.GetValue())
	}
	for _, d := range spec.DatabaseIds {
		locals.DatabaseIds = append(locals.DatabaseIds, d.GetValue())
	}

	if spec.ReadWriteEndpointFailoverPolicy != nil {
		switch spec.ReadWriteEndpointFailoverPolicy.Mode {
		case azuremssqlfailovergroupv1alpha1.AzureMssqlFailoverGroupFailoverMode_AUTOMATIC:
			locals.FailoverMode = "Automatic"
		case azuremssqlfailovergroupv1alpha1.AzureMssqlFailoverGroupFailoverMode_MANUAL:
			locals.FailoverMode = "Manual"
		}
	}

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureMssqlFailoverGroup.String()),
	}
	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}
	for k, v := range spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
