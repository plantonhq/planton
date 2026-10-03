package module

import (
	"fmt"
	"strconv"

	awssagemakerendpointv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakerendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awssagemakerendpointv1alpha1.AwsSagemakerEndpoint
	Spec   *awssagemakerendpointv1alpha1.AwsSagemakerEndpointSpec

	EndpointName string
	AwsTags      map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awssagemakerendpointv1alpha1.AwsSagemakerEndpointIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// The component's name IS the endpoint name.
	locals.EndpointName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerEndpoint.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}

// resolvedVariantName defaults variant names deterministically per
// position - identical to the Terraform module's locals, so both
// engines produce the same configuration.
func resolvedVariantName(v *awssagemakerendpointv1alpha1.AwsSagemakerEndpointVariant, index int, shadow bool) string {
	if v.VariantName != "" {
		return v.VariantName
	}
	if shadow {
		return fmt.Sprintf("shadow-variant-%d", index)
	}
	return fmt.Sprintf("variant-%d", index)
}
