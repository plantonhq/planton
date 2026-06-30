package module

import (
	"strconv"
	"strings"

	gcpcloudcdnv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudcdn/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
)

type Locals struct {
	// GcpLabels contains all the labels to be applied to GCP resources
	GcpLabels map[string]string
	// CDN resource naming
	BackendName        string
	BackendBucketName  string
	BackendServiceName string
	HealthCheckName    string
	UrlMapName         string
	HttpsProxyName     string
	HttpProxyName      string
	GlobalAddressName  string
	SslCertName        string
	// Backend configuration
	IsGcsBucket      bool
	IsComputeService bool
	IsCloudRun       bool
	IsExternalOrigin bool
	// CDN configuration
	CacheMode              string
	DefaultTtl             int
	MaxTtl                 int
	ClientTtl              int
	NegativeCachingEnabled bool
}

func initializeLocals(stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput) *Locals {
	locals := &Locals{}

	target := stackInput.Target

	// Create GCP labels from metadata
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpCloudCdn.String()),
	}

	if target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}

	// Generate resource names based on metadata
	cdnName := stackInput.Target.Metadata.Name
	locals.BackendName = cdnName
	locals.BackendBucketName = cdnName + "-backend-bucket"
	locals.BackendServiceName = cdnName + "-backend-service"
	locals.HealthCheckName = cdnName + "-health-check"
	locals.UrlMapName = cdnName + "-url-map"
	locals.HttpsProxyName = cdnName + "-https-proxy"
	locals.HttpProxyName = cdnName + "-http-proxy"
	locals.GlobalAddressName = cdnName + "-global-ip"
	locals.SslCertName = cdnName + "-ssl-cert"

	// Determine backend type
	backend := stackInput.Target.Spec.Backend
	if backend != nil {
		switch backend.BackendType.(type) {
		case *gcpcloudcdnv1.GcpCloudCdnBackend_GcsBucket:
			locals.IsGcsBucket = true
		case *gcpcloudcdnv1.GcpCloudCdnBackend_ComputeService:
			locals.IsComputeService = true
		case *gcpcloudcdnv1.GcpCloudCdnBackend_CloudRunService:
			locals.IsCloudRun = true
		case *gcpcloudcdnv1.GcpCloudCdnBackend_ExternalOrigin:
			locals.IsExternalOrigin = true
		}
	}

	// Set cache configuration with defaults
	spec := stackInput.Target.Spec

	// Cache mode (default: CACHE_ALL_STATIC)
	if spec.CacheMode != nil {
		locals.CacheMode = spec.CacheMode.String()
	} else {
		locals.CacheMode = gcpcloudcdnv1.CacheMode_CACHE_ALL_STATIC.String()
	}

	// Default TTL (default: 3600 seconds = 1 hour)
	if spec.DefaultTtlSeconds != nil {
		locals.DefaultTtl = int(*spec.DefaultTtlSeconds)
	} else {
		locals.DefaultTtl = 3600
	}

	// Max TTL (default: 86400 seconds = 1 day)
	if spec.MaxTtlSeconds != nil {
		locals.MaxTtl = int(*spec.MaxTtlSeconds)
	} else {
		locals.MaxTtl = 86400
	}

	// Client TTL (default: same as max TTL)
	if spec.ClientTtlSeconds != nil {
		locals.ClientTtl = int(*spec.ClientTtlSeconds)
	} else {
		locals.ClientTtl = locals.MaxTtl
	}

	// Negative caching (default: false)
	if spec.EnableNegativeCaching != nil {
		locals.NegativeCachingEnabled = *spec.EnableNegativeCaching
	} else {
		locals.NegativeCachingEnabled = false
	}

	return locals
}
