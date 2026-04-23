package module

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	gcpcloudrunv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpcloudrun/v1"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// service creates a Cloud Run v2 Service and applies optional public access.
func service(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*cloudrunv2.Service, error) {

	image := fmt.Sprintf("%s:%s",
		locals.GcpCloudRun.Spec.Container.Image.Repo,
		locals.GcpCloudRun.Spec.Container.Image.Tag)

	memory := fmt.Sprintf("%dMi", locals.GcpCloudRun.Spec.Container.Memory)
	cpu := strconv.Itoa(int(locals.GcpCloudRun.Spec.Container.Cpu))

	port := int32(8080)
	if locals.GcpCloudRun.Spec.Container.Port != 0 {
		port = locals.GcpCloudRun.Spec.Container.Port
	}

	// Determine service name: use spec.service_name if provided, otherwise metadata.name
	serviceName := locals.GcpCloudRun.Metadata.Name
	if locals.GcpCloudRun.Spec.ServiceName != "" {
		serviceName = locals.GcpCloudRun.Spec.ServiceName
	}

	// Service account: use provided or default
	var serviceAccountEmail pulumi.StringPtrInput
	if locals.GcpCloudRun.Spec.ServiceAccount != "" {
		serviceAccountEmail = pulumi.String(locals.GcpCloudRun.Spec.ServiceAccount)
	}

	// Max concurrency: use provided or default to 80
	maxConcurrency := int32(80)
	if locals.GcpCloudRun.Spec.MaxConcurrency != 0 {
		maxConcurrency = locals.GcpCloudRun.Spec.MaxConcurrency
	}

	// Timeout: use provided or default to 300 seconds
	timeout := int32(300)
	if locals.GcpCloudRun.Spec.TimeoutSeconds != 0 {
		timeout = locals.GcpCloudRun.Spec.TimeoutSeconds
	}

	// Ingress: convert enum to string value
	ingressValue := toIngressString(locals.GcpCloudRun.Spec.Ingress)

	// Execution environment: convert enum to string value
	executionEnv := toExecutionEnvironmentString(locals.GcpCloudRun.Spec.ExecutionEnvironment)

	// VPC access configuration
	var vpcAccess *cloudrunv2.ServiceTemplateVpcAccessArgs
	if locals.GcpCloudRun.Spec.VpcAccess != nil {
		networkValue := locals.GcpCloudRun.Spec.VpcAccess.Network.GetValue()
		subnetValue := locals.GcpCloudRun.Spec.VpcAccess.Subnet.GetValue()
		if networkValue != "" || subnetValue != "" {
			vpcAccess = &cloudrunv2.ServiceTemplateVpcAccessArgs{}
			if networkValue != "" {
				vpcAccess.NetworkInterfaces = cloudrunv2.ServiceTemplateVpcAccessNetworkInterfaceArray{
					&cloudrunv2.ServiceTemplateVpcAccessNetworkInterfaceArgs{
						Network:    pulumi.String(networkValue),
						Subnetwork: pulumi.String(subnetValue),
					},
				}
			}
			if locals.GcpCloudRun.Spec.VpcAccess.Egress != "" {
				vpcAccess.Egress = pulumi.String(locals.GcpCloudRun.Spec.VpcAccess.Egress)
			}
		}
	}

	// Build the main application container.
	appContainer := &cloudrunv2.ServiceTemplateContainerArgs{
		Image: pulumi.String(image),

		Ports: &cloudrunv2.ServiceTemplateContainerPortsArgs{
			ContainerPort: pulumi.Int(port),
		},

		Resources: &cloudrunv2.ServiceTemplateContainerResourcesArgs{
			Limits: pulumi.StringMap{
				"memory": pulumi.String(memory),
				"cpu":    pulumi.String(cpu),
			},
		},

		Envs: toEnvArray(locals),
	}

	containers := cloudrunv2.ServiceTemplateContainerArray{appContainer}
	var volumes cloudrunv2.ServiceTemplateVolumeArray

	// Cloud SQL connectivity configuration.
	if locals.GcpCloudRun.Spec.CloudSql != nil {
		cloudSql := locals.GcpCloudRun.Spec.CloudSql

		if cloudSql.Connection != nil {
			// Native Cloud SQL connection: volume mount with Unix sockets.
			instanceNames := collectCloudSqlInstanceNames(cloudSql.Connection.Instances)

			volumes = append(volumes, &cloudrunv2.ServiceTemplateVolumeArgs{
				Name: pulumi.String("cloudsql"),
				CloudSqlInstance: &cloudrunv2.ServiceTemplateVolumeCloudSqlInstanceArgs{
					Instances: instanceNames,
				},
			})

			appContainer.VolumeMounts = cloudrunv2.ServiceTemplateContainerVolumeMountArray{
				&cloudrunv2.ServiceTemplateContainerVolumeMountArgs{
					Name:      pulumi.String("cloudsql"),
					MountPath: pulumi.String("/cloudsql"),
				},
			}
		} else if cloudSql.AuthProxy != nil {
			// Auth Proxy sidecar: second container running cloud-sql-proxy.
			appContainer.Name = pulumi.StringPtr("app")
			appContainer.DependsOns = pulumi.StringArray{pulumi.String("cloud-sql-proxy")}

			proxyArgs := buildAuthProxyArgs(cloudSql.AuthProxy)
			containers = append(containers, &cloudrunv2.ServiceTemplateContainerArgs{
				Name:  pulumi.StringPtr("cloud-sql-proxy"),
				Image: pulumi.String("gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.15.2"),
				Args:  proxyArgs,
				Resources: &cloudrunv2.ServiceTemplateContainerResourcesArgs{
					Limits: pulumi.StringMap{
						"memory": pulumi.String("256Mi"),
						"cpu":    pulumi.String("1"),
					},
				},
			})
		}
	}

	template := &cloudrunv2.ServiceTemplateArgs{
		ServiceAccount:                serviceAccountEmail,
		Timeout:                       pulumi.String(fmt.Sprintf("%ds", timeout)),
		ExecutionEnvironment:          pulumi.String(executionEnv),
		MaxInstanceRequestConcurrency: pulumi.Int(maxConcurrency),
		VpcAccess:                     vpcAccess,
		Containers:                    containers,
		Scaling: &cloudrunv2.ServiceTemplateScalingArgs{
			MinInstanceCount: pulumi.Int(int(locals.GcpCloudRun.Spec.Container.Replicas.Min)),
			MaxInstanceCount: pulumi.Int(int(locals.GcpCloudRun.Spec.Container.Replicas.Max)),
		},
	}

	if len(volumes) > 0 {
		template.Volumes = volumes
	}

	createdService, err := cloudrunv2.NewService(ctx,
		locals.GcpCloudRun.Metadata.Name,
		&cloudrunv2.ServiceArgs{
			Project:            pulumi.String(locals.GcpCloudRun.Spec.ProjectId.GetValue()),
			Location:           pulumi.String(locals.GcpCloudRun.Spec.Region),
			Name:               pulumi.String(serviceName),
			InvokerIamDisabled: pulumi.BoolPtr(locals.GcpCloudRun.Spec.AllowUnauthenticated),
			Ingress:            pulumi.String(ingressValue),
			DeletionProtection: pulumi.Bool(locals.GcpCloudRun.Spec.DeleteProtection),
			Template:           template,
			Labels:             pulumi.ToStringMap(locals.GcpLabels),
		},
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Cloud Run v2 service")
	}

	return createdService, nil
}

