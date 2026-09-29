package module

import (
	"strings"

	azuremachinelearningdatastorev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningdatastore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMachineLearningDatastore *azuremachinelearningdatastorev1alpha1.AzureMachineLearningDatastore

	// WorkspaceId is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal ARM ID.
	WorkspaceId string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order. NOTE: datastore tags are ForceNew on
	// the provider -- changing any tag replaces the datastore object (the
	// data it points at is untouched).
	AzureTags map[string]string
}

// serviceDataIdentityWire maps the spec's service-data identity modes to
// the provider's wire values. Unspecified is absent -- the property is
// omitted so the provider applies its default, "None". The provider
// names this argument `service_data_auth_identity` on the blob resource
// and `service_data_identity` on the other two -- ONE spec field feeds
// both (recorded in the parity manifest).
var serviceDataIdentityWire = map[azuremachinelearningdatastorev1alpha1.AzureMachineLearningDatastoreServiceDataIdentity]string{
	azuremachinelearningdatastorev1alpha1.AzureMachineLearningDatastoreServiceDataIdentity_NONE:                               "None",
	azuremachinelearningdatastorev1alpha1.AzureMachineLearningDatastoreServiceDataIdentity_WORKSPACE_SYSTEM_ASSIGNED_IDENTITY: "WorkspaceSystemAssignedIdentity",
	azuremachinelearningdatastorev1alpha1.AzureMachineLearningDatastoreServiceDataIdentity_WORKSPACE_USER_ASSIGNED_IDENTITY:   "WorkspaceUserAssignedIdentity",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuremachinelearningdatastorev1alpha1.AzureMachineLearningDatastoreStackInput) *Locals {
	locals := &Locals{}

	locals.AzureMachineLearningDatastore = stackInput.Target
	target := stackInput.Target

	locals.WorkspaceId = target.Spec.WorkspaceId.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureMachineLearningDatastore.String()),
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

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
