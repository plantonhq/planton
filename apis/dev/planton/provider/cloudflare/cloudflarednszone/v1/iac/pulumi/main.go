// Package main provides the Pulumi program entrypoint for Cloudflare DNS Zone deployment.
// Binary releases are gzip-compressed to reduce download size.
package main

import (
	"github.com/pkg/errors"
	cloudflarednszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarednszone/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarednszone/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &cloudflarednszonev1.CloudflareDnsZoneStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