// collectCloudSqlInstanceNames extracts connection name strings from FK references.
func collectCloudSqlInstanceNames(instances []*foreignkeyv1.StringValueOrRef) pulumi.StringArrayInput {
	names := pulumi.StringArray{}
	for _, inst := range instances {
		names = append(names, pulumi.String(inst.GetValue()))
	}
	return names
}

// buildAuthProxyArgs constructs the command-line arguments for the cloud-sql-proxy sidecar.
func buildAuthProxyArgs(proxy *gcpcloudrunv1.GcpCloudRunCloudSqlAuthProxy) pulumi.StringArrayInput {
	args := pulumi.StringArray{}

	port := int32(5432)
	if proxy.Port != 0 {
		port = proxy.Port
	}
	args = append(args, pulumi.String(fmt.Sprintf("--port=%d", port)))

	if proxy.UsePrivateIp {
		args = append(args, pulumi.String("--private-ip"))
	}

	for _, inst := range proxy.Instances {
		args = append(args, pulumi.String(inst.GetValue()))
	}

	return args
}

// toEnvArray converts proto env / secret maps into the v2 container-env slice.
func toEnvArray(locals *Locals) cloudrunv2.ServiceTemplateContainerEnvArray {
	envs := cloudrunv2.ServiceTemplateContainerEnvArray{}

	if locals.GcpCloudRun.Spec.Container.Env == nil {
		return envs
	}

	for k, v := range locals.GcpCloudRun.Spec.Container.Env.Variables {
		envs = append(envs, &cloudrunv2.ServiceTemplateContainerEnvArgs{
			Name:  pulumi.String(k),
			Value: pulumi.String(v),
		})
	}

	for k, v := range locals.GcpCloudRun.Spec.Container.Env.Secrets {
		envs = append(envs, &cloudrunv2.ServiceTemplateContainerEnvArgs{
			Name: pulumi.String(k),
			ValueSource: &cloudrunv2.ServiceTemplateContainerEnvValueSourceArgs{
				SecretKeyRef: &cloudrunv2.ServiceTemplateContainerEnvValueSourceSecretKeyRefArgs{
					Secret: pulumi.String(v), // projects/*/secrets/*:version
				},
			},
		})
	}

	return envs
}

// toIngressString converts the proto enum to the GCP API string value
func toIngressString(ingress gcpcloudrunv1.GcpCloudRunIngress) string {
	switch ingress {
	case gcpcloudrunv1.GcpCloudRunIngress_INGRESS_TRAFFIC_INTERNAL_ONLY:
		return "INGRESS_TRAFFIC_INTERNAL_ONLY"
	case gcpcloudrunv1.GcpCloudRunIngress_INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER:
		return "INGRESS_TRAFFIC_INTERNAL_AND_CLOUD_LOAD_BALANCING"
	default:
		return "INGRESS_TRAFFIC_ALL"
	}
}

// toExecutionEnvironmentString converts the proto enum to the GCP API string value
func toExecutionEnvironmentString(env gcpcloudrunv1.GcpCloudRunExecutionEnvironment) string {
	switch env {
	case gcpcloudrunv1.GcpCloudRunExecutionEnvironment_EXECUTION_ENVIRONMENT_GEN1:
		return "EXECUTION_ENVIRONMENT_GEN1"
	default:
		return "EXECUTION_ENVIRONMENT_GEN2"
	}
}
