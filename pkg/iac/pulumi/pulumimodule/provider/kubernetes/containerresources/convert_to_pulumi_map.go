package containerresources

import (
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ConvertToPulumiMap(resources *kubernetes.ContainerResources) pulumi.Map {
	return pulumi.Map{
		"limits": pulumi.Map{
			"cpu":    pulumi.String(resources.Limits.Cpu),
			"memory": pulumi.String(resources.Limits.Memory),
		},
		"requests": pulumi.Map{
			"cpu":    pulumi.String(resources.Requests.Cpu),
			"memory": pulumi.String(resources.Requests.Memory),
		},
	}
}
