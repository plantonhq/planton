package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awscertv1 "github.com/plantonhq/planton/catalog/aws/awscertmanagercert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved input values and tag metadata for the Pulumi stack.
type Locals struct {
	AwsCertManagerCert *awscertv1.AwsCertManagerCert
	AwsTags            map[string]string
}

// initializeLocals prepares a Locals object by resolving IaC input and metadata-derived tags.
func initializeLocals(_ *pulumi.Context, iacInput *awscertv1.AwsCertManagerCertIacInput) *Locals {
	locals := &Locals{}
	locals.AwsCertManagerCert = iacInput.Target

	metadata := iacInput.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key. ACM
	// certificates have no AWS name -- metadata.name drives the Name tag and
	// consumers address the certificate through its ARN.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsCertManagerCert.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
