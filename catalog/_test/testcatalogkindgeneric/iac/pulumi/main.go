package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/_test/testcatalogkindgeneric/iac/pulumi/module"
	testcatalogkindgenericv1alpha2 "github.com/plantonhq/planton/catalog/_test/testcatalogkindgeneric/v1alpha2"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &testcatalogkindgenericv1alpha2.TestCatalogKindGenericIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
