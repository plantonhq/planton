package labelkeys

import (
	"testing"
)

func TestLabelConversionPrometheusFormat(t *testing.T) {
	testCases := []struct {
		testName                string
		inputLabel              string
		expectedPrometheusLabel string
	}{
		{
			testName:                "planton org label should be converted to prometheus label",
			inputLabel:              "planton.dev/org",
			expectedPrometheusLabel: "planton_org_org",
		},
		{
			testName:                "planton service label should be converted to prometheus label",
			inputLabel:              "planton.dev/service",
			expectedPrometheusLabel: "planton_org_service",
		},
		{
			testName:                "planton service-env label should be converted to prometheus label",
			inputLabel:              "planton.dev/env",
			expectedPrometheusLabel: "planton_org_env",
		},
		{
			testName:                "planton kind label should be converted to prometheus label",
			inputLabel:              "planton.dev/kind",
			expectedPrometheusLabel: "planton_org_kind",
		},
		{
			testName:                "planton id label should be converted to prometheus label",
			inputLabel:              "planton.dev/id",
			expectedPrometheusLabel: "planton_org_id",
		},
	}
	t.Run("test planton label conversion to prometheus format labels", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.testName, func(t *testing.T) {
				r := WithPrometheusFormat(tc.inputLabel)
				if r != tc.expectedPrometheusLabel {
					t.Errorf("expected: %s, got: %s", tc.expectedPrometheusLabel, r)
				}
			})
		}
	})
}
