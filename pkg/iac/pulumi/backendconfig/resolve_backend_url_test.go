package backendconfig

import (
	"testing"

	awsvpcv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsvpc/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumiannotationkeys"
	"github.com/plantonhq/planton/shared"
	"github.com/stretchr/testify/assert"
)

func manifestWithBackendUrlAnnotation(url string) *awsvpcv1alpha1.AwsVpc {
	annotations := map[string]string{}
	if url != "" {
		annotations[pulumiannotationkeys.BackendUrlAnnotationKey] = url
	}
	return &awsvpcv1alpha1.AwsVpc{
		Metadata: &shared.CatalogObjectMetadata{Annotations: annotations},
	}
}

func TestResolveBackendURL_FlagBeatsAnnotationAndEnv(t *testing.T) {
	t.Setenv(BackendUrlEnvVar, "s3://from-env")
	manifest := manifestWithBackendUrlAnnotation("s3://from-annotation")

	url, source := ResolveBackendURL(manifest, "s3://from-flag")

	assert.Equal(t, "s3://from-flag", url)
	assert.Equal(t, "flag", source)
}

func TestResolveBackendURL_AnnotationBeatsEnv(t *testing.T) {
	t.Setenv(BackendUrlEnvVar, "s3://from-env")
	manifest := manifestWithBackendUrlAnnotation("s3://from-annotation")

	url, source := ResolveBackendURL(manifest, "")

	assert.Equal(t, "s3://from-annotation", url)
	assert.Equal(t, "manifest annotation", source)
}

func TestResolveBackendURL_EnvIsTheLastLayer(t *testing.T) {
	t.Setenv(BackendUrlEnvVar, "gs://from-env")
	manifest := manifestWithBackendUrlAnnotation("")

	url, source := ResolveBackendURL(manifest, "")

	assert.Equal(t, "gs://from-env", url)
	assert.Contains(t, source, BackendUrlEnvVar)
}

func TestResolveBackendURL_NothingConfiguredMeansAmbientLogin(t *testing.T) {
	manifest := manifestWithBackendUrlAnnotation("")

	url, source := ResolveBackendURL(manifest, "")

	assert.Empty(t, url)
	assert.Empty(t, source)
}
