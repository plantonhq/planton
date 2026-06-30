package main

import (
	"github.com/pkg/errors"
	cloudflarecertificatepackv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarecertificatepack/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarecertificatepack/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &cloudflarecertificatepackv1.CloudflareCertificatePackStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
