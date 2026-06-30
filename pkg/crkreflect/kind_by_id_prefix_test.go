package crkreflect

import (
	"testing"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

func TestKindByIdPrefix(t *testing.T) {
	tests := []struct {
		name     string
		idPrefix string
		want     cloudresourcekind.CloudResourceKind
		wantErr  bool
	}{
		{
			name:     "AWS ECS Service",
			idPrefix: "ecssvc",
			want:     cloudresourcekind.CloudResourceKind_AwsEcsService,
			wantErr:  false,
		},
		{
			name:     "GCP GKE Cluster",
			idPrefix: "gke",
			want:     cloudresourcekind.CloudResourceKind_GcpGkeCluster,
			wantErr:  false,
		},
		{
			name:     "Azure AKS Cluster",
			idPrefix: "aks",
			want:     cloudresourcekind.CloudResourceKind_AzureAksCluster,
			wantErr:  false,
		},
		{
			name:     "Kubernetes Deployment",
			idPrefix: "k8sdpl",
			want:     cloudresourcekind.CloudResourceKind_KubernetesDeployment,
			wantErr:  false,
		},
		{
			name:     "Invalid prefix",
			idPrefix: "invalid",
			want:     cloudresourcekind.CloudResourceKind_unspecified,
			wantErr:  true,
		},
		{
			name:     "Empty prefix",
			idPrefix: "",
			want:     cloudresourcekind.CloudResourceKind_unspecified,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KindByIdPrefix(tt.idPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("KindByIdPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KindByIdPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
