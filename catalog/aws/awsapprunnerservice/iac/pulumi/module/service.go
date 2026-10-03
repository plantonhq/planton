package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apprunner"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/wafv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// service creates the App Runner service, its custom domain associations,
// and the optional WAF association, then exports outputs.
//
// The service's shared companions -- auto scaling configuration, VPC
// connector, and observability configuration -- are separate first-class
// resources referenced by ARN, never created here: each is designed by AWS
// to be shared across services, so embedding one per service would fork what
// should be tuned in one place.
func service(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsAppRunnerService.Spec
	serviceName := locals.AwsAppRunnerService.Metadata.Name

	// --- Source configuration -------------------------------------------
	// Exactly one of image/code repository is present (spec CEL). Port,
	// start command, and the env var/secret maps live at the spec top level
	// and are routed into whichever arm is active.
	sourceArgs := &apprunner.ServiceSourceConfigurationArgs{
		// Sent explicitly (never defaulted) because the honest default is
		// conditional on the source: AWS enables auto-deploy for code repos
		// and private ECR only, and REJECTS it for ECR_PUBLIC (spec CEL
		// guards the invalid combination before AWS would).
		AutoDeploymentsEnabled: pulumi.Bool(spec.AutoDeploymentsEnabled),
	}

	// Private ECR pulls use the access role; code repositories use the
	// out-of-band App Runner connection. The two never coexist because the
	// source arms are mutually exclusive.
	if spec.ImageSource != nil && spec.ImageSource.AccessRoleArn.GetValue() != "" {
		sourceArgs.AuthenticationConfiguration = &apprunner.ServiceSourceConfigurationAuthenticationConfigurationArgs{
			AccessRoleArn: pulumi.StringPtr(spec.ImageSource.AccessRoleArn.GetValue()),
		}
	}
	if spec.CodeSource != nil {
		sourceArgs.AuthenticationConfiguration = &apprunner.ServiceSourceConfigurationAuthenticationConfigurationArgs{
			ConnectionArn: pulumi.StringPtr(spec.CodeSource.ConnectionArn.GetValue()),
		}
	}

	// Runtime settings shared by both source arms.
	var port pulumi.StringPtrInput
	if spec.Port != nil {
		port = pulumi.StringPtr(*spec.Port)
	}
	var startCommand pulumi.StringPtrInput
	if spec.StartCommand != "" {
		startCommand = pulumi.StringPtr(spec.StartCommand)
	}
	var envVars pulumi.StringMapInput
	if len(spec.EnvironmentVariables) > 0 {
		envVars = pulumi.ToStringMap(spec.EnvironmentVariables)
	}
	var envSecrets pulumi.StringMapInput
	if len(spec.EnvironmentSecrets) > 0 {
		envSecrets = pulumi.ToStringMap(spec.EnvironmentSecrets)
	}

	if spec.ImageSource != nil {
		sourceArgs.ImageRepository = &apprunner.ServiceSourceConfigurationImageRepositoryArgs{
			ImageIdentifier:     pulumi.String(spec.ImageSource.ImageIdentifier),
			ImageRepositoryType: pulumi.String(spec.ImageSource.ImageRepositoryType),
			ImageConfiguration: &apprunner.ServiceSourceConfigurationImageRepositoryImageConfigurationArgs{
				Port:                        port,
				StartCommand:                startCommand,
				RuntimeEnvironmentVariables: envVars,
				RuntimeEnvironmentSecrets:   envSecrets,
			},
		}
	}

	if spec.CodeSource != nil {
		codeConfigArgs := &apprunner.ServiceSourceConfigurationCodeRepositoryCodeConfigurationArgs{
			ConfigurationSource: pulumi.String(spec.CodeSource.ConfigurationSource),
		}
		// Build settings apply only in API mode; in REPOSITORY mode App
		// Runner reads apprunner.yaml from the source directory and AWS
		// rejects inline values.
		if spec.CodeSource.ConfigurationSource == "API" {
			valuesArgs := &apprunner.ServiceSourceConfigurationCodeRepositoryCodeConfigurationCodeConfigurationValuesArgs{
				Runtime:                     pulumi.String(spec.CodeSource.Runtime),
				Port:                        port,
				StartCommand:                startCommand,
				RuntimeEnvironmentVariables: envVars,
				RuntimeEnvironmentSecrets:   envSecrets,
			}
			if spec.CodeSource.BuildCommand != "" {
				valuesArgs.BuildCommand = pulumi.StringPtr(spec.CodeSource.BuildCommand)
			}
			codeConfigArgs.CodeConfigurationValues = valuesArgs
		}

		codeRepoArgs := &apprunner.ServiceSourceConfigurationCodeRepositoryArgs{
			RepositoryUrl: pulumi.String(spec.CodeSource.RepositoryUrl),
			// BRANCH is the only source-code-version type AWS supports; the
			// spec models the branch directly rather than a one-value enum
			// wrapper.
			SourceCodeVersion: &apprunner.ServiceSourceConfigurationCodeRepositorySourceCodeVersionArgs{
				Type:  pulumi.String("BRANCH"),
				Value: pulumi.String(spec.CodeSource.Branch),
			},
			CodeConfiguration: codeConfigArgs,
		}
		if spec.CodeSource.SourceDirectory != "" {
			codeRepoArgs.SourceDirectory = pulumi.StringPtr(spec.CodeSource.SourceDirectory)
		}
		sourceArgs.CodeRepository = codeRepoArgs
	}

	// --- Instance configuration ------------------------------------------
	instanceArgs := &apprunner.ServiceInstanceConfigurationArgs{}
	if spec.Cpu != nil {
		instanceArgs.Cpu = pulumi.StringPtr(*spec.Cpu)
	}
	if spec.Memory != nil {
		instanceArgs.Memory = pulumi.StringPtr(*spec.Memory)
	}
	if spec.InstanceRoleArn.GetValue() != "" {
		instanceArgs.InstanceRoleArn = pulumi.StringPtr(spec.InstanceRoleArn.GetValue())
	}

	serviceArgs := &apprunner.ServiceArgs{
		// The cloud name is metadata.name -- the same basis the Terraform
		// module uses, so both engines create the same physical identity.
		// ForceNew: renaming replaces the service.
		ServiceName:           pulumi.String(serviceName),
		SourceConfiguration:   sourceArgs,
		InstanceConfiguration: instanceArgs,
		Tags:                  pulumi.ToStringMap(locals.AwsTags),
	}

	// --- Health check ------------------------------------------------------
	// Only sent when the spec configures it; AWS then applies TCP checks on
	// the service port with its own defaults. The path is meaningful only
	// for HTTP -- sending it alongside TCP would be silently ignored, so it
	// is omitted deliberately.
	if hc := spec.HealthCheck; hc != nil {
		hcArgs := &apprunner.ServiceHealthCheckConfigurationArgs{}
		protocol := "TCP"
		if hc.Protocol != nil {
			protocol = *hc.Protocol
		}
		hcArgs.Protocol = pulumi.StringPtr(protocol)
		if protocol == "HTTP" && hc.Path != nil {
			hcArgs.Path = pulumi.StringPtr(*hc.Path)
		}
		if hc.Interval != nil {
			hcArgs.Interval = pulumi.IntPtr(int(*hc.Interval))
		}
		if hc.Timeout != nil {
			hcArgs.Timeout = pulumi.IntPtr(int(*hc.Timeout))
		}
		if hc.HealthyThreshold != nil {
			hcArgs.HealthyThreshold = pulumi.IntPtr(int(*hc.HealthyThreshold))
		}
		if hc.UnhealthyThreshold != nil {
			hcArgs.UnhealthyThreshold = pulumi.IntPtr(int(*hc.UnhealthyThreshold))
		}
		serviceArgs.HealthCheckConfiguration = hcArgs
	}

	// --- Networking ---------------------------------------------------------
	// Always sent explicitly so the deployed shape never depends on AWS-side
	// defaults: egress routes through the referenced VPC connector when one
	// is set, ingress publicness and address family mirror the spec.
	egressArgs := &apprunner.ServiceNetworkConfigurationEgressConfigurationArgs{
		EgressType: pulumi.StringPtr("DEFAULT"),
	}
	if spec.VpcConnectorArn.GetValue() != "" {
		egressArgs.EgressType = pulumi.StringPtr("VPC")
		egressArgs.VpcConnectorArn = pulumi.StringPtr(spec.VpcConnectorArn.GetValue())
	}
	isPubliclyAccessible := true
	if spec.IsPubliclyAccessible != nil {
		isPubliclyAccessible = *spec.IsPubliclyAccessible
	}
	networkArgs := &apprunner.ServiceNetworkConfigurationArgs{
		EgressConfiguration: egressArgs,
		IngressConfiguration: &apprunner.ServiceNetworkConfigurationIngressConfigurationArgs{
			IsPubliclyAccessible: pulumi.BoolPtr(isPubliclyAccessible),
		},
	}
	if spec.IpAddressType != nil {
		networkArgs.IpAddressType = pulumi.StringPtr(*spec.IpAddressType)
	}
	serviceArgs.NetworkConfiguration = networkArgs

	// --- Encryption (ForceNew) ----------------------------------------------
	if spec.KmsKeyArn.GetValue() != "" {
		serviceArgs.EncryptionConfiguration = &apprunner.ServiceEncryptionConfigurationArgs{
			KmsKey: pulumi.String(spec.KmsKeyArn.GetValue()),
		}
	}

	// --- Observability --------------------------------------------------------
	// Presence of the configuration reference IS the enable switch -- there
	// is no separate toggle to drift out of sync. One-way on a live service
	// (upstream provider gap): the provider sends NOTHING for an absent
	// block on update, so removing the reference leaves tracing running at
	// AWS and the plan never converges. Sending a disabled block here
	// instead is NOT safe: AWS echoes NO observability member at all for a
	// tracing-less service (live DescribeService evidence, never
	// {ObservabilityEnabled: false}) and the provider flattens that nil to
	// an empty block, so an always-present disabled block perpetually
	// diffs on every tracing-less service. Disable requires service
	// replacement until the provider gains a disable path.
	if spec.ObservabilityConfigurationArn.GetValue() != "" {
		serviceArgs.ObservabilityConfiguration = &apprunner.ServiceObservabilityConfigurationArgs{
			ObservabilityEnabled:          pulumi.Bool(true),
			ObservabilityConfigurationArn: pulumi.StringPtr(spec.ObservabilityConfigurationArn.GetValue()),
		}
	}

	// --- Auto scaling -----------------------------------------------------------
	// Unset falls back to the account's default auto scaling configuration.
	if spec.AutoScalingConfigurationArn.GetValue() != "" {
		serviceArgs.AutoScalingConfigurationArn = pulumi.StringPtr(spec.AutoScalingConfigurationArn.GetValue())
	}

	createdService, err := apprunner.NewService(ctx, serviceName, serviceArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create App Runner service")
	}

	// --- Custom domain associations ---------------------------------------
	// One association per spec entry, keyed by domain name so entries add
	// and remove independently. The association returns as soon as
	// validation records are available -- it deliberately does not wait for
	// the domain to go active, because that requires DNS records this
	// module does not manage.
	domainOutputs := pulumi.Array{}
	for _, domain := range spec.CustomDomains {
		enableWww := true
		if domain.EnableWwwSubdomain != nil {
			enableWww = *domain.EnableWwwSubdomain
		}
		createdAssociation, err := apprunner.NewCustomDomainAssociation(ctx, serviceName+"-"+domain.DomainName, &apprunner.CustomDomainAssociationArgs{
			DomainName:         pulumi.String(domain.DomainName),
			ServiceArn:         createdService.Arn,
			EnableWwwSubdomain: pulumi.BoolPtr(enableWww),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to associate custom domain %q", domain.DomainName)
		}

		// Shaped key-for-key with the Terraform output so both engines
		// flatten onto the same outputs contract.
		records := createdAssociation.CertificateValidationRecords.ApplyT(func(recs []apprunner.CustomDomainAssociationCertificateValidationRecord) []map[string]string {
			out := make([]map[string]string, 0, len(recs))
			for _, r := range recs {
				record := map[string]string{"record_name": "", "record_type": "", "record_value": ""}
				if r.Name != nil {
					record["record_name"] = *r.Name
				}
				if r.Type != nil {
					record["record_type"] = *r.Type
				}
				if r.Value != nil {
					record["record_value"] = *r.Value
				}
				out = append(out, record)
			}
			return out
		})

		domainOutputs = append(domainOutputs, pulumi.Map{
			"domain_name":                    pulumi.String(domain.DomainName),
			"dns_target":                     createdAssociation.DnsTarget,
			"status":                         createdAssociation.Status,
			"certificate_validation_records": records,
		})
	}

	// --- WAF association ----------------------------------------------------
	// The protected resource points at the web ACL (the same direction
	// CloudFront models it) -- the association is glue with no identity of
	// its own, so it folds here rather than existing as a kind.
	if spec.WebAclArn.GetValue() != "" {
		_, err := wafv2.NewWebAclAssociation(ctx, serviceName+"-waf", &wafv2.WebAclAssociationArgs{
			ResourceArn: createdService.Arn,
			WebAclArn:   pulumi.String(spec.WebAclArn.GetValue()),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to associate WAF web ACL")
		}
	}

	// --- VPC Ingress Connections --------------------------------------------
	// One connection per spec entry, keyed by name so entries add and
	// remove independently. Each publishes the service into one VPC through
	// an interface VPC endpoint (AWS PrivateLink); clients inside that VPC
	// resolve the per-connection domain_name exported below. Every value is
	// create-time immutable by AWS design -- changing an entry replaces
	// that one connection.
	ingressOutputs := pulumi.Array{}
	for _, conn := range spec.VpcIngressConnections {
		createdConnection, err := apprunner.NewVpcIngressConnection(ctx, serviceName+"-"+conn.Name, &apprunner.VpcIngressConnectionArgs{
			Name:       pulumi.StringPtr(conn.Name),
			ServiceArn: createdService.Arn,
			IngressVpcConfiguration: &apprunner.VpcIngressConnectionIngressVpcConfigurationArgs{
				VpcId:         pulumi.StringPtr(conn.VpcId.GetValue()),
				VpcEndpointId: pulumi.StringPtr(conn.VpcEndpointId.GetValue()),
			},
			Tags: pulumi.ToStringMap(locals.AwsTags),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create VPC ingress connection %q", conn.Name)
		}

		// Shaped key-for-key with the Terraform output so both engines
		// flatten onto the same outputs contract.
		ingressOutputs = append(ingressOutputs, pulumi.Map{
			"name":        pulumi.String(conn.Name),
			"arn":         createdConnection.Arn,
			"domain_name": createdConnection.DomainName,
			"status":      createdConnection.Status,
		})
	}

	// Export outputs matching AwsAppRunnerServiceOutputs.
	ctx.Export(OpServiceArn, createdService.Arn)
	ctx.Export(OpServiceId, createdService.ServiceId)
	ctx.Export(OpServiceUrl, createdService.ServiceUrl)
	ctx.Export(OpServiceName, createdService.ServiceName)
	ctx.Export(OpServiceStatus, createdService.Status)
	ctx.Export(OpCustomDomains, domainOutputs)
	ctx.Export(OpVpcIngressConnections, ingressOutputs)

	return nil
}
