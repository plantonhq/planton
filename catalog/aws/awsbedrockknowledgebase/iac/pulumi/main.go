package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awsbedrockknowledgebase/iac/pulumi/module"
	awsbedrockknowledgebasev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockknowledgebase/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awsbedrockknowledgebasev1alpha1.AwsBedrockKnowledgeBaseIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}
		return module.Resources(ctx, iacInput)
	})
}
