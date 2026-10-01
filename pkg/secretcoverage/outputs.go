//go:build !codegen
// +build !codegen

package secretcoverage

import (
	"fmt"
	"sort"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A stack output marked `sensitive` is a secret the resource generates. Which outputs are secrets
// cannot be read off their names (an `access_key` is a key id in one kind and a PEM private key
// in another, and a credential-bearing `uri` looks like any address), so the name heuristic
// above stays on the spec, and `planton module verify` holds both engines to the schema's marks
// instead. What this file pins is the shape those marks may take, because the engines and the
// platform both decide secrecy per top-level output:
//
//   - `sensitive` sits only on a top-level field of the kind's StackOutputs;
//   - `sensitive_exempt_reason` never appears in outputs (it answers the name heuristic, which
//     does not run there);
//   - a marked output carries no value rule, since on Planton it holds a `$secret/` reference.

// OutputShapeViolation is one output field whose marks break the shape above.
type OutputShapeViolation struct {
	Kind   string
	Path   string
	Reason string
}

// OutputShapeViolations walks every production kind's stack outputs.
func OutputShapeViolations() []OutputShapeViolation {
	var out []OutputShapeViolation
	for _, kind := range crkreflect.KindsList() {
		provider := crkreflect.GetProvider(kind)
		if provider == cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified || provider.String()[0] == '_' {
			continue
		}
		msg, err := crkreflect.NewInstance(kind)
		if err != nil {
			continue
		}
		status := msg.ProtoReflect().Descriptor().Fields().ByName("status")
		if status == nil || status.Message() == nil {
			continue
		}
		outputs := status.Message().Fields().ByName("outputs")
		if outputs == nil || outputs.Message() == nil {
			continue
		}
		out = append(out, collectOutputShapeViolations(outputs.Message(), kind.String())...)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Path < out[j].Path
	})
	return out
}

// collectOutputShapeViolations walks one StackOutputs message descriptor, rooting paths at
// "status.outputs". Exposed to the tests through the package.
func collectOutputShapeViolations(outputsMd protoreflect.MessageDescriptor, kindName string) []OutputShapeViolation {
	var out []OutputShapeViolation
	visited := map[protoreflect.FullName]bool{}
	var walkOutputs func(md protoreflect.MessageDescriptor, prefix string, topLevel bool)
	walkOutputs = func(md protoreflect.MessageDescriptor, prefix string, topLevel bool) {
		if visited[md.FullName()] {
			return
		}
		visited[md.FullName()] = true
		fields := md.Fields()
		for i := 0; i < fields.Len(); i++ {
			fd := fields.Get(i)
			path := prefix + "." + string(fd.Name())
			sensitive, exemptReason := leafOptions(fd)
			if sensitive && !topLevel {
				out = append(out, OutputShapeViolation{kindName, path,
					"a secret output must be a top-level stack-outputs field: both engines decide secrecy per top-level output"})
			}
			if exemptReason != "" {
				out = append(out, OutputShapeViolation{kindName, path,
					"sensitive_exempt_reason answers the spec's name heuristic and has no meaning on an output"})
			}
			if sensitive {
				if rule := valueContentRule(fd); rule != "" {
					out = append(out, OutputShapeViolation{kindName, path, fmt.Sprintf(
						"a secret output holds a $secret/ reference on Planton, so its value rule (%s) would refuse it", rule)})
				}
			}
			var next protoreflect.MessageDescriptor
			if fd.IsMap() {
				next = fd.MapValue().Message()
			} else {
				next = fd.Message()
			}
			if next != nil {
				walkOutputs(next, path, false)
			}
		}
	}
	walkOutputs(outputsMd, "status.outputs", true)
	return out
}
