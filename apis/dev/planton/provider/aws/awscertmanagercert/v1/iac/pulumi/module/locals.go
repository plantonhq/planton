package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awscertv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscertmanagercert/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved input values and tag metadata for the Pulumi stack.
type Locals struct {
	AwsCertManagerCert *awscertv1.AwsCertManagerCert
	AwsTags            map[string]string
}

// initializeLocals prepares a Locals object by resolving stack input and metadata-derived tags.
func initializeLocals(ctx *pulumi.Context, stackInput *awscertv1.AwsCertManagerCertStackInput) *Locals {
	locals := &Locals{}
	locals.AwsCertManagerCert = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsCertManagerCert.Metadata.Org,
		awstagkeys.Environment:  locals.AwsCertManagerCert.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsCertManagerCert.String(),
		awstagkeys.ResourceId:   locals.AwsCertManagerCert.Metadata.Id,
	}

	return locals
}
