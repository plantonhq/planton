package catalogkindreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

func TestKindByIdPrefix(t *testing.T) {
	tests := []struct {
		name     string
		idPrefix string
		want     catalogkind.CatalogKind
		wantErr  bool
	}{
		{
			name:     "AWS ECS Service",
			idPrefix: "awsecss",
			want:     catalogkind.CatalogKind_AwsEcsService,
			wantErr:  false,
		},
		{
			name:     "GCP GKE Cluster",
			idPrefix: "gcpgke",
			want:     catalogkind.CatalogKind_GcpGkeCluster,
			wantErr:  false,
		},
		{
			name:     "Azure AKS Cluster",
			idPrefix: "aks",
			want:     catalogkind.CatalogKind_AzureAksCluster,
			wantErr:  false,
		},
		{
			name:     "Kubernetes Deployment",
			idPrefix: "k8sdpl",
			want:     catalogkind.CatalogKind_KubernetesDeployment,
			wantErr:  false,
		},
		{
			name:     "Invalid prefix",
			idPrefix: "invalid",
			want:     catalogkind.CatalogKind_unspecified,
			wantErr:  true,
		},
		{
			name:     "Empty prefix",
			idPrefix: "",
			want:     catalogkind.CatalogKind_unspecified,
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
