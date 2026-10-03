// Package main provides the Pulumi program entrypoint for AWS MemoryDB user deployment.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awsmemorydbuser/iac/pulumi/module"
	awsmemorydbuserv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmemorydbuser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awsmemorydbuserv1alpha1.AwsMemorydbUserIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
