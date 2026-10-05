package module

import (
	"github.com/pkg/errors"
	azureeventhubschemagroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubschemagroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventhubschemagroupv1alpha1.AzureEventHubSchemaGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHubSchemaGroup.Spec

	// The schema group, addressed by the parent namespace's ARM id. This
	// resource has NO update surface -- Azure exposes no mutable
	// properties on a schema group, so every field is ForceNew and any
	// change replaces the group (dropping the schemas registered inside
	// it). The registry's tier contract (STANDARD or higher namespace) is
	// enforced by Azure at apply time.
	schemaGroupArgs := &eventhub.NamespaceSchemaGroupArgs{
		// ForceNew: renaming replaces the group and drops its registered
		// schemas.
		Name:        pulumi.String(spec.SchemaGroupName),
		NamespaceId: pulumi.String(locals.NamespaceId),

		// Evolution policy and format, mapped from the spec enums to
		// ARM's wire values. Both ForceNew.
		SchemaCompatibility: pulumi.String(schemaCompatibilityStrings[spec.SchemaCompatibility]),
		SchemaType:          pulumi.String(schemaTypeStrings[spec.SchemaType]),
	}

	createdSchemaGroup, err := eventhub.NewNamespaceSchemaGroup(ctx,
		spec.SchemaGroupName,
		schemaGroupArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Event Hub schema group %s", spec.SchemaGroupName)
	}

	// Export outputs: what schema-registry serializers address at
	// runtime, alongside the namespace's fully-qualified hostname.
	ctx.Export(OpSchemaGroupId, createdSchemaGroup.ID())
	ctx.Export(OpSchemaGroupName, createdSchemaGroup.Name)

	return nil
}
