package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesTektonOperatorOutputs
// field.
const (
	OpNamespace              = "namespace"
	OpImageRegistry          = "image_registry"
	OpEntrypointImage        = "entrypoint_image"
	OpNopImage               = "nop_image"
	OpWorkingdirinitImage    = "workingdirinit_image"
	OpSidecarlogresultsImage = "sidecarlogresults_image"
)

// perBuildImages names, per output, the image-table entry of an image Tekton
// injects into every TaskRun pod. A cluster that cannot pull one of them
// fails every build, so they are the images a person mirrors or allow-lists
// first.
var perBuildImages = []struct{ output, entry string }{
	{OpEntrypointImage, "IMAGE_PIPELINES_ARG__ENTRYPOINT_IMAGE"},
	{OpNopImage, "IMAGE_PIPELINES_ARG__NOP_IMAGE"},
	{OpWorkingdirinitImage, "IMAGE_PIPELINES_ARG__WORKINGDIRINIT_IMAGE"},
	{OpSidecarlogresultsImage, "IMAGE_PIPELINES_ARG__SIDECARLOGRESULTS_IMAGE"},
}

// exportOutputs publishes the release manifest's fixed handles and the
// images the cluster pulls (the final apply group is passed so the export
// carries its dependency). The image table is read whether or not
// image_registry is set: without one the outputs name ghcr.io.
func exportOutputs(ctx *pulumi.Context, locals *Locals, _ pulumi.Resource) error {
	table, err := loadImageTable()
	if err != nil {
		return err
	}
	outputs, err := buildOutputs(table, locals.Spec.GetImageRegistry())
	if err != nil {
		return err
	}
	ctx.Export(OpNamespace, pulumi.String(vars.Namespace))
	for _, name := range []string{OpImageRegistry, OpEntrypointImage, OpNopImage, OpWorkingdirinitImage, OpSidecarlogresultsImage} {
		ctx.Export(name, pulumi.String(outputs[name]))
	}
	return nil
}

// buildOutputs resolves the image outputs from the table: the registry every
// Tekton component image is pulled from (the mirror when set, else ghcr.io)
// and each per-build image as the cluster pulls it, digest included. A table
// missing a per-build entry is refused rather than exported empty. Terraform
// twin: iac/tf/outputs.tf.
func buildOutputs(table *tektonImages, registry string) (map[string]string, error) {
	outputs := map[string]string{OpImageRegistry: upstreamRegistry}
	if registry != "" {
		outputs[OpImageRegistry] = registry
	}
	byName := make(map[string]string, len(table.Images))
	for _, entry := range table.Images {
		byName[entry.Name] = entry.Image
	}
	for _, image := range perBuildImages {
		upstream, ok := byName[image.entry]
		if !ok {
			return nil, errors.Errorf("the module's Tekton image table has no %s entry, so the %s output cannot name the image every build pulls; rebuild the table (see images.go)", image.entry, image.output)
		}
		outputs[image.output] = mirroredImage(upstream, registry)
	}
	return outputs, nil
}
