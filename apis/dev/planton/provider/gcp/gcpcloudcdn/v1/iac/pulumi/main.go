// Package main provides the Pulumi program entrypoint for GCP Cloud CDN deployment.
// Binary releases are gzip-compressed to reduce download size.
// Auto-release test: Multi-provider Pulumi change (GCP component).
package main

import (
	"github.com/pkg/errors"
	gcpcloudcdnv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudcdn/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudcdn/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpcloudcdnv1.GcpCloudCdnStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
