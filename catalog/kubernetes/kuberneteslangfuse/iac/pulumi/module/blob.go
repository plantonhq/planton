package module

import (
	"fmt"

	kuberneteslangfusev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslangfuse/v1alpha1"
)

// blobTarget is one upload type's resolved object storage target, in the
// shape the chart's s3.<uploadType> values take.
type blobTarget struct {
	// Family is the chart's storageProvider: "s3" (S3, R2 and every
	// S3-compatible store), "gcs" or "azure".
	Family string

	Bucket         string
	Region         string
	Endpoint       string
	ForcePathStyle bool

	// AccessKeyId / SecretAccessKey are the chart's {value} or
	// {secretKeyRef} blocks; nil = no static credentials (workload
	// identity or application default credentials).
	AccessKeyId     map[string]interface{}
	SecretAccessKey map[string]interface{}
}

// resolveBlobTargets resolves the event target, then the media and batch
// export targets - each the event target unless it declares its own.
// Upload types are named event, media and batch in the instance Secret's
// keys.
func (l *Locals) resolveBlobTargets() {
	storage := l.Spec.GetBlobStorage()
	l.EventTarget = l.resolveBlobTarget("event", storage.GetTarget())
	l.MediaTarget = l.EventTarget
	if t := storage.GetMediaUpload().GetTarget(); t != nil {
		l.MediaTarget = l.resolveBlobTarget("media", t)
	}
	l.BatchTarget = l.EventTarget
	if t := storage.GetBatchExport().GetTarget(); t != nil {
		l.BatchTarget = l.resolveBlobTarget("batch", t)
	}
	if json := l.gcsCredentials(); json != "" {
		l.declare(keyGcsCredentials, json)
	}
}

// gcsCredentials returns the first declared GCS service account key: every
// GCS target of one Langfuse shares one credential
// (LANGFUSE_GOOGLE_CLOUD_STORAGE_CREDENTIALS).
func (l *Locals) gcsCredentials() string {
	storage := l.Spec.GetBlobStorage()
	for _, t := range []*kuberneteslangfusev1alpha1.KubernetesLangfuseBlobTarget{
		storage.GetTarget(), storage.GetMediaUpload().GetTarget(), storage.GetBatchExport().GetTarget(),
	} {
		if json := t.GetGcs().GetCredentialsJson(); json != "" {
			return json
		}
	}
	return ""
}

func (l *Locals) resolveBlobTarget(uploadType string, t *kuberneteslangfusev1alpha1.KubernetesLangfuseBlobTarget) *blobTarget {
	secretKeyRef := func(key string) map[string]interface{} {
		return map[string]interface{}{"secretKeyRef": map[string]interface{}{"name": l.InstanceSecretName, "key": key}}
	}
	switch {
	case t.GetS3Compatible() != nil:
		c := t.GetS3Compatible()
		out := &blobTarget{
			Family:         "s3",
			Bucket:         c.GetBucket(),
			Region:         stringOr(c.Region, "auto"),
			Endpoint:       c.GetEndpoint().GetValue(),
			ForcePathStyle: c.ForcePathStyle == nil || c.GetForcePathStyle(),
		}
		if cs := c.GetCredentialsSecret(); cs != nil {
			out.AccessKeyId = map[string]interface{}{"secretKeyRef": map[string]interface{}{
				"name": cs.GetSecretName().GetValue(),
				"key":  stringOr(cs.AccessKeyIdKey, "admin_access_key_id"),
			}}
			out.SecretAccessKey = map[string]interface{}{"secretKeyRef": map[string]interface{}{
				"name": cs.GetSecretName().GetValue(),
				"key":  stringOr(cs.SecretAccessKeyKey, "admin_secret_access_key"),
			}}
		} else if ak := c.GetAccessKey(); ak != nil {
			key := fmt.Sprintf(keyBlobSecretAccessKeyTmpl, uploadType)
			l.declare(key, ak.GetSecretAccessKey())
			out.AccessKeyId = map[string]interface{}{"value": ak.GetAccessKeyId()}
			out.SecretAccessKey = secretKeyRef(key)
		}
		return out
	case t.GetS3() != nil:
		c := t.GetS3()
		out := &blobTarget{
			Family:   "s3",
			Bucket:   c.GetBucket().GetValue(),
			Region:   c.GetRegion().GetValue(),
			Endpoint: c.GetEndpoint(),
		}
		if ak := c.GetAccessKey(); ak != nil {
			key := fmt.Sprintf(keyBlobSecretAccessKeyTmpl, uploadType)
			l.declare(key, ak.GetSecretAccessKey())
			out.AccessKeyId = map[string]interface{}{"value": ak.GetAccessKeyId()}
			out.SecretAccessKey = secretKeyRef(key)
		}
		return out
	case t.GetR2() != nil:
		c := t.GetR2()
		key := fmt.Sprintf(keyBlobSecretAccessKeyTmpl, uploadType)
		l.declare(key, c.GetCredentials().GetSecretAccessKey().GetValue())
		// R2 serves path-style and virtual-hosted requests alike; path
		// style needs no per-bucket DNS name.
		return &blobTarget{
			Family:          "s3",
			Bucket:          c.GetBucket().GetValue(),
			Region:          "auto",
			Endpoint:        c.GetEndpoint().GetValue(),
			ForcePathStyle:  true,
			AccessKeyId:     map[string]interface{}{"value": c.GetCredentials().GetAccessKeyId().GetValue()},
			SecretAccessKey: secretKeyRef(key),
		}
	case t.GetGcs() != nil:
		return &blobTarget{Family: "gcs", Bucket: t.GetGcs().GetBucket().GetValue()}
	case t.GetAzureBlob() != nil:
		c := t.GetAzureBlob()
		key := fmt.Sprintf(keyBlobAzureAccountKeyTmpl, uploadType)
		l.declare(key, c.GetAccountKey())
		endpoint := c.GetEndpoint().GetValue()
		if endpoint == "" {
			endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", c.GetStorageAccount().GetValue())
		}
		// Langfuse's Azure driver reads the account name from the access key
		// id and the account key from the secret access key.
		return &blobTarget{
			Family:          "azure",
			Bucket:          c.GetContainer().GetValue(),
			Endpoint:        endpoint,
			AccessKeyId:     map[string]interface{}{"value": c.GetStorageAccount().GetValue()},
			SecretAccessKey: secretKeyRef(key),
		}
	}
	return &blobTarget{Family: "s3"}
}

// values renders the target into one s3.<uploadType> values block. The
// chart falls back from each upload type's keys to the top-level s3.* keys
// with Helm's `default`, which treats false as unset - so forcePathStyle is
// always rendered explicitly, and the top-level fallbacks are pinned empty
// (false for forcePathStyle) in buildHelmValues.
func (t *blobTarget) values(prefix string) map[string]interface{} {
	out := map[string]interface{}{
		"bucket":         t.Bucket,
		"prefix":         prefix,
		"forcePathStyle": t.ForcePathStyle,
	}
	setString(out, "region", t.Region)
	setString(out, "endpoint", t.Endpoint)
	if t.AccessKeyId != nil {
		out["accessKeyId"] = t.AccessKeyId
	}
	if t.SecretAccessKey != nil {
		out["secretAccessKey"] = t.SecretAccessKey
	}
	return out
}

func stringOr(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}
