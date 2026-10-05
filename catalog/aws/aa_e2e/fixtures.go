package aa_e2e

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	s3vectorstypes "github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	"github.com/pkg/errors"
)

// Standing S3 Vectors fixture for knowledge-base lanes. A Bedrock
// knowledge base of the VECTOR type needs a vector store, and S3 Vectors
// is the only self-contained one at the pin: the OpenSearch Serverless
// path needs a vector index no AWS-provider resource can create (upstream
// tests that path with the third-party opensearch provider), and no
// catalog kind wraps aws_s3vectors_* yet (a planned AwsS3VectorBucket
// kind carries the ledger judgment). Until that kind exists, the harness
// ensures the fixture through the SDK -- exactly how the upstream
// provider's own acceptance tests provision it -- and exports its index
// ARN as a PLANTON_E2E_ environment token for scenario manifests.
//
// The fixture is a STANDING account fixture (the seeded-S3-bucket class):
// S3 Vectors bills per use, so an empty bucket+index costs nothing at
// rest, and reusing one across runs avoids create/delete churn. It is
// intentionally NOT torn down per run; the account janitor policy governs
// it like the other seeded fixtures.
const (
	// S3VectorsIndexArnEnvVar is the ${E2E_ENV:...} token variable the
	// knowledge-base scenarios reference for the fixture index's ARN.
	S3VectorsIndexArnEnvVar = "PLANTON_E2E_S3VECTORS_INDEX_ARN"

	s3VectorsFixtureBucket = "planton-e2e-s3vectors"
	s3VectorsFixtureIndex  = "planton-e2e-kb-index"

	// The index shape pairs with Titan Text Embeddings V2 at 256
	// dimensions (the scenario pins dimensions: 256) -- the same shape the
	// upstream provider's own S3 Vectors acceptance fixture uses.
	s3VectorsFixtureDimension = 256
)

// EnsureS3VectorsKnowledgeBaseFixture idempotently creates the standing
// S3 Vectors bucket+index and exports S3VectorsIndexArnEnvVar for
// scenario token expansion. Call it from a kind's test entrypoint
// BEFORE running scenarios that reference the token.
func EnsureS3VectorsKnowledgeBaseFixture(ctx context.Context) error {
	if os.Getenv(S3VectorsIndexArnEnvVar) != "" {
		return nil
	}

	region := firstNonEmpty(os.Getenv("E2E_AWS_REGION"), os.Getenv("AWS_REGION"), defaultRegion)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return errors.Wrap(err, "failed to load AWS config for the S3 Vectors fixture")
	}
	client := s3vectors.NewFromConfig(cfg)

	if _, err := client.CreateVectorBucket(ctx, &s3vectors.CreateVectorBucketInput{
		VectorBucketName: aws.String(s3VectorsFixtureBucket),
	}); err != nil && !isS3VectorsConflict(err) {
		return errors.Wrapf(err, "create vector bucket %s", s3VectorsFixtureBucket)
	}

	if _, err := client.CreateIndex(ctx, &s3vectors.CreateIndexInput{
		VectorBucketName: aws.String(s3VectorsFixtureBucket),
		IndexName:        aws.String(s3VectorsFixtureIndex),
		DataType:         s3vectorstypes.DataTypeFloat32,
		Dimension:        aws.Int32(s3VectorsFixtureDimension),
		DistanceMetric:   s3vectorstypes.DistanceMetricEuclidean,
	}); err != nil && !isS3VectorsConflict(err) {
		return errors.Wrapf(err, "create vector index %s", s3VectorsFixtureIndex)
	}

	index, err := client.GetIndex(ctx, &s3vectors.GetIndexInput{
		VectorBucketName: aws.String(s3VectorsFixtureBucket),
		IndexName:        aws.String(s3VectorsFixtureIndex),
	})
	if err != nil {
		return errors.Wrapf(err, "get vector index %s", s3VectorsFixtureIndex)
	}

	arn := aws.ToString(index.Index.IndexArn)
	if err := os.Setenv(S3VectorsIndexArnEnvVar, arn); err != nil {
		return errors.Wrap(err, "export the fixture index ARN")
	}
	fmt.Printf("  [aws] S3 Vectors knowledge-base fixture ready: %s\n", arn)
	return nil
}

// isS3VectorsConflict matches the already-exists class of both fixture
// creates (idempotent ensure).
func isS3VectorsConflict(err error) bool {
	var conflict *s3vectorstypes.ConflictException
	return errors.As(err, &conflict) || strings.Contains(err.Error(), "ConflictException")
}

