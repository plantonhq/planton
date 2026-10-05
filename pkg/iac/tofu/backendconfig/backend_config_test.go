package backendconfig

import (
	"testing"

	awsvpcv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsvpc/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofuannotationkeys"
	"github.com/plantonhq/planton/shared"
	"github.com/stretchr/testify/assert"
)

func TestExtractFromManifest_TerraformProvisioner(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *awsvpcv1alpha1.AwsVpc
		want      *TofuBackendConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid s3 backend with terraform annotations",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"):   "s3",
						tofuannotationkeys.BackendBucketAnnotationKey("terraform"): "my-terraform-state",
						tofuannotationkeys.BackendKeyAnnotationKey("terraform"):    "aws-vpc/dev/terraform.tfstate",
						tofuannotationkeys.BackendRegionAnnotationKey("terraform"): "us-west-2",
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
			name: "valid gcs backend with terraform annotations",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"):   "gcs",
						tofuannotationkeys.BackendBucketAnnotationKey("terraform"): "my-gcs-bucket",
						tofuannotationkeys.BackendKeyAnnotationKey("terraform"):    "terraform/state",
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
			name: "valid azurerm backend with terraform annotations",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"):   "azurerm",
						tofuannotationkeys.BackendBucketAnnotationKey("terraform"): "my-container",
						tofuannotationkeys.BackendKeyAnnotationKey("terraform"):    "terraform/state",
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
			name: "valid local backend with terraform annotations",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"): "local",
						tofuannotationkeys.BackendKeyAnnotationKey("terraform"):  "/tmp/terraform.tfstate",
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
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"):     "s3",
						tofuannotationkeys.BackendBucketAnnotationKey("terraform"):   "my-r2-bucket",
						tofuannotationkeys.BackendKeyAnnotationKey("terraform"):      "state.tfstate",
						tofuannotationkeys.BackendRegionAnnotationKey("terraform"):   "auto",
						tofuannotationkeys.BackendEndpointAnnotationKey("terraform"): "https://account.r2.cloudflarestorage.com",
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
			name: "no backend annotations - returns nil without error",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						"other.annotation": "value",
					},
				},
			},
			want:      nil,
			wantError: false,
		},
		{
			name: "backend keys in labels are ignored (labels are cloud-tag territory)",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Labels: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"):   "s3",
						tofuannotationkeys.BackendBucketAnnotationKey("terraform"): "my-bucket",
						tofuannotationkeys.BackendKeyAnnotationKey("terraform"):    "some/path",
					},
				},
			},
			want:      nil,
			wantError: false,
		},
		{
			name: "missing backend key - returns partial config",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"):   "s3",
						tofuannotationkeys.BackendBucketAnnotationKey("terraform"): "my-bucket",
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
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"):   "unsupported",
						tofuannotationkeys.BackendBucketAnnotationKey("terraform"): "bucket",
						tofuannotationkeys.BackendKeyAnnotationKey("terraform"):    "some/path",
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
			name: "no annotations - returns nil without error",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{},
			},
			want:      nil,
			wantError: false,
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
		manifest  *awsvpcv1alpha1.AwsVpc
		want      *TofuBackendConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid s3 backend with tofu annotations",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("tofu"):   "s3",
						tofuannotationkeys.BackendBucketAnnotationKey("tofu"): "my-tofu-state",
						tofuannotationkeys.BackendKeyAnnotationKey("tofu"):    "aws-vpc/dev/terraform.tfstate",
						tofuannotationkeys.BackendRegionAnnotationKey("tofu"): "us-east-1",
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
			name: "valid gcs backend with tofu annotations",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("tofu"):   "gcs",
						tofuannotationkeys.BackendBucketAnnotationKey("tofu"): "my-gcs-bucket",
						tofuannotationkeys.BackendKeyAnnotationKey("tofu"):    "tofu/state",
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
			name: "terraform-prefixed annotations are NOT read for the tofu provisioner",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						// The key prefix must match the provisioner; there is no cross-prefix fallback.
						tofuannotationkeys.BackendTypeAnnotationKey("terraform"):   "s3",
						tofuannotationkeys.BackendBucketAnnotationKey("terraform"): "terraform-bucket",
						tofuannotationkeys.BackendKeyAnnotationKey("terraform"):    "state.tfstate",
					},
				},
			},
			want:      nil,
			wantError: false,
		},
		{
			name: "missing backend key with tofu annotations - returns partial config",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						tofuannotationkeys.BackendTypeAnnotationKey("tofu"):   "s3",
						tofuannotationkeys.BackendBucketAnnotationKey("tofu"): "my-bucket",
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
