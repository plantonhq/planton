package backendconfig

import (
	"testing"

	awsvpcv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsvpc/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumilabels"
	"github.com/stretchr/testify/assert"
)

func TestExtractFromManifest(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *awsvpcv1.AwsVpc
		want      *PulumiBackendConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "stack.fqdn takes precedence",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						pulumilabels.StackFqdnLabelKey:    "demo-org/aws-examples/dev",
						pulumilabels.OrganizationLabelKey: "should-be-ignored",
						pulumilabels.ProjectLabelKey:      "should-be-ignored",
						pulumilabels.StackNameLabelKey:    "should-be-ignored",
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
			name: "individual labels when stack.fqdn not present",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						pulumilabels.OrganizationLabelKey: "my-org",
						pulumilabels.ProjectLabelKey:      "my-project",
						pulumilabels.StackNameLabelKey:    "production",
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
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						pulumilabels.StackFqdnLabelKey: "invalid-format",
					},
				},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "invalid stack.fqdn format",
		},
		{
			name: "missing required labels",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						pulumilabels.OrganizationLabelKey: "my-org",
						pulumilabels.ProjectLabelKey:      "my-project",
						// Missing stack name
					},
				},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "missing required Pulumi backend labels",
		},
		{
			name: "empty label values",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						pulumilabels.OrganizationLabelKey: "my-org",
						pulumilabels.ProjectLabelKey:      "",
						pulumilabels.StackNameLabelKey:    "dev",
					},
				},
			},
			want:      nil,
			wantError: true,
			errorMsg:  "Pulumi backend labels cannot be empty",
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
		{
			name: "empty stack.fqdn components",
			manifest: &awsvpcv1.AwsVpc{
				Metadata: &shared.CloudResourceMetadata{
					Labels: map[string]string{
						pulumilabels.StackFqdnLabelKey: "org//stack", // Missing project
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