// Standing AgentCore code-bundle fixture for agent-runtime lanes. An
// AgentCore runtime's code artifact is a zip in S3 the service reads at
// create; no upstream provider resource seeds one (the provider's own
// acceptance tests take the bundle location from environment variables).
// The harness ensures a tiny valid bundle -- one main.py that satisfies
// the managed-runtime entrypoint contract -- and exports the bucket name
// as a PLANTON_E2E_ environment token for scenario manifests (the key is
// the constant the scenarios spell literally).
//
// The fixture is a STANDING account fixture (the seeded-S3-bucket class):
// one ~300-byte object costs nothing at rest, and reusing it across runs
// avoids create/upload churn. It is intentionally NOT torn down per run;
// the account janitor policy governs it like the other seeded fixtures.
const (
	// AgentCoreCodeBucketEnvVar is the ${E2E_ENV:...} token variable the
	// agent-runtime scenarios reference for the fixture bucket's name.
	AgentCoreCodeBucketEnvVar = "PLANTON_E2E_AGENTCORE_CODE_BUCKET"

	agentCoreCodeFixtureBucket = "planton-e2e-agentcore-code"

	// AgentCoreCodeFixtureKey is the bundle's object key -- scenarios
	// spell it literally in spec.artifact.code.s3.prefix.
	AgentCoreCodeFixtureKey = "bundles/e2e-agent.zip"
)

// EnsureAgentCoreCodeBundleFixture idempotently creates the standing
// code-bundle bucket and object and exports AgentCoreCodeBucketEnvVar for
// scenario token expansion. Call it from a kind's test entrypoint
// BEFORE running scenarios that reference the token.
func EnsureAgentCoreCodeBundleFixture(ctx context.Context) error {
	if os.Getenv(AgentCoreCodeBucketEnvVar) != "" {
		return nil
	}

	region := firstNonEmpty(os.Getenv("E2E_AWS_REGION"), os.Getenv("AWS_REGION"), defaultRegion)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return errors.Wrap(err, "failed to load AWS config for the AgentCore code fixture")
	}
	client := s3.NewFromConfig(cfg)

	createInput := &s3.CreateBucketInput{Bucket: aws.String(agentCoreCodeFixtureBucket)}
	// us-east-1 is the only region CreateBucket forbids a location
	// constraint for.
	if region != "us-east-1" {
		createInput.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}
	if _, err := client.CreateBucket(ctx, createInput); err != nil && !isS3BucketExists(err) {
		return errors.Wrapf(err, "create code-bundle bucket %s", agentCoreCodeFixtureBucket)
	}

	bundle, err := agentCoreCodeBundle()
	if err != nil {
		return errors.Wrap(err, "assemble the code bundle zip")
	}
	if _, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(agentCoreCodeFixtureBucket),
		Key:    aws.String(AgentCoreCodeFixtureKey),
		Body:   bytes.NewReader(bundle),
	}); err != nil {
		return errors.Wrapf(err, "put code bundle %s", AgentCoreCodeFixtureKey)
	}

	if err := os.Setenv(AgentCoreCodeBucketEnvVar, agentCoreCodeFixtureBucket); err != nil {
		return errors.Wrap(err, "export the fixture bucket name")
	}
	fmt.Printf("  [aws] AgentCore code-bundle fixture ready: s3://%s/%s\n", agentCoreCodeFixtureBucket, AgentCoreCodeFixtureKey)
	return nil
}

