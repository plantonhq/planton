package module

import (
	"github.com/pkg/errors"
	azuremonitordatacollectionruleassociationv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitordatacollectionruleassociation/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremonitordatacollectionruleassociationv1alpha1.AzureMonitorDataCollectionRuleAssociationIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMonitorDataCollectionRuleAssociation.Spec

	// Attach the machine to a rule or an endpoint. The association is an
	// extension resource living ON the target machine -- creating it
	// puts the machine under the rule, destroying it detaches the
	// machine without touching the rule.
	associationArgs := &monitoring.DataCollectionRuleAssociationArgs{
		TargetResourceId: pulumi.String(locals.TargetResourceId),
	}

	// Left unset for endpoint bindings so the provider applies Azure's
	// mandated fixed name ("configurationAccessEndpoint"); required (and
	// spec-enforced) for rule bindings.
	if spec.Name != "" {
		associationArgs.Name = pulumi.String(spec.Name)
	}

	// Exactly one of the two bindings (spec CEL mirrors the provider's
	// ExactlyOneOf).
	if locals.DataCollectionRuleId != "" {
		associationArgs.DataCollectionRuleId = pulumi.String(locals.DataCollectionRuleId)
	}
	if locals.DataCollectionEndpointId != "" {
		associationArgs.DataCollectionEndpointId = pulumi.String(locals.DataCollectionEndpointId)
	}

	// Sent only when non-empty for a clean plan; Azure treats an absent
	// and an empty description identically.
	if spec.Description != "" {
		associationArgs.Description = pulumi.String(spec.Description)
	}

	createdAssociation, err := monitoring.NewDataCollectionRuleAssociation(ctx,
		locals.AzureMonitorDataCollectionRuleAssociation.Metadata.Name,
		associationArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data collection rule association %s",
			locals.AzureMonitorDataCollectionRuleAssociation.Metadata.Name)
	}

	ctx.Export(OpDataCollectionRuleAssociationId, createdAssociation.ID())
	ctx.Export(OpDataCollectionRuleAssociationName, createdAssociation.Name)

	return nil
}
