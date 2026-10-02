package module

import (
	"github.com/pkg/errors"
	gcpdialogflowcxagentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdialogflowcxagent/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpdialogflowcxagentv1alpha1.GcpDialogflowCxAgentStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdAgent, err := agent(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create dialogflow cx agent")
	}

	// The folded children, each parented to the agent so they are created
	// after it and destroyed before it.
	if err := webhooks(ctx, locals, gcpProvider, createdAgent); err != nil {
		return errors.Wrap(err, "failed to create dialogflow cx webhooks")
	}
	if err := tools(ctx, locals, gcpProvider, createdAgent); err != nil {
		return errors.Wrap(err, "failed to create dialogflow cx tools")
	}
	if err := versionsAndEnvironments(ctx, locals, gcpProvider, createdAgent); err != nil {
		return errors.Wrap(err, "failed to create dialogflow cx versions and environments")
	}
	if err := generativeSettings(ctx, locals, gcpProvider, createdAgent); err != nil {
		return errors.Wrap(err, "failed to create dialogflow cx generative settings")
	}

	return nil
}
