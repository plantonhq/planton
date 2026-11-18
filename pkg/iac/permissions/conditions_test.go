package permissions

import (
	"encoding/json"
	"os"
	"testing"
)

// The case table is the platform's, copied verbatim: the platform's
// PermissionConditionsTest and the console's permission-conditions test read
// the same cases, so a condition decides the same way in all three readers.
func TestConditionNeededAgreesWithTheSharedCaseTable(t *testing.T) {
	raw, err := os.ReadFile("testdata/permission-conditions.cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		Cases []struct {
			Name      string                   `json:"name"`
			Condition *string                  `json:"condition"`
			Specs     []map[string]interface{} `json:"specs"`
			Needed    bool                     `json:"needed"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		t.Fatal(err)
	}
	if len(table.Cases) == 0 {
		t.Fatal("the case table holds no cases")
	}
	for _, c := range table.Cases {
		condition := ""
		if c.Condition != nil {
			condition = *c.Condition
		}
		if got := ConditionNeeded(condition, c.Specs); got != c.Needed {
			t.Errorf("%s: needed = %v, want %v", c.Name, got, c.Needed)
		}
	}
}
