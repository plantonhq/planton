package catalogkindreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

func TestKindFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected catalogkind.CatalogKind
	}{
		// KubernetesDeployment tests
		{
			name:     "KubernetesDeployment - PascalCase",
			input:    "KubernetesDeployment",
			expected: catalogkind.CatalogKind_KubernetesDeployment,
		},
		{
			name:     "KubernetesDeployment - kebab-case",
			input:    "kubernetes-deployment",
			expected: catalogkind.CatalogKind_KubernetesDeployment,
		},
		{
			name:     "KubernetesDeployment - snake_case",
			input:    "kubernetes_deployment",
			expected: catalogkind.CatalogKind_KubernetesDeployment,
		},
		{
			name:     "KubernetesDeployment - UPPERCASE",
			input:    "KUBERNETESDEPLOYMENT",
			expected: catalogkind.CatalogKind_KubernetesDeployment,
		},
		{
			name:     "KubernetesDeployment - mixed case with hyphens",
			input:    "Kubernetes-Deployment",
			expected: catalogkind.CatalogKind_KubernetesDeployment,
		},
		{
			name:     "KubernetesDeployment - mixed case with underscores",
			input:    "Kubernetes_Deployment",
			expected: catalogkind.CatalogKind_KubernetesDeployment,
		},

		// AwsEcsService tests
		{
			name:     "AwsEcsService - PascalCase",
			input:    "AwsEcsService",
			expected: catalogkind.CatalogKind_AwsEcsService,
		},
		{
			name:     "AwsEcsService - kebab-case",
			input:    "aws-ecs-service",
			expected: catalogkind.CatalogKind_AwsEcsService,
		},
		{
			name:     "AwsEcsService - snake_case",
			input:    "aws_ecs_service",
			expected: catalogkind.CatalogKind_AwsEcsService,
		},
		{
			name:     "AwsEcsService - UPPERCASE",
			input:    "AWSECSSERVICE",
			expected: catalogkind.CatalogKind_AwsEcsService,
		},

		// GcpGkeCluster tests
		{
			name:     "GcpGkeCluster - PascalCase",
			input:    "GcpGkeCluster",
			expected: catalogkind.CatalogKind_GcpGkeCluster,
		},
		{
			name:     "GcpGkeCluster - kebab-case",
			input:    "gcp-gke-cluster",
			expected: catalogkind.CatalogKind_GcpGkeCluster,
		},
		{
			name:     "GcpGkeCluster - snake_case",
			input:    "gcp_gke_cluster",
			expected: catalogkind.CatalogKind_GcpGkeCluster,
		},

		// Invalid input tests
		{
			name:     "Empty string",
			input:    "",
			expected: catalogkind.CatalogKind_unspecified,
		},
		{
			name:     "Unknown resource",
			input:    "UnknownResource",
			expected: catalogkind.CatalogKind_unspecified,
		},
		{
			name:     "Random string",
			input:    "random-string-123",
			expected: catalogkind.CatalogKind_unspecified,
		},
	}

	// Also test if aliases in AliasMap work (if any are defined)
	// Note: This test assumes AliasMap might be populated elsewhere
	if len(AliasMap) > 0 {
		for kind, aliases := range AliasMap {
			for _, alias := range aliases {
				tests = append(tests, struct {
					name     string
					input    string
					expected catalogkind.CatalogKind
				}{
					name:     "Alias for " + kind.String() + ": " + alias,
					input:    alias,
					expected: kind,
				})
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KindFromString(tt.input)
			if result != tt.expected {
				t.Errorf("KindFromString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestKindFromStringNormalization specifically tests the normalization logic
func TestKindFromStringNormalization(t *testing.T) {
	// Test that all different representations of the same kind resolve to the same value
	testCases := [][]string{
		{
			"KubernetesDeployment",
			"kubernetesdeployment",
			"KUBERNETESDEPLOYMENT",
			"kubernetes-deployment",
			"kubernetes_deployment",
			"Kubernetes-Deployment",
			"Kubernetes_Deployment",
			"KUBERNETES-DEPLOYMENT",
			"KUBERNETES_DEPLOYMENT",
		},
		{
			"AwsEcsService",
			"awsecsservice",
			"AWSECSSERVICE",
			"aws-ecs-service",
			"aws_ecs_service",
			"Aws-Ecs-Service",
			"AWS_ECS_SERVICE",
		},
	}

	for _, variants := range testCases {
		if len(variants) == 0 {
			continue
		}

		// Get the expected kind from the first variant (PascalCase)
		expected := KindFromString(variants[0])
		if expected == catalogkind.CatalogKind_unspecified {
			t.Errorf("Failed to get expected kind for %s", variants[0])
			continue
		}

		// Test that all variants resolve to the same kind
		for _, variant := range variants {
			result := KindFromString(variant)
			if result != expected {
				t.Errorf("KindFromString(%q) = %v, want %v (same as %q)",
					variant, result, expected, variants[0])
			}
		}
	}
}
