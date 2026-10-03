// Package main provides the Pulumi program entrypoint for AWS MemoryDB ACL deployment.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awsmemorydbacl/iac/pulumi/module"
	awsmemorydbaclv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmemorydbacl/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awsmemorydbaclv1alpha1.AwsMemorydbAclIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
