package backendconfig

import (
	"testing"

	awsvpcv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsvpc/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofulabels"
	"github.com/stretchr/testify/assert"
)

func TestExtractFromManifest_TerraformProvisioner(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *awsvpcv1.AwsVpc
		want      *TofuBackendConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid s3 backend with terraform labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("terraform"):   "s3",
						tofulabels.BackendBucketLabelKey("terraform"): "my-terraform-state",
						tofulabels.BackendKeyLabelKey("terraform"):    "aws-vpc/dev/terraform.tfstate",
						tofulabels.BackendRegionLabelKey("terraform"): "us-west-2",
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:   "s3",
				BackendBucket: "my-terraform-state",
				BackendKey:    "aws-vpc/dev/terraform.tfstate",
				BackendRegion: "us-west-2",
			},
			wantError: false,
		},
		{
			name: "valid gcs backend with terraform labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("terraform"):   "gcs",
						tofulabels.BackendBucketLabelKey("terraform"): "my-gcs-bucket",
						tofulabels.BackendKeyLabelKey("terraform"):    "terraform/state",
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:   "gcs",
				BackendBucket: "my-gcs-bucket",
				BackendKey:    "terraform/state",
			},
			wantError: false,
		},
		{
			name: "valid azurerm backend with terraform labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("terraform"):   "azurerm",
						tofulabels.BackendBucketLabelKey("terraform"): "my-container",
						tofulabels.BackendKeyLabelKey("terraform"):    "terraform/state",
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:   "azurerm",
				BackendBucket: "my-container",
				BackendKey:    "terraform/state",
			},
			wantError: false,
		},
		{
			name: "valid local backend with terraform labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("terraform"): "local",
						tofulabels.BackendKeyLabelKey("terraform"):  "/tmp/terraform.tfstate",
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType: "local",
				BackendKey:  "/tmp/terraform.tfstate",
			},
			wantError: false,
		},
		{
			name: "s3-compatible backend with endpoint",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("terraform"):     "s3",
						tofulabels.BackendBucketLabelKey("terraform"):   "my-r2-bucket",
						tofulabels.BackendKeyLabelKey("terraform"):      "state.tfstate",
						tofulabels.BackendRegionLabelKey("terraform"):   "auto",
						tofulabels.BackendEndpointLabelKey("terraform"): "https://account.r2.cloudflarestorage.com",
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:     "s3",
				BackendBucket:   "my-r2-bucket",
				BackendKey:      "state.tfstate",
				BackendRegion:   "auto",
				BackendEndpoint: "https://account.r2.cloudflarestorage.com",
				S3Compatible:    true,
			},
			wantError: false,
		},
		{
			name: "no backend labels - returns nil without error",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						"other.label": "value",
					},
				},
			},
			want:      nil,
			wantError: false,
		},
		{
			name: "missing backend key - returns partial config",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("terraform"):   "s3",
						tofulabels.BackendBucketLabelKey("terraform"): "my-bucket",
						// Missing backend key - pure extraction returns partial config
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:   "s3",
				BackendBucket: "my-bucket",
			},
			wantError: false,
		},
		{
			name: "unsupported backend type - returns config (validation happens later)",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("terraform"):   "unsupported",
						tofulabels.BackendBucketLabelKey("terraform"): "bucket",
						tofulabels.BackendKeyLabelKey("terraform"):    "some/path",
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:   "unsupported",
				BackendBucket: "bucket",
				BackendKey:    "some/path",
			},
			wantError: false,
		},
		{
			name: "no labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "no labels found in manifest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractFromManifest(tt.manifest, "terraform")

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExtractFromManifest_TofuProvisioner(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *awsvpcv1.AwsVpc
		want      *TofuBackendConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid s3 backend with tofu labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("tofu"):   "s3",
						tofulabels.BackendBucketLabelKey("tofu"): "my-tofu-state",
						tofulabels.BackendKeyLabelKey("tofu"):    "aws-vpc/dev/terraform.tfstate",
						tofulabels.BackendRegionLabelKey("tofu"): "us-east-1",
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:   "s3",
				BackendBucket: "my-tofu-state",
				BackendKey:    "aws-vpc/dev/terraform.tfstate",
				BackendRegion: "us-east-1",
			},
			wantError: false,
		},
		{
			name: "valid gcs backend with tofu labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("tofu"):   "gcs",
						tofulabels.BackendBucketLabelKey("tofu"): "my-gcs-bucket",
						tofulabels.BackendKeyLabelKey("tofu"):    "tofu/state",
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:   "gcs",
				BackendBucket: "my-gcs-bucket",
				BackendKey:    "tofu/state",
			},
			wantError: false,
		},
		{
			name: "missing backend key with tofu labels - returns partial config",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						tofulabels.BackendTypeLabelKey("tofu"):   "s3",
						tofulabels.BackendBucketLabelKey("tofu"): "my-bucket",
						// Missing backend key - pure extraction returns partial config
					},
				},
			},
			want: &TofuBackendConfig{
				BackendType:   "s3",
				BackendBucket: "my-bucket",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractFromManifest(tt.manifest, "tofu")

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExtractFromManifest_LegacyFallback(t *testing.T) {
	tests := []struct {
		name            string
		manifest        *awsvpcv1.AwsVpc
		provisionerType string
		want            *TofuBackendConfig
		wantError       bool
		errorMsg        string
	}{
		{
			name: "tofu provisioner falls back to legacy terraform labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						// Using legacy terraform.* labels
						tofulabels.LegacyBackendTypeLabelKey:   "s3",
						tofulabels.LegacyBackendBucketLabelKey: "legacy-bucket",
						tofulabels.LegacyBackendKeyLabelKey:    "aws-vpc/prod/state.tfstate",
						tofulabels.LegacyBackendRegionLabelKey: "us-west-2",
					},
				},
			},
			provisionerType: "tofu",
			want: &TofuBackendConfig{
				BackendType:   "s3",
				BackendBucket: "legacy-bucket",
				BackendKey:    "aws-vpc/prod/state.tfstate",
				BackendRegion: "us-west-2",
			},
			wantError: false,
		},
		{
			name: "terraform provisioner uses terraform labels directly (same as legacy)",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						// terraform.* labels are both provisioner-specific AND legacy
						tofulabels.LegacyBackendTypeLabelKey:   "gcs",
						tofulabels.LegacyBackendBucketLabelKey: "terraform-bucket",
						tofulabels.LegacyBackendKeyLabelKey:    "aws-vpc/staging",
					},
				},
			},
			provisionerType: "terraform",
			want: &TofuBackendConfig{
				BackendType:   "gcs",
				BackendBucket: "terraform-bucket",
				BackendKey:    "aws-vpc/staging",
			},
			wantError: false,
		},
		{
			name: "provisioner-specific labels take precedence over legacy",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						// Both tofu.* and terraform.* labels present
						tofulabels.BackendTypeLabelKey("tofu"):   "s3",
						tofulabels.BackendBucketLabelKey("tofu"): "tofu-specific-bucket",
						tofulabels.BackendKeyLabelKey("tofu"):    "tofu-state.tfstate",
						tofulabels.LegacyBackendTypeLabelKey:     "gcs",
						tofulabels.LegacyBackendBucketLabelKey:   "legacy-bucket",
						tofulabels.LegacyBackendKeyLabelKey:      "legacy-state",
					},
				},
			},
			provisionerType: "tofu",
			want: &TofuBackendConfig{
				BackendType:   "s3",
				BackendBucket: "tofu-specific-bucket",
				BackendKey:    "tofu-state.tfstate",
			},
			wantError: false,
		},
		{
			name: "legacy fallback with partial labels returns partial config",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						// Only type and bucket, missing key - pure extraction returns partial config
						tofulabels.LegacyBackendTypeLabelKey:   "s3",
						tofulabels.LegacyBackendBucketLabelKey: "my-bucket",
					},
				},
			},
			provisionerType: "tofu",
			want: &TofuBackendConfig{
				BackendType:   "s3",
				BackendBucket: "my-bucket",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractFromManifest(tt.manifest, tt.provisionerType)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIsS3Compatible(t *testing.T) {
	tests := []struct {
		name   string
		config *TofuBackendConfig
		want   bool
	}{
		{
			name: "endpoint set - S3 compatible",
			config: &TofuBackendConfig{
				BackendType:     "s3",
				BackendEndpoint: "https://account.r2.cloudflarestorage.com",
			},
			want: true,
		},
		{
			name: "region auto - S3 compatible",
			config: &TofuBackendConfig{
				BackendType:   "s3",
				BackendRegion: "auto",
			},
			want: true,
		},
		{
			name: "both endpoint and auto region - S3 compatible",
			config: &TofuBackendConfig{
				BackendType:     "s3",
				BackendRegion:   "auto",
				BackendEndpoint: "https://account.r2.cloudflarestorage.com",
			},
			want: true,
		},
		{
			name: "standard AWS S3 - not S3 compatible",
			config: &TofuBackendConfig{
				BackendType:   "s3",
				BackendRegion: "us-west-2",
			},
			want: false,
		},
		{
			name: "GCS backend - not S3 compatible",
			config: &TofuBackendConfig{
				BackendType: "gcs",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsS3Compatible()
			assert.Equal(t, tt.want, got)
		})
	}
}
