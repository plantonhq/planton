package gcpcloudrunworkerpoolv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudRunWorkerPoolSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpCloudRunWorkerPoolSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// One worker container, manual scaling left to Google's default.
	minimal := func() *GcpCloudRunWorkerPool {
		return &GcpCloudRunWorkerPool{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudRunWorkerPool",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders-worker"},
			Spec: &GcpCloudRunWorkerPoolSpec{
				Region: "us-central1",
				Containers: []*GcpCloudRunWorkerPoolContainer{{
					Image: "us-docker.pkg.dev/p/repo/worker:1.0.0",
				}},
			},
		}
	}

	ginkgo.It("should accept the minimal worker pool", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a blank image slot for pipeline injection", func() {
		msg := minimal()
		msg.Spec.Containers[0].Image = ""
		// The artifact_image_slot contract leaves the image blank for the
		// pipeline; the spec still requires a non-empty image at validate
		// time, exactly as the Cloud Run service does.
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept every optional lever together", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("workers-project")
		msg.Spec.WorkerPoolName = "orders-worker"
		msg.Spec.Description = "pulls order events"
		msg.Spec.Labels = map[string]string{"team": "orders"}
		msg.Spec.Annotations = map[string]string{"owner": "orders"}
		msg.Spec.Containers = []*GcpCloudRunWorkerPoolContainer{
			{
				Name:       "worker",
				Image:      "us-docker.pkg.dev/p/repo/worker:1.0.0",
				Command:    []string{"/bin/worker"},
				Args:       []string{"--pull"},
				WorkingDir: "/app",
				Env: []*GcpCloudRunWorkerPoolEnvVar{
					{Name: "SUBSCRIPTION", Value: "orders-sub"},
					{Name: "DB_PASSWORD", ValueFromSecret: &GcpCloudRunWorkerPoolSecretEnvSource{Secret: "db-password", Version: "latest"}},
				},
				Resources:    &GcpCloudRunWorkerPoolContainerResources{Cpu: "2", Memory: "1Gi"},
				VolumeMounts: []*GcpCloudRunWorkerPoolVolumeMount{{Name: "scratch", MountPath: "/tmp/scratch"}},
				StartupProbe: &GcpCloudRunWorkerPoolStartupProbe{
					PeriodSeconds:    proto.Int32(10),
					FailureThreshold: proto.Int32(6),
					Handler:          &GcpCloudRunWorkerPoolStartupProbe_TcpSocket{TcpSocket: &GcpCloudRunWorkerPoolTcpSocketAction{Port: proto.Int32(8081)}},
				},
				LivenessProbe: &GcpCloudRunWorkerPoolLivenessProbe{
					Handler: &GcpCloudRunWorkerPoolLivenessProbe_HttpGet{HttpGet: &GcpCloudRunWorkerPoolHttpGetAction{
						Path: "/healthz", Port: proto.Int32(8081),
						HttpHeaders: []*GcpCloudRunWorkerPoolHttpHeader{{Name: "X-Probe", Value: "1"}},
					}},
				},
				DependsOn: []string{"otel"},
			},
			{Name: "otel", Image: "otel/opentelemetry-collector:0.100.0"},
		}
		msg.Spec.Volumes = []*GcpCloudRunWorkerPoolVolume{
			{Name: "scratch", Source: &GcpCloudRunWorkerPoolVolume_EmptyDir{EmptyDir: &GcpCloudRunWorkerPoolVolumeEmptyDir{Medium: "MEMORY", SizeLimit: "256Mi"}}},
			{Name: "db", Source: &GcpCloudRunWorkerPoolVolume_CloudSqlInstance{CloudSqlInstance: &GcpCloudRunWorkerPoolVolumeCloudSql{Instances: []*foreignkeyv1.StringValueOrRef{nameRef("orders-db")}}}},
			{Name: "certs", Source: &GcpCloudRunWorkerPoolVolume_Secret{Secret: &GcpCloudRunWorkerPoolVolumeSecret{Secret: "tls", DefaultMode: proto.Int32(292), Items: []*GcpCloudRunWorkerPoolVolumeSecretItem{{Path: "tls.crt", Version: "3", Mode: proto.Int32(256)}}}}},
			{Name: "media", Source: &GcpCloudRunWorkerPoolVolume_Gcs{Gcs: &GcpCloudRunWorkerPoolVolumeGcs{Bucket: nameRef("media-bucket"), ReadOnly: true, MountOptions: []string{"implicit-dirs"}}}},
			{Name: "share", Source: &GcpCloudRunWorkerPoolVolume_Nfs{Nfs: &GcpCloudRunWorkerPoolVolumeNfs{Server: "10.0.0.2", Path: "/share1", ReadOnly: false}}},
		}
		msg.Spec.ServiceAccount = nameRef("orders-worker-sa")
		msg.Spec.Scaling = &GcpCloudRunWorkerPoolScaling{ScalingMode: "AUTOMATIC", MinInstanceCount: proto.Int32(1), MaxInstanceCount: proto.Int32(10)}
		msg.Spec.InstanceSplits = []*GcpCloudRunWorkerPoolInstanceSplit{
			{Type: "INSTANCE_SPLIT_ALLOCATION_TYPE_LATEST", Percent: proto.Int32(90)},
			{Type: "INSTANCE_SPLIT_ALLOCATION_TYPE_REVISION", Revision: "orders-worker-v41", Percent: proto.Int32(10)},
		}
		msg.Spec.EncryptionKey = nameRef("cmek")
		msg.Spec.EncryptionKeyRevocationAction = "SHUTDOWN"
		msg.Spec.EncryptionKeyShutdownDuration = "3600s"
		msg.Spec.Revision = "orders-worker-v42"
		msg.Spec.RevisionLabels = map[string]string{"build": "42"}
		msg.Spec.RevisionAnnotations = map[string]string{"commit": "abc"}
		msg.Spec.VpcAccess = &GcpCloudRunWorkerPoolVpcAccess{
			NetworkInterfaces: []*GcpCloudRunWorkerPoolNetworkInterface{{Network: nameRef("vpc"), Subnetwork: nameRef("subnet"), Tags: []string{"worker"}}},
			Egress:            "ALL_TRAFFIC",
		}
		msg.Spec.NodeSelector = &GcpCloudRunWorkerPoolNodeSelector{Accelerator: "nvidia-l4"}
		msg.Spec.GpuZonalRedundancyDisabled = true
		msg.Spec.LaunchStage = "BETA"
		msg.Spec.BinaryAuthorization = &GcpCloudRunWorkerPoolBinaryAuthorization{UseDefault: true, BreakglassJustification: "hotfix"}
		msg.Spec.DeletionProtection = proto.Bool(false)
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require region and at least one container", func() {
		msg := minimal()
		msg.Spec.Region = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Containers = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a global region -- worker pools are regional only", func() {
		msg := minimal()
		msg.Spec.Region = "global"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should tie manual and automatic scaling fields to their modes", func() {
		msg := minimal()
		msg.Spec.Scaling = &GcpCloudRunWorkerPoolScaling{ManualInstanceCount: proto.Int32(3)}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), "unset mode is MANUAL")

		msg.Spec.Scaling = &GcpCloudRunWorkerPoolScaling{ScalingMode: "AUTOMATIC", ManualInstanceCount: proto.Int32(3)}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg.Spec.Scaling = &GcpCloudRunWorkerPoolScaling{ScalingMode: "MANUAL", MinInstanceCount: proto.Int32(1)}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg.Spec.Scaling = &GcpCloudRunWorkerPoolScaling{ScalingMode: "AUTOMATIC", MinInstanceCount: proto.Int32(5), MaxInstanceCount: proto.Int32(2)}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg.Spec.Scaling = &GcpCloudRunWorkerPoolScaling{ScalingMode: "MANUAL", ManualInstanceCount: proto.Int32(0)}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), "zero parks the pool")
	})

	ginkgo.It("should enforce instance split coherence", func() {
		msg := minimal()
		msg.Spec.InstanceSplits = []*GcpCloudRunWorkerPoolInstanceSplit{{Type: "INSTANCE_SPLIT_ALLOCATION_TYPE_REVISION", Percent: proto.Int32(100)}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "REVISION without a revision")
		msg.Spec.InstanceSplits = []*GcpCloudRunWorkerPoolInstanceSplit{{Type: "INSTANCE_SPLIT_ALLOCATION_TYPE_LATEST", Revision: "x", Percent: proto.Int32(100)}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "LATEST with a revision")
		msg.Spec.InstanceSplits = []*GcpCloudRunWorkerPoolInstanceSplit{{Type: "LATEST"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "an unknown type")
		msg.Spec.InstanceSplits = []*GcpCloudRunWorkerPoolInstanceSplit{{Type: "INSTANCE_SPLIT_ALLOCATION_TYPE_LATEST", Percent: proto.Int32(101)}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "percent above 100")
	})

	ginkgo.It("should tie the encryption revocation levers together", func() {
		msg := minimal()
		msg.Spec.EncryptionKeyRevocationAction = "SHUTDOWN"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "an action without a key")
		msg.Spec.EncryptionKey = nameRef("cmek")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.EncryptionKeyRevocationAction = "PREVENT_NEW"
		msg.Spec.EncryptionKeyShutdownDuration = "3600s"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a shutdown duration without SHUTDOWN")
		msg.Spec.EncryptionKeyRevocationAction = "SHUTDOWN"
		msg.Spec.EncryptionKeyShutdownDuration = "1h"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a duration not in seconds")
		msg.Spec.EncryptionKeyRevocationAction = "DELETE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "an unknown action")
	})

	ginkgo.It("should require an accelerator before disabling GPU zonal redundancy", func() {
		msg := minimal()
		msg.Spec.GpuZonalRedundancyDisabled = true
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a connector beside network interfaces", func() {
		msg := minimal()
		msg.Spec.VpcAccess = &GcpCloudRunWorkerPoolVpcAccess{
			Connector:         nameRef("conn"),
			NetworkInterfaces: []*GcpCloudRunWorkerPoolNetworkInterface{{Subnetwork: nameRef("subnet")}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.VpcAccess.Connector = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.VpcAccess.Egress = "PUBLIC"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an env var with both a value and a secret", func() {
		msg := minimal()
		msg.Spec.Containers[0].Env = []*GcpCloudRunWorkerPoolEnvVar{{
			Name: "X", Value: "literal",
			ValueFromSecret: &GcpCloudRunWorkerPoolSecretEnvSource{Secret: "s"},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Containers[0].Env = []*GcpCloudRunWorkerPoolEnvVar{{Name: "1X", Value: "v"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a name starting with a digit")
	})

	ginkgo.It("should take exactly one of a value, a secret reference, or a secret value", func() {
		msg := minimal()
		msg.Spec.Containers[0].Env = []*GcpCloudRunWorkerPoolEnvVar{{Name: "TOKEN", SecretValue: "s3cr3t"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), "a secret value alone")
		msg.Spec.Containers[0].Env = []*GcpCloudRunWorkerPoolEnvVar{{Name: "TOKEN", Value: "literal", SecretValue: "s3cr3t"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a value beside a secret value")
		msg.Spec.Containers[0].Env = []*GcpCloudRunWorkerPoolEnvVar{{
			Name: "TOKEN", SecretValue: "s3cr3t",
			ValueFromSecret: &GcpCloudRunWorkerPoolSecretEnvSource{Secret: "s"},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a secret reference beside a secret value")
	})

	ginkgo.It("should enforce the probe timing rules", func() {
		msg := minimal()
		msg.Spec.Containers[0].StartupProbe = &GcpCloudRunWorkerPoolStartupProbe{
			PeriodSeconds:    proto.Int32(60),
			FailureThreshold: proto.Int32(5),
			Handler:          &GcpCloudRunWorkerPoolStartupProbe_TcpSocket{TcpSocket: &GcpCloudRunWorkerPoolTcpSocketAction{Port: proto.Int32(8081)}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a 300-second startup window")
		msg.Spec.Containers[0].StartupProbe = &GcpCloudRunWorkerPoolStartupProbe{
			TimeoutSeconds: proto.Int32(20), PeriodSeconds: proto.Int32(10),
			Handler: &GcpCloudRunWorkerPoolStartupProbe_TcpSocket{TcpSocket: &GcpCloudRunWorkerPoolTcpSocketAction{Port: proto.Int32(8081)}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "timeout above period")
		msg.Spec.Containers[0].StartupProbe = &GcpCloudRunWorkerPoolStartupProbe{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a probe without a handler")
		msg.Spec.Containers[0].StartupProbe = nil
		msg.Spec.Containers[0].LivenessProbe = &GcpCloudRunWorkerPoolLivenessProbe{
			Handler: &GcpCloudRunWorkerPoolLivenessProbe_Grpc{Grpc: &GcpCloudRunWorkerPoolGrpcAction{Port: proto.Int32(9090), Service: "worker"}},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one volume source and valid mounts", func() {
		msg := minimal()
		msg.Spec.Volumes = []*GcpCloudRunWorkerPoolVolume{{Name: "empty"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a volume without a source")
		msg.Spec.Volumes = []*GcpCloudRunWorkerPoolVolume{{Name: "Bad_Name", Source: &GcpCloudRunWorkerPoolVolume_EmptyDir{EmptyDir: &GcpCloudRunWorkerPoolVolumeEmptyDir{}}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a volume name that is not a DNS label")
		msg.Spec.Volumes = nil
		msg.Spec.Containers[0].VolumeMounts = []*GcpCloudRunWorkerPoolVolumeMount{{Name: "x", MountPath: "relative"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a relative mount path")
	})

	ginkgo.It("should validate resource limit formats", func() {
		msg := minimal()
		msg.Spec.Containers[0].Resources = &GcpCloudRunWorkerPoolContainerResources{Cpu: "two"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Containers[0].Resources = &GcpCloudRunWorkerPoolContainerResources{Memory: "512MB"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should cap probe headers at one until the SDK widens", func() {
		msg := minimal()
		msg.Spec.Containers[0].LivenessProbe = &GcpCloudRunWorkerPoolLivenessProbe{
			Handler: &GcpCloudRunWorkerPoolLivenessProbe_HttpGet{HttpGet: &GcpCloudRunWorkerPoolHttpGetAction{
				Path: "/healthz", Port: proto.Int32(8081),
				HttpHeaders: []*GcpCloudRunWorkerPoolHttpHeader{{Name: "A", Value: "1"}, {Name: "B", Value: "2"}},
			}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a launch stage Cloud Run does not support", func() {
		msg := minimal()
		msg.Spec.LaunchStage = "PRELAUNCH"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a binary authorization block naming both the default and a policy", func() {
		msg := minimal()
		msg.Spec.BinaryAuthorization = &GcpCloudRunWorkerPoolBinaryAuthorization{UseDefault: true, Policy: "projects/p/platforms/cloudRun/policies/x"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
