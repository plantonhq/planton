package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	kuberneteslokiv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesloki/v1alpha1"
	"github.com/plantonhq/planton/pkg/cloudflare/r2"
)

// r2Storage is the rendering of the r2 storage arm: the chart's S3 values
// (R2 speaks S3; loki.storage.type stays "s3"), the bucket names, and the
// data of the module-owned `<name>-r2-credentials` Secret the key pair
// reaches Loki from.
//
// Everything S3-shaped is composed here from R2's own vocabulary: the
// endpoint from the account and the bucket's jurisdiction (pkg/cloudflare/r2
// is the one home of that host table), region "auto", path-style
// addressing. The key pair never enters the values: the config carries
// ${VAR} placeholders that -config.expand-env expands from environment
// variables sourced from the Secret.
//
// PARITY: the Terraform module renders the same values (locals.tf, the r2
// block, whose host table names pkg/cloudflare/r2 as its source of truth).
type r2Storage struct {
	S3Values    map[string]interface{}
	BucketNames map[string]interface{}
	SecretData  map[string]string
}

// buildR2Storage returns nil unless the r2 arm is declared.
func buildR2Storage(storage *kuberneteslokiv1alpha1.KubernetesLokiStorage) *r2Storage {
	arm := storage.GetR2()
	if arm == nil {
		return nil
	}
	bucket := arm.GetBucket().GetValue()
	return &r2Storage{
		S3Values: map[string]interface{}{
			"endpoint":         r2.S3Endpoint(arm.GetAccountId().GetValue(), arm.GetJurisdiction().GetValue()),
			"region":           r2.Region,
			"s3ForcePathStyle": true,
			"accessKeyId":      fmt.Sprintf("${%s}", vars.EnvS3AccessKeyId),
			"secretAccessKey":  fmt.Sprintf("${%s}", vars.EnvS3SecretAccessKey),
		},
		BucketNames: bucketNames(bucket, ""),
		SecretData: map[string]string{
			vars.R2AccessKeyIdKey:     arm.GetCredentials().GetAccessKeyId().GetValue(),
			vars.R2SecretAccessKeyKey: arm.GetCredentials().GetSecretAccessKey().GetValue(),
		},
	}
}

// r2CredentialEnv renders the two variables Loki's config expands, read
// from the module-owned credentials Secret.
func r2CredentialEnv(secretName string) []interface{} {
	return []interface{}{
		moduleSecretEnvVar(vars.EnvS3AccessKeyId, secretName, vars.R2AccessKeyIdKey),
		moduleSecretEnvVar(vars.EnvS3SecretAccessKey, secretName, vars.R2SecretAccessKeyKey),
	}
}

func moduleSecretEnvVar(name, secretName, key string) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
		"valueFrom": map[string]interface{}{
			"secretKeyRef": map[string]interface{}{
				"name": secretName,
				"key":  key,
			},
		},
	}
}

// credentialsChecksum fingerprints a module-owned Secret's data for a pod
// annotation. Loki reads its credentials only at start, so a changed value
// must change the pod template for the next apply to roll it. The
// fingerprint is the SHA-256 of the data's JSON encoding (keys sorted),
// exactly what OpenTofu's sha256(jsonencode(map)) computes.
func credentialsChecksum(data map[string]string) string {
	encoded, _ := json.Marshal(data)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
