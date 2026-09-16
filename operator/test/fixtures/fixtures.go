// Package fixtures carries the in-cluster lab infrastructure the Kind suites
// under test/ stand a platform beside: an S3-compatible store to archive to
// and a key service to seal against. Each fixture is a plain manifest,
// pinned and embedded the way the operator embeds the charts it renders, so
// a suite applies it with kubectl and reads its names from here rather than
// re-typing them; the package's own test decodes every manifest and holds
// the image pins, which is how a stale pin or a typo is caught by `make
// test` and never by a two-hour lane. Nothing here is a product component
// and nothing here is a real secret.
package fixtures

import (
	_ "embed"
)

// The S3-compatible store (MinIO), one replica on an emptyDir, and the
// bucket a lane archives to. Any suite that needs an object store on Kind
// applies this and reads the names below.
const (
	MinIONamespace = "planton-lane-minio"
	// MinIOEndpoint is the store as pods reach it; a platform's backup
	// declaration points its s3 arm here.
	MinIOEndpoint  = "http://minio.planton-lane-minio.svc:9000"
	MinIOBucket    = "planton-backups"
	MinIOAccessKey = "laneadmin"
	MinIOSecretKey = "laneadmin-secret-not-a-real-secret"
	// MinIOBucketJob is the one-shot Job that creates the bucket; a lane
	// waits for it to complete before declaring anything against the store.
	MinIOBucketJob = "make-bucket"
)

//go:embed minio.yaml
var minio []byte

// MinIO returns the store's manifests (Namespace, Deployment, Service, and
// the bucket Job), ready for `kubectl apply -f -`.
func MinIO() []byte { return minio }

// The key service: a dev-mode OpenBao with a fixed root token, in memory. A
// lane mounts its transit engine and creates KeyHolderTransitKey on it
// before declaring a platform whose vault seals against that key.
const (
	KeyHolderNamespace = "planton-lane-key-holder"
	// KeyHolderAddress is the key service as the vault reaches it -- the
	// transit seal's `address`.
	KeyHolderAddress   = "http://key-holder.planton-lane-key-holder.svc:8200"
	KeyHolderRootToken = "lane-holder-root"
	// KeyHolderPodSelector selects the key holder's pod for kubectl exec.
	KeyHolderPodSelector = "app=key-holder"
	// KeyHolderTransitMount and KeyHolderTransitKey are the engine and the
	// key a lane creates on the holder; the seal declaration names both.
	KeyHolderTransitMount = "transit"
	KeyHolderTransitKey   = "planton-unseal"
)

//go:embed key-holder.yaml
var keyHolder []byte

// KeyHolder returns the key service's manifests (Namespace, Deployment,
// Service), ready for `kubectl apply -f -`.
func KeyHolder() []byte { return keyHolder }
