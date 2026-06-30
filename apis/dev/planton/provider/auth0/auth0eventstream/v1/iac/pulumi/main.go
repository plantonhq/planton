package main

import (
	"github.com/pkg/errors"
	auth0eventstreamv1 "github.com/plantonhq/planton/apis/dev/planton/provider/auth0/auth0eventstream/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/auth0/auth0eventstream/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &auth0eventstreamv1.Auth0EventStreamStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
