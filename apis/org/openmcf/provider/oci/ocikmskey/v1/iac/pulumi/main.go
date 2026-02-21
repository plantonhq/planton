package main

import (
	"github.com/pkg/errors"
	ocikmskeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocikmskey/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocikmskey/v1/iac/pulumi/module"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &ocikmskeyv1.OciKmsKeyStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
