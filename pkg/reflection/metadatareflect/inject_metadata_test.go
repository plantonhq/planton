package metadatareflect

import (
	"testing"

	awss3bucketv1alpha1 "github.com/plantonhq/planton/catalog/aws/awss3bucket/v1alpha1"

	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestInjectMetadata(t *testing.T) {
	tests := []struct {
		name  string
		input proto.Message
		meta  *shared.CatalogObjectMetadata
		want  proto.Message
	}{
		{
			name:  "when metadata is injected it should appear in the output",
			input: &awss3bucketv1alpha1.AwsS3Bucket{},
			meta:  &shared.CatalogObjectMetadata{Id: "test-id"},
			want: &awss3bucketv1alpha1.AwsS3Bucket{
				Metadata: &shared.CatalogObjectMetadata{Id: "test-id"},
			},
		},
		{
			name:  "when meta is nil the message must stay unchanged",
			input: &awss3bucketv1alpha1.AwsS3Bucket{},
			meta:  nil,
			want:  &awss3bucketv1alpha1.AwsS3Bucket{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InjectMetadata(proto.Clone(tt.input), tt.meta) // clone to avoid mutating test data
			if !proto.Equal(got, tt.want) {
				t.Errorf("InjectMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
