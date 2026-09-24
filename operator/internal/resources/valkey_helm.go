package resources

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// The platform's cache-server role is named "redis" (the wire protocol every
// consumer speaks), but the engine that serves it is Valkey -- the
// BSD-3-Clause, Linux-Foundation fork that ships a redis-compatible server.
// Valkey is chosen over Redis deliberately: Redis 8+ is tri-licensed
// (RSALv2/SSPLv1/AGPLv3), a family self-hosted customers' legal teams often
// cannot accept, while Valkey stays permissive. This mirrors the desktop
// instance, which supervises a valkey-server behind REDIS_* configuration.
// The role keeps the "redis" name across the CRD, status, Secret, Service,
// and connection surfaces so the engine choice is invisible to consumers.
const (
	ValkeyHelmChartVersion = "3.0.31"
	RedisPort              = 6379

	// Valkey image coordinates. Chart v3.0.31 (appVersion 8.1.3) defaults to
	// bitnami/valkey, which was removed from Docker Hub when Bitnami
	// deprecated the free registry -- so the image is pinned to the frozen
	// bitnamilegacy mirror at the matching app version.
	valkeyImageRegistry   = "docker.io"
	valkeyImageRepository = "bitnamilegacy/valkey"
	valkeyImageTag        = "8.1.3-debian-12-r3"

	// RedisSecretKey is the data key inside the operator-generated credential
	// Secret that holds the cache-server password. The Valkey chart is
	// configured to read this key via auth.existingSecret +
	// auth.existingSecretPasswordKey.
	RedisSecretKey = "redis-password"

	// The store's sizing, chosen here rather than left to the chart. The
	// chart's own default is its "nano" preset (a 192Mi limit its header
	// calls "for basic testing"), no maxmemory, and noeviction -- so a
	// dataset that outgrew the preset was OOM-killed and then reloaded more
	// persisted data than it may hold on every restart, a crash loop behind a
	// front door still answering 200 (met live 2026-09-18, 37 hours after a
	// fresh install). The numbers are the hosted product's for the same store
	// role: a ceiling inside the limit with eviction, so what is persisted
	// always reloads. No CPU limit (requests-only, the house pattern).
	ValkeyDefaultMaxMemory       = "768mb"
	ValkeyDefaultMaxMemoryPolicy = "allkeys-lru"
	valkeyDefaultCPURequest      = "100m"
	valkeyDefaultMemoryRequest   = "256Mi"
	valkeyDefaultMemoryLimit     = "1Gi"

	// The chart's resourcesPreset silently wins over nothing; naming "none"
	// beside explicit resources makes the override unambiguous.
	valkeyResourcesPresetNone = "none"

	// valkey.conf spells booleans as yes/no.
	valkeyConfYes = "yes"
	valkeyConfNo  = "no"
)

// ValkeyHelmOptions is everything the store's render needs, resolved by the
// component (spec field, then platform-wide setting, then the default above)
// before it gets here.
type ValkeyHelmOptions struct {
	CRName string

	// Persistence keeps the dataset on a volume and replays it (append-only
	// file) after a restart. Off renders an emptyDir and no claim template.
	Persistence bool
	// StorageSize and StorageClass shape the volume; read only with
	// Persistence. An empty StorageClass is OMITTED so the cluster default
	// provisions (an explicit "" would disable dynamic provisioning in the
	// Bitnami convention).
	StorageSize  string
	StorageClass string

	// MaxMemory and MaxMemoryPolicy are the dataset ceiling and what happens
	// at it, in Valkey's own words.
	MaxMemory       string
	MaxMemoryPolicy string

	// Resources is the container's sizing.
	Resources corev1.ResourceRequirements
}

// ValkeyDefaultResources is the container sizing every install gets unless
// spec.database.redis.resources says otherwise.
func ValkeyDefaultResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(valkeyDefaultCPURequest),
			corev1.ResourceMemory: resource.MustParse(valkeyDefaultMemoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(valkeyDefaultMemoryLimit),
		},
	}
}

// ValkeyHelmValues builds the Helm values map for rendering the Bitnami Valkey
// chart in standalone architecture with password authentication via an
// operator-managed Secret:
//   - fullnameOverride keeps the role-named resources ("{crName}-redis-*")
//   - standalone architecture (no sentinel/replica overhead)
//   - Password from existingSecret (operator-generated, never in values)
//   - the container's resources set explicitly, the chart's preset named off
//   - the server configuration (valkey.conf) carrying the memory ceiling, the
//     eviction policy, and the persistence posture
//   - a volume only when persistence is on
//
// The Valkey chart names the data-serving workload "primary" (StatefulSet and
// Service "{fullname}-primary"), where the Redis chart said "master" -- the
// readiness check and RedisServiceHost follow that naming.
func ValkeyHelmValues(opts ValkeyHelmOptions) map[string]any {
	persistence := map[string]any{"enabled": opts.Persistence}
	if opts.Persistence {
		persistence["size"] = opts.StorageSize
		if opts.StorageClass != "" {
			persistence["storageClass"] = opts.StorageClass
		}
	}
	return map[string]any{
		"fullnameOverride": redisReleaseName(opts.CRName),
		"architecture":     "standalone",
		"image": map[string]any{
			"registry":   valkeyImageRegistry,
			"repository": valkeyImageRepository,
			"tag":        valkeyImageTag,
		},
		"auth": map[string]any{
			"existingSecret":            RedisSecretName(opts.CRName),
			"existingSecretPasswordKey": RedisSecretKey,
		},
		// The chart renders commonConfiguration verbatim into valkey.conf;
		// setting it REPLACES the chart's own two lines (appendonly yes,
		// save ""), so every line the server needs is written here.
		"commonConfiguration": valkeyServerConfiguration(opts),
		"primary": map[string]any{
			"resourcesPreset": valkeyResourcesPresetNone,
			"resources":       helmResourceValues(opts.Resources),
			"persistence":     persistence,
		},
	}
}

// valkeyServerConfiguration is the valkey.conf the chart mounts: the ceiling
// and the eviction policy always; the append-only file only with a volume to
// keep it on (an AOF on an emptyDir is rewritten for nothing). RDB snapshots
// stay off in both postures, the chart's own choice.
func valkeyServerConfiguration(opts ValkeyHelmOptions) string {
	appendOnly := valkeyConfNo
	if opts.Persistence {
		appendOnly = valkeyConfYes
	}
	return strings.Join([]string{
		"# The dataset ceiling and what happens at it -- always set for this store.",
		"maxmemory " + opts.MaxMemory,
		"maxmemory-policy " + opts.MaxMemoryPolicy,
		"# Append-only-file persistence, on only when a volume backs it.",
		"appendonly " + appendOnly,
		"# RDB snapshots off: the AOF is the durability posture, or none is.",
		`save ""`,
	}, "\n")
}

// redisReleaseName returns the Helm release name: "{crName}-redis".
func redisReleaseName(crName string) string {
	return fmt.Sprintf("%s-redis", crName)
}

// RedisSecretName returns the credential Secret name: "{crName}-redis-credentials".
func RedisSecretName(crName string) string {
	return fmt.Sprintf("%s-redis-credentials", crName)
}

// RedisServiceHost returns the in-cluster DNS hostname for the cache server's
// primary Service created by the Valkey chart:
// "{crName}-redis-primary.{namespace}.svc.cluster.local".
func RedisServiceHost(crName, namespace string) string {
	return fmt.Sprintf("%s-redis-primary.%s.svc.cluster.local", crName, namespace)
}
