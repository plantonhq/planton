package module

import (
	"strings"
	"testing"
)

// The per-build images are the ones a person mirrors or allow-lists before a
// build can start: Tekton injects them into every TaskRun pod, so a cluster
// that cannot pull one fails every build. The outputs name each exactly as
// the cluster pulls it -- at the mirror when image_registry is set, else on
// ghcr.io -- with its digest. The Terraform module builds the same values
// from the same table (iac/tf/outputs.tf).

// perBuildOutputs maps each per-build output to the image-table entry it names.
var perBuildOutputs = map[string]string{
	OpEntrypointImage:        "IMAGE_PIPELINES_ARG__ENTRYPOINT_IMAGE",
	OpNopImage:               "IMAGE_PIPELINES_ARG__NOP_IMAGE",
	OpWorkingdirinitImage:    "IMAGE_PIPELINES_ARG__WORKINGDIRINIT_IMAGE",
	OpSidecarlogresultsImage: "IMAGE_PIPELINES_ARG__SIDECARLOGRESULTS_IMAGE",
}

func tableImage(t *testing.T, table *tektonImages, name string) string {
	t.Helper()
	for _, entry := range table.Images {
		if entry.Name == name {
			return entry.Image
		}
	}
	t.Fatalf("the image table has no %s entry", name)
	return ""
}

func TestOutputsNameThePerBuildImagesAtTheMirror(t *testing.T) {
	table, err := loadImageTable()
	if err != nil {
		t.Fatal(err)
	}
	outputs, err := buildOutputs(table, mirror)
	if err != nil {
		t.Fatal(err)
	}
	if got := outputs[OpImageRegistry]; got != mirror {
		t.Fatalf("%s = %q, want the mirror %q", OpImageRegistry, got, mirror)
	}
	for output, entry := range perBuildOutputs {
		upstream := tableImage(t, table, entry)
		want := mirror + strings.TrimPrefix(upstream, "ghcr.io")
		if got := outputs[output]; got != want {
			t.Fatalf("%s = %q, want %q", output, got, want)
		}
		if !strings.Contains(outputs[output], "@sha256:") {
			t.Fatalf("%s = %q carries no digest", output, outputs[output])
		}
	}
}

func TestOutputsNameGhcrWithoutARegistry(t *testing.T) {
	table, err := loadImageTable()
	if err != nil {
		t.Fatal(err)
	}
	outputs, err := buildOutputs(table, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := outputs[OpImageRegistry]; got != "ghcr.io" {
		t.Fatalf("%s = %q, want ghcr.io", OpImageRegistry, got)
	}
	for output, entry := range perBuildOutputs {
		if got, want := outputs[output], tableImage(t, table, entry); got != want {
			t.Fatalf("%s = %q, want %q", output, got, want)
		}
	}
}

func TestOutputsRefuseATableWithoutAPerBuildImage(t *testing.T) {
	table, err := loadImageTable()
	if err != nil {
		t.Fatal(err)
	}
	var kept []tektonImage
	for _, entry := range table.Images {
		if entry.Name != "IMAGE_PIPELINES_ARG__NOP_IMAGE" {
			kept = append(kept, entry)
		}
	}
	table.Images = kept
	if _, err := buildOutputs(table, ""); err == nil || !strings.Contains(err.Error(), "IMAGE_PIPELINES_ARG__NOP_IMAGE") {
		t.Fatalf("buildOutputs without the nop entry returned %v, want an error naming it", err)
	}
}
