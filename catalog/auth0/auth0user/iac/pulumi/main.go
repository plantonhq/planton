package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/auth0/auth0user/iac/pulumi/module"
	auth0userv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0user/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &auth0userv1alpha1.Auth0UserIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
