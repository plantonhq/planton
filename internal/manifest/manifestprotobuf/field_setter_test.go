package manifestprotobuf

import (
	kubernetesredisv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesredis/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestSetProtoField(t *testing.T) {
	tests := []struct {
		name      string
		message   proto.Message
		fieldPath string
		value     interface{}
		expected  proto.Message
		expectErr bool
	}{
		{
			name: "Set existing string field in snake case",
			message: &kubernetesredisv1.KubernetesRedis{Spec: &kubernetesredisv1.KubernetesRedisSpec{
				Container: &kubernetesredisv1.KubernetesRedisContainer{
					DiskSize: "1Gi",
				},
			}},
			fieldPath: "spec.container.disk_size",
			value:     "2Gi",
			expected: &kubernetesredisv1.KubernetesRedis{Spec: &kubernetesredisv1.KubernetesRedisSpec{
				Container: &kubernetesredisv1.KubernetesRedisContainer{
					DiskSize: "2Gi",
				},
			}},
			expectErr: false,
		},
		{
			name: "Set existing string field in camel case",
			message: &kubernetesredisv1.KubernetesRedis{Spec: &kubernetesredisv1.KubernetesRedisSpec{
				Container: &kubernetesredisv1.KubernetesRedisContainer{
					DiskSize: "1Gi",
				},
			}},
			fieldPath: "spec.container.diskSize",
			value:     "2Gi",
			expected: &kubernetesredisv1.KubernetesRedis{Spec: &kubernetesredisv1.KubernetesRedisSpec{
				Container: &kubernetesredisv1.KubernetesRedisContainer{
					DiskSize: "2Gi",
				},
			}},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetProtoField(tt.message, tt.fieldPath, tt.value)
			if tt.expectErr {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
				assert.True(t, proto.Equal(tt.expected, result), "Expected %v but got %v", tt.expected, result)
			}
		})
	}
}
