package backendconfig

import (
	"testing"

	awsvpcv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsvpc/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumiannotationkeys"
	"github.com/plantonhq/planton/shared"
	"github.com/stretchr/testify/assert"
)

func TestExtractFromManifest(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *awsvpcv1alpha1.AwsVpc
		want      *PulumiBackendConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "stack.fqdn takes precedence",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						pulumiannotationkeys.StackFqdnAnnotationKey:    "demo-org/aws-examples/dev",
						pulumiannotationkeys.OrganizationAnnotationKey: "should-be-ignored",
						pulumiannotationkeys.ProjectAnnotationKey:      "should-be-ignored",
						pulumiannotationkeys.StackNameAnnotationKey:    "should-be-ignored",
					},
				},
			},
			want: &PulumiBackendConfig{
				StackFqdn:    "demo-org/aws-examples/dev",
				Organization: "demo-org",
				Project:      "aws-examples",
				StackName:    "dev",
			},
			wantError: false,
		},
		{
			name: "individual annotations when stack.fqdn not present",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						pulumiannotationkeys.OrganizationAnnotationKey: "my-org",
						pulumiannotationkeys.ProjectAnnotationKey:      "my-project",
						pulumiannotationkeys.StackNameAnnotationKey:    "production",
					},
				},
			},
			want: &PulumiBackendConfig{
				StackFqdn:    "my-org/my-project/production",
				Organization: "my-org",
				Project:      "my-project",
				StackName:    "production",
			},
			wantError: false,
		},
		{
			name: "invalid stack.fqdn format",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						pulumiannotationkeys.StackFqdnAnnotationKey: "invalid-format",
					},
				},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "invalid stack.fqdn format",
		},
		{
			name: "missing required annotations",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						pulumiannotationkeys.OrganizationAnnotationKey: "my-org",
						pulumiannotationkeys.ProjectAnnotationKey:      "my-project",
						// Missing stack name
					},
				},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "missing required Pulumi backend annotations",
		},
		{
			name: "empty annotation values",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						pulumiannotationkeys.OrganizationAnnotationKey: "my-org",
						pulumiannotationkeys.ProjectAnnotationKey:      "",
						pulumiannotationkeys.StackNameAnnotationKey:    "dev",
					},
				},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "Pulumi backend annotations cannot be empty",
		},
		{
			name: "no annotations",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "no annotations found in manifest",
		},
		{
			name: "empty stack.fqdn components",
			manifest: &awsvpcv1alpha1.AwsVpc{
				Metadata: &shared.CatalogObjectMetadata{
					Annotations: map[string]string{
						pulumiannotationkeys.StackFqdnAnnotationKey: "org//stack", // Missing project
					},
				},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "stack FQDN components cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractFromManifest(tt.manifest)

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

func TestParseStackFqdn(t *testing.T) {
	tests := []struct {
		name      string
		fqdn      string
		wantOrg   string
		wantProj  string
		wantStack string
		wantError bool
	}{
		{
			name:      "valid fqdn",
			fqdn:      "my-org/my-project/my-stack",
			wantOrg:   "my-org",
			wantProj:  "my-project",
			wantStack: "my-stack",
			wantError: false,
		},
		{
			name:      "valid fqdn with spaces",
			fqdn:      " my-org / my-project / my-stack ",
			wantOrg:   "my-org",
			wantProj:  "my-project",
			wantStack: "my-stack",
			wantError: false,
		},
		{
			name:      "too few parts",
			fqdn:      "my-org/my-project",
			wantError: true,
		},
		{
			name:      "too many parts",
			fqdn:      "my-org/my-project/my-stack/extra",
			wantError: true,
		},
		{
			name:      "empty component",
			fqdn:      "my-org//my-stack",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, proj, stack, err := parseStackFqdn(tt.fqdn)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOrg, org)
				assert.Equal(t, tt.wantProj, proj)
				assert.Equal(t, tt.wantStack, stack)
			}
		})
	}
}
