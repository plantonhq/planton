package tfoverride

import (
	"os"
	"path/filepath"
	"testing"

	awsprovider "github.com/plantonhq/planton/catalog/aws"
	awsvpcv1 "github.com/plantonhq/planton/catalog/aws/awsvpc/v1alpha1"
	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	dovpcv1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanvpc/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/iacinput"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// awsIacInputYaml builds an IaC input through the production path
// (BuildIacInputYaml + BuildFromProto) so the test exercises the same
// protojson/YAML round-trip a live run takes.
func awsIacInputYaml(t *testing.T, cfg *awsprovider.AwsProviderConfig) string {
	t.Helper()
	manifest := &awsvpcv1.AwsVpc{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsVpc",
		Metadata:   &shared.CatalogObjectMetadata{Name: "override-test-vpc"},
	}
	providerConfig, cleanup, err := iacinputproviderconfig.BuildFromProto(
		cfg, catalogkind.CatalogProvider_aws)
	require.NoError(t, err)
	t.Cleanup(cleanup)
	yaml, err := iacinput.BuildIacInputYaml(manifest, providerConfig)
	require.NoError(t, err)
	return yaml
}

func fullProviderBlockConfig() *awsprovider.AwsProviderConfig {
	return &awsprovider.AwsProviderConfig{
		AccountId:       "123456789012",
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY00",
		Region:          "us-east-1",
		AssumeRoleChain: []*awsprovider.AwsAssumeRole{
			{
				RoleArn:     "arn:aws:iam::111111111111:role/intermediate",
				SessionName: "planton-hop-1",
			},
			{
				RoleArn:           "arn:aws:iam::222222222222:role/deploy",
				ExternalId:        "expected-external-id",
				Duration:          "1h",
				Tags:              map[string]string{"Team": "platform"},
				TransitiveTagKeys: []string{"Team"},
				SourceIdentity:    "platform-engineer",
			},
		},
		DefaultTags: &awsprovider.AwsDefaultTags{
			Tags: map[string]string{"CostCenter": "eng", "ManagedBy": "planton"},
		},
		Endpoints:  map[string]string{"sts": "https://sts.internal.example.com", "s3": "https://s3.internal.example.com"},
		MaxRetries: proto.Int32(10),
		RetryMode:  "adaptive",
	}
}

func TestWriteProviderOverrideFile_FullSurface(t *testing.T) {
	moduleDir := t.TempDir()

	wrote, err := WriteProviderOverrideFile(moduleDir, awsIacInputYaml(t, fullProviderBlockConfig()))
	require.NoError(t, err)
	assert.True(t, wrote)

	content, err := os.ReadFile(filepath.Join(moduleDir, OverrideFileName))
	require.NoError(t, err)
	got := string(content)

	// One provider block carrying the whole surface, hops in declaration order.
	assert.Contains(t, got, `provider "aws" {`)
	first := `role_arn     = "arn:aws:iam::111111111111:role/intermediate"`
	second := `role_arn    = "arn:aws:iam::222222222222:role/deploy"`
	assert.Contains(t, got, first)
	assert.Contains(t, got, second)
	assert.Less(t, indexOf(t, got, first), indexOf(t, got, second),
		"assume_role hops must be emitted in chain order")
	assert.Contains(t, got, `external_id`)
	assert.Contains(t, got, `transitive_tag_keys = ["Team"]`)
	assert.Contains(t, got, `source_identity`)
	assert.Contains(t, got, "default_tags {")
	assert.Contains(t, got, `CostCenter = "eng"`)
	assert.Contains(t, got, "endpoints {")
	assert.Contains(t, got, `sts = "https://sts.internal.example.com"`)
	assert.Contains(t, got, `max_retries = 10`)
	assert.Contains(t, got, `retry_mode  = "adaptive"`)

	// The credential law and the region rule: never on disk, never in the block.
	assert.NotContains(t, got, "access_key =")
	assert.NotContains(t, got, "secret_key")
	assert.NotContains(t, got, "AKIAIOSFODNN7EXAMPLE")
	assert.NotContains(t, got, "region =")
	assert.NotContains(t, got, "us-east-1")
}

func TestWriteProviderOverrideFile_Deterministic(t *testing.T) {
	moduleDir := t.TempDir()
	yaml := awsIacInputYaml(t, fullProviderBlockConfig())

	_, err := WriteProviderOverrideFile(moduleDir, yaml)
	require.NoError(t, err)
	first, err := os.ReadFile(filepath.Join(moduleDir, OverrideFileName))
	require.NoError(t, err)

	_, err = WriteProviderOverrideFile(moduleDir, yaml)
	require.NoError(t, err)
	second, err := os.ReadFile(filepath.Join(moduleDir, OverrideFileName))
	require.NoError(t, err)

	assert.Equal(t, string(first), string(second), "map iteration must not leak into output")
}

func TestWriteProviderOverrideFile_NoArgs_RemovesStaleFile(t *testing.T) {
	moduleDir := t.TempDir()
	stalePath := filepath.Join(moduleDir, OverrideFileName)
	require.NoError(t, os.WriteFile(stalePath, []byte("stale from a previous run"), 0644))

	// A config with credentials but no provider-block args: the file must go away
	// (the CLI's zip cache and user checkouts outlive runs).
	cfg := &awsprovider.AwsProviderConfig{
		AccountId:       "123456789012",
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY00",
		Region:          "us-east-1",
	}
	wrote, err := WriteProviderOverrideFile(moduleDir, awsIacInputYaml(t, cfg))
	require.NoError(t, err)
	assert.False(t, wrote)
	_, statErr := os.Stat(stalePath)
	assert.True(t, os.IsNotExist(statErr), "stale override file must be removed")
}

func TestWriteProviderOverrideFile_NoProviderConfig_NoWrite(t *testing.T) {
	moduleDir := t.TempDir()
	manifest := &awsvpcv1.AwsVpc{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsVpc",
		Metadata:   &shared.CatalogObjectMetadata{Name: "ambient-vpc"},
	}
	yaml, err := iacinput.BuildIacInputYaml(manifest, nil)
	require.NoError(t, err)

	wrote, err := WriteProviderOverrideFile(moduleDir, yaml)
	require.NoError(t, err)
	assert.False(t, wrote)
	_, statErr := os.Stat(filepath.Join(moduleDir, OverrideFileName))
	assert.True(t, os.IsNotExist(statErr))
}

func TestWriteProviderOverrideFile_NonAwsProvider_NoOp(t *testing.T) {
	moduleDir := t.TempDir()
	manifest := &dovpcv1.DigitalOceanVpc{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanVpc",
		Metadata:   &shared.CatalogObjectMetadata{Name: "do-vpc"},
	}
	providerConfig, cleanup, err := iacinputproviderconfig.BuildFromProto(
		&digitaloceanprovider.DigitalOceanProviderConfig{ApiToken: "dop_v1_token"},
		catalogkind.CatalogProvider_digital_ocean)
	require.NoError(t, err)
	t.Cleanup(cleanup)
	yaml, err := iacinput.BuildIacInputYaml(manifest, providerConfig)
	require.NoError(t, err)

	wrote, err := WriteProviderOverrideFile(moduleDir, yaml)
	require.NoError(t, err)
	assert.False(t, wrote)
}

func indexOf(t *testing.T, haystack, needle string) int {
	t.Helper()
	idx := len([]byte(haystack))
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return idx
}
