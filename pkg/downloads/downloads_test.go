package downloads

import "testing"

// These tests pin the R2 key shapes to what the release workflows upload
// (release.terraform-modules.yaml, release.pulumi-modules.yaml). The expected
// strings are deliberate literals: if either side changes its shape, exactly
// one of the two must fail — that failure is the contract doing its job.
// Update the workflows and these literals together, never one alone.

func TestBuildTerraformDownloadURL(t *testing.T) {
	tests := []struct {
		name    string
		kindDir string
		release string
		want    string
	}{
		{
			name:    "canonical kind",
			kindDir: "AwsEcsService",
			release: "v0.3.50",
			want:    "https://downloads.planton.dev/releases/v0.3.50/modules/terraform/awsecsservice/module.zip",
		},
		{
			// One live module set per kind: the key stays identical when
			// a kind graduates API versions -- pin that invariance.
			name:    "key carries no API version segment",
			kindDir: "AwsS3Bucket",
			release: "v0.4.0",
			want:    "https://downloads.planton.dev/releases/v0.4.0/modules/terraform/awss3bucket/module.zip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildTerraformDownloadURL(tt.kindDir, tt.release); got != tt.want {
				t.Errorf("BuildTerraformDownloadURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildPulumiDownloadURL(t *testing.T) {
	tests := []struct {
		name     string
		kindDir  string
		release  string
		platform string
		want     string
	}{
		{
			name:     "darwin arm64",
			kindDir:  "AwsEcsService",
			release:  "v0.3.50",
			platform: "darwin_arm64",
			want:     "https://downloads.planton.dev/releases/v0.3.50/modules/pulumi/awsecsservice/darwin_arm64.gz",
		},
		{
			name:     "linux amd64",
			kindDir:  "AwsS3Bucket",
			release:  "v0.3.50",
			platform: "linux_amd64",
			want:     "https://downloads.planton.dev/releases/v0.3.50/modules/pulumi/awss3bucket/linux_amd64.gz",
		},
		{
			// Every ARM runner the platform ships (Graviton, kind on an Apple
			// Silicon Mac) asks for this shape.
			name:     "linux arm64",
			kindDir:  "KubernetesDeployment",
			release:  "v0.3.50",
			platform: "linux_arm64",
			want:     "https://downloads.planton.dev/releases/v0.3.50/modules/pulumi/kubernetesdeployment/linux_arm64.gz",
		},
		{
			// The release lane gzips "{kind}.exe" on windows, so the
			// remote artifact carries ".exe.gz" — pin it so the windows fast
			// path cannot regress to the extensionless shape.
			name:     "windows carries the executable suffix",
			kindDir:  "AwsEcsService",
			release:  "v0.3.50",
			platform: "windows_amd64",
			want:     "https://downloads.planton.dev/releases/v0.3.50/modules/pulumi/awsecsservice/windows_amd64.exe.gz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildPulumiDownloadURL(tt.kindDir, tt.release, tt.platform); got != tt.want {
				t.Errorf("BuildPulumiDownloadURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildPulumiSourceDownloadURL(t *testing.T) {
	tests := []struct {
		name    string
		kindDir string
		release string
		want    string
	}{
		{
			name:    "canonical kind",
			kindDir: "AwsEcsService",
			release: "v0.3.50",
			want:    "https://downloads.planton.dev/releases/v0.3.50/modules/pulumi/awsecsservice/source.zip",
		},
		{
			// Source is platform-independent -- pin that the key carries no
			// platform segment (the binary key beside it does).
			name:    "key carries no platform segment",
			kindDir: "AwsS3Bucket",
			release: "v0.4.0",
			want:    "https://downloads.planton.dev/releases/v0.4.0/modules/pulumi/awss3bucket/source.zip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildPulumiSourceDownloadURL(tt.kindDir, tt.release); got != tt.want {
				t.Errorf("BuildPulumiSourceDownloadURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildDefinitionsDownloadURL(t *testing.T) {
	want := "https://downloads.planton.dev/releases/v0.4.0/definitions/definitions-manifest.json"
	if got := BuildDefinitionsDownloadURL("v0.4.0", "definitions-manifest.json"); got != want {
		t.Errorf("BuildDefinitionsDownloadURL() = %q, want %q", got, want)
	}
}

func TestBuildDefinitionsExplodedFileURL(t *testing.T) {
	want := "https://downloads.planton.dev/releases/v0.4.0/definitions/exploded/planton/references/chart-format.md"
	if got := BuildDefinitionsExplodedFileURL("v0.4.0", "planton", "references/chart-format.md"); got != want {
		t.Errorf("BuildDefinitionsExplodedFileURL() = %q, want %q", got, want)
	}
}
