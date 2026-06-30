package valuefrom

import v1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

func ToStringArray(input []*v1.StringValueOrRef) []string {
	resp := make([]string, 0)
	if input == nil || len(input) == 0 {
		return resp
	}
	for _, item := range input {
		if item == nil || item.GetValue() == "" {
			continue
		}
		resp = append(resp, item.GetValue())
	}
	return resp
}
