package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/auth0/auth0clientfrommetadatadocument/iac/pulumi/module"
	auth0clientfrommetadatadocumentv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0clientfrommetadatadocument/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
