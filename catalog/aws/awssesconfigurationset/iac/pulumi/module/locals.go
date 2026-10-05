package module

import (
	"strconv"

	awssesconfigurationsetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssesconfigurationset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	AwsSesConfigurationSet *awssesconfigurationsetv1alpha1.AwsSesConfigurationSet
	AwsTags                map[string]string
	// SendingEnabled is tri-state in the contract (proto optional): nil means
	// "not specified", and the catalog default is TRUE.
	SendingEnabled bool
	// EventDestinations maps destination name -> spec entry for per-name satellites.
	EventDestinations map[string]*awssesconfigurationsetv1alpha1.AwsSesConfigurationSetEventDestination
}

func initializeLocals(_ *pulumi.Context, iacInput *awssesconfigurationsetv1alpha1.AwsSesConfigurationSetIacInput) *Locals {
	locals := &Locals{}
	locals.AwsSesConfigurationSet = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsSesConfigurationSet.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsSesConfigurationSet.Metadata.Org,
		awstagkeys.Environment:  locals.AwsSesConfigurationSet.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSesConfigurationSet.String(),
		awstagkeys.ResourceId:   locals.AwsSesConfigurationSet.Metadata.Id,
	}

	spec := locals.AwsSesConfigurationSet.Spec
	locals.SendingEnabled = true
	if spec.SendingEnabled != nil {
		locals.SendingEnabled = spec.GetSendingEnabled()
	}

	locals.EventDestinations = make(map[string]*awssesconfigurationsetv1alpha1.AwsSesConfigurationSetEventDestination)
	for _, dest := range spec.GetEventDestinations() {
		locals.EventDestinations[dest.Name] = dest
	}

	return locals
}
