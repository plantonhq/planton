package module

import (
	"github.com/pkg/errors"
	azurecontainerappenvironmentdaprcomponentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironmentdaprcomponent/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecontainerappenvironmentdaprcomponentv1alpha1.AzureContainerAppEnvironmentDaprComponentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerAppEnvironmentDaprComponent.Spec

	// The Dapr component is a pluggable backend (state store, pub/sub
	// broker, secret store, binding) registered once on the environment
	// and consumed by Dapr-enabled apps whose dapr.app_id appears in
	// scopes. Name, type, and environment are ForceNew.
	componentArgs := &containerapp.EnvironmentDaprComponentArgs{
		Name:                      pulumi.String(spec.ComponentName),
		ContainerAppEnvironmentId: pulumi.String(spec.ContainerAppEnvironmentId.GetValue()),
		ComponentType:             pulumi.String(spec.ComponentType),
		Version:                   pulumi.String(spec.Version),
	}

	// Sent value-or-default: Azure applies 5s itself, but sending the
	// documented default keeps both engines byte-identical on the wire.
	initTimeout := "5s"
	if spec.InitTimeout != nil {
		initTimeout = spec.GetInitTimeout()
	}
	componentArgs.InitTimeout = pulumi.String(initTimeout)

	// Left false so a broken component fails loudly at sidecar startup
	// instead of surfacing as runtime errors on first use.
	if spec.IgnoreErrors != nil {
		componentArgs.IgnoreErrors = pulumi.Bool(spec.GetIgnoreErrors())
	}

	// Connection strings and keys travel as component secrets referenced
	// from metadata by secret_name -- never as plain metadata values.
	if len(spec.Secrets) > 0 {
		secrets := make(containerapp.EnvironmentDaprComponentSecretArray, 0, len(spec.Secrets))
		for _, secret := range spec.Secrets {
			secrets = append(secrets, &containerapp.EnvironmentDaprComponentSecretArgs{
				Name:  pulumi.String(secret.Name),
				Value: pulumi.String(secret.Value),
			})
		}
		componentArgs.Secrets = secrets
	}

	// The component's configuration entries; keys depend on the component
	// type. The spec's CEL guarantees value XOR secret_name per entry.
	// Values may be foreign-key references (e.g. an azureClientId entry
	// tracking a managed identity's client_id); they arrive pre-resolved
	// to literals.
	if len(spec.Metadata) > 0 {
		metadataEntries := make(containerapp.EnvironmentDaprComponentMetadataArray, 0, len(spec.Metadata))
		for _, metadataEntry := range spec.Metadata {
			entryArgs := &containerapp.EnvironmentDaprComponentMetadataArgs{
				Name: pulumi.String(metadataEntry.Name),
			}
			if metadataEntry.SecretName != "" {
				entryArgs.SecretName = pulumi.StringPtr(metadataEntry.SecretName)
			} else if metadataEntry.Value.GetValue() != "" {
				entryArgs.Value = pulumi.StringPtr(metadataEntry.Value.GetValue())
			}
			metadataEntries = append(metadataEntries, entryArgs)
		}
		componentArgs.Metadatas = metadataEntries
	}

	// An empty scopes list exposes the component to every Dapr-enabled
	// app in the environment -- scope production components deliberately.
	if len(spec.Scopes) > 0 {
		componentArgs.Scopes = pulumi.ToStringArray(spec.Scopes)
	}

	createdComponent, err := containerapp.NewEnvironmentDaprComponent(ctx,
		spec.ComponentName,
		componentArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Dapr component %s", spec.ComponentName)
	}

	// Export outputs. Apps consume the component through Dapr's
	// runtime by its name.
	ctx.Export(OpDaprComponentId, createdComponent.ID())
	ctx.Export(OpComponentName, createdComponent.Name)

	return nil
}
