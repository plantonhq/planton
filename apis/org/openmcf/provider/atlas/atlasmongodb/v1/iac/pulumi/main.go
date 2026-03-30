package main

import (
	"github.com/pkg/errors"
	atlasmongodbv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/atlas/atlasmongodb/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/atlas/atlasmongodb/v1/iac/pulumi/module"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &atlasmongodbv1.AtlasMongodbStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
