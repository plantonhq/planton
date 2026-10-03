package metadatareflect

import (
	"testing"

	awss3bucketv1alpha1 "github.com/plantonhq/planton/catalog/aws/awss3bucket/v1alpha1"

	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestExtractMetadata(t *testing.T) {
	tests := []struct {
		name  string
		input proto.Message
		want  *shared.CatalogObjectMetadata
	}{
		{
			name: "when metadata is set should return the metadata from input",
			input: &awss3bucketv1alpha1.AwsS3Bucket{
				Metadata: &shared.CatalogObjectMetadata{
					Id: "test-id",
				},
			},
			want: &shared.CatalogObjectMetadata{Id: "test-id"},
		}, {
			name: "when metadata object is empty in input, should return empty metadata object",
			input: &awss3bucketv1alpha1.AwsS3Bucket{
				Metadata: &shared.CatalogObjectMetadata{},
			},
			want: &shared.CatalogObjectMetadata{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractMetadata(tt.input)
			if !proto.Equal(got, tt.want) {
				t.Errorf("Extractmetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
