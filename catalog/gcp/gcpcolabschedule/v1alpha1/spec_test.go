package gcpcolabschedulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpColabScheduleSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpColabScheduleSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	notebook := func() *GcpColabScheduleNotebookExecutionJob {
		return &GcpColabScheduleNotebookExecutionJob{
			GcsNotebookSource:                   &GcpColabScheduleGcsNotebookSource{Uri: "gs://ml-notebooks/report.ipynb"},
			NotebookRuntimeTemplateResourceName: litRef("projects/p/locations/us-central1/notebookRuntimeTemplates/standard"),
			GcsOutputUri:                        litRef("gs://ml-notebook-runs"),
			ServiceAccount:                      litRef("notebooks@p.iam.gserviceaccount.com"),
		}
	}

	minimal := func() *GcpColabSchedule {
		return &GcpColabSchedule{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpColabSchedule",
			Metadata:   &shared.CloudResourceMetadata{Name: "nightly-report"},
			Spec: &GcpColabScheduleSpec{
				Location:              "us-central1",
				Cron:                  "0 6 * * *",
				MaxConcurrentRunCount: 1,
				NotebookExecutionJob:  notebook(),
			},
		}
	}

	ginkgo.It("should accept the smallest notebook schedule", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every notebook field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.DisplayName = "Nightly report"
		msg.Spec.Cron = "TZ=America/New_York 0 9 * * 1-5"
		msg.Spec.AllowQueueing = true
		msg.Spec.StartTime = "2026-10-01T00:00:00Z"
		msg.Spec.EndTime = "2027-10-01T00:00:00.5+05:30"
		msg.Spec.MaxRunCount = 365
		msg.Spec.DesiredState = "PAUSED"
		msg.Spec.DeletionPolicy = "PREVENT"
		job := msg.Spec.NotebookExecutionJob
		job.DisplayName = "Report run"
		job.GcsNotebookSource = nil
		job.DataformRepositorySource = &GcpColabScheduleDataformRepositorySource{DataformRepositoryResourceName: "projects/p/locations/us-central1/repositories/notebooks", CommitSha: "abc123"}
		job.NotebookRuntimeTemplateResourceName = nil
		job.CustomEnvironmentSpec = &GcpColabScheduleCustomEnvironmentSpec{
			MachineSpec: &GcpColabScheduleMachineSpec{
				MachineType: "a2-highgpu-1g", AcceleratorType: "NVIDIA_TESLA_A100", AcceleratorCount: 1, GpuPartitionSize: "1g.5gb", TpuTopology: "",
				ReservationAffinity: &GcpColabScheduleReservationAffinity{ReservationAffinityType: "SPECIFIC_RESERVATION", Key: "compute.googleapis.com/reservation-name", Values: []string{"projects/p/zones/us-central1-a/reservations/r"}, UseReservationPool: false},
			},
			NetworkSpec:        &GcpColabScheduleNetworkSpec{EnableInternetAccess: true, Network: litRef("projects/p/global/networks/n"), Subnetwork: litRef("projects/p/regions/us-central1/subnetworks/s")},
			PersistentDiskSpec: &GcpColabSchedulePersistentDiskSpec{DiskType: "pd-ssd", DiskSizeGb: 200},
		}
		job.WorkbenchRuntime = true
		job.ServiceAccount = nil
		job.ExecutionUser = "alice@example.com"
		job.ExecutionTimeout = "3600s"
		job.KernelName = "python3"
		job.Labels = map[string]string{"report": "daily"}
		job.KmsKeyName = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every pipeline field set", func() {
		msg := minimal()
		msg.Spec.NotebookExecutionJob = nil
		msg.Spec.MaxConcurrentActiveRunCount = 2
		msg.Spec.PipelineJob = &GcpColabSchedulePipelineJob{
			DisplayName:          "Retrain",
			PipelineSpec:         `{"pipelineInfo":{"name":"retrain"}}`,
			TemplateUri:          "https://us-central1-kfp.pkg.dev/p/pipelines/retrain/v1",
			RuntimeConfig:        &GcpColabSchedulePipelineRuntimeConfig{GcsOutputDirectory: "gs://ml-pipelines/root", FailurePolicy: "PIPELINE_FAILURE_POLICY_FAIL_FAST", ParameterValues: map[string]string{"epochs": "10"}},
			ServiceAccount:       litRef("pipelines@p.iam.gserviceaccount.com"),
			Network:              litRef("projects/123456789012/global/networks/ml-vpc"),
			ReservedIpRanges:     []string{"vertex-ai-ip-range"},
			PreflightValidations: true,
			Labels:               map[string]string{"model": "churn"},
			KmsKeyName:           litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k"),
			PscInterfaceConfig: &GcpColabSchedulePscInterfaceConfig{
				NetworkAttachment: "projects/p/regions/us-central1/networkAttachments/vertex",
				DnsPeeringConfigs: []*GcpColabScheduleDnsPeeringConfig{{Domain: "corp.example.com.", TargetNetwork: litRef("corp-vpc"), TargetProject: litRef("corp-dns")}},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one run arm", func() {
		msg := minimal()
		msg.Spec.NotebookExecutionJob = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.PipelineJob = &GcpColabSchedulePipelineJob{RuntimeConfig: &GcpColabSchedulePipelineRuntimeConfig{GcsOutputDirectory: "gs://b"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a cron and at least one concurrent run", func() {
		msg := minimal()
		msg.Spec.Cron = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.MaxConcurrentRunCount = 0
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one notebook source", func() {
		msg := minimal()
		msg.Spec.NotebookExecutionJob.GcsNotebookSource = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NotebookExecutionJob.DataformRepositorySource = &GcpColabScheduleDataformRepositorySource{DataformRepositoryResourceName: "projects/p/locations/l/repositories/r"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one environment", func() {
		msg := minimal()
		msg.Spec.NotebookExecutionJob.NotebookRuntimeTemplateResourceName = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NotebookExecutionJob.CustomEnvironmentSpec = &GcpColabScheduleCustomEnvironmentSpec{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one identity", func() {
		msg := minimal()
		msg.Spec.NotebookExecutionJob.ServiceAccount = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NotebookExecutionJob.ExecutionUser = "alice@example.com"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NotebookExecutionJob.ServiceAccount = nil
		msg.Spec.NotebookExecutionJob.ExecutionUser = "alice"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require an output location and a gs:// notebook", func() {
		msg := minimal()
		msg.Spec.NotebookExecutionJob.GcsOutputUri = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NotebookExecutionJob.GcsNotebookSource.Uri = "https://example.com/report.ipynb"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject malformed timestamps, timeouts, and generations", func() {
		msg := minimal()
		msg.Spec.StartTime = "2026-10-01"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NotebookExecutionJob.ExecutionTimeout = "1h"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NotebookExecutionJob.GcsNotebookSource.Generation = "latest"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown desired state and reservation affinity", func() {
		msg := minimal()
		msg.Spec.DesiredState = "STOPPED"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NotebookExecutionJob.NotebookRuntimeTemplateResourceName = nil
		msg.Spec.NotebookExecutionJob.CustomEnvironmentSpec = &GcpColabScheduleCustomEnvironmentSpec{MachineSpec: &GcpColabScheduleMachineSpec{ReservationAffinity: &GcpColabScheduleReservationAffinity{ReservationAffinityType: "ANY"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a pipeline output root and a dotted DNS domain", func() {
		msg := minimal()
		msg.Spec.NotebookExecutionJob = nil
		msg.Spec.PipelineJob = &GcpColabSchedulePipelineJob{RuntimeConfig: &GcpColabSchedulePipelineRuntimeConfig{}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.PipelineJob = &GcpColabSchedulePipelineJob{
			RuntimeConfig:      &GcpColabSchedulePipelineRuntimeConfig{GcsOutputDirectory: "gs://b"},
			PscInterfaceConfig: &GcpColabSchedulePscInterfaceConfig{DnsPeeringConfigs: []*GcpColabScheduleDnsPeeringConfig{{Domain: "corp.example.com", TargetNetwork: litRef("n"), TargetProject: litRef("p")}}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