// agentCoreCodeBundle assembles the minimal managed-runtime bundle: one
// main.py exposing the HTTP contract the PYTHON_3_13 managed runtime
// expects. The lanes never invoke the runtime, so the handler body only
// needs to be syntactically plausible.
func agentCoreCodeBundle() ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	entry, err := writer.Create("main.py")
	if err != nil {
		return nil, err
	}
	if _, err := entry.Write([]byte(
		"def handler(event, context):\n" +
			"    return {\"status\": \"ok\", \"echo\": event}\n",
	)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// isS3BucketExists matches the already-exists class of the bucket create
// (idempotent ensure; BucketAlreadyOwnedByYou is the same-account form).
func isS3BucketExists(err error) bool {
	var owned *s3types.BucketAlreadyOwnedByYou
	var exists *s3types.BucketAlreadyExists
	return errors.As(err, &owned) || errors.As(err, &exists)
}

// Standing KMS key fixture for the AgentCore token-vault lanes. The
// vault's revert path carries a server contract no per-lane fixture can
// satisfy: SetTokenVaultCMK validates the PREVIOUS key's state on every
// write ("Old KMS Key validation failed ... expected KeyState:ENABLED"),
// so reverting a CustomerManagedKey vault to ServiceManagedKey requires
// the OLD customer key to still be ENABLED. A chain-deployed key dies
// into its deletion window at the CMK lane's fixture teardown, wedging
// the vault before the revert lane runs (live-caught). The standing key
// stays ENABLED across lanes and runs, so the CMK lane binds it and the
// service-managed lane's revert always finds it alive.
//
// The fixture is a STANDING account fixture (the S3 Vectors class): one
// symmetric key costs USD 1/month at rest, and the alias makes the
// ensure idempotent (KMS keys are unnamed -- the alias is the findable
// handle). It is intentionally NOT torn down per run; the account
// janitor policy governs it like the other seeded fixtures. The ensure
// self-heals a key found disabled or inside its deletion window
// (cancel + re-enable) -- the same rescue recipe the kind's GUIDE
// teaches for a wedged vault.
const (
	// TokenVaultKMSKeyArnEnvVar is the ${E2E_ENV:...} token variable the
	// token-vault CMK scenario references for the standing key's ARN.
	TokenVaultKMSKeyArnEnvVar = "PLANTON_E2E_ACETV_KMS_KEY_ARN"

	tokenVaultKMSKeyFixtureAlias = "alias/planton-e2e-acetv"
)

// EnsureTokenVaultKMSKeyFixture idempotently creates (or revives) the
// standing token-vault KMS key and exports TokenVaultKMSKeyArnEnvVar for
// scenario token expansion. Call it from the token-vault test entrypoint
// BEFORE running scenarios.
func EnsureTokenVaultKMSKeyFixture(ctx context.Context) error {
	if os.Getenv(TokenVaultKMSKeyArnEnvVar) != "" {
		return nil
	}

	region := firstNonEmpty(os.Getenv("E2E_AWS_REGION"), os.Getenv("AWS_REGION"), defaultRegion)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return errors.Wrap(err, "failed to load AWS config for the token-vault KMS fixture")
	}
	client := kms.NewFromConfig(cfg)

	var keyArn string
	describe, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: aws.String(tokenVaultKMSKeyFixtureAlias),
	})
	switch {
	case err == nil:
		meta := describe.KeyMetadata
		keyArn = aws.ToString(meta.Arn)
		// Self-heal: a prior teardown may have scheduled deletion or
		// disabled the key -- the vault contract needs ENABLED.
		if meta.KeyState == kmstypes.KeyStatePendingDeletion {
			if _, err := client.CancelKeyDeletion(ctx, &kms.CancelKeyDeletionInput{KeyId: meta.KeyId}); err != nil {
				return errors.Wrap(err, "cancel the standing key's scheduled deletion")
			}
		}
		if meta.KeyState != kmstypes.KeyStateEnabled {
			if _, err := client.EnableKey(ctx, &kms.EnableKeyInput{KeyId: meta.KeyId}); err != nil {
				return errors.Wrap(err, "re-enable the standing key")
			}
		}
	case isKMSNotFound(err):
		created, err := client.CreateKey(ctx, &kms.CreateKeyInput{
			Description: aws.String("Planton E2E standing key for AgentCore token-vault lanes"),
			Tags: []kmstypes.Tag{
				{TagKey: aws.String("managed-by"), TagValue: aws.String("planton-e2e")},
			},
		})
		if err != nil {
			return errors.Wrap(err, "create the standing token-vault key")
		}
		keyArn = aws.ToString(created.KeyMetadata.Arn)
		if _, err := client.CreateAlias(ctx, &kms.CreateAliasInput{
			AliasName:   aws.String(tokenVaultKMSKeyFixtureAlias),
			TargetKeyId: created.KeyMetadata.KeyId,
		}); err != nil {
			return errors.Wrapf(err, "alias the standing key as %s", tokenVaultKMSKeyFixtureAlias)
		}
	default:
		return errors.Wrapf(err, "describe %s", tokenVaultKMSKeyFixtureAlias)
	}

	if err := os.Setenv(TokenVaultKMSKeyArnEnvVar, keyArn); err != nil {
		return errors.Wrap(err, "export the standing key ARN")
	}
	fmt.Printf("  [aws] token-vault KMS key fixture ready: %s\n", keyArn)
	return nil
}

// isKMSNotFound matches the alias-not-found class of the describe
// (idempotent ensure; a missing alias means create key + alias).
func isKMSNotFound(err error) bool {
	var notFound *kmstypes.NotFoundException
	return errors.As(err, &notFound)
}
